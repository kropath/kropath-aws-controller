// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/kropath/kropath-aws-controller/api/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newPolicyDocumentScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}
	return scheme
}

func policyDocumentWithReadyReason(name, reason string) *v1alpha1.PolicyDocument {
	return &v1alpha1.PolicyDocument{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
		Status: v1alpha1.PolicyDocumentStatus{
			Conditions: []metav1.Condition{
				{Type: v1alpha1.ConditionReady, Status: metav1.ConditionTrue, Reason: reason, Message: "test fixture"},
			},
		},
	}
}

// compareCollectorFixture asserts collector's exposition for metricName
// matches the .prom fixture at fixturePath (repo standard: testutil.CollectAndCompare
// against internal/registry/metrics_test.go and internal/version/metrics_test.go's precedent).
func compareCollectorFixture(t *testing.T, collector prometheus.Collector, fixturePath, metricName string) {
	t.Helper()
	f, err := os.Open(fixturePath)
	if err != nil {
		t.Fatalf("open fixture %s: %v", fixturePath, err)
	}
	defer func() { _ = f.Close() }()

	if err := testutil.CollectAndCompare(collector, f, metricName); err != nil {
		t.Errorf("unexpected collector output for %s: %v", metricName, err)
	}
}

// failingReader implements client.Reader and fails every List and Get call,
// simulating a collector's backing kind being unreachable at scrape time.
type failingReader struct{}

func (failingReader) Get(_ context.Context, _ client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
	return errors.New("simulated get failure")
}

func (failingReader) List(_ context.Context, _ client.ObjectList, _ ...client.ListOption) error {
	return errors.New("simulated list failure")
}

func TestDocumentsCollector_ThreeReasons(t *testing.T) {
	scheme := newPolicyDocumentScheme(t)
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
		policyDocumentWithReadyReason("doc-a", "DocumentResolved"),
		policyDocumentWithReadyReason("doc-b", "SourceNotReady"),
		policyDocumentWithReadyReason("doc-c", "SidConflict"),
	).Build()

	collector := newDocumentsCollector(c, nil)
	compareCollectorFixture(t, collector, "testdata/documents_three_reasons.prom", "kropath_policydocument_documents")
}

func TestDocumentsCollector_AfterDelete(t *testing.T) {
	scheme := newPolicyDocumentScheme(t)
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
		policyDocumentWithReadyReason("doc-b", "SourceNotReady"),
		policyDocumentWithReadyReason("doc-c", "SidConflict"),
	).Build()

	collector := newDocumentsCollector(c, nil)
	compareCollectorFixture(t, collector, "testdata/documents_after_delete.prom", "kropath_policydocument_documents")
}

func TestDocumentsCollector_ListError(t *testing.T) {
	before := readCounterValue(t, metricsCollectErrorsTotal, prometheus.Labels{"collector": documentsCollectorName})
	collector := newDocumentsCollector(failingReader{}, nil)

	if n := testutil.CollectAndCount(collector, "kropath_policydocument_documents"); n != 0 {
		t.Errorf("expected no samples on a List error, got %d", n)
	}
	if delta := readCounterValue(t, metricsCollectErrorsTotal, prometheus.Labels{"collector": documentsCollectorName}) - before; delta != 1 {
		t.Errorf("kropath_metrics_collect_errors_total delta: want 1, got %v", delta)
	}
}
