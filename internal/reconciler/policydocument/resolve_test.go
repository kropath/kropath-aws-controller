// Copyright 2026 The kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package policydocument

import (
	"context"
	"fmt"
	"testing"

	"github.com/kropath/kropath-aws-controller/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dto "github.com/prometheus/client_model/go"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// refResolutionCounterValue reads the current value of
// kropath_policydocument_ref_resolutions_total{kind,field,outcome} from the
// shared controller-runtime registry both packages register into. Tests
// compare before/after deltas rather than an absolute value, since the
// counter is process-global and cumulative across the whole test binary.
func refResolutionCounterValue(t *testing.T, kind, field, outcome string) float64 {
	t.Helper()
	families, err := ctrlmetrics.Registry.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}
	for _, family := range families {
		if family.GetName() != "kropath_policydocument_ref_resolutions_total" {
			continue
		}
		for _, m := range family.GetMetric() {
			if matchesLabels(m, map[string]string{"kind": kind, "field": field, "outcome": outcome}) {
				return m.GetCounter().GetValue()
			}
		}
	}
	return 0
}

func matchesLabels(m *dto.Metric, want map[string]string) bool {
	if len(m.GetLabel()) != len(want) {
		return false
	}
	for _, lp := range m.GetLabel() {
		if want[lp.GetName()] != lp.GetValue() {
			return false
		}
	}
	return true
}

// notFoundClient always reports NotFound on Get, simulating an installed
// kind whose referenced object does not exist yet -- resolveRef's "pending"
// outcome, distinct from noMatchKindClient's "crd_absent".
type notFoundClient struct {
	client.Client
}

func (notFoundClient) Get(_ context.Context, key client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
	return apierrors.NewNotFound(schema.GroupResource{Group: "aws.kropath.run"}, key.Name)
}

// erroringGetClient reports an arbitrary hard error on Get, distinct from
// NotFound and NoMatchError, for resolveRef's outcome="error" path.
type erroringGetClient struct {
	client.Client
}

func (erroringGetClient) Get(_ context.Context, _ client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
	return fmt.Errorf("simulated transient API error")
}

