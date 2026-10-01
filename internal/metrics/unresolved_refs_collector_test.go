// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"context"
	"fmt"
	"testing"

	"github.com/kropath/kropath-aws-controller/api/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// alwaysNotFoundReader lists through to the wrapped lister but reports every
// Get as NotFound, simulating a ref whose target object does not exist --
// the same "pending" outcome resolveRef treats a NotFound Get as.
type alwaysNotFoundReader struct {
	lister client.Reader
}

func (r alwaysNotFoundReader) Get(_ context.Context, key client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
	return apierrors.NewNotFound(schema.GroupResource{Group: "aws.kropath.run"}, key.Name)
}

func (r alwaysNotFoundReader) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return r.lister.List(ctx, list, opts...)
}

func resourceRefStatement(refs ...*v1alpha1.PolicyRef) v1alpha1.PolicyStatement {
	resources := make([]v1alpha1.PolicyResource, len(refs))
	for i, ref := range refs {
		resources[i] = v1alpha1.PolicyResource{Ref: ref}
	}
	return v1alpha1.PolicyStatement{Effect: "Allow", Actions: []string{"s3:GetObject"}, Resources: resources}
}

func TestUnresolvedRefsCollector_ThreePending(t *testing.T) {
	scheme := newPolicyDocumentScheme(t)
	doc := &v1alpha1.PolicyDocument{
		ObjectMeta: metav1.ObjectMeta{Name: "ac09-doc", Namespace: "default"},
		Spec: v1alpha1.PolicyDocumentSpec{
			Statements: []v1alpha1.PolicyStatement{
				resourceRefStatement(
					&v1alpha1.PolicyRef{Kind: "AWSIAMRole", Name: "role-a"},
					&v1alpha1.PolicyRef{Kind: "AWSS3Bucket", Name: "bucket-a"},
					&v1alpha1.PolicyRef{Kind: "AWSLambdaFunction", Name: "fn-a"},
				),
			},
		},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(doc).Build()
	reader := alwaysNotFoundReader{lister: fakeClient}

	collector := newUnresolvedRefsCollector(reader, nil)
	compareCollectorFixture(t, collector, "testdata/unresolved_refs_three_pending.prom", "kropath_policydocument_unresolved_refs")
}

func TestUnresolvedRefsCollector_KindOther(t *testing.T) {
	scheme := newPolicyDocumentScheme(t)
	doc := &v1alpha1.PolicyDocument{
		ObjectMeta: metav1.ObjectMeta{Name: "ac11-doc", Namespace: "default"},
		Spec: v1alpha1.PolicyDocumentSpec{
			Statements: []v1alpha1.PolicyStatement{
				resourceRefStatement(&v1alpha1.PolicyRef{Kind: "NotAKropathKind", Name: "whatever"}),
			},
		},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(doc).Build()
	reader := alwaysNotFoundReader{lister: fakeClient}

	collector := newUnresolvedRefsCollector(reader, nil)
	compareCollectorFixture(t, collector, "testdata/unresolved_refs_kind_other.prom", "kropath_policydocument_unresolved_refs")
}

// TestUnresolvedRefsCollector_TenBogusKindsBucketToOne asserts the §2.4
// bucketing bound: ten distinct unbounded ref.Kind values still cost exactly
// one "other" series, not one per typo.
func TestUnresolvedRefsCollector_TenBogusKindsBucketToOne(t *testing.T) {
	scheme := newPolicyDocumentScheme(t)
	refs := make([]*v1alpha1.PolicyRef, 10)
	for i := range refs {
		refs[i] = &v1alpha1.PolicyRef{Kind: fmt.Sprintf("BogusKind%d", i), Name: fmt.Sprintf("obj-%d", i)}
	}
	doc := &v1alpha1.PolicyDocument{
		ObjectMeta: metav1.ObjectMeta{Name: "ac11-ten-bogus", Namespace: "default"},
		Spec:       v1alpha1.PolicyDocumentSpec{Statements: []v1alpha1.PolicyStatement{resourceRefStatement(refs...)}},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(doc).Build()
	reader := alwaysNotFoundReader{lister: fakeClient}

	collector := newUnresolvedRefsCollector(reader, nil)
	compareCollectorFixture(t, collector, "testdata/unresolved_refs_ten_bogus_kinds.prom", "kropath_policydocument_unresolved_refs")
}

func TestUnresolvedRefsCollector_ListError(t *testing.T) {
	before := readCounterValue(t, metricsCollectErrorsTotal, prometheus.Labels{"collector": unresolvedRefsCollectorName})
	collector := newUnresolvedRefsCollector(failingReader{}, nil)

	if n := testutil.CollectAndCount(collector, "kropath_policydocument_unresolved_refs"); n != 0 {
		t.Errorf("expected no samples on a List error, got %d", n)
	}
	if delta := readCounterValue(t, metricsCollectErrorsTotal, prometheus.Labels{"collector": unresolvedRefsCollectorName}) - before; delta != 1 {
		t.Errorf("kropath_metrics_collect_errors_total delta: want 1, got %v", delta)
	}
}
