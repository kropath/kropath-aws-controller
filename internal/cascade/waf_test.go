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

// zeroKropathWAF is a zero-value WAFKropathSection (absent source).
var zeroKropathWAF = cascade.WAFKropathSection{}

// zeroWAFCfg is a zero-value WAFConfigSection (absent source).
var zeroWAFCfg = cascade.WAFConfigSection{}

// mergeWAFAll calls MergeWAFCascade with all eight inputs.
func mergeWAFAll(
	globalKropathMandatory,
	localKropathMandatory cascade.WAFKropathSection,
	globalWAFCfgMandatory,
	localWAFCfgMandatory,
	localWAFCfgDefaults,
	globalWAFCfgDefaults cascade.WAFConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.WAFKropathSection,
) cascade.EffectiveWAFConfig {
	return cascade.MergeWAFCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalWAFCfgMandatory,
		localWAFCfgMandatory,
		localWAFCfgDefaults,
		globalWAFCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeWAFCascade_AC1 — globalKropathConfig.mandatory.waf.scope at level 1
// propagates to effCfg.mandatory.scope.
func TestMergeWAFCascade_AC1(t *testing.T) {
	got := mergeWAFAll(
		cascade.WAFKropathSection{Scope: "REGIONAL"}, // level 1
		zeroKropathWAF,
		zeroWAFCfg,
		zeroWAFCfg,
		zeroWAFCfg,
		zeroWAFCfg,
		zeroKropathWAF,
		zeroKropathWAF,
	)

	if got.Mandatory.Scope != "REGIONAL" {
		t.Errorf("AC-1: mandatory.scope = %q, want REGIONAL (level 1 wins)", got.Mandatory.Scope)
	}
	if got.Defaults.Scope != "" {
		t.Errorf("AC-1: defaults.scope = %q, must not bleed from mandatory", got.Defaults.Scope)
	}
}

// TestMergeWAFCascade_AC2 — globalWAFConfig.mandatory.scope at level 3 wins
// when levels 1-2 are empty.
func TestMergeWAFCascade_AC2(t *testing.T) {
	got := mergeWAFAll(
		zeroKropathWAF,
		zeroKropathWAF,
		cascade.WAFConfigSection{Scope: "CLOUDFRONT"}, // level 3
		zeroWAFCfg,
		zeroWAFCfg,
		zeroWAFCfg,
		zeroKropathWAF,
		zeroKropathWAF,
	)

	if got.Mandatory.Scope != "CLOUDFRONT" {
		t.Errorf("AC-2: mandatory.scope = %q, want CLOUDFRONT (level 3 wins when 1-2 empty)", got.Mandatory.Scope)
	}
}

// TestMergeWAFCascade_AC3 — localWAFConfig.defaults.scope at level 6 propagates;
// mandatory stays empty.
func TestMergeWAFCascade_AC3(t *testing.T) {
	got := mergeWAFAll(
		zeroKropathWAF,
		zeroKropathWAF,
		zeroWAFCfg,
		zeroWAFCfg,
		cascade.WAFConfigSection{Scope: "REGIONAL"}, // level 6
		zeroWAFCfg,
		zeroKropathWAF,
		zeroKropathWAF,
	)

	if got.Mandatory.Scope != "" {
		t.Errorf("AC-3: mandatory.scope = %q, want empty", got.Mandatory.Scope)
	}
	if got.Defaults.Scope != "REGIONAL" {
		t.Errorf("AC-3: defaults.scope = %q, want REGIONAL (level 6)", got.Defaults.Scope)
	}
}

// TestMergeWAFCascade_AC4 — globalKropathConfig.mandatory.waf.cloudWatchMetricsEnabled=true
// at level 1 propagates to effCfg.mandatory.cloudWatchMetricsEnabled.
func TestMergeWAFCascade_AC4(t *testing.T) {
	got := mergeWAFAll(
		cascade.WAFKropathSection{CloudWatchMetricsEnabled: true}, // level 1
		zeroKropathWAF,
		zeroWAFCfg,
		zeroWAFCfg,
		zeroWAFCfg,
		zeroWAFCfg,
		zeroKropathWAF,
		zeroKropathWAF,
	)

	if !got.Mandatory.CloudWatchMetricsEnabled {
		t.Errorf("AC-4: mandatory.cloudWatchMetricsEnabled = false, want true (level 1 wins)")
	}
	if got.Defaults.CloudWatchMetricsEnabled {
		t.Errorf("AC-4: defaults.cloudWatchMetricsEnabled must not bleed from mandatory")
	}
}

// TestMergeWAFCascade_AC5 — globalWAFConfig.defaults.sampledRequestsEnabled=true
// at level 7 propagates to effCfg.defaults.sampledRequestsEnabled.
func TestMergeWAFCascade_AC5(t *testing.T) {
	got := mergeWAFAll(
		zeroKropathWAF,
		zeroKropathWAF,
		zeroWAFCfg,
		zeroWAFCfg,
		zeroWAFCfg,
		cascade.WAFConfigSection{SampledRequestsEnabled: true}, // level 7
		zeroKropathWAF,
		zeroKropathWAF,
	)

	if got.Mandatory.SampledRequestsEnabled {
		t.Errorf("AC-5: mandatory.sampledRequestsEnabled must not bleed from defaults")
	}
	if !got.Defaults.SampledRequestsEnabled {
		t.Errorf("AC-5: defaults.sampledRequestsEnabled = false, want true (level 7)")
	}
}

// TestMergeWAFCascade_AC6 — globalWAFConfig.mandatory.namingTemplate at level 3
// propagates to effCfg.mandatory.namingTemplate.
func TestMergeWAFCascade_AC6(t *testing.T) {
	got := mergeWAFAll(
		zeroKropathWAF,
		zeroKropathWAF,
		cascade.WAFConfigSection{NamingTemplate: "corp-{namespace}-{name}"}, // level 3
		zeroWAFCfg,
		zeroWAFCfg,
		zeroWAFCfg,
		zeroKropathWAF,
		zeroKropathWAF,
	)

	if got.Mandatory.NamingTemplate != "corp-{namespace}-{name}" {
		t.Errorf("AC-6: mandatory.namingTemplate = %q, want corp-{namespace}-{name} (level 3)", got.Mandatory.NamingTemplate)
	}
}

// TestMergeWAFCascade_AC7 — KropathConfig.mandatory.tags and WAFConfig.mandatory.tags
// are union-merged into effCfg.mandatory.tags.
func TestMergeWAFCascade_AC7(t *testing.T) {
	got := mergeWAFAll(
		cascade.WAFKropathSection{Tags: map[string]string{"cost-centre": "infra"}},   // level 1
		zeroKropathWAF,
		zeroWAFCfg,
		cascade.WAFConfigSection{Tags: map[string]string{"waf-class": "production"}}, // level 4
		zeroWAFCfg,
		zeroWAFCfg,
		zeroKropathWAF,
		zeroKropathWAF,
	)

	if got.Mandatory.Tags["cost-centre"] != "infra" {
		t.Errorf("AC-7: mandatory.tags[cost-centre] = %q, want infra", got.Mandatory.Tags["cost-centre"])
	}
	if got.Mandatory.Tags["waf-class"] != "production" {
		t.Errorf("AC-7: mandatory.tags[waf-class] = %q, want production", got.Mandatory.Tags["waf-class"])
	}
}

// TestMergeWAFCascade_AC8 — WAFConfig.mandatory.syncedLabels at level 3 propagates
// to effCfg.mandatory.syncedLabels.
func TestMergeWAFCascade_AC8(t *testing.T) {
	got := mergeWAFAll(
		zeroKropathWAF,
		zeroKropathWAF,
		cascade.WAFConfigSection{SyncedLabels: map[string]string{"data-class": "internal"}}, // level 3
		zeroWAFCfg,
		zeroWAFCfg,
		zeroWAFCfg,
		zeroKropathWAF,
		zeroKropathWAF,
	)

	if got.Mandatory.SyncedLabels["data-class"] != "internal" {
		t.Errorf("AC-8: mandatory.syncedLabels[data-class] = %q, want internal", got.Mandatory.SyncedLabels["data-class"])
	}
}

// TestMergeWAFCascade_MandatoryPriorityOrder — verifies mandatory priority order
// for scope: level 1 > 2 > 3 > 4.
func TestMergeWAFCascade_MandatoryPriorityOrder(t *testing.T) {
	cases := []struct {
		name                   string
		globalKropathMandatory cascade.WAFKropathSection
		localKropathMandatory  cascade.WAFKropathSection
		globalWAFCfgMandatory  cascade.WAFConfigSection
		localWAFCfgMandatory   cascade.WAFConfigSection
		wantScope              string
	}{
		{
			name:                   "level1-wins",
			globalKropathMandatory: cascade.WAFKropathSection{Scope: "level1"},
			localKropathMandatory:  cascade.WAFKropathSection{Scope: "level2"},
			globalWAFCfgMandatory:  cascade.WAFConfigSection{Scope: "level3"},
			localWAFCfgMandatory:   cascade.WAFConfigSection{Scope: "level4"},
			wantScope:              "level1",
		},
		{
			name:                   "level2-wins-when-1-absent",
			globalKropathMandatory: zeroKropathWAF,
			localKropathMandatory:  cascade.WAFKropathSection{Scope: "level2"},
			globalWAFCfgMandatory:  cascade.WAFConfigSection{Scope: "level3"},
			localWAFCfgMandatory:   cascade.WAFConfigSection{Scope: "level4"},
			wantScope:              "level2",
		},
		{
			name:                   "level3-wins-when-1-2-absent",
			globalKropathMandatory: zeroKropathWAF,
			localKropathMandatory:  zeroKropathWAF,
			globalWAFCfgMandatory:  cascade.WAFConfigSection{Scope: "level3"},
			localWAFCfgMandatory:   cascade.WAFConfigSection{Scope: "level4"},
			wantScope:              "level3",
		},
		{
			name:                   "level4-wins-when-1-2-3-absent",
			globalKropathMandatory: zeroKropathWAF,
			localKropathMandatory:  zeroKropathWAF,
			globalWAFCfgMandatory:  zeroWAFCfg,
			localWAFCfgMandatory:   cascade.WAFConfigSection{Scope: "level4"},
			wantScope:              "level4",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeWAFAll(
				tc.globalKropathMandatory,
				tc.localKropathMandatory,
				tc.globalWAFCfgMandatory,
				tc.localWAFCfgMandatory,
				zeroWAFCfg,
				zeroWAFCfg,
				zeroKropathWAF,
				zeroKropathWAF,
			)
			if got.Mandatory.Scope != tc.wantScope {
				t.Errorf("mandatory.scope = %q, want %q", got.Mandatory.Scope, tc.wantScope)
			}
		})
	}
}