func TestResolveRefOutcomes(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	predictedArnObj := &unstructured.Unstructured{}
	predictedArnObj.SetGroupVersionKind(schema.GroupVersionKind{Group: "aws.kropath.run", Version: "v1alpha1", Kind: "AWSIAMRole"})
	predictedArnObj.SetNamespace("default")
	predictedArnObj.SetName("resolved-role")
	if err := unstructured.SetNestedField(predictedArnObj.Object, "arn:aws:iam::123456789012:role/example", "status", "predictedArn"); err != nil {
		t.Fatalf("seed predictedArn: %v", err)
	}

	pendingFieldObj := &unstructured.Unstructured{}
	pendingFieldObj.SetGroupVersionKind(schema.GroupVersionKind{Group: "aws.kropath.run", Version: "v1alpha1", Kind: "AWSIAMRole"})
	pendingFieldObj.SetNamespace("default")
	pendingFieldObj.SetName("pending-role")

	resolvedArnObj := &unstructured.Unstructured{}
	resolvedArnObj.SetGroupVersionKind(schema.GroupVersionKind{Group: "aws.kropath.run", Version: "v1alpha1", Kind: "AWSIAMRole"})
	resolvedArnObj.SetNamespace("default")
	resolvedArnObj.SetName("resolved-role-arn")
	if err := unstructured.SetNestedField(resolvedArnObj.Object, "arn:aws:iam::123456789012:role/resolved-via-arn", "status", "arn"); err != nil {
		t.Fatalf("seed arn: %v", err)
	}

	pendingArnObj := &unstructured.Unstructured{}
	pendingArnObj.SetGroupVersionKind(schema.GroupVersionKind{Group: "aws.kropath.run", Version: "v1alpha1", Kind: "AWSIAMRole"})
	pendingArnObj.SetNamespace("default")
	pendingArnObj.SetName("pending-role-arn")

	badArnTypeObj := &unstructured.Unstructured{}
	badArnTypeObj.SetGroupVersionKind(schema.GroupVersionKind{Group: "aws.kropath.run", Version: "v1alpha1", Kind: "AWSIAMRole"})
	badArnTypeObj.SetNamespace("default")
	badArnTypeObj.SetName("bad-arn-type-role")
	if err := unstructured.SetNestedStringMap(badArnTypeObj.Object, map[string]string{"nested": "not-a-string"}, "status", "arn"); err != nil {
		t.Fatalf("seed bad arn type: %v", err)
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
		predictedArnObj, pendingFieldObj, resolvedArnObj, pendingArnObj, badArnTypeObj,
	).Build()

	tests := []struct {
		name        string
		client      client.Client
		ref         *v1alpha1.PolicyRef
		wantPending bool
		wantErr     bool
		outcome     string
		kind        string
		field       string
	}{
		{
			name:        "crd absent",
			client:      noMatchKindClient{},
			ref:         &v1alpha1.PolicyRef{Kind: "AWSLambdaFunction", Name: "fn"},
			wantPending: true,
			outcome:     "crd_absent",
			kind:        "AWSLambdaFunction",
			field:       "predictedArn",
		},
		{
			name:        "object not found -> pending",
			client:      notFoundClient{},
			ref:         &v1alpha1.PolicyRef{Kind: "AWSIAMRole", Name: "missing-role"},
			wantPending: true,
			outcome:     "pending",
			kind:        "AWSIAMRole",
			field:       "predictedArn",
		},
		{
			name:    "hard Get error",
			client:  erroringGetClient{},
			ref:     &v1alpha1.PolicyRef{Kind: "AWSIAMRole", Name: "role"},
			wantErr: true,
			outcome: "error",
			kind:    "AWSIAMRole",
			field:   "predictedArn",
		},
		{
			name:        "field absent -> pending",
			client:      fakeClient,
			ref:         &v1alpha1.PolicyRef{Kind: "AWSIAMRole", Name: "pending-role"},
			wantPending: true,
			outcome:     "pending",
			kind:        "AWSIAMRole",
			field:       "predictedArn",
		},
		{
			name:    "field populated -> resolved",
			client:  fakeClient,
			ref:     &v1alpha1.PolicyRef{Kind: "AWSIAMRole", Name: "resolved-role"},
			outcome: "resolved",
			kind:    "AWSIAMRole",
			field:   "predictedArn",
		},
		{
			name:        "unsupported field",
			client:      fakeClient,
			ref:         &v1alpha1.PolicyRef{Kind: "AWSIAMRole", Name: "resolved-role", Field: "bogus"},
			wantPending: true,
			wantErr:     true,
			outcome:     "error",
			kind:        "AWSIAMRole",
			field:       "unsupported",
		},
		{
			name:        "unbounded kind buckets to other",
			client:      notFoundClient{},
			ref:         &v1alpha1.PolicyRef{Kind: "NotAKropathKind", Name: "whatever"},
			wantPending: true,
			outcome:     "pending",
			kind:        "other",
			field:       "predictedArn",
		},
		{
			name:        "arn field absent -> pending",
			client:      fakeClient,
			ref:         &v1alpha1.PolicyRef{Kind: "AWSIAMRole", Name: "pending-role-arn", Field: "arn"},
			wantPending: true,
			outcome:     "pending",
			kind:        "AWSIAMRole",
			field:       "arn",
		},
		{
			name:    "arn field populated -> resolved",
			client:  fakeClient,
			ref:     &v1alpha1.PolicyRef{Kind: "AWSIAMRole", Name: "resolved-role-arn", Field: "arn"},
			outcome: "resolved",
			kind:    "AWSIAMRole",
			field:   "arn",
		},
		{
			name:    "arn field wrong type -> error",
			client:  fakeClient,
			ref:     &v1alpha1.PolicyRef{Kind: "AWSIAMRole", Name: "bad-arn-type-role", Field: "arn"},
			wantErr: true,
			outcome: "error",
			kind:    "AWSIAMRole",
			field:   "arn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := refResolutionCounterValue(t, tt.kind, tt.field, tt.outcome)

			_, pending, err := resolveRef(context.Background(), tt.client, "default", tt.ref)

			if (err != nil) != tt.wantErr {
				t.Fatalf("resolveRef error = %v, wantErr %v", err, tt.wantErr)
			}
			if pending != tt.wantPending {
				t.Fatalf("resolveRef pending = %v, want %v", pending, tt.wantPending)
			}
			if delta := refResolutionCounterValue(t, tt.kind, tt.field, tt.outcome) - before; delta != 1 {
				t.Errorf("ref_resolutions_total{kind=%q,field=%q,outcome=%q} delta: want 1, got %v", tt.kind, tt.field, tt.outcome, delta)
			}
		})
	}
}

// TestResolveRefInputValidationEmitsNoMetric covers §2.4.1's two rows that
// must emit nothing: both fire before a kind or field is known, so the only
// labels available would be meaningless (kind="other", field="unsupported").
func TestResolveRefInputValidationEmitsNoMetric(t *testing.T) {
	before := refResolutionCounterValue(t, "other", "unsupported", "error")

	if _, _, err := resolveRef(context.Background(), fake.NewClientBuilder().Build(), "default", nil); err == nil {
		t.Fatal("expected error for nil ref")
	}
	if _, _, err := resolveRef(context.Background(), fake.NewClientBuilder().Build(), "default", &v1alpha1.PolicyRef{}); err == nil {
		t.Fatal("expected error for blank kind/name")
	}

	if delta := refResolutionCounterValue(t, "other", "unsupported", "error") - before; delta != 0 {
		t.Errorf("expected no metric emission for input-validation failures, got delta %v", delta)
	}
}

func TestBucketKindTenBogusValuesMapToOther(t *testing.T) {
	for i := 0; i < 10; i++ {
		if got := bucketKind(fmt.Sprintf("BogusKind%d", i)); got != "other" {
			t.Errorf("bucketKind(BogusKind%d) = %q, want %q", i, got, "other")
		}
	}
}

func TestBucketKindKnownKindsPassThrough(t *testing.T) {
	for _, k := range policyDocumentRefKindNames {
		if got := bucketKind(k); got != k {
			t.Errorf("bucketKind(%q) = %q, want unchanged", k, got)
		}
	}
}
