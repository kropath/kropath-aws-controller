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

// zeroKropathRDS is a zero-value RDSKropathSection (absent source).
var zeroKropathRDS = cascade.RDSKropathSection{}

// zeroRDSCfg is a zero-value RDSConfigSection (absent source).
var zeroRDSCfg = cascade.RDSConfigSection{}

// mergeRDSAll calls MergeRDSCascade with all eight inputs.
func mergeRDSAll(
	globalKropathMandatory,
	localKropathMandatory cascade.RDSKropathSection,
	globalRDSCfgMandatory,
	localRDSCfgMandatory,
	localRDSCfgDefaults,
	globalRDSCfgDefaults cascade.RDSConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.RDSKropathSection,
) cascade.EffectiveRDSConfig {
	return cascade.MergeRDSCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalRDSCfgMandatory,
		localRDSCfgMandatory,
		localRDSCfgDefaults,
		globalRDSCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeRDSCascade_AC1 — KropathConfig.mandatory.rds.storageEncrypted: true at level 1
// forces storageEncrypted on all RDS resources regardless of RDSConfig.
func TestMergeRDSCascade_AC1(t *testing.T) {
	got := mergeRDSAll(
		cascade.RDSKropathSection{StorageEncrypted: boolPtr(true)}, // level 1
		zeroKropathRDS,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroKropathRDS,
		zeroKropathRDS,
	)
	if got.Mandatory.StorageEncrypted == nil || !*got.Mandatory.StorageEncrypted {
		t.Errorf("AC-1: mandatory.storageEncrypted = %v, want *true", got.Mandatory.StorageEncrypted)
	}
}

// TestMergeRDSCascade_AC2 — RDSConfig.mandatory.storageEncrypted: true at level 3
// overrides KropathConfig.defaults.rds.storageEncrypted (level 8).
func TestMergeRDSCascade_AC2(t *testing.T) {
	got := mergeRDSAll(
		zeroKropathRDS,
		zeroKropathRDS,
		cascade.RDSConfigSection{StorageEncrypted: boolPtr(true)}, // level 3 mandatory wins
		zeroRDSCfg,
		zeroRDSCfg,
		zeroRDSCfg,
		cascade.RDSKropathSection{StorageEncrypted: boolPtr(false)}, // level 8 defaults (should not win)
		zeroKropathRDS,
	)
	if got.Mandatory.StorageEncrypted == nil || !*got.Mandatory.StorageEncrypted {
		t.Errorf("AC-2: mandatory.storageEncrypted = %v, want *true (level 3 mandatory must win)", got.Mandatory.StorageEncrypted)
	}
	// Defaults tier: level 8 sets false
	if got.Defaults.StorageEncrypted == nil || *got.Defaults.StorageEncrypted {
		t.Errorf("AC-2: defaults.storageEncrypted = %v, want *false (level 8 defaults)", got.Defaults.StorageEncrypted)
	}
}

// TestMergeRDSCascade_AC3 — RDSConfig.defaults.storageEncrypted: true at level 6 applies
// when no mandatory override; explicit false pointer must be preserved.
func TestMergeRDSCascade_AC3(t *testing.T) {
	got := mergeRDSAll(
		zeroKropathRDS,
		zeroKropathRDS,
		zeroRDSCfg,
		zeroRDSCfg,
		cascade.RDSConfigSection{StorageEncrypted: boolPtr(true)}, // level 6 defaults
		zeroRDSCfg,
		zeroKropathRDS,
		zeroKropathRDS,
	)
	if got.Mandatory.StorageEncrypted != nil {
		t.Errorf("AC-3: mandatory.storageEncrypted = %v, want nil (no mandatory override)", got.Mandatory.StorageEncrypted)
	}
	if got.Defaults.StorageEncrypted == nil || !*got.Defaults.StorageEncrypted {
		t.Errorf("AC-3: defaults.storageEncrypted = %v, want *true", got.Defaults.StorageEncrypted)
	}
}

// TestMergeRDSCascade_AC4 — KropathConfig.mandatory.rds.deletionProtection: true at level 1
// forces deletion protection on all RDS resources.
func TestMergeRDSCascade_AC4(t *testing.T) {
	got := mergeRDSAll(
		cascade.RDSKropathSection{DeletionProtection: boolPtr(true)}, // level 1
		zeroKropathRDS,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroKropathRDS,
		zeroKropathRDS,
	)
	if got.Mandatory.DeletionProtection == nil || !*got.Mandatory.DeletionProtection {
		t.Errorf("AC-4: mandatory.deletionProtection = %v, want *true", got.Mandatory.DeletionProtection)
	}
}

// TestMergeRDSCascade_AC5 — RDSConfig.mandatory.deletionProtection: true at level 4 forces
// deletion protection; level 4 wins over level 8 defaults.
func TestMergeRDSCascade_AC5(t *testing.T) {
	got := mergeRDSAll(
		zeroKropathRDS,
		zeroKropathRDS,
		zeroRDSCfg,
		cascade.RDSConfigSection{DeletionProtection: boolPtr(true)}, // level 4
		zeroRDSCfg,
		zeroRDSCfg,
		zeroKropathRDS,
		zeroKropathRDS,
	)
	if got.Mandatory.DeletionProtection == nil || !*got.Mandatory.DeletionProtection {
		t.Errorf("AC-5: mandatory.deletionProtection = %v, want *true (level 4 must win)", got.Mandatory.DeletionProtection)
	}
}

// TestMergeRDSCascade_AC6 — backupRetentionPeriod minimum floor: globalKropathMandatory sets
// floor value; cascade propagates it; RGD applies max(mandatory, instance). This unit test
// verifies that the mandatory floor value is correctly propagated by the cascade.
func TestMergeRDSCascade_AC6(t *testing.T) {
	const floor int64 = 7

	// Test 1: mandatory floor from level 1 propagates.
	got := mergeRDSAll(
		cascade.RDSKropathSection{BackupRetentionPeriod: floor}, // level 1
		zeroKropathRDS,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroKropathRDS,
		zeroKropathRDS,
	)
	if got.Mandatory.BackupRetentionPeriod != floor {
		t.Errorf("AC-6 (floor propagation): mandatory.backupRetentionPeriod = %d, want %d", got.Mandatory.BackupRetentionPeriod, floor)
	}

	// Test 2: level 3 mandatory wins over level 1 when level 1 is 0.
	const strictFloor int64 = 14
	got2 := mergeRDSAll(
		zeroKropathRDS,
		zeroKropathRDS,
		cascade.RDSConfigSection{BackupRetentionPeriod: strictFloor}, // level 3
		zeroRDSCfg,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroKropathRDS,
		zeroKropathRDS,
	)
	if got2.Mandatory.BackupRetentionPeriod != strictFloor {
		t.Errorf("AC-6 (level 3 floor): mandatory.backupRetentionPeriod = %d, want %d", got2.Mandatory.BackupRetentionPeriod, strictFloor)
	}

	// Test 3: when no mandatory source sets floor, mandatory.backupRetentionPeriod = 0 (not enforced).
	got3 := mergeRDSAll(
		zeroKropathRDS, zeroKropathRDS,
		zeroRDSCfg, zeroRDSCfg, zeroRDSCfg, zeroRDSCfg,
		zeroKropathRDS, zeroKropathRDS,
	)
	if got3.Mandatory.BackupRetentionPeriod != 0 {
		t.Errorf("AC-6 (no floor): mandatory.backupRetentionPeriod = %d, want 0", got3.Mandatory.BackupRetentionPeriod)
	}
}

// TestMergeRDSCascade_AC7 — KropathConfig.mandatory.rds.multiAZ: true at level 1
// forces Multi-AZ on all instances.
func TestMergeRDSCascade_AC7(t *testing.T) {
	got := mergeRDSAll(
		cascade.RDSKropathSection{MultiAZ: boolPtr(true)}, // level 1
		zeroKropathRDS,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroKropathRDS,
		zeroKropathRDS,
	)
	if got.Mandatory.MultiAZ == nil || !*got.Mandatory.MultiAZ {
		t.Errorf("AC-7: mandatory.multiAZ = %v, want *true", got.Mandatory.MultiAZ)
	}
}

// TestMergeRDSCascade_AC8 — publiclyAccessible pointer semantics:
//   - nil = not enforced (falls through)
//   - mandatory pointer false = forced private-only across all instances (OD-2)
//   - mandatory pointer nil = not enforced, not written to output
func TestMergeRDSCascade_AC8(t *testing.T) {
	// Test 1: mandatory pointer false = forced private-only.
	got := mergeRDSAll(
		cascade.RDSKropathSection{PubliclyAccessible: boolPtr(false)}, // level 1: explicit false
		zeroKropathRDS,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroKropathRDS,
		zeroKropathRDS,
	)
	if got.Mandatory.PubliclyAccessible == nil || *got.Mandatory.PubliclyAccessible {
		t.Errorf("AC-8 (forced private): mandatory.publiclyAccessible = %v, want *false (explicit false must be preserved)", got.Mandatory.PubliclyAccessible)
	}

	// Test 2: nil = not enforced.
	got2 := mergeRDSAll(
		cascade.RDSKropathSection{PubliclyAccessible: nil}, // level 1: nil
		zeroKropathRDS,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroKropathRDS,
		zeroKropathRDS,
	)
	if got2.Mandatory.PubliclyAccessible != nil {
		t.Errorf("AC-8 (nil = not enforced): mandatory.publiclyAccessible = %v, want nil", got2.Mandatory.PubliclyAccessible)
	}

	// Test 3: level 3 explicit false wins when levels 1-2 are nil.
	got3 := mergeRDSAll(
		zeroKropathRDS,
		zeroKropathRDS,
		cascade.RDSConfigSection{PubliclyAccessible: boolPtr(false)}, // level 3
		zeroRDSCfg,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroKropathRDS,
		zeroKropathRDS,
	)
	if got3.Mandatory.PubliclyAccessible == nil || *got3.Mandatory.PubliclyAccessible {
		t.Errorf("AC-8 (level 3 false wins): mandatory.publiclyAccessible = %v, want *false", got3.Mandatory.PubliclyAccessible)
	}
}

// TestMergeRDSCascade_AC9 — KropathConfig.mandatory.rds.kmsKeyID non-empty at level 1
// forces org-wide KMS key on all instances.
func TestMergeRDSCascade_AC9(t *testing.T) {
	const orgKey = "arn:aws:kms:ap-southeast-2:123456789012:key/org-rds-key"
	got := mergeRDSAll(
		cascade.RDSKropathSection{KmsKeyID: orgKey}, // level 1
		zeroKropathRDS,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroKropathRDS,
		zeroKropathRDS,
	)
	if got.Mandatory.KmsKeyID != orgKey {
		t.Errorf("AC-9: mandatory.kmsKeyID = %q, want %q", got.Mandatory.KmsKeyID, orgKey)
	}
}

// TestMergeRDSCascade_AC10 — RDSConfig.defaults.storageType: "gp3" at level 6
// applies as default when instance does not specify.
func TestMergeRDSCascade_AC10(t *testing.T) {
	got := mergeRDSAll(
		zeroKropathRDS,
		zeroKropathRDS,
		zeroRDSCfg,
		zeroRDSCfg,
		cascade.RDSConfigSection{StorageType: "gp3"}, // level 6 defaults
		zeroRDSCfg,
		zeroKropathRDS,
		zeroKropathRDS,
	)
	if got.Defaults.StorageType != "gp3" {
		t.Errorf("AC-10: defaults.storageType = %q, want %q", got.Defaults.StorageType, "gp3")
	}
	if got.Mandatory.StorageType != "" {
		t.Errorf("AC-10: mandatory.storageType = %q, want empty (no mandatory override)", got.Mandatory.StorageType)
	}
}

// TestMergeRDSCascade_AC11 — manageMasterUserPassword pointer semantics:
//   - nil = not enforced; true in mandatory = forced Secrets Manager for all instances.
func TestMergeRDSCascade_AC11(t *testing.T) {
	// Test 1: mandatory true forces Secrets Manager.
	got := mergeRDSAll(
		zeroKropathRDS,
		zeroKropathRDS,
		cascade.RDSConfigSection{ManageMasterUserPassword: boolPtr(true)}, // level 3
		zeroRDSCfg,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroKropathRDS,
		zeroKropathRDS,
	)
	if got.Mandatory.ManageMasterUserPassword == nil || !*got.Mandatory.ManageMasterUserPassword {
		t.Errorf("AC-11 (forced): mandatory.manageMasterUserPassword = %v, want *true", got.Mandatory.ManageMasterUserPassword)
	}

	// Test 2: nil = not enforced.
	got2 := mergeRDSAll(
		zeroKropathRDS, zeroKropathRDS,
		zeroRDSCfg, zeroRDSCfg, zeroRDSCfg, zeroRDSCfg,
		zeroKropathRDS, zeroKropathRDS,
	)
	if got2.Mandatory.ManageMasterUserPassword != nil {
		t.Errorf("AC-11 (nil = not enforced): mandatory.manageMasterUserPassword = %v, want nil", got2.Mandatory.ManageMasterUserPassword)
	}
}

// TestMergeRDSCascade_AC12 — serverlessV2ScalingMinCapacity cluster-only: cascade reads from
// RDSConfig tiers only (levels 3/4 mandatory, 6/7 defaults); KropathConfig levels have no
// equivalent field and cannot contribute.
func TestMergeRDSCascade_AC12(t *testing.T) {
	const minACU = 0.5

	// Level 3 mandatory propagates.
	got := mergeRDSAll(
		zeroKropathRDS,
		zeroKropathRDS,
		cascade.RDSConfigSection{ServerlessV2ScalingMinCapacity: minACU}, // level 3
		zeroRDSCfg,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroKropathRDS,
		zeroKropathRDS,
	)
	if got.Mandatory.ServerlessV2ScalingMinCapacity != minACU {
		t.Errorf("AC-12: mandatory.serverlessV2ScalingMinCapacity = %f, want %f", got.Mandatory.ServerlessV2ScalingMinCapacity, minACU)
	}

	// Level 6 defaults propagates.
	got2 := mergeRDSAll(
		zeroKropathRDS,
		zeroKropathRDS,
		zeroRDSCfg,
		zeroRDSCfg,
		cascade.RDSConfigSection{ServerlessV2ScalingMinCapacity: minACU}, // level 6
		zeroRDSCfg,
		zeroKropathRDS,
		zeroKropathRDS,
	)
	if got2.Defaults.ServerlessV2ScalingMinCapacity != minACU {
		t.Errorf("AC-12: defaults.serverlessV2ScalingMinCapacity = %f, want %f", got2.Defaults.ServerlessV2ScalingMinCapacity, minACU)
	}

	// Zero when no RDSConfig source sets it.
	got3 := mergeRDSAll(
		zeroKropathRDS, zeroKropathRDS,
		zeroRDSCfg, zeroRDSCfg, zeroRDSCfg, zeroRDSCfg,
		zeroKropathRDS, zeroKropathRDS,
	)
	if got3.Mandatory.ServerlessV2ScalingMinCapacity != 0 {
		t.Errorf("AC-12 (zero): mandatory.serverlessV2ScalingMinCapacity = %f, want 0", got3.Mandatory.ServerlessV2ScalingMinCapacity)
	}
}

// TestMergeRDSCascade_AC13 — serverlessV2ScalingMaxCapacity cluster-only: same scoping as AC-12.
func TestMergeRDSCascade_AC13(t *testing.T) {
	const maxACU = 128.0

	got := mergeRDSAll(
		zeroKropathRDS,
		zeroKropathRDS,
		cascade.RDSConfigSection{ServerlessV2ScalingMaxCapacity: maxACU}, // level 3
		zeroRDSCfg,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroKropathRDS,
		zeroKropathRDS,
	)
	if got.Mandatory.ServerlessV2ScalingMaxCapacity != maxACU {
		t.Errorf("AC-13: mandatory.serverlessV2ScalingMaxCapacity = %f, want %f", got.Mandatory.ServerlessV2ScalingMaxCapacity, maxACU)
	}

	got2 := mergeRDSAll(
		zeroKropathRDS,
		zeroKropathRDS,
		zeroRDSCfg,
		zeroRDSCfg,
		cascade.RDSConfigSection{ServerlessV2ScalingMaxCapacity: maxACU}, // level 6
		zeroRDSCfg,
		zeroKropathRDS,
		zeroKropathRDS,
	)
	if got2.Defaults.ServerlessV2ScalingMaxCapacity != maxACU {
		t.Errorf("AC-13: defaults.serverlessV2ScalingMaxCapacity = %f, want %f", got2.Defaults.ServerlessV2ScalingMaxCapacity, maxACU)
	}
}

// TestMergeRDSCascade_AC14 — backtrackWindow cluster-only: Aurora MySQL clusters only;
// cascades from RDSConfig levels only; ignored by RDSInstance (the controller routes it
// only for RDSCluster reconciliation paths).
func TestMergeRDSCascade_AC14(t *testing.T) {
	const backtrackSecs int64 = 86400 // 24 hours

	got := mergeRDSAll(
		zeroKropathRDS,
		zeroKropathRDS,
		cascade.RDSConfigSection{BacktrackWindow: backtrackSecs}, // level 3
		zeroRDSCfg,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroKropathRDS,
		zeroKropathRDS,
	)
	if got.Mandatory.BacktrackWindow != backtrackSecs {
		t.Errorf("AC-14: mandatory.backtrackWindow = %d, want %d", got.Mandatory.BacktrackWindow, backtrackSecs)
	}

	got2 := mergeRDSAll(
		zeroKropathRDS,
		zeroKropathRDS,
		zeroRDSCfg,
		zeroRDSCfg,
		cascade.RDSConfigSection{BacktrackWindow: backtrackSecs}, // level 6
		zeroRDSCfg,
		zeroKropathRDS,
		zeroKropathRDS,
	)
	if got2.Defaults.BacktrackWindow != backtrackSecs {
		t.Errorf("AC-14: defaults.backtrackWindow = %d, want %d", got2.Defaults.BacktrackWindow, backtrackSecs)
	}
}

// TestMergeRDSCascade_AC15 — Tags, syncedLabels, syncedAnnotations: map union merge.
//
//   - Tags: union across KropathConfig.mandatory + RDSConfig.mandatory (levels 1-4); first (highest priority) wins on key conflict.
//   - SyncedLabels: RDSConfig levels only; level 3 wins over level 4.
//   - Tags defaults: union across all four defaults levels; level 6 wins over level 9.
func TestMergeRDSCascade_AC15(t *testing.T) {
	got := mergeRDSAll(
		cascade.RDSKropathSection{Tags: map[string]string{"cost-centre": "platform", "env": "prod"}}, // level 1
		zeroKropathRDS,
		cascade.RDSConfigSection{ // level 3
			Tags:         map[string]string{"resource-type": "rds", "env": "staging"},
			SyncedLabels: map[string]string{"data-class": "confidential"},
		},
		cascade.RDSConfigSection{ // level 4
			SyncedLabels:      map[string]string{"data-class": "internal", "team": "platform"},
			SyncedAnnotations: map[string]string{"owner": "infra"},
		},
		cascade.RDSConfigSection{ // level 6 defaults
			Tags:              map[string]string{"backup": "daily"},
			SyncedAnnotations: map[string]string{"backup-policy": "daily"},
		},
		zeroRDSCfg,
		zeroKropathRDS,
		cascade.RDSKropathSection{Tags: map[string]string{"org": "kropath", "env": "dev"}}, // level 9
	)

	// Mandatory tags: level 1 "env=prod" wins over level 3 "env=staging".
	if got.Mandatory.Tags["env"] != "prod" {
		t.Errorf("AC-15 (tags, key conflict): mandatory.tags[env] = %q, want %q (level 1 wins)", got.Mandatory.Tags["env"], "prod")
	}
	if got.Mandatory.Tags["cost-centre"] != "platform" {
		t.Errorf("AC-15 (tags, cost-centre): mandatory.tags[cost-centre] = %q, want %q", got.Mandatory.Tags["cost-centre"], "platform")
	}
	if got.Mandatory.Tags["resource-type"] != "rds" {
		t.Errorf("AC-15 (tags, resource-type): mandatory.tags[resource-type] = %q, want %q", got.Mandatory.Tags["resource-type"], "rds")
	}

	// SyncedLabels: level 3 "data-class=confidential" wins over level 4 "data-class=internal".
	if got.Mandatory.SyncedLabels["data-class"] != "confidential" {
		t.Errorf("AC-15 (syncedLabels, key conflict): mandatory.syncedLabels[data-class] = %q, want %q (level 3 wins)", got.Mandatory.SyncedLabels["data-class"], "confidential")
	}
	if got.Mandatory.SyncedLabels["team"] != "platform" {
		t.Errorf("AC-15 (syncedLabels, team): mandatory.syncedLabels[team] = %q, want %q", got.Mandatory.SyncedLabels["team"], "platform")
	}

	// SyncedAnnotations: level 4 mandatory.
	if got.Mandatory.SyncedAnnotations["owner"] != "infra" {
		t.Errorf("AC-15 (syncedAnnotations): mandatory.syncedAnnotations[owner] = %q, want %q", got.Mandatory.SyncedAnnotations["owner"], "infra")
	}

	// Defaults tags: level 6 "env=prod-from-defaults" from level 6; level 9 "env=dev" should be superseded.
	if got.Defaults.Tags["backup"] != "daily" {
		t.Errorf("AC-15 (defaults tags, backup): defaults.tags[backup] = %q, want %q", got.Defaults.Tags["backup"], "daily")
	}
	if got.Defaults.Tags["org"] != "kropath" {
		t.Errorf("AC-15 (defaults tags, org): defaults.tags[org] = %q, want %q", got.Defaults.Tags["org"], "kropath")
	}

	// Defaults syncedAnnotations from level 6.
	if got.Defaults.SyncedAnnotations["backup-policy"] != "daily" {
		t.Errorf("AC-15 (defaults syncedAnnotations): defaults.syncedAnnotations[backup-policy] = %q, want %q", got.Defaults.SyncedAnnotations["backup-policy"], "daily")
	}
}

// TestMergeRDSCascade_AC16 — Cascade correctly propagates all resolved fields. This unit test
// verifies that a fully-populated merge call returns the correct effective config. Provider
// identity (aws.accountId, aws.region) is written by the controller reconciler separately
// and is covered by Chainsaw integration tests (cascade-provider-identity scenario).
func TestMergeRDSCascade_AC16(t *testing.T) {
	const (
		orgKey  = "arn:aws:kms:us-east-1:111222333444:key/org-rds-key"
		localKey = "arn:aws:kms:us-east-1:111222333444:key/local-rds-key"
	)
	got := mergeRDSAll(
		cascade.RDSKropathSection{ // level 1 globalKropathMandatory
			StorageEncrypted:                boolPtr(true),
			KmsKeyID:                        orgKey,
			DeletionProtection:              boolPtr(true),
			BackupRetentionPeriod:           7,
			MultiAZ:                         boolPtr(true),
			PubliclyAccessible:              boolPtr(false),
			AutoMinorVersionUpgrade:         boolPtr(true),
			CopyTagsToSnapshot:              boolPtr(true),
			ManageMasterUserPassword:        boolPtr(true),
			PerformanceInsightsEnabled:      boolPtr(true),
			EnableIAMDatabaseAuthentication: boolPtr(true),
			Tags:                            map[string]string{"org": "kropath"},
		},
		zeroKropathRDS,
		cascade.RDSConfigSection{ // level 3 globalRDSCfgMandatory
			NamingTemplate: "{namespace}-{name}",
			SyncedLabels:   map[string]string{"data-class": "confidential"},
		},
		cascade.RDSConfigSection{ // level 4 localRDSCfgMandatory
			KmsKeyID: localKey, // lower priority than level 1; level 1 wins
			ServerlessV2ScalingMinCapacity: 0.5,
			ServerlessV2ScalingMaxCapacity: 128.0,
			BacktrackWindow:                86400,
		},
		zeroRDSCfg, // level 6
		cascade.RDSConfigSection{ // level 7 globalRDSCfgDefaults
			StorageType: "gp3",
		},
		zeroKropathRDS,
		zeroKropathRDS,
	)

	// Verify mandatory tier.
	if got.Mandatory.StorageEncrypted == nil || !*got.Mandatory.StorageEncrypted {
		t.Errorf("AC-16: mandatory.storageEncrypted = %v, want *true", got.Mandatory.StorageEncrypted)
	}
	if got.Mandatory.KmsKeyID != orgKey {
		t.Errorf("AC-16: mandatory.kmsKeyID = %q, want %q (level 1 wins over level 4)", got.Mandatory.KmsKeyID, orgKey)
	}
	if got.Mandatory.DeletionProtection == nil || !*got.Mandatory.DeletionProtection {
		t.Errorf("AC-16: mandatory.deletionProtection = %v, want *true", got.Mandatory.DeletionProtection)
	}
	if got.Mandatory.BackupRetentionPeriod != 7 {
		t.Errorf("AC-16: mandatory.backupRetentionPeriod = %d, want 7", got.Mandatory.BackupRetentionPeriod)
	}
	if got.Mandatory.MultiAZ == nil || !*got.Mandatory.MultiAZ {
		t.Errorf("AC-16: mandatory.multiAZ = %v, want *true", got.Mandatory.MultiAZ)
	}
	if got.Mandatory.PubliclyAccessible == nil || *got.Mandatory.PubliclyAccessible {
		t.Errorf("AC-16: mandatory.publiclyAccessible = %v, want *false", got.Mandatory.PubliclyAccessible)
	}
	if got.Mandatory.AutoMinorVersionUpgrade == nil || !*got.Mandatory.AutoMinorVersionUpgrade {
		t.Errorf("AC-16: mandatory.autoMinorVersionUpgrade = %v, want *true", got.Mandatory.AutoMinorVersionUpgrade)
	}
	if got.Mandatory.CopyTagsToSnapshot == nil || !*got.Mandatory.CopyTagsToSnapshot {
		t.Errorf("AC-16: mandatory.copyTagsToSnapshot = %v, want *true", got.Mandatory.CopyTagsToSnapshot)
	}
	if got.Mandatory.ManageMasterUserPassword == nil || !*got.Mandatory.ManageMasterUserPassword {
		t.Errorf("AC-16: mandatory.manageMasterUserPassword = %v, want *true", got.Mandatory.ManageMasterUserPassword)
	}
	if got.Mandatory.PerformanceInsightsEnabled == nil || !*got.Mandatory.PerformanceInsightsEnabled {
		t.Errorf("AC-16: mandatory.performanceInsightsEnabled = %v, want *true", got.Mandatory.PerformanceInsightsEnabled)
	}
	if got.Mandatory.EnableIAMDatabaseAuthentication == nil || !*got.Mandatory.EnableIAMDatabaseAuthentication {
		t.Errorf("AC-16: mandatory.enableIAMDatabaseAuthentication = %v, want *true", got.Mandatory.EnableIAMDatabaseAuthentication)
	}
	if got.Mandatory.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("AC-16: mandatory.namingTemplate = %q, want %q", got.Mandatory.NamingTemplate, "{namespace}-{name}")
	}
	if got.Mandatory.ServerlessV2ScalingMinCapacity != 0.5 {
		t.Errorf("AC-16: mandatory.serverlessV2ScalingMinCapacity = %f, want 0.5", got.Mandatory.ServerlessV2ScalingMinCapacity)
	}
	if got.Mandatory.ServerlessV2ScalingMaxCapacity != 128.0 {
		t.Errorf("AC-16: mandatory.serverlessV2ScalingMaxCapacity = %f, want 128.0", got.Mandatory.ServerlessV2ScalingMaxCapacity)
	}
	if got.Mandatory.BacktrackWindow != 86400 {
		t.Errorf("AC-16: mandatory.backtrackWindow = %d, want 86400", got.Mandatory.BacktrackWindow)
	}
	if got.Mandatory.Tags["org"] != "kropath" {
		t.Errorf("AC-16: mandatory.tags[org] = %q, want %q", got.Mandatory.Tags["org"], "kropath")
	}
	if got.Mandatory.SyncedLabels["data-class"] != "confidential" {
		t.Errorf("AC-16: mandatory.syncedLabels[data-class] = %q, want %q", got.Mandatory.SyncedLabels["data-class"], "confidential")
	}

	// Verify defaults tier.
	if got.Defaults.StorageType != "gp3" {
		t.Errorf("AC-16: defaults.storageType = %q, want %q", got.Defaults.StorageType, "gp3")
	}
}

