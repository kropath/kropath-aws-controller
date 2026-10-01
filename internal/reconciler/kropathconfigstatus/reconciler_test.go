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

package kropathconfigstatus

import (
	"context"
	"testing"

	"github.com/kropath/kropath-aws-controller/api/v1alpha1"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const globalNS = "kro-system"

func testScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme v1alpha1: %v", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme corev1: %v", err)
	}
	return scheme
}

func testReconciler(t *testing.T, objs ...runtime.Object) (*Reconciler, client.Client) {
	t.Helper()
	scheme := testScheme(t)
	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(objs...).
		WithStatusSubresource(&v1alpha1.KropathConfig{}).
		Build()
	return &Reconciler{Client: c, Reader: c, Scheme: scheme}, c
}

// namespace returns a namespace annotated to resolve its global tier to
// globalNS. AC-12 deleted the old "kro-system" default (spec §5.5): a
// namespace lacking the annotation is governance-only, not a silent fallback.
func namespace(name string) *corev1.Namespace {
	return namespaceWithAnnotation(name, globalNS)
}

func namespaceWithAnnotation(name, globalConfigNS string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: map[string]string{util.GlobalConfigNamespaceAnnotation: globalConfigNS},
		},
	}
}

func kropathConfig(ns string) *v1alpha1.KropathConfig {
	return &v1alpha1.KropathConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "KropathConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.KropathConfigName,
			Namespace: ns,
		},
	}
}

func s3Config(ns, name string) *v1alpha1.S3Config {
	return &v1alpha1.S3Config{
		TypeMeta:   metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "S3Config"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
	}
}

func iamConfig(ns, name string) *v1alpha1.IAMConfig {
	return &v1alpha1.IAMConfig{
		TypeMeta:   metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "IAMConfig"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
	}
}

func reconcileFor(t *testing.T, r *Reconciler, ns string) ctrl.Result {
	t.Helper()
	res, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: ns, Name: util.KropathConfigName},
	})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	return res
}

func getKPC(t *testing.T, c client.Client, ns string) *v1alpha1.KropathConfig {
	t.Helper()
	got := &v1alpha1.KropathConfig{}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: ns, Name: util.KropathConfigName}, got); err != nil {
		t.Fatalf("Get KropathConfig: %v", err)
	}
	return got
}

func findCondition(conds []metav1.Condition, condType string) *metav1.Condition {
	for i := range conds {
		if conds[i].Type == condType {
			return &conds[i]
		}
	}
	return nil
}

func TestReconcile_GlobalTier(t *testing.T) {
	// team-a is annotated to resolve its global tier to kro-system (AC-12: no
	// default, so this must be explicit). An S3Config living there makes the
	// kro-system KropathConfig resolve as team-a's global tier.
	r, c := testReconciler(t,
		namespace("team-a"),
		kropathConfig(globalNS),
		s3Config("team-a", "general-policy"),
	)

	reconcileFor(t, r, globalNS)

	got := getKPC(t, c, globalNS)
	cond := findCondition(got.Status.Conditions, ReconciledConditionType)
	if cond == nil {
		t.Fatal("expected a Reconciled condition, got none")
	}
	if cond.Status != metav1.ConditionTrue {
		t.Errorf("Status = %v, want True", cond.Status)
	}
	if cond.Reason != "GlobalTier" {
		t.Errorf("Reason = %q, want GlobalTier", cond.Reason)
	}
	if got.Status.ObservedGeneration != got.Generation {
		t.Errorf("ObservedGeneration = %d, want %d", got.Status.ObservedGeneration, got.Generation)
	}
}

func TestReconcile_LocalTier(t *testing.T) {
	// The KropathConfig lives in team-a itself, alongside an IAMConfig in the
	// same namespace: local tier, since team-a's own namespace (not its
	// resolved global namespace) is where the object lives.
	r, c := testReconciler(t,
		namespace("team-a"),
		kropathConfig("team-a"),
		iamConfig("team-a", "general-policy"),
	)

	reconcileFor(t, r, "team-a")

	got := getKPC(t, c, "team-a")
	cond := findCondition(got.Status.Conditions, ReconciledConditionType)
	if cond == nil {
		t.Fatal("expected a Reconciled condition, got none")
	}
	if cond.Status != metav1.ConditionTrue {
		t.Errorf("Status = %v, want True", cond.Status)
	}
	if cond.Reason != "LocalTier" {
		t.Errorf("Reason = %q, want LocalTier", cond.Reason)
	}
}

