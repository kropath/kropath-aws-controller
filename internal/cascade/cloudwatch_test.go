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

// zeroKropathCW is a zero-value CloudWatchKropathSection (absent source).
var zeroKropathCW = cascade.CloudWatchKropathSection{}

// zeroCWCfg is a zero-value CloudWatchConfigSection (absent source).
var zeroCWCfg = cascade.CloudWatchConfigSection{}

// mergeCWAll calls MergeCloudWatchCascade with all eight inputs.
func mergeCWAll(
	globalKropathMandatory,
	localKropathMandatory cascade.CloudWatchKropathSection,
	globalCWCfgMandatory,
	localCWCfgMandatory,
	localCWCfgDefaults,
	globalCWCfgDefaults cascade.CloudWatchConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.CloudWatchKropathSection,
) cascade.EffectiveCloudWatchConfig {
	return cascade.MergeCloudWatchCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalCWCfgMandatory,
		localCWCfgMandatory,
		localCWCfgDefaults,
		globalCWCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// boolPtrEq returns true if both pointers point to the same bool value or are both nil.
func boolPtrEq(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// TestMergeCloudWatchCascade_AC1 — globalKropathConfig.mandatory.cloudwatch.actionsEnabled=true
// at level 1 propagates to effCfg.mandatory.actionsEnabled.
func TestMergeCloudWatchCascade_AC1(t *testing.T) {
	got := mergeCWAll(
		cascade.CloudWatchKropathSection{ActionsEnabled: boolPtr(true)}, // level 1
		zeroKropathCW,
		zeroCWCfg,
		zeroCWCfg,
		zeroCWCfg,
		zeroCWCfg,
		zeroKropathCW,
		zeroKropathCW,
	)

	if !boolPtrEq(got.Mandatory.ActionsEnabled, boolPtr(true)) {
		t.Errorf("AC-1: mandatory.actionsEnabled = %v, want true (level 1 wins)", got.Mandatory.ActionsEnabled)
	}
	if got.Defaults.ActionsEnabled != nil {
		t.Errorf("AC-1: defaults.actionsEnabled = %v, must not bleed from mandatory", got.Defaults.ActionsEnabled)
	}
}

// TestMergeCloudWatchCascade_AC2 — globalCWConfig.mandatory.actionsEnabled=false at level 3 wins
// when levels 1-2 are nil (not set). Boolean false is a valid and distinct sentinel from nil.
func TestMergeCloudWatchCascade_AC2(t *testing.T) {
	got := mergeCWAll(
		zeroKropathCW,
		zeroKropathCW,
		cascade.CloudWatchConfigSection{ActionsEnabled: boolPtr(false)}, // level 3 — false is set, not nil
		zeroCWCfg,
		zeroCWCfg,
		zeroCWCfg,
		zeroKropathCW,
		zeroKropathCW,
	)

	if !boolPtrEq(got.Mandatory.ActionsEnabled, boolPtr(false)) {
		t.Errorf("AC-2: mandatory.actionsEnabled = %v, want false (level 3 wins, nil skipped)", got.Mandatory.ActionsEnabled)
	}
}

// TestMergeCloudWatchCascade_AC3 — globalKropathConfig.mandatory.cloudwatch.treatMissingData="breaching"
// at level 1 propagates to effCfg.mandatory.treatMissingData.
func TestMergeCloudWatchCascade_AC3(t *testing.T) {
	got := mergeCWAll(
		cascade.CloudWatchKropathSection{TreatMissingData: "breaching"}, // level 1
		zeroKropathCW,
		zeroCWCfg,
		zeroCWCfg,
		zeroCWCfg,
		zeroCWCfg,
		zeroKropathCW,
		zeroKropathCW,
	)

	if got.Mandatory.TreatMissingData != "breaching" {
		t.Errorf("AC-3: mandatory.treatMissingData = %q, want breaching (level 1 wins)", got.Mandatory.TreatMissingData)
	}
	if got.Defaults.TreatMissingData != "" {
		t.Errorf("AC-3: defaults.treatMissingData = %q, must not bleed from mandatory", got.Defaults.TreatMissingData)
	}
}

// TestMergeCloudWatchCascade_AC4 — localCWConfig.mandatory.treatMissingData="notBreaching" at
// level 4 wins when levels 1-3 are empty.
func TestMergeCloudWatchCascade_AC4(t *testing.T) {
	got := mergeCWAll(
		zeroKropathCW,
		zeroKropathCW,
		zeroCWCfg,
		cascade.CloudWatchConfigSection{TreatMissingData: "notBreaching"}, // level 4
		zeroCWCfg,
		zeroCWCfg,
		zeroKropathCW,
		zeroKropathCW,
	)

	if got.Mandatory.TreatMissingData != "notBreaching" {
		t.Errorf("AC-4: mandatory.treatMissingData = %q, want notBreaching (level 4 wins when 1-3 empty)", got.Mandatory.TreatMissingData)
	}
}

// TestMergeCloudWatchCascade_AC5 — globalKropathConfig.mandatory.cloudwatch.outputFormat="json"
// at level 1 propagates to effCfg.mandatory.outputFormat.
func TestMergeCloudWatchCascade_AC5(t *testing.T) {
	got := mergeCWAll(
		cascade.CloudWatchKropathSection{OutputFormat: "json"}, // level 1
		zeroKropathCW,
		zeroCWCfg,
		zeroCWCfg,
		zeroCWCfg,
		zeroCWCfg,
		zeroKropathCW,
		zeroKropathCW,
	)

	if got.Mandatory.OutputFormat != "json" {
		t.Errorf("AC-5: mandatory.outputFormat = %q, want json (level 1 wins)", got.Mandatory.OutputFormat)
	}
	if got.Defaults.OutputFormat != "" {
		t.Errorf("AC-5: defaults.outputFormat = %q, must not bleed from mandatory", got.Defaults.OutputFormat)
	}
}

// TestMergeCloudWatchCascade_AC6 — localCWConfig.mandatory.outputFormat="opentelemetry1.0" at
// level 4 wins when levels 1-3 are empty.
func TestMergeCloudWatchCascade_AC6(t *testing.T) {
	got := mergeCWAll(
		zeroKropathCW,
		zeroKropathCW,
		zeroCWCfg,
		cascade.CloudWatchConfigSection{OutputFormat: "opentelemetry1.0"}, // level 4
		zeroCWCfg,
		zeroCWCfg,
		zeroKropathCW,
		zeroKropathCW,
	)

	if got.Mandatory.OutputFormat != "opentelemetry1.0" {
		t.Errorf("AC-6: mandatory.outputFormat = %q, want opentelemetry1.0 (level 4 wins when 1-3 empty)", got.Mandatory.OutputFormat)
	}
}

// TestMergeCloudWatchCascade_AC7 — defaults priority: localCWCfgDefaults (L6) beats
// globalCWCfgDefaults (L7) beats localKropathDefaults (L8) beats globalKropathDefaults (L9).
func TestMergeCloudWatchCascade_AC7_DefaultsPriority(t *testing.T) {
	cases := []struct {
		name                  string
		localCWCfgDefaults    cascade.CloudWatchConfigSection
		globalCWCfgDefaults   cascade.CloudWatchConfigSection
		localKropathDefaults  cascade.CloudWatchKropathSection
		globalKropathDefaults cascade.CloudWatchKropathSection
		wantTreatMissingData  string
	}{
		{
			name:                  "level6-wins",
			localCWCfgDefaults:    cascade.CloudWatchConfigSection{TreatMissingData: "breaching"},
			globalCWCfgDefaults:   cascade.CloudWatchConfigSection{TreatMissingData: "ignore"},
			localKropathDefaults:  cascade.CloudWatchKropathSection{TreatMissingData: "notBreaching"},
			globalKropathDefaults: cascade.CloudWatchKropathSection{TreatMissingData: "missing"},
			wantTreatMissingData:  "breaching",
		},
		{
			name:                  "level7-wins-when-6-absent",
			localCWCfgDefaults:    zeroCWCfg,
			globalCWCfgDefaults:   cascade.CloudWatchConfigSection{TreatMissingData: "ignore"},
			localKropathDefaults:  cascade.CloudWatchKropathSection{TreatMissingData: "notBreaching"},
			globalKropathDefaults: cascade.CloudWatchKropathSection{TreatMissingData: "missing"},
			wantTreatMissingData:  "ignore",
		},
		{
			name:                  "level8-wins-when-6-7-absent",
			localCWCfgDefaults:    zeroCWCfg,
			globalCWCfgDefaults:   zeroCWCfg,
			localKropathDefaults:  cascade.CloudWatchKropathSection{TreatMissingData: "notBreaching"},
			globalKropathDefaults: cascade.CloudWatchKropathSection{TreatMissingData: "missing"},
			wantTreatMissingData:  "notBreaching",
		},
		{
			name:                  "level9-wins-when-6-7-8-absent",
			localCWCfgDefaults:    zeroCWCfg,
			globalCWCfgDefaults:   zeroCWCfg,
			localKropathDefaults:  zeroKropathCW,
			globalKropathDefaults: cascade.CloudWatchKropathSection{TreatMissingData: "missing"},
			wantTreatMissingData:  "missing",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeCWAll(
				zeroKropathCW,
				zeroKropathCW,
				zeroCWCfg,
				zeroCWCfg,
				tc.localCWCfgDefaults,
				tc.globalCWCfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
			)
			if got.Defaults.TreatMissingData != tc.wantTreatMissingData {
				t.Errorf("defaults.treatMissingData = %q, want %q", got.Defaults.TreatMissingData, tc.wantTreatMissingData)
			}
		})
	}
}

// TestMergeCloudWatchCascade_AC8 — globalCWConfig.mandatory.namingTemplate="corp-{namespace}-{name}"
// at level 3 propagates to effCfg.mandatory.namingTemplate.
// KropathConfig.cloudwatch has no namingTemplate field — only CWConfig levels (3, 4) participate.
func TestMergeCloudWatchCascade_AC8(t *testing.T) {
	got := mergeCWAll(
		zeroKropathCW,
		zeroKropathCW,
		cascade.CloudWatchConfigSection{NamingTemplate: "corp-{namespace}-{name}"}, // level 3
		zeroCWCfg,
		zeroCWCfg,
		zeroCWCfg,
		zeroKropathCW,
		zeroKropathCW,
	)

	if got.Mandatory.NamingTemplate != "corp-{namespace}-{name}" {
		t.Errorf("AC-8: mandatory.namingTemplate = %q, want corp-{namespace}-{name} (level 3)", got.Mandatory.NamingTemplate)
	}
}

// TestMergeCloudWatchCascade_AC9 — localCWConfig.defaults.namingTemplate="{namespace}-{name}"
// at level 6 propagates to effCfg.defaults.namingTemplate.
// KropathConfig.cloudwatch has no namingTemplate — defaults only from CWConfig levels (6, 7).
func TestMergeCloudWatchCascade_AC9(t *testing.T) {
	got := mergeCWAll(
		zeroKropathCW,
		zeroKropathCW,
		zeroCWCfg,
		zeroCWCfg,
		cascade.CloudWatchConfigSection{NamingTemplate: "{namespace}-{name}"}, // level 6
		zeroCWCfg,
		zeroKropathCW,
		zeroKropathCW,
	)

	if got.Defaults.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("AC-9: defaults.namingTemplate = %q, want {namespace}-{name} (level 6)", got.Defaults.NamingTemplate)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("AC-9: mandatory.namingTemplate = %q, must not bleed from defaults", got.Mandatory.NamingTemplate)
	}
}

