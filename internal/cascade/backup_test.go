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

// zeroBackupKropath is a zero-value BackupKropathSection (absent source).
var zeroBackupKropath = cascade.BackupKropathSection{}

// zeroBackupCfg is a zero-value BackupConfigSection (absent source).
var zeroBackupCfg = cascade.BackupConfigSection{}

// mergeBackupAll calls MergeBackupCascade with all eight inputs.
func mergeBackupAll(
	globalKropathMandatory,
	localKropathMandatory cascade.BackupKropathSection,
	globalBackupCfgMandatory,
	localBackupCfgMandatory,
	localBackupCfgDefaults,
	globalBackupCfgDefaults cascade.BackupConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.BackupKropathSection,
) cascade.EffectiveBackupConfig {
	return cascade.MergeBackupCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalBackupCfgMandatory,
		localBackupCfgMandatory,
		localBackupCfgDefaults,
		globalBackupCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeBackupCascade_AC12 — globalKropathConfig.mandatory.backup.encryptionKeyARN
// propagates to effCfg.mandatory.encryptionKeyARN (spec AC-12).
func TestMergeBackupCascade_AC12(t *testing.T) {
	got := mergeBackupAll(
		cascade.BackupKropathSection{EncryptionKeyARN: "arn:aws:kms:us-east-1:123456789012:key/org-key"},
		zeroBackupKropath,
		zeroBackupCfg,
		zeroBackupCfg,
		zeroBackupCfg,
		zeroBackupCfg,
		zeroBackupKropath,
		zeroBackupKropath,
	)

	const want = "arn:aws:kms:us-east-1:123456789012:key/org-key"
	if got.Mandatory.EncryptionKeyARN != want {
		t.Errorf("AC-12: mandatory.encryptionKeyARN = %q, want %q", got.Mandatory.EncryptionKeyARN, want)
	}
	if got.Defaults.EncryptionKeyARN != "" {
		t.Errorf("AC-12: defaults.encryptionKeyARN must not bleed from mandatory, got %q",
			got.Defaults.EncryptionKeyARN)
	}
}

// TestMergeBackupCascade_AC13 — BackupConfig.defaults.tags and KropathConfig.defaults.tags
// are union-merged into effCfg.defaults.tags (spec AC-13).
func TestMergeBackupCascade_AC13(t *testing.T) {
	got := mergeBackupAll(
		zeroBackupKropath,
		zeroBackupKropath,
		zeroBackupCfg,
		zeroBackupCfg,
		cascade.BackupConfigSection{Tags: map[string]string{"team": "backup"}},           // level 6 local BackupConfig defaults
		zeroBackupCfg,
		zeroBackupKropath,
		cascade.BackupKropathSection{Tags: map[string]string{"env": "prod"}},             // level 9 global KropathConfig defaults
	)

	if got.Defaults.Tags["team"] != "backup" {
		t.Errorf("AC-13: defaults.tags[team] = %q, want backup", got.Defaults.Tags["team"])
	}
	if got.Defaults.Tags["env"] != "prod" {
		t.Errorf("AC-13: defaults.tags[env] = %q, want prod", got.Defaults.Tags["env"])
	}
	if got.Mandatory.Tags != nil {
		t.Errorf("AC-13: mandatory.tags must be nil when not set, got %v", got.Mandatory.Tags)
	}
}

// TestMergeBackupCascade_AllAbsent — when all sources are zero, effectiveConfig
// fields are all zero (permissive; no governance enforced).
func TestMergeBackupCascade_AllAbsent(t *testing.T) {
	got := mergeBackupAll(
		zeroBackupKropath, zeroBackupKropath,
		zeroBackupCfg, zeroBackupCfg, zeroBackupCfg, zeroBackupCfg,
		zeroBackupKropath, zeroBackupKropath,
	)

	if got.Mandatory.EncryptionKeyARN != "" {
		t.Errorf("all-absent: mandatory.encryptionKeyARN = %q, want empty", got.Mandatory.EncryptionKeyARN)
	}
	if got.Mandatory.VaultLockMinRetentionDays != 0 {
		t.Errorf("all-absent: mandatory.vaultLockMinRetentionDays = %d, want 0", got.Mandatory.VaultLockMinRetentionDays)
	}
	if got.Mandatory.EnableContinuousBackup {
		t.Error("all-absent: mandatory.enableContinuousBackup should be false")
	}
	if got.Mandatory.EnableMalwareScan {
		t.Error("all-absent: mandatory.enableMalwareScan should be false")
	}
	if got.Defaults.EncryptionKeyARN != "" {
		t.Errorf("all-absent: defaults.encryptionKeyARN = %q, want empty", got.Defaults.EncryptionKeyARN)
	}
	if got.Defaults.DefaultLifecycleDeleteAfterDays != 0 {
		t.Errorf("all-absent: defaults.defaultLifecycleDeleteAfterDays = %d, want 0", got.Defaults.DefaultLifecycleDeleteAfterDays)
	}
}

// TestMergeBackupCascade_MandatoryCascadeOrder_EncryptionKeyARN — verifies mandatory
// priority order for encryptionKeyARN (level 1 > 2 > 3 > 4).
func TestMergeBackupCascade_MandatoryCascadeOrder_EncryptionKeyARN(t *testing.T) {
	cases := []struct {
		name                     string
		globalKropathMandatory   cascade.BackupKropathSection
		localKropathMandatory    cascade.BackupKropathSection
		globalBackupCfgMandatory cascade.BackupConfigSection
		localBackupCfgMandatory  cascade.BackupConfigSection
		wantARN                  string
	}{
		{
			name:                     "level1-wins",
			globalKropathMandatory:   cascade.BackupKropathSection{EncryptionKeyARN: "L1"},
			localKropathMandatory:    cascade.BackupKropathSection{EncryptionKeyARN: "L2"},
			globalBackupCfgMandatory: cascade.BackupConfigSection{EncryptionKeyARN: "L3"},
			localBackupCfgMandatory:  cascade.BackupConfigSection{EncryptionKeyARN: "L4"},
			wantARN:                  "L1",
		},
		{
			name:                     "level2-wins-when-1-absent",
			globalKropathMandatory:   zeroBackupKropath,
			localKropathMandatory:    cascade.BackupKropathSection{EncryptionKeyARN: "L2"},
			globalBackupCfgMandatory: cascade.BackupConfigSection{EncryptionKeyARN: "L3"},
			localBackupCfgMandatory:  cascade.BackupConfigSection{EncryptionKeyARN: "L4"},
			wantARN:                  "L2",
		},
		{
			name:                     "level3-wins-when-1-2-absent",
			globalKropathMandatory:   zeroBackupKropath,
			localKropathMandatory:    zeroBackupKropath,
			globalBackupCfgMandatory: cascade.BackupConfigSection{EncryptionKeyARN: "L3"},
			localBackupCfgMandatory:  cascade.BackupConfigSection{EncryptionKeyARN: "L4"},
			wantARN:                  "L3",
		},
		{
			name:                     "level4-wins-when-1-2-3-absent",
			globalKropathMandatory:   zeroBackupKropath,
			localKropathMandatory:    zeroBackupKropath,
			globalBackupCfgMandatory: zeroBackupCfg,
			localBackupCfgMandatory:  cascade.BackupConfigSection{EncryptionKeyARN: "L4"},
			wantARN:                  "L4",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeBackupAll(
				tc.globalKropathMandatory,
				tc.localKropathMandatory,
				tc.globalBackupCfgMandatory,
				tc.localBackupCfgMandatory,
				zeroBackupCfg, zeroBackupCfg,
				zeroBackupKropath, zeroBackupKropath,
			)
			if got.Mandatory.EncryptionKeyARN != tc.wantARN {
				t.Errorf("mandatory.encryptionKeyARN = %q, want %q", got.Mandatory.EncryptionKeyARN, tc.wantARN)
			}
		})
	}
}

