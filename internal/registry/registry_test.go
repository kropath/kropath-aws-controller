// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package registry_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kropath/kropath-aws-controller/internal/registry"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// ---- helpers ---------------------------------------------------------------

func noopBuild(_ registry.BuildCtx, _ []schema.GroupVersionKind) (controller.Controller, error) {
	return nil, nil
}

func failBuild(_ registry.BuildCtx, _ []schema.GroupVersionKind) (controller.Controller, error) {
	return nil, fmt.Errorf("build intentionally failed")
}

func awsGVK(kind string) schema.GroupVersionKind {
	return schema.GroupVersionKind{Group: "aws.kropath.run", Version: "v1alpha1", Kind: kind}
}

func testBuildCtx() registry.BuildCtx {
	return registry.BuildCtx{
		Manager: nil, // tests that don't call Build don't need a real manager
		Log:     logr.Discard(),
	}
}

// stubLister implements registry.ResourceLister for tests.
type stubLister struct {
	resources []*metav1.APIResourceList
	err       error
}

func (s *stubLister) ServerPreferredResources() ([]*metav1.APIResourceList, error) {
	return s.resources, s.err
}

// stubListerFromGVKs builds a stubLister that reports the given GVKs as served.
func stubListerFromGVKs(gvks []schema.GroupVersionKind) *stubLister {
	byGV := map[string][]metav1.APIResource{}
	for _, gvk := range gvks {
		gv := gvk.Group + "/" + gvk.Version
		byGV[gv] = append(byGV[gv], metav1.APIResource{Name: strings.ToLower(gvk.Kind) + "s", Kind: gvk.Kind})
	}
	var lists []*metav1.APIResourceList
	for gv, resources := range byGV {
		lists = append(lists, &metav1.APIResourceList{
			GroupVersion: gv,
			APIResources: resources,
		})
	}
	return &stubLister{resources: lists}
}

// ---- GatherServedGVKs tests ------------------------------------------------

func TestGatherServedGVKs_Empty(t *testing.T) {
	dc := stubListerFromGVKs(nil)
	got, err := registry.GatherServedGVKs(dc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty map, got %d entries", len(got))
	}
}

