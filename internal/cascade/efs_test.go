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

// zeroKropathEFS is a zero-value EFSKropathSection (absent source, no org-wide tags).
var zeroKropathEFS = cascade.EFSKropathSection{}

// zeroEFSCfg is a zero-value EFSConfigSection (absent source).
var zeroEFSCfg = cascade.EFSConfigSection{}

// mergeEFSAll calls MergeEFSCascade with all eight inputs.
func mergeEFSAll(
	globalKropathMandatory,
	localKropathMandatory cascade.EFSKropathSection,
	globalEFSCfgMandatory,
	localEFSCfgMandatory,
	localEFSCfgDefaults,
	globalEFSCfgDefaults cascade.EFSConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.EFSKropathSection,
) cascade.EffectiveEFSConfig {
	return cascade.MergeEFSCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalEFSCfgMandatory,
		localEFSCfgMandatory,
		localEFSCfgDefaults,
		globalEFSCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeEFSCascade_AC_C1 — EFSConfig.mandatory.encrypted=true at level 3
// propagates to effectiveConfig.mandatory.encrypted=true.
func TestMergeEFSCascade_AC_C1(t *testing.T) {
	got := mergeEFSAll(
		zeroKropathEFS,
		zeroKropathEFS,
		cascade.EFSConfigSection{Encrypted: boolPtr(true)}, // level 3
		zeroEFSCfg,
		zeroEFSCfg,
		zeroEFSCfg,
		zeroKropathEFS,
		zeroKropathEFS,
	)
	if got.Mandatory.Encrypted == nil || !*got.Mandatory.Encrypted {
		t.Errorf("AC-C1: mandatory.encrypted = %v, want *true", got.Mandatory.Encrypted)
	}
}

// TestMergeEFSCascade_AC_C1_LocalOverridesGlobal — localEFSConfig.mandatory.encrypted at level 4
// wins over globalEFSConfig.mandatory.encrypted at level 3 when level 4 is set first.
// NOTE: level 3 (global) wins over level 4 (local) in mandatory priority chain.
func TestMergeEFSCascade_AC_C1_GlobalMandatoryWinsOverLocal(t *testing.T) {
	got := mergeEFSAll(
		zeroKropathEFS,
		zeroKropathEFS,
		cascade.EFSConfigSection{Encrypted: boolPtr(true)}, // level 3
		cascade.EFSConfigSection{Encrypted: boolPtr(false)}, // level 4 (lower priority)
		zeroEFSCfg,
		zeroEFSCfg,
		zeroKropathEFS,
		zeroKropathEFS,
	)
	if got.Mandatory.Encrypted == nil || !*got.Mandatory.Encrypted {
		t.Errorf("mandatory priority: encrypted = %v, want *true (level 3 must win over level 4)", got.Mandatory.Encrypted)
	}
}

// TestMergeEFSCascade_AC_C2 — EFSConfig.mandatory.kmsKeyId propagates.
func TestMergeEFSCascade_AC_C2(t *testing.T) {
	const keyArn = "arn:aws:kms:us-east-1:123456789012:key/my-key"
	got := mergeEFSAll(
		zeroKropathEFS,
		zeroKropathEFS,
		cascade.EFSConfigSection{KmsKeyId: keyArn}, // level 3
		zeroEFSCfg,
		zeroEFSCfg,
		zeroEFSCfg,
		zeroKropathEFS,
		zeroKropathEFS,
	)
	if got.Mandatory.KmsKeyId != keyArn {
		t.Errorf("AC-C2: mandatory.kmsKeyId = %q, want %q", got.Mandatory.KmsKeyId, keyArn)
	}
}

// TestMergeEFSCascade_AC_C3 — defaults.encrypted=false at level 6 is preserved;
// explicit false is NOT treated as nil/not-set.
func TestMergeEFSCascade_AC_C3(t *testing.T) {
	got := mergeEFSAll(
		zeroKropathEFS,
		zeroKropathEFS,
		zeroEFSCfg,
		zeroEFSCfg,
		cascade.EFSConfigSection{Encrypted: boolPtr(false)}, // level 6
		zeroEFSCfg,
		zeroKropathEFS,
		zeroKropathEFS,
	)
	if got.Defaults.Encrypted == nil || *got.Defaults.Encrypted {
		t.Errorf("AC-C3: defaults.encrypted = %v, want *false (explicit false must be preserved)", got.Defaults.Encrypted)
	}
	if got.Mandatory.Encrypted != nil {
		t.Errorf("AC-C3: mandatory.encrypted = %v, want nil", got.Mandatory.Encrypted)
	}
}

// TestMergeEFSCascade_AC_C4 — defaults.performanceMode propagates from level 6.
func TestMergeEFSCascade_AC_C4(t *testing.T) {
	got := mergeEFSAll(
		zeroKropathEFS,
		zeroKropathEFS,
		zeroEFSCfg,
		zeroEFSCfg,
		cascade.EFSConfigSection{PerformanceMode: "generalPurpose"}, // level 6
		zeroEFSCfg,
		zeroKropathEFS,
		zeroKropathEFS,
	)
	if got.Defaults.PerformanceMode != "generalPurpose" {
		t.Errorf("AC-C4: defaults.performanceMode = %q, want %q", got.Defaults.PerformanceMode, "generalPurpose")
	}
}

// TestMergeEFSCascade_AC_C5 — defaults.throughputMode propagates from level 6.
func TestMergeEFSCascade_AC_C5(t *testing.T) {
	got := mergeEFSAll(
		zeroKropathEFS,
		zeroKropathEFS,
		zeroEFSCfg,
		zeroEFSCfg,
		cascade.EFSConfigSection{ThroughputMode: "elastic"}, // level 6
		zeroEFSCfg,
		zeroKropathEFS,
		zeroKropathEFS,
	)
	if got.Defaults.ThroughputMode != "elastic" {
		t.Errorf("AC-C5: defaults.throughputMode = %q, want %q", got.Defaults.ThroughputMode, "elastic")
	}
}

// TestMergeEFSCascade_AC_C6 — mandatory.backupEnabled=true propagates from level 3.
func TestMergeEFSCascade_AC_C6(t *testing.T) {
	got := mergeEFSAll(
		zeroKropathEFS,
		zeroKropathEFS,
		cascade.EFSConfigSection{BackupEnabled: boolPtr(true)}, // level 3
		zeroEFSCfg,
		zeroEFSCfg,
		zeroEFSCfg,
		zeroKropathEFS,
		zeroKropathEFS,
	)
	if got.Mandatory.BackupEnabled == nil || !*got.Mandatory.BackupEnabled {
		t.Errorf("AC-C6: mandatory.backupEnabled = %v, want *true", got.Mandatory.BackupEnabled)
	}
}

// TestMergeEFSCascade_AC_C7 — defaults.replicationOverwriteProtection propagates from level 6.
func TestMergeEFSCascade_AC_C7(t *testing.T) {
	got := mergeEFSAll(
		zeroKropathEFS,
		zeroKropathEFS,
		zeroEFSCfg,
		zeroEFSCfg,
		cascade.EFSConfigSection{ReplicationOverwriteProtection: "ENABLED"}, // level 6
		zeroEFSCfg,
		zeroKropathEFS,
		zeroKropathEFS,
	)
	if got.Defaults.ReplicationOverwriteProtection != "ENABLED" {
		t.Errorf("AC-C7: defaults.replicationOverwriteProtection = %q, want %q", got.Defaults.ReplicationOverwriteProtection, "ENABLED")
	}
}

// TestMergeEFSCascade_AC_C8 — tags union merge: KropathConfig.mandatory.tags and
// EFSConfig.mandatory.tags both appear in effectiveConfig.mandatory.tags.
func TestMergeEFSCascade_AC_C8(t *testing.T) {
	got := mergeEFSAll(
		cascade.EFSKropathSection{Tags: map[string]string{"cost-centre": "platform"}}, // level 1
		zeroKropathEFS,
		cascade.EFSConfigSection{Tags: map[string]string{"resource-type": "storage"}}, // level 3
		zeroEFSCfg,
		zeroEFSCfg,
		zeroEFSCfg,
		zeroKropathEFS,
		zeroKropathEFS,
	)
	if got.Mandatory.Tags["cost-centre"] != "platform" {
		t.Errorf("AC-C8: mandatory.tags[cost-centre] = %q, want %q", got.Mandatory.Tags["cost-centre"], "platform")
	}
	if got.Mandatory.Tags["resource-type"] != "storage" {
		t.Errorf("AC-C8: mandatory.tags[resource-type] = %q, want %q", got.Mandatory.Tags["resource-type"], "storage")
	}
}

// TestMergeEFSCascade_AC_C9 — syncedLabels from EFSConfig.mandatory propagates.
func TestMergeEFSCascade_AC_C9(t *testing.T) {
	got := mergeEFSAll(
		zeroKropathEFS,
		zeroKropathEFS,
		cascade.EFSConfigSection{SyncedLabels: map[string]string{"data-class": "sensitive"}}, // level 3
		zeroEFSCfg,
		zeroEFSCfg,
		zeroEFSCfg,
		zeroKropathEFS,
		zeroKropathEFS,
	)
	if got.Mandatory.SyncedLabels["data-class"] != "sensitive" {
		t.Errorf("AC-C9: mandatory.syncedLabels[data-class] = %q, want %q", got.Mandatory.SyncedLabels["data-class"], "sensitive")
	}
}

// TestMergeEFSCascade_DefaultsTagsUnion — tags union in defaults tier covers all four levels.
func TestMergeEFSCascade_DefaultsTagsUnion(t *testing.T) {
	got := mergeEFSAll(
		zeroKropathEFS,
		zeroKropathEFS,
		zeroEFSCfg,
		zeroEFSCfg,
		cascade.EFSConfigSection{Tags: map[string]string{"tier": "gold"}},    // level 6
		cascade.EFSConfigSection{Tags: map[string]string{"tier": "silver"}},   // level 7 (lower priority)
		zeroKropathEFS,
		cascade.EFSKropathSection{Tags: map[string]string{"env": "prod"}}, // level 9
	)
	// level 6 wins over level 7 on key conflict
	if got.Defaults.Tags["tier"] != "gold" {
		t.Errorf("defaults tags priority: tier = %q, want %q (level 6 must win over level 7)", got.Defaults.Tags["tier"], "gold")
	}
	if got.Defaults.Tags["env"] != "prod" {
		t.Errorf("defaults tags union: env = %q, want %q", got.Defaults.Tags["env"], "prod")
	}
}

// TestMergeEFSCascade_MandatoryTagsConflict — KropathConfig.mandatory.tags wins over
// EFSConfig.mandatory.tags on key conflict (level 1 wins over level 3).
func TestMergeEFSCascade_MandatoryTagsConflict(t *testing.T) {
	got := mergeEFSAll(
		cascade.EFSKropathSection{Tags: map[string]string{"env": "prod"}},     // level 1
		zeroKropathEFS,
		cascade.EFSConfigSection{Tags: map[string]string{"env": "staging"}},    // level 3 (lower priority)
		zeroEFSCfg,
		zeroEFSCfg,
		zeroEFSCfg,
		zeroKropathEFS,
		zeroKropathEFS,
	)
	if got.Mandatory.Tags["env"] != "prod" {
		t.Errorf("tags conflict: mandatory.tags[env] = %q, want %q (level 1 must win over level 3)", got.Mandatory.Tags["env"], "prod")
	}
}

// TestMergeEFSCascade_BoolNilFallsThrough — nil falls through to next candidate for *bool fields.
func TestMergeEFSCascade_BoolNilFallsThrough(t *testing.T) {
	got := mergeEFSAll(
		zeroKropathEFS,
		zeroKropathEFS,
		cascade.EFSConfigSection{Encrypted: nil},           // level 3: nil
		cascade.EFSConfigSection{Encrypted: boolPtr(false)}, // level 4: explicit false
		zeroEFSCfg,
		zeroEFSCfg,
		zeroKropathEFS,
		zeroKropathEFS,
	)
	if got.Mandatory.Encrypted == nil || *got.Mandatory.Encrypted {
		t.Errorf("nil-falls-through: mandatory.encrypted = %v, want *false (level 4 explicit-false must win over nil level 3)", got.Mandatory.Encrypted)
	}
}

// TestMergeEFSCascade_AllZero — all inputs zero returns zero-value effective config.
func TestMergeEFSCascade_AllZero(t *testing.T) {
	got := mergeEFSAll(
		zeroKropathEFS, zeroKropathEFS,
		zeroEFSCfg, zeroEFSCfg, zeroEFSCfg, zeroEFSCfg,
		zeroKropathEFS, zeroKropathEFS,
	)
	if got.Mandatory.Encrypted != nil {
		t.Errorf("all-zero: mandatory.encrypted = %v, want nil", got.Mandatory.Encrypted)
	}
	if got.Mandatory.PerformanceMode != "" {
		t.Errorf("all-zero: mandatory.performanceMode = %q, want empty", got.Mandatory.PerformanceMode)
	}
	if got.Defaults.Encrypted != nil {
		t.Errorf("all-zero: defaults.encrypted = %v, want nil", got.Defaults.Encrypted)
	}
	if got.Defaults.ThroughputMode != "" {
		t.Errorf("all-zero: defaults.throughputMode = %q, want empty", got.Defaults.ThroughputMode)
	}
}

// TestMergeEFSCascade_TransitionFields — lifecycle policy string fields propagate.
func TestMergeEFSCascade_TransitionFields(t *testing.T) {
	got := mergeEFSAll(
		zeroKropathEFS,
		zeroKropathEFS,
		cascade.EFSConfigSection{TransitionToIA: "AFTER_30_DAYS"}, // level 3
		zeroEFSCfg,
		cascade.EFSConfigSection{TransitionToArchive: "AFTER_90_DAYS"},                 // level 6
		cascade.EFSConfigSection{TransitionToPrimaryStorage: "AFTER_1_ACCESS_IN_7_DAYS"}, // level 7
		zeroKropathEFS,
		zeroKropathEFS,
	)
	if got.Mandatory.TransitionToIA != "AFTER_30_DAYS" {
		t.Errorf("mandatory.transitionToIA = %q, want %q", got.Mandatory.TransitionToIA, "AFTER_30_DAYS")
	}
	if got.Defaults.TransitionToArchive != "AFTER_90_DAYS" {
		t.Errorf("defaults.transitionToArchive = %q, want %q", got.Defaults.TransitionToArchive, "AFTER_90_DAYS")
	}
	// level 6 wins over level 7 for defaults.transitionToPrimaryStorage (both empty in level 6, so level 7 applies)
	if got.Defaults.TransitionToPrimaryStorage != "AFTER_1_ACCESS_IN_7_DAYS" {
		t.Errorf("defaults.transitionToPrimaryStorage = %q, want %q", got.Defaults.TransitionToPrimaryStorage, "AFTER_1_ACCESS_IN_7_DAYS")
	}
}
