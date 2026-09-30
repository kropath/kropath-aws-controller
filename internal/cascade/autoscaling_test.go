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

package cascade_test

import (
	"testing"

	"github.com/kropath/kropath-aws-controller/internal/cascade"
)

// zeroKropathAS is a zero-value AutoScalingKropathSection (absent source).
var zeroKropathAS = cascade.AutoScalingKropathSection{}

// zeroASCfg is a zero-value AutoScalingConfigSection (absent source).
var zeroASCfg = cascade.AutoScalingConfigSection{}

// mergeASAll calls MergeAutoScalingCascade with all eight inputs.
func mergeASAll(
	globalKropathMandatory,
	localKropathMandatory cascade.AutoScalingKropathSection,
	globalASCfgMandatory,
	localASCfgMandatory,
	localASCfgDefaults,
	globalASCfgDefaults cascade.AutoScalingConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.AutoScalingKropathSection,
) cascade.EffectiveAutoScalingConfig {
	return cascade.MergeAutoScalingCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalASCfgMandatory,
		localASCfgMandatory,
		localASCfgDefaults,
		globalASCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeAutoScalingCascade_AC_C1 — globalKropathConfig.mandatory.autoscaling.
// newInstancesProtectedFromScaleIn=true at level 1 propagates to
// effectiveConfig.mandatory.newInstancesProtectedFromScaleIn.
func TestMergeAutoScalingCascade_AC_C1(t *testing.T) {
	got := mergeASAll(
		cascade.AutoScalingKropathSection{NewInstancesProtectedFromScaleIn: true}, // level 1
		zeroKropathAS,
		zeroASCfg,
		zeroASCfg,
		zeroASCfg,
		zeroASCfg,
		zeroKropathAS,
		zeroKropathAS,
	)

	if !got.Mandatory.NewInstancesProtectedFromScaleIn {
		t.Error("AC-C1: mandatory.newInstancesProtectedFromScaleIn should be true when set at level 1")
	}
	if got.Defaults.NewInstancesProtectedFromScaleIn {
		t.Error("AC-C1: defaults.newInstancesProtectedFromScaleIn must not bleed from mandatory")
	}
}

// TestMergeAutoScalingCascade_AC_C2 — level-1 KropathConfig mandatory wins over
// level-3 AutoScalingConfig mandatory for newInstancesProtectedFromScaleIn.
func TestMergeAutoScalingCascade_AC_C2(t *testing.T) {
	got := mergeASAll(
		cascade.AutoScalingKropathSection{NewInstancesProtectedFromScaleIn: true}, // level 1
		zeroKropathAS,
		cascade.AutoScalingConfigSection{NewInstancesProtectedFromScaleIn: false}, // level 3 (zero = not enforced)
		zeroASCfg,
		zeroASCfg,
		zeroASCfg,
		zeroKropathAS,
		zeroKropathAS,
	)

	if !got.Mandatory.NewInstancesProtectedFromScaleIn {
		t.Error("AC-C2: level-1 KropathConfig must win; mandatory.newInstancesProtectedFromScaleIn should be true")
	}
}

// TestMergeAutoScalingCascade_AC_C3 — only globalASCfg.defaults.capacityRebalance=true;
// mandatory must be false, defaults must be true.
func TestMergeAutoScalingCascade_AC_C3(t *testing.T) {
	got := mergeASAll(
		zeroKropathAS,
		zeroKropathAS,
		zeroASCfg,
		zeroASCfg,
		zeroASCfg,
		cascade.AutoScalingConfigSection{CapacityRebalance: true}, // level 7 global defaults
		zeroKropathAS,
		zeroKropathAS,
	)

	if got.Mandatory.CapacityRebalance {
		t.Error("AC-C3: mandatory.capacityRebalance should be false when only defaults set")
	}
	if !got.Defaults.CapacityRebalance {
		t.Error("AC-C3: defaults.capacityRebalance should be true")
	}
}

// TestMergeAutoScalingCascade_AC_C4 — globalASCfg.mandatory.healthCheckType="EC2,ELB"
// (level 3) wins over localASCfg.mandatory.healthCheckType="EC2" (level 4).
func TestMergeAutoScalingCascade_AC_C4(t *testing.T) {
	got := mergeASAll(
		zeroKropathAS,
		zeroKropathAS,
		cascade.AutoScalingConfigSection{HealthCheckType: "EC2,ELB"}, // level 3
		cascade.AutoScalingConfigSection{HealthCheckType: "EC2"},     // level 4
		zeroASCfg,
		zeroASCfg,
		zeroKropathAS,
		zeroKropathAS,
	)

	if got.Mandatory.HealthCheckType != "EC2,ELB" {
		t.Errorf("AC-C4: mandatory.healthCheckType = %q, want EC2,ELB", got.Mandatory.HealthCheckType)
	}
}

// TestMergeAutoScalingCascade_AC_C5 — only globalASCfg.defaults.healthCheckGracePeriod=600;
// mandatory must be 0, defaults must be 600.
func TestMergeAutoScalingCascade_AC_C5(t *testing.T) {
	got := mergeASAll(
		zeroKropathAS,
		zeroKropathAS,
		zeroASCfg,
		zeroASCfg,
		zeroASCfg,
		cascade.AutoScalingConfigSection{HealthCheckGracePeriod: 600}, // level 7
		zeroKropathAS,
		zeroKropathAS,
	)

	if got.Mandatory.HealthCheckGracePeriod != 0 {
		t.Errorf("AC-C5: mandatory.healthCheckGracePeriod = %d, want 0", got.Mandatory.HealthCheckGracePeriod)
	}
	if got.Defaults.HealthCheckGracePeriod != 600 {
		t.Errorf("AC-C5: defaults.healthCheckGracePeriod = %d, want 600", got.Defaults.HealthCheckGracePeriod)
	}
}

// TestMergeAutoScalingCascade_AC_C6 — globalASCfg.mandatory.maxInstanceLifetime=604800
// propagates to effectiveConfig.mandatory.maxInstanceLifetime.
func TestMergeAutoScalingCascade_AC_C6(t *testing.T) {
	got := mergeASAll(
		zeroKropathAS,
		zeroKropathAS,
		cascade.AutoScalingConfigSection{MaxInstanceLifetime: 604800}, // level 3
		zeroASCfg,
		zeroASCfg,
		zeroASCfg,
		zeroKropathAS,
		zeroKropathAS,
	)

	if got.Mandatory.MaxInstanceLifetime != 604800 {
		t.Errorf("AC-C6: mandatory.maxInstanceLifetime = %d, want 604800", got.Mandatory.MaxInstanceLifetime)
	}
}

// TestMergeAutoScalingCascade_AC_C7 — globalASCfg.mandatory.namingTemplate="{namespace}-{name}"
// propagates to effectiveConfig.mandatory.namingTemplate.
func TestMergeAutoScalingCascade_AC_C7(t *testing.T) {
	got := mergeASAll(
		zeroKropathAS,
		zeroKropathAS,
		cascade.AutoScalingConfigSection{NamingTemplate: "{namespace}-{name}"}, // level 3
		zeroASCfg,
		zeroASCfg,
		zeroASCfg,
		zeroKropathAS,
		zeroKropathAS,
	)

	if got.Mandatory.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("AC-C7: mandatory.namingTemplate = %q, want {namespace}-{name}", got.Mandatory.NamingTemplate)
	}
}

