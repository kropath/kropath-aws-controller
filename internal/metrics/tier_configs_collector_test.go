// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"context"
	"strings"
	"testing"

	"github.com/kropath/kropath-controller/api/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func kropathConfig(name, namespace, reason string) *v1alpha1.KropathConfig {
	return &v1alpha1.KropathConfig{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Status: v1alpha1.KropathConfigStatus{
			Conditions: []metav1.Condition{{Type: "Reconciled", Status: metav1.ConditionTrue, Reason: reason}},
		},
	}
}

// readPlainGaugeValue reads a bare prometheus.Gauge (not a GaugeVec) --
// distinct from metrics_labeloperator_test.go's readGaugeValue, which reads a
// *GaugeVec by label set.
func readPlainGaugeValue(t *testing.T, g prometheus.Gauge) float64 {
	t.Helper()
	var m dto.Metric
	if err := g.Write(&m); err != nil {
		t.Fatalf("Write: %v", err)
	}
	return m.GetGauge().GetValue()
}

// drainCollect runs Collect and discards every sample -- used when a test
// only cares about a side-effect gauge (KropathConfigFamilyKindsUnavailable),
// not the collector-emitted series themselves.
func drainCollect(c prometheus.Collector) {
	ch := make(chan prometheus.Metric, 64)
	c.Collect(ch)
	close(ch)
	for range ch {
	}
}

func TestTierConfigsCollector_FourReasons(t *testing.T) {
	scheme := withheldTestScheme(t)
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
		kropathConfig("global-and-local", "ns1", "GlobalAndLocalTier"),
		kropathConfig("global", "ns2", "GlobalTier"),
		kropathConfig("local", "ns3", "LocalTier"),
		kropathConfig("unref", "ns4", "Unreferenced"),
	).Build()

	collector := &tierConfigsCollector{apiReader: c}
	want := `
		# HELP kropath_kropathconfigstatus_configs Number of KropathConfig objects currently in each tier-reference reason, sampled at scrape time; reason="Unreferenced" means no <Family>Config resolves this config.
		# TYPE kropath_kropathconfigstatus_configs gauge
		kropath_kropathconfigstatus_configs{reason="GlobalAndLocalTier"} 1
		kropath_kropathconfigstatus_configs{reason="GlobalTier"} 1
		kropath_kropathconfigstatus_configs{reason="LocalTier"} 1
		kropath_kropathconfigstatus_configs{reason="Unreferenced"} 1
	`
	if err := testutil.CollectAndCompare(collector, strings.NewReader(want), "kropath_kropathconfigstatus_configs"); err != nil {
		t.Fatal(err)
	}
}

func TestTierConfigsCollector_DeletedObjectRemovesSeries(t *testing.T) {
	scheme := withheldTestScheme(t)
	kpc := kropathConfig("baseline", "ns1", "Unreferenced")
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(kpc).Build()
	collector := &tierConfigsCollector{apiReader: c}

	present := `
		# HELP kropath_kropathconfigstatus_configs Number of KropathConfig objects currently in each tier-reference reason, sampled at scrape time; reason="Unreferenced" means no <Family>Config resolves this config.
		# TYPE kropath_kropathconfigstatus_configs gauge
		kropath_kropathconfigstatus_configs{reason="Unreferenced"} 1
	`
	if err := testutil.CollectAndCompare(collector, strings.NewReader(present), "kropath_kropathconfigstatus_configs"); err != nil {
		t.Fatalf("before delete: %v", err)
	}

	if err := c.Delete(context.Background(), kpc); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if err := testutil.CollectAndCompare(collector, strings.NewReader(""), "kropath_kropathconfigstatus_configs"); err != nil {
		t.Fatalf("after delete: %v", err)
	}
}

func TestTierConfigsCollector_FamilyKindsUnavailableGauge(t *testing.T) {
	scheme := withheldTestScheme(t)
	kpc := kropathConfig("baseline", "ns1", "Unreferenced")
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(kpc).Build()
	reader := noMatchKindsReader{Reader: c, kinds: map[string]bool{
		"MemoryDBConfigList": true,
		"KMSConfigList":      true,
	}}
	collector := &tierConfigsCollector{apiReader: reader}

	drainCollect(collector)

	if got := readPlainGaugeValue(t, kropathconfigstatusFamilyKindsUnavailable); got != 2 {
		t.Errorf("kropath_kropathconfigstatus_family_kinds_unavailable: want 2, got %v", got)
	}

	errAfter := readCounterValue(t, metricsCollectErrorsTotal, prometheus.Labels{"collector": tierConfigsCollectorLabel})
	if errAfter != 0 {
		t.Errorf("kropath_metrics_collect_errors_total{collector=%q}: want 0, got %v", tierConfigsCollectorLabel, errAfter)
	}
}

func TestTierConfigsCollector_ListErrorEmitsNoSamples(t *testing.T) {
	collector := &tierConfigsCollector{apiReader: erroringReader{}}

	errBefore := readCounterValue(t, metricsCollectErrorsTotal, prometheus.Labels{"collector": tierConfigsCollectorLabel})

	if err := testutil.CollectAndCompare(collector, strings.NewReader(""), "kropath_kropathconfigstatus_configs"); err != nil {
		t.Fatal(err)
	}

	errAfter := readCounterValue(t, metricsCollectErrorsTotal, prometheus.Labels{"collector": tierConfigsCollectorLabel})
	if delta := errAfter - errBefore; delta != 1 {
		t.Errorf("kropath_metrics_collect_errors_total delta: want 1, got %v", delta)
	}
}
