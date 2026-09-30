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

// zeroKropathCT is a zero-value CloudTrailKropathSection (absent source).
var zeroKropathCT = cascade.CloudTrailKropathSection{}

// zeroCTCfg is a zero-value CloudTrailConfigSection (absent source).
var zeroCTCfg = cascade.CloudTrailConfigSection{}

// mergeCTAll calls MergeCloudTrailCascade with all eight inputs.
func mergeCTAll(
	globalKropathMandatory,
	localKropathMandatory cascade.CloudTrailKropathSection,
	globalCTCfgMandatory,
	localCTCfgMandatory,
	localCTCfgDefaults,
	globalCTCfgDefaults cascade.CloudTrailConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.CloudTrailKropathSection,
) cascade.EffectiveCloudTrailConfig {
	return cascade.MergeCloudTrailCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalCTCfgMandatory,
		localCTCfgMandatory,
		localCTCfgDefaults,
		globalCTCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeCloudTrailCascade_IsMultiRegionTrail_Level1 — globalKropathConfig.mandatory.cloudtrail.
// isMultiRegionTrail=true at level 1 propagates to effCfg.mandatory.isMultiRegionTrail.
func TestMergeCloudTrailCascade_IsMultiRegionTrail_Level1(t *testing.T) {
	got := mergeCTAll(
		cascade.CloudTrailKropathSection{IsMultiRegionTrail: true}, // level 1
		zeroKropathCT,
		zeroCTCfg,
		zeroCTCfg,
		zeroCTCfg,
		zeroCTCfg,
		zeroKropathCT,
		zeroKropathCT,
	)

	if !got.Mandatory.IsMultiRegionTrail {
		t.Errorf("isMultiRegionTrail level1: mandatory = %v, want true (level 1 wins)", got.Mandatory.IsMultiRegionTrail)
	}
	if got.Defaults.IsMultiRegionTrail {
		t.Errorf("isMultiRegionTrail level1: defaults = %v, must not bleed from mandatory", got.Defaults.IsMultiRegionTrail)
	}
}

// TestMergeCloudTrailCascade_IsMultiRegionTrail_Level3 — globalCTConfig.mandatory.isMultiRegionTrail=true
// at level 3 wins when levels 1-2 are false.
func TestMergeCloudTrailCascade_IsMultiRegionTrail_Level3(t *testing.T) {
	got := mergeCTAll(
		zeroKropathCT,
		zeroKropathCT,
		cascade.CloudTrailConfigSection{IsMultiRegionTrail: true}, // level 3
		zeroCTCfg,
		zeroCTCfg,
		zeroCTCfg,
		zeroKropathCT,
		zeroKropathCT,
	)

	if !got.Mandatory.IsMultiRegionTrail {
		t.Errorf("isMultiRegionTrail level3: mandatory = %v, want true (level 3 wins when 1-2 false)", got.Mandatory.IsMultiRegionTrail)
	}
}

// TestMergeCloudTrailCascade_IsMultiRegionTrail_DefaultsLevel6 — localCTConfig.defaults.
// isMultiRegionTrail=true at level 6 propagates; mandatory stays false.
func TestMergeCloudTrailCascade_IsMultiRegionTrail_DefaultsLevel6(t *testing.T) {
	got := mergeCTAll(
		zeroKropathCT,
		zeroKropathCT,
		zeroCTCfg,
		zeroCTCfg,
		cascade.CloudTrailConfigSection{IsMultiRegionTrail: true}, // level 6
		zeroCTCfg,
		zeroKropathCT,
		zeroKropathCT,
	)

	if got.Mandatory.IsMultiRegionTrail {
		t.Errorf("isMultiRegionTrail defaults level6: mandatory = %v, want false", got.Mandatory.IsMultiRegionTrail)
	}
	if !got.Defaults.IsMultiRegionTrail {
		t.Errorf("isMultiRegionTrail defaults level6: defaults = %v, want true (level 6)", got.Defaults.IsMultiRegionTrail)
	}
}

// TestMergeCloudTrailCascade_EnableLogFileValidation_Level1 — globalKropathConfig.mandatory.cloudtrail.
// enableLogFileValidation=true at level 1 propagates.
func TestMergeCloudTrailCascade_EnableLogFileValidation_Level1(t *testing.T) {
	got := mergeCTAll(
		cascade.CloudTrailKropathSection{EnableLogFileValidation: true}, // level 1
		zeroKropathCT,
		zeroCTCfg,
		zeroCTCfg,
		zeroCTCfg,
		zeroCTCfg,
		zeroKropathCT,
		zeroKropathCT,
	)

	if !got.Mandatory.EnableLogFileValidation {
		t.Errorf("enableLogFileValidation level1: mandatory = %v, want true (level 1 wins)", got.Mandatory.EnableLogFileValidation)
	}
	if got.Defaults.EnableLogFileValidation {
		t.Errorf("enableLogFileValidation level1: defaults = %v, must not bleed from mandatory", got.Defaults.EnableLogFileValidation)
	}
}

// TestMergeCloudTrailCascade_EnableLogFileValidation_DefaultsFromCTConfig — globalCTConfig.defaults.
// enableLogFileValidation=true at level 7 propagates; mandatory stays false.
func TestMergeCloudTrailCascade_EnableLogFileValidation_DefaultsFromCTConfig(t *testing.T) {
	got := mergeCTAll(
		zeroKropathCT,
		zeroKropathCT,
		zeroCTCfg,
		zeroCTCfg,
		zeroCTCfg,
		cascade.CloudTrailConfigSection{EnableLogFileValidation: true}, // level 7
		zeroKropathCT,
		zeroKropathCT,
	)

	if got.Mandatory.EnableLogFileValidation {
		t.Errorf("enableLogFileValidation defaults level7: mandatory = %v, want false", got.Mandatory.EnableLogFileValidation)
	}
	if !got.Defaults.EnableLogFileValidation {
		t.Errorf("enableLogFileValidation defaults level7: defaults = %v, want true (level 7)", got.Defaults.EnableLogFileValidation)
	}
}

// TestMergeCloudTrailCascade_RetentionPeriod_Level1 — globalKropathConfig.mandatory.cloudtrail.
// retentionPeriod=2557 at level 1 propagates.
func TestMergeCloudTrailCascade_RetentionPeriod_Level1(t *testing.T) {
	got := mergeCTAll(
		cascade.CloudTrailKropathSection{RetentionPeriod: 2557}, // level 1
		zeroKropathCT,
		zeroCTCfg,
		zeroCTCfg,
		zeroCTCfg,
		zeroCTCfg,
		zeroKropathCT,
		zeroKropathCT,
	)

	if got.Mandatory.RetentionPeriod != 2557 {
		t.Errorf("retentionPeriod level1: mandatory = %d, want 2557 (level 1 wins)", got.Mandatory.RetentionPeriod)
	}
	if got.Defaults.RetentionPeriod != 0 {
		t.Errorf("retentionPeriod level1: defaults = %d, must not bleed from mandatory", got.Defaults.RetentionPeriod)
	}
}

// TestMergeCloudTrailCascade_RetentionPeriod_Level3WinsWhen1And2Zero — globalCTConfig.mandatory.
// retentionPeriod=365 at level 3 wins when KropathConfig.mandatory.cloudtrail.retentionPeriod=0 (levels 1-2).
func TestMergeCloudTrailCascade_RetentionPeriod_Level3WinsWhen1And2Zero(t *testing.T) {
	got := mergeCTAll(
		zeroKropathCT,
		zeroKropathCT,
		cascade.CloudTrailConfigSection{RetentionPeriod: 365}, // level 3
		zeroCTCfg,
		zeroCTCfg,
		zeroCTCfg,
		zeroKropathCT,
		zeroKropathCT,
	)

	if got.Mandatory.RetentionPeriod != 365 {
		t.Errorf("retentionPeriod level3: mandatory = %d, want 365 (level 3 wins when 1-2 are 0)", got.Mandatory.RetentionPeriod)
	}
}

// TestMergeCloudTrailCascade_RetentionPeriod_DefaultsLevel6 — localCTConfig.defaults.
// retentionPeriod=90 at level 6 propagates; mandatory stays 0.
func TestMergeCloudTrailCascade_RetentionPeriod_DefaultsLevel6(t *testing.T) {
	got := mergeCTAll(
		zeroKropathCT,
		zeroKropathCT,
		zeroCTCfg,
		zeroCTCfg,
		cascade.CloudTrailConfigSection{RetentionPeriod: 90}, // level 6
		zeroCTCfg,
		zeroKropathCT,
		zeroKropathCT,
	)

	if got.Mandatory.RetentionPeriod != 0 {
		t.Errorf("retentionPeriod defaults level6: mandatory = %d, want 0", got.Mandatory.RetentionPeriod)
	}
	if got.Defaults.RetentionPeriod != 90 {
		t.Errorf("retentionPeriod defaults level6: defaults = %d, want 90 (level 6)", got.Defaults.RetentionPeriod)
	}
}

// TestMergeCloudTrailCascade_RetentionPeriod_DefaultsPriorityOrder — defaults priority:
// level 6 > 7 > 8 > 9 for retentionPeriod.
func TestMergeCloudTrailCascade_RetentionPeriod_DefaultsPriorityOrder(t *testing.T) {
	cases := []struct {
		name                  string
		localCTCfgDefaults    cascade.CloudTrailConfigSection
		globalCTCfgDefaults   cascade.CloudTrailConfigSection
		localKropathDefaults  cascade.CloudTrailKropathSection
		globalKropathDefaults cascade.CloudTrailKropathSection
		wantRetentionPeriod   int64
	}{
		{
			name:                  "level6-wins",
			localCTCfgDefaults:    cascade.CloudTrailConfigSection{RetentionPeriod: 30},
			globalCTCfgDefaults:   cascade.CloudTrailConfigSection{RetentionPeriod: 90},
			localKropathDefaults:  cascade.CloudTrailKropathSection{RetentionPeriod: 180},
			globalKropathDefaults: cascade.CloudTrailKropathSection{RetentionPeriod: 365},
			wantRetentionPeriod:   30,
		},
		{
			name:                  "level7-wins-when-6-absent",
			localCTCfgDefaults:    zeroCTCfg,
			globalCTCfgDefaults:   cascade.CloudTrailConfigSection{RetentionPeriod: 90},
			localKropathDefaults:  cascade.CloudTrailKropathSection{RetentionPeriod: 180},
			globalKropathDefaults: cascade.CloudTrailKropathSection{RetentionPeriod: 365},
			wantRetentionPeriod:   90,
		},
		{
			name:                  "level8-wins-when-6-7-absent",
			localCTCfgDefaults:    zeroCTCfg,
			globalCTCfgDefaults:   zeroCTCfg,
			localKropathDefaults:  cascade.CloudTrailKropathSection{RetentionPeriod: 180},
			globalKropathDefaults: cascade.CloudTrailKropathSection{RetentionPeriod: 365},
			wantRetentionPeriod:   180,
		},
		{
			name:                  "level9-wins-when-6-7-8-absent",
			localCTCfgDefaults:    zeroCTCfg,
			globalCTCfgDefaults:   zeroCTCfg,
			localKropathDefaults:  zeroKropathCT,
			globalKropathDefaults: cascade.CloudTrailKropathSection{RetentionPeriod: 365},
			wantRetentionPeriod:   365,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeCTAll(
				zeroKropathCT,
				zeroKropathCT,
				zeroCTCfg,
				zeroCTCfg,
				tc.localCTCfgDefaults,
				tc.globalCTCfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
			)
			if got.Defaults.RetentionPeriod != tc.wantRetentionPeriod {
				t.Errorf("defaults.retentionPeriod = %d, want %d", got.Defaults.RetentionPeriod, tc.wantRetentionPeriod)
			}
		})
	}
}

