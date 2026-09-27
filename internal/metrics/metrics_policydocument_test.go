// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestPolicyDocumentRefResolution(t *testing.T) {
	labels := prometheus.Labels{"kind": "AWSIAMRole", "field": "predictedArn", "outcome": "resolved"}
	before := readCounterValue(t, policydocumentRefResolutionsTotal, labels)
	PolicyDocumentRefResolution("AWSIAMRole", "predictedArn", "resolved")
	if delta := readCounterValue(t, policydocumentRefResolutionsTotal, labels) - before; delta != 1 {
		t.Errorf("policydocumentRefResolutionsTotal delta: want 1, got %v", delta)
	}
}

func TestPolicyDocumentSidConflict(t *testing.T) {
	before := readSidConflictCounter(t)
	PolicyDocumentSidConflict()
	if delta := readSidConflictCounter(t) - before; delta != 1 {
		t.Errorf("policydocumentSidConflictsTotal delta: want 1, got %v", delta)
	}
}

func readSidConflictCounter(t *testing.T) float64 {
	t.Helper()
	var m dto.Metric
	if err := policydocumentSidConflictsTotal.Write(&m); err != nil {
		t.Fatalf("metric Write: %v", err)
	}
	return m.GetCounter().GetValue()
}