func TestGatherServedGVKs_PopulatesMap(t *testing.T) {
	want := []schema.GroupVersionKind{
		awsGVK("S3Config"),
		awsGVK("KropathConfig"),
	}
	dc := stubListerFromGVKs(want)
	got, err := registry.GatherServedGVKs(dc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, gvk := range want {
		if !got[gvk] {
			t.Errorf("expected %v in served set, not found", gvk)
		}
	}
}

// ---- RunGate: KropathConfig precondition -----------------------------------

func TestRunGate_KropathConfigAbsent_ReturnsError(t *testing.T) {
	coord := &registry.Coordinator{}
	coord.Add(registry.Entry{
		Package:  "s3config",
		Required: []schema.GroupVersionKind{awsGVK("S3Config"), awsGVK("KropathConfig")},
		Build:    noopBuild,
	})
	err := coord.RunGate(testBuildCtx(), map[schema.GroupVersionKind]bool{})
	if err == nil {
		t.Fatal("expected error when KropathConfig is absent, got nil")
	}
	if !strings.Contains(err.Error(), "KropathConfig") {
		t.Errorf("error should mention KropathConfig, got: %v", err)
	}
}

func TestRunGate_KropathConfigAbsent_NoControllersBuilt(t *testing.T) {
	built := false
	coord := &registry.Coordinator{}
	coord.Add(registry.Entry{
		Package:  "s3config",
		Required: []schema.GroupVersionKind{awsGVK("S3Config"), awsGVK("KropathConfig")},
		Build: func(_ registry.BuildCtx, _ []schema.GroupVersionKind) (controller.Controller, error) {
			built = true
			return nil, nil
		},
	})
	_ = coord.RunGate(testBuildCtx(), map[schema.GroupVersionKind]bool{})
	if built {
		t.Error("Build should not have been called when KropathConfig precondition fails")
	}
}

// ---- RunGate: gating logic -------------------------------------------------

func TestRunGate_AllRequiredPresent_BecomesActive(t *testing.T) {
	coord := &registry.Coordinator{}
	coord.Add(registry.Entry{
		Package:  "s3config",
		Required: []schema.GroupVersionKind{awsGVK("S3Config"), awsGVK("KropathConfig")},
		Build:    noopBuild,
	})
	served := map[schema.GroupVersionKind]bool{
		awsGVK("KropathConfig"): true,
		awsGVK("S3Config"):      true,
	}
	if err := coord.RunGate(testBuildCtx(), served); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !coord.ActivePackages()["s3config"] {
		t.Error("s3config should be active")
	}
}

func TestRunGate_MissingRequired_StaysPending(t *testing.T) {
	coord := &registry.Coordinator{}
	coord.Add(registry.Entry{
		Package:  "elbconfig",
		Required: []schema.GroupVersionKind{awsGVK("ELBConfig"), awsGVK("KropathConfig")},
		Build:    noopBuild,
	})
	// KropathConfig present, but ELBConfig absent
	served := map[schema.GroupVersionKind]bool{
		awsGVK("KropathConfig"): true,
	}
	if err := coord.RunGate(testBuildCtx(), served); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if coord.ActivePackages()["elbconfig"] {
		t.Error("elbconfig should remain pending when ELBConfig CRD is absent")
	}
}

func TestRunGate_PartialRequired_StaysPending(t *testing.T) {
	coord := &registry.Coordinator{}
	coord.Add(registry.Entry{
		Package:  "multi",
		Required: []schema.GroupVersionKind{awsGVK("AConfig"), awsGVK("BConfig"), awsGVK("KropathConfig")},
		Build:    noopBuild,
	})
	// Only AConfig served (besides KropathConfig)
	served := map[schema.GroupVersionKind]bool{
		awsGVK("KropathConfig"): true,
		awsGVK("AConfig"):       true,
	}
	if err := coord.RunGate(testBuildCtx(), served); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if coord.ActivePackages()["multi"] {
		t.Error("entry with partial Required should remain pending")
	}
}

func TestRunGate_BuildError_Propagates(t *testing.T) {
	coord := &registry.Coordinator{}
	coord.Add(registry.Entry{
		Package:  "broken",
		Required: []schema.GroupVersionKind{awsGVK("KropathConfig")},
		Build:    failBuild,
	})
	served := map[schema.GroupVersionKind]bool{awsGVK("KropathConfig"): true}
	err := coord.RunGate(testBuildCtx(), served)
	if err == nil {
		t.Fatal("expected error from failing Build, got nil")
	}
	if !strings.Contains(err.Error(), "broken") {
		t.Errorf("error should name the package, got: %v", err)
	}
}

// ---- RunUnconditional tests ------------------------------------------------

func TestRunUnconditional_AllActive(t *testing.T) {
	coord := &registry.Coordinator{}
	for _, pkg := range []string{"a", "b", "c"} {
		p := pkg
		coord.Add(registry.Entry{
			Package: p,
			Build:   noopBuild,
		})
	}
	if err := coord.RunUnconditional(testBuildCtx()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	active := coord.ActivePackages()
	for _, pkg := range []string{"a", "b", "c"} {
		if !active[pkg] {
			t.Errorf("package %q should be active after RunUnconditional", pkg)
		}
	}
}

func TestRunUnconditional_SkipsAlreadyActive(t *testing.T) {
	buildCount := 0
	coord := &registry.Coordinator{}
	coord.Add(registry.Entry{
		Package: "x",
		Build: func(_ registry.BuildCtx, _ []schema.GroupVersionKind) (controller.Controller, error) {
			buildCount++
			return nil, nil
		},
	})
	if err := coord.RunUnconditional(testBuildCtx()); err != nil {
		t.Fatalf("first run error: %v", err)
	}
	if err := coord.RunUnconditional(testBuildCtx()); err != nil {
		t.Fatalf("second run error: %v", err)
	}
	if buildCount != 1 {
		t.Errorf("Build should be called exactly once, got %d", buildCount)
	}
}

// ---- Concurrent access (race detector) ------------------------------------

func TestCoordinator_ConcurrentRunGate(t *testing.T) {
	coord := &registry.Coordinator{}
	for i := 0; i < 20; i++ {
		n := fmt.Sprintf("pkg%d", i)
		coord.Add(registry.Entry{
			Package:  n,
			Required: []schema.GroupVersionKind{awsGVK("KropathConfig")},
			Build:    noopBuild,
		})
	}
	served := map[schema.GroupVersionKind]bool{awsGVK("KropathConfig"): true}

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = coord.RunGate(testBuildCtx(), served)
		}()
	}
	wg.Wait()
}

