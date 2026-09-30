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

package sqsconfig

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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	ctrl "sigs.k8s.io/controller-runtime"
)

// AC-1: globalKropathConfig.mandatory.sqs.encryptionType="kms" propagates (level 1 wins).
func TestReconcileAC1GlobalKropathEncryptionTypeLevel1(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig(v1alpha1.KropathConfigTier{
			SQS: cascade.SQSKropathSection{EncryptionType: "kms"},
		}),
		localSQSConfig("payments-prod", "general-policy", cascade.SQSConfigSection{}, cascade.SQSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSQSConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.EncryptionType; got != "kms" {
		t.Fatalf("mandatory.encryptionType = %q, want kms", got)
	}
}

// AC-2: globalSQSConfig.mandatory.encryptionType="kms" wins when levels 1-2 are empty (level 3 wins).
func TestReconcileAC2GlobalSQSConfigEncryptionTypeLevel3(t *testing.T) {
	rec, _ := testReconciler(t,
		globalSQSConfig("general-policy", cascade.SQSConfigSection{EncryptionType: "kms"}, cascade.SQSConfigSection{}),
		localSQSConfig("payments-prod", "general-policy", cascade.SQSConfigSection{}, cascade.SQSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSQSConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.EncryptionType; got != "kms" {
		t.Fatalf("mandatory.encryptionType = %q, want kms", got)
	}
}

// AC-3: localSQSConfig.defaults.encryptionType="sqs-managed" propagates (level 6).
func TestReconcileAC3LocalSQSConfigDefaultsEncryptionType(t *testing.T) {
	rec, _ := testReconciler(t,
		localSQSConfig("payments-prod", "general-policy",
			cascade.SQSConfigSection{},
			cascade.SQSConfigSection{EncryptionType: "sqs-managed"},
		),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSQSConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Defaults.EncryptionType; got != "sqs-managed" {
		t.Fatalf("defaults.encryptionType = %q, want sqs-managed", got)
	}
	if got := updated.Status.EffectiveConfig.Mandatory.EncryptionType; got != "" {
		t.Fatalf("mandatory.encryptionType = %q, want empty", got)
	}
}

// AC-4: globalKropathConfig.mandatory.sqs.visibilityTimeout=120 propagates (level 1 wins).
func TestReconcileAC4GlobalKropathVisibilityTimeoutLevel1(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig(v1alpha1.KropathConfigTier{
			SQS: cascade.SQSKropathSection{VisibilityTimeout: 120},
		}),
		localSQSConfig("payments-prod", "general-policy", cascade.SQSConfigSection{}, cascade.SQSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSQSConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.VisibilityTimeout; got != 120 {
		t.Fatalf("mandatory.visibilityTimeout = %d, want 120", got)
	}
}

// AC-5: globalSQSConfig.defaults.visibilityTimeout=30 (level 7) wins over
// globalKropathConfig.defaults.sqs.visibilityTimeout=60 (level 9).
func TestReconcileAC5GlobalSQSConfigDefaultsWinsOverKropathDefaults(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfigWithDefaults(
			v1alpha1.KropathConfigTier{},
			v1alpha1.KropathConfigTier{SQS: cascade.SQSKropathSection{VisibilityTimeout: 60}},
		),
		globalSQSConfig("general-policy", cascade.SQSConfigSection{}, cascade.SQSConfigSection{VisibilityTimeout: 30}),
		localSQSConfig("payments-prod", "general-policy", cascade.SQSConfigSection{}, cascade.SQSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSQSConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Defaults.VisibilityTimeout; got != 30 {
		t.Fatalf("defaults.visibilityTimeout = %d, want 30 (level 7 beats level 9)", got)
	}
}

// AC-6: globalKropathConfig.mandatory.sqs.messageRetentionPeriod=604800 propagates.
func TestReconcileAC6GlobalKropathMessageRetentionPeriod(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig(v1alpha1.KropathConfigTier{
			SQS: cascade.SQSKropathSection{MessageRetentionPeriod: 604800},
		}),
		localSQSConfig("payments-prod", "general-policy", cascade.SQSConfigSection{}, cascade.SQSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSQSConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.MessageRetentionPeriod; got != 604800 {
		t.Fatalf("mandatory.messageRetentionPeriod = %d, want 604800", got)
	}
}

// AC-7: globalSQSConfig.mandatory.delaySeconds=60 propagates (level 3).
func TestReconcileAC7GlobalSQSConfigDelaySeconds(t *testing.T) {
	rec, _ := testReconciler(t,
		globalSQSConfig("general-policy", cascade.SQSConfigSection{DelaySeconds: 60}, cascade.SQSConfigSection{}),
		localSQSConfig("payments-prod", "general-policy", cascade.SQSConfigSection{}, cascade.SQSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSQSConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.DelaySeconds; got != 60 {
		t.Fatalf("mandatory.delaySeconds = %d, want 60", got)
	}
}

// AC-8: globalSQSConfig.defaults.maximumMessageSize=131072 propagates (level 7).
func TestReconcileAC8GlobalSQSConfigDefaultsMaximumMessageSize(t *testing.T) {
	rec, _ := testReconciler(t,
		globalSQSConfig("general-policy", cascade.SQSConfigSection{}, cascade.SQSConfigSection{MaximumMessageSize: 131072}),
		localSQSConfig("payments-prod", "general-policy", cascade.SQSConfigSection{}, cascade.SQSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSQSConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Defaults.MaximumMessageSize; got != 131072 {
		t.Fatalf("defaults.maximumMessageSize = %d, want 131072", got)
	}
}

// AC-9: KropathConfig.mandatory.tags and SQSConfig.mandatory.tags are union-merged.
func TestReconcileAC9TagUnionMerge(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig(v1alpha1.KropathConfigTier{
			Tags: map[string]string{"cost-centre": "infra", "shared-key": "from-global-kropath"},
		}),
		localSQSConfig("payments-prod", "general-policy",
			cascade.SQSConfigSection{Tags: map[string]string{"queue-type": "messaging", "shared-key": "from-local-sqsconfig"}},
			cascade.SQSConfigSection{},
		),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSQSConfig(t, rec.Client, "payments-prod", "general-policy")
	tags := updated.Status.EffectiveConfig.Mandatory.Tags
	if tags["cost-centre"] != "infra" {
		t.Fatalf("tags[cost-centre] = %q, want infra", tags["cost-centre"])
	}
	if tags["queue-type"] != "messaging" {
		t.Fatalf("tags[queue-type] = %q, want messaging", tags["queue-type"])
	}
	// Level 1 (global KropathConfig mandatory) wins over level 4 (local SQSConfig mandatory).
	if tags["shared-key"] != "from-global-kropath" {
		t.Fatalf("tags[shared-key] = %q, want from-global-kropath (level-1 wins)", tags["shared-key"])
	}
}

// AC-10: globalKropathConfig.mandatory.sqs.kmsMasterKeyId propagates.
func TestReconcileAC10GlobalKropathKmsMasterKeyId(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig(v1alpha1.KropathConfigTier{
			SQS: cascade.SQSKropathSection{KmsMasterKeyId: "arn:aws:kms:ap-southeast-2:123:key/org-key"},
		}),
		localSQSConfig("payments-prod", "general-policy", cascade.SQSConfigSection{}, cascade.SQSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSQSConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.KmsMasterKeyId; got != "arn:aws:kms:ap-southeast-2:123:key/org-key" {
		t.Fatalf("mandatory.kmsMasterKeyId = %q, want arn:aws:kms:ap-southeast-2:123:key/org-key", got)
	}
}

// AC-11: globalSQSConfig.defaults.namingTemplate="{namespace}-{name}" propagates (level 7).
// KropathConfig.sqs has no namingTemplate field.
func TestReconcileAC11GlobalSQSConfigDefaultsNamingTemplate(t *testing.T) {
	rec, _ := testReconciler(t,
		globalSQSConfig("general-policy",
			cascade.SQSConfigSection{},
			cascade.SQSConfigSection{NamingTemplate: "{namespace}-{name}"},
		),
		localSQSConfig("payments-prod", "general-policy", cascade.SQSConfigSection{}, cascade.SQSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSQSConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Defaults.NamingTemplate; got != "{namespace}-{name}" {
		t.Fatalf("defaults.namingTemplate = %q, want {namespace}-{name}", got)
	}
}

// AC-12: globalSQSConfig.mandatory.syncedLabels={data-class: internal} propagates (level 3).
func TestReconcileAC12GlobalSQSConfigSyncedLabels(t *testing.T) {
	rec, _ := testReconciler(t,
		globalSQSConfig("general-policy",
			cascade.SQSConfigSection{SyncedLabels: map[string]string{"data-class": "internal"}},
			cascade.SQSConfigSection{},
		),
		localSQSConfig("payments-prod", "general-policy", cascade.SQSConfigSection{}, cascade.SQSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSQSConfig(t, rec.Client, "payments-prod", "general-policy")
	labels := updated.Status.EffectiveConfig.Mandatory.SyncedLabels
	if labels["data-class"] != "internal" {
		t.Fatalf("mandatory.syncedLabels[data-class] = %q, want internal", labels["data-class"])
	}
}

// AC-13: Provider identity from globalKropathConfig propagates to effCfg.aws.*.
func TestReconcileAC13ProviderIdentityPropagates(t *testing.T) {
	rec, _ := testReconciler(t,
		namespaceWithIdentity("payments-prod", "123456789012", "ap-southeast-2"),
		localSQSConfig("payments-prod", "general-policy", cascade.SQSConfigSection{}, cascade.SQSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSQSConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.AWS.AccountID; got != "123456789012" {
		t.Fatalf("aws.accountId = %q, want 123456789012", got)
	}
	if got := updated.Status.EffectiveConfig.AWS.Region; got != "ap-southeast-2" {
		t.Fatalf("aws.region = %q, want ap-southeast-2", got)
	}
}

func TestRequestsForKropathConfigChangeGlobal(t *testing.T) {
	rec, _ := testReconciler(t,
		localSQSConfig("payments-prod", "general-policy", cascade.SQSConfigSection{}, cascade.SQSConfigSection{}),
		localSQSConfig("sandbox", "general-policy", cascade.SQSConfigSection{}, cascade.SQSConfigSection{}),
		localSQSConfig("payments-prod", "other-policy", cascade.SQSConfigSection{}, cascade.SQSConfigSection{}),
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
		localSQSConfig("payments-prod", "general-policy", cascade.SQSConfigSection{}, cascade.SQSConfigSection{}),
		localSQSConfig("sandbox", "general-policy", cascade.SQSConfigSection{}, cascade.SQSConfigSection{}),
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
		localSQSConfig("payments-prod", "general-policy", cascade.SQSConfigSection{}, cascade.SQSConfigSection{}),
		localSQSConfig("payments-prod", "other-policy", cascade.SQSConfigSection{}, cascade.SQSConfigSection{}),
		localSQSConfig("sandbox", "general-policy", cascade.SQSConfigSection{}, cascade.SQSConfigSection{}),
	)

	got := rec.requestsForKropathConfigChange(context.Background(), &v1alpha1.KropathConfig{
		ObjectMeta: metav1.ObjectMeta{Name: util.KropathConfigName, Namespace: "payments-prod"},
	})

	if len(got) != 2 {
		t.Fatalf("requests len = %d, want 2 — local KPC triggers all configs in its namespace (%#v)", len(got), got)
	}
}

func TestRequestsForSQSConfigChangeGlobal(t *testing.T) {
	rec, _ := testReconciler(t,
		localSQSConfig("payments-prod", "general-policy", cascade.SQSConfigSection{}, cascade.SQSConfigSection{}),
		localSQSConfig("data-prod", "general-policy", cascade.SQSConfigSection{}, cascade.SQSConfigSection{}),
	)

	got := rec.requestsForSQSConfigChange(context.Background(), &v1alpha1.SQSConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "general-policy", Namespace: "kro-system"},
	})

	if len(got) != 2 {
		t.Fatalf("requests len = %d, want 2 (%#v)", len(got), got)
	}
}

func TestRequestsForSQSConfigChangeNonGlobalIgnored(t *testing.T) {
	rec, _ := testReconciler(t,
		localSQSConfig("payments-prod", "general-policy", cascade.SQSConfigSection{}, cascade.SQSConfigSection{}),
	)

	got := rec.requestsForSQSConfigChange(context.Background(), &v1alpha1.SQSConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "general-policy", Namespace: "payments-prod"},
	})

	if len(got) != 0 {
		t.Fatalf("requests len = %d, want 0 (%#v)", len(got), got)
	}
}

// ─── Test fixtures ──────────────────────────────────────────────────────────

func testReconciler(t *testing.T, objs ...runtime.Object) (*Reconciler, *v1alpha1.SQSConfig) {
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
		WithStatusSubresource(&v1alpha1.SQSConfig{})
	for _, obj := range objs {
		builder = builder.WithRuntimeObjects(obj)
	}
	cl := builder.Build()
	cfg := &v1alpha1.SQSConfig{}
	if err := cl.Get(context.Background(), client.ObjectKey{Namespace: "payments-prod", Name: "general-policy"}, cfg); err != nil {
		t.Fatalf("seed local SQSConfig: %v", err)
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

func globalKropathConfigWithDefaults(mandatory, defaults v1alpha1.KropathConfigTier) *v1alpha1.KropathConfig {
	return &v1alpha1.KropathConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "KropathConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.KropathConfigName,
			Namespace: "kro-system",
		},
		Spec: v1alpha1.KropathConfigSpec{
			Mandatory: mandatory,
			Defaults:  defaults,
		},
	}
}

func globalSQSConfig(name string, mandatory, defaults cascade.SQSConfigSection) *v1alpha1.SQSConfig {
	return &v1alpha1.SQSConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "SQSConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kro-system",
		},
		Spec: v1alpha1.SQSConfigSpec{
			Mandatory: mandatory,
			Defaults:  defaults,
		},
	}
}

func localSQSConfig(namespace, name string, mandatory, defaults cascade.SQSConfigSection) *v1alpha1.SQSConfig {
	cfg := globalSQSConfig(name, mandatory, defaults)
	cfg.Namespace = namespace
	return cfg
}

func getSQSConfig(t *testing.T, c client.Client, namespace, name string) *v1alpha1.SQSConfig {
	t.Helper()
	cfg := &v1alpha1.SQSConfig{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("SQSConfig"))
	if err := c.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, cfg); err != nil {
		t.Fatalf("get SQSConfig %s/%s: %v", namespace, name, err)
	}
	return cfg
}

func req(namespace, name string) ctrl.Request {
	return ctrl.Request{NamespacedName: client.ObjectKey{Namespace: namespace, Name: name}}
}