// TestMergeCloudTrailCascade_ZeroSentinel_IsMultiRegionTrail — when all sources are false,
// effectiveConfig.mandatory.isMultiRegionTrail is false (not enforced).
func TestMergeCloudTrailCascade_ZeroSentinel_IsMultiRegionTrail(t *testing.T) {
	got := mergeCTAll(
		zeroKropathCT, zeroKropathCT,
		zeroCTCfg, zeroCTCfg, zeroCTCfg, zeroCTCfg,
		zeroKropathCT, zeroKropathCT,
	)

	if got.Mandatory.IsMultiRegionTrail {
		t.Errorf("zero-sentinel isMultiRegionTrail: mandatory = %v, want false (not enforced)", got.Mandatory.IsMultiRegionTrail)
	}
	if got.Defaults.IsMultiRegionTrail {
		t.Errorf("zero-sentinel isMultiRegionTrail: defaults = %v, want false (not enforced)", got.Defaults.IsMultiRegionTrail)
	}
}

// TestMergeCloudTrailCascade_ZeroSentinel_RetentionPeriod — when all sources are 0,
// effectiveConfig.mandatory.retentionPeriod is 0 (not enforced).
func TestMergeCloudTrailCascade_ZeroSentinel_RetentionPeriod(t *testing.T) {
	got := mergeCTAll(
		zeroKropathCT, zeroKropathCT,
		zeroCTCfg, zeroCTCfg, zeroCTCfg, zeroCTCfg,
		zeroKropathCT, zeroKropathCT,
	)

	if got.Mandatory.RetentionPeriod != 0 {
		t.Errorf("zero-sentinel retentionPeriod: mandatory = %d, want 0 (not enforced)", got.Mandatory.RetentionPeriod)
	}
	if got.Defaults.RetentionPeriod != 0 {
		t.Errorf("zero-sentinel retentionPeriod: defaults = %d, want 0 (not enforced)", got.Defaults.RetentionPeriod)
	}
}

