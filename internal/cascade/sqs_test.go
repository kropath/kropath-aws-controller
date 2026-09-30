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

// zeroKropathSQS is a zero-value SQSKropathSection (absent source).
var zeroKropathSQS = cascade.SQSKropathSection{}

// zeroSQSCfg is a zero-value SQSConfigSection (absent source).
var zeroSQSCfg = cascade.SQSConfigSection{}

// mergeSQSAll calls MergeSQSCascade with all eight inputs.
func mergeSQSAll(
	globalKropathMandatory,
	localKropathMandatory cascade.SQSKropathSection,
	globalSQSCfgMandatory,
	localSQSCfgMandatory,
	localSQSCfgDefaults,
	globalSQSCfgDefaults cascade.SQSConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.SQSKropathSection,
) cascade.EffectiveSQSConfig {
	return cascade.MergeSQSCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalSQSCfgMandatory,
		localSQSCfgMandatory,
		localSQSCfgDefaults,
		globalSQSCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeSQSCascade_AC1 — globalKropathConfig.mandatory.sqs.encryptionType="kms" at
// level 1 propagates to effCfg.mandatory.encryptionType (level 1 wins).
func TestMergeSQSCascade_AC1(t *testing.T) {
	got := mergeSQSAll(
		cascade.SQSKropathSection{EncryptionType: "kms"}, // level 1
		zeroKropathSQS,
		zeroSQSCfg,
		zeroSQSCfg,
		zeroSQSCfg,
		zeroSQSCfg,
		zeroKropathSQS,
		zeroKropathSQS,
	)

	if got.Mandatory.EncryptionType != "kms" {
		t.Errorf("AC-1: mandatory.encryptionType = %q, want kms", got.Mandatory.EncryptionType)
	}
	if got.Defaults.EncryptionType != "" {
		t.Errorf("AC-1: defaults.encryptionType = %q, must not bleed from mandatory", got.Defaults.EncryptionType)
	}
}

// TestMergeSQSCascade_AC2 — only globalSQSConfig.mandatory.encryptionType="kms" set at
// level 3 (levels 1-2 empty); controller propagates level 3 to effCfg.mandatory.encryptionType.
func TestMergeSQSCascade_AC2(t *testing.T) {
	got := mergeSQSAll(
		zeroKropathSQS,
		zeroKropathSQS,
		cascade.SQSConfigSection{EncryptionType: "kms"}, // level 3
		zeroSQSCfg,
		zeroSQSCfg,
		zeroSQSCfg,
		zeroKropathSQS,
		zeroKropathSQS,
	)

	if got.Mandatory.EncryptionType != "kms" {
		t.Errorf("AC-2: mandatory.encryptionType = %q, want kms (level 3 wins when 1-2 absent)", got.Mandatory.EncryptionType)
	}
}

// TestMergeSQSCascade_AC3 — only localSQSConfig.defaults.encryptionType="sqs-managed" set
// at level 6; mandatory must be empty.
func TestMergeSQSCascade_AC3(t *testing.T) {
	got := mergeSQSAll(
		zeroKropathSQS,
		zeroKropathSQS,
		zeroSQSCfg,
		zeroSQSCfg,
		cascade.SQSConfigSection{EncryptionType: "sqs-managed"}, // level 6
		zeroSQSCfg,
		zeroKropathSQS,
		zeroKropathSQS,
	)

	if got.Mandatory.EncryptionType != "" {
		t.Errorf("AC-3: mandatory.encryptionType = %q, want empty", got.Mandatory.EncryptionType)
	}
	if got.Defaults.EncryptionType != "sqs-managed" {
		t.Errorf("AC-3: defaults.encryptionType = %q, want sqs-managed (level 6)", got.Defaults.EncryptionType)
	}
}

// TestMergeSQSCascade_AC4 — globalKropathConfig.mandatory.sqs.visibilityTimeout=120
// at level 1 propagates to effCfg.mandatory.visibilityTimeout.
func TestMergeSQSCascade_AC4(t *testing.T) {
	got := mergeSQSAll(
		cascade.SQSKropathSection{VisibilityTimeout: 120}, // level 1
		zeroKropathSQS,
		zeroSQSCfg,
		zeroSQSCfg,
		zeroSQSCfg,
		zeroSQSCfg,
		zeroKropathSQS,
		zeroKropathSQS,
	)

	if got.Mandatory.VisibilityTimeout != 120 {
		t.Errorf("AC-4: mandatory.visibilityTimeout = %d, want 120", got.Mandatory.VisibilityTimeout)
	}
	if got.Defaults.VisibilityTimeout != 0 {
		t.Errorf("AC-4: defaults.visibilityTimeout = %d, must not bleed from mandatory", got.Defaults.VisibilityTimeout)
	}
}

// TestMergeSQSCascade_AC5 — globalSQSConfig.defaults.visibilityTimeout=30 (level 7)
// wins over KropathConfig.defaults.sqs.visibilityTimeout=60 (level 9).
func TestMergeSQSCascade_AC5(t *testing.T) {
	got := mergeSQSAll(
		zeroKropathSQS,
		zeroKropathSQS,
		zeroSQSCfg,
		zeroSQSCfg,
		zeroSQSCfg,
		cascade.SQSConfigSection{VisibilityTimeout: 30},  // level 7
		zeroKropathSQS,
		cascade.SQSKropathSection{VisibilityTimeout: 60}, // level 9
	)

	if got.Defaults.VisibilityTimeout != 30 {
		t.Errorf("AC-5: defaults.visibilityTimeout = %d, want 30 (level 7 wins over level 9)", got.Defaults.VisibilityTimeout)
	}
}

// TestMergeSQSCascade_AC6 — globalKropathConfig.mandatory.sqs.messageRetentionPeriod=604800
// propagates to effCfg.mandatory.messageRetentionPeriod.
func TestMergeSQSCascade_AC6(t *testing.T) {
	got := mergeSQSAll(
		cascade.SQSKropathSection{MessageRetentionPeriod: 604800}, // level 1
		zeroKropathSQS,
		zeroSQSCfg,
		zeroSQSCfg,
		zeroSQSCfg,
		zeroSQSCfg,
		zeroKropathSQS,
		zeroKropathSQS,
	)

	if got.Mandatory.MessageRetentionPeriod != 604800 {
		t.Errorf("AC-6: mandatory.messageRetentionPeriod = %d, want 604800", got.Mandatory.MessageRetentionPeriod)
	}
}

// TestMergeSQSCascade_AC7 — globalSQSConfig.mandatory.delaySeconds=60 (level 3) propagates
// to effCfg.mandatory.delaySeconds.
func TestMergeSQSCascade_AC7(t *testing.T) {
	got := mergeSQSAll(
		zeroKropathSQS,
		zeroKropathSQS,
		cascade.SQSConfigSection{DelaySeconds: 60}, // level 3
		zeroSQSCfg,
		zeroSQSCfg,
		zeroSQSCfg,
		zeroKropathSQS,
		zeroKropathSQS,
	)

	if got.Mandatory.DelaySeconds != 60 {
		t.Errorf("AC-7: mandatory.delaySeconds = %d, want 60", got.Mandatory.DelaySeconds)
	}
}

// TestMergeSQSCascade_AC8 — globalSQSConfig.defaults.maximumMessageSize=131072 (level 7)
// propagates to effCfg.defaults.maximumMessageSize.
func TestMergeSQSCascade_AC8(t *testing.T) {
	got := mergeSQSAll(
		zeroKropathSQS,
		zeroKropathSQS,
		zeroSQSCfg,
		zeroSQSCfg,
		zeroSQSCfg,
		cascade.SQSConfigSection{MaximumMessageSize: 131072}, // level 7
		zeroKropathSQS,
		zeroKropathSQS,
	)

	if got.Defaults.MaximumMessageSize != 131072 {
		t.Errorf("AC-8: defaults.maximumMessageSize = %d, want 131072", got.Defaults.MaximumMessageSize)
	}
}

// TestMergeSQSCascade_AC9 — KropathConfig.mandatory.tags and SQSConfig.mandatory.tags are
// union-merged into effCfg.mandatory.tags; KropathConfig keys win on conflict.
func TestMergeSQSCascade_AC9(t *testing.T) {
	got := mergeSQSAll(
		cascade.SQSKropathSection{Tags: map[string]string{"cost-centre": "infra"}},        // level 1
		zeroKropathSQS,
		zeroSQSCfg,
		cascade.SQSConfigSection{Tags: map[string]string{"queue-type": "messaging"}}, // level 4
		zeroSQSCfg,
		zeroSQSCfg,
		zeroKropathSQS,
		zeroKropathSQS,
	)

	if got.Mandatory.Tags["cost-centre"] != "infra" {
		t.Errorf("AC-9: mandatory.tags[cost-centre] = %q, want infra", got.Mandatory.Tags["cost-centre"])
	}
	if got.Mandatory.Tags["queue-type"] != "messaging" {
		t.Errorf("AC-9: mandatory.tags[queue-type] = %q, want messaging", got.Mandatory.Tags["queue-type"])
	}
}

// TestMergeSQSCascade_AC10 — globalKropathConfig.mandatory.sqs.kmsMasterKeyId propagates
// to effCfg.mandatory.kmsMasterKeyId (level 1).
func TestMergeSQSCascade_AC10(t *testing.T) {
	keyID := "arn:aws:kms:ap-southeast-2:123456789012:key/org-key"
	got := mergeSQSAll(
		cascade.SQSKropathSection{KmsMasterKeyId: keyID}, // level 1
		zeroKropathSQS,
		zeroSQSCfg,
		zeroSQSCfg,
		zeroSQSCfg,
		zeroSQSCfg,
		zeroKropathSQS,
		zeroKropathSQS,
	)

	if got.Mandatory.KmsMasterKeyId != keyID {
		t.Errorf("AC-10: mandatory.kmsMasterKeyId = %q, want %q", got.Mandatory.KmsMasterKeyId, keyID)
	}
}

// TestMergeSQSCascade_AC11 — globalSQSConfig.defaults.namingTemplate="{namespace}-{name}"
// at level 7 propagates to effCfg.defaults.namingTemplate.
// KropathConfig.sqs has no namingTemplate field -- SQSConfig-only.
func TestMergeSQSCascade_AC11(t *testing.T) {
	got := mergeSQSAll(
		zeroKropathSQS,
		zeroKropathSQS,
		zeroSQSCfg,
		zeroSQSCfg,
		zeroSQSCfg,
		cascade.SQSConfigSection{NamingTemplate: "{namespace}-{name}"}, // level 7
		zeroKropathSQS,
		zeroKropathSQS,
	)

	if got.Defaults.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("AC-11: defaults.namingTemplate = %q, want {namespace}-{name}", got.Defaults.NamingTemplate)
	}
}

