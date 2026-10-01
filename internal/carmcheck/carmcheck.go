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

// Package carmcheck inspects an existing ACK/kro install and reports whether
// it satisfies the CARM preconditions kropath's account/region placement
// model depends on (ADR-015 §5.8.4, KRO-1139 C-3/C-4).
//
// kropath has no runtime dependency on CARM but a total correctness
// dependency on it: kropath and ACK are two independent resolvers of the
// same placement question, and they agree only when every precondition here
// holds. When one is violated, kropath computes a name and a predictedArn
// that describe a resource ACK places somewhere else, and nothing errors —
// the exact bug class ADR-019 exists to eliminate.
//
// This package is a best-effort, install-time diagnostic, never a runtime
// gate. ACK's own configuration (its --enable-carm and --watch-namespace
// flags in particular) can come from a flag, an environment variable, or a
// Helm feature gate; reading Deployment container args, as this package
// does, can therefore produce false negatives (SeverityUnknown), never a
// false sense of confidence. Findings are meant to be surfaced to an
// operator, not consumed by a reconcile loop.
package carmcheck

import (
	"context"
	"fmt"
	"regexp"
	"sort"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kropath/kropath-aws-controller/internal/reconciler/util"
)

// Severity classifies a Finding's urgency.
type Severity string

const (
	// SeverityBlocking means kropath and ACK will disagree about placement
	// for at least one resource — the computed name/predictedArn is wrong.
	SeverityBlocking Severity = "blocking"
	// SeverityWarning means a configuration is suspicious but not confirmed
	// to cause a placement mismatch.
	SeverityWarning Severity = "warning"
	// SeverityOK means the precondition holds as far as this checker can tell.
	SeverityOK Severity = "ok"
	// SeverityUnknown means the checker could not determine the answer from
	// what it can read. ADR-015 §5.8.4: "kropath does not and cannot verify
	// [these preconditions]" — this is the expected outcome for anything
	// configured outside a Deployment's container args.
	SeverityUnknown Severity = "unknown"
)

// Finding is one precondition's verdict.
type Finding struct {
	// Check names the precondition, e.g. "enable-carm", "role-account-map".
	Check    string
	Severity Severity
	Message  string
	// Namespace is set when the finding concerns one specific
	// kropath-managed namespace, and empty for an install-wide finding.
	Namespace string
}

// Report is the full result of a conformance Run.
type Report struct {
	Findings []Finding
}

// HasBlocking reports whether any finding is SeverityBlocking.
func (r *Report) HasBlocking() bool {
	for _, f := range r.Findings {
		if f.Severity == SeverityBlocking {
			return true
		}
	}
	return false
}

// HasUnknown reports whether any finding is SeverityUnknown.
func (r *Report) HasUnknown() bool {
	for _, f := range r.Findings {
		if f.Severity == SeverityUnknown {
			return true
		}
	}
	return false
}

// Options configures a Run.
type Options struct {
	// ACKNamespace is where the ACK controller Deployments and the
	// ack-role-account-map ConfigMap live. Required.
	ACKNamespace string
	// RoleAccountMapName overrides the default ConfigMap name
	// ("ack-role-account-map").
	RoleAccountMapName string
	// Namespaces overrides namespace auto-discovery with an explicit list.
	// When empty, Run discovers kropath-managed resource namespaces by
	// listing namespaces that carry util.GlobalConfigNamespaceAnnotation
	// (ADR-015 §5.8.2 rule 2 / KRO-1139 H-5: a namespace carrying that
	// annotation is a resource namespace; a namespace lacking it is
	// governance-only and out of scope for placement preconditions).
	Namespaces []string
}

const defaultRoleAccountMapName = "ack-role-account-map"

// ownerAccountIDAnnotation is the ACK-native namespace annotation ADR-015
// §5.8.1 names as the account source. It is intentionally not read from the
// util package: that package's constants are kropath-owned runtime contracts,
// and this key is ACK's, not kropath's.
const ownerAccountIDAnnotation = "services.k8s.aws/owner-account-id"

// TeamIDAnnotation is ACK's team-level CARM marker. Its presence on a
// kropath-managed namespace is unsupported (ADR-015 §5.8.2 rule 5): under
// ACK's TeamLevelCARM gate the team branch resolves the role from the team
// map, and the owner-account-id branch this precondition depends on never
// executes.
const TeamIDAnnotation = "services.k8s.aws/team-id"

var accountIDPattern = regexp.MustCompile(`^[0-9]{12}$`)

// roleARNPattern captures the account ID embedded in an IAM role ARN, e.g.
// "arn:aws:iam::111122223333:role/ACKRole" -> "111122223333". It accepts the
// aws-cn/aws-us-gov partitions too.
var roleARNPattern = regexp.MustCompile(`^arn:aws[a-zA-Z0-9-]*:iam::([0-9]{12}):role/`)

// Run inspects the cluster reachable through c and returns every precondition
// finding. It returns an error only when it cannot talk to the API server at
// all (e.g. listing Namespaces fails) — a missing or misconfigured
// precondition is always a Finding, never an error.
func Run(ctx context.Context, c client.Client, opts Options) (*Report, error) {
	namespaces, err := managedNamespaces(ctx, c, opts.Namespaces)
	if err != nil {
		return nil, fmt.Errorf("discover kropath-managed namespaces: %w", err)
	}

	report := &Report{}
	report.Findings = append(report.Findings, checkEnableCARM(ctx, c, opts.ACKNamespace)...)
	report.Findings = append(report.Findings, checkWatchNamespaceScope(ctx, c, opts.ACKNamespace, namespaces)...)
	report.Findings = append(report.Findings, checkNamespaceIgnoreList(namespaces)...)
	report.Findings = append(report.Findings, checkRoleAccountMap(ctx, c, opts, namespaces)...)
	report.Findings = append(report.Findings, checkTeamID(namespaces)...)
	report.Findings = append(report.Findings, checkIAMRoleSelector(ctx, c, namespaces)...)

	sort.SliceStable(report.Findings, func(i, j int) bool {
		if report.Findings[i].Check != report.Findings[j].Check {
			return report.Findings[i].Check < report.Findings[j].Check
		}
		return report.Findings[i].Namespace < report.Findings[j].Namespace
	})
	return report, nil
}

// managedNamespaces resolves the set of kropath-managed resource namespaces
// to check, either from an explicit override or by auto-discovery.
func managedNamespaces(ctx context.Context, c client.Client, override []string) ([]corev1.Namespace, error) {
	if len(override) > 0 {
		result := make([]corev1.Namespace, 0, len(override))
		for _, name := range override {
			var ns corev1.Namespace
			if err := c.Get(ctx, client.ObjectKey{Name: name}, &ns); err != nil {
				return nil, fmt.Errorf("get namespace %q: %w", name, err)
			}
			result = append(result, ns)
		}
		return result, nil
	}

	var list corev1.NamespaceList
	if err := c.List(ctx, &list); err != nil {
		return nil, err
	}
	var result []corev1.Namespace
	for _, ns := range list.Items {
		if ns.Annotations[util.GlobalConfigNamespaceAnnotation] != "" {
			result = append(result, ns)
		}
	}
	return result, nil
}