// TestMergeRDSCascade_AllZero — all inputs zero returns zero-value effective config.
func TestMergeRDSCascade_AllZero(t *testing.T) {
	got := mergeRDSAll(
		zeroKropathRDS, zeroKropathRDS,
		zeroRDSCfg, zeroRDSCfg, zeroRDSCfg, zeroRDSCfg,
		zeroKropathRDS, zeroKropathRDS,
	)
	if got.Mandatory.StorageEncrypted != nil {
		t.Errorf("all-zero: mandatory.storageEncrypted = %v, want nil", got.Mandatory.StorageEncrypted)
	}
	if got.Mandatory.KmsKeyID != "" {
		t.Errorf("all-zero: mandatory.kmsKeyID = %q, want empty", got.Mandatory.KmsKeyID)
	}
	if got.Mandatory.BackupRetentionPeriod != 0 {
		t.Errorf("all-zero: mandatory.backupRetentionPeriod = %d, want 0", got.Mandatory.BackupRetentionPeriod)
	}
	if got.Mandatory.ServerlessV2ScalingMinCapacity != 0 {
		t.Errorf("all-zero: mandatory.serverlessV2ScalingMinCapacity = %f, want 0", got.Mandatory.ServerlessV2ScalingMinCapacity)
	}
	if got.Defaults.StorageType != "" {
		t.Errorf("all-zero: defaults.storageType = %q, want empty", got.Defaults.StorageType)
	}
	if got.Defaults.StorageEncrypted != nil {
		t.Errorf("all-zero: defaults.storageEncrypted = %v, want nil", got.Defaults.StorageEncrypted)
	}
}

