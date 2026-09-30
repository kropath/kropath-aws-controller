// Copyright 2026 kropath Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package util provides shared helpers for reconcilers.
package util

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/kropath/kropath-aws-controller/api/v1alpha1"
	"github.com/kropath/kropath-aws-controller/internal/metrics"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// GlobalConfigNamespaceAnnotation is the annotation key on a Namespace that
	// designates which namespace holds the global-tier KropathConfig for resources
	// in that namespace. Its presence is what makes a namespace a "resource
	// namespace" (RoleResource); its absence makes it "governance-only"
	// (RoleGovernanceOnly) -- there is no default (spec §5.5, AC-12).
	GlobalConfigNamespaceAnnotation = "aws.kropath.run/global-config-namespace"

	// KropathConfigName is the enforced singleton name for KropathConfig objects,
	// in both the global and the local tier (ADR-018 D-1, ADR-015 §5.7). The CRD
	// rejects any other name via x-kubernetes-validations. Role is determined by
	// namespace, never by this name.
	KropathConfigName = "baseline"

	// OwnerAccountIDAnnotation is the ACK-native namespace annotation kropath
	// reads for the resource's AWS account (spec §4.4). Written by the platform
	// operator, never by kropath.
	OwnerAccountIDAnnotation = "services.k8s.aws/owner-account-id"

	// DefaultRegionAnnotation is the ACK-native namespace annotation kropath
	// reads for the resource's AWS region (spec §4.4).
	DefaultRegionAnnotation = "services.k8s.aws/default-region"

	// TeamIDAnnotation, when present on a resource namespace, means ACK's
	// TeamLevelCARM gate resolves the account from the team map instead of
	// owner-account-id -- kropath cannot observe that resolution, so it rejects
	// rather than silently computing an account ACK will never use (spec §6.3).
	TeamIDAnnotation = "services.k8s.aws/team-id"

	// PlacementStatusAnnotation is the controller-owned verdict annotation
	// published on every resource Namespace (spec §4.4, §6.2). It is write-only
	// for the controller: never read back as input, so a hand edit is overwritten
	// on the next evaluation rather than believed.
	PlacementStatusAnnotation = "aws.kropath.run/placement-status"

	// PlacementFieldManager is the server-side-apply field manager that owns
	// PlacementStatusAnnotation, so a GitOps tool managing the Namespace never
	// contends with the controller over the object (spec §6.2).
	PlacementFieldManager = "kropath-placement"

	// PlacementStatusOK is the value published on PlacementStatusAnnotation when
	// placement resolves successfully.
	PlacementStatusOK = "ok"
)

// Reason codes for placement resolution (spec §6.2, §6.3, §6.4). These strings are
// normative: they appear verbatim as condition/annotation/Event reasons and are
// asserted directly by AC-3 - AC-6, AC-8, AC-10.
const (
	ReasonTeamAnnotationUnsupported = "TeamAnnotationUnsupported"
	ReasonMissingAccountAnnotation  = "MissingAccountAnnotation"
	ReasonInvalidAccountAnnotation  = "InvalidAccountAnnotation"
	ReasonMissingRegionAnnotation   = "MissingRegionAnnotation"
	ReasonNamespaceUnreadable       = "NamespaceUnreadable"
	ReasonGlobalTierInput           = "GlobalTierInput"
	ReasonResolvedFromNamespace     = "ResolvedFromNamespace"
	ReasonPlacementResolved         = "PlacementResolved"
)

// PlacementReasons returns the closed set of reason values
// kropath_placement_resolutions_total can carry (spec §2.1, §2.2, §5.2) --
// the eight values above, in declaration order. ResolvedFromNamespace is
// included because it is a normative member of the contract set even though
// no call site ever emits it (spec §2.2): omitting it here would make the
// §10.2.1 label-value check reject the design §9 matcher that names it.
func PlacementReasons() []string {
	return []string{
		ReasonTeamAnnotationUnsupported,
		ReasonMissingAccountAnnotation,
		ReasonInvalidAccountAnnotation,
		ReasonMissingRegionAnnotation,
		ReasonNamespaceUnreadable,
		ReasonGlobalTierInput,
		ReasonResolvedFromNamespace,
		ReasonPlacementResolved,
	}
}

// PlacementResolvedConditionType is the condition every <ResourceFamily>Config
// reconciler publishes alongside Reconciled, naming which placement was resolved
// and from where (spec §7.2).
const PlacementResolvedConditionType = "PlacementResolved"