// TestMergeAutoScalingCascade_AC_C8 — KropathConfig.mandatory.tags and
// AutoScalingConfig.mandatory.tags are union-merged into effectiveConfig.mandatory.tags,
// with KropathConfig winning on key conflicts.
func TestMergeAutoScalingCascade_AC_C8(t *testing.T) {
	got := mergeASAll(
		cascade.AutoScalingKropathSection{Tags: map[string]string{"env": "mandatory", "owner": "platform"}}, // level 1
		zeroKropathAS,
		cascade.AutoScalingConfigSection{Tags: map[string]string{"env": "cfg", "cost-centre": "infra"}}, // level 3
		zeroASCfg,
		zeroASCfg,
		zeroASCfg,
		zeroKropathAS,
		zeroKropathAS,
	)

	if got.Mandatory.Tags["env"] != "mandatory" {
		t.Errorf("AC-C8: level-1 KropathConfig must win on key conflict; tags[env] = %q, want mandatory", got.Mandatory.Tags["env"])
	}
	if got.Mandatory.Tags["owner"] != "platform" {
		t.Errorf("AC-C8: tags[owner] = %q, want platform", got.Mandatory.Tags["owner"])
	}
	if got.Mandatory.Tags["cost-centre"] != "infra" {
		t.Errorf("AC-C8: tags[cost-centre] = %q, want infra", got.Mandatory.Tags["cost-centre"])
	}
}