// TestMergeSQSCascade_AC12 — SQSConfig.mandatory.syncedLabels={data-class: internal}
// propagates to effCfg.mandatory.syncedLabels.
func TestMergeSQSCascade_AC12(t *testing.T) {
	got := mergeSQSAll(
		zeroKropathSQS,
		zeroKropathSQS,
		cascade.SQSConfigSection{SyncedLabels: map[string]string{"data-class": "internal"}}, // level 3
		zeroSQSCfg,
		zeroSQSCfg,
		zeroSQSCfg,
		zeroKropathSQS,
		zeroKropathSQS,
	)

	if got.Mandatory.SyncedLabels["data-class"] != "internal" {
		t.Errorf("AC-12: mandatory.syncedLabels[data-class] = %q, want internal", got.Mandatory.SyncedLabels["data-class"])
	}
}

// TestMergeSQSCascade_AllAbsent — when all sources are zero, effectiveConfig is all-zero
// (permissive; no governance enforced).
func TestMergeSQSCascade_AllAbsent(t *testing.T) {
	got := mergeSQSAll(
		zeroKropathSQS, zeroKropathSQS,
		zeroSQSCfg, zeroSQSCfg, zeroSQSCfg, zeroSQSCfg,
		zeroKropathSQS, zeroKropathSQS,
	)

	if got.Mandatory.EncryptionType != "" {
		t.Errorf("all-absent: mandatory.encryptionType = %q, want empty", got.Mandatory.EncryptionType)
	}
	if got.Mandatory.KmsMasterKeyId != "" {
		t.Errorf("all-absent: mandatory.kmsMasterKeyId = %q, want empty", got.Mandatory.KmsMasterKeyId)
	}
	if got.Mandatory.VisibilityTimeout != 0 {
		t.Errorf("all-absent: mandatory.visibilityTimeout = %d, want 0", got.Mandatory.VisibilityTimeout)
	}
	if got.Mandatory.MessageRetentionPeriod != 0 {
		t.Errorf("all-absent: mandatory.messageRetentionPeriod = %d, want 0", got.Mandatory.MessageRetentionPeriod)
	}
	if got.Mandatory.DelaySeconds != 0 {
		t.Errorf("all-absent: mandatory.delaySeconds = %d, want 0", got.Mandatory.DelaySeconds)
	}
	if got.Mandatory.MaximumMessageSize != 0 {
		t.Errorf("all-absent: mandatory.maximumMessageSize = %d, want 0", got.Mandatory.MaximumMessageSize)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("all-absent: mandatory.namingTemplate = %q, want empty", got.Mandatory.NamingTemplate)
	}
	if len(got.Mandatory.Tags) != 0 {
		t.Errorf("all-absent: mandatory.tags = %v, want empty", got.Mandatory.Tags)
	}
	if len(got.Defaults.Tags) != 0 {
		t.Errorf("all-absent: defaults.tags = %v, want empty", got.Defaults.Tags)
	}
}

