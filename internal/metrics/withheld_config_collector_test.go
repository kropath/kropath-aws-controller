// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/kropath/kropath-controller/api/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func withheldTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme: %v", err)
	}
	return scheme
}

func withheldCondition(reason string) []metav1.Condition {
	return []metav1.Condition{{Type: "Reconciled", Status: metav1.ConditionFalse, Reason: reason}}
}

// noMatchKindsReader wraps a real client.Reader and reports a
// *meta.NoKindMatchError for any List whose object list type name is in
// kinds, mirroring policydocument's noMatchKindClient fixture.
type noMatchKindsReader struct {
	client.Reader
	kinds map[string]bool
}

func (r noMatchKindsReader) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	typeName := reflect.TypeOf(list).Elem().Name()
	if r.kinds[typeName] {
		return &apimeta.NoKindMatchError{
			GroupKind:        schema.GroupKind{Group: "aws.kropath.run", Kind: strings.TrimSuffix(typeName, "List")},
			SearchedVersions: []string{"v1alpha1"},
		}
	}
	return r.Reader.List(ctx, list, opts...)
}

// erroringReader fails every List with a plain error -- a real collect
// failure, never meta.NoKindMatchError.
type erroringReader struct{ client.Reader }

func (erroringReader) List(_ context.Context, _ client.ObjectList, _ ...client.ListOption) error {
	return errors.New("simulated list failure")
}

func TestWithheldConfigCollector_TwoFamilies(t *testing.T) {
	scheme := withheldTestScheme(t)
	s3 := &v1alpha1.S3Config{
		ObjectMeta: metav1.ObjectMeta{Name: "general-policy", Namespace: "app"},
		Status:     v1alpha1.S3ConfigStatus{Conditions: withheldCondition("NamespaceUnreadable")},
	}
	iam := &v1alpha1.IAMConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "general-policy", Namespace: "gov"},
		Status:     v1alpha1.IAMConfigStatus{Conditions: withheldCondition("GlobalTierInput")},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(s3, iam).Build()

	collector := &withheldConfigCollector{apiReader: c}
	want := `
		# HELP kropath_cascade_effective_config_withheld Number of <Family>Config objects currently publishing no status.effectiveConfig, by family and withholding reason, sampled at scrape time.
		# TYPE kropath_cascade_effective_config_withheld gauge
		kropath_cascade_effective_config_withheld{family="iamconfig",reason="GlobalTierInput"} 1
		kropath_cascade_effective_config_withheld{family="s3config",reason="NamespaceUnreadable"} 1
	`
	if err := testutil.CollectAndCompare(collector, strings.NewReader(want), "kropath_cascade_effective_config_withheld"); err != nil {
		t.Fatal(err)
	}
}

func TestWithheldConfigCollector_ResolvedConfigNotCounted(t *testing.T) {
	scheme := withheldTestScheme(t)
	s3 := &v1alpha1.S3Config{
		ObjectMeta: metav1.ObjectMeta{Name: "general-policy", Namespace: "app"},
		Status: v1alpha1.S3ConfigStatus{
			Conditions:       []metav1.Condition{{Type: "Reconciled", Status: metav1.ConditionTrue, Reason: "CascadeMerged"}},
			EffectiveConfig:  v1alpha1.EffectiveS3Config{AWS: v1alpha1.ProviderIdentity{AccountID: "123456789012", Region: "us-east-1"}},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(s3).Build()

	collector := &withheldConfigCollector{apiReader: c}
	if err := testutil.CollectAndCompare(collector, strings.NewReader(""), "kropath_cascade_effective_config_withheld"); err != nil {
		t.Fatal(err)
	}
}

func TestWithheldConfigCollector_DeletedObjectRemovesSeries(t *testing.T) {
	scheme := withheldTestScheme(t)
	s3 := &v1alpha1.S3Config{
		ObjectMeta: metav1.ObjectMeta{Name: "general-policy", Namespace: "app"},
		Status:     v1alpha1.S3ConfigStatus{Conditions: withheldCondition("NamespaceUnreadable")},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(s3).Build()
	collector := &withheldConfigCollector{apiReader: c}

	present := `
		# HELP kropath_cascade_effective_config_withheld Number of <Family>Config objects currently publishing no status.effectiveConfig, by family and withholding reason, sampled at scrape time.
		# TYPE kropath_cascade_effective_config_withheld gauge
		kropath_cascade_effective_config_withheld{family="s3config",reason="NamespaceUnreadable"} 1
	`
	if err := testutil.CollectAndCompare(collector, strings.NewReader(present), "kropath_cascade_effective_config_withheld"); err != nil {
		t.Fatalf("before delete: %v", err)
	}

	if err := c.Delete(context.Background(), s3); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if err := testutil.CollectAndCompare(collector, strings.NewReader(""), "kropath_cascade_effective_config_withheld"); err != nil {
		t.Fatalf("after delete: %v", err)
	}
}

func TestWithheldConfigCollector_CRDAbsentIsNotACollectError(t *testing.T) {
	scheme := withheldTestScheme(t)
	s3 := &v1alpha1.S3Config{
		ObjectMeta: metav1.ObjectMeta{Name: "general-policy", Namespace: "app"},
		Status:     v1alpha1.S3ConfigStatus{Conditions: withheldCondition("NamespaceUnreadable")},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(s3).Build()
	reader := noMatchKindsReader{Reader: c, kinds: map[string]bool{"MemoryDBConfigList": true}}
	collector := &withheldConfigCollector{apiReader: reader}

	errBefore := readCounterValue(t, metricsCollectErrorsTotal, prometheus.Labels{"collector": withheldCollectorLabel})

	want := `
		# HELP kropath_cascade_effective_config_withheld Number of <Family>Config objects currently publishing no status.effectiveConfig, by family and withholding reason, sampled at scrape time.
		# TYPE kropath_cascade_effective_config_withheld gauge
		kropath_cascade_effective_config_withheld{family="s3config",reason="NamespaceUnreadable"} 1
	`
	if err := testutil.CollectAndCompare(collector, strings.NewReader(want), "kropath_cascade_effective_config_withheld"); err != nil {
		t.Fatal(err)
	}

	errAfter := readCounterValue(t, metricsCollectErrorsTotal, prometheus.Labels{"collector": withheldCollectorLabel})
	if delta := errAfter - errBefore; delta != 0 {
		t.Errorf("kropath_metrics_collect_errors_total delta: want 0 (NoKindMatchError is not a collect error), got %v", delta)
	}
}

func TestWithheldConfigCollector_ListErrorEmitsNoSamples(t *testing.T) {
	collector := &withheldConfigCollector{apiReader: erroringReader{}}

	errBefore := readCounterValue(t, metricsCollectErrorsTotal, prometheus.Labels{"collector": withheldCollectorLabel})

	if err := testutil.CollectAndCompare(collector, strings.NewReader(""), "kropath_cascade_effective_config_withheld"); err != nil {
		t.Fatal(err)
	}

	errAfter := readCounterValue(t, metricsCollectErrorsTotal, prometheus.Labels{"collector": withheldCollectorLabel})
	if delta := errAfter - errBefore; delta != 1 {
		t.Errorf("kropath_metrics_collect_errors_total delta: want 1, got %v", delta)
	}
}
