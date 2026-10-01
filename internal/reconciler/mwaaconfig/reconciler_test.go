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

package mwaaconfig

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

// AC-1: globalKropathConfig.mandatory.mwaa.webserverAccessMode="PRIVATE_ONLY" propagates (level 1 wins).
func TestReconcileAC1GlobalKropathWebserverAccessModeLevel1(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig(v1alpha1.KropathConfigTier{
			MWAA: cascade.MWAAKropathSection{WebserverAccessMode: "PRIVATE_ONLY"},
		}),
		localMWAAConfig("payments-prod", "general-policy", cascade.MWAAConfigSection{}, cascade.MWAAConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getMWAAConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.WebserverAccessMode; got != "PRIVATE_ONLY" {
		t.Fatalf("mandatory.webserverAccessMode = %q, want PRIVATE_ONLY", got)
	}
}

// AC-1b: globalMWAAConfig.defaults.environmentClass="mw1.small" propagates (level 7).
func TestReconcileAC1bGlobalMWAAConfigDefaultsEnvironmentClass(t *testing.T) {
	rec, _ := testReconciler(t,
		globalMWAAConfig("general-policy", cascade.MWAAConfigSection{}, cascade.MWAAConfigSection{EnvironmentClass: "mw1.small"}),
		localMWAAConfig("payments-prod", "general-policy", cascade.MWAAConfigSection{}, cascade.MWAAConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getMWAAConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Defaults.EnvironmentClass; got != "mw1.small" {
		t.Fatalf("defaults.environmentClass = %q, want mw1.small", got)
	}
}

// AC-1c: localMWAAConfig.defaults.airflowVersion="2.10.3" propagates (level 6 strongest defaults).
func TestReconcileAC1cLocalMWAAConfigDefaultsAirflowVersion(t *testing.T) {
	rec, _ := testReconciler(t,
		globalMWAAConfig("general-policy", cascade.MWAAConfigSection{}, cascade.MWAAConfigSection{AirflowVersion: "2.9.0"}),
		localMWAAConfig("payments-prod", "general-policy", cascade.MWAAConfigSection{}, cascade.MWAAConfigSection{AirflowVersion: "2.10.3"}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getMWAAConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Defaults.AirflowVersion; got != "2.10.3" {
		t.Fatalf("defaults.airflowVersion = %q, want 2.10.3 (level 6 beats level 7)", got)
	}
}

// AC-5: globalKropathConfig.mandatory.mwaa.maxWorkers=5 propagates (level 1 wins).
func TestReconcileAC5GlobalKropathMaxWorkersLevel1(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig(v1alpha1.KropathConfigTier{
			MWAA: cascade.MWAAKropathSection{MaxWorkers: 5},
		}),
		localMWAAConfig("payments-prod", "general-policy", cascade.MWAAConfigSection{}, cascade.MWAAConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getMWAAConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Mandatory.MaxWorkers; got != 5 {
		t.Fatalf("mandatory.maxWorkers = %d, want 5", got)
	}
}

// AC-6: globalKropathConfig.mandatory.mwaa.dagProcessingLogsEnabled=true propagates (*bool, level 1 wins).
func TestReconcileAC6GlobalKropathDagProcessingLogsEnabledLevel1(t *testing.T) {
	trueVal := true
	rec, _ := testReconciler(t,
		globalKropathConfig(v1alpha1.KropathConfigTier{
			MWAA: cascade.MWAAKropathSection{DagProcessingLogsEnabled: &trueVal},
		}),
		localMWAAConfig("payments-prod", "general-policy", cascade.MWAAConfigSection{}, cascade.MWAAConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getMWAAConfig(t, rec.Client, "payments-prod", "general-policy")
	if updated.Status.EffectiveConfig.Mandatory.DagProcessingLogsEnabled == nil {
		t.Fatal("mandatory.dagProcessingLogsEnabled = nil, want &true")
	}
	if got := *updated.Status.EffectiveConfig.Mandatory.DagProcessingLogsEnabled; !got {
		t.Fatalf("mandatory.dagProcessingLogsEnabled = %v, want true", got)
	}
}

// AC-6b: localMWAAConfig.defaults.dagProcessingLogsEnabled=false overrides global KropathConfig defaults (level 6 beats level 9).
func TestReconcileAC6bLocalDefaultsBoolPtrBeatsGlobalKropathDefaults(t *testing.T) {
	falseVal := false
	trueVal := true
	rec, _ := testReconciler(t,
		globalKropathConfigWithDefaults(
			v1alpha1.KropathConfigTier{},
			v1alpha1.KropathConfigTier{MWAA: cascade.MWAAKropathSection{DagProcessingLogsEnabled: &trueVal}},
		),
		localMWAAConfig("payments-prod", "general-policy",
			cascade.MWAAConfigSection{},
			cascade.MWAAConfigSection{DagProcessingLogsEnabled: &falseVal},
		),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getMWAAConfig(t, rec.Client, "payments-prod", "general-policy")
	if updated.Status.EffectiveConfig.Defaults.DagProcessingLogsEnabled == nil {
		t.Fatal("defaults.dagProcessingLogsEnabled = nil, want &false")
	}
	if got := *updated.Status.EffectiveConfig.Defaults.DagProcessingLogsEnabled; got {
		t.Fatalf("defaults.dagProcessingLogsEnabled = %v, want false (level 6 beats level 9)", got)
	}
}

// AC-7: airflowConfigurationOptions merge — mandatory and defaults each accumulate independently.
func TestReconcileAC7AirflowConfigurationOptionsMerge(t *testing.T) {
	rec, _ := testReconciler(t,
		localMWAAConfig("payments-prod", "general-policy",
			cascade.MWAAConfigSection{
				AirflowConfigurationOptions: map[string]string{"core.dags_are_paused_at_creation": "True"},
			},
			cascade.MWAAConfigSection{
				AirflowConfigurationOptions: map[string]string{"core.dag_run_conf_overrides_params": "False"},
			},
		),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getMWAAConfig(t, rec.Client, "payments-prod", "general-policy")
	mandatoryOpts := updated.Status.EffectiveConfig.Mandatory.AirflowConfigurationOptions
	if mandatoryOpts["core.dags_are_paused_at_creation"] != "True" {
		t.Fatalf("mandatory airflowConfigurationOptions[core.dags_are_paused_at_creation] = %q, want True", mandatoryOpts["core.dags_are_paused_at_creation"])
	}
	defaultsOpts := updated.Status.EffectiveConfig.Defaults.AirflowConfigurationOptions
	if defaultsOpts["core.dag_run_conf_overrides_params"] != "False" {
		t.Fatalf("defaults airflowConfigurationOptions[core.dag_run_conf_overrides_params] = %q, want False", defaultsOpts["core.dag_run_conf_overrides_params"])
	}
}

// AC-8: Tags from KropathConfig tier-level and MWAAConfig are union-merged.
func TestReconcileAC8TagUnionMerge(t *testing.T) {
	rec, _ := testReconciler(t,
		globalKropathConfig(v1alpha1.KropathConfigTier{
			Tags: map[string]string{"cost-centre": "infra", "shared-key": "from-global-kropath"},
		}),
		localMWAAConfig("payments-prod", "general-policy",
			cascade.MWAAConfigSection{Tags: map[string]string{"env": "prod", "shared-key": "from-local-mwaaconfig"}},
			cascade.MWAAConfigSection{},
		),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getMWAAConfig(t, rec.Client, "payments-prod", "general-policy")
	tags := updated.Status.EffectiveConfig.Mandatory.Tags
	if tags["cost-centre"] != "infra" {
		t.Fatalf("tags[cost-centre] = %q, want infra", tags["cost-centre"])
	}
	if tags["env"] != "prod" {
		t.Fatalf("tags[env] = %q, want prod", tags["env"])
	}
	// Level 1 (global KropathConfig mandatory) wins over level 4 (local MWAAConfig mandatory).
	if tags["shared-key"] != "from-global-kropath" {
		t.Fatalf("tags[shared-key] = %q, want from-global-kropath (level-1 wins)", tags["shared-key"])
	}
}

// AC-10: All 10 per-component logging fields propagate from localMWAAConfig.defaults (level 6).
func TestReconcileAC10AllLoggingFieldsPropagate(t *testing.T) {
	trueVal := true
	rec, _ := testReconciler(t,
		localMWAAConfig("payments-prod", "general-policy",
			cascade.MWAAConfigSection{},
			cascade.MWAAConfigSection{
				DagProcessingLogsEnabled: &trueVal,
				DagProcessingLogsLevel:   "INFO",
				SchedulerLogsEnabled:     &trueVal,
				SchedulerLogsLevel:       "INFO",
				TaskLogsEnabled:          &trueVal,
				TaskLogsLevel:            "INFO",
				WebserverLogsEnabled:     &trueVal,
				WebserverLogsLevel:       "INFO",
				WorkerLogsEnabled:        &trueVal,
				WorkerLogsLevel:          "INFO",
			},
		),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getMWAAConfig(t, rec.Client, "payments-prod", "general-policy")
	d := updated.Status.EffectiveConfig.Defaults
	if d.DagProcessingLogsEnabled == nil || !*d.DagProcessingLogsEnabled {
		t.Fatal("defaults.dagProcessingLogsEnabled should be &true")
	}
	if d.DagProcessingLogsLevel != "INFO" {
		t.Fatalf("defaults.dagProcessingLogsLevel = %q, want INFO", d.DagProcessingLogsLevel)
	}
	if d.SchedulerLogsEnabled == nil || !*d.SchedulerLogsEnabled {
		t.Fatal("defaults.schedulerLogsEnabled should be &true")
	}
	if d.SchedulerLogsLevel != "INFO" {
		t.Fatalf("defaults.schedulerLogsLevel = %q, want INFO", d.SchedulerLogsLevel)
	}
	if d.TaskLogsEnabled == nil || !*d.TaskLogsEnabled {
		t.Fatal("defaults.taskLogsEnabled should be &true")
	}
	if d.TaskLogsLevel != "INFO" {
		t.Fatalf("defaults.taskLogsLevel = %q, want INFO", d.TaskLogsLevel)
	}
	if d.WebserverLogsEnabled == nil || !*d.WebserverLogsEnabled {
		t.Fatal("defaults.webserverLogsEnabled should be &true")
	}
	if d.WebserverLogsLevel != "INFO" {
		t.Fatalf("defaults.webserverLogsLevel = %q, want INFO", d.WebserverLogsLevel)
	}
	if d.WorkerLogsEnabled == nil || !*d.WorkerLogsEnabled {
		t.Fatal("defaults.workerLogsEnabled should be &true")
	}
	if d.WorkerLogsLevel != "INFO" {
		t.Fatalf("defaults.workerLogsLevel = %q, want INFO", d.WorkerLogsLevel)
	}
}

// TestNamingTemplateOnlyFromMWAAConfig confirms namingTemplate is not sourced from KropathConfig.
func TestNamingTemplateOnlyFromMWAAConfig(t *testing.T) {
	rec, _ := testReconciler(t,
		globalMWAAConfig("general-policy",
			cascade.MWAAConfigSection{},
			cascade.MWAAConfigSection{NamingTemplate: "mwaa-{namespace}-{name}"},
		),
		localMWAAConfig("payments-prod", "general-policy", cascade.MWAAConfigSection{}, cascade.MWAAConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getMWAAConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.Defaults.NamingTemplate; got != "mwaa-{namespace}-{name}" {
		t.Fatalf("defaults.namingTemplate = %q, want mwaa-{namespace}-{name}", got)
	}
}

func TestRequestsForKropathConfigChangeGlobal(t *testing.T) {
	rec, _ := testReconciler(t,
		localMWAAConfig("payments-prod", "general-policy", cascade.MWAAConfigSection{}, cascade.MWAAConfigSection{}),
		localMWAAConfig("sandbox", "general-policy", cascade.MWAAConfigSection{}, cascade.MWAAConfigSection{}),
		localMWAAConfig("payments-prod", "other-policy", cascade.MWAAConfigSection{}, cascade.MWAAConfigSection{}),
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

func TestRequestsForKropathConfigChangeLocalNamespaceEnqueuesAllConfigsInNamespace(t *testing.T) {
	rec, _ := testReconciler(t,
		localMWAAConfig("payments-prod", "general-policy", cascade.MWAAConfigSection{}, cascade.MWAAConfigSection{}),
		localMWAAConfig("payments-prod", "other-policy", cascade.MWAAConfigSection{}, cascade.MWAAConfigSection{}),
		localMWAAConfig("sandbox", "general-policy", cascade.MWAAConfigSection{}, cascade.MWAAConfigSection{}),
	)

	got := rec.requestsForKropathConfigChange(context.Background(), &v1alpha1.KropathConfig{
		ObjectMeta: metav1.ObjectMeta{Name: util.KropathConfigName, Namespace: "payments-prod"},
	})

	if len(got) != 2 {
		t.Fatalf("requests len = %d, want 2 — default KPC triggers all configs in its namespace (%#v)", len(got), got)
	}
}

func TestRequestsForMWAAConfigChangeGlobal(t *testing.T) {
	rec, _ := testReconciler(t,
		localMWAAConfig("payments-prod", "general-policy", cascade.MWAAConfigSection{}, cascade.MWAAConfigSection{}),
		localMWAAConfig("data-prod", "general-policy", cascade.MWAAConfigSection{}, cascade.MWAAConfigSection{}),
	)

	got := rec.requestsForMWAAConfigChange(context.Background(), &v1alpha1.MWAAConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "general-policy", Namespace: "kro-system"},
	})

	if len(got) != 2 {
		t.Fatalf("requests len = %d, want 2 (%#v)", len(got), got)
	}
}

func TestRequestsForMWAAConfigChangeNonGlobalIgnored(t *testing.T) {
	rec, _ := testReconciler(t,
		localMWAAConfig("payments-prod", "general-policy", cascade.MWAAConfigSection{}, cascade.MWAAConfigSection{}),
	)

	got := rec.requestsForMWAAConfigChange(context.Background(), &v1alpha1.MWAAConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "general-policy", Namespace: "payments-prod"},
	})

	if len(got) != 0 {
		t.Fatalf("requests len = %d, want 0 (%#v)", len(got), got)
	}
}

func TestProviderIdentityPropagates(t *testing.T) {
	rec, _ := testReconciler(t,
		namespaceWithIdentity("payments-prod", "123456789012", "ap-southeast-2"),
		localMWAAConfig("payments-prod", "general-policy", cascade.MWAAConfigSection{}, cascade.MWAAConfigSection{}),
	)

	if _, err := rec.Reconcile(context.Background(), req("payments-prod", "general-policy")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated := getMWAAConfig(t, rec.Client, "payments-prod", "general-policy")
	if got := updated.Status.EffectiveConfig.AWS.AccountID; got != "123456789012" {
		t.Fatalf("aws.accountId = %q, want 123456789012", got)
	}
	if got := updated.Status.EffectiveConfig.AWS.Region; got != "ap-southeast-2" {
		t.Fatalf("aws.region = %q, want ap-southeast-2", got)
	}
}

// ─── Test fixtures ──────────────────────────────────────────────────────────

func testReconciler(t *testing.T, objs ...runtime.Object) (*Reconciler, *v1alpha1.MWAAConfig) {
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
		WithStatusSubresource(&v1alpha1.MWAAConfig{})
	for _, obj := range objs {
		builder = builder.WithRuntimeObjects(obj)
	}
	cl := builder.Build()
	cfg := &v1alpha1.MWAAConfig{}
	if err := cl.Get(context.Background(), client.ObjectKey{Namespace: "payments-prod", Name: "general-policy"}, cfg); err != nil {
		t.Fatalf("seed local MWAAConfig: %v", err)
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

func globalMWAAConfig(name string, mandatory, defaults cascade.MWAAConfigSection) *v1alpha1.MWAAConfig {
	return &v1alpha1.MWAAConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "MWAAConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kro-system",
		},
		Spec: v1alpha1.MWAAConfigSpec{
			Mandatory: mandatory,
			Defaults:  defaults,
		},
	}
}

func localMWAAConfig(namespace, name string, mandatory, defaults cascade.MWAAConfigSection) *v1alpha1.MWAAConfig {
	cfg := globalMWAAConfig(name, mandatory, defaults)
	cfg.Namespace = namespace
	return cfg
}

func getMWAAConfig(t *testing.T, c client.Client, namespace, name string) *v1alpha1.MWAAConfig {
	t.Helper()
	cfg := &v1alpha1.MWAAConfig{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("MWAAConfig"))
	if err := c.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, cfg); err != nil {
		t.Fatalf("get MWAAConfig %s/%s: %v", namespace, name, err)
	}
	return cfg
}

func req(namespace, name string) ctrl.Request {
	return ctrl.Request{NamespacedName: client.ObjectKey{Namespace: namespace, Name: name}}
}
