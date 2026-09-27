// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func readGaugeValue(t *testing.T, vec *prometheus.GaugeVec, labels prometheus.Labels) float64 {
	t.Helper()
	g, err := vec.GetMetricWith(labels)
	if err != nil {
		t.Fatalf("GetMetricWith: %v", err)
	}
	var m dto.Metric
	if err := g.Write(&m); err != nil {
		t.Fatalf("metric Write: %v", err)
	}
	return m.GetGauge().GetValue()
}

func TestLabelOperatorWatchedKinds(t *testing.T) {
	LabelOperatorWatchedKinds("mtest.kropath.run", 3)
	if got := readGaugeValue(t, labeloperatorWatchedKinds, prometheus.Labels{"group": "mtest.kropath.run"}); got != 3 {
		t.Errorf("labeloperatorWatchedKinds = %v, want 3", got)
	}
	LabelOperatorWatchedKinds("mtest.kropath.run", 0)
	if got := readGaugeValue(t, labeloperatorWatchedKinds, prometheus.Labels{"group": "mtest.kropath.run"}); got != 0 {
		t.Errorf("labeloperatorWatchedKinds = %v, want 0 after re-set", got)
	}
}

func TestLabelOperatorGroupDiscovery(t *testing.T) {
	for _, outcome := range []string{"discovered", "empty", "error"} {
		before := readCounterValue(t, labeloperatorGroupDiscoveryTotal, prometheus.Labels{"group": "mtest.kropath.run", "outcome": outcome})
		LabelOperatorGroupDiscovery("mtest.kropath.run", outcome)
		if delta := readCounterValue(t, labeloperatorGroupDiscoveryTotal, prometheus.Labels{"group": "mtest.kropath.run", "outcome": outcome}) - before; delta != 1 {
			t.Errorf("labeloperatorGroupDiscoveryTotal[%s] delta: want 1, got %v", outcome, delta)
		}
	}
}

func TestLabelOperatorPatch(t *testing.T) {
	for _, outcome := range []string{"patched", "not_found", "error"} {
		before := readCounterValue(t, labeloperatorPatchesTotal, prometheus.Labels{"group": "mtest.kropath.run", "outcome": outcome})
		LabelOperatorPatch("mtest.kropath.run", outcome)
		if delta := readCounterValue(t, labeloperatorPatchesTotal, prometheus.Labels{"group": "mtest.kropath.run", "outcome": outcome}) - before; delta != 1 {
			t.Errorf("labeloperatorPatchesTotal[%s] delta: want 1, got %v", outcome, delta)
		}
	}
}