// TestMergeCloudTrailCascade_AllAbsent — when all sources are zero, effectiveConfig is
// all-zero (permissive; no governance enforced).
func TestMergeCloudTrailCascade_AllAbsent(t *testing.T) {
	got := mergeCTAll(
		zeroKropathCT, zeroKropathCT,
		zeroCTCfg, zeroCTCfg, zeroCTCfg, zeroCTCfg,
		zeroKropathCT, zeroKropathCT,
	)

	if got.Mandatory.IsMultiRegionTrail {
		t.Errorf("all-absent: mandatory.isMultiRegionTrail = %v, want false", got.Mandatory.IsMultiRegionTrail)
	}
	if got.Mandatory.EnableLogFileValidation {
		t.Errorf("all-absent: mandatory.enableLogFileValidation = %v, want false", got.Mandatory.EnableLogFileValidation)
	}
	if got.Mandatory.RetentionPeriod != 0 {
		t.Errorf("all-absent: mandatory.retentionPeriod = %d, want 0", got.Mandatory.RetentionPeriod)
	}
	if got.Mandatory.S3BucketName != "" {
		t.Errorf("all-absent: mandatory.s3BucketName = %q, want empty", got.Mandatory.S3BucketName)
	}
	if got.Mandatory.KmsKeyID != "" {
		t.Errorf("all-absent: mandatory.kmsKeyID = %q, want empty", got.Mandatory.KmsKeyID)
	}
	if len(got.Mandatory.Tags) != 0 {
		t.Errorf("all-absent: mandatory.tags = %v, want empty", got.Mandatory.Tags)
	}
	if got.Defaults.IsMultiRegionTrail {
		t.Errorf("all-absent: defaults.isMultiRegionTrail = %v, want false", got.Defaults.IsMultiRegionTrail)
	}
	if got.Defaults.RetentionPeriod != 0 {
		t.Errorf("all-absent: defaults.retentionPeriod = %d, want 0", got.Defaults.RetentionPeriod)
	}
	if len(got.Defaults.Tags) != 0 {
		t.Errorf("all-absent: defaults.tags = %v, want empty", got.Defaults.Tags)
	}
}