// ---- OnGVKServable tests ---------------------------------------------------

func TestOnGVKServable_ActivatesPendingEntry(t *testing.T) {
	coord := &registry.Coordinator{}
	coord.Add(registry.Entry{
		Package:  "elbconfig",
		Required: []schema.GroupVersionKind{awsGVK("ELBConfig"), awsGVK("KropathConfig")},
		Build:    noopBuild,
	})
	// Startup: KropathConfig present but ELBConfig absent → elbconfig pending.
	served := map[schema.GroupVersionKind]bool{awsGVK("KropathConfig"): true}
	if err := coord.RunGate(testBuildCtx(), served); err != nil {
		t.Fatalf("RunGate error: %v", err)
	}
	if coord.ActivePackages()["elbconfig"] {
		t.Fatal("elbconfig should be pending after startup")
	}

	// CRD watcher signals ELBConfig is now served.
	if err := coord.OnGVKServable(testBuildCtx(), awsGVK("ELBConfig")); err != nil {
		t.Fatalf("OnGVKServable error: %v", err)
	}
	if !coord.ActivePackages()["elbconfig"] {
		t.Error("elbconfig should be active after ELBConfig CRD arrives")
	}
}

func TestOnGVKServable_DoubleRegistration_IsNoop(t *testing.T) {
	buildCount := 0
	coord := &registry.Coordinator{}
	coord.Add(registry.Entry{
		Package:  "elbconfig",
		Required: []schema.GroupVersionKind{awsGVK("ELBConfig"), awsGVK("KropathConfig")},
		Build: func(_ registry.BuildCtx, _ []schema.GroupVersionKind) (controller.Controller, error) {
			buildCount++
			return nil, nil
		},
	})
	served := map[schema.GroupVersionKind]bool{awsGVK("KropathConfig"): true}
	if err := coord.RunGate(testBuildCtx(), served); err != nil {
		t.Fatalf("RunGate error: %v", err)
	}

	// Signal ELBConfig twice — Build must only run once.
	if err := coord.OnGVKServable(testBuildCtx(), awsGVK("ELBConfig")); err != nil {
		t.Fatalf("first OnGVKServable error: %v", err)
	}
	if err := coord.OnGVKServable(testBuildCtx(), awsGVK("ELBConfig")); err != nil {
		t.Fatalf("second OnGVKServable error: %v", err)
	}
	if buildCount != 1 {
		t.Errorf("Build should be called exactly once, got %d", buildCount)
	}
}

func TestOnGVKServable_AttachesOptionalToActiveEntry(t *testing.T) {
	watchCalls := 0
	coord := &registry.Coordinator{}
	coord.Add(registry.Entry{
		Package:  "policydocument",
		Required: []schema.GroupVersionKind{awsGVK("PolicyDocument"), awsGVK("KropathConfig")},
		Optional: []schema.GroupVersionKind{awsGVK("AWSLambdaFunction")},
		Build:    noopBuild,
		AddKindWatch: func(_ controller.Controller, _ schema.GroupVersionKind) error {
			watchCalls++
			return nil
		},
	})
	// Both Required served at startup → policydocument active.
	served := map[schema.GroupVersionKind]bool{
		awsGVK("PolicyDocument"): true,
		awsGVK("KropathConfig"):  true,
	}
	if err := coord.RunGate(testBuildCtx(), served); err != nil {
		t.Fatalf("RunGate error: %v", err)
	}
	if !coord.ActivePackages()["policydocument"] {
		t.Fatal("policydocument should be active")
	}

	// Optional kind arrives later.
	if err := coord.OnGVKServable(testBuildCtx(), awsGVK("AWSLambdaFunction")); err != nil {
		t.Fatalf("OnGVKServable error: %v", err)
	}
	if watchCalls != 1 {
		t.Errorf("AddKindWatch should be called once, got %d", watchCalls)
	}

	// Calling again should be a no-op.
	if err := coord.OnGVKServable(testBuildCtx(), awsGVK("AWSLambdaFunction")); err != nil {
		t.Fatalf("second OnGVKServable error: %v", err)
	}
	if watchCalls != 1 {
		t.Errorf("AddKindWatch should not be called again, got %d total", watchCalls)
	}
}