// TestMergeAutoScalingCascade_AllAbsent — when all sources are zero, effectiveConfig
// fields are all zero (permissive; no governance enforced).
func TestMergeAutoScalingCascade_AllAbsent(t *testing.T) {
	got := mergeASAll(
		zeroKropathAS, zeroKropathAS,
		zeroASCfg, zeroASCfg, zeroASCfg, zeroASCfg,
		zeroKropathAS, zeroKropathAS,
	)

	if got.Mandatory.NewInstancesProtectedFromScaleIn {
		t.Error("all-absent: mandatory.newInstancesProtectedFromScaleIn should be false")
	}
	if got.Mandatory.CapacityRebalance {
		t.Error("all-absent: mandatory.capacityRebalance should be false")
	}
	if got.Mandatory.HealthCheckType != "" {
		t.Errorf("all-absent: mandatory.healthCheckType = %q, want empty", got.Mandatory.HealthCheckType)
	}
	if got.Mandatory.HealthCheckGracePeriod != 0 {
		t.Errorf("all-absent: mandatory.healthCheckGracePeriod = %d, want 0", got.Mandatory.HealthCheckGracePeriod)
	}
	if got.Mandatory.MaxInstanceLifetime != 0 {
		t.Errorf("all-absent: mandatory.maxInstanceLifetime = %d, want 0", got.Mandatory.MaxInstanceLifetime)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("all-absent: mandatory.namingTemplate = %q, want empty", got.Mandatory.NamingTemplate)
	}
	if got.Defaults.NewInstancesProtectedFromScaleIn {
		t.Error("all-absent: defaults.newInstancesProtectedFromScaleIn should be false")
	}
	if got.Defaults.CapacityRebalance {
		t.Error("all-absent: defaults.capacityRebalance should be false")
	}
}

// TestMergeAutoScalingCascade_MandatoryCascadeOrder — verifies mandatory priority order
// for newInstancesProtectedFromScaleIn (level 1 > 2 > 3 > 4).
func TestMergeAutoScalingCascade_MandatoryCascadeOrder(t *testing.T) {
	cases := []struct {
		name                   string
		globalKropathMandatory cascade.AutoScalingKropathSection
		localKropathMandatory  cascade.AutoScalingKropathSection
		globalASCfgMandatory   cascade.AutoScalingConfigSection
		localASCfgMandatory    cascade.AutoScalingConfigSection
		wantScaleIn            bool
	}{
		{
			name:                   "level1-sets-true",
			globalKropathMandatory: cascade.AutoScalingKropathSection{NewInstancesProtectedFromScaleIn: true},
			wantScaleIn:            true,
		},
		{
			name:                  "level2-sets-true-when-1-absent",
			localKropathMandatory: cascade.AutoScalingKropathSection{NewInstancesProtectedFromScaleIn: true},
			wantScaleIn:           true,
		},
		{
			name:                 "level3-sets-true-when-1-2-absent",
			globalASCfgMandatory: cascade.AutoScalingConfigSection{NewInstancesProtectedFromScaleIn: true},
			wantScaleIn:          true,
		},
		{
			name:                "level4-sets-true-when-1-2-3-absent",
			localASCfgMandatory: cascade.AutoScalingConfigSection{NewInstancesProtectedFromScaleIn: true},
			wantScaleIn:         true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeASAll(
				tc.globalKropathMandatory,
				tc.localKropathMandatory,
				tc.globalASCfgMandatory,
				tc.localASCfgMandatory,
				zeroASCfg, zeroASCfg,
				zeroKropathAS, zeroKropathAS,
			)
			if got.Mandatory.NewInstancesProtectedFromScaleIn != tc.wantScaleIn {
				t.Errorf("newInstancesProtectedFromScaleIn = %v, want %v",
					got.Mandatory.NewInstancesProtectedFromScaleIn, tc.wantScaleIn)
			}
		})
	}
}

// TestMergeAutoScalingCascade_DefaultsCascadeOrder — verifies defaults priority order
// for capacityRebalance (level 6 > 7 > 8 > 9).
func TestMergeAutoScalingCascade_DefaultsCascadeOrder(t *testing.T) {
	cases := []struct {
		name                  string
		localASCfgDefaults    cascade.AutoScalingConfigSection
		globalASCfgDefaults   cascade.AutoScalingConfigSection
		localKropathDefaults  cascade.AutoScalingKropathSection
		globalKropathDefaults cascade.AutoScalingKropathSection
		wantCapacity          bool
	}{
		{
			name:               "level6-wins",
			localASCfgDefaults: cascade.AutoScalingConfigSection{CapacityRebalance: true},
			wantCapacity:       true,
		},
		{
			name:                "level7-wins-when-6-absent",
			globalASCfgDefaults: cascade.AutoScalingConfigSection{CapacityRebalance: true},
			wantCapacity:        true,
		},
		{
			name:                 "level8-wins-when-6-7-absent",
			localKropathDefaults: cascade.AutoScalingKropathSection{CapacityRebalance: true},
			wantCapacity:         true,
		},
		{
			name:                  "level9-fallback",
			globalKropathDefaults: cascade.AutoScalingKropathSection{CapacityRebalance: true},
			wantCapacity:          true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeASAll(
				zeroKropathAS, zeroKropathAS,
				zeroASCfg, zeroASCfg,
				tc.localASCfgDefaults,
				tc.globalASCfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
			)
			if got.Defaults.CapacityRebalance != tc.wantCapacity {
				t.Errorf("defaults.capacityRebalance = %v, want %v",
					got.Defaults.CapacityRebalance, tc.wantCapacity)
			}
		})
	}
}

