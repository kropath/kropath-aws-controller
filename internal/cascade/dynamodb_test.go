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

// boolPtr returns a pointer to the given bool value.
func boolPtr(b bool) *bool { return &b }

// zeroKropathDDB is a zero-value DynamoDBKropathSection (absent source).
var zeroKropathDDB = cascade.DynamoDBKropathSection{}

// zeroDDBCfg is a zero-value DynamoDBConfigSection (absent source).
var zeroDDBCfg = cascade.DynamoDBConfigSection{}

// mergeDDBAll calls MergeDynamoDBCascade with all eight inputs.
func mergeDDBAll(
	globalKropathMandatory,
	localKropathMandatory cascade.DynamoDBKropathSection,
	globalDDBCfgMandatory,
	localDDBCfgMandatory,
	localDDBCfgDefaults,
	globalDDBCfgDefaults cascade.DynamoDBConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.DynamoDBKropathSection,
) cascade.EffectiveDynamoDBConfig {
	return cascade.MergeDynamoDBCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalDDBCfgMandatory,
		localDDBCfgMandatory,
		localDDBCfgDefaults,
		globalDDBCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeDynamoDBCascade_AC1 — globalKropathConfig.mandatory.dynamodb.encryptionEnabled=true
// at level 1 propagates to effCfg.mandatory.encryptionEnabled=true (level 1 wins).
func TestMergeDynamoDBCascade_AC1(t *testing.T) {
	got := mergeDDBAll(
		cascade.DynamoDBKropathSection{EncryptionEnabled: boolPtr(true)}, // level 1
		zeroKropathDDB,
		zeroDDBCfg,
		zeroDDBCfg,
		zeroDDBCfg,
		zeroDDBCfg,
		zeroKropathDDB,
		zeroKropathDDB,
	)
	if got.Mandatory.EncryptionEnabled == nil || !*got.Mandatory.EncryptionEnabled {
		t.Errorf("AC-1: mandatory.encryptionEnabled = %v, want *true", got.Mandatory.EncryptionEnabled)
	}
}

// TestMergeDynamoDBCascade_AC2 — globalDynamoDBConfig.mandatory.encryptionEnabled=true at
// level 3 wins when levels 1-2 are nil.
func TestMergeDynamoDBCascade_AC2(t *testing.T) {
	got := mergeDDBAll(
		zeroKropathDDB,
		zeroKropathDDB,
		cascade.DynamoDBConfigSection{EncryptionEnabled: boolPtr(true)}, // level 3
		zeroDDBCfg,
		zeroDDBCfg,
		zeroDDBCfg,
		zeroKropathDDB,
		zeroKropathDDB,
	)
	if got.Mandatory.EncryptionEnabled == nil || !*got.Mandatory.EncryptionEnabled {
		t.Errorf("AC-2: mandatory.encryptionEnabled = %v, want *true", got.Mandatory.EncryptionEnabled)
	}
}

// TestMergeDynamoDBCascade_AC3 — localDynamoDBConfig.defaults.encryptionEnabled=false at
// level 6 propagates; explicit false is preserved (not treated as nil).
func TestMergeDynamoDBCascade_AC3(t *testing.T) {
	got := mergeDDBAll(
		zeroKropathDDB,
		zeroKropathDDB,
		zeroDDBCfg,
		zeroDDBCfg,
		cascade.DynamoDBConfigSection{EncryptionEnabled: boolPtr(false)}, // level 6
		zeroDDBCfg,
		zeroKropathDDB,
		zeroKropathDDB,
	)
	if got.Defaults.EncryptionEnabled == nil || *got.Defaults.EncryptionEnabled {
		t.Errorf("AC-3: defaults.encryptionEnabled = %v, want *false (explicit false must be preserved)", got.Defaults.EncryptionEnabled)
	}
	if got.Mandatory.EncryptionEnabled != nil {
		t.Errorf("AC-3: mandatory.encryptionEnabled = %v, want nil", got.Mandatory.EncryptionEnabled)
	}
}

// TestMergeDynamoDBCascade_AC4 — globalKropathConfig.mandatory.dynamodb.billingMode="PROVISIONED"
// at level 1 wins.
func TestMergeDynamoDBCascade_AC4(t *testing.T) {
	got := mergeDDBAll(
		cascade.DynamoDBKropathSection{BillingMode: "PROVISIONED"}, // level 1
		zeroKropathDDB,
		zeroDDBCfg,
		zeroDDBCfg,
		zeroDDBCfg,
		zeroDDBCfg,
		zeroKropathDDB,
		zeroKropathDDB,
	)
	if got.Mandatory.BillingMode != "PROVISIONED" {
		t.Errorf("AC-4: mandatory.billingMode = %q, want %q", got.Mandatory.BillingMode, "PROVISIONED")
	}
}

// TestMergeDynamoDBCascade_AC5 — globalDynamoDBConfig.defaults.billingMode="PAY_PER_REQUEST"
// at level 7 wins over globalKropathConfig.defaults.dynamodb.billingMode="PROVISIONED" at level 9.
func TestMergeDynamoDBCascade_AC5(t *testing.T) {
	got := mergeDDBAll(
		zeroKropathDDB,
		zeroKropathDDB,
		zeroDDBCfg,
		zeroDDBCfg,
		zeroDDBCfg,
		cascade.DynamoDBConfigSection{BillingMode: "PAY_PER_REQUEST"}, // level 7
		zeroKropathDDB,
		cascade.DynamoDBKropathSection{BillingMode: "PROVISIONED"}, // level 9
	)
	if got.Defaults.BillingMode != "PAY_PER_REQUEST" {
		t.Errorf("AC-5: defaults.billingMode = %q, want %q (level 7 must win over level 9)", got.Defaults.BillingMode, "PAY_PER_REQUEST")
	}
}

// TestMergeDynamoDBCascade_AC6 — globalKropathConfig.mandatory.dynamodb.deletionProtectionEnabled=true
// at level 1 propagates.
func TestMergeDynamoDBCascade_AC6(t *testing.T) {
	got := mergeDDBAll(
		cascade.DynamoDBKropathSection{DeletionProtectionEnabled: boolPtr(true)}, // level 1
		zeroKropathDDB,
		zeroDDBCfg,
		zeroDDBCfg,
		zeroDDBCfg,
		zeroDDBCfg,
		zeroKropathDDB,
		zeroKropathDDB,
	)
	if got.Mandatory.DeletionProtectionEnabled == nil || !*got.Mandatory.DeletionProtectionEnabled {
		t.Errorf("AC-6: mandatory.deletionProtectionEnabled = %v, want *true", got.Mandatory.DeletionProtectionEnabled)
	}
}

// TestMergeDynamoDBCascade_AC7 — globalDynamoDBConfig.mandatory.pointInTimeRecoveryEnabled=true
// at level 3 propagates when levels 1-2 are nil.
func TestMergeDynamoDBCascade_AC7(t *testing.T) {
	got := mergeDDBAll(
		zeroKropathDDB,
		zeroKropathDDB,
		cascade.DynamoDBConfigSection{PointInTimeRecoveryEnabled: boolPtr(true)}, // level 3
		zeroDDBCfg,
		zeroDDBCfg,
		zeroDDBCfg,
		zeroKropathDDB,
		zeroKropathDDB,
	)
	if got.Mandatory.PointInTimeRecoveryEnabled == nil || !*got.Mandatory.PointInTimeRecoveryEnabled {
		t.Errorf("AC-7: mandatory.pointInTimeRecoveryEnabled = %v, want *true", got.Mandatory.PointInTimeRecoveryEnabled)
	}
}

// TestMergeDynamoDBCascade_AC8 — globalDynamoDBConfig.defaults.tableClass="STANDARD_INFREQUENT_ACCESS"
// at level 7 propagates.
func TestMergeDynamoDBCascade_AC8(t *testing.T) {
	got := mergeDDBAll(
		zeroKropathDDB,
		zeroKropathDDB,
		zeroDDBCfg,
		zeroDDBCfg,
		zeroDDBCfg,
		cascade.DynamoDBConfigSection{TableClass: "STANDARD_INFREQUENT_ACCESS"}, // level 7
		zeroKropathDDB,
		zeroKropathDDB,
	)
	if got.Defaults.TableClass != "STANDARD_INFREQUENT_ACCESS" {
		t.Errorf("AC-8: defaults.tableClass = %q, want %q", got.Defaults.TableClass, "STANDARD_INFREQUENT_ACCESS")
	}
}

// TestMergeDynamoDBCascade_AC9 — globalKropathConfig.mandatory.dynamodb.contributorInsights="ENABLE"
// at level 1 wins over DynamoDBConfig.mandatory.contributorInsights="" at levels 3-4.
func TestMergeDynamoDBCascade_AC9(t *testing.T) {
	got := mergeDDBAll(
		cascade.DynamoDBKropathSection{ContributorInsights: "ENABLE"}, // level 1
		zeroKropathDDB,
		zeroDDBCfg, // level 3 empty
		zeroDDBCfg, // level 4 empty
		zeroDDBCfg,
		zeroDDBCfg,
		zeroKropathDDB,
		zeroKropathDDB,
	)
	if got.Mandatory.ContributorInsights != "ENABLE" {
		t.Errorf("AC-9: mandatory.contributorInsights = %q, want %q", got.Mandatory.ContributorInsights, "ENABLE")
	}
}

// TestMergeDynamoDBCascade_AC10 — Tags union merge: KropathConfig.mandatory.tags and
// DynamoDBConfig.mandatory.tags both appear in effCfg.mandatory.tags.
func TestMergeDynamoDBCascade_AC10(t *testing.T) {
	got := mergeDDBAll(
		cascade.DynamoDBKropathSection{Tags: map[string]string{"cost-centre": "infra"}}, // level 1
		zeroKropathDDB,
		cascade.DynamoDBConfigSection{Tags: map[string]string{"resource-type": "database"}}, // level 3
		zeroDDBCfg,
		zeroDDBCfg,
		zeroDDBCfg,
		zeroKropathDDB,
		zeroKropathDDB,
	)
	if got.Mandatory.Tags["cost-centre"] != "infra" {
		t.Errorf("AC-10: mandatory.tags[cost-centre] = %q, want %q", got.Mandatory.Tags["cost-centre"], "infra")
	}
	if got.Mandatory.Tags["resource-type"] != "database" {
		t.Errorf("AC-10: mandatory.tags[resource-type] = %q, want %q", got.Mandatory.Tags["resource-type"], "database")
	}
}

// TestMergeDynamoDBCascade_AC11 — globalKropathConfig.mandatory.dynamodb.kmsMasterKeyId
// at level 1 propagates.
func TestMergeDynamoDBCascade_AC11(t *testing.T) {
	const keyArn = "arn:aws:kms:ap-southeast-2:123:key/org-key"
	got := mergeDDBAll(
		cascade.DynamoDBKropathSection{KmsMasterKeyId: keyArn}, // level 1
		zeroKropathDDB,
		zeroDDBCfg,
		zeroDDBCfg,
		zeroDDBCfg,
		zeroDDBCfg,
		zeroKropathDDB,
		zeroKropathDDB,
	)
	if got.Mandatory.KmsMasterKeyId != keyArn {
		t.Errorf("AC-11: mandatory.kmsMasterKeyId = %q, want %q", got.Mandatory.KmsMasterKeyId, keyArn)
	}
}

// TestMergeDynamoDBCascade_AC12 — globalDynamoDBConfig.defaults.namingTemplate at level 7
// propagates. KropathConfig.dynamodb has no namingTemplate field (levels 1-2, 8-9 absent).
func TestMergeDynamoDBCascade_AC12(t *testing.T) {
	got := mergeDDBAll(
		zeroKropathDDB,
		zeroKropathDDB,
		zeroDDBCfg,
		zeroDDBCfg,
		zeroDDBCfg,
		cascade.DynamoDBConfigSection{NamingTemplate: "{namespace}-{name}"}, // level 7
		zeroKropathDDB,
		zeroKropathDDB,
	)
	if got.Defaults.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("AC-12: defaults.namingTemplate = %q, want %q", got.Defaults.NamingTemplate, "{namespace}-{name}")
	}
}

