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

package util_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kropath/kropath-aws-controller/internal/reconciler/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

// readPlacementResolutionsTotal reads kropath_placement_resolutions_total{reason}
// from the same global registry production code registers into (internal/metrics
// exports no counter internals, by design, so this package -- external to both
// util and metrics -- gathers the registry directly, exactly as Prometheus would).
func readPlacementResolutionsTotal(t *testing.T, reason string) float64 {
	t.Helper()
	mfs, err := ctrlmetrics.Registry.Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	for _, mf := range mfs {
		if mf.GetName() != "kropath_placement_resolutions_total" {
			continue
		}
		for _, m := range mf.GetMetric() {
			for _, lp := range m.GetLabel() {
				if lp.GetName() == "reason" && lp.GetValue() == reason {
					return m.GetCounter().GetValue()
				}
			}
		}
	}
	return 0
}

func testClient(t *testing.T, objs ...runtime.Object) *fake.ClientBuilder {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme: %v", err)
	}
	return fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objs...)
}

func namespaceWithAnnotations(name string, annotations map[string]string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Annotations: annotations},
	}
}

// AC-12: no kro-system default; a namespace without the annotation resolves to
// RoleGovernanceOnly, never a fallback namespace name.
func TestResolveNamespaceRoleNoKroSystemDefault(t *testing.T) {
	ns := namespaceWithAnnotations("payments-prod", nil)
	c := testClient(t, ns).Build()

	role, globalNS, err := util.ResolveNamespaceRole(context.Background(), c, "payments-prod")
	if err != nil {
		t.Fatalf("ResolveNamespaceRole: unexpected error: %v", err)
	}
	if role != util.RoleGovernanceOnly {
		t.Errorf("role = %v, want RoleGovernanceOnly", role)
	}
	if globalNS != "" {
		t.Errorf("globalNS = %q, want empty (no kro-system default)", globalNS)
	}
}

func TestResolveNamespaceRoleEmptyAnnotationIsGovernanceOnly(t *testing.T) {
	ns := namespaceWithAnnotations("payments-prod", map[string]string{util.GlobalConfigNamespaceAnnotation: ""})
	c := testClient(t, ns).Build()

	role, _, err := util.ResolveNamespaceRole(context.Background(), c, "payments-prod")
	if err != nil {
		t.Fatalf("ResolveNamespaceRole: unexpected error: %v", err)
	}
	if role != util.RoleGovernanceOnly {
		t.Errorf("role = %v, want RoleGovernanceOnly", role)
	}
}

func TestResolveNamespaceRoleResourceNamespace(t *testing.T) {
	ns := namespaceWithAnnotations("payments-prod", map[string]string{util.GlobalConfigNamespaceAnnotation: "platform-config"})
	c := testClient(t, ns).Build()

	role, globalNS, err := util.ResolveNamespaceRole(context.Background(), c, "payments-prod")
	if err != nil {
		t.Fatalf("ResolveNamespaceRole: unexpected error: %v", err)
	}
	if role != util.RoleResource {
		t.Errorf("role = %v, want RoleResource", role)
	}
	if globalNS != "platform-config" {
		t.Errorf("globalNS = %q, want %q", globalNS, "platform-config")
	}
}

// A read failure is an error, never a role (spec §5.5) -- this is the
// NamespaceUnreadable trigger at the config-object site (§6.3).
func TestResolveNamespaceRoleMissingNamespaceIsError(t *testing.T) {
	c := testClient(t).Build()

	_, _, err := util.ResolveNamespaceRole(context.Background(), c, "nonexistent")
	if err == nil {
		t.Fatal("ResolveNamespaceRole: expected error for missing namespace, got nil")
	}
}

// AC-11: DerivePartition longest-prefix table, us-isob-east-1 pinned against a
// first-match bug (it also matches the shorter us-iso- prefix).
func TestDerivePartitionLongestPrefixTable(t *testing.T) {
	cases := []struct {
		region string
		want   string
	}{
		{"ap-southeast-2", "aws"},
		{"cn-north-1", "aws-cn"},
		{"us-gov-west-1", "aws-us-gov"},
		{"us-iso-east-1", "aws-iso"},
		{"us-isob-east-1", "aws-iso-b"},
		{"eu-isoe-west-1", "aws-iso-e"},
		{"us-isof-south-1", "aws-iso-f"},
		{"eu-isof-south-1", "aws-iso-f"},
		{"", ""},
	}
	for _, tc := range cases {
		if got := util.DerivePartition(tc.region); got != tc.want {
			t.Errorf("DerivePartition(%q) = %q, want %q", tc.region, got, tc.want)
		}
	}
}

