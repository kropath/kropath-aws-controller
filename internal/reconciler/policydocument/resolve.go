// Copyright 2026 The kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package policydocument

import (
	"context"
	"fmt"
	"strings"

	"github.com/kropath/kropath-controller/api/v1alpha1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
		if apimeta.IsNoMatchError(err) || client.IgnoreNotFound(err) == nil {
			return "", true, nil
		}
		return "", false, err
	}

	switch field {
	case "predictedArn":
		arn, found, err := unstructured.NestedString(obj.Object, "status", "predictedArn")
		if err != nil {
			return "", false, err
		}
		if !found || strings.TrimSpace(arn) == "" {
			return "", true, nil
		}
		return arn, false, nil
	case "arn":
		arn, found, err := unstructured.NestedString(obj.Object, "status", "arn")
		if err != nil {
			return "", false, err
		}
		if !found || strings.TrimSpace(arn) == "" {
			return "", true, nil
		}
		return arn, false, nil
	default:
		return "", true, fmt.Errorf("unsupported ref field %q", field)
	}
}