// TestMergeBackupCascade_DefaultsCascadeOrder_EncryptionKeyARN — verifies defaults
// priority order for encryptionKeyARN (level 6 > 7 > 8 > 9).
func TestMergeBackupCascade_DefaultsCascadeOrder_EncryptionKeyARN(t *testing.T) {
	cases := []struct {
		name                    string
		localBackupCfgDefaults  cascade.BackupConfigSection
		globalBackupCfgDefaults cascade.BackupConfigSection
		localKropathDefaults    cascade.BackupKropathSection
		globalKropathDefaults   cascade.BackupKropathSection
		wantARN                 string
	}{
		{
			name:                    "level6-wins",
			localBackupCfgDefaults:  cascade.BackupConfigSection{EncryptionKeyARN: "L6"},
			globalBackupCfgDefaults: cascade.BackupConfigSection{EncryptionKeyARN: "L7"},
			localKropathDefaults:    cascade.BackupKropathSection{EncryptionKeyARN: "L8"},
			globalKropathDefaults:   cascade.BackupKropathSection{EncryptionKeyARN: "L9"},
			wantARN:                 "L6",
		},
		{
			name:                    "level7-wins-when-6-absent",
			localBackupCfgDefaults:  zeroBackupCfg,
			globalBackupCfgDefaults: cascade.BackupConfigSection{EncryptionKeyARN: "L7"},
			localKropathDefaults:    cascade.BackupKropathSection{EncryptionKeyARN: "L8"},
			globalKropathDefaults:   cascade.BackupKropathSection{EncryptionKeyARN: "L9"},
			wantARN:                 "L7",
		},
		{
			name:                    "level8-wins-when-6-7-absent",
			localBackupCfgDefaults:  zeroBackupCfg,
			globalBackupCfgDefaults: zeroBackupCfg,
			localKropathDefaults:    cascade.BackupKropathSection{EncryptionKeyARN: "L8"},
			globalKropathDefaults:   cascade.BackupKropathSection{EncryptionKeyARN: "L9"},
			wantARN:                 "L8",
		},
		{
			name:                    "level9-fallback",
			localBackupCfgDefaults:  zeroBackupCfg,
			globalBackupCfgDefaults: zeroBackupCfg,
			localKropathDefaults:    zeroBackupKropath,
			globalKropathDefaults:   cascade.BackupKropathSection{EncryptionKeyARN: "L9"},
			wantARN:                 "L9",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeBackupAll(
				zeroBackupKropath, zeroBackupKropath,
				zeroBackupCfg, zeroBackupCfg,
				tc.localBackupCfgDefaults,
				tc.globalBackupCfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
			)
			if got.Defaults.EncryptionKeyARN != tc.wantARN {
				t.Errorf("defaults.encryptionKeyARN = %q, want %q", got.Defaults.EncryptionKeyARN, tc.wantARN)
			}
		})
	}
}

