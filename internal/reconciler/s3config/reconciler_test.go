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

package s3config

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/kropath/kropath-aws-controller/api/v1alpha1"
	"github.com/kropath/kropath-aws-controller/internal/cascade"
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
		WithStatusSubresource(&v1alpha1.S3Config{}).
		Build()
	return &Reconciler{Client: c, Scheme: scheme}, c
}

// --- Object factory helpers ---

// namespace returns a resource namespace resolving its global tier to
// globalNS, with a valid account/region so placement resolves successfully.
// Most tests in this file exercise cascade-merge logic, not placement itself,
// so this keeps them from repeating that boilerplate.
func namespace(name string) *corev1.Namespace {
	return namespaceWithAnnotation(name, globalNS)
}

func namespaceWithAnnotation(name, globalConfigNS string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				util.GlobalConfigNamespaceAnnotation: globalConfigNS,
				util.OwnerAccountIDAnnotation:        "111122223333",
				util.DefaultRegionAnnotation:         "ap-southeast-2",
			},
		},
	}
}

// governanceOnlyNamespace carries none of the placement annotations (spec
// §6.4, AC-12): it is exempt from every placement rule and its config objects
// never receive status.effectiveConfig.
func governanceOnlyNamespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
}

func globalKropathConfig(s3 cascade.S3Section) *v1alpha1.KropathConfig {
	return &v1alpha1.KropathConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "KropathConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.KropathConfigName,
			Namespace: globalNS,
		},
		Spec: v1alpha1.KropathConfigSpec{
			Mandatory: v1alpha1.KropathConfigTier{S3: s3},
		},
	}
}

// localKropathConfig creates the local-tier KropathConfig in a namespace,
// always named util.KropathConfigName (ADR-018 D-1).
func localKropathConfig(ns string, s3 cascade.S3Section) *v1alpha1.KropathConfig {
	return &v1alpha1.KropathConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "KropathConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.KropathConfigName,
			Namespace: ns,
		},
		Spec: v1alpha1.KropathConfigSpec{
			Mandatory: v1alpha1.KropathConfigTier{S3: s3},
		},
	}
}

func globalS3Config(name string, mandatory, defaults cascade.S3ConfigSection) *v1alpha1.S3Config {
	return &v1alpha1.S3Config{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "S3Config"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: globalNS,
		},
		Spec: v1alpha1.S3ConfigSpec{
			Mandatory: mandatory,
			Defaults:  defaults,
		},
	}
}

func localS3Config(ns, name string, mandatory, defaults cascade.S3ConfigSection) *v1alpha1.S3Config {
	return &v1alpha1.S3Config{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "S3Config"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: v1alpha1.S3ConfigSpec{
			Mandatory: mandatory,
			Defaults:  defaults,
		},
	}
}

func getS3Config(t *testing.T, c client.Client, ns, name string) *v1alpha1.S3Config {
	t.Helper()
	cfg := &v1alpha1.S3Config{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("S3Config"))
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: ns, Name: name}, cfg); err != nil {
		t.Fatalf("Get S3Config %s/%s: %v", ns, name, err)
	}
	return cfg
}

// --- Core reconcile tests ---

func TestReconcilerReconcile(t *testing.T) {
	rec, c := testReconciler(t,
		namespace("payments-prod"),
		globalKropathConfig(cascade.S3Section{EncryptionAlgorithm: "aws:kms"}),
		localKropathConfig("payments-prod", cascade.S3Section{BlockPublicAccess: true}),
		globalS3Config("general-policy",
			cascade.S3ConfigSection{KmsKeyArn: "arn:global-s3"},
			cascade.S3ConfigSection{Versioning: "Enabled"},
		),
		localS3Config("payments-prod", "general-policy",
			cascade.S3ConfigSection{Versioning: "Suspended"},
			cascade.S3ConfigSection{LogDeliveryBucket: "logs"},
		),
	)

	_, err := rec.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "payments-prod", Name: "general-policy"},
	})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	got := getS3Config(t, c, "payments-prod", "general-policy")
	if got.Status.EffectiveConfig.Mandatory.EncryptionAlgorithm != "aws:kms" {
		t.Errorf("mandatory.encryptionAlgorithm = %q, want aws:kms", got.Status.EffectiveConfig.Mandatory.EncryptionAlgorithm)
	}
	if !got.Status.EffectiveConfig.Mandatory.BlockPublicAccess {
		t.Error("mandatory.blockPublicAccess should be true")
	}
	if got.Status.EffectiveConfig.Mandatory.KmsKeyArn != "arn:global-s3" {
		t.Errorf("mandatory.kmsKeyArn = %q, want arn:global-s3", got.Status.EffectiveConfig.Mandatory.KmsKeyArn)
	}
	if got.Status.EffectiveConfig.Defaults.Versioning != "Enabled" {
		t.Errorf("defaults.versioning = %q, want Enabled", got.Status.EffectiveConfig.Defaults.Versioning)
	}
	if got.Status.EffectiveConfig.Defaults.LogDeliveryBucket != "logs" {
		t.Errorf("defaults.logDeliveryBucket = %q, want logs", got.Status.EffectiveConfig.Defaults.LogDeliveryBucket)
	}
}

