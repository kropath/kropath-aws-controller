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

package kmsconfig

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kropath/kropath-aws-controller/api/v1alpha1"
	"github.com/kropath/kropath-aws-controller/internal/cascade"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestReconcileAC1GlobalMandatoryEnableRotationWins(t *testing.T) {
	rec, cfg := testReconciler(t,
		globalKropathConfig(v1alpha1.KropathConfigTier{
			KMS: cascade.KMSKropathSection{EnableKeyRotation: true},
		}),
		localKMSConfig("payments-prod", "general-policy", cascade.KMSConfigSection{}, cascade.KMSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getKMSConfig(t, rec.Client, "payments-prod", "general-policy")
	if !updated.Status.EffectiveConfig.Mandatory.EnableKeyRotation {
		t.Fatal("mandatory.enableKeyRotation = false, want true")
	}
	if updated.Status.ObservedGeneration != cfg.Generation {
		t.Fatalf("observedGeneration = %d, want %d", updated.Status.ObservedGeneration, cfg.Generation)
	}
}

func TestReconcileAC2Level1MandatoryWinsOverLevel3(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig(v1alpha1.KropathConfigTier{
			KMS: cascade.KMSKropathSection{AllowedKeySpecs: []string{"SYMMETRIC_DEFAULT"}},
		}),
		globalKMSConfig("general-policy",
			cascade.KMSConfigSection{AllowedKeySpecs: []string{"RSA_4096"}},
			cascade.KMSConfigSection{},
		),
		localKMSConfig("payments-prod", "general-policy", cascade.KMSConfigSection{}, cascade.KMSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getKMSConfig(t, rec.Client, "payments-prod", "general-policy")
	got := updated.Status.EffectiveConfig.Mandatory.AllowedKeySpecs
	if len(got) != 1 || got[0] != "SYMMETRIC_DEFAULT" {
		t.Fatalf("mandatory.allowedKeySpecs = %v, want [SYMMETRIC_DEFAULT] (level-1 value)", got)
	}
}

func TestReconcileAC4KeySpecFromGlobalKMSConfig(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKMSConfig("general-policy", cascade.KMSConfigSection{KeySpec: "SYMMETRIC_DEFAULT"}, cascade.KMSConfigSection{}),
		localKMSConfig("payments-prod", "general-policy", cascade.KMSConfigSection{}, cascade.KMSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getKMSConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.KeySpec; got != "SYMMETRIC_DEFAULT" {
		t.Fatalf("mandatory.keySpec = %q, want SYMMETRIC_DEFAULT", got)
	}
}

func TestReconcileAC5DefaultsOnlyKeySpec(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKMSConfig("general-policy", cascade.KMSConfigSection{}, cascade.KMSConfigSection{KeySpec: "RSA_2048"}),
		localKMSConfig("payments-prod", "general-policy", cascade.KMSConfigSection{}, cascade.KMSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getKMSConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Defaults.KeySpec; got != "RSA_2048" {
		t.Fatalf("defaults.keySpec = %q, want RSA_2048", got)
	}
	if got := updated.Status.EffectiveConfig.Mandatory.KeySpec; got != "" {
		t.Fatalf("mandatory.keySpec = %q, want empty", got)
	}
}

func TestReconcileAC9InvalidKeySpecSetsInvalidCondition(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig(v1alpha1.KropathConfigTier{
			KMS: cascade.KMSKropathSection{AllowedKeySpecs: []string{"SYMMETRIC_DEFAULT"}},
		}),
		localKMSConfig("payments-prod", "general-policy", cascade.KMSConfigSection{KeySpec: "RSA_4096"}, cascade.KMSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getKMSConfig(t, rec.Client, "payments-prod", "general-policy")
	cond := findCondition(updated.Status.Conditions, "Valid")
	if cond == nil {
		t.Fatal("missing Valid condition")
	}
	if cond.Status != metav1.ConditionFalse {
		t.Fatalf("Valid condition status = %v, want False", cond.Status)
	}
	if cond.Reason != "InvalidKeySpecNotInAllowedList" {
		t.Fatalf("Valid condition reason = %q, want InvalidKeySpecNotInAllowedList", cond.Reason)
	}
	// On validation failure the reconciler must not write status.effectiveConfig.
	if updated.Status.EffectiveConfig.Mandatory.KeySpec != "" {
		t.Fatalf("effectiveConfig.mandatory.keySpec = %q, want untouched (empty)", updated.Status.EffectiveConfig.Mandatory.KeySpec)
	}
}

func TestReconcileAC10ValidKeySpecSetsValidCondition(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig(v1alpha1.KropathConfigTier{
			KMS: cascade.KMSKropathSection{AllowedKeySpecs: []string{"SYMMETRIC_DEFAULT", "RSA_4096"}},
		}),
		localKMSConfig("payments-prod", "general-policy", cascade.KMSConfigSection{KeySpec: "RSA_4096"}, cascade.KMSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getKMSConfig(t, rec.Client, "payments-prod", "general-policy")
	cond := findCondition(updated.Status.Conditions, "Valid")
	if cond == nil || cond.Status != metav1.ConditionTrue {
		t.Fatalf("Valid condition = %+v, want True", cond)
	}
	if got := updated.Status.EffectiveConfig.Mandatory.KeySpec; got != "RSA_4096" {
		t.Fatalf("effectiveConfig.mandatory.keySpec = %q, want RSA_4096", got)
	}
}

func TestReconcileAC15TagUnionMerge(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig(v1alpha1.KropathConfigTier{
			Tags: map[string]string{"owner": "platform-team", "shared-key": "from-global-kropath"},
		}),
		localKMSConfig("payments-prod", "general-policy",
			cascade.KMSConfigSection{Tags: map[string]string{"team": "payments", "shared-key": "from-local-kmsconfig"}},
			cascade.KMSConfigSection{},
		),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getKMSConfig(t, rec.Client, "payments-prod", "general-policy")
	tags := updated.Status.EffectiveConfig.Mandatory.Tags
	if tags["owner"] != "platform-team" {
		t.Fatalf("tags[owner] = %q, want platform-team", tags["owner"])
	}
	if tags["team"] != "payments" {
		t.Fatalf("tags[team] = %q, want payments", tags["team"])
	}
	// Level 1 (global KropathConfig mandatory) wins over level 4 (local KMSConfig mandatory)
	// on key conflicts.
	if tags["shared-key"] != "from-global-kropath" {
		t.Fatalf("tags[shared-key] = %q, want from-global-kropath (level-1 wins)", tags["shared-key"])
	}
}

func TestReconcileCopiesAWSIdentity(t *testing.T) {
	rec, _ := testReconciler(t,
		namespaceWithIdentity("payments-prod", "123456789012", "us-east-1"),
		localKMSConfig("payments-prod", "general-policy", cascade.KMSConfigSection{}, cascade.KMSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getKMSConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.AWS.AccountID; got != "123456789012" {
		t.Fatalf("aws.accountId = %q, want 123456789012", got)
	}
	if got := updated.Status.EffectiveConfig.AWS.Region; got != "us-east-1" {
		t.Fatalf("aws.region = %q, want us-east-1", got)
	}
}

func TestRequestsForKropathConfigChangeGlobal(t *testing.T) {
	rec, _ := testReconciler(t,
		localKMSConfig("payments-prod", "general-policy", cascade.KMSConfigSection{}, cascade.KMSConfigSection{}),
		localKMSConfig("sandbox", "general-policy", cascade.KMSConfigSection{}, cascade.KMSConfigSection{}),
		localKMSConfig("payments-prod", "other-policy", cascade.KMSConfigSection{}, cascade.KMSConfigSection{}),
	)

	got := rec.requestsForKropathConfigChange(context.Background(), &v1alpha1.KropathConfig{
		ObjectMeta: metav1.ObjectMeta{Name: util.KropathConfigName, Namespace: "kro-system"},
	})

	want := map[string]bool{
		"payments-prod/general-policy": false,
		"sandbox/general-policy":       false,
		"payments-prod/other-policy":   false,
	}
	if len(got) != len(want) {
		t.Fatalf("requests len = %d, want %d (%#v)", len(got), len(want), got)
	}
	for _, r := range got {
		key := r.Namespace + "/" + r.Name
		if _, ok := want[key]; !ok {
			t.Fatalf("unexpected request %q", key)
		}
		want[key] = true
	}
	for key, seen := range want {
		if !seen {
			t.Fatalf("missing request %q", key)
		}
	}
}

func TestRequestsForKropathConfigChangeUnrelatedNamespaceIgnored(t *testing.T) {
	// Classification is by namespace only (ADR-018 D-1): a KropathConfig in a
	// namespace that is neither an item's own namespace nor the resolved
	// global namespace for any item triggers 0 requests.
	rec, _ := testReconciler(t,
		localKMSConfig("payments-prod", "general-policy", cascade.KMSConfigSection{}, cascade.KMSConfigSection{}),
		localKMSConfig("sandbox", "general-policy", cascade.KMSConfigSection{}, cascade.KMSConfigSection{}),
	)

	got := rec.requestsForKropathConfigChange(context.Background(), &v1alpha1.KropathConfig{
		ObjectMeta: metav1.ObjectMeta{Name: util.KropathConfigName, Namespace: "unrelated-ns"},
	})

	if len(got) != 0 {
		t.Fatalf("requests len = %d, want 0 — unrelated namespace triggers nothing (%#v)", len(got), got)
	}
}

func TestRequestsForKropathConfigChangeLocalNamespaceEnqueuesAllConfigsInNamespace(t *testing.T) {
	// A KropathConfig in namespace X is the local tier for every config in X,
	// regardless of its name (the name is now a fixed singleton).
	rec, _ := testReconciler(t,
		localKMSConfig("payments-prod", "general-policy", cascade.KMSConfigSection{}, cascade.KMSConfigSection{}),
		localKMSConfig("payments-prod", "other-policy", cascade.KMSConfigSection{}, cascade.KMSConfigSection{}),
		localKMSConfig("sandbox", "general-policy", cascade.KMSConfigSection{}, cascade.KMSConfigSection{}),
	)

	got := rec.requestsForKropathConfigChange(context.Background(), &v1alpha1.KropathConfig{
		ObjectMeta: metav1.ObjectMeta{Name: util.KropathConfigName, Namespace: "payments-prod"},
	})

	if len(got) != 2 {
		t.Fatalf("requests len = %d, want 2 — local KPC triggers all configs in its namespace (%#v)", len(got), got)
	}
}

func TestRequestsForKMSConfigChangeGlobal(t *testing.T) {
	rec, _ := testReconciler(t,
		localKMSConfig("payments-prod", "general-policy", cascade.KMSConfigSection{}, cascade.KMSConfigSection{}),
		localKMSConfig("data-prod", "general-policy", cascade.KMSConfigSection{}, cascade.KMSConfigSection{}),
	)

	got := rec.requestsForKMSConfigChange(context.Background(), &v1alpha1.KMSConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "general-policy", Namespace: "kro-system"},
	})

	if len(got) != 2 {
		t.Fatalf("requests len = %d, want 2 (%#v)", len(got), got)
	}
}

func TestRequestsForKMSConfigChangeNonGlobalIgnored(t *testing.T) {
	rec, _ := testReconciler(t,
		localKMSConfig("payments-prod", "general-policy", cascade.KMSConfigSection{}, cascade.KMSConfigSection{}),
	)

	got := rec.requestsForKMSConfigChange(context.Background(), &v1alpha1.KMSConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "general-policy", Namespace: "payments-prod"},
	})

	if len(got) != 0 {
		t.Fatalf("requests len = %d, want 0 (%#v)", len(got), got)
	}
}

// ─── Test fixtures ──────────────────────────────────────────────────────────

func testReconciler(t *testing.T, objs ...runtime.Object) (*Reconciler, *v1alpha1.KMSConfig) {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("add corev1 scheme: %v", err)
	}
	objs = withDefaultNamespaces(objs)
	builder := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&v1alpha1.KMSConfig{})
	for _, obj := range objs {
		builder = builder.WithRuntimeObjects(obj)
	}
	cl := builder.Build()
	cfg := &v1alpha1.KMSConfig{}
	if err := cl.Get(context.Background(), client.ObjectKey{Namespace: "payments-prod", Name: "general-policy"}, cfg); err != nil {
		t.Fatalf("seed local KMSConfig: %v", err)
	}
	return &Reconciler{Client: cl, Log: logr.Discard(), Scheme: scheme}, cfg
}

// withDefaultNamespaces auto-seeds a resource namespace (valid account/region,
// global tier "kro-system") for every namespace referenced by objs that isn't
// "kro-system" and doesn't already have an explicit Namespace object among objs.
// Most tests in this file exercise cascade-merge logic and don't care about
// placement specifics, so this keeps them from repeating that boilerplate.
func withDefaultNamespaces(objs []runtime.Object) []runtime.Object {
	explicit := map[string]bool{}
	namespaces := map[string]bool{}
	for _, obj := range objs {
		if ns, ok := obj.(*corev1.Namespace); ok {
			explicit[ns.Name] = true
			continue
		}
		co, ok := obj.(client.Object)
		if !ok {
			continue
		}
		if ns := co.GetNamespace(); ns != "" && ns != "kro-system" {
			namespaces[ns] = true
		}
	}
	for ns := range namespaces {
		if !explicit[ns] {
			objs = append(objs, defaultResourceNamespace(ns))
		}
	}
	return objs
}

func namespaceWithIdentity(name, accountID, region string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				util.GlobalConfigNamespaceAnnotation: "kro-system",
				util.OwnerAccountIDAnnotation:         accountID,
				util.DefaultRegionAnnotation:          region,
			},
		},
	}
}

