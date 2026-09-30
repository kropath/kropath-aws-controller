// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"testing"

	"github.com/kropath/kropath-aws-controller/internal/features"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func gatherFamily(t *testing.T, reg *prometheus.Registry, name string) *dto.MetricFamily {
	t.Helper()
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	for _, mf := range mfs {
		if mf.GetName() == name {
			return mf
		}
	}
	return nil
}

func TestBuildInfoLabelCardinality(t *testing.T) {
	reg := prometheus.NewRegistry()
	registerMetrics(reg, "v1.2.3", "abc1234", "go1.24.0", nil)

	mf := gatherFamily(t, reg, "kropath_build_info")
	if mf == nil {
		t.Fatal("kropath_build_info metric family not found")
	}
	if len(mf.GetMetric()) != 1 {
		t.Fatalf("want 1 time series, got %d", len(mf.GetMetric()))
	}
	m := mf.GetMetric()[0]
	if m.GetGauge().GetValue() != 1 {
		t.Errorf("want gauge value 1, got %f", m.GetGauge().GetValue())
	}

	labels := labelsMap(m)
	if labels["version"] != "v1.2.3" {
		t.Errorf("version label: got %q, want %q", labels["version"], "v1.2.3")
	}
	if labels["git_commit"] != "abc1234" {
		t.Errorf("git_commit label: got %q, want %q", labels["git_commit"], "abc1234")
	}
	if labels["go_version"] != "go1.24.0" {
		t.Errorf("go_version label: got %q, want %q", labels["go_version"], "go1.24.0")
	}
}

func TestFeatureEnabledLabelCardinality(t *testing.T) {
	testFeatures := []features.Reconciler{
		{Package: "iamconfig", Stability: "stable"},
		{Package: "s3config", Stability: "stable"},
		{Package: "kmsconfig", Stability: "alpha"},
	}
	reg := prometheus.NewRegistry()
	registerMetrics(reg, "dev", "none", "go1.24.0", testFeatures)

	mf := gatherFamily(t, reg, "kropath_feature_enabled")
	if mf == nil {
		t.Fatal("kropath_feature_enabled metric family not found")
	}
	if len(mf.GetMetric()) != len(testFeatures) {
		t.Fatalf("want %d time series, got %d", len(testFeatures), len(mf.GetMetric()))
	}

	seen := make(map[string]string) // package → stability
	for _, m := range mf.GetMetric() {
		if m.GetGauge().GetValue() != 1 {
			t.Errorf("want gauge value 1, got %f", m.GetGauge().GetValue())
		}
		lm := labelsMap(m)
		seen[lm["feature"]] = lm["stability"]
	}
	for _, f := range testFeatures {
		got, ok := seen[f.Package]
		if !ok {
			t.Errorf("feature %q not found in metrics", f.Package)
			continue
		}
		if got != f.Stability {
			t.Errorf("feature %q: stability got %q, want %q", f.Package, got, f.Stability)
		}
	}
}

func TestFeatureEnabledFullRegistryCardinality(t *testing.T) {
	reg := prometheus.NewRegistry()
	registerMetrics(reg, "dev", "none", "go1.24.0", features.All)

	mf := gatherFamily(t, reg, "kropath_feature_enabled")
	if mf == nil {
		t.Fatal("kropath_feature_enabled metric family not found")
	}
	if len(mf.GetMetric()) != len(features.All) {
		t.Fatalf("want %d time series (one per features.All entry), got %d",
			len(features.All), len(mf.GetMetric()))
	}
}

func labelsMap(m *dto.Metric) map[string]string {
	out := make(map[string]string, len(m.GetLabel()))
	for _, lp := range m.GetLabel() {
		out[lp.GetName()] = lp.GetValue()
	}
	return out
}
