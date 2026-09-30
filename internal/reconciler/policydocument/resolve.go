// Copyright 2026 The kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package policydocument

import (
	"context"
	"fmt"
	"strings"

	"github.com/kropath/kropath-aws-controller/api/v1alpha1"
	"github.com/kropath/kropath-aws-controller/internal/metrics"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// bucketKind buckets ref.Kind against the six installable kinds
// policydocument watches (policyDocumentRefKindSet, built from the same list
// as defaultRefGVKs, and mirrored by internal/registry/entries.go:85's
// policyDocumentRefGVKs), applied here at the emit site rather than in the
// metrics package, so a tenant's unbounded ref.kind value never reaches a
// label. A kind outside the set costs one "other" series, not one series per
// typo (spec §2.4).
func bucketKind(kind string) string {
	if policyDocumentRefKindSet[kind] {
		return kind
	}
	return "other"
}

func resolveRef(ctx context.Context, c client.Client, namespace string, ref *v1alpha1.PolicyRef) (string, bool, error) {
	if ref == nil {
		return "", true, fmt.Errorf("ref must not be nil")
	}
	if strings.TrimSpace(ref.Kind) == "" || strings.TrimSpace(ref.Name) == "" {
		return "", true, fmt.Errorf("ref.kind and ref.name are required")
	}

	field := strings.TrimSpace(ref.Field)
	if field == "" {
		field = "predictedArn"
	}

	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "aws.kropath.run", Version: "v1alpha1", Kind: ref.Kind})
	if err := c.Get(ctx, types.NamespacedName{Namespace: namespace, Name: ref.Name}, obj); err != nil {
		// The ref's CRD may not be registered yet (an optional kind not yet
		// installed, e.g. AWSLambdaFunction before AC-9 installs its CRD).
		// Treat that the same as "object not found" — pending, not an error —
		// so a dynamic-detection suite doesn't surface as a Reconciler error.
		// Split into two sequential ifs (rather than the prior ||) so the
		// distinguishing err is still in scope to tell the two outcomes apart
		// for the counter below (spec §2.4) — both arms still return the
		// identical ("", true, nil) tuple, so this is a readability change
		// only and M-11's swallow is untouched.
		if apimeta.IsNoMatchError(err) {
			metrics.PolicyDocumentRefResolution(bucketKind(ref.Kind), field, "crd_absent")
			return "", true, nil
		}
		if client.IgnoreNotFound(err) == nil {
			metrics.PolicyDocumentRefResolution(bucketKind(ref.Kind), field, "pending")
			return "", true, nil
		}
		metrics.PolicyDocumentRefResolution(bucketKind(ref.Kind), field, "error")
		return "", false, err
	}

	switch field {
	case "predictedArn":
		arn, found, err := unstructured.NestedString(obj.Object, "status", "predictedArn")
		if err != nil {
			metrics.PolicyDocumentRefResolution(bucketKind(ref.Kind), field, "error")
			return "", false, err
		}
		if !found || strings.TrimSpace(arn) == "" {
			metrics.PolicyDocumentRefResolution(bucketKind(ref.Kind), field, "pending")
			return "", true, nil
		}
		metrics.PolicyDocumentRefResolution(bucketKind(ref.Kind), field, "resolved")
		return arn, false, nil
	case "arn":
		arn, found, err := unstructured.NestedString(obj.Object, "status", "arn")
		if err != nil {
			metrics.PolicyDocumentRefResolution(bucketKind(ref.Kind), field, "error")
			return "", false, err
		}
		if !found || strings.TrimSpace(arn) == "" {
			metrics.PolicyDocumentRefResolution(bucketKind(ref.Kind), field, "pending")
			return "", true, nil
		}
		metrics.PolicyDocumentRefResolution(bucketKind(ref.Kind), field, "resolved")
		return arn, false, nil
	default:
		metrics.PolicyDocumentRefResolution(bucketKind(ref.Kind), "unsupported", "error")
		return "", true, fmt.Errorf("unsupported ref field %q", field)
	}
}