// The reconciler resolves the local KropathConfig by the fixed singleton name
// (ADR-018 D-1), not by cfg.Name. A leftover object under the old profile-name
// convention (e.g. from before the singleton migration) must be ignored.
func TestLocalKropathConfigLooksUpSingletonName(t *testing.T) {
	rec, c := testReconciler(t,
		namespace("payments-prod"),
		// KPC named util.KropathConfigName (local) — the one the reconciler should pick up
		localKropathConfig("payments-prod", cascade.S3Section{BlockPublicAccess: true}),
		// KPC named "general-policy" — a pre-migration leftover; must NOT be picked up
		&v1alpha1.KropathConfig{
			TypeMeta:   metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "KropathConfig"},
			ObjectMeta: metav1.ObjectMeta{Name: "general-policy", Namespace: "payments-prod"},
			Spec: v1alpha1.KropathConfigSpec{
				Mandatory: v1alpha1.KropathConfigTier{S3: cascade.S3Section{EnforceHttpsOnly: true}},
			},
		},
		localS3Config("payments-prod", "general-policy",
			cascade.S3ConfigSection{}, cascade.S3ConfigSection{},
		),
	)

	_, err := rec.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "payments-prod", Name: "general-policy"},
	})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	got := getS3Config(t, c, "payments-prod", "general-policy")
	if !got.Status.EffectiveConfig.Mandatory.BlockPublicAccess {
		t.Errorf("mandatory.blockPublicAccess should be true — set only on KPC/%s", util.KropathConfigName)
	}
	// EnforceHttpsOnly is from the pre-migration "general-policy" object, which is NOT the singleton.
	if got.Status.EffectiveConfig.Mandatory.EnforceHttpsOnly {
		t.Error("mandatory.enforceHttpsOnly should be false — KPC/'general-policy' is not the local singleton")
	}
}

// Gap 1: when namespace carries aws.kropath.run/global-config-namespace annotation,
// the reconciler resolves the global KropathConfig from the annotated namespace.
func TestResolveGlobalNamespaceFromAnnotation(t *testing.T) {
	const customGlobalNS = "platform-config"

	rec, c := testReconciler(t,
		// Annotate the resource namespace to point to a custom global namespace
		namespaceWithAnnotation("payments-prod", customGlobalNS),
		namespace(customGlobalNS),
		// Global KPC in the custom namespace (NOT kro-system)
		&v1alpha1.KropathConfig{
			TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "KropathConfig"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      util.KropathConfigName,
				Namespace: customGlobalNS,
			},
			Spec: v1alpha1.KropathConfigSpec{
				Mandatory: v1alpha1.KropathConfigTier{S3: cascade.S3Section{EncryptionAlgorithm: "aws:kms"}},
			},
		},
		localKropathConfig("payments-prod", cascade.S3Section{}),
		localS3Config("payments-prod", "general-policy",
			cascade.S3ConfigSection{}, cascade.S3ConfigSection{},
		),
	)

	_, err := rec.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "payments-prod", Name: "general-policy"},
	})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	got := getS3Config(t, c, "payments-prod", "general-policy")
	if got.Status.EffectiveConfig.Mandatory.EncryptionAlgorithm != "aws:kms" {
		t.Errorf("mandatory.encryptionAlgorithm = %q, want aws:kms — global KPC in annotated namespace should apply",
			got.Status.EffectiveConfig.Mandatory.EncryptionAlgorithm)
	}
}