func TestResolvePlacementTeamAnnotationUnsupported(t *testing.T) {
	ns := namespaceWithAnnotations("payments-prod", map[string]string{
		util.OwnerAccountIDAnnotation: "111122223333",
		util.DefaultRegionAnnotation:  "ap-southeast-2",
		util.TeamIDAnnotation:         "payments",
	})
	c := testClient(t, ns).Build()

	_, placementErr, err := util.ResolvePlacement(context.Background(), c, "payments-prod")
	if err != nil {
		t.Fatalf("ResolvePlacement: unexpected error: %v", err)
	}
	if placementErr == nil || placementErr.Reason != util.ReasonTeamAnnotationUnsupported {
		t.Fatalf("placementErr = %+v, want reason %s", placementErr, util.ReasonTeamAnnotationUnsupported)
	}
}

func TestResolvePlacementMissingAccountAnnotation(t *testing.T) {
	ns := namespaceWithAnnotations("payments-prod", map[string]string{util.DefaultRegionAnnotation: "ap-southeast-2"})
	c := testClient(t, ns).Build()

	_, placementErr, err := util.ResolvePlacement(context.Background(), c, "payments-prod")
	if err != nil {
		t.Fatalf("ResolvePlacement: unexpected error: %v", err)
	}
	if placementErr == nil || placementErr.Reason != util.ReasonMissingAccountAnnotation {
		t.Fatalf("placementErr = %+v, want reason %s", placementErr, util.ReasonMissingAccountAnnotation)
	}
}

// Absent versus present-but-empty are treated identically (spec §4.2).
func TestResolvePlacementEmptyAccountAnnotationIsMissing(t *testing.T) {
	ns := namespaceWithAnnotations("payments-prod", map[string]string{
		util.OwnerAccountIDAnnotation: "",
		util.DefaultRegionAnnotation:  "ap-southeast-2",
	})
	c := testClient(t, ns).Build()

	_, placementErr, err := util.ResolvePlacement(context.Background(), c, "payments-prod")
	if err != nil {
		t.Fatalf("ResolvePlacement: unexpected error: %v", err)
	}
	if placementErr == nil || placementErr.Reason != util.ReasonMissingAccountAnnotation {
		t.Fatalf("placementErr = %+v, want reason %s", placementErr, util.ReasonMissingAccountAnnotation)
	}
}

func TestResolvePlacementInvalidAccountAnnotation(t *testing.T) {
	ns := namespaceWithAnnotations("payments-prod", map[string]string{
		util.OwnerAccountIDAnnotation: "acct-1234",
		util.DefaultRegionAnnotation:  "ap-southeast-2",
	})
	c := testClient(t, ns).Build()

	_, placementErr, err := util.ResolvePlacement(context.Background(), c, "payments-prod")
	if err != nil {
		t.Fatalf("ResolvePlacement: unexpected error: %v", err)
	}
	if placementErr == nil || placementErr.Reason != util.ReasonInvalidAccountAnnotation {
		t.Fatalf("placementErr = %+v, want reason %s", placementErr, util.ReasonInvalidAccountAnnotation)
	}
	want := `namespace "payments-prod" annotation "services.k8s.aws/owner-account-id" has value "acct-1234", which is not a 12-digit AWS account ID`
	if placementErr.Message != want {
		t.Errorf("message = %q, want %q", placementErr.Message, want)
	}
}

func TestResolvePlacementMissingRegionAnnotation(t *testing.T) {
	ns := namespaceWithAnnotations("payments-prod", map[string]string{util.OwnerAccountIDAnnotation: "111122223333"})
	c := testClient(t, ns).Build()

	_, placementErr, err := util.ResolvePlacement(context.Background(), c, "payments-prod")
	if err != nil {
		t.Fatalf("ResolvePlacement: unexpected error: %v", err)
	}
	if placementErr == nil || placementErr.Reason != util.ReasonMissingRegionAnnotation {
		t.Fatalf("placementErr = %+v, want reason %s", placementErr, util.ReasonMissingRegionAnnotation)
	}
}