// TestMergeCloudWatchCascade_AC10 — KropathConfig.mandatory.tags and
// CWConfig.mandatory.tags are union-merged into effCfg.mandatory.tags.
// Level 1 wins on key conflict.
func TestMergeCloudWatchCascade_AC10(t *testing.T) {
	got := mergeCWAll(
		cascade.CloudWatchKropathSection{Tags: map[string]string{"cost-centre": "platform", "env": "org-level"}}, // level 1
		zeroKropathCW,
		zeroCWCfg,
		cascade.CloudWatchConfigSection{Tags: map[string]string{"team": "ops", "env": "config-level"}}, // level 4
		zeroCWCfg,
		zeroCWCfg,
		zeroKropathCW,
		zeroKropathCW,
	)

	if got.Mandatory.Tags["cost-centre"] != "platform" {
		t.Errorf("AC-10: mandatory.tags[cost-centre] = %q, want platform (from level 1)", got.Mandatory.Tags["cost-centre"])
	}
	if got.Mandatory.Tags["team"] != "ops" {
		t.Errorf("AC-10: mandatory.tags[team] = %q, want ops (from level 4)", got.Mandatory.Tags["team"])
	}
	// Level 1 wins key conflict
	if got.Mandatory.Tags["env"] != "org-level" {
		t.Errorf("AC-10: mandatory.tags[env] = %q, want org-level (level 1 wins over level 4)", got.Mandatory.Tags["env"])
	}
}