// AC-12: no kro-system default. A namespace lacking the global-config-namespace
// annotation is governance-only -- exempt from placement, no effectiveConfig
// written -- never a silent fallback to any namespace (spec §5.5, §6.4).
func TestGovernanceOnlyNamespaceReceivesNoEffectiveConfig(t *testing.T) {
	rec, c := testReconciler(t,
		governanceOnlyNamespace("platform-shared"),
		globalKropathConfig(cascade.S3Section{EncryptionAlgorithm: "aws:kms"}),
		localS3Config("platform-shared", "general-policy",
			cascade.S3ConfigSection{}, cascade.S3ConfigSection{},
		),
	)

	_, err := rec.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "platform-shared", Name: "general-policy"},
	})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	got := getS3Config(t, c, "platform-shared", "general-policy")
	var zero v1alpha1.EffectiveS3Config
	if !reflect.DeepEqual(got.Status.EffectiveConfig, zero) {
		t.Fatalf("EffectiveConfig = %+v, want zero value for a governance-only namespace", got.Status.EffectiveConfig)
	}
	// Go-value equality against the zero value is not sufficient (KRO-1199):
	// encoding/json's omitempty is a no-op on non-pointer struct fields, so a
	// zero-valued EffectiveConfig previously still serialized as
	// "effectiveConfig":{"aws":{},"defaults":{},"mandatory":{}} instead of
	// being absent from the wire payload the API server actually stores.
	// Assert the marshaled JSON directly so this regression can't slip back in.
	statusJSON, err := json.Marshal(got.Status)
	if err != nil {
		t.Fatalf("json.Marshal(Status): %v", err)
	}
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(statusJSON, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(Status): %v", err)
	}
	if _, present := decoded["effectiveConfig"]; present {
		t.Fatalf("marshaled status carries an \"effectiveConfig\" key for a governance-only namespace; want it absent: %s", statusJSON)
	}
	cond := findCondition(got.Status.Conditions, "Reconciled")
	if cond == nil || cond.Status != metav1.ConditionTrue || cond.Reason != util.ReasonGlobalTierInput {
		t.Fatalf("Reconciled condition = %+v, want True/%s", cond, util.ReasonGlobalTierInput)
	}
}

// Gap 3: NamingTemplate from S3Config (not KropathConfig) is carried through to effective config.
func TestNamingTemplateFromS3ConfigApplied(t *testing.T) {
	rec, c := testReconciler(t,
		namespace("payments-prod"),
		localKropathConfig("payments-prod", cascade.S3Section{}),
		localS3Config("payments-prod", "general-policy",
			cascade.S3ConfigSection{NamingTemplate: "{account_id}-{namespace}-{name}"},
			cascade.S3ConfigSection{},
		),
	)

	_, err := rec.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "payments-prod", Name: "general-policy"},
	})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	got := getS3Config(t, c, "payments-prod", "general-policy")
	want := "{account_id}-{namespace}-{name}"
	if got.Status.EffectiveConfig.Mandatory.NamingTemplate != want {
		t.Errorf("mandatory.namingTemplate = %q, want %q", got.Status.EffectiveConfig.Mandatory.NamingTemplate, want)
	}
}

// --- Fan-out tests ---

func TestRequestsForKropathConfigChangeIncludesGlobalAndLocalMatches(t *testing.T) {
	rec, _ := testReconciler(t,
		namespace("payments-prod"),
		namespace("other-ns"),
		globalS3Config("general-policy", cascade.S3ConfigSection{}, cascade.S3ConfigSection{}),
		localS3Config("payments-prod", "general-policy", cascade.S3ConfigSection{}, cascade.S3ConfigSection{}),
		localS3Config("other-ns", "general-policy", cascade.S3ConfigSection{}, cascade.S3ConfigSection{}),
		localS3Config("payments-prod", "other-policy", cascade.S3ConfigSection{}, cascade.S3ConfigSection{}),
	)

	// A global KPC (in kro-system) enqueues every S3Config whose resolved
	// global namespace is kro-system, classified by namespace only (ADR-018
	// D-1) -- not filtered by name, since the name is now a fixed singleton.
	requests := rec.requestsForKropathConfigChange(context.Background(), &v1alpha1.KropathConfig{
		ObjectMeta: metav1.ObjectMeta{Name: util.KropathConfigName, Namespace: globalNS},
	})

	// kro-system/general-policy, payments-prod/general-policy, payments-prod/other-policy, other-ns/general-policy
	if len(requests) != 4 {
		t.Fatalf("requests len = %d, want 4; got %#v", len(requests), requests)
	}
	got := map[string]bool{}
	for _, req := range requests {
		got[req.Namespace+"/"+req.Name] = true
	}
	for _, want := range []string{
		globalNS + "/general-policy",
		"payments-prod/general-policy",
		"payments-prod/other-policy",
		"other-ns/general-policy",
	} {
		if !got[want] {
			t.Errorf("missing request %q in %#v", want, requests)
		}
	}
}