func TestResolvePlacementSuccess(t *testing.T) {
	ns := namespaceWithAnnotations("payments-prod", map[string]string{
		util.OwnerAccountIDAnnotation: "111122223333",
		util.DefaultRegionAnnotation:  "ap-southeast-2",
	})
	c := testClient(t, ns).Build()

	identity, placementErr, err := util.ResolvePlacement(context.Background(), c, "payments-prod")
	if err != nil {
		t.Fatalf("ResolvePlacement: unexpected error: %v", err)
	}
	if placementErr != nil {
		t.Fatalf("placementErr = %+v, want nil", placementErr)
	}
	// AC-14: the whole struct is populated or none of it is.
	if identity.AccountID == "" || identity.Region == "" || identity.Partition == "" {
		t.Errorf("identity = %+v, want no empty field on success", identity)
	}
	if identity.AccountID != "111122223333" || identity.Region != "ap-southeast-2" || identity.Partition != "aws" {
		t.Errorf("identity = %+v, want {111122223333 ap-southeast-2 aws}", identity)
	}
}

func TestResolvePlacementMissingNamespaceIsError(t *testing.T) {
	c := testClient(t).Build()

	_, placementErr, err := util.ResolvePlacement(context.Background(), c, "nonexistent")
	if err == nil {
		t.Fatal("ResolvePlacement: expected error for missing namespace, got nil")
	}
	if placementErr != nil {
		t.Errorf("placementErr = %+v, want nil when the Get itself failed", placementErr)
	}
}

// AC-10: the Namespace Event message and the config Reconciled message must be
// byte-identical -- both must render from the same PlacementError.
func TestResolvePlacementMessagesAreDeterministic(t *testing.T) {
	ns := namespaceWithAnnotations("payments-prod", map[string]string{util.OwnerAccountIDAnnotation: "111122223333"})
	c := testClient(t, ns).Build()

	_, err1, e1 := util.ResolvePlacement(context.Background(), c, "payments-prod")
	_, err2, e2 := util.ResolvePlacement(context.Background(), c, "payments-prod")
	if e1 != nil || e2 != nil {
		t.Fatalf("unexpected errors: %v, %v", e1, e2)
	}
	if err1.Message != err2.Message || err1.Reason != err2.Reason {
		t.Errorf("two evaluations diverged: %+v vs %+v", err1, err2)
	}
}