// TestMergeDynamoDBCascade_AC13 — globalDynamoDBConfig.mandatory.syncedLabels at level 3
// propagates into effCfg.mandatory.syncedLabels.
func TestMergeDynamoDBCascade_AC13(t *testing.T) {
	got := mergeDDBAll(
		zeroKropathDDB,
		zeroKropathDDB,
		cascade.DynamoDBConfigSection{SyncedLabels: map[string]string{"data-class": "internal"}}, // level 3
		zeroDDBCfg,
		zeroDDBCfg,
		zeroDDBCfg,
		zeroKropathDDB,
		zeroKropathDDB,
	)
	if got.Mandatory.SyncedLabels["data-class"] != "internal" {
		t.Errorf("AC-13: mandatory.syncedLabels[data-class] = %q, want %q", got.Mandatory.SyncedLabels["data-class"], "internal")
	}
}

// TestFirstNonNilBoolPtr_NilFallsThrough verifies that nil falls through to the next candidate.
func TestFirstNonNilBoolPtr_NilFallsThrough(t *testing.T) {
	got := mergeDDBAll(
		cascade.DynamoDBKropathSection{EncryptionEnabled: nil}, // level 1: nil
		zeroKropathDDB,                                          // level 2: nil
		cascade.DynamoDBConfigSection{EncryptionEnabled: boolPtr(false)}, // level 3: explicit false
		zeroDDBCfg,
		zeroDDBCfg,
		zeroDDBCfg,
		zeroKropathDDB,
		zeroKropathDDB,
	)
	if got.Mandatory.EncryptionEnabled == nil || *got.Mandatory.EncryptionEnabled {
		t.Errorf("nil-falls-through: mandatory.encryptionEnabled = %v, want *false (level 3 explicit-false must win over nil levels 1-2)", got.Mandatory.EncryptionEnabled)
	}
}

