// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package metrics

import "testing"

func TestOutcomeValues_KnownMetric(t *testing.T) {
	got := OutcomeValues("kropath_labeloperator_patches_total")
	want := map[string]bool{"patched": true, "not_found": true, "error": true}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %d values matching %v", got, len(want), want)
	}
	for _, v := range got {
		if !want[v] {
			t.Errorf("unexpected value %q", v)
		}
	}
}

func TestOutcomeValues_UnknownMetric(t *testing.T) {
	if got := OutcomeValues("kropath_does_not_exist"); got != nil {
		t.Fatalf("got %v, want nil for an unrecognized metric", got)
	}
}

// TestOutcomeValues_ReturnsCopy guards against a caller's mutation of the
// returned slice corrupting the package-level closed set for every future
// caller (including the §10.2.1 label-value check, which unions this output).
func TestOutcomeValues_ReturnsCopy(t *testing.T) {
	got := OutcomeValues("kropath_cascade_mapfunc_errors_total")
	if len(got) == 0 {
		t.Fatal("expected a non-empty closed set")
	}
	got[0] = "mutated"

	again := OutcomeValues("kropath_cascade_mapfunc_errors_total")
	for _, v := range again {
		if v == "mutated" {
			t.Fatal("mutating a returned slice corrupted the package-level closed set")
		}
	}
}

func TestOutcomeMetrics_CoversEveryMapEntry(t *testing.T) {
	names := OutcomeMetrics()
	if len(names) != len(outcomeValueSets) {
		t.Fatalf("OutcomeMetrics() returned %d names, want %d (one per outcomeValueSets entry)", len(names), len(outcomeValueSets))
	}
	for _, name := range names {
		if OutcomeValues(name) == nil {
			t.Errorf("OutcomeMetrics() returned %q, but OutcomeValues(%q) is nil", name, name)
		}
	}
}
