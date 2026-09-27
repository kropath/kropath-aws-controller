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
	"testing"

	"github.com/kropath/kropath-controller/api/v1alpha1"
	"github.com/kropath/kropath-controller/internal/cascade"
	"github.com/kropath/kropath-controller/internal/reconciler/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

func profileTestClient(t *testing.T, objs ...runtime.Object) *fake.ClientBuilder {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme: %v", err)
	}
	return fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objs...)
}

func s3Config(namespace, name string, blockPublicAccess bool) *v1alpha1.S3Config {
	return &v1alpha1.S3Config{
		TypeMeta:   metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "S3Config"},
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Spec: v1alpha1.S3ConfigSpec{
			Mandatory: cascade.S3ConfigSection{BlockPublicAccess: blockPublicAccess},
		},
	}
}

func TestLoadConfigWithFallthrough_RequestedProfileExists(t *testing.T) {
	c := profileTestClient(t, s3Config("tenant", "pci", true)).Build()

	obj, found, viaFallthrough, err := util.LoadConfigWithFallthrough[v1alpha1.S3Config](
		context.Background(), c, v1alpha1.GroupVersion.WithKind("S3Config"), "tenant", "pci", util.DefaultConfigProfile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found || viaFallthrough {
		t.Fatalf("found=%v viaFallthrough=%v, want found=true viaFallthrough=false", found, viaFallthrough)
	}
	if !obj.Spec.Mandatory.BlockPublicAccess {
		t.Fatalf("expected the requested profile's own config, not the fallthrough target")
	}
}

func TestLoadConfigWithFallthrough_FallsThroughWhenProfileMissing(t *testing.T) {
	c := profileTestClient(t, s3Config("tenant", util.DefaultConfigProfile, true)).Build()

	obj, found, viaFallthrough, err := util.LoadConfigWithFallthrough[v1alpha1.S3Config](
		context.Background(), c, v1alpha1.GroupVersion.WithKind("S3Config"), "tenant", "pci", util.DefaultConfigProfile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found || !viaFallthrough {
		t.Fatalf("found=%v viaFallthrough=%v, want found=true viaFallthrough=true", found, viaFallthrough)
	}
	if !obj.Spec.Mandatory.BlockPublicAccess {
		t.Fatalf("expected the fallthrough target's config to be merged")
	}
}

func TestLoadConfigWithFallthrough_NeitherExists(t *testing.T) {
	c := profileTestClient(t).Build()

	obj, found, viaFallthrough, err := util.LoadConfigWithFallthrough[v1alpha1.S3Config](
		context.Background(), c, v1alpha1.GroupVersion.WithKind("S3Config"), "tenant", "pci", util.DefaultConfigProfile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found || viaFallthrough {
		t.Fatalf("found=%v viaFallthrough=%v, want found=false viaFallthrough=false", found, viaFallthrough)
	}
	if obj == nil {
		t.Fatalf("expected a non-nil zero-value object so callers can merge an empty tier")
	}
}

func TestLoadConfigWithFallthrough_RequestedProfileIsAlreadyDefault(t *testing.T) {
	// configRef == "general-policy" and it does not exist: no second Get should be attempted,
	// and the outcome must still be found=false (not a false fallthrough success).
	c := profileTestClient(t).Build()

	obj, found, viaFallthrough, err := util.LoadConfigWithFallthrough[v1alpha1.S3Config](
		context.Background(), c, v1alpha1.GroupVersion.WithKind("S3Config"), "tenant", util.DefaultConfigProfile, util.DefaultConfigProfile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found || viaFallthrough {
		t.Fatalf("found=%v viaFallthrough=%v, want found=false viaFallthrough=false", found, viaFallthrough)
	}
	if obj == nil {
		t.Fatalf("expected a non-nil zero-value object")
	}
}

// readConfigProfileResolutionsTotal reads
// kropath_config_profile_resolutions_total{family,reason} from the same
// global registry production code registers into. See
// readPlacementResolutionsTotal (namespace_test.go) for why this gathers the
// registry rather than a package-internal var.
func readConfigProfileResolutionsTotal(t *testing.T, family, reason string) float64 {
	t.Helper()
	mfs, err := ctrlmetrics.Registry.Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	for _, mf := range mfs {
		if mf.GetName() != "kropath_config_profile_resolutions_total" {
			continue
		}
		for _, m := range mf.GetMetric() {
			var gotFamily, gotReason string
			for _, lp := range m.GetLabel() {
				switch lp.GetName() {
				case "family":
					gotFamily = lp.GetValue()
				case "reason":
					gotReason = lp.GetValue()
				}
			}
			if gotFamily == family && gotReason == reason {
				return m.GetCounter().GetValue()
			}
		}
	}
	return 0
}

// spec §2.1: ConfigProfileReasons() is the closed set the §10.2.1
// label-value check enumerates against.
func TestConfigProfileReasonsReturnsAllThreeValues(t *testing.T) {
	got := util.ConfigProfileReasons()
	want := []string{util.ReasonProfileFound, util.ReasonProfileFallthrough, util.ReasonProfileUnresolved}
	if len(got) != len(want) {
		t.Fatalf("ConfigProfileReasons() = %v (len %d), want len %d", got, len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ConfigProfileReasons()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// spec §2.6: LoadConfigWithFallthrough increments exactly one of the three
// reasons per call, keyed by the lowercase Kind as the family label, and its
// two error returns emit nothing.
func TestLoadConfigWithFallthrough_IncrementsConfigProfileResolutionsTotal(t *testing.T) {
	const family = "s3config" // strings.ToLower("S3Config")

	t.Run("direct hit increments ProfileFound", func(t *testing.T) {
		c := profileTestClient(t, s3Config("tenant", "pci", true)).Build()
		before := readConfigProfileResolutionsTotal(t, family, util.ReasonProfileFound)

		if _, found, viaFallthrough, err := util.LoadConfigWithFallthrough[v1alpha1.S3Config](
			context.Background(), c, v1alpha1.GroupVersion.WithKind("S3Config"), "tenant", "pci", util.DefaultConfigProfile,
		); err != nil || !found || viaFallthrough {
			t.Fatalf("LoadConfigWithFallthrough: found=%v viaFallthrough=%v err=%v", found, viaFallthrough, err)
		}
		if delta := readConfigProfileResolutionsTotal(t, family, util.ReasonProfileFound) - before; delta != 1 {
			t.Errorf("kropath_config_profile_resolutions_total{family=%q,reason=%q} delta = %v, want 1", family, util.ReasonProfileFound, delta)
		}
	})

	t.Run("fallthrough hit increments ProfileFallthrough", func(t *testing.T) {
		c := profileTestClient(t, s3Config("tenant", util.DefaultConfigProfile, true)).Build()
		before := readConfigProfileResolutionsTotal(t, family, util.ReasonProfileFallthrough)

		if _, found, viaFallthrough, err := util.LoadConfigWithFallthrough[v1alpha1.S3Config](
			context.Background(), c, v1alpha1.GroupVersion.WithKind("S3Config"), "tenant", "pci", util.DefaultConfigProfile,
		); err != nil || !found || !viaFallthrough {
			t.Fatalf("LoadConfigWithFallthrough: found=%v viaFallthrough=%v err=%v", found, viaFallthrough, err)
		}
		if delta := readConfigProfileResolutionsTotal(t, family, util.ReasonProfileFallthrough) - before; delta != 1 {
			t.Errorf("kropath_config_profile_resolutions_total{family=%q,reason=%q} delta = %v, want 1", family, util.ReasonProfileFallthrough, delta)
		}
	})

	t.Run("neither exists increments ProfileUnresolved", func(t *testing.T) {
		c := profileTestClient(t).Build()
		before := readConfigProfileResolutionsTotal(t, family, util.ReasonProfileUnresolved)

		if _, found, viaFallthrough, err := util.LoadConfigWithFallthrough[v1alpha1.S3Config](
			context.Background(), c, v1alpha1.GroupVersion.WithKind("S3Config"), "tenant", "pci", util.DefaultConfigProfile,
		); err != nil || found || viaFallthrough {
			t.Fatalf("LoadConfigWithFallthrough: found=%v viaFallthrough=%v err=%v", found, viaFallthrough, err)
		}
		if delta := readConfigProfileResolutionsTotal(t, family, util.ReasonProfileUnresolved) - before; delta != 1 {
			t.Errorf("kropath_config_profile_resolutions_total{family=%q,reason=%q} delta = %v, want 1", family, util.ReasonProfileUnresolved, delta)
		}
	})

	t.Run("Get error emits nothing", func(t *testing.T) {
		beforeFound := readConfigProfileResolutionsTotal(t, family, util.ReasonProfileFound)
		beforeFallthrough := readConfigProfileResolutionsTotal(t, family, util.ReasonProfileFallthrough)
		beforeUnresolved := readConfigProfileResolutionsTotal(t, family, util.ReasonProfileUnresolved)

		// A client built against a scheme with no v1alpha1 types registered
		// makes Get fail with "no kind is registered for the type ... in
		// scheme" -- a non-NotFound error, exercising LoadConfigWithFallthrough's
		// error return without depending on any specific error text.
		c := fake.NewClientBuilder().WithScheme(runtime.NewScheme()).Build()
		_, _, _, err := util.LoadConfigWithFallthrough[v1alpha1.S3Config](
			context.Background(), c, v1alpha1.GroupVersion.WithKind("S3Config"), "tenant", "pci", util.DefaultConfigProfile)
		if err == nil {
			t.Fatal("LoadConfigWithFallthrough: expected an error for an unregistered scheme")
		}

		if got := readConfigProfileResolutionsTotal(t, family, util.ReasonProfileFound); got != beforeFound {
			t.Errorf("ProfileFound changed on error: before %v, after %v", beforeFound, got)
		}
		if got := readConfigProfileResolutionsTotal(t, family, util.ReasonProfileFallthrough); got != beforeFallthrough {
			t.Errorf("ProfileFallthrough changed on error: before %v, after %v", beforeFallthrough, got)
		}
		if got := readConfigProfileResolutionsTotal(t, family, util.ReasonProfileUnresolved); got != beforeUnresolved {
			t.Errorf("ProfileUnresolved changed on error: before %v, after %v", beforeUnresolved, got)
		}
	})
}

func TestConfigProfileResolvedCondition(t *testing.T) {
	now := metav1.Now()

	cases := []struct {
		name           string
		found          bool
		viaFallthrough bool
		wantStatus     metav1.ConditionStatus
		wantReason     string
	}{
		{"direct hit", true, false, metav1.ConditionTrue, "ProfileFound"},
		{"fallthrough hit", true, true, metav1.ConditionTrue, "ProfileFallthrough"},
		{"unresolved", false, false, metav1.ConditionFalse, "ProfileUnresolved"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cond := util.ConfigProfileResolvedCondition("pci", tc.found, tc.viaFallthrough, 3, now)
			if cond.Type != util.ConfigProfileResolvedConditionType {
				t.Fatalf("Type = %q, want %q", cond.Type, util.ConfigProfileResolvedConditionType)
			}
			if cond.Status != tc.wantStatus {
				t.Fatalf("Status = %q, want %q", cond.Status, tc.wantStatus)
			}
			if cond.Reason != tc.wantReason {
				t.Fatalf("Reason = %q, want %q", cond.Reason, tc.wantReason)
			}
			if cond.ObservedGeneration != 3 {
				t.Fatalf("ObservedGeneration = %d, want 3", cond.ObservedGeneration)
			}
		})
	}
}