// accountIDPattern is the shape check for services.k8s.aws/owner-account-id (spec
// §5.1): form only, anchored, no STS call and no existence check.
var accountIDPattern = regexp.MustCompile(`^[0-9]{12}$`)

// NamespaceRole classifies a Namespace for placement purposes (spec §5.5).
type NamespaceRole int

const (
	// RoleResource is a namespace carrying GlobalConfigNamespaceAnnotation
	// (non-empty). The placement gate applies and its <ResourceFamily>Config
	// objects receive status.effectiveConfig.
	RoleResource NamespaceRole = iota
	// RoleGovernanceOnly is a namespace lacking GlobalConfigNamespaceAnnotation
	// (absent or ""). It is exempt from every placement rule; its
	// <ResourceFamily>Config objects are global-tier inputs and never receive
	// status.effectiveConfig (spec §6.4).
	RoleGovernanceOnly
)

// PlacementError carries a normative reason code and rendered message returned by
// ResolvePlacement (spec §5.6, §6.2). It is not a NamespaceUnreadable condition --
// that case is reported through the separate `error` return of ResolveNamespaceRole
// and ResolvePlacement, since a Namespace that cannot be read cannot be evaluated
// against any rule (spec §6.3).
type PlacementError struct {
	Reason  string
	Message string
}

func (e *PlacementError) Error() string { return e.Message }

// ResolveNamespaceRole reports whether namespace is a resource namespace and, if
// so, which namespace holds its global tier (spec §5.5). A read failure is an
// error and is never reported as a role -- the previous ResolveGlobalNamespace
// collapsed "annotation absent" and "namespace unreadable" into the same silent
// kro-system fallback; this replacement separates all three explicitly.
func ResolveNamespaceRole(ctx context.Context, c client.Client, namespace string) (role NamespaceRole, globalNS string, err error) {
	var ns corev1.Namespace
	if err := c.Get(ctx, types.NamespacedName{Name: namespace}, &ns); err != nil {
		return RoleGovernanceOnly, "", err
	}
	ann := ns.Annotations[GlobalConfigNamespaceAnnotation]
	if ann == "" {
		return RoleGovernanceOnly, "", nil
	}
	return RoleResource, ann, nil
}

// DerivePartition maps a region to its AWS partition by longest-prefix match
// (spec §4.3). Total over non-empty input; returns "" only for "". Matching is
// longest-prefix, not first-match: us-isob-east-1 also has the prefix us-iso-, so
// candidates are enumerated longest-first to avoid returning aws-iso for it.
func DerivePartition(region string) string {
	if region == "" {
		return ""
	}
	// Ordered longest-prefix-first. Do not reorder without preserving this
	// property -- see the us-isob-east-1 / us-iso- trap above and AC-11.
	switch {
	case strings.HasPrefix(region, "us-gov-"):
		return "aws-us-gov"
	case strings.HasPrefix(region, "cn-"):
		return "aws-cn"
	case strings.HasPrefix(region, "us-isob-"):
		return "aws-iso-b"
	case strings.HasPrefix(region, "us-isof-"), strings.HasPrefix(region, "eu-isof-"):
		return "aws-iso-f"
	case strings.HasPrefix(region, "eu-isoe-"):
		return "aws-iso-e"
	case strings.HasPrefix(region, "us-iso-"):
		return "aws-iso"
	default:
		return "aws"
	}
}

