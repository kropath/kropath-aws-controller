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

package apigatewayconfig

import (
	"context"
	"reflect"
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

const globalNS = "platform-config"

// AC-1: globalKropathConfig.mandatory.apigateway.endpointType="REGIONAL" propagates (level 1 wins).
func TestReconcileAC1GlobalKropathEndpointTypeLevel1(t *testing.T) {
	rec, _ := testReconciler(t,
		resourceNamespace("payments-prod"),
		globalKropathConfig(v1alpha1.KropathConfigTier{
			ApiGateway: cascade.ApiGatewayKropathSection{EndpointType: "REGIONAL"},
		}),
		localApiGatewayConfig("payments-prod", "general-policy", cascade.ApiGatewayConfigSection{}, cascade.ApiGatewayConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getApiGatewayConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.EndpointType; got != "REGIONAL" {
		t.Fatalf("mandatory.endpointType = %q, want REGIONAL", got)
	}
}

// AC-2: globalApiGatewayConfig.mandatory.apiKeySource="HEADER" propagates (level 3 wins when L1-L2 absent).
func TestReconcileAC2GlobalApiGatewayConfigApiKeySourceLevel3(t *testing.T) {
	rec, _ := testReconciler(t,
		resourceNamespace("payments-prod"),
		globalApiGatewayConfig("general-policy", cascade.ApiGatewayConfigSection{ApiKeySource: "HEADER"}, cascade.ApiGatewayConfigSection{}),
		localApiGatewayConfig("payments-prod", "general-policy", cascade.ApiGatewayConfigSection{}, cascade.ApiGatewayConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getApiGatewayConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.ApiKeySource; got != "HEADER" {
		t.Fatalf("mandatory.apiKeySource = %q, want HEADER", got)
	}
}

// AC-3: localApiGatewayConfig.defaults.namingTemplate="{namespace}-{name}" propagates (level 6).
func TestReconcileAC3LocalApiGatewayConfigDefaultsNamingTemplate(t *testing.T) {
	rec, _ := testReconciler(t,
		resourceNamespace("payments-prod"),
		localApiGatewayConfig("payments-prod", "general-policy",
			cascade.ApiGatewayConfigSection{},
			cascade.ApiGatewayConfigSection{NamingTemplate: "{namespace}-{name}"},
		),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getApiGatewayConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Defaults.NamingTemplate; got != "{namespace}-{name}" {
		t.Fatalf("defaults.namingTemplate = %q, want {namespace}-{name}", got)
	}
	if got := updated.Status.EffectiveConfig.Mandatory.NamingTemplate; got != "" {
		t.Fatalf("mandatory.namingTemplate = %q, want empty", got)
	}
}

// AC-4: globalKropathConfig.mandatory.tags augmented into KropathSection tags cascade.
func TestReconcileAC4GlobalKropathTagsAugmented(t *testing.T) {
	rec, _ := testReconciler(t,
		resourceNamespace("payments-prod"),
		globalKropathConfig(v1alpha1.KropathConfigTier{
			Tags: map[string]string{"cost-centre": "infra"},
		}),
		localApiGatewayConfig("payments-prod", "general-policy",
			cascade.ApiGatewayConfigSection{Tags: map[string]string{"service": "api"}},
			cascade.ApiGatewayConfigSection{},
		),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getApiGatewayConfig(t, rec.Client, "payments-prod", "general-policy")
	tags := updated.Status.EffectiveConfig.Mandatory.Tags
	if tags["cost-centre"] != "infra" {
		t.Fatalf("tags[cost-centre] = %q, want infra (from global KropathConfig)", tags["cost-centre"])
	}
	if tags["service"] != "api" {
		t.Fatalf("tags[service] = %q, want api (from local ApiGatewayConfig mandatory)", tags["service"])
	}
}

// AC-5: account/region identity resolves from the namespace annotations (ADR-019 D-3
// — KropathConfig.spec.aws no longer exists as a source), not from any config CR.
func TestReconcileAC5ProviderIdentityResolvesFromNamespace(t *testing.T) {
	rec, _ := testReconciler(t,
		namespaceWithIdentity("payments-prod", "123456789012", "ap-southeast-2"),
		localApiGatewayConfig("payments-prod", "general-policy", cascade.ApiGatewayConfigSection{}, cascade.ApiGatewayConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getApiGatewayConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.AWS.AccountID; got != "123456789012" {
		t.Fatalf("aws.accountId = %q, want 123456789012", got)
	}
	if got := updated.Status.EffectiveConfig.AWS.Region; got != "ap-southeast-2" {
		t.Fatalf("aws.region = %q, want ap-southeast-2", got)
	}
	if got := updated.Status.EffectiveConfig.AWS.Partition; got != "aws" {
		t.Fatalf("aws.partition = %q, want aws", got)
	}
}

// AC-6: disableExecuteApiEndpoint enforced from global ApiGatewayConfig mandatory (level 3).
func TestReconcileAC6DisableExecuteApiEndpointLevel3(t *testing.T) {
	rec, _ := testReconciler(t,
		resourceNamespace("payments-prod"),
		globalApiGatewayConfig("general-policy",
			cascade.ApiGatewayConfigSection{DisableExecuteApiEndpoint: true},
			cascade.ApiGatewayConfigSection{},
		),
		localApiGatewayConfig("payments-prod", "general-policy", cascade.ApiGatewayConfigSection{}, cascade.ApiGatewayConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getApiGatewayConfig(t, rec.Client, "payments-prod", "general-policy")
	if !updated.Status.EffectiveConfig.Mandatory.DisableExecuteApiEndpoint {
		t.Fatalf("mandatory.disableExecuteApiEndpoint = false, want true")
	}
}

// A resource namespace missing the account annotation gets no effectiveConfig and a
// Reconciled=False/MissingAccountAnnotation condition (spec §6.2, AC-3 shape).
func TestReconcileMissingAccountAnnotationWithholdsEffectiveConfig(t *testing.T) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "payments-prod",
			Annotations: map[string]string{util.GlobalConfigNamespaceAnnotation: globalNS, util.DefaultRegionAnnotation: "ap-southeast-2"},
		},
	}
	rec, _ := testReconciler(t, ns, localApiGatewayConfig("payments-prod", "general-policy", cascade.ApiGatewayConfigSection{}, cascade.ApiGatewayConfigSection{}))

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getApiGatewayConfig(t, rec.Client, "payments-prod", "general-policy")
	var zero v1alpha1.EffectiveAPIGatewayConfig
	if !reflect.DeepEqual(updated.Status.EffectiveConfig, zero) {
		t.Fatalf("EffectiveConfig = %+v, want zero value", updated.Status.EffectiveConfig)
	}
	found := false
	for _, c := range updated.Status.Conditions {
		if c.Type == "Reconciled" {
			found = true
			if c.Status != metav1.ConditionFalse || c.Reason != util.ReasonMissingAccountAnnotation {
				t.Fatalf("Reconciled condition = %+v, want False/%s", c, util.ReasonMissingAccountAnnotation)
			}
		}
	}
	if !found {
		t.Fatal("no Reconciled condition published")
	}
}

// A governance-only namespace (no global-config-namespace annotation) is exempt: its
// config objects receive no effectiveConfig (spec §6.4, AC-8 shape).
func TestReconcileGovernanceOnlyNamespaceExempt(t *testing.T) {
	rec, _ := testReconciler(t,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "platform-shared"}},
		localApiGatewayConfig("platform-shared", "general-policy", cascade.ApiGatewayConfigSection{}, cascade.ApiGatewayConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("platform-shared", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getApiGatewayConfig(t, rec.Client, "platform-shared", "general-policy")
	var zero v1alpha1.EffectiveAPIGatewayConfig
	if !reflect.DeepEqual(updated.Status.EffectiveConfig, zero) {
		t.Fatalf("EffectiveConfig = %+v, want zero value for governance-only namespace", updated.Status.EffectiveConfig)
	}
	for _, c := range updated.Status.Conditions {
		if c.Type == "Reconciled" && (c.Status != metav1.ConditionTrue || c.Reason != util.ReasonGlobalTierInput) {
			t.Fatalf("Reconciled condition = %+v, want True/%s", c, util.ReasonGlobalTierInput)
		}
	}
}

func TestRequestsForKropathConfigChangeGlobal(t *testing.T) {
	rec, _ := testReconciler(t,
		resourceNamespace("payments-prod"),
		resourceNamespace("sandbox"),
		localApiGatewayConfig("payments-prod", "general-policy", cascade.ApiGatewayConfigSection{}, cascade.ApiGatewayConfigSection{}),
		localApiGatewayConfig("sandbox", "general-policy", cascade.ApiGatewayConfigSection{}, cascade.ApiGatewayConfigSection{}),
		localApiGatewayConfig("payments-prod", "other-policy", cascade.ApiGatewayConfigSection{}, cascade.ApiGatewayConfigSection{}),
	)

	got := rec.requestsForKropathConfigChange(context.Background(), &v1alpha1.KropathConfig{
		ObjectMeta: metav1.ObjectMeta{Name: util.KropathConfigName, Namespace: globalNS},
	})

	// Classified by namespace, not name (ADR-018 D-1): a global-tier change
	// enqueues every config whose resolved global namespace matches.
	if len(got) != 3 {
		t.Fatalf("requests len = %d, want 3 (%#v)", len(got), got)
	}
}

func TestRequestsForKropathConfigChangeLocalNamespaceEnqueuesAllConfigsInNamespace(t *testing.T) {
	rec, _ := testReconciler(t,
		resourceNamespace("payments-prod"),
		resourceNamespace("sandbox"),
		localApiGatewayConfig("payments-prod", "general-policy", cascade.ApiGatewayConfigSection{}, cascade.ApiGatewayConfigSection{}),
		localApiGatewayConfig("payments-prod", "other-policy", cascade.ApiGatewayConfigSection{}, cascade.ApiGatewayConfigSection{}),
		localApiGatewayConfig("sandbox", "general-policy", cascade.ApiGatewayConfigSection{}, cascade.ApiGatewayConfigSection{}),
	)

	got := rec.requestsForKropathConfigChange(context.Background(), &v1alpha1.KropathConfig{
		ObjectMeta: metav1.ObjectMeta{Name: util.KropathConfigName, Namespace: "payments-prod"},
	})

	if len(got) != 2 {
		t.Fatalf("requests len = %d, want 2 — local KPC triggers all configs in its namespace (%#v)", len(got), got)
	}
}

func TestRequestsForApiGatewayConfigChangeNonGlobalIgnored(t *testing.T) {
	rec, _ := testReconciler(t,
		resourceNamespace("payments-prod"),
		localApiGatewayConfig("payments-prod", "general-policy", cascade.ApiGatewayConfigSection{}, cascade.ApiGatewayConfigSection{}),
	)

	got := rec.requestsForApiGatewayConfigChange(context.Background(), &v1alpha1.APIGatewayConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "general-policy", Namespace: "payments-prod"},
	})

	if len(got) != 0 {
		t.Fatalf("requests len = %d, want 0 (%#v)", len(got), got)
	}
}

// ─── Test fixtures ──────────────────────────────────────────────────────────

func testReconciler(t *testing.T, objs ...runtime.Object) (*Reconciler, *v1alpha1.APIGatewayConfig) {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("add corev1 scheme: %v", err)
	}
	builder := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&v1alpha1.APIGatewayConfig{})
	for _, obj := range objs {
		builder = builder.WithRuntimeObjects(obj)
	}
	cl := builder.Build()
	cfg := &v1alpha1.APIGatewayConfig{}
	_ = cl.Get(context.Background(), client.ObjectKey{Namespace: "payments-prod", Name: "general-policy"}, cfg)
	return &Reconciler{Client: cl, Log: logr.Discard(), Scheme: scheme}, cfg
}

// resourceNamespace returns a namespace annotated as a resource namespace with a
// valid account/region, resolving its global tier to globalNS.
func resourceNamespace(name string) *corev1.Namespace {
	return namespaceWithIdentity(name, "111122223333", "ap-southeast-2")
}

func namespaceWithIdentity(name, accountID, region string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				util.GlobalConfigNamespaceAnnotation: globalNS,
				util.OwnerAccountIDAnnotation:        accountID,
				util.DefaultRegionAnnotation:         region,
			},
		},
	}
}

func globalKropathConfig(tier v1alpha1.KropathConfigTier) *v1alpha1.KropathConfig {
	return &v1alpha1.KropathConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "KropathConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.KropathConfigName,
			Namespace: globalNS,
		},
		Spec: v1alpha1.KropathConfigSpec{
			Mandatory: tier,
		},
	}
}

func globalApiGatewayConfig(name string, mandatory, defaults cascade.ApiGatewayConfigSection) *v1alpha1.APIGatewayConfig {
	return &v1alpha1.APIGatewayConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "APIGatewayConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: globalNS,
		},
		Spec: v1alpha1.APIGatewayConfigSpec{
			Mandatory: mandatory,
			Defaults:  defaults,
		},
	}
}

func localApiGatewayConfig(namespace, name string, mandatory, defaults cascade.ApiGatewayConfigSection) *v1alpha1.APIGatewayConfig {
	cfg := globalApiGatewayConfig(name, mandatory, defaults)
	cfg.Namespace = namespace
	return cfg
}

func getApiGatewayConfig(t *testing.T, c client.Client, namespace, name string) *v1alpha1.APIGatewayConfig {
	t.Helper()
	cfg := &v1alpha1.APIGatewayConfig{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("APIGatewayConfig"))
	if err := c.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, cfg); err != nil {
		t.Fatalf("get ApiGatewayConfig %s/%s: %v", namespace, name, err)
	}
	return cfg
}

func req(namespace, name string) ctrl.Request {
	return ctrl.Request{NamespacedName: client.ObjectKey{Namespace: namespace, Name: name}}
}