// TestMergeCloudWatchCascade_AllAbsent — when all sources are zero, effectiveConfig is
// all-zero (permissive; no governance enforced).
func TestMergeCloudWatchCascade_AllAbsent(t *testing.T) {
	got := mergeCWAll(
		zeroKropathCW, zeroKropathCW,
		zeroCWCfg, zeroCWCfg, zeroCWCfg, zeroCWCfg,
		zeroKropathCW, zeroKropathCW,
	)

	if got.Mandatory.ActionsEnabled != nil {
		t.Errorf("all-absent: mandatory.actionsEnabled = %v, want nil", got.Mandatory.ActionsEnabled)
	}
	if got.Mandatory.TreatMissingData != "" {
		t.Errorf("all-absent: mandatory.treatMissingData = %q, want empty", got.Mandatory.TreatMissingData)
	}
	if got.Mandatory.OutputFormat != "" {
		t.Errorf("all-absent: mandatory.outputFormat = %q, want empty", got.Mandatory.OutputFormat)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("all-absent: mandatory.namingTemplate = %q, want empty", got.Mandatory.NamingTemplate)
	}
	if len(got.Mandatory.Tags) != 0 {
		t.Errorf("all-absent: mandatory.tags = %v, want empty", got.Mandatory.Tags)
	}
	if got.Defaults.ActionsEnabled != nil {
		t.Errorf("all-absent: defaults.actionsEnabled = %v, want nil", got.Defaults.ActionsEnabled)
	}
	if got.Defaults.TreatMissingData != "" {
		t.Errorf("all-absent: defaults.treatMissingData = %q, want empty", got.Defaults.TreatMissingData)
	}
	if len(got.Defaults.Tags) != 0 {
		t.Errorf("all-absent: defaults.tags = %v, want empty", got.Defaults.Tags)
	}
}