// TestMergeBackupCascade_MandatoryIsolatedFromDefaults — mandatory fields must not
// bleed into defaults and vice versa.
func TestMergeBackupCascade_MandatoryIsolatedFromDefaults(t *testing.T) {
	got := mergeBackupAll(
		cascade.BackupKropathSection{
			EncryptionKeyARN:      "arn:mandatory",
			EnableContinuousBackup: true,
			VaultLockMinRetentionDays: 90,
		},
		zeroBackupKropath,
		zeroBackupCfg,
		zeroBackupCfg,
		cascade.BackupConfigSection{ // defaults level 6
			EncryptionKeyARN:             "arn:defaults",
			DefaultLifecycleDeleteAfterDays: 35,
		},
		zeroBackupCfg,
		zeroBackupKropath,
		zeroBackupKropath,
	)

	if got.Mandatory.EncryptionKeyARN != "arn:mandatory" {
		t.Errorf("mandatory.encryptionKeyARN = %q, want arn:mandatory", got.Mandatory.EncryptionKeyARN)
	}
	if !got.Mandatory.EnableContinuousBackup {
		t.Error("mandatory.enableContinuousBackup should be true")
	}
	if got.Mandatory.VaultLockMinRetentionDays != 90 {
		t.Errorf("mandatory.vaultLockMinRetentionDays = %d, want 90", got.Mandatory.VaultLockMinRetentionDays)
	}
	// Defaults must not bleed from mandatory
	if got.Defaults.EnableContinuousBackup {
		t.Error("defaults.enableContinuousBackup must not bleed from mandatory")
	}
	if got.Defaults.VaultLockMinRetentionDays != 0 {
		t.Errorf("defaults.vaultLockMinRetentionDays must not bleed from mandatory, got %d", got.Defaults.VaultLockMinRetentionDays)
	}
	// Defaults level 6 values must be present
	if got.Defaults.EncryptionKeyARN != "arn:defaults" {
		t.Errorf("defaults.encryptionKeyARN = %q, want arn:defaults", got.Defaults.EncryptionKeyARN)
	}
	if got.Defaults.DefaultLifecycleDeleteAfterDays != 35 {
		t.Errorf("defaults.defaultLifecycleDeleteAfterDays = %d, want 35", got.Defaults.DefaultLifecycleDeleteAfterDays)
	}
}