// spec §2.1, §5.2: PlacementReasons() is the closed set the §10.2.1
// label-value check enumerates against. All 8 values, including
// ResolvedFromNamespace, which no call site ever emits (spec §2.2) but which
// stays in the contract set so the design §9 matcher naming it still passes.
func TestPlacementReasonsReturnsAllEightValues(t *testing.T) {
	got := util.PlacementReasons()
	want := []string{
		util.ReasonTeamAnnotationUnsupported,
		util.ReasonMissingAccountAnnotation,
		util.ReasonInvalidAccountAnnotation,
		util.ReasonMissingRegionAnnotation,
		util.ReasonNamespaceUnreadable,
		util.ReasonGlobalTierInput,
		util.ReasonResolvedFromNamespace,
		util.ReasonPlacementResolved,
	}
	if len(got) != len(want) {
		t.Fatalf("PlacementReasons() = %v (len %d), want len %d", got, len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("PlacementReasons()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// spec §2.2: ResolvePlacement owns 6 of the 7 emitted reason values,
// table-driven over each.
func TestResolvePlacementIncrementsPlacementResolutionsTotal(t *testing.T) {
	cases := []struct {
		name        string
		annotations map[string]string
		wantReason  string
	}{
		{"team annotation unsupported", map[string]string{
			util.OwnerAccountIDAnnotation: "111122223333",
			util.DefaultRegionAnnotation:  "ap-southeast-2",
			util.TeamIDAnnotation:         "payments",
		}, util.ReasonTeamAnnotationUnsupported},
		{"missing account annotation", map[string]string{
			util.DefaultRegionAnnotation: "ap-southeast-2",
		}, util.ReasonMissingAccountAnnotation},
		{"invalid account annotation", map[string]string{
			util.OwnerAccountIDAnnotation: "acct-1234",
			util.DefaultRegionAnnotation:  "ap-southeast-2",
		}, util.ReasonInvalidAccountAnnotation},
		{"missing region annotation", map[string]string{
			util.OwnerAccountIDAnnotation: "111122223333",
		}, util.ReasonMissingRegionAnnotation},
		{"success", map[string]string{
			util.OwnerAccountIDAnnotation: "111122223333",
			util.DefaultRegionAnnotation:  "ap-southeast-2",
		}, util.ReasonPlacementResolved},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ns := namespaceWithAnnotations("payments-prod", tc.annotations)
			c := testClient(t, ns).Build()

			before := readPlacementResolutionsTotal(t, tc.wantReason)
			if _, _, err := util.ResolvePlacement(context.Background(), c, "payments-prod"); err != nil {
				t.Fatalf("ResolvePlacement: unexpected error: %v", err)
			}
			if delta := readPlacementResolutionsTotal(t, tc.wantReason) - before; delta != 1 {
				t.Errorf("kropath_placement_resolutions_total{reason=%q} delta = %v, want 1", tc.wantReason, delta)
			}
		})
	}
}

// The NamespaceUnreadable emit site is ResolvePlacement's own Get failure,
// which needs a missing namespace rather than an annotations table entry.
func TestResolvePlacementNamespaceUnreadableIncrementsMetric(t *testing.T) {
	c := testClient(t).Build()

	before := readPlacementResolutionsTotal(t, util.ReasonNamespaceUnreadable)
	if _, _, err := util.ResolvePlacement(context.Background(), c, "nonexistent"); err == nil {
		t.Fatal("ResolvePlacement: expected error for missing namespace, got nil")
	}
	if delta := readPlacementResolutionsTotal(t, util.ReasonNamespaceUnreadable) - before; delta != 1 {
		t.Errorf("kropath_placement_resolutions_total{reason=%q} delta = %v, want 1", util.ReasonNamespaceUnreadable, delta)
	}
}

// spec §2.2: GlobalTierInput is ResolveFamilyPlacement's own emit, on the
// governance-only early return that never reaches ResolvePlacement.
func TestResolveFamilyPlacementGovernanceOnlyIncrementsGlobalTierInput(t *testing.T) {
	ns := namespaceWithAnnotations("governance-only-ns", nil)
	c := testClient(t, ns).Build()

	before := readPlacementResolutionsTotal(t, util.ReasonGlobalTierInput)
	result, err := util.ResolveFamilyPlacement(context.Background(), c, "governance-only-ns", 1, metav1.Now())
	if err != nil {
		t.Fatalf("ResolveFamilyPlacement: unexpected error: %v", err)
	}
	if result.Role != util.RoleGovernanceOnly {
		t.Fatalf("Role = %v, want RoleGovernanceOnly", result.Role)
	}
	if delta := readPlacementResolutionsTotal(t, util.ReasonGlobalTierInput) - before; delta != 1 {
		t.Errorf("kropath_placement_resolutions_total{reason=%q} delta = %v, want 1", util.ReasonGlobalTierInput, delta)
	}
}

// spec §2.2's no-double-count guarantee: ResolveFamilyPlacement calls
// ResolvePlacement internally, so a successful family resolution must
// increment PlacementResolved exactly once, not twice.
func TestResolveFamilyPlacementSuccessIncrementsPlacementResolvedExactlyOnce(t *testing.T) {
	ns := namespaceWithAnnotations("payments-prod", map[string]string{
		util.GlobalConfigNamespaceAnnotation: "platform-config",
		util.OwnerAccountIDAnnotation:        "111122223333",
		util.DefaultRegionAnnotation:         "ap-southeast-2",
	})
	c := testClient(t, ns).Build()

	before := readPlacementResolutionsTotal(t, util.ReasonPlacementResolved)
	result, err := util.ResolveFamilyPlacement(context.Background(), c, "payments-prod", 1, metav1.Now())
	if err != nil {
		t.Fatalf("ResolveFamilyPlacement: unexpected error: %v", err)
	}
	if result.Role != util.RoleResource {
		t.Fatalf("Role = %v, want RoleResource", result.Role)
	}
	if delta := readPlacementResolutionsTotal(t, util.ReasonPlacementResolved) - before; delta != 1 {
		t.Errorf("kropath_placement_resolutions_total{reason=%q} delta = %v, want exactly 1 (no double count)", util.ReasonPlacementResolved, delta)
	}
}

func TestNamespaceUnreadableConditionMirrorsReconciledAndPlacement(t *testing.T) {
	now := metav1.Now()
	readErr := errors.New("etcdserver: request timed out")
	reconciled, placement := util.NamespaceUnreadableCondition("payments-prod", readErr, 3, now)

	if reconciled.Reason != util.ReasonNamespaceUnreadable || placement.Reason != util.ReasonNamespaceUnreadable {
		t.Fatalf("reasons = %q / %q, want both %q", reconciled.Reason, placement.Reason, util.ReasonNamespaceUnreadable)
	}
	if reconciled.Message != placement.Message {
		t.Errorf("messages diverged: %q vs %q", reconciled.Message, placement.Message)
	}
	if reconciled.Status != metav1.ConditionFalse || placement.Status != metav1.ConditionFalse {
		t.Errorf("expected both conditions False, got %v / %v", reconciled.Status, placement.Status)
	}
}
