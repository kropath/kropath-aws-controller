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

package carmcheck

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// checkRoleAccountMap implements ADR-015 §5.8.4 preconditions 3 and 4: an
// ack-role-account-map entry must exist for every kropath-managed namespace's
// annotated account, and that entry's role ARN must live in the account it
// is keyed by. Precondition 4 is the most likely vector of the five — unlike
// the others it is a data-entry error made on every account onboarding, not
// an install flag set once by an expert.
func checkRoleAccountMap(ctx context.Context, c client.Client, opts Options, managed []corev1.Namespace) []Finding {
	name := opts.RoleAccountMapName
	if name == "" {
		name = defaultRoleAccountMapName
	}

	var cm corev1.ConfigMap
	if err := c.Get(ctx, client.ObjectKey{Namespace: opts.ACKNamespace, Name: name}, &cm); err != nil {
		return []Finding{{
			Check:    "role-account-map",
			Severity: SeverityUnknown,
			Message:  fmt.Sprintf("could not read ConfigMap %s/%s: %v", opts.ACKNamespace, name, err),
		}}
	}

	var findings []Finding
	// accountHasEntry records which account keys resolved to a role ARN
	// that lives in that same account — the only entries precondition 3
	// should count as "covering" that account.
	accountHasEntry := make(map[string]bool)

	for account, arn := range cm.Data {
		if !accountIDPattern.MatchString(account) {
			findings = append(findings, Finding{
				Check:    "role-account-map",
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("%s/%s: key %q is not a 12-digit AWS account ID", opts.ACKNamespace, name, account),
			})
			continue
		}

		m := roleARNPattern.FindStringSubmatch(arn)
		if m == nil {
			findings = append(findings, Finding{
				Check:    "role-account-map",
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("%s/%s: entry %q has a value that is not an IAM role ARN: %q", opts.ACKNamespace, name, account, arn),
			})
			continue
		}

		roleAccount := m[1]
		if roleAccount != account {
			findings = append(findings, Finding{
				Check:    "role-account-map",
				Severity: SeverityBlocking,
				Message: fmt.Sprintf(
					"%s/%s: entry %q maps to %s, a role in account %s — ACK overwrites the account with the resolved role's account, so it will place resources for account %s in %s instead (ADR-015 §5.8.4 precondition 4)",
					opts.ACKNamespace, name, account, arn, roleAccount, account, roleAccount,
				),
			})
			continue
		}

		accountHasEntry[account] = true
		findings = append(findings, Finding{
			Check:    "role-account-map",
			Severity: SeverityOK,
			Message:  fmt.Sprintf("%s/%s: entry %q correctly maps to a role in the same account", opts.ACKNamespace, name, account),
		})
	}

	for _, ns := range managed {
		account := ns.Annotations[ownerAccountIDAnnotation]
		if account == "" {
			// Annotation presence/validity is ADR-015 §5.8.2 rule 1, a
			// runtime concern for kropath-aws-controller's own reconciler
			// (KRO-1128), not this install-conformance checker.
			continue
		}
		if !accountHasEntry[account] {
			findings = append(findings, Finding{
				Check:     "role-account-map",
				Severity:  SeverityBlocking,
				Namespace: ns.Name,
				Message: fmt.Sprintf(
					"namespace %q declares owner-account-id=%s but %s/%s has no valid entry for that account (ADR-015 §5.8.4 precondition 3)",
					ns.Name, account, opts.ACKNamespace, name,
				),
			})
		}
	}

	return findings
}