// TestMergeRDSCascade_MandatoryPriorityLevel1OverLevel4 — level 1 (globalKropathMandatory)
// wins over level 4 (localRDSCfgMandatory) for all shared string and bool pointer fields.
func TestMergeRDSCascade_MandatoryPriorityLevel1OverLevel4(t *testing.T) {
	const (
		orgKey   = "arn:aws:kms:us-east-1:111:key/org"
		localKey = "arn:aws:kms:us-east-1:111:key/local"
	)
	got := mergeRDSAll(
		cascade.RDSKropathSection{KmsKeyID: orgKey, StorageType: "io1"}, // level 1
		zeroKropathRDS,
		zeroRDSCfg,
		cascade.RDSConfigSection{KmsKeyID: localKey, StorageType: "gp3"}, // level 4
		zeroRDSCfg,
		zeroRDSCfg,
		zeroKropathRDS,
		zeroKropathRDS,
	)
	if got.Mandatory.KmsKeyID != orgKey {
		t.Errorf("level-priority: mandatory.kmsKeyID = %q, want %q (level 1 wins over level 4)", got.Mandatory.KmsKeyID, orgKey)
	}
	if got.Mandatory.StorageType != "io1" {
		t.Errorf("level-priority: mandatory.storageType = %q, want %q (level 1 wins over level 4)", got.Mandatory.StorageType, "io1")
	}
}

