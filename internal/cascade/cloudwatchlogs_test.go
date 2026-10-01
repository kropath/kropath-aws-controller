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

// zeroKropathCWL is a zero-value CloudWatchLogsKropathSection (absent source).
var zeroKropathCWL = cascade.CloudWatchLogsKropathSection{}

// zeroCWLCfg is a zero-value CloudWatchLogsConfigSection (absent source).
var zeroCWLCfg = cascade.CloudWatchLogsConfigSection{}

// mergeCWLAll calls MergeCloudWatchLogsCascade with all eight inputs.
func mergeCWLAll(
	globalKropathMandatory,
	localKropathMandatory cascade.CloudWatchLogsKropathSection,
	globalCWLCfgMandatory,
	localCWLCfgMandatory,
	localCWLCfgDefaults,
	globalCWLCfgDefaults cascade.CloudWatchLogsConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.CloudWatchLogsKropathSection,
) cascade.EffectiveCloudWatchLogsConfig {
	return cascade.MergeCloudWatchLogsCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalCWLCfgMandatory,
		localCWLCfgMandatory,
		localCWLCfgDefaults,
		globalCWLCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeCloudWatchLogsCascade_AC1 — globalKropathConfig.mandatory.cloudwatchlogs.kmsKeyId
// at level 1 propagates to effCfg.mandatory.kmsKeyId (level 1 wins).
func TestMergeCloudWatchLogsCascade_AC1(t *testing.T) {
	keyID := "arn:aws:kms:ap-southeast-2:123:key/org-key"
	got := mergeCWLAll(
		cascade.CloudWatchLogsKropathSection{KmsKeyId: keyID}, // level 1
		zeroKropathCWL,
		zeroCWLCfg,
		zeroCWLCfg,
		zeroCWLCfg,
		zeroCWLCfg,
		zeroKropathCWL,
		zeroKropathCWL,
	)

	if got.Mandatory.KmsKeyId != keyID {
		t.Errorf("AC-1: mandatory.kmsKeyId = %q, want %q (level 1 wins)", got.Mandatory.KmsKeyId, keyID)
	}
	if got.Defaults.KmsKeyId != "" {
		t.Errorf("AC-1: defaults.kmsKeyId = %q, must not bleed from mandatory", got.Defaults.KmsKeyId)
	}
}

// TestMergeCloudWatchLogsCascade_AC2 — globalCWLConfig.mandatory.kmsKeyId at level 3 wins
// when levels 1-2 are empty.
func TestMergeCloudWatchLogsCascade_AC2(t *testing.T) {
	keyID := "arn:aws:kms:ap-southeast-2:123:key/rc-key"
	got := mergeCWLAll(
		zeroKropathCWL,
		zeroKropathCWL,
		cascade.CloudWatchLogsConfigSection{KmsKeyId: keyID}, // level 3
		zeroCWLCfg,
		zeroCWLCfg,
		zeroCWLCfg,
		zeroKropathCWL,
		zeroKropathCWL,
	)

	if got.Mandatory.KmsKeyId != keyID {
		t.Errorf("AC-2: mandatory.kmsKeyId = %q, want %q (level 3 wins when 1-2 empty)", got.Mandatory.KmsKeyId, keyID)
	}
}

// TestMergeCloudWatchLogsCascade_AC3 — localCWLConfig.defaults.kmsKeyId at level 6 propagates;
// mandatory stays empty.
func TestMergeCloudWatchLogsCascade_AC3(t *testing.T) {
	keyID := "arn:aws:kms:ap-southeast-2:123:key/default-key"
	got := mergeCWLAll(
		zeroKropathCWL,
		zeroKropathCWL,
		zeroCWLCfg,
		zeroCWLCfg,
		cascade.CloudWatchLogsConfigSection{KmsKeyId: keyID}, // level 6
		zeroCWLCfg,
		zeroKropathCWL,
		zeroKropathCWL,
	)

	if got.Mandatory.KmsKeyId != "" {
		t.Errorf("AC-3: mandatory.kmsKeyId = %q, want empty", got.Mandatory.KmsKeyId)
	}
	if got.Defaults.KmsKeyId != keyID {
		t.Errorf("AC-3: defaults.kmsKeyId = %q, want %q (level 6)", got.Defaults.KmsKeyId, keyID)
	}
}

// TestMergeCloudWatchLogsCascade_AC4 — globalKropathConfig.mandatory.cloudwatchlogs.retentionDays=365
// at level 1 propagates to effCfg.mandatory.retentionDays.
func TestMergeCloudWatchLogsCascade_AC4(t *testing.T) {
	got := mergeCWLAll(
		cascade.CloudWatchLogsKropathSection{RetentionDays: 365}, // level 1
		zeroKropathCWL,
		zeroCWLCfg,
		zeroCWLCfg,
		zeroCWLCfg,
		zeroCWLCfg,
		zeroKropathCWL,
		zeroKropathCWL,
	)

	if got.Mandatory.RetentionDays != 365 {
		t.Errorf("AC-4: mandatory.retentionDays = %d, want 365 (level 1 wins)", got.Mandatory.RetentionDays)
	}
	if got.Defaults.RetentionDays != 0 {
		t.Errorf("AC-4: defaults.retentionDays = %d, must not bleed from mandatory", got.Defaults.RetentionDays)
	}
}

// TestMergeCloudWatchLogsCascade_AC5 — globalCWLConfig.mandatory.retentionDays=2557 (level 3)
// wins when KropathConfig.mandatory.cloudwatchlogs.retentionDays=0 (levels 1-2).
func TestMergeCloudWatchLogsCascade_AC5(t *testing.T) {
	got := mergeCWLAll(
		zeroKropathCWL,
		zeroKropathCWL,
		cascade.CloudWatchLogsConfigSection{RetentionDays: 2557}, // level 3
		zeroCWLCfg,
		zeroCWLCfg,
		zeroCWLCfg,
		zeroKropathCWL,
		zeroKropathCWL,
	)

	if got.Mandatory.RetentionDays != 2557 {
		t.Errorf("AC-5: mandatory.retentionDays = %d, want 2557 (level 3 wins when 1-2 are 0)", got.Mandatory.RetentionDays)
	}
}

// TestMergeCloudWatchLogsCascade_AC6 — localCWLConfig.defaults.retentionDays=90 at level 6
// propagates; mandatory stays 0.
func TestMergeCloudWatchLogsCascade_AC6(t *testing.T) {
	got := mergeCWLAll(
		zeroKropathCWL,
		zeroKropathCWL,
		zeroCWLCfg,
		zeroCWLCfg,
		cascade.CloudWatchLogsConfigSection{RetentionDays: 90}, // level 6
		zeroCWLCfg,
		zeroKropathCWL,
		zeroKropathCWL,
	)

	if got.Mandatory.RetentionDays != 0 {
		t.Errorf("AC-6: mandatory.retentionDays = %d, want 0", got.Mandatory.RetentionDays)
	}
	if got.Defaults.RetentionDays != 90 {
		t.Errorf("AC-6: defaults.retentionDays = %d, want 90 (level 6)", got.Defaults.RetentionDays)
	}
}

// TestMergeCloudWatchLogsCascade_AC7 — globalCWLConfig.defaults.namingTemplate="{namespace}-{name}"
// at level 7 propagates. KropathConfig.cloudwatchlogs has no namingTemplate field.
func TestMergeCloudWatchLogsCascade_AC7(t *testing.T) {
	got := mergeCWLAll(
		zeroKropathCWL,
		zeroKropathCWL,
		zeroCWLCfg,
		zeroCWLCfg,
		zeroCWLCfg,
		cascade.CloudWatchLogsConfigSection{NamingTemplate: "{namespace}-{name}"}, // level 7
		zeroKropathCWL,
		zeroKropathCWL,
	)

	if got.Defaults.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("AC-7: defaults.namingTemplate = %q, want {namespace}-{name} (level 7)", got.Defaults.NamingTemplate)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("AC-7: mandatory.namingTemplate = %q, must be empty", got.Mandatory.NamingTemplate)
	}
}