// TestMergeCloudWatchCascade_MandatoryPriorityOrder — verifies mandatory priority order
// for actionsEnabled: level 1 > 2 > 3 > 4. Boolean false is a valid set value (not a sentinel).
func TestMergeCloudWatchCascade_MandatoryPriorityOrder(t *testing.T) {
	cases := []struct {
		name                   string
		globalKropathMandatory cascade.CloudWatchKropathSection
		localKropathMandatory  cascade.CloudWatchKropathSection
		globalCWCfgMandatory   cascade.CloudWatchConfigSection
		localCWCfgMandatory    cascade.CloudWatchConfigSection
		wantActionsEnabled     *bool
	}{
		{
			name:                   "level1-wins",
			globalKropathMandatory: cascade.CloudWatchKropathSection{ActionsEnabled: boolPtr(true)},
			localKropathMandatory:  cascade.CloudWatchKropathSection{ActionsEnabled: boolPtr(false)},
			globalCWCfgMandatory:   cascade.CloudWatchConfigSection{ActionsEnabled: boolPtr(true)},
			localCWCfgMandatory:    cascade.CloudWatchConfigSection{ActionsEnabled: boolPtr(false)},
			wantActionsEnabled:     boolPtr(true),
		},
		{
			name:                   "level2-wins-when-1-nil",
			globalKropathMandatory: zeroKropathCW,
			localKropathMandatory:  cascade.CloudWatchKropathSection{ActionsEnabled: boolPtr(false)},
			globalCWCfgMandatory:   cascade.CloudWatchConfigSection{ActionsEnabled: boolPtr(true)},
			localCWCfgMandatory:    cascade.CloudWatchConfigSection{ActionsEnabled: boolPtr(true)},
			wantActionsEnabled:     boolPtr(false),
		},
		{
			name:                   "level3-wins-when-1-2-nil",
			globalKropathMandatory: zeroKropathCW,
			localKropathMandatory:  zeroKropathCW,
			globalCWCfgMandatory:   cascade.CloudWatchConfigSection{ActionsEnabled: boolPtr(true)},
			localCWCfgMandatory:    cascade.CloudWatchConfigSection{ActionsEnabled: boolPtr(false)},
			wantActionsEnabled:     boolPtr(true),
		},
		{
			name:                   "level4-wins-when-1-2-3-nil",
			globalKropathMandatory: zeroKropathCW,
			localKropathMandatory:  zeroKropathCW,
			globalCWCfgMandatory:   zeroCWCfg,
			localCWCfgMandatory:    cascade.CloudWatchConfigSection{ActionsEnabled: boolPtr(false)},
			wantActionsEnabled:     boolPtr(false),
		},
		{
			name:                   "all-nil-returns-nil",
			globalKropathMandatory: zeroKropathCW,
			localKropathMandatory:  zeroKropathCW,
			globalCWCfgMandatory:   zeroCWCfg,
			localCWCfgMandatory:    zeroCWCfg,
			wantActionsEnabled:     nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeCWAll(
				tc.globalKropathMandatory,
				tc.localKropathMandatory,
				tc.globalCWCfgMandatory,
				tc.localCWCfgMandatory,
				zeroCWCfg,
				zeroCWCfg,
				zeroKropathCW,
				zeroKropathCW,
			)
			if !boolPtrEq(got.Mandatory.ActionsEnabled, tc.wantActionsEnabled) {
				gotStr := "<nil>"
				wantStr := "<nil>"
				if got.Mandatory.ActionsEnabled != nil {
					gotStr = "false"
					if *got.Mandatory.ActionsEnabled {
						gotStr = "true"
					}
				}
				if tc.wantActionsEnabled != nil {
					wantStr = "false"
					if *tc.wantActionsEnabled {
						wantStr = "true"
					}
				}
				t.Errorf("mandatory.actionsEnabled = %s, want %s", gotStr, wantStr)
			}
		})
	}
}

