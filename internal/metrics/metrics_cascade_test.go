// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestCascadeMapFuncError(t *testing.T) {
	before := readCounterValue(t, cascadeMapfuncErrorsTotal, prometheus.Labels{"family": "s3config", "trigger": "kropathconfig"})

	CascadeMapFuncError("s3config", "kropathconfig")

	after := readCounterValue(t, cascadeMapfuncErrorsTotal, prometheus.Labels{"family": "s3config", "trigger": "kropathconfig"})
	if delta := after - before; delta != 1 {
		t.Errorf("cascadeMapfuncErrorsTotal delta: want 1, got %v", delta)
	}
}