// ---- ReconcilerState tests -------------------------------------------------

func TestReconcilerState_ActiveEntry(t *testing.T) {
	coord := &registry.Coordinator{}
	coord.Add(registry.Entry{
		Package:  "s3config",
		Required: []schema.GroupVersionKind{awsGVK("S3Config"), awsGVK("KropathConfig")},
		Build:    noopBuild,
	})
	served := map[schema.GroupVersionKind]bool{
		awsGVK("KropathConfig"): true,
		awsGVK("S3Config"):      true,
	}
	if err := coord.RunGate(testBuildCtx(), served); err != nil {
		t.Fatalf("RunGate error: %v", err)
	}
	state, missing := coord.ReconcilerState("s3config")
	if state != "active" {
		t.Errorf("state: got %q, want active", state)
	}
	if len(missing) != 0 {
		t.Errorf("missingKinds: got %v, want empty", missing)
	}
}

func TestReconcilerState_PendingEntry(t *testing.T) {
	coord := &registry.Coordinator{}
	coord.Add(registry.Entry{
		Package:  "elbconfig",
		Required: []schema.GroupVersionKind{awsGVK("ELBConfig"), awsGVK("KropathConfig")},
		Build:    noopBuild,
	})
	served := map[schema.GroupVersionKind]bool{awsGVK("KropathConfig"): true}
	if err := coord.RunGate(testBuildCtx(), served); err != nil {
		t.Fatalf("RunGate error: %v", err)
	}
	state, missing := coord.ReconcilerState("elbconfig")
	if state != "pending" {
		t.Errorf("state: got %q, want pending", state)
	}
	if len(missing) == 0 {
		t.Error("missingKinds should not be empty for pending entry")
	}
	found := false
	for _, k := range missing {
		if k == "ELBConfig" {
			found = true
		}
	}
	if !found {
		t.Errorf("missingKinds should contain ELBConfig, got %v", missing)
	}
}

func TestReconcilerState_UnknownPackage(t *testing.T) {
	coord := &registry.Coordinator{}
	state, missing := coord.ReconcilerState("doesnotexist")
	if state != "" {
		t.Errorf("state: got %q, want empty string", state)
	}
	if len(missing) != 0 {
		t.Errorf("missingKinds: got %v, want nil", missing)
	}
}

// ---- Registry completeness: every entry declares Required/Optional ---------

func TestAllEntries_CascadeHaveKropathConfigRequired(t *testing.T) {
	kpc := awsGVK("KropathConfig")
	for _, e := range registry.All() {
		if e.Package == "labeloperator" || e.Package == "namespaceplacement" {
			continue // no Required by design: labeloperator spans provider groups;
			// namespaceplacement watches only the core Namespace kind
		}
		found := false
		for _, gvk := range e.Required {
			if gvk == kpc {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("entry %q is missing KropathConfig in Required", e.Package)
		}
	}
}

func TestAllEntries_PolicyDocumentHasOptionalRefGVKs(t *testing.T) {
	for _, e := range registry.All() {
		if e.Package != "policydocument" {
			continue
		}
		if len(e.Optional) == 0 {
			t.Error("policydocument entry must have Optional ref GVKs")
		}
		// All optional GVKs must be in aws.kropath.run
		for _, gvk := range e.Optional {
			if gvk.Group != "aws.kropath.run" {
				t.Errorf("policydocument optional GVK %v is not in aws.kropath.run", gvk)
			}
		}
		return
	}
	t.Error("policydocument entry not found in registry.All()")
}

// TestAllEntries_NoDuplicatePackages verifies package names are unique.
func TestAllEntries_NoDuplicatePackages(t *testing.T) {
	seen := map[string]bool{}
	for _, e := range registry.All() {
		if seen[e.Package] {
			t.Errorf("duplicate package %q in registry.All()", e.Package)
		}
		seen[e.Package] = true
	}
}

// Ensure ctrl import is used (BuildCtx carries ctrl.Manager).
var _ ctrl.Manager
