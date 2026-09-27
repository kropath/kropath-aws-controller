// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func readCounterValue(t *testing.T, vec *prometheus.CounterVec, labels prometheus.Labels) float64 {
	t.Helper()
	c, err := vec.GetMetricWith(labels)
	if err != nil {
		t.Fatalf("GetMetricWith: %v", err)
	}
	var m dto.Metric
	if err := c.Write(&m); err != nil {
		t.Fatalf("metric Write: %v", err)
	}
	return m.GetCounter().GetValue()
}

func TestPlacementResolved(t *testing.T) {
	before := readCounterValue(t, placementResolutionsTotal, prometheus.Labels{"reason": "mtest-reason-a"})
	PlacementResolved("mtest-reason-a")
	if delta := readCounterValue(t, placementResolutionsTotal, prometheus.Labels{"reason": "mtest-reason-a"}) - before; delta != 1 {
		t.Errorf("placementResolutionsTotal delta: want 1, got %v", delta)
	}
}

func TestConfigProfileResolved(t *testing.T) {
	before := readCounterValue(t, configProfileResolutionsTotal, prometheus.Labels{"family": "mtestconfig", "reason": "ProfileFound"})
	ConfigProfileResolved("mtestconfig", "ProfileFound")
	if delta := readCounterValue(t, configProfileResolutionsTotal, prometheus.Labels{"family": "mtestconfig", "reason": "ProfileFound"}) - before; delta != 1 {
		t.Errorf("configProfileResolutionsTotal delta: want 1, got %v", delta)
	}
}
