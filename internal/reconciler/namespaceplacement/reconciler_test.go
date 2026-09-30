// Copyright 2026 kropath Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package namespaceplacement

import (
	"context"
	"testing"

	"github.com/kropath/kropath-aws-controller/internal/reconciler/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

type fakeRecorder struct {
	events []event
}

type event struct {
	eventType, reason, message string
}

func (f *fakeRecorder) Event(object runtime.Object, eventtype, reason, message string) {
	f.events = append(f.events, event{eventtype, reason, message})
}

func testReconciler(t *testing.T, objs ...runtime.Object) (*Reconciler, client.Client, *fakeRecorder) {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme: %v", err)
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objs...).Build()
	rec := &fakeRecorder{}
	return &Reconciler{Client: c, Recorder: rec, Scheme: scheme}, c, rec
}

func namespaceWithAnnotations(name string, annotations map[string]string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Annotations: annotations},
	}
}

func getNamespace(t *testing.T, c client.Client, name string) *corev1.Namespace {
	t.Helper()
	ns := &corev1.Namespace{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: name}, ns); err != nil {
		t.Fatalf("get namespace %s: %v", name, err)
	}
	return ns
}

func reconcile(t *testing.T, r *Reconciler, name string) {
	t.Helper()
	if _, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{Name: name}}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
}

// AC-1 / AC-9: a fully annotated resource namespace with no local config object
// still gets a verdict published on the Namespace -- the case a config-object
// gate misses entirely.
func TestReconcileVerdictWithoutLocalConfig(t *testing.T) {
	r, c, rec := testReconciler(t, namespaceWithAnnotations("payments-prod", map[string]string{
		util.GlobalConfigNamespaceAnnotation: "platform-config",
		util.OwnerAccountIDAnnotation:        "111122223333",
		util.DefaultRegionAnnotation:         "ap-southeast-2",
	}))

	reconcile(t, r, "payments-prod")

	ns := getNamespace(t, c, "payments-prod")
	if got := ns.Annotations[util.PlacementStatusAnnotation]; got != util.PlacementStatusOK {
		t.Fatalf("placement-status = %q, want %q", got, util.PlacementStatusOK)
	}
	if len(rec.events) != 1 || rec.events[0].eventType != corev1.EventTypeNormal || rec.events[0].reason != util.ReasonPlacementResolved {
		t.Fatalf("events = %+v, want one Normal/%s event", rec.events, util.ReasonPlacementResolved)
	}
}

func TestReconcileMissingAccountAnnotation(t *testing.T) {
	r, c, rec := testReconciler(t, namespaceWithAnnotations("payments-prod", map[string]string{
		util.GlobalConfigNamespaceAnnotation: "platform-config",
		util.DefaultRegionAnnotation:         "ap-southeast-2",
	}))

	reconcile(t, r, "payments-prod")

	ns := getNamespace(t, c, "payments-prod")
	if got := ns.Annotations[util.PlacementStatusAnnotation]; got != util.ReasonMissingAccountAnnotation {
		t.Fatalf("placement-status = %q, want %q", got, util.ReasonMissingAccountAnnotation)
	}
	if len(rec.events) != 1 || rec.events[0].eventType != corev1.EventTypeWarning || rec.events[0].reason != util.ReasonMissingAccountAnnotation {
		t.Fatalf("events = %+v, want one Warning/%s event", rec.events, util.ReasonMissingAccountAnnotation)
	}
}

func TestReconcileTeamAnnotationGated(t *testing.T) {
	r, c, _ := testReconciler(t, namespaceWithAnnotations("payments-prod", map[string]string{
		util.GlobalConfigNamespaceAnnotation: "platform-config",
		util.OwnerAccountIDAnnotation:        "111122223333",
		util.DefaultRegionAnnotation:         "ap-southeast-2",
		util.TeamIDAnnotation:                "payments",
	}))

	reconcile(t, r, "payments-prod")

	ns := getNamespace(t, c, "payments-prod")
	if got := ns.Annotations[util.PlacementStatusAnnotation]; got != util.ReasonTeamAnnotationUnsupported {
		t.Fatalf("placement-status = %q, want %q — not %q", got, util.ReasonTeamAnnotationUnsupported, util.PlacementStatusOK)
	}
}