// TestMergeWAFCascade_DefaultsPriorityOrder — verifies defaults priority order
// for scope: level 6 > 7 > 8 > 9.
func TestMergeWAFCascade_DefaultsPriorityOrder(t *testing.T) {
	cases := []struct {
		name                  string
		localWAFCfgDefaults   cascade.WAFConfigSection
		globalWAFCfgDefaults  cascade.WAFConfigSection
		localKropathDefaults  cascade.WAFKropathSection
		globalKropathDefaults cascade.WAFKropathSection
		wantScope             string
	}{
		{
			name:                  "level6-wins",
			localWAFCfgDefaults:   cascade.WAFConfigSection{Scope: "level6"},
			globalWAFCfgDefaults:  cascade.WAFConfigSection{Scope: "level7"},
			localKropathDefaults:  cascade.WAFKropathSection{Scope: "level8"},
			globalKropathDefaults: cascade.WAFKropathSection{Scope: "level9"},
			wantScope:             "level6",
		},
		{
			name:                  "level7-wins-when-6-absent",
			localWAFCfgDefaults:   zeroWAFCfg,
			globalWAFCfgDefaults:  cascade.WAFConfigSection{Scope: "level7"},
			localKropathDefaults:  cascade.WAFKropathSection{Scope: "level8"},
			globalKropathDefaults: cascade.WAFKropathSection{Scope: "level9"},
			wantScope:             "level7",
		},
		{
			name:                  "level8-wins-when-6-7-absent",
			localWAFCfgDefaults:   zeroWAFCfg,
			globalWAFCfgDefaults:  zeroWAFCfg,
			localKropathDefaults:  cascade.WAFKropathSection{Scope: "level8"},
			globalKropathDefaults: cascade.WAFKropathSection{Scope: "level9"},
			wantScope:             "level8",
		},
		{
			name:                  "level9-wins-when-6-7-8-absent",
			localWAFCfgDefaults:   zeroWAFCfg,
			globalWAFCfgDefaults:  zeroWAFCfg,
			localKropathDefaults:  zeroKropathWAF,
			globalKropathDefaults: cascade.WAFKropathSection{Scope: "level9"},
			wantScope:             "level9",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeWAFAll(
				zeroKropathWAF,
				zeroKropathWAF,
				zeroWAFCfg,
				zeroWAFCfg,
				tc.localWAFCfgDefaults,
				tc.globalWAFCfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
			)
			if got.Defaults.Scope != tc.wantScope {
				t.Errorf("defaults.scope = %q, want %q", got.Defaults.Scope, tc.wantScope)
			}
		})
	}
}