// TestMergeCloudTrailCascade_NamingTemplate_Level7 — globalCTConfig.defaults.
// namingTemplate="{namespace}-{name}" at level 7 propagates.
// KropathConfig.cloudtrail has no namingTemplate field.
func TestMergeCloudTrailCascade_NamingTemplate_Level7(t *testing.T) {
	got := mergeCTAll(
		zeroKropathCT,
		zeroKropathCT,
		zeroCTCfg,
		zeroCTCfg,
		zeroCTCfg,
		cascade.CloudTrailConfigSection{NamingTemplate: "{namespace}-{name}"}, // level 7
		zeroKropathCT,
		zeroKropathCT,
	)

	if got.Defaults.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("namingTemplate level7: defaults = %q, want {namespace}-{name} (level 7)", got.Defaults.NamingTemplate)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("namingTemplate level7: mandatory = %q, must be empty", got.Mandatory.NamingTemplate)
	}
}

// TestMergeCloudTrailCascade_NamingTemplate_MandatoryLevel3 — globalCTConfig.mandatory.
// namingTemplate="corp-{namespace}-{name}" at level 3 propagates to effCfg.mandatory.namingTemplate.
func TestMergeCloudTrailCascade_NamingTemplate_MandatoryLevel3(t *testing.T) {
	got := mergeCTAll(
		zeroKropathCT,
		zeroKropathCT,
		cascade.CloudTrailConfigSection{NamingTemplate: "corp-{namespace}-{name}"}, // level 3
		zeroCTCfg,
		zeroCTCfg,
		zeroCTCfg,
		zeroKropathCT,
		zeroKropathCT,
	)

	if got.Mandatory.NamingTemplate != "corp-{namespace}-{name}" {
		t.Errorf("namingTemplate level3: mandatory = %q, want corp-{namespace}-{name} (level 3)", got.Mandatory.NamingTemplate)
	}
}