// TestMergeCloudWatchCascade_TagsKeyConflict — on key conflict, lower level number wins.
// Level 1 (globalKropathMandatory) wins over level 4 (localCWCfgMandatory).
func TestMergeCloudWatchCascade_TagsKeyConflict(t *testing.T) {
	got := mergeCWAll(
		cascade.CloudWatchKropathSection{Tags: map[string]string{"env": "org-level"}},      // level 1
		zeroKropathCW,
		zeroCWCfg,
		cascade.CloudWatchConfigSection{Tags: map[string]string{"env": "config-level"}}, // level 4
		zeroCWCfg,
		zeroCWCfg,
		zeroKropathCW,
		zeroKropathCW,
	)

	if got.Mandatory.Tags["env"] != "org-level" {
		t.Errorf("tag-conflict: mandatory.tags[env] = %q, want org-level (level 1 wins over level 4)", got.Mandatory.Tags["env"])
	}
}

// TestMergeCloudWatchCascade_SyncedLabelsUnion — SyncedLabels from global (L3) and
// local (L4) CWConfig mandatory tiers are union-merged; L3 wins on key conflict.
// KropathConfig.cloudwatch does NOT carry syncedLabels.
func TestMergeCloudWatchCascade_SyncedLabelsUnion(t *testing.T) {
	got := mergeCWAll(
		zeroKropathCW,
		zeroKropathCW,
		cascade.CloudWatchConfigSection{SyncedLabels: map[string]string{"tier": "global", "data-class": "internal"}}, // level 3
		cascade.CloudWatchConfigSection{SyncedLabels: map[string]string{"tier": "local", "env": "prod"}},             // level 4
		zeroCWCfg,
		zeroCWCfg,
		zeroKropathCW,
		zeroKropathCW,
	)

	if got.Mandatory.SyncedLabels["data-class"] != "internal" {
		t.Errorf("synced-labels-union: mandatory.syncedLabels[data-class] = %q, want internal (from L3)", got.Mandatory.SyncedLabels["data-class"])
	}
	if got.Mandatory.SyncedLabels["env"] != "prod" {
		t.Errorf("synced-labels-union: mandatory.syncedLabels[env] = %q, want prod (from L4)", got.Mandatory.SyncedLabels["env"])
	}
	if got.Mandatory.SyncedLabels["tier"] != "global" {
		t.Errorf("synced-labels-union: mandatory.syncedLabels[tier] = %q, want global (L3 wins over L4)", got.Mandatory.SyncedLabels["tier"])
	}
}

// TestMergeCloudWatchCascade_NamingTemplateKropathNotParticipating — setting a KropathSection
// with all scalar fields set does NOT produce a namingTemplate since that field is absent from
// KropathConfig.cloudwatch entirely.
func TestMergeCloudWatchCascade_NamingTemplateKropathNotParticipating(t *testing.T) {
	got := mergeCWAll(
		cascade.CloudWatchKropathSection{TreatMissingData: "breaching", OutputFormat: "json"}, // level 1, no NamingTemplate
		zeroKropathCW,
		zeroCWCfg,
		zeroCWCfg,
		zeroCWCfg,
		zeroCWCfg,
		zeroKropathCW,
		zeroKropathCW,
	)

	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("naming-template-kc-not-participating: mandatory.namingTemplate = %q, must be empty (KC has no namingTemplate)", got.Mandatory.NamingTemplate)
	}
}