// TestMergeSQSCascade_MandatoryPriorityOrder — verifies mandatory priority order
// for encryptionType (level 1 > 2 > 3 > 4).
func TestMergeSQSCascade_MandatoryPriorityOrder(t *testing.T) {
	cases := []struct {
		name                   string
		globalKropathMandatory cascade.SQSKropathSection
		localKropathMandatory  cascade.SQSKropathSection
		globalSQSCfgMandatory  cascade.SQSConfigSection
		localSQSCfgMandatory   cascade.SQSConfigSection
		wantEncryptionType     string
	}{
		{
			name:                   "level1-wins",
			globalKropathMandatory: cascade.SQSKropathSection{EncryptionType: "level1"},
			localKropathMandatory:  cascade.SQSKropathSection{EncryptionType: "level2"},
			globalSQSCfgMandatory:  cascade.SQSConfigSection{EncryptionType: "level3"},
			localSQSCfgMandatory:   cascade.SQSConfigSection{EncryptionType: "level4"},
			wantEncryptionType:     "level1",
		},
		{
			name:                   "level2-wins-when-1-absent",
			globalKropathMandatory: zeroKropathSQS,
			localKropathMandatory:  cascade.SQSKropathSection{EncryptionType: "level2"},
			globalSQSCfgMandatory:  cascade.SQSConfigSection{EncryptionType: "level3"},
			localSQSCfgMandatory:   cascade.SQSConfigSection{EncryptionType: "level4"},
			wantEncryptionType:     "level2",
		},
		{
			name:                   "level3-wins-when-1-2-absent",
			globalKropathMandatory: zeroKropathSQS,
			localKropathMandatory:  zeroKropathSQS,
			globalSQSCfgMandatory:  cascade.SQSConfigSection{EncryptionType: "level3"},
			localSQSCfgMandatory:   cascade.SQSConfigSection{EncryptionType: "level4"},
			wantEncryptionType:     "level3",
		},
		{
			name:                   "level4-wins-when-1-2-3-absent",
			globalKropathMandatory: zeroKropathSQS,
			localKropathMandatory:  zeroKropathSQS,
			globalSQSCfgMandatory:  zeroSQSCfg,
			localSQSCfgMandatory:   cascade.SQSConfigSection{EncryptionType: "level4"},
			wantEncryptionType:     "level4",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeSQSAll(
				tc.globalKropathMandatory,
				tc.localKropathMandatory,
				tc.globalSQSCfgMandatory,
				tc.localSQSCfgMandatory,
				zeroSQSCfg, zeroSQSCfg,
				zeroKropathSQS, zeroKropathSQS,
			)
			if got.Mandatory.EncryptionType != tc.wantEncryptionType {
				t.Errorf("mandatory.encryptionType = %q, want %q", got.Mandatory.EncryptionType, tc.wantEncryptionType)
			}
		})
	}
}