// TestMergeCloudTrailCascade_TagsUnionMerge — KropathConfig.mandatory.tags and
// CTConfig.mandatory.tags are union-merged into effCfg.mandatory.tags.
func TestMergeCloudTrailCascade_TagsUnionMerge(t *testing.T) {
	got := mergeCTAll(
		cascade.CloudTrailKropathSection{Tags: map[string]string{"cost-centre": "infra"}},    // level 1
		zeroKropathCT,
		zeroCTCfg,
		cascade.CloudTrailConfigSection{Tags: map[string]string{"audit-class": "pci-dss"}}, // level 4
		zeroCTCfg,
		zeroCTCfg,
		zeroKropathCT,
		zeroKropathCT,
	)

	if got.Mandatory.Tags["cost-centre"] != "infra" {
		t.Errorf("tags-union: mandatory.tags[cost-centre] = %q, want infra", got.Mandatory.Tags["cost-centre"])
	}
	if got.Mandatory.Tags["audit-class"] != "pci-dss" {
		t.Errorf("tags-union: mandatory.tags[audit-class] = %q, want pci-dss", got.Mandatory.Tags["audit-class"])
	}
}

// TestMergeCloudTrailCascade_TagsKeyConflict — on key conflict, lower level number wins.
// Level 1 (globalKropathMandatory) wins over level 4 (localCTCfgMandatory).
func TestMergeCloudTrailCascade_TagsKeyConflict(t *testing.T) {
	got := mergeCTAll(
		cascade.CloudTrailKropathSection{Tags: map[string]string{"env": "org-level"}},     // level 1
		zeroKropathCT,
		zeroCTCfg,
		cascade.CloudTrailConfigSection{Tags: map[string]string{"env": "config-level"}}, // level 4
		zeroCTCfg,
		zeroCTCfg,
		zeroKropathCT,
		zeroKropathCT,
	)

	if got.Mandatory.Tags["env"] != "org-level" {
		t.Errorf("tag-conflict: mandatory.tags[env] = %q, want org-level (level 1 wins over level 4)", got.Mandatory.Tags["env"])
	}
}

// TestMergeCloudTrailCascade_SyncedLabelsUnion — SyncedLabels from global (L3) and
// local (L4) CTConfig mandatory tiers are union-merged; L3 wins on key conflict.
func TestMergeCloudTrailCascade_SyncedLabelsUnion(t *testing.T) {
	got := mergeCTAll(
		zeroKropathCT,
		zeroKropathCT,
		cascade.CloudTrailConfigSection{SyncedLabels: map[string]string{"tier": "global", "audit-class": "pci"}}, // level 3
		cascade.CloudTrailConfigSection{SyncedLabels: map[string]string{"tier": "local", "env": "prod"}},         // level 4
		zeroCTCfg,
		zeroCTCfg,
		zeroKropathCT,
		zeroKropathCT,
	)

	if got.Mandatory.SyncedLabels["audit-class"] != "pci" {
		t.Errorf("synced-labels-union: mandatory.syncedLabels[audit-class] = %q, want pci", got.Mandatory.SyncedLabels["audit-class"])
	}
	if got.Mandatory.SyncedLabels["env"] != "prod" {
		t.Errorf("synced-labels-union: mandatory.syncedLabels[env] = %q, want prod", got.Mandatory.SyncedLabels["env"])
	}
	if got.Mandatory.SyncedLabels["tier"] != "global" {
		t.Errorf("synced-labels-union: mandatory.syncedLabels[tier] = %q, want global (L3 wins over L4)", got.Mandatory.SyncedLabels["tier"])
	}
}

