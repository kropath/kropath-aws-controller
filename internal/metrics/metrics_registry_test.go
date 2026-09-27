// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"math"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// gaugeVecHasSeries reports whether vec currently has a child with exactly
// labels, without creating one as a side effect (unlike GetMetricWith /
// WithLabelValues). Used to assert the M-17 delete-on-activation contract:
// an absent series, not a Set(0), is what "not pending" must look like.
func gaugeVecHasSeries(vec *prometheus.GaugeVec, labels prometheus.Labels) bool {
	ch := make(chan prometheus.Metric, 16)
	go func() {
		vec.Collect(ch)
		close(ch)
	}()
	for m := range ch {
		var dtoM dto.Metric
		if err := m.Write(&dtoM); err != nil {
			continue
		}
		if len(dtoM.GetLabel()) != len(labels) {
			continue
		}
		match := true
		for _, lp := range dtoM.GetLabel() {
			if v, ok := labels[lp.GetName()]; !ok || v != lp.GetValue() {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func TestReconcilerPendingSince_RecordsCurrentTimestamp(t *testing.T) {
	pkg := "mtest-pending-since"
	now := time.Now()
	ReconcilerPendingSince(pkg, now)
	defer ReconcilerPendingClear(pkg)

	if !gaugeVecHasSeries(registryReconcilerPendingSinceTimestampSeconds, prometheus.Labels{"package": pkg}) {
		t.Fatal("expected a pending-since series after ReconcilerPendingSince")
	}
	got := readGaugeValue(t, registryReconcilerPendingSinceTimestampSeconds, prometheus.Labels{"package": pkg})
	if math.Abs(got-float64(now.Unix())) > 5 {
		t.Errorf("pending-since value = %v, want within 5s of %v", got, now.Unix())
	}
}

// TestReconcilerPendingClear_DeletesSeries is the property M-17 turns on: a
// leftover child after activation is not a wrong number, it is a
// permanently-firing alert once §9's min-by aggregation picks it up.
func TestReconcilerPendingClear_DeletesSeries(t *testing.T) {
	pkg := "mtest-pending-clear"
	ReconcilerPendingSince(pkg, time.Now())
	if !gaugeVecHasSeries(registryReconcilerPendingSinceTimestampSeconds, prometheus.Labels{"package": pkg}) {
		t.Fatal("setup: expected series to exist before clearing")
	}

	ReconcilerPendingClear(pkg)

	if gaugeVecHasSeries(registryReconcilerPendingSinceTimestampSeconds, prometheus.Labels{"package": pkg}) {
		t.Error("expected series to be absent after ReconcilerPendingClear, not merely zeroed")
	}
}

func TestCRDWatchEvent(t *testing.T) {
	for _, outcome := range []string{"activated", "store_miss", "not_servable", "cast_failed", "already_active"} {
		before := readCounterValue(t, registryCRDWatchEventsTotal, prometheus.Labels{"outcome": outcome})
		CRDWatchEvent(outcome)
		if delta := readCounterValue(t, registryCRDWatchEventsTotal, prometheus.Labels{"outcome": outcome}) - before; delta != 1 {
			t.Errorf("registryCRDWatchEventsTotal[%s] delta: want 1, got %v", outcome, delta)
		}
	}
}

func TestOptionalKindsAttached(t *testing.T) {
	pkg := "mtest-optional-attached"
	OptionalKindsAttached(pkg, 3)
	if got := readGaugeValue(t, registryOptionalKindsAttached, prometheus.Labels{"package": pkg}); got != 3 {
		t.Errorf("registryOptionalKindsAttached = %v, want 3", got)
	}
	OptionalKindsAttached(pkg, 0)
	if got := readGaugeValue(t, registryOptionalKindsAttached, prometheus.Labels{"package": pkg}); got != 0 {
		t.Errorf("registryOptionalKindsAttached = %v, want 0 after re-set", got)
	}
}