// ResolvePlacement resolves the complete AWS placement identity for a resource
// namespace, evaluating the reason codes of spec §6.2 in fixed order (first
// match wins): TeamAnnotationUnsupported, MissingAccountAnnotation,
// InvalidAccountAnnotation, MissingRegionAnnotation. It is one function called
// from both the namespace reconciler and every <ResourceFamily>Config reconciler
// (spec §5.6) so the two publication sites can never drift.
//
// Callers must only invoke this for a namespace already known to be
// RoleResource (i.e. after ResolveNamespaceRole). The `error` return is
// reserved for a Namespace Get failure -- the NamespaceUnreadable case (spec
// §6.3) -- and is distinct from the *PlacementError return, which reports an
// evaluated-and-failed rule.
func ResolvePlacement(ctx context.Context, c client.Client, namespace string) (v1alpha1.ProviderIdentity, *PlacementError, error) {
	var ns corev1.Namespace
	if err := c.Get(ctx, types.NamespacedName{Name: namespace}, &ns); err != nil {
		metrics.PlacementResolved(ReasonNamespaceUnreadable)
		return v1alpha1.ProviderIdentity{}, nil, err
	}

	if _, hasTeamID := ns.Annotations[TeamIDAnnotation]; hasTeamID {
		metrics.PlacementResolved(ReasonTeamAnnotationUnsupported)
		return v1alpha1.ProviderIdentity{}, &PlacementError{
			Reason: ReasonTeamAnnotationUnsupported,
			Message: fmt.Sprintf(
				`namespace %q carries annotation "services.k8s.aws/team-id"; under ACK's TeamLevelCARM the owner-account-id branch never executes, so kropath cannot resolve the account. Team-level CARM requires its own ADR (ADR-015 §5.8.2 rule 5)`,
				namespace),
		}, nil
	}

	accountID := ns.Annotations[OwnerAccountIDAnnotation]
	if accountID == "" {
		metrics.PlacementResolved(ReasonMissingAccountAnnotation)
		return v1alpha1.ProviderIdentity{}, &PlacementError{
			Reason: ReasonMissingAccountAnnotation,
			Message: fmt.Sprintf(
				`namespace %q is a resource namespace but does not carry annotation "services.k8s.aws/owner-account-id"; placement cannot be resolved and no effectiveConfig is written`,
				namespace),
		}, nil
	}
	if !accountIDPattern.MatchString(accountID) {
		metrics.PlacementResolved(ReasonInvalidAccountAnnotation)
		return v1alpha1.ProviderIdentity{}, &PlacementError{
			Reason: ReasonInvalidAccountAnnotation,
			Message: fmt.Sprintf(
				`namespace %q annotation "services.k8s.aws/owner-account-id" has value %q, which is not a 12-digit AWS account ID`,
				namespace, accountID),
		}, nil
	}

	region := ns.Annotations[DefaultRegionAnnotation]
	if region == "" {
		metrics.PlacementResolved(ReasonMissingRegionAnnotation)
		return v1alpha1.ProviderIdentity{}, &PlacementError{
			Reason: ReasonMissingRegionAnnotation,
			Message: fmt.Sprintf(
				`namespace %q is a resource namespace but does not carry annotation "services.k8s.aws/default-region"; without it ACK falls through to the controller's own --aws-region flag, which kropath cannot observe`,
				namespace),
		}, nil
	}

	metrics.PlacementResolved(ReasonPlacementResolved)
	return v1alpha1.ProviderIdentity{
		AccountID: accountID,
		Region:    region,
		Partition: DerivePartition(region),
	}, nil, nil
}

// NamespaceUnreadableMessage renders the byte-identical §6.2 message for a
// Namespace Get failure at a <ResourceFamily>Config reconcile (spec §6.3 -- this
// verdict is config-object-site only and never arises on the Namespace itself).
func NamespaceUnreadableMessage(namespace string, err error) string {
	return fmt.Sprintf(`namespace %q could not be read: %v; effectiveConfig withheld and reconcile requeued`, namespace, err)
}

// GlobalTierInputMessage renders the byte-identical §6.4 message published on a
// governance-only namespace's <ResourceFamily>Config objects.
func GlobalTierInputMessage(namespace string) string {
	return fmt.Sprintf(
		`namespace %q does not carry annotation "aws.kropath.run/global-config-namespace"; objects in it are global-tier inputs and receive no effectiveConfig`,
		namespace)
}

// PlacementProvenanceMessage renders the §7.2 PlacementResolved success message,
// naming which namespace and annotations the identity was resolved from.
func PlacementProvenanceMessage(identity v1alpha1.ProviderIdentity, namespace string) string {
	return fmt.Sprintf(
		`account %q and region %q resolved from namespace %q annotations "services.k8s.aws/owner-account-id" and "services.k8s.aws/default-region"; partition %q derived from region`,
		identity.AccountID, identity.Region, namespace, identity.Partition)
}

