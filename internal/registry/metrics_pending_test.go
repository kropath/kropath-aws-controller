// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

// pendingSeriesExists reports whether the global registry currently exposes
// kropath_registry_reconciler_pending_since_timestamp_seconds{package=pkg}.
// Gathering the real ctrlmetrics.Registry (rather than reaching into
// internal/metrics' unexported vars) exercises exactly what a scrape would
// see, which is the property M-17 cares about: an absent series, not a
// zeroed one.
func pendingSeriesExists(t *testing.T, pkg string) bool {
	t.Helper()
	families, err := ctrlmetrics.Registry.Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	for _, fam := range families {
		if fam.GetName() != "kropath_registry_reconciler_pending_since_timestamp_seconds" {
			continue
		}
		for _, m := range fam.GetMetric() {
			for _, lp := range m.GetLabel() {
				if lp.GetName() == "package" && lp.GetValue() == pkg {
					return true
				}
			}
		}
	}
	return false
}

func optionalKindsAttachedValue(t *testing.T, pkg string) (float64, bool) {
	t.Helper()
	families, err := ctrlmetrics.Registry.Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	for _, fam := range families {
		if fam.GetName() != "kropath_registry_optional_kinds_attached" {
			continue
		}
		for _, m := range fam.GetMetric() {
			for _, lp := range m.GetLabel() {
				if lp.GetName() == "package" && lp.GetValue() == pkg {
					return m.GetGauge().GetValue(), true
				}
			}
		}
	}
	return 0, false
}

// TestRunGate_PendingThenActive_ClearsPendingTimestamp asserts the M-17
// startup-gate half of the delete-on-activation contract (gate.go): a
// reconciler observed pending on one RunGate call must have no
// pending-since series once a later RunGate call finds all its Required
// GVKs served.
func TestRunGate_PendingThenActive_ClearsPendingTimestamp(t *testing.T) {
	pkg := "mtest-pending-clears-gate"
	gvk := schema.GroupVersionKind{Group: "mtest.kropath.run", Version: "v1alpha1", Kind: "MtestGatePending"}

	coord := &Coordinator{}
	coord.Add(Entry{
		Package:  pkg,
		Required: []schema.GroupVersionKind{gvk},
		Build:    metricsNoopBuild,
	})

	// First pass: gvk absent, entry stays pending.
	if err := coord.RunGate(metricsBuildCtx(), map[schema.GroupVersionKind]bool{KropathConfigGVK: true}); err != nil {
		t.Fatalf("RunGate (pending): %v", err)
	}
	if !pendingSeriesExists(t, pkg) {
		t.Fatal("expected pending-since series after first RunGate with the required GVK absent")
	}

	// Second pass: gvk now served, entry activates.
	if err := coord.RunGate(metricsBuildCtx(), map[schema.GroupVersionKind]bool{KropathConfigGVK: true, gvk: true}); err != nil {
		t.Fatalf("RunGate (active): %v", err)
	}
	if pendingSeriesExists(t, pkg) {
		t.Error("expected pending-since series to be absent once the reconciler activates via RunGate")
	}
}

// TestOnGVKServable_ActivatesPendingEntry_ClearsPendingTimestamp asserts the
// M-17 runtime-activation half (registry.go): a reconciler parked pending by
// RunGate must have its pending-since series deleted when the CRD watcher's
// OnGVKServable later activates it.
func TestOnGVKServable_ActivatesPendingEntry_ClearsPendingTimestamp(t *testing.T) {
	pkg := "mtest-pending-clears-ongvkservable"
	gvk := schema.GroupVersionKind{Group: "mtest.kropath.run", Version: "v1alpha1", Kind: "MtestRuntimePending"}

	coord := &Coordinator{}
	coord.Add(Entry{
		Package:  pkg,
		Required: []schema.GroupVersionKind{gvk},
		Build:    metricsNoopBuild,
	})

	if err := coord.RunGate(metricsBuildCtx(), map[schema.GroupVersionKind]bool{KropathConfigGVK: true}); err != nil {
		t.Fatalf("RunGate: %v", err)
	}
	if !pendingSeriesExists(t, pkg) {
		t.Fatal("expected pending-since series after RunGate with the required GVK absent")
	}

	if err := coord.OnGVKServable(metricsBuildCtx(), gvk); err != nil {
		t.Fatalf("OnGVKServable: %v", err)
	}
	if pendingSeriesExists(t, pkg) {
		t.Error("expected pending-since series to be absent once OnGVKServable activates the reconciler")
	}
}