func TestRequestsForKropathConfigChangeLocalNamespaceEnqueuesAllConfigsInNamespace(t *testing.T) {
	rec, _ := testReconciler(t,
		namespace("payments-prod"),
		localS3Config("payments-prod", "general-policy", cascade.S3ConfigSection{}, cascade.S3ConfigSection{}),
		localS3Config("payments-prod", "other-policy", cascade.S3ConfigSection{}, cascade.S3ConfigSection{}),
		// Different namespace — should NOT be enqueued by this local KPC change
		localS3Config("other-ns", "general-policy", cascade.S3ConfigSection{}, cascade.S3ConfigSection{}),
	)

	// Local KPC in "payments-prod" — should enqueue all S3Configs in that namespace
	requests := rec.requestsForKropathConfigChange(context.Background(), &v1alpha1.KropathConfig{
		ObjectMeta: metav1.ObjectMeta{Name: util.KropathConfigName, Namespace: "payments-prod"},
	})

	if len(requests) != 2 {
		t.Fatalf("requests len = %d, want 2; got %#v", len(requests), requests)
	}
}

func TestRequestsForS3ConfigChangeQueuesNamespaceMatchesForGlobalSource(t *testing.T) {
	rec, _ := testReconciler(t,
		namespace("payments-prod"),
		namespace("data-prod"),
		globalS3Config("general-policy", cascade.S3ConfigSection{}, cascade.S3ConfigSection{}),
		localS3Config("payments-prod", "general-policy", cascade.S3ConfigSection{}, cascade.S3ConfigSection{}),
		localS3Config("data-prod", "general-policy", cascade.S3ConfigSection{}, cascade.S3ConfigSection{}),
	)

	requests := rec.requestsForS3ConfigChange(context.Background(), &v1alpha1.S3Config{
		ObjectMeta: metav1.ObjectMeta{Name: "general-policy", Namespace: globalNS},
	})

	if len(requests) != 2 {
		t.Fatalf("requests len = %d, want 2; got %#v", len(requests), requests)
	}
}

// --- ADR-015 §5.9 profile fallthrough tests ---

func findCondition(conditions []metav1.Condition, condType string) *metav1.Condition {
	for i := range conditions {
		if conditions[i].Type == condType {
			return &conditions[i]
		}
	}
	return nil
}

// Unit: configRef "pci" with no global S3Config/pci merges S3Config/general-policy (ADR-015 §5.9).
func TestGlobalProfileFallthroughWhenRequestedProfileMissing(t *testing.T) {
	rec, c := testReconciler(t,
		namespace("payments-prod"),
		localKropathConfig("payments-prod", cascade.S3Section{}),
		// Only the fallthrough target exists globally — no S3Config/pci in kro-system.
		globalS3Config("general-policy",
			cascade.S3ConfigSection{KmsKeyArn: "arn:general-policy"},
			cascade.S3ConfigSection{},
		),
		// Local S3Config/pci — this is the object being reconciled.
		localS3Config("payments-prod", "pci", cascade.S3ConfigSection{}, cascade.S3ConfigSection{}),
	)

	_, err := rec.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "payments-prod", Name: "pci"},
	})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	got := getS3Config(t, c, "payments-prod", "pci")
	if got.Status.EffectiveConfig.Mandatory.KmsKeyArn != "arn:general-policy" {
		t.Fatalf("mandatory.kmsKeyArn = %q, want arn:general-policy (fallthrough was not applied)", got.Status.EffectiveConfig.Mandatory.KmsKeyArn)
	}

	cond := findCondition(got.Status.Conditions, util.ConfigProfileResolvedConditionType)
	if cond == nil {
		t.Fatalf("expected a %s condition", util.ConfigProfileResolvedConditionType)
	}
	if cond.Status != metav1.ConditionTrue || cond.Reason != "ProfileFallthrough" {
		t.Fatalf("condition = %+v, want Status=True Reason=ProfileFallthrough", cond)
	}
}