// TestMergeBackupCascade_IntegerFieldsZeroIsNotSet — int64 fields: 0 = not enforced;
// a non-zero value at a higher level wins.
func TestMergeBackupCascade_IntegerFieldsZeroIsNotSet(t *testing.T) {
	got := mergeBackupAll(
		zeroBackupKropath,
		zeroBackupKropath,
		cascade.BackupConfigSection{VaultLockMinRetentionDays: 90},  // level 3
		cascade.BackupConfigSection{VaultLockMinRetentionDays: 30},  // level 4
		zeroBackupCfg,
		zeroBackupCfg,
		zeroBackupKropath,
		zeroBackupKropath,
	)

	if got.Mandatory.VaultLockMinRetentionDays != 90 {
		t.Errorf("mandatory.vaultLockMinRetentionDays = %d, want 90 (level 3 wins over level 4)", got.Mandatory.VaultLockMinRetentionDays)
	}
}

// TestMergeBackupCascade_BooleanFieldsCascade — bool fields: false = not enforced,
// true at a higher level wins.
func TestMergeBackupCascade_BooleanFieldsCascade(t *testing.T) {
	got := mergeBackupAll(
		zeroBackupKropath,
		zeroBackupKropath,
		zeroBackupCfg,
		cascade.BackupConfigSection{EnableMalwareScan: true}, // level 4 mandatory
		cascade.BackupConfigSection{EnableContinuousBackup: true}, // level 6 defaults
		zeroBackupCfg,
		zeroBackupKropath,
		zeroBackupKropath,
	)

	if !got.Mandatory.EnableMalwareScan {
		t.Error("mandatory.enableMalwareScan should be true from level 4")
	}
	if got.Mandatory.EnableContinuousBackup {
		t.Error("mandatory.enableContinuousBackup must not bleed from defaults")
	}
	if !got.Defaults.EnableContinuousBackup {
		t.Error("defaults.enableContinuousBackup should be true from level 6")
	}
	if got.Defaults.EnableMalwareScan {
		t.Error("defaults.enableMalwareScan must not bleed from mandatory")
	}
}

// TestMergeBackupCascade_IAMRoleARN_NotInKropathConfig — iamRoleARN is not in
// KropathConfig; it only appears at levels 3/4 (mandatory) and 6/7 (defaults).
func TestMergeBackupCascade_IAMRoleARN_NotInKropathConfig(t *testing.T) {
	const globalRole = "arn:aws:iam::123456789012:role/GlobalBackup"
	const localRole = "arn:aws:iam::123456789012:role/LocalBackup"

	// Level 3 (global BackupConfig mandatory) provides the role.
	got := mergeBackupAll(
		cascade.BackupKropathSection{EncryptionKeyARN: "arn:key"}, // level 1 — no IAMRoleARN
		zeroBackupKropath,
		cascade.BackupConfigSection{IAMRoleARN: globalRole}, // level 3
		cascade.BackupConfigSection{IAMRoleARN: localRole},  // level 4
		zeroBackupCfg,
		zeroBackupCfg,
		zeroBackupKropath,
		zeroBackupKropath,
	)

	if got.Mandatory.IAMRoleARN != globalRole {
		t.Errorf("mandatory.iamRoleARN = %q, want %q (level 3 wins over level 4)", got.Mandatory.IAMRoleARN, globalRole)
	}

	// Defaults path: level 6 and 7 only.
	const defaultRole = "arn:aws:iam::123456789012:role/DefaultBackup"
	got2 := mergeBackupAll(
		zeroBackupKropath,
		cascade.BackupKropathSection{ScannerRoleARN: "arn:scanner"}, // level 2 — no IAMRoleARN
		zeroBackupCfg,
		zeroBackupCfg,
		cascade.BackupConfigSection{IAMRoleARN: defaultRole}, // level 6
		zeroBackupCfg,
		zeroBackupKropath,
		cascade.BackupKropathSection{ScannerRoleARN: "arn:scanner"}, // level 9 — no IAMRoleARN
	)

	if got2.Defaults.IAMRoleARN != defaultRole {
		t.Errorf("defaults.iamRoleARN = %q, want %q (level 6)", got2.Defaults.IAMRoleARN, defaultRole)
	}
}

