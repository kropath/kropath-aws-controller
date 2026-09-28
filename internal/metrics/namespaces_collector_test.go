// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"context"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func namespacesTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme: %v", err)
	}
	return scheme
}

func resourceNamespace(name, placementStatus string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				globalConfigNamespaceAnnotation: name,
				placementStatusAnnotation:       placementStatus,
			},
		},
	}
}

func TestNamespacesCollector_OkAndMissingRegion(t *testing.T) {
	scheme := namespacesTestScheme(t)
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
		resourceNamespace("ok-ns", "ok"),
		resourceNamespace("missing-region-ns", "MissingRegionAnnotation"),
		// Governance-only: no globalConfigNamespaceAnnotation, so it is never
		// evaluated by namespaceplacement and must not appear here even
		// though it happens to carry a stray placement-status annotation.
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
			Name:        "governance-only-ns",
			Annotations: map[string]string{placementStatusAnnotation: "ok"},
		}},
	).Build()

	collector := &namespacesCollector{reader: c}
	want := `
		# HELP kropath_namespaceplacement_namespaces Number of resource namespaces currently carrying each placement-status annotation value, sampled at scrape time.
		# TYPE kropath_namespaceplacement_namespaces gauge
		kropath_namespaceplacement_namespaces{status="MissingRegionAnnotation"} 1
		kropath_namespaceplacement_namespaces{status="ok"} 1
	`
	if err := testutil.CollectAndCompare(collector, strings.NewReader(want), "kropath_namespaceplacement_namespaces"); err != nil {
		t.Fatal(err)
	}
}

func TestNamespacesCollector_DeletedObjectRemovesSeries(t *testing.T) {
	scheme := namespacesTestScheme(t)
	ns := resourceNamespace("ok-ns", "ok")
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ns).Build()
	collector := &namespacesCollector{reader: c}

	present := `
		# HELP kropath_namespaceplacement_namespaces Number of resource namespaces currently carrying each placement-status annotation value, sampled at scrape time.
		# TYPE kropath_namespaceplacement_namespaces gauge
		kropath_namespaceplacement_namespaces{status="ok"} 1
	`
	if err := testutil.CollectAndCompare(collector, strings.NewReader(present), "kropath_namespaceplacement_namespaces"); err != nil {
		t.Fatalf("before delete: %v", err)
	}

	if err := c.Delete(context.Background(), ns); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if err := testutil.CollectAndCompare(collector, strings.NewReader(""), "kropath_namespaceplacement_namespaces"); err != nil {
		t.Fatalf("after delete: %v", err)
	}
}

func TestNamespacesCollector_ListErrorEmitsNoSamples(t *testing.T) {
	collector := &namespacesCollector{reader: erroringReader{}}

	errBefore := readCounterValue(t, metricsCollectErrorsTotal, prometheus.Labels{"collector": namespacesCollectorLabel})

	if err := testutil.CollectAndCompare(collector, strings.NewReader(""), "kropath_namespaceplacement_namespaces"); err != nil {
		t.Fatal(err)
	}

	errAfter := readCounterValue(t, metricsCollectErrorsTotal, prometheus.Labels{"collector": namespacesCollectorLabel})
	if delta := errAfter - errBefore; delta != 1 {
		t.Errorf("kropath_metrics_collect_errors_total delta: want 1, got %v", delta)
	}
}
