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

// Package kropathconfigstatus publishes a Reconciled condition onto every
// KropathConfig object, naming which tier it was interpreted as and how many
// <ResourceFamily>Config kinds actually consume it (KRO-1121).
//
// Before this reconciler existed, every other cascade reconciler read
// KropathConfig but never wrote back to it, so a KropathConfig that no
// <ResourceFamily>Config resolves — wrong namespace, or created before the
// singleton name rule existed — was silently indistinguishable from a
// correctly-resolved one. That invisibility, combined with an ArgoCD health
// check that reported every KropathConfig Healthy unconditionally, wedged
// every tenant Application for ~3.5h during KRO-1104.
package kropathconfigstatus

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/kropath/kropath-aws-controller/api/v1alpha1"
	"github.com/kropath/kropath-aws-controller/internal/features"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// ReconciledConditionType is the condition this reconciler publishes.
const ReconciledConditionType = "Reconciled"

// Reason codes for the Reconciled condition this reconciler publishes (spec
// §2.1). These strings are normative: they appear verbatim as the
// condition's Reason and are the closed set TierReasons() returns.
const (
	ReasonGlobalAndLocalTier = "GlobalAndLocalTier"
	ReasonGlobalTier         = "GlobalTier"
	ReasonLocalTier          = "LocalTier"
	ReasonUnreferenced       = "Unreferenced"
)

// TierReasons returns the closed set of reason values
// kropath_kropathconfigstatus_configs can carry (spec §2.1), in the order
// reconciledCondition evaluates them.
func TierReasons() []string {
	return []string{ReasonGlobalAndLocalTier, ReasonGlobalTier, ReasonLocalTier, ReasonUnreferenced}
}

// requeueInterval re-evaluates consumer counts even though this reconciler
// does not watch the ~57 family Config kinds directly (watching all of them
// would add this repo's §1 missing-CRD-kills-the-manager risk to every
// family, for a status field that can tolerate being briefly stale). A
// KropathConfig's own create/update/delete still triggers an immediate
// reconcile via the Watch below; this interval only covers the case where a
// family config appears or disappears without the KropathConfig itself
// changing.
const requeueInterval = 15 * time.Second

type Reconciler struct {
	// Client is used for the Get/Status().Update() of the KropathConfig this
	// reconciler owns; it goes through the manager's cache like every other
	// reconciler's client.
	Client client.Client
	// Reader lists the ~57 family Config kinds. It is deliberately the
	// manager's *uncached* API reader, not Client: a cached List() for a kind
	// with no running informer (e.g. a family whose CRD briefly is not
	// installed) starts one lazily and blocks the calling reconcile until it
	// syncs — which, per docs/frequent-chainsaw-errors.md §1, never happens
	// for a kind the API server does not serve. A direct read instead gets a
	// same-call NoKindMatchError, which listFamilyConfig already treats as
	// zero consumers.
	Reader client.Reader
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	kpc := &v1alpha1.KropathConfig{}
	kpc.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("KropathConfig"))
	if err := r.Client.Get(ctx, req.NamespacedName, kpc); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	globalKinds, localKinds, err := r.consumingFamilyKinds(ctx, kpc.Namespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("kropathconfigstatus: determining consumers of %s/%s: %w", kpc.Namespace, kpc.Name, err)
	}

	now := metav1.Now()
	newCond := reconciledCondition(globalKinds, localKinds, kpc.Generation, now)

	if !conditionNeedsUpdate(kpc.Status.Conditions, newCond) && kpc.Status.ObservedGeneration == kpc.Generation {
		return ctrl.Result{RequeueAfter: requeueInterval}, nil
	}

	kpc.Status.Conditions = setCondition(kpc.Status.Conditions, newCond)
	kpc.Status.ObservedGeneration = kpc.Generation
	kpc.Status.SyncedTimestamp = now.UTC().Format(time.RFC3339)

	if err := r.Client.Status().Update(ctx, kpc); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: requeueInterval}, nil
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	_, err := r.BuildWithManager(mgr)
	return err
}

func (r *Reconciler) BuildWithManager(mgr ctrl.Manager) (controller.Controller, error) {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.KropathConfig{}).
		Build(r)
}

// consumingFamilyKinds returns the sorted family Config kind names that
// resolve kpcNamespace as their global tier, and the ones that resolve it as
// their local tier. A kind can appear in both (kpcNamespace happens to also
// be some other namespace's resolved global-config-namespace) or in neither
// — the "unreferenced" case this reconciler exists to surface.
func (r *Reconciler) consumingFamilyKinds(ctx context.Context, kpcNamespace string) (global, local []string, err error) {
	globalNSCache := map[string]string{}
	resolveGlobal := func(ns string) string {
		if g, ok := globalNSCache[ns]; ok {
			return g
		}
		_, g, err := util.ResolveNamespaceRole(ctx, r.Client, ns)
		if err != nil {
			g = ""
		}
		globalNSCache[ns] = g
		return g
	}

	for _, kind := range familyConfigKinds() {
		items, err := r.listFamilyConfig(ctx, kind)
		if err != nil {
			return nil, nil, fmt.Errorf("listing %s: %w", kind, err)
		}

		var hasLocal, hasGlobal bool
		for _, item := range items {
			ns := item.GetNamespace()
			if ns == kpcNamespace {
				hasLocal = true
			}
			if resolveGlobal(ns) == kpcNamespace {
				hasGlobal = true
			}
			if hasLocal && hasGlobal {
				break
			}
		}
		if hasLocal {
			local = append(local, kind)
		}
		if hasGlobal {
			global = append(global, kind)
		}
	}

	sort.Strings(global)
	sort.Strings(local)
	return global, local, nil
}