// TestMergeWAFCascade_TagsKeyConflict — on key conflict, lower level number wins.
// Level 1 (globalKropathMandatory) wins over level 4 (localWAFCfgMandatory).
func TestMergeWAFCascade_TagsKeyConflict(t *testing.T) {
	got := mergeWAFAll(
		cascade.WAFKropathSection{Tags: map[string]string{"env": "org-level"}},   // level 1
		zeroKropathWAF,
		zeroWAFCfg,
		cascade.WAFConfigSection{Tags: map[string]string{"env": "config-level"}}, // level 4
		zeroWAFCfg,
		zeroWAFCfg,
		zeroKropathWAF,
		zeroKropathWAF,
	)

	if got.Mandatory.Tags["env"] != "org-level" {
		t.Errorf("tag-conflict: mandatory.tags[env] = %q, want org-level (level 1 wins over level 4)", got.Mandatory.Tags["env"])
	}
}

// TestMergeWAFCascade_SyncedLabelsUnion — SyncedLabels from global (L3) and
// local (L4) WAFConfig mandatory tiers are union-merged; L3 wins on key conflict.
func TestMergeWAFCascade_SyncedLabelsUnion(t *testing.T) {
	got := mergeWAFAll(
		zeroKropathWAF,
		zeroKropathWAF,
		cascade.WAFConfigSection{SyncedLabels: map[string]string{"tier": "global", "data-class": "internal"}}, // level 3
		cascade.WAFConfigSection{SyncedLabels: map[string]string{"tier": "local", "env": "prod"}},             // level 4
		zeroWAFCfg,
		zeroWAFCfg,
		zeroKropathWAF,
		zeroKropathWAF,
	)

	if got.Mandatory.SyncedLabels["data-class"] != "internal" {
		t.Errorf("synced-labels-union: mandatory.syncedLabels[data-class] = %q, want internal", got.Mandatory.SyncedLabels["data-class"])
	}
	if got.Mandatory.SyncedLabels["env"] != "prod" {
		t.Errorf("synced-labels-union: mandatory.syncedLabels[env] = %q, want prod", got.Mandatory.SyncedLabels["env"])
	}
	if got.Mandatory.SyncedLabels["tier"] != "global" {
		t.Errorf("synced-labels-union: mandatory.syncedLabels[tier] = %q, want global (L3 wins over L4)", got.Mandatory.SyncedLabels["tier"])
	}
}

