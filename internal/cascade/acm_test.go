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

// zeroACMKropath is a zero-value ACMKropathSection (absent source).
var zeroACMKropath = cascade.ACMKropathSection{}

// zeroACMCfg is a zero-value ACMConfigSection (absent source).
var zeroACMCfg = cascade.ACMConfigSection{}

// mergeACMAll calls MergeACMCascade with all eight inputs.
func mergeACMAll(
	globalKropathMandatory, localKropathMandatory cascade.ACMKropathSection,
	globalACMCfgMandatory, localACMCfgMandatory cascade.ACMConfigSection,
	localACMCfgDefaults, globalACMCfgDefaults cascade.ACMConfigSection,
	localKropathDefaults, globalKropathDefaults cascade.ACMKropathSection,
) cascade.EffectiveACMConfig {
	return cascade.MergeACMCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalACMCfgMandatory,
		localACMCfgMandatory,
		localACMCfgDefaults,
		globalACMCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeACMCascade_AC1_KropathMandatoryKeyAlgorithmLevel1 — globalKropathConfig.mandatory.
// certificateManager.keyAlgorithm (level 1) propagates to effCfg.mandatory.keyAlgorithm.
func TestMergeACMCascade_AC1_KropathMandatoryKeyAlgorithmLevel1(t *testing.T) {
	got := mergeACMAll(
		cascade.ACMKropathSection{KeyAlgorithm: "EC_prime256v1"}, // level 1
		zeroACMKropath,
		zeroACMCfg,
		zeroACMCfg,
		zeroACMCfg,
		zeroACMCfg,
		zeroACMKropath,
		zeroACMKropath,
	)

	if got.Mandatory.KeyAlgorithm != "EC_prime256v1" {
		t.Errorf("AC1: mandatory.keyAlgorithm = %q, want EC_prime256v1", got.Mandatory.KeyAlgorithm)
	}
	if got.Defaults.KeyAlgorithm != "" {
		t.Errorf("AC1: defaults.keyAlgorithm must not bleed from mandatory, got %q", got.Defaults.KeyAlgorithm)
	}
}

// TestMergeACMCascade_AC2_Level1WinsOverLevel3 — globalKropathConfig (level 1) wins over
// globalACMConfig.mandatory (level 3) for keyAlgorithm.
func TestMergeACMCascade_AC2_Level1WinsOverLevel3(t *testing.T) {
	got := mergeACMAll(
		cascade.ACMKropathSection{KeyAlgorithm: "EC_prime256v1"}, // level 1
		zeroACMKropath,
		cascade.ACMConfigSection{KeyAlgorithm: "RSA_2048"}, // level 3
		zeroACMCfg,
		zeroACMCfg,
		zeroACMCfg,
		zeroACMKropath,
		zeroACMKropath,
	)

	if got.Mandatory.KeyAlgorithm != "EC_prime256v1" {
		t.Errorf("AC2: level 1 KropathConfig must win; mandatory.keyAlgorithm = %q, want EC_prime256v1", got.Mandatory.KeyAlgorithm)
	}
}

// TestMergeACMCascade_AC3_GlobalACMCfgMandatoryKeyAlgorithmLevel3 — globalACMConfig.mandatory.
// keyAlgorithm (level 3) propagates when no KropathConfig override.
func TestMergeACMCascade_AC3_GlobalACMCfgMandatoryKeyAlgorithmLevel3(t *testing.T) {
	got := mergeACMAll(
		zeroACMKropath,
		zeroACMKropath,
		cascade.ACMConfigSection{KeyAlgorithm: "EC_secp384r1"}, // level 3
		zeroACMCfg,
		zeroACMCfg,
		zeroACMCfg,
		zeroACMKropath,
		zeroACMKropath,
	)

	if got.Mandatory.KeyAlgorithm != "EC_secp384r1" {
		t.Errorf("AC3: mandatory.keyAlgorithm = %q, want EC_secp384r1", got.Mandatory.KeyAlgorithm)
	}
}

// TestMergeACMCascade_AC4_Level3WinsOverLevel4 — globalACMConfig.mandatory (level 3)
// wins over localACMConfig.mandatory (level 4) for keyAlgorithm.
func TestMergeACMCascade_AC4_Level3WinsOverLevel4(t *testing.T) {
	got := mergeACMAll(
		zeroACMKropath,
		zeroACMKropath,
		cascade.ACMConfigSection{KeyAlgorithm: "EC_prime256v1"}, // level 3
		cascade.ACMConfigSection{KeyAlgorithm: "RSA_2048"},      // level 4
		zeroACMCfg,
		zeroACMCfg,
		zeroACMKropath,
		zeroACMKropath,
	)

	if got.Mandatory.KeyAlgorithm != "EC_prime256v1" {
		t.Errorf("AC4: mandatory.keyAlgorithm = %q, want EC_prime256v1 (level 3 wins)", got.Mandatory.KeyAlgorithm)
	}
}

// TestMergeACMCascade_AC5_DefaultsKeyAlgorithmFromLevel6 — localACMConfig.defaults.
// keyAlgorithm (level 6) propagates; mandatory must be empty.
func TestMergeACMCascade_AC5_DefaultsKeyAlgorithmFromLevel6(t *testing.T) {
	got := mergeACMAll(
		zeroACMKropath,
		zeroACMKropath,
		zeroACMCfg,
		zeroACMCfg,
		cascade.ACMConfigSection{KeyAlgorithm: "RSA_2048"}, // level 6
		zeroACMCfg,
		zeroACMKropath,
		zeroACMKropath,
	)

	if got.Mandatory.KeyAlgorithm != "" {
		t.Errorf("AC5: mandatory.keyAlgorithm = %q, want empty", got.Mandatory.KeyAlgorithm)
	}
	if got.Defaults.KeyAlgorithm != "RSA_2048" {
		t.Errorf("AC5: defaults.keyAlgorithm = %q, want RSA_2048", got.Defaults.KeyAlgorithm)
	}
}

// TestMergeACMCascade_AC6_DefaultsLevel6WinsOverLevel9 — level 6 wins over level 9 for defaults.
func TestMergeACMCascade_AC6_DefaultsLevel6WinsOverLevel9(t *testing.T) {
	got := mergeACMAll(
		zeroACMKropath,
		zeroACMKropath,
		zeroACMCfg,
		zeroACMCfg,
		cascade.ACMConfigSection{KeyAlgorithm: "EC_prime256v1"}, // level 6
		zeroACMCfg,
		zeroACMKropath,
		cascade.ACMKropathSection{KeyAlgorithm: "RSA_2048"}, // level 9
	)

	if got.Defaults.KeyAlgorithm != "EC_prime256v1" {
		t.Errorf("AC6: defaults level 6 must win; keyAlgorithm = %q, want EC_prime256v1", got.Defaults.KeyAlgorithm)
	}
}

// TestMergeACMCascade_AC7_MandatoryCertificateTransparencyLogging — level 1
// certificateTransparencyLogging propagates to mandatory.
func TestMergeACMCascade_AC7_MandatoryCertificateTransparencyLogging(t *testing.T) {
	got := mergeACMAll(
		cascade.ACMKropathSection{CertificateTransparencyLogging: "ENABLED"}, // level 1
		zeroACMKropath,
		zeroACMCfg,
		zeroACMCfg,
		zeroACMCfg,
		zeroACMCfg,
		zeroACMKropath,
		zeroACMKropath,
	)

	if got.Mandatory.CertificateTransparencyLogging != "ENABLED" {
		t.Errorf("AC7: mandatory.certificateTransparencyLogging = %q, want ENABLED", got.Mandatory.CertificateTransparencyLogging)
	}
	if got.Defaults.CertificateTransparencyLogging != "" {
		t.Errorf("AC7: defaults.certificateTransparencyLogging must not bleed from mandatory")
	}
}

// TestMergeACMCascade_AC8_ACMConfigOnlyUsageMode — usageMode only from ACMConfig levels 3/4.
func TestMergeACMCascade_AC8_ACMConfigOnlyUsageMode(t *testing.T) {
	got := mergeACMAll(
		zeroACMKropath,
		zeroACMKropath,
		cascade.ACMConfigSection{UsageMode: "GENERAL_PURPOSE"}, // level 3
		zeroACMCfg,
		zeroACMCfg,
		zeroACMCfg,
		zeroACMKropath,
		zeroACMKropath,
	)

	if got.Mandatory.UsageMode != "GENERAL_PURPOSE" {
		t.Errorf("AC8: mandatory.usageMode = %q, want GENERAL_PURPOSE", got.Mandatory.UsageMode)
	}
}

// TestMergeACMCascade_AC9_DefaultsUsageModeFromLevel6 — usageMode from defaults (level 6).
func TestMergeACMCascade_AC9_DefaultsUsageModeFromLevel6(t *testing.T) {
	got := mergeACMAll(
		zeroACMKropath,
		zeroACMKropath,
		zeroACMCfg,
		zeroACMCfg,
		cascade.ACMConfigSection{UsageMode: "SHORT_LIVED_CERTIFICATE"}, // level 6
		zeroACMCfg,
		zeroACMKropath,
		zeroACMKropath,
	)

	if got.Mandatory.UsageMode != "" {
		t.Errorf("AC9: mandatory.usageMode = %q, want empty", got.Mandatory.UsageMode)
	}
	if got.Defaults.UsageMode != "SHORT_LIVED_CERTIFICATE" {
		t.Errorf("AC9: defaults.usageMode = %q, want SHORT_LIVED_CERTIFICATE", got.Defaults.UsageMode)
	}
}

// TestMergeACMCascade_AC10_ACMConfigOnlyKeyStorageSecurityStandard — keyStorageSecurityStandard
// only from ACMConfig (no KropathConfig equivalent).
func TestMergeACMCascade_AC10_ACMConfigOnlyKeyStorageSecurityStandard(t *testing.T) {
	got := mergeACMAll(
		zeroACMKropath,
		zeroACMKropath,
		cascade.ACMConfigSection{KeyStorageSecurityStandard: "FIPS_140_2_LEVEL_3_OR_HIGHER"}, // level 3
		zeroACMCfg,
		zeroACMCfg,
		zeroACMCfg,
		zeroACMKropath,
		zeroACMKropath,
	)

	if got.Mandatory.KeyStorageSecurityStandard != "FIPS_140_2_LEVEL_3_OR_HIGHER" {
		t.Errorf("AC10: mandatory.keyStorageSecurityStandard = %q, want FIPS_140_2_LEVEL_3_OR_HIGHER",
			got.Mandatory.KeyStorageSecurityStandard)
	}
}

// TestMergeACMCascade_AC11_MandatoryTagUnionMerge — KropathConfig.mandatory.tags (level 1)
// and ACMConfig.mandatory.tags (level 4) union-merge; level 1 wins on key conflict.
func TestMergeACMCascade_AC11_MandatoryTagUnionMerge(t *testing.T) {
	got := mergeACMAll(
		cascade.ACMKropathSection{Tags: map[string]string{"cost-centre": "platform"}}, // level 1
		zeroACMKropath,
		zeroACMCfg,
		cascade.ACMConfigSection{Tags: map[string]string{"cert-team": "security"}}, // level 4
		zeroACMCfg,
		zeroACMCfg,
		zeroACMKropath,
		zeroACMKropath,
	)

	if got.Mandatory.Tags["cost-centre"] != "platform" {
		t.Errorf("AC11: mandatory.tags[cost-centre] = %q, want platform", got.Mandatory.Tags["cost-centre"])
	}
	if got.Mandatory.Tags["cert-team"] != "security" {
		t.Errorf("AC11: mandatory.tags[cert-team] = %q, want security", got.Mandatory.Tags["cert-team"])
	}
}

// TestMergeACMCascade_AC12_MandatoryTagKeyConflictLevel1Wins — when both level 1 and level 4
// set the same tag key, level 1 wins.
func TestMergeACMCascade_AC12_MandatoryTagKeyConflictLevel1Wins(t *testing.T) {
	got := mergeACMAll(
		cascade.ACMKropathSection{Tags: map[string]string{"env": "prod"}}, // level 1
		zeroACMKropath,
		zeroACMCfg,
		cascade.ACMConfigSection{Tags: map[string]string{"env": "staging"}}, // level 4
		zeroACMCfg,
		zeroACMCfg,
		zeroACMKropath,
		zeroACMKropath,
	)

	if got.Mandatory.Tags["env"] != "prod" {
		t.Errorf("AC12: level 1 must win on key conflict; mandatory.tags[env] = %q, want prod", got.Mandatory.Tags["env"])
	}
}

// TestMergeACMCascade_AC13_DefaultsTagUnionMerge — tags union across all defaults sources;
// level 6 wins on key conflict.
func TestMergeACMCascade_AC13_DefaultsTagUnionMerge(t *testing.T) {
	got := mergeACMAll(
		zeroACMKropath,
		zeroACMKropath,
		zeroACMCfg,
		zeroACMCfg,
		cascade.ACMConfigSection{Tags: map[string]string{"tier": "gold"}},      // level 6
		zeroACMCfg,
		zeroACMKropath,
		cascade.ACMKropathSection{Tags: map[string]string{"org": "kropath"}}, // level 9
	)

	if got.Defaults.Tags["tier"] != "gold" {
		t.Errorf("AC13: defaults.tags[tier] = %q, want gold", got.Defaults.Tags["tier"])
	}
	if got.Defaults.Tags["org"] != "kropath" {
		t.Errorf("AC13: defaults.tags[org] = %q, want kropath", got.Defaults.Tags["org"])
	}
}

// TestMergeACMCascade_AC14_MandatorySyncedLabelsFromACMConfig — syncedLabels from
// ACMConfig mandatory levels only; KropathConfig does not contribute.
func TestMergeACMCascade_AC14_MandatorySyncedLabelsFromACMConfig(t *testing.T) {
	got := mergeACMAll(
		zeroACMKropath,
		zeroACMKropath,
		cascade.ACMConfigSection{SyncedLabels: map[string]string{"compliance": "pci"}}, // level 3
		zeroACMCfg,
		zeroACMCfg,
		zeroACMCfg,
		zeroACMKropath,
		zeroACMKropath,
	)

	if got.Mandatory.SyncedLabels["compliance"] != "pci" {
		t.Errorf("AC14: mandatory.syncedLabels[compliance] = %q, want pci", got.Mandatory.SyncedLabels["compliance"])
	}
}

// TestMergeACMCascade_AC15_SyncedLabelsMandatoryLevel3WinsLevel4 — level 3 (globalACMConfig)
// wins over level 4 (localACMConfig) for syncedLabels.
func TestMergeACMCascade_AC15_SyncedLabelsMandatoryLevel3WinsLevel4(t *testing.T) {
	got := mergeACMAll(
		zeroACMKropath,
		zeroACMKropath,
		cascade.ACMConfigSection{SyncedLabels: map[string]string{"team": "security"}}, // level 3
		cascade.ACMConfigSection{SyncedLabels: map[string]string{"team": "platform"}}, // level 4
		zeroACMCfg,
		zeroACMCfg,
		zeroACMKropath,
		zeroACMKropath,
	)

	if got.Mandatory.SyncedLabels["team"] != "security" {
		t.Errorf("AC15: level 3 must win; mandatory.syncedLabels[team] = %q, want security", got.Mandatory.SyncedLabels["team"])
	}
}

// TestMergeACMCascade_AC16_DefaultsSyncedLabelsFromLevel6 — syncedLabels in defaults from
// ACMConfig level 6.
func TestMergeACMCascade_AC16_DefaultsSyncedLabelsFromLevel6(t *testing.T) {
	got := mergeACMAll(
		zeroACMKropath,
		zeroACMKropath,
		zeroACMCfg,
		zeroACMCfg,
		cascade.ACMConfigSection{SyncedLabels: map[string]string{"managed-by": "kropath"}}, // level 6
		zeroACMCfg,
		zeroACMKropath,
		zeroACMKropath,
	)

	if got.Mandatory.SyncedLabels != nil {
		t.Errorf("AC16: mandatory.syncedLabels must be nil when only defaults set, got %v", got.Mandatory.SyncedLabels)
	}
	if got.Defaults.SyncedLabels["managed-by"] != "kropath" {
		t.Errorf("AC16: defaults.syncedLabels[managed-by] = %q, want kropath", got.Defaults.SyncedLabels["managed-by"])
	}
}

// TestMergeACMCascade_AC17_MandatorySyncedAnnotations — syncedAnnotations from ACMConfig
// mandatory levels (L3 wins over L4).
func TestMergeACMCascade_AC17_MandatorySyncedAnnotations(t *testing.T) {
	got := mergeACMAll(
		zeroACMKropath,
		zeroACMKropath,
		cascade.ACMConfigSection{SyncedAnnotations: map[string]string{"cert.kropath.run/policy": "pci"}}, // level 3
		zeroACMCfg,
		zeroACMCfg,
		zeroACMCfg,
		zeroACMKropath,
		zeroACMKropath,
	)

	if got.Mandatory.SyncedAnnotations["cert.kropath.run/policy"] != "pci" {
		t.Errorf("AC17: mandatory.syncedAnnotations[cert.kropath.run/policy] = %q, want pci",
			got.Mandatory.SyncedAnnotations["cert.kropath.run/policy"])
	}
}

// TestMergeACMCascade_AC18_AllFieldsZeroYieldsEmptyResult — zero inputs produce empty output.
func TestMergeACMCascade_AC18_AllFieldsZeroYieldsEmptyResult(t *testing.T) {
	got := mergeACMAll(
		zeroACMKropath, zeroACMKropath, zeroACMCfg, zeroACMCfg,
		zeroACMCfg, zeroACMCfg, zeroACMKropath, zeroACMKropath,
	)

	if got.Mandatory.KeyAlgorithm != "" {
		t.Errorf("AC18: mandatory.keyAlgorithm = %q, want empty", got.Mandatory.KeyAlgorithm)
	}
	if got.Mandatory.UsageMode != "" {
		t.Errorf("AC18: mandatory.usageMode = %q, want empty", got.Mandatory.UsageMode)
	}
	if got.Mandatory.Tags != nil {
		t.Errorf("AC18: mandatory.tags = %v, want nil", got.Mandatory.Tags)
	}
	if got.Mandatory.SyncedLabels != nil {
		t.Errorf("AC18: mandatory.syncedLabels = %v, want nil", got.Mandatory.SyncedLabels)
	}
	if got.Defaults.KeyAlgorithm != "" {
		t.Errorf("AC18: defaults.keyAlgorithm = %q, want empty", got.Defaults.KeyAlgorithm)
	}
	if got.Defaults.Tags != nil {
		t.Errorf("AC18: defaults.tags = %v, want nil", got.Defaults.Tags)
	}
}

// TestMergeACMCascade_AC19_MandatoryDoesNotBleedIntoDefaults — a mandatory-only source
// must not affect defaults.
func TestMergeACMCascade_AC19_MandatoryDoesNotBleedIntoDefaults(t *testing.T) {
	got := mergeACMAll(
		cascade.ACMKropathSection{
			KeyAlgorithm:                   "EC_prime256v1",
			CertificateTransparencyLogging: "ENABLED",
			Tags:                           map[string]string{"k": "v"},
		},
		zeroACMKropath,
		zeroACMCfg,
		zeroACMCfg,
		zeroACMCfg,
		zeroACMCfg,
		zeroACMKropath,
		zeroACMKropath,
	)

	if got.Defaults.KeyAlgorithm != "" {
		t.Errorf("AC19: defaults.keyAlgorithm must not bleed from mandatory-only level 1")
	}
	if got.Defaults.CertificateTransparencyLogging != "" {
		t.Errorf("AC19: defaults.certificateTransparencyLogging must not bleed from mandatory-only level 1")
	}
	if got.Defaults.Tags != nil {
		t.Errorf("AC19: defaults.tags must not bleed from mandatory-only level 1")
	}
}

// TestMergeACMCascade_AC20_DefaultsACMConfigOnlyKeyStorageSecurityStandard — level 6 defaults.
func TestMergeACMCascade_AC20_DefaultsACMConfigOnlyKeyStorageSecurityStandard(t *testing.T) {
	got := mergeACMAll(
		zeroACMKropath,
		zeroACMKropath,
		zeroACMCfg,
		zeroACMCfg,
		cascade.ACMConfigSection{KeyStorageSecurityStandard: "FIPS_140_2_LEVEL_2_OR_HIGHER"}, // level 6
		zeroACMCfg,
		zeroACMKropath,
		zeroACMKropath,
	)

	if got.Mandatory.KeyStorageSecurityStandard != "" {
		t.Errorf("AC20: mandatory.keyStorageSecurityStandard = %q, want empty when only defaults set",
			got.Mandatory.KeyStorageSecurityStandard)
	}
	if got.Defaults.KeyStorageSecurityStandard != "FIPS_140_2_LEVEL_2_OR_HIGHER" {
		t.Errorf("AC20: defaults.keyStorageSecurityStandard = %q, want FIPS_140_2_LEVEL_2_OR_HIGHER",
			got.Defaults.KeyStorageSecurityStandard)
	}
}