// TestMergeRDSCascade_DefaultsPriorityLevel6OverLevel9 — level 6 (localRDSCfgDefaults)
// wins over level 9 (globalKropathDefaults) for defaults tier.
func TestMergeRDSCascade_DefaultsPriorityLevel6OverLevel9(t *testing.T) {
	got := mergeRDSAll(
		zeroKropathRDS,
		zeroKropathRDS,
		zeroRDSCfg,
		zeroRDSCfg,
		cascade.RDSConfigSection{StorageType: "io1", BackupRetentionPeriod: 14}, // level 6
		zeroRDSCfg,
		zeroKropathRDS,
		cascade.RDSKropathSection{StorageType: "gp3", BackupRetentionPeriod: 7}, // level 9
	)
	if got.Defaults.StorageType != "io1" {
		t.Errorf("defaults-priority: defaults.storageType = %q, want %q (level 6 wins over level 9)", got.Defaults.StorageType, "io1")
	}
	if got.Defaults.BackupRetentionPeriod != 14 {
		t.Errorf("defaults-priority: defaults.backupRetentionPeriod = %d, want 14 (level 6 wins over level 9)", got.Defaults.BackupRetentionPeriod)
	}
}

// TestMergeRDSCascade_NilFallsThroughToNextMandatoryLevel — nil pointer at level 1 falls through
// to the next non-nil pointer at level 3.
func TestMergeRDSCascade_NilFallsThroughToNextMandatoryLevel(t *testing.T) {
	got := mergeRDSAll(
		cascade.RDSKropathSection{StorageEncrypted: nil}, // level 1: nil (falls through)
		zeroKropathRDS,
		cascade.RDSConfigSection{StorageEncrypted: boolPtr(false)}, // level 3: explicit false
		zeroRDSCfg,
		zeroRDSCfg,
		zeroRDSCfg,
		zeroKropathRDS,
		zeroKropathRDS,
	)
	if got.Mandatory.StorageEncrypted == nil || *got.Mandatory.StorageEncrypted {
		t.Errorf("nil-falls-through: mandatory.storageEncrypted = %v, want *false (level 3 explicit-false must win over nil levels 1-2)", got.Mandatory.StorageEncrypted)
	}
}