// TestMergeWAFCascade_AllAbsent — when all sources are zero, effectiveConfig is
// all-zero (permissive; no governance enforced).
func TestMergeWAFCascade_AllAbsent(t *testing.T) {
	got := mergeWAFAll(
		zeroKropathWAF, zeroKropathWAF,
		zeroWAFCfg, zeroWAFCfg, zeroWAFCfg, zeroWAFCfg,
		zeroKropathWAF, zeroKropathWAF,
	)

	if got.Mandatory.Scope != "" {
		t.Errorf("all-absent: mandatory.scope = %q, want empty", got.Mandatory.Scope)
	}
	if got.Mandatory.DefaultAction != "" {
		t.Errorf("all-absent: mandatory.defaultAction = %q, want empty", got.Mandatory.DefaultAction)
	}
	if got.Mandatory.CloudWatchMetricsEnabled {
		t.Errorf("all-absent: mandatory.cloudWatchMetricsEnabled = true, want false")
	}
	if got.Mandatory.SampledRequestsEnabled {
		t.Errorf("all-absent: mandatory.sampledRequestsEnabled = true, want false")
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("all-absent: mandatory.namingTemplate = %q, want empty", got.Mandatory.NamingTemplate)
	}
	if len(got.Mandatory.Tags) != 0 {
		t.Errorf("all-absent: mandatory.tags = %v, want empty", got.Mandatory.Tags)
	}
	if got.Defaults.Scope != "" {
		t.Errorf("all-absent: defaults.scope = %q, want empty", got.Defaults.Scope)
	}
	if got.Defaults.CloudWatchMetricsEnabled {
		t.Errorf("all-absent: defaults.cloudWatchMetricsEnabled = true, want false")
	}
	if len(got.Defaults.Tags) != 0 {
		t.Errorf("all-absent: defaults.tags = %v, want empty", got.Defaults.Tags)
	}
}
