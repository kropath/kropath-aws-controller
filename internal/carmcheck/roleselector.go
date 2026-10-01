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
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// iamRoleSelectorListGVK identifies ACK's cluster-scoped CARM-alternative
// CRD (group services.k8s.aws, not iam.services.k8s.aws — verified against
// aws-controllers-k8s/iam-controller's
// helm/crds/services.k8s.aws_iamroleselectors.yaml). kropath-aws-controller does
// not vendor ACK's IAM API types, so this is read as unstructured.
var iamRoleSelectorListGVK = schema.GroupVersionKind{
	Group:   "services.k8s.aws",
	Version: "v1alpha1",
	Kind:    "IAMRoleSelectorList",
}

// checkIAMRoleSelector implements ADR-015 §5.8.4 precondition 5
// (IAMRoleSelector half): no IAMRoleSelector may match a kropath-managed
// namespace. An IAMRoleSelector matching the namespace, or a resource in it,
// means ACK may resolve the role from the selector instead of from CARM's
// account map, and kropath cannot observe which one will apply.
func checkIAMRoleSelector(ctx context.Context, c client.Client, managed []corev1.Namespace) []Finding {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(iamRoleSelectorListGVK)

	if err := c.List(ctx, list); err != nil {
		if meta.IsNoMatchError(err) {
			return []Finding{{
				Check:    "iam-role-selector",
				Severity: SeverityOK,
				Message:  "IAMRoleSelector CRD not installed — ACK documents IAMRoleSelector and CARM as mutually exclusive, so this precondition holds by construction",
			}}
		}
		return []Finding{{
			Check:    "iam-role-selector",
			Severity: SeverityUnknown,
			Message:  fmt.Sprintf("could not list IAMRoleSelector objects: %v", err),
		}}
	}

	if len(list.Items) == 0 {
		return []Finding{{
			Check:    "iam-role-selector",
			Severity: SeverityOK,
			Message:  "no IAMRoleSelector objects defined",
		}}
	}

	var findings []Finding
	for _, item := range list.Items {
		for _, ns := range managed {
			if iamRoleSelectorMatchesNamespace(item, ns) {
				findings = append(findings, Finding{
					Check:     "iam-role-selector",
					Severity:  SeverityBlocking,
					Namespace: ns.Name,
					Message: fmt.Sprintf(
						"IAMRoleSelector %q matches namespace %q; kropath cannot verify whether ACK resolves the role via CARM's account map or this selector for resources here (ADR-015 §5.8.4 precondition 5)",
						item.GetName(), ns.Name,
					),
				})
			}
		}
	}
	if len(findings) == 0 {
		findings = append(findings, Finding{
			Check:    "iam-role-selector",
			Severity: SeverityOK,
			Message:  fmt.Sprintf("%d IAMRoleSelector object(s) defined but none match a kropath-managed namespace", len(list.Items)),
		})
	}
	return findings
}

// iamRoleSelectorMatchesNamespace reports whether item's namespaceSelector
// matches ns, by exact name or by a label subset match. It deliberately does
// not narrow by spec.resourceTypeSelector or spec.resourceLabelSelector:
// those can only make a selector match *fewer* resources within a matched
// namespace, never zero out a namespace match, and this checker is
// conservative by design — a false "matches" costs an operator a minute of
// verification, a false "does not match" hides the bug ADR-019 exists to
// eliminate.
func iamRoleSelectorMatchesNamespace(item unstructured.Unstructured, ns corev1.Namespace) bool {
	names, _, _ := unstructured.NestedStringSlice(item.Object, "spec", "namespaceSelector", "names")
	for _, n := range names {
		if n == ns.Name {
			return true
		}
	}

	matchLabels, _, _ := unstructured.NestedStringMap(item.Object, "spec", "namespaceSelector", "labelSelector", "matchLabels")
	if len(matchLabels) == 0 {
		return false
	}
	for k, v := range matchLabels {
		if ns.Labels[k] != v {
			return false
		}
	}
	return true
}