// listFamilyConfig lists every instance of a <ResourceFamily>Config kind
// cluster-wide, using the scheme to construct the right typed *List object
// generically instead of a 57-way type switch.
func (r *Reconciler) listFamilyConfig(ctx context.Context, kind string) ([]metav1.Object, error) {
	listGVK := v1alpha1.GroupVersion.WithKind(kind + "List")
	obj, err := r.Scheme.New(listGVK)
	if err != nil {
		return nil, fmt.Errorf("constructing %s: %w", listGVK, err)
	}
	list, ok := obj.(client.ObjectList)
	if !ok {
		return nil, fmt.Errorf("%s does not implement client.ObjectList", listGVK)
	}

	if err := r.Reader.List(ctx, list); err != nil {
		if apimeta.IsNoMatchError(err) || apierrors.IsNotFound(err) {
			// This family's CRD is not installed in this cluster. Every other
			// cascade reconciler already treats its own missing CRD as fatal to
			// the whole manager (docs/frequent-chainsaw-errors.md §1) — but that
			// rule exists for *watches*, not for this reconciler's uncached,
			// on-demand List, so degrade to "zero consumers of this kind"
			// instead of failing the whole KropathConfig status reconcile.
			return nil, nil
		}
		return nil, err
	}

	items, err := apimeta.ExtractList(list)
	if err != nil {
		return nil, fmt.Errorf("extracting %s items: %w", listGVK, err)
	}
	out := make([]metav1.Object, 0, len(items))
	for _, item := range items {
		accessor, err := apimeta.Accessor(item)
		if err != nil {
			return nil, err
		}
		out = append(out, accessor)
	}
	return out, nil
}

// familyConfigKinds derives the list of <ResourceFamily>Config kinds from
// features.All instead of a hand-maintained list, so a new cascade family is
// picked up automatically instead of silently missing from consumer counts.
func familyConfigKinds() []string {
	kinds := make([]string, 0, len(features.All))
	for _, entry := range features.All {
		if entry.Name == "KropathConfig" || !strings.HasSuffix(entry.Name, "Config") {
			continue
		}
		kinds = append(kinds, entry.Name)
	}
	return kinds
}

// reconciledCondition builds the Reconciled condition from the observed
// consumers. Status is True whenever at least one family config resolves
// this object as either tier, and False when none does (KRO-1121's
// "unreferenced" case).
func reconciledCondition(globalKinds, localKinds []string, observedGeneration int64, now metav1.Time) metav1.Condition {
	cond := metav1.Condition{
		Type:               ReconciledConditionType,
		ObservedGeneration: observedGeneration,
		LastTransitionTime: now,
	}

	switch {
	case len(globalKinds) > 0 && len(localKinds) > 0:
		cond.Status = metav1.ConditionTrue
		cond.Reason = ReasonGlobalAndLocalTier
		cond.Message = fmt.Sprintf(
			"Resolved as the global tier by %d family config kind(s) (%s) and the local tier by %d (%s).",
			len(globalKinds), strings.Join(globalKinds, ", "), len(localKinds), strings.Join(localKinds, ", "))
	case len(globalKinds) > 0:
		cond.Status = metav1.ConditionTrue
		cond.Reason = ReasonGlobalTier
		cond.Message = fmt.Sprintf(
			"Resolved as the global tier by %d family config kind(s): %s.",
			len(globalKinds), strings.Join(globalKinds, ", "))
	case len(localKinds) > 0:
		cond.Status = metav1.ConditionTrue
		cond.Reason = ReasonLocalTier
		cond.Message = fmt.Sprintf(
			"Resolved as the local tier by %d family config kind(s): %s.",
			len(localKinds), strings.Join(localKinds, ", "))
	default:
		cond.Status = metav1.ConditionFalse
		cond.Reason = ReasonUnreferenced
		cond.Message = "No <ResourceFamily>Config resolves this KropathConfig as its global or local tier. " +
			"Verify the namespace and the aws.kropath.run/global-config-namespace annotation."
	}
	return cond
}

// conditionNeedsUpdate returns true when no existing condition matches Type+Status+Reason+Message.
func conditionNeedsUpdate(conditions []metav1.Condition, newCond metav1.Condition) bool {
	for _, c := range conditions {
		if c.Type == newCond.Type {
			return c.Status != newCond.Status || c.Reason != newCond.Reason || c.Message != newCond.Message
		}
	}
	return true
}

// setCondition upserts a condition by Type, preserving LastTransitionTime when status is unchanged.
func setCondition(conditions []metav1.Condition, newCond metav1.Condition) []metav1.Condition {
	for i, c := range conditions {
		if c.Type == newCond.Type {
			if c.Status == newCond.Status {
				newCond.LastTransitionTime = c.LastTransitionTime
			}
			conditions[i] = newCond
			return conditions
		}
	}
	return append(conditions, newCond)
}