// TestMergeCloudWatchLogsCascade_AC8 — globalCWLConfig.mandatory.namingTemplate="corp-{namespace}-{name}"
// at level 3 propagates to effCfg.mandatory.namingTemplate.
func TestMergeCloudWatchLogsCascade_AC8(t *testing.T) {
	got := mergeCWLAll(
		zeroKropathCWL,
		zeroKropathCWL,
		cascade.CloudWatchLogsConfigSection{NamingTemplate: "corp-{namespace}-{name}"}, // level 3
		zeroCWLCfg,
		zeroCWLCfg,
		zeroCWLCfg,
		zeroKropathCWL,
		zeroKropathCWL,
	)

	if got.Mandatory.NamingTemplate != "corp-{namespace}-{name}" {
		t.Errorf("AC-8: mandatory.namingTemplate = %q, want corp-{namespace}-{name} (level 3)", got.Mandatory.NamingTemplate)
	}
}

// TestMergeCloudWatchLogsCascade_AC9 — KropathConfig.mandatory.tags and
// CWLConfig.mandatory.tags are union-merged into effCfg.mandatory.tags.
func TestMergeCloudWatchLogsCascade_AC9(t *testing.T) {
	got := mergeCWLAll(
		cascade.CloudWatchLogsKropathSection{Tags: map[string]string{"cost-centre": "infra"}},     // level 1
		zeroKropathCWL,
		zeroCWLCfg,
		cascade.CloudWatchLogsConfigSection{Tags: map[string]string{"log-class": "production"}}, // level 4
		zeroCWLCfg,
		zeroCWLCfg,
		zeroKropathCWL,
		zeroKropathCWL,
	)

	if got.Mandatory.Tags["cost-centre"] != "infra" {
		t.Errorf("AC-9: mandatory.tags[cost-centre] = %q, want infra", got.Mandatory.Tags["cost-centre"])
	}
	if got.Mandatory.Tags["log-class"] != "production" {
		t.Errorf("AC-9: mandatory.tags[log-class] = %q, want production", got.Mandatory.Tags["log-class"])
	}
}

