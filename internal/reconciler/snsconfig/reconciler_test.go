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

package snsconfig

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

// AC-1: globalKropathConfig.mandatory.sns.kmsMasterKeyId propagates (level 1 wins).
func TestReconcileAC1GlobalKropathKmsMasterKeyIdLevel1(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig(v1alpha1.KropathConfigTier{
			SNS: cascade.SNSKropathSection{KmsMasterKeyId: "arn:aws:kms:ap-southeast-2:123456789012:key/org-key"},
		}),
		localSNSConfig("events-prod", "general-policy", cascade.SNSConfigSection{}, cascade.SNSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("events-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSNSConfig(t, rec.Client, "events-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.KmsMasterKeyId; got != "arn:aws:kms:ap-southeast-2:123456789012:key/org-key" {
		t.Fatalf("mandatory.kmsMasterKeyId = %q, want arn:aws:kms:ap-southeast-2:123456789012:key/org-key", got)
	}
}

// AC-2: globalSNSConfig.mandatory.kmsMasterKeyId wins when levels 1-2 are empty (level 3 wins).
func TestReconcileAC2GlobalSNSConfigKmsMasterKeyIdLevel3(t *testing.T) {
	rec, _ := testReconciler(t,
		globalSNSConfig("general-policy",
			cascade.SNSConfigSection{KmsMasterKeyId: "arn:aws:kms:ap-southeast-2:123456789012:key/org-key"},
			cascade.SNSConfigSection{},
		),
		localSNSConfig("events-prod", "general-policy", cascade.SNSConfigSection{}, cascade.SNSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("events-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSNSConfig(t, rec.Client, "events-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.KmsMasterKeyId; got != "arn:aws:kms:ap-southeast-2:123456789012:key/org-key" {
		t.Fatalf("mandatory.kmsMasterKeyId = %q, want arn:aws:kms:ap-southeast-2:123456789012:key/org-key", got)
	}
}

// AC-3: localSNSConfig.defaults.kmsMasterKeyId propagates (level 6); mandatory stays empty.
func TestReconcileAC3LocalSNSConfigDefaultsKmsMasterKeyId(t *testing.T) {
	rec, _ := testReconciler(t,
		localSNSConfig("events-prod", "general-policy",
			cascade.SNSConfigSection{},
			cascade.SNSConfigSection{KmsMasterKeyId: "alias/aws/sns"},
		),
	)

	if _, err := rec.Reconcile(context.Background(), req("events-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSNSConfig(t, rec.Client, "events-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Defaults.KmsMasterKeyId; got != "alias/aws/sns" {
		t.Fatalf("defaults.kmsMasterKeyId = %q, want alias/aws/sns", got)
	}
	if got := updated.Status.EffectiveConfig.Mandatory.KmsMasterKeyId; got != "" {
		t.Fatalf("mandatory.kmsMasterKeyId = %q, want empty", got)
	}
}

// AC-4: globalKropathConfig.mandatory.sns.signatureVersion="2" propagates (level 1 wins).
func TestReconcileAC4GlobalKropathSignatureVersionLevel1(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig(v1alpha1.KropathConfigTier{
			SNS: cascade.SNSKropathSection{SignatureVersion: "2"},
		}),
		localSNSConfig("events-prod", "general-policy", cascade.SNSConfigSection{}, cascade.SNSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("events-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSNSConfig(t, rec.Client, "events-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.SignatureVersion; got != "2" {
		t.Fatalf("mandatory.signatureVersion = %q, want 2", got)
	}
}

// AC-5: globalSNSConfig.defaults.signatureVersion="2" (level 7) wins over
// globalKropathConfig.defaults.sns.signatureVersion="1" (level 9).
func TestReconcileAC5GlobalSNSConfigDefaultsWinsOverKropathDefaults(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfigWithDefaults(
			v1alpha1.KropathConfigTier{},
			v1alpha1.KropathConfigTier{SNS: cascade.SNSKropathSection{SignatureVersion: "1"}},
		),
		globalSNSConfig("general-policy",
			cascade.SNSConfigSection{},
			cascade.SNSConfigSection{SignatureVersion: "2"},
		),
		localSNSConfig("events-prod", "general-policy", cascade.SNSConfigSection{}, cascade.SNSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("events-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSNSConfig(t, rec.Client, "events-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Defaults.SignatureVersion; got != "2" {
		t.Fatalf("defaults.signatureVersion = %q, want 2 (level 7 beats level 9)", got)
	}
}

// AC-6: globalKropathConfig.mandatory.sns.tracingConfig="Active" propagates (level 1 wins).
func TestReconcileAC6GlobalKropathTracingConfigLevel1(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig(v1alpha1.KropathConfigTier{
			SNS: cascade.SNSKropathSection{TracingConfig: "Active"},
		}),
		localSNSConfig("events-prod", "general-policy", cascade.SNSConfigSection{}, cascade.SNSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("events-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSNSConfig(t, rec.Client, "events-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.TracingConfig; got != "Active" {
		t.Fatalf("mandatory.tracingConfig = %q, want Active", got)
	}
}

// AC-7: globalSNSConfig.mandatory.dataProtectionPolicy propagates (level 3).
func TestReconcileAC7GlobalSNSConfigDataProtectionPolicyLevel3(t *testing.T) {
	policy := `{"Name":"org-policy","Statements":[]}`
	rec, _ := testReconciler(t,
		globalSNSConfig("general-policy",
			cascade.SNSConfigSection{DataProtectionPolicy: policy},
			cascade.SNSConfigSection{},
		),
		localSNSConfig("events-prod", "general-policy", cascade.SNSConfigSection{}, cascade.SNSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("events-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSNSConfig(t, rec.Client, "events-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.DataProtectionPolicy; got != policy {
		t.Fatalf("mandatory.dataProtectionPolicy = %q, want %q", got, policy)
	}
}

// AC-8: globalSNSConfig.mandatory.deliveryFeedback.http all three fields propagate (level 3).
func TestReconcileAC8GlobalSNSConfigDeliveryFeedbackHTTPLevel3(t *testing.T) {
	http := &cascade.DeliveryFeedbackProtocol{
		SuccessFeedbackRoleArn:    "arn:aws:iam::123456789012:role/http-success",
		FailureFeedbackRoleArn:    "arn:aws:iam::123456789012:role/http-failure",
		SuccessFeedbackSampleRate: "100",
	}
	rec, _ := testReconciler(t,
		globalSNSConfig("general-policy",
			cascade.SNSConfigSection{DeliveryFeedback: cascade.DeliveryFeedback{HTTP: http}},
			cascade.SNSConfigSection{},
		),
		localSNSConfig("events-prod", "general-policy", cascade.SNSConfigSection{}, cascade.SNSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("events-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSNSConfig(t, rec.Client, "events-prod", "general-policy")
	got := updated.Status.EffectiveConfig.Mandatory.DeliveryFeedback
	if got == nil || got.HTTP == nil {
		t.Fatal("mandatory.deliveryFeedback.http is nil, want populated")
	}
	if got.HTTP.SuccessFeedbackRoleArn != http.SuccessFeedbackRoleArn {
		t.Fatalf("successFeedbackRoleArn = %q, want %q", got.HTTP.SuccessFeedbackRoleArn, http.SuccessFeedbackRoleArn)
	}
	if got.HTTP.FailureFeedbackRoleArn != http.FailureFeedbackRoleArn {
		t.Fatalf("failureFeedbackRoleArn = %q, want %q", got.HTTP.FailureFeedbackRoleArn, http.FailureFeedbackRoleArn)
	}
	if got.HTTP.SuccessFeedbackSampleRate != http.SuccessFeedbackSampleRate {
		t.Fatalf("successFeedbackSampleRate = %q, want %q", got.HTTP.SuccessFeedbackSampleRate, http.SuccessFeedbackSampleRate)
	}
}

// AC-9: globalSNSConfig.defaults.deliveryFeedback.sqs propagates (level 7).
func TestReconcileAC9GlobalSNSConfigDefaultsDeliveryFeedbackSQSLevel7(t *testing.T) {
	sqs := &cascade.DeliveryFeedbackProtocol{
		SuccessFeedbackRoleArn:    "arn:aws:iam::123456789012:role/sqs-success-default",
		SuccessFeedbackSampleRate: "50",
	}
	rec, _ := testReconciler(t,
		globalSNSConfig("general-policy",
			cascade.SNSConfigSection{},
			cascade.SNSConfigSection{DeliveryFeedback: cascade.DeliveryFeedback{SQS: sqs}},
		),
		localSNSConfig("events-prod", "general-policy", cascade.SNSConfigSection{}, cascade.SNSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("events-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSNSConfig(t, rec.Client, "events-prod", "general-policy")
	got := updated.Status.EffectiveConfig.Defaults.DeliveryFeedback
	if got == nil || got.SQS == nil {
		t.Fatal("defaults.deliveryFeedback.sqs is nil, want populated")
	}
	if got.SQS.SuccessFeedbackRoleArn != sqs.SuccessFeedbackRoleArn {
		t.Fatalf("sqs.successFeedbackRoleArn = %q, want %q", got.SQS.SuccessFeedbackRoleArn, sqs.SuccessFeedbackRoleArn)
	}
	if got.SQS.SuccessFeedbackSampleRate != sqs.SuccessFeedbackSampleRate {
		t.Fatalf("sqs.successFeedbackSampleRate = %q, want %q", got.SQS.SuccessFeedbackSampleRate, sqs.SuccessFeedbackSampleRate)
	}
}

// AC-10: globalSNSConfig.defaults.namingTemplate="{namespace}-{name}" propagates (level 7).
func TestReconcileAC10GlobalSNSConfigDefaultsNamingTemplateLevel7(t *testing.T) {
	rec, _ := testReconciler(t,
		globalSNSConfig("general-policy",
			cascade.SNSConfigSection{},
			cascade.SNSConfigSection{NamingTemplate: "{namespace}-{name}"},
		),
		localSNSConfig("events-prod", "general-policy", cascade.SNSConfigSection{}, cascade.SNSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("events-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSNSConfig(t, rec.Client, "events-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Defaults.NamingTemplate; got != "{namespace}-{name}" {
		t.Fatalf("defaults.namingTemplate = %q, want {namespace}-{name}", got)
	}
}

// AC-11: KropathConfig.mandatory.tags and SNSConfig.mandatory.tags are union-merged.
func TestReconcileAC11TagUnionMerge(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig(v1alpha1.KropathConfigTier{
			Tags: map[string]string{"cost-centre": "infra", "shared-key": "from-global-kropath"},
		}),
		localSNSConfig("events-prod", "general-policy",
			cascade.SNSConfigSection{Tags: map[string]string{"topic-type": "events", "shared-key": "from-local-snsconfig"}},
			cascade.SNSConfigSection{},
		),
	)

	if _, err := rec.Reconcile(context.Background(), req("events-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSNSConfig(t, rec.Client, "events-prod", "general-policy")
	tags := updated.Status.EffectiveConfig.Mandatory.Tags
	if tags["cost-centre"] != "infra" {
		t.Fatalf("tags[cost-centre] = %q, want infra", tags["cost-centre"])
	}
	if tags["topic-type"] != "events" {
		t.Fatalf("tags[topic-type] = %q, want events", tags["topic-type"])
	}
	// Level 1 (global KropathConfig mandatory) wins over level 4 (local SNSConfig mandatory).
	if tags["shared-key"] != "from-global-kropath" {
		t.Fatalf("tags[shared-key] = %q, want from-global-kropath (level-1 wins)", tags["shared-key"])
	}
}

// AC-12: globalSNSConfig.mandatory.syncedLabels={data-class: internal} propagates (level 3).
func TestReconcileAC12GlobalSNSConfigSyncedLabels(t *testing.T) {
	rec, _ := testReconciler(t,
		globalSNSConfig("general-policy",
			cascade.SNSConfigSection{SyncedLabels: map[string]string{"data-class": "internal"}},
			cascade.SNSConfigSection{},
		),
		localSNSConfig("events-prod", "general-policy", cascade.SNSConfigSection{}, cascade.SNSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("events-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSNSConfig(t, rec.Client, "events-prod", "general-policy")
	labels := updated.Status.EffectiveConfig.Mandatory.SyncedLabels
	if labels["data-class"] != "internal" {
		t.Fatalf("mandatory.syncedLabels[data-class] = %q, want internal", labels["data-class"])
	}
}

// AC-13: Provider identity from globalKropathConfig propagates to effCfg.aws.*.
func TestReconcileAC13ProviderIdentityPropagates(t *testing.T) {
	rec, _ := testReconciler(t,
		namespaceWithIdentity("events-prod", "123456789012", "ap-southeast-2"),
		localSNSConfig("events-prod", "general-policy", cascade.SNSConfigSection{}, cascade.SNSConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("events-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getSNSConfig(t, rec.Client, "events-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.AWS.AccountID; got != "123456789012" {
		t.Fatalf("aws.accountId = %q, want 123456789012", got)
	}
	if got := updated.Status.EffectiveConfig.AWS.Region; got != "ap-southeast-2" {
		t.Fatalf("aws.region = %q, want ap-southeast-2", got)
	}
}

func TestRequestsForKropathConfigChangeGlobal(t *testing.T) {
	rec, _ := testReconciler(t,
		localSNSConfig("events-prod", "general-policy", cascade.SNSConfigSection{}, cascade.SNSConfigSection{}),
		localSNSConfig("sandbox", "general-policy", cascade.SNSConfigSection{}, cascade.SNSConfigSection{}),
		localSNSConfig("events-prod", "other-policy", cascade.SNSConfigSection{}, cascade.SNSConfigSection{}),
	)

	got := rec.requestsForKropathConfigChange(context.Background(), &v1alpha1.KropathConfig{
		ObjectMeta: metav1.ObjectMeta{Name: util.KropathConfigName, Namespace: "kro-system"},
	})

	want := map[string]bool{
		"events-prod/general-policy": false,
		"sandbox/general-policy":     false,
		"events-prod/other-policy":   false,
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
		localSNSConfig("events-prod", "general-policy", cascade.SNSConfigSection{}, cascade.SNSConfigSection{}),
		localSNSConfig("sandbox", "general-policy", cascade.SNSConfigSection{}, cascade.SNSConfigSection{}),
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
		localSNSConfig("events-prod", "general-policy", cascade.SNSConfigSection{}, cascade.SNSConfigSection{}),
		localSNSConfig("events-prod", "other-policy", cascade.SNSConfigSection{}, cascade.SNSConfigSection{}),
		localSNSConfig("sandbox", "general-policy", cascade.SNSConfigSection{}, cascade.SNSConfigSection{}),
	)

	got := rec.requestsForKropathConfigChange(context.Background(), &v1alpha1.KropathConfig{
		ObjectMeta: metav1.ObjectMeta{Name: util.KropathConfigName, Namespace: "events-prod"},
	})

	if len(got) != 2 {
		t.Fatalf("requests len = %d, want 2 — local KPC triggers all configs in its namespace (%#v)", len(got), got)
	}
}

func TestRequestsForSNSConfigChangeGlobal(t *testing.T) {
	rec, _ := testReconciler(t,
		localSNSConfig("events-prod", "general-policy", cascade.SNSConfigSection{}, cascade.SNSConfigSection{}),
		localSNSConfig("data-prod", "general-policy", cascade.SNSConfigSection{}, cascade.SNSConfigSection{}),
	)

	got := rec.requestsForSNSConfigChange(context.Background(), &v1alpha1.SNSConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "general-policy", Namespace: "kro-system"},
	})

	if len(got) != 2 {
		t.Fatalf("requests len = %d, want 2 (%#v)", len(got), got)
	}
}

func TestRequestsForSNSConfigChangeNonGlobalIgnored(t *testing.T) {
	rec, _ := testReconciler(t,
		localSNSConfig("events-prod", "general-policy", cascade.SNSConfigSection{}, cascade.SNSConfigSection{}),
	)

	got := rec.requestsForSNSConfigChange(context.Background(), &v1alpha1.SNSConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "general-policy", Namespace: "events-prod"},
	})

	if len(got) != 0 {
		t.Fatalf("requests len = %d, want 0 (%#v)", len(got), got)
	}
}

// ─── Test fixtures ──────────────────────────────────────────────────────────

func testReconciler(t *testing.T, objs ...runtime.Object) (*Reconciler, *v1alpha1.SNSConfig) {
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
		WithStatusSubresource(&v1alpha1.SNSConfig{})
	for _, obj := range objs {
		builder = builder.WithRuntimeObjects(obj)
	}
	cl := builder.Build()
	cfg := &v1alpha1.SNSConfig{}
	if err := cl.Get(context.Background(), client.ObjectKey{Namespace: "events-prod", Name: "general-policy"}, cfg); err != nil {
		t.Fatalf("seed local SNSConfig: %v", err)
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

func globalSNSConfig(name string, mandatory, defaults cascade.SNSConfigSection) *v1alpha1.SNSConfig {
	return &v1alpha1.SNSConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "SNSConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kro-system",
		},
		Spec: v1alpha1.SNSConfigSpec{
			Mandatory: mandatory,
			Defaults:  defaults,
		},
	}
}

func localSNSConfig(namespace, name string, mandatory, defaults cascade.SNSConfigSection) *v1alpha1.SNSConfig {
	cfg := globalSNSConfig(name, mandatory, defaults)
	cfg.Namespace = namespace
	return cfg
}

func getSNSConfig(t *testing.T, c client.Client, namespace, name string) *v1alpha1.SNSConfig {
	t.Helper()
	cfg := &v1alpha1.SNSConfig{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("SNSConfig"))
	if err := c.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, cfg); err != nil {
		t.Fatalf("get SNSConfig %s/%s: %v", namespace, name, err)
	}
	return cfg
}

func req(namespace, name string) ctrl.Request {
	return ctrl.Request{NamespacedName: client.ObjectKey{Namespace: namespace, Name: name}}
}