// TestMergeSQSCascade_DefaultsPriorityOrder — verifies defaults priority order
// for encryptionType (level 6 > 7 > 8 > 9).
func TestMergeSQSCascade_DefaultsPriorityOrder(t *testing.T) {
	cases := []struct {
		name                  string
		localSQSCfgDefaults   cascade.SQSConfigSection
		globalSQSCfgDefaults  cascade.SQSConfigSection
		localKropathDefaults  cascade.SQSKropathSection
		globalKropathDefaults cascade.SQSKropathSection
		wantEncryptionType    string
	}{
		{
			name:                  "level6-wins",
			localSQSCfgDefaults:   cascade.SQSConfigSection{EncryptionType: "level6"},
			globalSQSCfgDefaults:  cascade.SQSConfigSection{EncryptionType: "level7"},
			localKropathDefaults:  cascade.SQSKropathSection{EncryptionType: "level8"},
			globalKropathDefaults: cascade.SQSKropathSection{EncryptionType: "level9"},
			wantEncryptionType:    "level6",
		},
		{
			name:                  "level7-wins-when-6-absent",
			localSQSCfgDefaults:   zeroSQSCfg,
			globalSQSCfgDefaults:  cascade.SQSConfigSection{EncryptionType: "level7"},
			localKropathDefaults:  cascade.SQSKropathSection{EncryptionType: "level8"},
			globalKropathDefaults: cascade.SQSKropathSection{EncryptionType: "level9"},
			wantEncryptionType:    "level7",
		},
		{
			name:                  "level8-wins-when-6-7-absent",
			localSQSCfgDefaults:   zeroSQSCfg,
			globalSQSCfgDefaults:  zeroSQSCfg,
			localKropathDefaults:  cascade.SQSKropathSection{EncryptionType: "level8"},
			globalKropathDefaults: cascade.SQSKropathSection{EncryptionType: "level9"},
			wantEncryptionType:    "level8",
		},
		{
			name:                  "level9-fallback",
			localSQSCfgDefaults:   zeroSQSCfg,
			globalSQSCfgDefaults:  zeroSQSCfg,
			localKropathDefaults:  zeroKropathSQS,
			globalKropathDefaults: cascade.SQSKropathSection{EncryptionType: "level9"},
			wantEncryptionType:    "level9",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeSQSAll(
				zeroKropathSQS, zeroKropathSQS,
				zeroSQSCfg, zeroSQSCfg,
				tc.localSQSCfgDefaults,
				tc.globalSQSCfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
			)
			if got.Defaults.EncryptionType != tc.wantEncryptionType {
				t.Errorf("defaults.encryptionType = %q, want %q", got.Defaults.EncryptionType, tc.wantEncryptionType)
			}
		})
	}
}

