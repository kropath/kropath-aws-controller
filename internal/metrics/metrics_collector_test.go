// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func TestCollectError(t *testing.T) {
	before := readCounterValue(t, metricsCollectErrorsTotal, prometheus.Labels{"collector": "mtest-collector"})
	CollectError("mtest-collector")
	if delta := readCounterValue(t, metricsCollectErrorsTotal, prometheus.Labels{"collector": "mtest-collector"}) - before; delta != 1 {
		t.Errorf("metricsCollectErrorsTotal delta: want 1, got %v", delta)
	}
}

func TestCollectContextDeadline(t *testing.T) {
	ctx, cancel := CollectContext(context.Background())
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("CollectContext: expected a deadline")
	}
	if got := time.Until(deadline); got <= 0 || got > CollectTimeout {
		t.Errorf("CollectContext: deadline %v from now, want in (0, %v]", got, CollectTimeout)
	}
}