// TestMergeCloudWatchLogsCascade_AC10 — CWLConfig.mandatory.syncedLabels={data-class: internal}
// at level 3 propagates to effCfg.mandatory.syncedLabels.
func TestMergeCloudWatchLogsCascade_AC10(t *testing.T) {
	got := mergeCWLAll(
		zeroKropathCWL,
		zeroKropathCWL,
		cascade.CloudWatchLogsConfigSection{SyncedLabels: map[string]string{"data-class": "internal"}}, // level 3
		zeroCWLCfg,
		zeroCWLCfg,
		zeroCWLCfg,
		zeroKropathCWL,
		zeroKropathCWL,
	)

	if got.Mandatory.SyncedLabels["data-class"] != "internal" {
		t.Errorf("AC-10: mandatory.syncedLabels[data-class] = %q, want internal", got.Mandatory.SyncedLabels["data-class"])
	}
}

// TestMergeCloudWatchLogsCascade_AllAbsent — when all sources are zero, effectiveConfig is
// all-zero (permissive; no governance enforced).
func TestMergeCloudWatchLogsCascade_AllAbsent(t *testing.T) {
	got := mergeCWLAll(
		zeroKropathCWL, zeroKropathCWL,
		zeroCWLCfg, zeroCWLCfg, zeroCWLCfg, zeroCWLCfg,
		zeroKropathCWL, zeroKropathCWL,
	)

	if got.Mandatory.KmsKeyId != "" {
		t.Errorf("all-absent: mandatory.kmsKeyId = %q, want empty", got.Mandatory.KmsKeyId)
	}
	if got.Mandatory.RetentionDays != 0 {
		t.Errorf("all-absent: mandatory.retentionDays = %d, want 0", got.Mandatory.RetentionDays)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("all-absent: mandatory.namingTemplate = %q, want empty", got.Mandatory.NamingTemplate)
	}
	if len(got.Mandatory.Tags) != 0 {
		t.Errorf("all-absent: mandatory.tags = %v, want empty", got.Mandatory.Tags)
	}
	if got.Defaults.KmsKeyId != "" {
		t.Errorf("all-absent: defaults.kmsKeyId = %q, want empty", got.Defaults.KmsKeyId)
	}
	if got.Defaults.RetentionDays != 0 {
		t.Errorf("all-absent: defaults.retentionDays = %d, want 0", got.Defaults.RetentionDays)
	}
	if len(got.Defaults.Tags) != 0 {
		t.Errorf("all-absent: defaults.tags = %v, want empty", got.Defaults.Tags)
	}
}