// TestOnGVKServable_AttachOptional_SetsOptionalKindsAttachedGauge verifies
// registry.go reports the attached-optional-kind count via the new gauge
// when the CRD watcher attaches an optional kind to an already-active entry.
func TestOnGVKServable_AttachOptional_SetsOptionalKindsAttachedGauge(t *testing.T) {
	pkg := "mtest-optional-gauge-ongvkservable"
	reqGVK := schema.GroupVersionKind{Group: "mtest.kropath.run", Version: "v1alpha1", Kind: "MtestOptBase"}
	optGVK := schema.GroupVersionKind{Group: "mtest.kropath.run", Version: "v1alpha1", Kind: "MtestOptExtra"}

	coord := &Coordinator{}
	coord.Add(Entry{
		Package:  pkg,
		Required: []schema.GroupVersionKind{reqGVK},
		Optional: []schema.GroupVersionKind{optGVK},
		Build:    metricsNoopBuild,
		AddKindWatch: func(_ controller.Controller, _ schema.GroupVersionKind) error {
			return nil
		},
	})

	// Activate the entry first (reqGVK served, optGVK not yet).
	if err := coord.OnGVKServable(metricsBuildCtx(), reqGVK); err != nil {
		t.Fatalf("OnGVKServable (activate): %v", err)
	}
	if got, ok := optionalKindsAttachedValue(t, pkg); ok && got != 0 {
		t.Errorf("optional-kinds-attached before any attach = %v, want 0 or absent", got)
	}

	// Now attach the optional kind.
	if err := coord.OnGVKServable(metricsBuildCtx(), optGVK); err != nil {
		t.Fatalf("OnGVKServable (attach optional): %v", err)
	}
	got, ok := optionalKindsAttachedValue(t, pkg)
	if !ok {
		t.Fatal("expected an optional-kinds-attached series after attaching one optional kind")
	}
	if got != 1 {
		t.Errorf("optional-kinds-attached = %v, want 1", got)
	}
}

// TestRunGate_ServedOptionalAtStartup_SetsOptionalKindsAttachedGauge verifies
// gate.go reports the attached count when an entry activates at startup with
// some Optional GVKs already served.
func TestRunGate_ServedOptionalAtStartup_SetsOptionalKindsAttachedGauge(t *testing.T) {
	pkg := "mtest-optional-gauge-rungate"
	reqGVK := schema.GroupVersionKind{Group: "mtest.kropath.run", Version: "v1alpha1", Kind: "MtestGateOptBase"}
	optGVK := schema.GroupVersionKind{Group: "mtest.kropath.run", Version: "v1alpha1", Kind: "MtestGateOptExtra"}

	coord := &Coordinator{}
	coord.Add(Entry{
		Package:  pkg,
		Required: []schema.GroupVersionKind{reqGVK},
		Optional: []schema.GroupVersionKind{optGVK},
		Build:    metricsNoopBuild,
	})

	served := map[schema.GroupVersionKind]bool{
		KropathConfigGVK: true,
		reqGVK:           true,
		optGVK:           true,
	}
	if err := coord.RunGate(metricsBuildCtx(), served); err != nil {
		t.Fatalf("RunGate: %v", err)
	}

	got, ok := optionalKindsAttachedValue(t, pkg)
	if !ok {
		t.Fatal("expected an optional-kinds-attached series after RunGate with the optional GVK already served")
	}
	if got != 1 {
		t.Errorf("optional-kinds-attached = %v, want 1", got)
	}
}