// TestMergeSQSCascade_MandatoryIsolatedFromDefaults — mandatory fields must not bleed
// into defaults and vice versa.
func TestMergeSQSCascade_MandatoryIsolatedFromDefaults(t *testing.T) {
	got := mergeSQSAll(
		cascade.SQSKropathSection{EncryptionType: "kms", VisibilityTimeout: 120},
		zeroKropathSQS,
		cascade.SQSConfigSection{DelaySeconds: 30},
		zeroSQSCfg,
		cascade.SQSConfigSection{EncryptionType: "sqs-managed", MaximumMessageSize: 65536}, // defaults level 6
		zeroSQSCfg,
		zeroKropathSQS,
		zeroKropathSQS,
	)

	if got.Mandatory.EncryptionType != "kms" {
		t.Errorf("mandatory.encryptionType = %q, want kms", got.Mandatory.EncryptionType)
	}
	if got.Mandatory.VisibilityTimeout != 120 {
		t.Errorf("mandatory.visibilityTimeout = %d, want 120", got.Mandatory.VisibilityTimeout)
	}
	if got.Mandatory.DelaySeconds != 30 {
		t.Errorf("mandatory.delaySeconds = %d, want 30", got.Mandatory.DelaySeconds)
	}
	if got.Defaults.EncryptionType != "sqs-managed" {
		t.Errorf("defaults.encryptionType = %q, want sqs-managed", got.Defaults.EncryptionType)
	}
	if got.Defaults.MaximumMessageSize != 65536 {
		t.Errorf("defaults.maximumMessageSize = %d, want 65536", got.Defaults.MaximumMessageSize)
	}
	if got.Defaults.VisibilityTimeout != 0 {
		t.Errorf("defaults.visibilityTimeout = %d, must not bleed from mandatory", got.Defaults.VisibilityTimeout)
	}
}

