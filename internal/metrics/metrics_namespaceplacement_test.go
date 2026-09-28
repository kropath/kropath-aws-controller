// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNamespacePlacementTransition(t *testing.T) {
	before := readCounterValue(t, namespaceplacementTransitionsTotal, prometheus.Labels{"to": "ok"})

	NamespacePlacementTransition("ok")

	after := readCounterValue(t, namespaceplacementTransitionsTotal, prometheus.Labels{"to": "ok"})
	if delta := after - before; delta != 1 {
		t.Errorf("namespaceplacementTransitionsTotal delta: want 1, got %v", delta)
	}
}