// TestMergeBackupCascade_NamingTemplate_NotInKropathConfig — namingTemplate is not in
// KropathConfig; it only appears at levels 3/4 (mandatory) and 6/7 (defaults).
func TestMergeBackupCascade_NamingTemplate_NotInKropathConfig(t *testing.T) {
	const globalTemplate = "{namespace}-global-{name}"
	const defaultTemplate = "{namespace}-{name}"

	got := mergeBackupAll(
		zeroBackupKropath,
		zeroBackupKropath,
		cascade.BackupConfigSection{NamingTemplate: globalTemplate}, // level 3
		zeroBackupCfg,
		cascade.BackupConfigSection{NamingTemplate: defaultTemplate}, // level 6
		zeroBackupCfg,
		zeroBackupKropath,
		zeroBackupKropath,
	)

	if got.Mandatory.NamingTemplate != globalTemplate {
		t.Errorf("mandatory.namingTemplate = %q, want %q", got.Mandatory.NamingTemplate, globalTemplate)
	}
	if got.Defaults.NamingTemplate != defaultTemplate {
		t.Errorf("defaults.namingTemplate = %q, want %q", got.Defaults.NamingTemplate, defaultTemplate)
	}
}

// TestMergeBackupCascade_MandatoryTagsMerge — mandatory tags are union-merged across
// all levels; level 1 wins on key conflicts.
func TestMergeBackupCascade_MandatoryTagsMerge(t *testing.T) {
	got := mergeBackupAll(
		cascade.BackupKropathSection{Tags: map[string]string{"env": "prod", "owner": "kropath"}}, // level 1
		zeroBackupKropath,
		zeroBackupCfg,
		cascade.BackupConfigSection{Tags: map[string]string{"env": "dev", "service": "backup"}}, // level 4
		zeroBackupCfg,
		zeroBackupCfg,
		zeroBackupKropath,
		zeroBackupKropath,
	)

	// Level 1 wins on "env" conflict.
	if got.Mandatory.Tags["env"] != "prod" {
		t.Errorf("mandatory.tags[env] = %q, want prod (level 1 wins)", got.Mandatory.Tags["env"])
	}
	if got.Mandatory.Tags["owner"] != "kropath" {
		t.Errorf("mandatory.tags[owner] = %q, want kropath", got.Mandatory.Tags["owner"])
	}
	if got.Mandatory.Tags["service"] != "backup" {
		t.Errorf("mandatory.tags[service] = %q, want backup", got.Mandatory.Tags["service"])
	}
}

// TestMergeBackupCascade_DefaultsTagsMerge — defaults tags are union-merged across all
// levels; level 6 wins on key conflicts.
func TestMergeBackupCascade_DefaultsTagsMerge(t *testing.T) {
	got := mergeBackupAll(
		zeroBackupKropath,
		zeroBackupKropath,
		zeroBackupCfg,
		zeroBackupCfg,
		cascade.BackupConfigSection{Tags: map[string]string{"cost-centre": "backup", "env": "prod"}}, // level 6
		zeroBackupCfg,
		zeroBackupKropath,
		cascade.BackupKropathSection{Tags: map[string]string{"env": "dev", "org": "kropath"}}, // level 9
	)

	// Level 6 wins on "env" conflict.
	if got.Defaults.Tags["env"] != "prod" {
		t.Errorf("defaults.tags[env] = %q, want prod (level 6 wins)", got.Defaults.Tags["env"])
	}
	if got.Defaults.Tags["cost-centre"] != "backup" {
		t.Errorf("defaults.tags[cost-centre] = %q, want backup", got.Defaults.Tags["cost-centre"])
	}
	if got.Defaults.Tags["org"] != "kropath" {
		t.Errorf("defaults.tags[org] = %q, want kropath", got.Defaults.Tags["org"])
	}
}