// FamilyPlacementResult is what every <ResourceFamily>Config reconciler needs
// from namespace-role and placement resolution in one round trip (spec §5.6). It
// exposes every case as data instead of requiring each of the 57 reconcilers to
// reimplement the branching, so the AC-10 byte-identical-message guarantee holds
// by construction rather than by convention.
type FamilyPlacementResult struct {
	// Role is RoleResource or RoleGovernanceOnly.
	Role NamespaceRole
	// GlobalNamespace is the resolved global-tier config namespace. Only
	// meaningful when Role == RoleResource.
	GlobalNamespace string
	// Identity is the resolved AWS identity. Only valid when Role ==
	// RoleResource and ReconciledOverride == nil.
	Identity v1alpha1.ProviderIdentity
	// PlacementCondition is the PlacementResolved condition to publish (spec
	// §7.2). Populated only when Role == RoleResource -- a governance-only
	// object gets no PlacementResolved condition at all, since nothing was
	// evaluated for it (spec §6.4: "placement rules not evaluated").
	PlacementCondition *metav1.Condition
	// ReconciledOverride, when non-nil, is the Reconciled condition the caller
	// must publish INSTEAD OF the normal CascadeMerged success condition, and
	// signals that status.effectiveConfig must be withheld (placement failure)
	// or cleared (governance-only) rather than written from the cascade merge.
	ReconciledOverride *metav1.Condition
}

// ResolveFamilyPlacement resolves role and (for resource namespaces) placement
// for a <ResourceFamily>Config reconcile. The returned error is non-nil only for
// a Namespace Get failure (NamespaceUnreadable, spec §6.3): callers must publish
// Reconciled=False/NamespaceUnreadable, withhold effectiveConfig, attempt no
// Namespace patch, and return the error so controller-runtime requeues with
// backoff.
func ResolveFamilyPlacement(ctx context.Context, c client.Client, namespace string, generation int64, now metav1.Time) (FamilyPlacementResult, error) {
	role, globalNS, err := ResolveNamespaceRole(ctx, c, namespace)
	if err != nil {
		return FamilyPlacementResult{}, err
	}

	if role == RoleGovernanceOnly {
		metrics.PlacementResolved(ReasonGlobalTierInput)
		cond := &metav1.Condition{
			Type:               "Reconciled",
			Status:             metav1.ConditionTrue,
			Reason:             ReasonGlobalTierInput,
			Message:            GlobalTierInputMessage(namespace),
			ObservedGeneration: generation,
			LastTransitionTime: now,
		}
		return FamilyPlacementResult{Role: RoleGovernanceOnly, ReconciledOverride: cond}, nil
	}

	identity, placementErr, err := ResolvePlacement(ctx, c, namespace)
	if err != nil {
		return FamilyPlacementResult{}, err
	}
	if placementErr != nil {
		reconciledCond := &metav1.Condition{
			Type:               "Reconciled",
			Status:             metav1.ConditionFalse,
			Reason:             placementErr.Reason,
			Message:            placementErr.Message,
			ObservedGeneration: generation,
			LastTransitionTime: now,
		}
		placementCond := &metav1.Condition{
			Type:               PlacementResolvedConditionType,
			Status:             metav1.ConditionFalse,
			Reason:             placementErr.Reason,
			Message:            placementErr.Message,
			ObservedGeneration: generation,
			LastTransitionTime: now,
		}
		return FamilyPlacementResult{
			Role:               RoleResource,
			GlobalNamespace:    globalNS,
			PlacementCondition: placementCond,
			ReconciledOverride: reconciledCond,
		}, nil
	}

	placementCond := &metav1.Condition{
		Type:               PlacementResolvedConditionType,
		Status:             metav1.ConditionTrue,
		Reason:             ReasonResolvedFromNamespace,
		Message:            PlacementProvenanceMessage(identity, namespace),
		ObservedGeneration: generation,
		LastTransitionTime: now,
	}
	return FamilyPlacementResult{
		Role:               RoleResource,
		GlobalNamespace:    globalNS,
		Identity:           identity,
		PlacementCondition: placementCond,
	}, nil
}

// NamespaceUnreadableCondition builds the byte-identical Reconciled and
// PlacementResolved conditions a <ResourceFamily>Config reconciler publishes
// when ResolveFamilyPlacement returns a non-nil error (spec §6.3).
func NamespaceUnreadableCondition(namespace string, err error, generation int64, now metav1.Time) (reconciled, placement metav1.Condition) {
	msg := NamespaceUnreadableMessage(namespace, err)
	reconciled = metav1.Condition{
		Type:               "Reconciled",
		Status:             metav1.ConditionFalse,
		Reason:             ReasonNamespaceUnreadable,
		Message:            msg,
		ObservedGeneration: generation,
		LastTransitionTime: now,
	}
	placement = metav1.Condition{
		Type:               PlacementResolvedConditionType,
		Status:             metav1.ConditionFalse,
		Reason:             ReasonNamespaceUnreadable,
		Message:            msg,
		ObservedGeneration: generation,
		LastTransitionTime: now,
	}
	return reconciled, placement
}