// TestMergeSQSCascade_TagUnionMerge — tags from multiple tiers are union-merged;
// higher-priority (lower-level number) source wins on key conflict.
func TestMergeSQSCascade_TagUnionMerge(t *testing.T) {
	got := mergeSQSAll(
		cascade.SQSKropathSection{Tags: map[string]string{"owner": "platform", "cost-centre": "infra"}},   // level 1 (wins)
		zeroKropathSQS,
		cascade.SQSConfigSection{Tags: map[string]string{"cost-centre": "payments", "env": "prod"}}, // level 3
		zeroSQSCfg,
		zeroSQSCfg,
		zeroSQSCfg,
		zeroKropathSQS,
		zeroKropathSQS,
	)

	if got.Mandatory.Tags["cost-centre"] != "infra" {
		t.Errorf("tag union: mandatory.tags[cost-centre] = %q, want infra (level 1 wins)", got.Mandatory.Tags["cost-centre"])
	}
	if got.Mandatory.Tags["owner"] != "platform" {
		t.Errorf("tag union: mandatory.tags[owner] = %q, want platform", got.Mandatory.Tags["owner"])
	}
	if got.Mandatory.Tags["env"] != "prod" {
		t.Errorf("tag union: mandatory.tags[env] = %q, want prod (additive)", got.Mandatory.Tags["env"])
	}
}

// TestMergeSQSCascade_SyncedLabelsUnionMerge — syncedLabels from SQSConfig levels only
// are union-merged; level 3 wins over level 4 on key conflict.
func TestMergeSQSCascade_SyncedLabelsUnionMerge(t *testing.T) {
	got := mergeSQSAll(
		zeroKropathSQS,
		zeroKropathSQS,
		cascade.SQSConfigSection{SyncedLabels: map[string]string{"team": "platform", "data-class": "public"}},   // level 3 (wins)
		cascade.SQSConfigSection{SyncedLabels: map[string]string{"data-class": "internal", "region": "ap"}}, // level 4
		zeroSQSCfg,
		zeroSQSCfg,
		zeroKropathSQS,
		zeroKropathSQS,
	)

	if got.Mandatory.SyncedLabels["data-class"] != "public" {
		t.Errorf("syncedLabels: mandatory.syncedLabels[data-class] = %q, want public (level 3 wins)", got.Mandatory.SyncedLabels["data-class"])
	}
	if got.Mandatory.SyncedLabels["team"] != "platform" {
		t.Errorf("syncedLabels: mandatory.syncedLabels[team] = %q, want platform", got.Mandatory.SyncedLabels["team"])
	}
	if got.Mandatory.SyncedLabels["region"] != "ap" {
		t.Errorf("syncedLabels: mandatory.syncedLabels[region] = %q, want ap (additive)", got.Mandatory.SyncedLabels["region"])
	}
}

// TestMergeSQSCascade_NamingTemplateLevel3WinsLevel4 — global SQSConfig.mandatory.namingTemplate
// (level 3) wins over local SQSConfig.mandatory.namingTemplate (level 4).
func TestMergeSQSCascade_NamingTemplateLevel3WinsLevel4(t *testing.T) {
	got := mergeSQSAll(
		zeroKropathSQS,
		zeroKropathSQS,
		cascade.SQSConfigSection{NamingTemplate: "{namespace}-{name}-global"}, // level 3
		cascade.SQSConfigSection{NamingTemplate: "{namespace}-{name}-local"},  // level 4
		zeroSQSCfg,
		zeroSQSCfg,
		zeroKropathSQS,
		zeroKropathSQS,
	)

	if got.Mandatory.NamingTemplate != "{namespace}-{name}-global" {
		t.Errorf("namingTemplate: mandatory.namingTemplate = %q, want {namespace}-{name}-global (level 3 wins)",
			got.Mandatory.NamingTemplate)
	}
}