// TestMergeCloudTrailCascade_MandatoryDoesNotBleedIntoDefaults — a value set only in
// mandatory must not appear in defaults.
func TestMergeCloudTrailCascade_MandatoryDoesNotBleedIntoDefaults(t *testing.T) {
	got := mergeCTAll(
		cascade.CloudTrailKropathSection{IsMultiRegionTrail: true, RetentionPeriod: 2557}, // level 1
		zeroKropathCT,
		cascade.CloudTrailConfigSection{S3BucketName: "audit-logs", KmsKeyID: "arn:aws:kms:us-east-1:123:key/abc"}, // level 3
		zeroCTCfg,
		zeroCTCfg,
		zeroCTCfg,
		zeroKropathCT,
		zeroKropathCT,
	)

	if got.Defaults.IsMultiRegionTrail {
		t.Errorf("mandatory-bleed: defaults.isMultiRegionTrail = %v, want false", got.Defaults.IsMultiRegionTrail)
	}
	if got.Defaults.RetentionPeriod != 0 {
		t.Errorf("mandatory-bleed: defaults.retentionPeriod = %d, want 0", got.Defaults.RetentionPeriod)
	}
	if got.Defaults.S3BucketName != "" {
		t.Errorf("mandatory-bleed: defaults.s3BucketName = %q, want empty", got.Defaults.S3BucketName)
	}
	if got.Defaults.KmsKeyID != "" {
		t.Errorf("mandatory-bleed: defaults.kmsKeyID = %q, want empty", got.Defaults.KmsKeyID)
	}
}

// TestMergeCloudTrailCascade_IsMultiRegionTrail_MandatoryPriorityOrder — verifies mandatory
// priority order: level 1 > 2 > 3 > 4 for isMultiRegionTrail.
func TestMergeCloudTrailCascade_IsMultiRegionTrail_MandatoryPriorityOrder(t *testing.T) {
	// Only level 1 set: expect true
	got := mergeCTAll(
		cascade.CloudTrailKropathSection{IsMultiRegionTrail: true}, // level 1 (only source)
		zeroKropathCT,
		zeroCTCfg,
		zeroCTCfg,
		zeroCTCfg,
		zeroCTCfg,
		zeroKropathCT,
		zeroKropathCT,
	)
	if !got.Mandatory.IsMultiRegionTrail {
		t.Errorf("mandatory-priority L1: isMultiRegionTrail = %v, want true", got.Mandatory.IsMultiRegionTrail)
	}

	// Level 1 absent, level 3 set: expect true
	got = mergeCTAll(
		zeroKropathCT,
		zeroKropathCT,
		cascade.CloudTrailConfigSection{IsMultiRegionTrail: true}, // level 3 (only source)
		zeroCTCfg,
		zeroCTCfg,
		zeroCTCfg,
		zeroKropathCT,
		zeroKropathCT,
	)
	if !got.Mandatory.IsMultiRegionTrail {
		t.Errorf("mandatory-priority L3: isMultiRegionTrail = %v, want true (level 3 wins when 1-2 false)", got.Mandatory.IsMultiRegionTrail)
	}
}

// TestMergeCloudTrailCascade_IncludeGlobalServiceEvents_CTConfigOnly — includeGlobalServiceEvents
// is CloudTrailConfig-only; KropathConfig has no such field.
func TestMergeCloudTrailCascade_IncludeGlobalServiceEvents_CTConfigOnly(t *testing.T) {
	got := mergeCTAll(
		zeroKropathCT,
		zeroKropathCT,
		cascade.CloudTrailConfigSection{IncludeGlobalServiceEvents: true}, // level 3
		zeroCTCfg,
		zeroCTCfg,
		zeroCTCfg,
		zeroKropathCT,
		zeroKropathCT,
	)

	if !got.Mandatory.IncludeGlobalServiceEvents {
		t.Errorf("includeGlobalServiceEvents level3: mandatory = %v, want true", got.Mandatory.IncludeGlobalServiceEvents)
	}
}

// TestMergeCloudTrailCascade_TerminationProtectionEnabled_DefaultsLevel7 — globalCTConfig.defaults.
// terminationProtectionEnabled=true at level 7 propagates; mandatory stays false.
func TestMergeCloudTrailCascade_TerminationProtectionEnabled_DefaultsLevel7(t *testing.T) {
	got := mergeCTAll(
		zeroKropathCT,
		zeroKropathCT,
		zeroCTCfg,
		zeroCTCfg,
		zeroCTCfg,
		cascade.CloudTrailConfigSection{TerminationProtectionEnabled: true}, // level 7
		zeroKropathCT,
		zeroKropathCT,
	)

	if got.Mandatory.TerminationProtectionEnabled {
		t.Errorf("terminationProtection defaults level7: mandatory = %v, want false", got.Mandatory.TerminationProtectionEnabled)
	}
	if !got.Defaults.TerminationProtectionEnabled {
		t.Errorf("terminationProtection defaults level7: defaults = %v, want true (level 7)", got.Defaults.TerminationProtectionEnabled)
	}
}