// TestMergeAutoScalingCascade_MandatoryIsolatedFromDefaults — mandatory fields must not
// bleed into defaults and vice versa.
func TestMergeAutoScalingCascade_MandatoryIsolatedFromDefaults(t *testing.T) {
	got := mergeASAll(
		cascade.AutoScalingKropathSection{NewInstancesProtectedFromScaleIn: true, CapacityRebalance: true},
		zeroKropathAS,
		cascade.AutoScalingConfigSection{HealthCheckType: "EC2,ELB", HealthCheckGracePeriod: 600},
		zeroASCfg,
		cascade.AutoScalingConfigSection{HealthCheckType: "EC2", HealthCheckGracePeriod: 300}, // level 6
		zeroASCfg,
		zeroKropathAS,
		zeroKropathAS,
	)

	if !got.Mandatory.NewInstancesProtectedFromScaleIn {
		t.Error("mandatory.newInstancesProtectedFromScaleIn should be true")
	}
	if !got.Mandatory.CapacityRebalance {
		t.Error("mandatory.capacityRebalance should be true")
	}
	if got.Mandatory.HealthCheckType != "EC2,ELB" {
		t.Errorf("mandatory.healthCheckType = %q, want EC2,ELB", got.Mandatory.HealthCheckType)
	}
	if got.Mandatory.HealthCheckGracePeriod != 600 {
		t.Errorf("mandatory.healthCheckGracePeriod = %d, want 600", got.Mandatory.HealthCheckGracePeriod)
	}
	if got.Defaults.NewInstancesProtectedFromScaleIn {
		t.Error("defaults.newInstancesProtectedFromScaleIn must not bleed from mandatory")
	}
	if got.Defaults.CapacityRebalance {
		t.Error("defaults.capacityRebalance must not bleed from mandatory")
	}
	if got.Defaults.HealthCheckType != "EC2" {
		t.Errorf("defaults.healthCheckType = %q, want EC2", got.Defaults.HealthCheckType)
	}
	if got.Defaults.HealthCheckGracePeriod != 300 {
		t.Errorf("defaults.healthCheckGracePeriod = %d, want 300", got.Defaults.HealthCheckGracePeriod)
	}
}

// TestMergeAutoScalingCascade_SyncedLabels — syncedLabels merge from AutoScalingConfig
// levels only; globalASCfg wins on key conflicts in mandatory tier.
func TestMergeAutoScalingCascade_SyncedLabels(t *testing.T) {
	got := mergeASAll(
		zeroKropathAS,
		zeroKropathAS,
		cascade.AutoScalingConfigSection{SyncedLabels: map[string]string{"tier": "global", "team": "platform"}}, // level 3
		cascade.AutoScalingConfigSection{SyncedLabels: map[string]string{"tier": "local"}},                      // level 4
		zeroASCfg,
		zeroASCfg,
		zeroKropathAS,
		zeroKropathAS,
	)

	if got.Mandatory.SyncedLabels["tier"] != "global" {
		t.Errorf("mandatory.syncedLabels[tier] = %q, want global (level-3 wins)", got.Mandatory.SyncedLabels["tier"])
	}
	if got.Mandatory.SyncedLabels["team"] != "platform" {
		t.Errorf("mandatory.syncedLabels[team] = %q, want platform", got.Mandatory.SyncedLabels["team"])
	}
}

// TestMergeAutoScalingCascade_DefaultsTagUnionMerge — defaults tags union from all
// four defaults levels; level 6 wins on key conflicts.
func TestMergeAutoScalingCascade_DefaultsTagUnionMerge(t *testing.T) {
	got := mergeASAll(
		zeroKropathAS,
		zeroKropathAS,
		zeroASCfg,
		zeroASCfg,
		cascade.AutoScalingConfigSection{Tags: map[string]string{"env": "local-cfg", "owner": "app"}}, // level 6
		cascade.AutoScalingConfigSection{Tags: map[string]string{"env": "global-cfg"}},                // level 7
		cascade.AutoScalingKropathSection{Tags: map[string]string{"env": "local-kpc"}},               // level 8
		cascade.AutoScalingKropathSection{Tags: map[string]string{"env": "global-kpc", "dept": "eng"}}, // level 9
	)

	if got.Defaults.Tags["env"] != "local-cfg" {
		t.Errorf("defaults.tags[env] = %q, want local-cfg (level-6 wins)", got.Defaults.Tags["env"])
	}
	if got.Defaults.Tags["owner"] != "app" {
		t.Errorf("defaults.tags[owner] = %q, want app", got.Defaults.Tags["owner"])
	}
	if got.Defaults.Tags["dept"] != "eng" {
		t.Errorf("defaults.tags[dept] = %q, want eng (from level-9)", got.Defaults.Tags["dept"])
	}
}