// AC-8: a governance-only namespace (no global-config-namespace) is exempt --
// no placement-status annotation is ever written.
func TestReconcileGovernanceOnlyNamespaceNotAnnotated(t *testing.T) {
	r, c, rec := testReconciler(t, namespaceWithAnnotations("platform-shared", nil))

	reconcile(t, r, "platform-shared")

	ns := getNamespace(t, c, "platform-shared")
	if _, ok := ns.Annotations[util.PlacementStatusAnnotation]; ok {
		t.Fatalf("placement-status annotation = %q, want absent for a governance-only namespace", ns.Annotations[util.PlacementStatusAnnotation])
	}
	if len(rec.events) != 0 {
		t.Fatalf("events = %+v, want none for a governance-only namespace", rec.events)
	}
}

// Events fire only on a verdict transition -- reconciling a namespace stuck in
// the same failure state twice must not produce a second Event (spec §6.2).
func TestReconcileEventOnlyOnTransition(t *testing.T) {
	r, c, rec := testReconciler(t, namespaceWithAnnotations("payments-prod", map[string]string{
		util.GlobalConfigNamespaceAnnotation: "platform-config",
		util.DefaultRegionAnnotation:         "ap-southeast-2",
	}))

	reconcile(t, r, "payments-prod")
	if len(rec.events) != 1 {
		t.Fatalf("after first reconcile: events = %+v, want 1", rec.events)
	}

	// Re-fetch to pick up the annotation written by the first reconcile, exactly
	// as a real watch-driven resync would see it.
	ns := getNamespace(t, c, "payments-prod")
	r2, _, rec2 := testReconciler(t, ns)
	rec2.events = rec.events // continue the same event log across the two "reconciles"
	reconcile(t, r2, "payments-prod")

	if len(rec2.events) != 1 {
		t.Fatalf("after second reconcile with unchanged verdict: events = %+v, want still 1 (no new Event)", rec2.events)
	}
}

func TestReconcileNotFoundIsNoop(t *testing.T) {
	r, _, rec := testReconciler(t)
	reconcile(t, r, "does-not-exist")
	if len(rec.events) != 0 {
		t.Fatalf("events = %+v, want none", rec.events)
	}
}

// transitionCounterValue reads kropath_namespaceplacement_transitions_total
// for the given "to" label straight from the shared registry --
// internal/metrics registers its own counter and does not export it, so this
// is the only black-box way to assert the increment from this package.
func transitionCounterValue(t *testing.T, to string) float64 {
	t.Helper()
	families, err := ctrlmetrics.Registry.Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	for _, mf := range families {
		if mf.GetName() != "kropath_namespaceplacement_transitions_total" {
			continue
		}
		for _, m := range mf.GetMetric() {
			for _, lp := range m.GetLabel() {
				if lp.GetName() == "to" && lp.GetValue() == to {
					return m.GetCounter().GetValue()
				}
			}
		}
	}
	return 0
}

// The counter shares the same prevStatus != newStatus guard as the Event
// (spec §2.5), so it must fire exactly once on the first reconcile (a
// transition into MissingAccountAnnotation) and not again on a second
// reconcile that reaches the same verdict.
func TestReconcileTransitionCounterFiresOnlyOnceOnTransition(t *testing.T) {
	const reason = "MissingAccountAnnotation"
	before := transitionCounterValue(t, reason)

	r, c, _ := testReconciler(t, namespaceWithAnnotations("obs-transition-counter-ns", map[string]string{
		util.GlobalConfigNamespaceAnnotation: "platform-config",
		util.DefaultRegionAnnotation:         "ap-southeast-2",
	}))
	reconcile(t, r, "obs-transition-counter-ns")

	afterFirst := transitionCounterValue(t, reason)
	if delta := afterFirst - before; delta != 1 {
		t.Fatalf("after first reconcile: %s delta = %v, want 1", reason, delta)
	}

	ns := getNamespace(t, c, "obs-transition-counter-ns")
	r2, _, _ := testReconciler(t, ns)
	reconcile(t, r2, "obs-transition-counter-ns")

	afterSecond := transitionCounterValue(t, reason)
	if delta := afterSecond - afterFirst; delta != 0 {
		t.Fatalf("after second reconcile with unchanged verdict: %s delta = %v, want 0 (no new increment)", reason, delta)
	}
}