func TestReconcile_GlobalAndLocalTier(t *testing.T) {
	// "hub" hosts both an S3Config of its own (making its KropathConfig the
	// local tier) and is team-b's designated global-config-namespace, with an
	// IAMConfig living in team-b (making hub's KropathConfig the global tier
	// as seen from team-b).
	r, c := testReconciler(t,
		namespace("hub"),
		namespaceWithAnnotation("team-b", "hub"),
		kropathConfig("hub"),
		s3Config("hub", "general-policy"),
		iamConfig("team-b", "general-policy"),
	)

	reconcileFor(t, r, "hub")

	got := getKPC(t, c, "hub")
	cond := findCondition(got.Status.Conditions, ReconciledConditionType)
	if cond == nil {
		t.Fatal("expected a Reconciled condition, got none")
	}
	if cond.Status != metav1.ConditionTrue {
		t.Errorf("Status = %v, want True", cond.Status)
	}
	if cond.Reason != "GlobalAndLocalTier" {
		t.Errorf("Reason = %q, want GlobalAndLocalTier", cond.Reason)
	}
}

func TestReconcile_Unreferenced(t *testing.T) {
	// No family config anywhere resolves this namespace as either tier.
	r, c := testReconciler(t,
		namespace("orphan"),
		kropathConfig("orphan"),
	)

	reconcileFor(t, r, "orphan")

	got := getKPC(t, c, "orphan")
	cond := findCondition(got.Status.Conditions, ReconciledConditionType)
	if cond == nil {
		t.Fatal("expected a Reconciled condition, got none")
	}
	if cond.Status != metav1.ConditionFalse {
		t.Errorf("Status = %v, want False", cond.Status)
	}
	if cond.Reason != "Unreferenced" {
		t.Errorf("Reason = %q, want Unreferenced", cond.Reason)
	}
}

func TestReconcile_NotFound_NoError(t *testing.T) {
	r, _ := testReconciler(t)
	res := reconcileFor(t, r, "does-not-exist")
	if res.RequeueAfter != 0 {
		t.Errorf("RequeueAfter = %v, want 0 for a deleted object", res.RequeueAfter)
	}
}

func TestReconcile_RequeuesForFreshness(t *testing.T) {
	r, _ := testReconciler(t,
		namespace("orphan"),
		kropathConfig("orphan"),
	)
	res := reconcileFor(t, r, "orphan")
	if res.RequeueAfter != requeueInterval {
		t.Errorf("RequeueAfter = %v, want %v", res.RequeueAfter, requeueInterval)
	}
}

func TestReconcile_Idempotent_PreservesLastTransitionTime(t *testing.T) {
	r, c := testReconciler(t,
		namespace("team-a"),
		kropathConfig(globalNS),
		s3Config("team-a", "general-policy"),
	)

	reconcileFor(t, r, globalNS)
	first := findCondition(getKPC(t, c, globalNS).Status.Conditions, ReconciledConditionType)
	firstTransition := first.LastTransitionTime

	reconcileFor(t, r, globalNS)
	second := findCondition(getKPC(t, c, globalNS).Status.Conditions, ReconciledConditionType)

	if !second.LastTransitionTime.Equal(&firstTransition) {
		t.Errorf("LastTransitionTime changed on a no-op reconcile: %v -> %v", firstTransition, second.LastTransitionTime)
	}
}

func TestFamilyConfigKinds_ExcludesNonFamilyEntries(t *testing.T) {
	kinds := familyConfigKinds()
	if len(kinds) == 0 {
		t.Fatal("expected at least one family config kind")
	}
	for _, excluded := range []string{"KropathConfig", "PolicyDocument", "LabelOperator", "KropathConfigStatus"} {
		for _, k := range kinds {
			if k == excluded {
				t.Errorf("familyConfigKinds() unexpectedly includes %q", excluded)
			}
		}
	}
	var hasS3, hasIAM bool
	for _, k := range kinds {
		hasS3 = hasS3 || k == "S3Config"
		hasIAM = hasIAM || k == "IAMConfig"
	}
	if !hasS3 || !hasIAM {
		t.Errorf("familyConfigKinds() = %v, want it to include S3Config and IAMConfig", kinds)
	}
}

// spec §2.1: TierReasons() is the closed set the §10.2.1 label-value check
// enumerates against, and must match reconciledCondition's own reason
// literals exactly (asserted elsewhere in this file by TestReconcile_*).
func TestTierReasonsReturnsAllFourValues(t *testing.T) {
	got := TierReasons()
	want := []string{ReasonGlobalAndLocalTier, ReasonGlobalTier, ReasonLocalTier, ReasonUnreferenced}
	if len(got) != len(want) {
		t.Fatalf("TierReasons() = %v (len %d), want len %d", got, len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("TierReasons()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