// TestMergeCloudWatchLogsCascade_MandatoryPriorityOrder — verifies mandatory priority order
// for kmsKeyId: level 1 > 2 > 3 > 4.
func TestMergeCloudWatchLogsCascade_MandatoryPriorityOrder(t *testing.T) {
	cases := []struct {
		name                   string
		globalKropathMandatory cascade.CloudWatchLogsKropathSection
		localKropathMandatory  cascade.CloudWatchLogsKropathSection
		globalCWLCfgMandatory  cascade.CloudWatchLogsConfigSection
		localCWLCfgMandatory   cascade.CloudWatchLogsConfigSection
		wantKmsKeyId           string
	}{
		{
			name:                   "level1-wins",
			globalKropathMandatory: cascade.CloudWatchLogsKropathSection{KmsKeyId: "level1"},
			localKropathMandatory:  cascade.CloudWatchLogsKropathSection{KmsKeyId: "level2"},
			globalCWLCfgMandatory:  cascade.CloudWatchLogsConfigSection{KmsKeyId: "level3"},
			localCWLCfgMandatory:   cascade.CloudWatchLogsConfigSection{KmsKeyId: "level4"},
			wantKmsKeyId:           "level1",
		},
		{
			name:                   "level2-wins-when-1-absent",
			globalKropathMandatory: zeroKropathCWL,
			localKropathMandatory:  cascade.CloudWatchLogsKropathSection{KmsKeyId: "level2"},
			globalCWLCfgMandatory:  cascade.CloudWatchLogsConfigSection{KmsKeyId: "level3"},
			localCWLCfgMandatory:   cascade.CloudWatchLogsConfigSection{KmsKeyId: "level4"},
			wantKmsKeyId:           "level2",
		},
		{
			name:                   "level3-wins-when-1-2-absent",
			globalKropathMandatory: zeroKropathCWL,
			localKropathMandatory:  zeroKropathCWL,
			globalCWLCfgMandatory:  cascade.CloudWatchLogsConfigSection{KmsKeyId: "level3"},
			localCWLCfgMandatory:   cascade.CloudWatchLogsConfigSection{KmsKeyId: "level4"},
			wantKmsKeyId:           "level3",
		},
		{
			name:                   "level4-wins-when-1-2-3-absent",
			globalKropathMandatory: zeroKropathCWL,
			localKropathMandatory:  zeroKropathCWL,
			globalCWLCfgMandatory:  zeroCWLCfg,
			localCWLCfgMandatory:   cascade.CloudWatchLogsConfigSection{KmsKeyId: "level4"},
			wantKmsKeyId:           "level4",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeCWLAll(
				tc.globalKropathMandatory,
				tc.localKropathMandatory,
				tc.globalCWLCfgMandatory,
				tc.localCWLCfgMandatory,
				zeroCWLCfg,
				zeroCWLCfg,
				zeroKropathCWL,
				zeroKropathCWL,
			)
			if got.Mandatory.KmsKeyId != tc.wantKmsKeyId {
				t.Errorf("mandatory.kmsKeyId = %q, want %q", got.Mandatory.KmsKeyId, tc.wantKmsKeyId)
			}
		})
	}
}