// Unit: configRef "pci" with S3Config/pci present globally merges that config, not the
// fallthrough target (ADR-015 §5.9 — the fallthrough hop only fires on NotFound).
func TestGlobalProfileNoFallthroughWhenRequestedProfileExists(t *testing.T) {
	rec, c := testReconciler(t,
		namespace("payments-prod"),
		localKropathConfig("payments-prod", cascade.S3Section{}),
		globalS3Config("pci",
			cascade.S3ConfigSection{KmsKeyArn: "arn:pci"},
			cascade.S3ConfigSection{},
		),
		globalS3Config("general-policy",
			cascade.S3ConfigSection{KmsKeyArn: "arn:general-policy"},
			cascade.S3ConfigSection{},
		),
		localS3Config("payments-prod", "pci", cascade.S3ConfigSection{}, cascade.S3ConfigSection{}),
	)

	_, err := rec.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "payments-prod", Name: "pci"},
	})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	got := getS3Config(t, c, "payments-prod", "pci")
	if got.Status.EffectiveConfig.Mandatory.KmsKeyArn != "arn:pci" {
		t.Fatalf("mandatory.kmsKeyArn = %q, want arn:pci (fallthrough must not override an existing profile)", got.Status.EffectiveConfig.Mandatory.KmsKeyArn)
	}

	cond := findCondition(got.Status.Conditions, util.ConfigProfileResolvedConditionType)
	if cond == nil {
		t.Fatalf("expected a %s condition", util.ConfigProfileResolvedConditionType)
	}
	if cond.Status != metav1.ConditionTrue || cond.Reason != "ProfileFound" {
		t.Fatalf("condition = %+v, want Status=True Reason=ProfileFound", cond)
	}
}

// Unit: a nonexistent profile at the global tier — neither the requested profile nor the
// fallthrough target exists — merges an empty global tier (ADR-002 D-1: absence is silently
// skipped, not an error) but is observable via the ProfileUnresolved condition (ADR-015 §5.9).
func TestGlobalProfileUnresolvedIsObservable(t *testing.T) {
	rec, c := testReconciler(t,
		namespace("payments-prod"),
		localKropathConfig("payments-prod", cascade.S3Section{}),
		// No global S3Config at all — neither "pci" nor "general-policy".
		localS3Config("payments-prod", "pci", cascade.S3ConfigSection{KmsKeyArn: "arn:local-pci"}, cascade.S3ConfigSection{}),
	)

	_, err := rec.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "payments-prod", Name: "pci"},
	})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	got := getS3Config(t, c, "payments-prod", "pci")
	// The global tier is empty, but the local tier (level 4) still merges — confirms this
	// is a skipped tier, not a reconcile failure.
	if got.Status.EffectiveConfig.Mandatory.KmsKeyArn != "arn:local-pci" {
		t.Fatalf("mandatory.kmsKeyArn = %q, want arn:local-pci (local tier must still merge)", got.Status.EffectiveConfig.Mandatory.KmsKeyArn)
	}

	cond := findCondition(got.Status.Conditions, util.ConfigProfileResolvedConditionType)
	if cond == nil {
		t.Fatalf("expected a %s condition", util.ConfigProfileResolvedConditionType)
	}
	if cond.Status != metav1.ConditionFalse || cond.Reason != "ProfileUnresolved" {
		t.Fatalf("condition = %+v, want Status=False Reason=ProfileUnresolved", cond)
	}
}

// Integration-style: a tenant namespace on a non-default profile still receives the
// org-wide guardrails via the global fallthrough (ADR-015 §5.9).
func TestNonDefaultProfileReceivesOrgWideGuardrailsViaFallthrough(t *testing.T) {
	rec, c := testReconciler(t,
		namespace("payments-prod"),
		localKropathConfig("payments-prod", cascade.S3Section{}),
		globalS3Config("general-policy",
			cascade.S3ConfigSection{BlockPublicAccess: true, EncryptionAlgorithm: "aws:kms"},
			cascade.S3ConfigSection{},
		),
		localS3Config("payments-prod", "pci", cascade.S3ConfigSection{}, cascade.S3ConfigSection{}),
	)

	_, err := rec.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "payments-prod", Name: "pci"},
	})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	got := getS3Config(t, c, "payments-prod", "pci")
	if !got.Status.EffectiveConfig.Mandatory.BlockPublicAccess {
		t.Error("mandatory.blockPublicAccess should be true — org-wide guardrail lost on a non-default profile")
	}
	if got.Status.EffectiveConfig.Mandatory.EncryptionAlgorithm != "aws:kms" {
		t.Errorf("mandatory.encryptionAlgorithm = %q, want aws:kms — org-wide guardrail lost on a non-default profile", got.Status.EffectiveConfig.Mandatory.EncryptionAlgorithm)
	}
}