func defaultResourceNamespace(name string) *corev1.Namespace {
	return namespaceWithIdentity(name, "111122223333", "ap-southeast-2")
}

func globalKropathConfig(tier v1alpha1.KropathConfigTier) *v1alpha1.KropathConfig {
	return &v1alpha1.KropathConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "KropathConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.KropathConfigName,
			Namespace: "kro-system",
		},
		Spec: v1alpha1.KropathConfigSpec{
			Mandatory: tier,
		},
	}
}

func globalKMSConfig(name string, mandatory, defaults cascade.KMSConfigSection) *v1alpha1.KMSConfig {
	return &v1alpha1.KMSConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "KMSConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kro-system",
		},
		Spec: v1alpha1.KMSConfigSpec{
			Mandatory: mandatory,
			Defaults:  defaults,
		},
	}
}

func localKMSConfig(namespace, name string, mandatory, defaults cascade.KMSConfigSection) *v1alpha1.KMSConfig {
	cfg := globalKMSConfig(name, mandatory, defaults)
	cfg.Namespace = namespace
	return cfg
}

func getKMSConfig(t *testing.T, c client.Client, namespace, name string) *v1alpha1.KMSConfig {
	t.Helper()
	cfg := &v1alpha1.KMSConfig{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("KMSConfig"))
	if err := c.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, cfg); err != nil {
		t.Fatalf("get KMSConfig %s/%s: %v", namespace, name, err)
	}
	return cfg
}

func findCondition(conditions []metav1.Condition, condType string) *metav1.Condition {
	for i := range conditions {
		if conditions[i].Type == condType {
			return &conditions[i]
		}
	}
	return nil
}

func req(namespace, name string) ctrl.Request {
	return ctrl.Request{NamespacedName: client.ObjectKey{Namespace: namespace, Name: name}}
}