// TestMergeCloudWatchLogsCascade_DefaultsPriorityOrder — verifies defaults priority order
// for retentionDays: level 6 > 7 > 8 > 9.
func TestMergeCloudWatchLogsCascade_DefaultsPriorityOrder(t *testing.T) {
	cases := []struct {
		name                  string
		localCWLCfgDefaults   cascade.CloudWatchLogsConfigSection
		globalCWLCfgDefaults  cascade.CloudWatchLogsConfigSection
		localKropathDefaults  cascade.CloudWatchLogsKropathSection
		globalKropathDefaults cascade.CloudWatchLogsKropathSection
		wantRetentionDays     int64
	}{
		{
			name:                  "level6-wins",
			localCWLCfgDefaults:   cascade.CloudWatchLogsConfigSection{RetentionDays: 30},
			globalCWLCfgDefaults:  cascade.CloudWatchLogsConfigSection{RetentionDays: 90},
			localKropathDefaults:  cascade.CloudWatchLogsKropathSection{RetentionDays: 180},
			globalKropathDefaults: cascade.CloudWatchLogsKropathSection{RetentionDays: 365},
			wantRetentionDays:     30,
		},
		{
			name:                  "level7-wins-when-6-absent",
			localCWLCfgDefaults:   zeroCWLCfg,
			globalCWLCfgDefaults:  cascade.CloudWatchLogsConfigSection{RetentionDays: 90},
			localKropathDefaults:  cascade.CloudWatchLogsKropathSection{RetentionDays: 180},
			globalKropathDefaults: cascade.CloudWatchLogsKropathSection{RetentionDays: 365},
			wantRetentionDays:     90,
		},
		{
			name:                  "level8-wins-when-6-7-absent",
			localCWLCfgDefaults:   zeroCWLCfg,
			globalCWLCfgDefaults:  zeroCWLCfg,
			localKropathDefaults:  cascade.CloudWatchLogsKropathSection{RetentionDays: 180},
			globalKropathDefaults: cascade.CloudWatchLogsKropathSection{RetentionDays: 365},
			wantRetentionDays:     180,
		},
		{
			name:                  "level9-wins-when-6-7-8-absent",
			localCWLCfgDefaults:   zeroCWLCfg,
			globalCWLCfgDefaults:  zeroCWLCfg,
			localKropathDefaults:  zeroKropathCWL,
			globalKropathDefaults: cascade.CloudWatchLogsKropathSection{RetentionDays: 365},
			wantRetentionDays:     365,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeCWLAll(
				zeroKropathCWL,
				zeroKropathCWL,
				zeroCWLCfg,
				zeroCWLCfg,
				tc.localCWLCfgDefaults,
				tc.globalCWLCfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
			)
			if got.Defaults.RetentionDays != tc.wantRetentionDays {
				t.Errorf("defaults.retentionDays = %d, want %d", got.Defaults.RetentionDays, tc.wantRetentionDays)
			}
		})
	}
}

// TestMergeCloudWatchLogsCascade_TagsKeyConflict — on key conflict, lower level number wins.
// Level 1 (globalKropathMandatory) wins over level 4 (localCWLCfgMandatory).
func TestMergeCloudWatchLogsCascade_TagsKeyConflict(t *testing.T) {
	got := mergeCWLAll(
		cascade.CloudWatchLogsKropathSection{Tags: map[string]string{"env": "org-level"}},      // level 1
		zeroKropathCWL,
		zeroCWLCfg,
		cascade.CloudWatchLogsConfigSection{Tags: map[string]string{"env": "config-level"}}, // level 4
		zeroCWLCfg,
		zeroCWLCfg,
		zeroKropathCWL,
		zeroKropathCWL,
	)

	if got.Mandatory.Tags["env"] != "org-level" {
		t.Errorf("tag-conflict: mandatory.tags[env] = %q, want org-level (level 1 wins over level 4)", got.Mandatory.Tags["env"])
	}
}

// TestMergeCloudWatchLogsCascade_SyncedLabelsUnion — SyncedLabels from global (L3) and
// local (L4) CWLConfig mandatory tiers are union-merged; L3 wins on key conflict.
func TestMergeCloudWatchLogsCascade_SyncedLabelsUnion(t *testing.T) {
	got := mergeCWLAll(
		zeroKropathCWL,
		zeroKropathCWL,
		cascade.CloudWatchLogsConfigSection{SyncedLabels: map[string]string{"tier": "global", "data-class": "internal"}}, // level 3
		cascade.CloudWatchLogsConfigSection{SyncedLabels: map[string]string{"tier": "local", "env": "prod"}},             // level 4
		zeroCWLCfg,
		zeroCWLCfg,
		zeroKropathCWL,
		zeroKropathCWL,
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