// TestMergeDynamoDBCascade_BillingModeLevel1WinsOverLevel3 verifies mandatory priority.
func TestMergeDynamoDBCascade_BillingModeLevel1WinsOverLevel3(t *testing.T) {
	got := mergeDDBAll(
		cascade.DynamoDBKropathSection{BillingMode: "PROVISIONED"}, // level 1
		zeroKropathDDB,
		cascade.DynamoDBConfigSection{BillingMode: "PAY_PER_REQUEST"}, // level 3
		zeroDDBCfg,
		zeroDDBCfg,
		zeroDDBCfg,
		zeroKropathDDB,
		zeroKropathDDB,
	)
	if got.Mandatory.BillingMode != "PROVISIONED" {
		t.Errorf("mandatory priority: billingMode = %q, want %q (level 1 must win over level 3)", got.Mandatory.BillingMode, "PROVISIONED")
	}
}

// TestMergeDynamoDBCascade_TagsUnionDefaults verifies tag union in defaults tier.
func TestMergeDynamoDBCascade_TagsUnionDefaults(t *testing.T) {
	got := mergeDDBAll(
		zeroKropathDDB,
		zeroKropathDDB,
		zeroDDBCfg,
		zeroDDBCfg,
		cascade.DynamoDBConfigSection{Tags: map[string]string{"tier": "gold"}}, // level 6
		zeroDDBCfg,
		zeroKropathDDB,
		cascade.DynamoDBKropathSection{Tags: map[string]string{"env": "prod"}}, // level 9
	)
	if got.Defaults.Tags["tier"] != "gold" {
		t.Errorf("defaults tags: tier = %q, want %q", got.Defaults.Tags["tier"], "gold")
	}
	if got.Defaults.Tags["env"] != "prod" {
		t.Errorf("defaults tags: env = %q, want %q", got.Defaults.Tags["env"], "prod")
	}
}

// TestMergeDynamoDBCascade_AllZero — all inputs zero returns zero-value effective config.
func TestMergeDynamoDBCascade_AllZero(t *testing.T) {
	got := mergeDDBAll(
		zeroKropathDDB, zeroKropathDDB,
		zeroDDBCfg, zeroDDBCfg, zeroDDBCfg, zeroDDBCfg,
		zeroKropathDDB, zeroKropathDDB,
	)
	if got.Mandatory.EncryptionEnabled != nil {
		t.Errorf("all-zero: mandatory.encryptionEnabled = %v, want nil", got.Mandatory.EncryptionEnabled)
	}
	if got.Mandatory.BillingMode != "" {
		t.Errorf("all-zero: mandatory.billingMode = %q, want empty", got.Mandatory.BillingMode)
	}
	if got.Defaults.EncryptionEnabled != nil {
		t.Errorf("all-zero: defaults.encryptionEnabled = %v, want nil", got.Defaults.EncryptionEnabled)
	}
}
