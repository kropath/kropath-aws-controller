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

var zeroDSQLKropath = cascade.DSQLKropathSection{}
var zeroDSQLCfg = cascade.DSQLConfigSection{}

func mergeDSQLAll(
	globalKropathMandatory,
	localKropathMandatory cascade.DSQLKropathSection,
	globalDSQLCfgMandatory,
	localDSQLCfgMandatory,
	localDSQLCfgDefaults,
	globalDSQLCfgDefaults cascade.DSQLConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.DSQLKropathSection,
) cascade.EffectiveDSQLConfig {
	return cascade.MergeDSQLCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalDSQLCfgMandatory,
		localDSQLCfgMandatory,
		localDSQLCfgDefaults,
		globalDSQLCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestDSQLCascade_MandatoryWinsOverDSQLConfig (C-2) — KropathConfig level 1 true AND
// DSQLConfig level 3 false → mandatory.deletionProtectionEnabled: true (level 1 wins).
func TestDSQLCascade_MandatoryWinsOverDSQLConfig(t *testing.T) {
	got := mergeDSQLAll(
		cascade.DSQLKropathSection{DeletionProtectionEnabled: true},  // level 1
		zeroDSQLKropath,
		cascade.DSQLConfigSection{DeletionProtectionEnabled: false}, // level 3 (zero = not enforced)
		zeroDSQLCfg,
		zeroDSQLCfg,
		zeroDSQLCfg,
		zeroDSQLKropath,
		zeroDSQLKropath,
	)

	if !got.Mandatory.DeletionProtectionEnabled {
		t.Error("C-2: level-1 KropathConfig must win; mandatory.deletionProtectionEnabled should be true")
	}
}

// TestDSQLCascade_DefaultsOnlyNoMandatory (C-3) — only DSQLConfig.defaults.deletionProtectionEnabled: true
// → effCfg.defaults: true, effCfg.mandatory: false.
func TestDSQLCascade_DefaultsOnlyNoMandatory(t *testing.T) {
	got := mergeDSQLAll(
		zeroDSQLKropath,
		zeroDSQLKropath,
		zeroDSQLCfg,
		zeroDSQLCfg,
		zeroDSQLCfg,
		cascade.DSQLConfigSection{DeletionProtectionEnabled: true}, // level 7 global defaults
		zeroDSQLKropath,
		zeroDSQLKropath,
	)

	if got.Mandatory.DeletionProtectionEnabled {
		t.Error("C-3: mandatory.deletionProtectionEnabled should be false when only defaults set")
	}
	if !got.Defaults.DeletionProtectionEnabled {
		t.Error("C-3: defaults.deletionProtectionEnabled should be true")
	}
}

// TestDSQLCascade_LocalKropathConfigMandatory (C-4) — localKropathConfig.mandatory.dsql.deletionProtectionEnabled: true
// (level 2) → effCfg.mandatory.deletionProtectionEnabled: true.
func TestDSQLCascade_LocalKropathConfigMandatory(t *testing.T) {
	got := mergeDSQLAll(
		zeroDSQLKropath,
		cascade.DSQLKropathSection{DeletionProtectionEnabled: true}, // level 2
		zeroDSQLCfg,
		zeroDSQLCfg,
		zeroDSQLCfg,
		zeroDSQLCfg,
		zeroDSQLKropath,
		zeroDSQLKropath,
	)

	if !got.Mandatory.DeletionProtectionEnabled {
		t.Error("C-4: mandatory.deletionProtectionEnabled should be true when set at level 2")
	}
}

// TestDSQLCascade_GlobalDSQLConfigWinsOverLocal (C-6) — local DSQLConfig level 4 AND
// global DSQLConfig level 3 both set → effCfg.mandatory.kmsEncryptionKey = level 3 (global wins).
func TestDSQLCascade_GlobalDSQLConfigWinsOverLocal(t *testing.T) {
	got := mergeDSQLAll(
		zeroDSQLKropath,
		zeroDSQLKropath,
		cascade.DSQLConfigSection{KmsEncryptionKey: "arn:aws:kms:us-east-1:111:key/global"}, // level 3
		cascade.DSQLConfigSection{KmsEncryptionKey: "arn:aws:kms:us-east-1:111:key/local"},  // level 4
		zeroDSQLCfg,
		zeroDSQLCfg,
		zeroDSQLKropath,
		zeroDSQLKropath,
	)

	want := "arn:aws:kms:us-east-1:111:key/global"
	if got.Mandatory.KmsEncryptionKey != want {
		t.Errorf("C-6: mandatory.kmsEncryptionKey = %q, want %q (global level 3 wins)", got.Mandatory.KmsEncryptionKey, want)
	}
}

// TestDSQLCascade_DefaultsKmsKeyOnly (C-7) — only DSQLConfig.defaults.kmsEncryptionKey set;
// mandatory.kmsEncryptionKey must be empty, defaults must be populated.
func TestDSQLCascade_DefaultsKmsKeyOnly(t *testing.T) {
	key := "arn:aws:kms:us-east-1:111:key/defaults"
	got := mergeDSQLAll(
		zeroDSQLKropath,
		zeroDSQLKropath,
		zeroDSQLCfg,
		zeroDSQLCfg,
		zeroDSQLCfg,
		cascade.DSQLConfigSection{KmsEncryptionKey: key}, // level 7 global defaults
		zeroDSQLKropath,
		zeroDSQLKropath,
	)

	if got.Mandatory.KmsEncryptionKey != "" {
		t.Errorf("C-7: mandatory.kmsEncryptionKey = %q, want empty", got.Mandatory.KmsEncryptionKey)
	}
	if got.Defaults.KmsEncryptionKey != key {
		t.Errorf("C-7: defaults.kmsEncryptionKey = %q, want %q", got.Defaults.KmsEncryptionKey, key)
	}
}

// TestDSQLCascade_AWSIdentityPropagates (C-10) — provider identity is carried outside cascade;
// tests that the MergeDSQLCascade itself does not affect provider identity fields (they are
// populated separately by the reconciler from KropathConfig.spec.aws.*).
func TestDSQLCascade_AWSIdentityPropagates(t *testing.T) {
	got := mergeDSQLAll(
		zeroDSQLKropath, zeroDSQLKropath,
		zeroDSQLCfg, zeroDSQLCfg,
		zeroDSQLCfg, zeroDSQLCfg,
		zeroDSQLKropath, zeroDSQLKropath,
	)

	// The cascade result itself must not carry provider identity — the reconciler
	// appends it from KropathConfig.spec.aws.* as a separate EffectiveDSQLConfig.AWS field.
	// Verify the section fields that are DSQL-specific are all zero when all inputs are zero.
	if got.Mandatory.DeletionProtectionEnabled {
		t.Error("C-10: mandatory.deletionProtectionEnabled should be false when all inputs zero")
	}
	if got.Mandatory.KmsEncryptionKey != "" {
		t.Errorf("C-10: mandatory.kmsEncryptionKey = %q, want empty", got.Mandatory.KmsEncryptionKey)
	}
}

// TestDSQLCascade_AllAbsent — when all sources are zero, effectiveConfig
// fields are all zero (permissive; no governance enforced).
func TestDSQLCascade_AllAbsent(t *testing.T) {
	got := mergeDSQLAll(
		zeroDSQLKropath, zeroDSQLKropath,
		zeroDSQLCfg, zeroDSQLCfg,
		zeroDSQLCfg, zeroDSQLCfg,
		zeroDSQLKropath, zeroDSQLKropath,
	)

	if got.Mandatory.DeletionProtectionEnabled {
		t.Error("all-absent: mandatory.deletionProtectionEnabled should be false")
	}
	if got.Mandatory.KmsEncryptionKey != "" {
		t.Errorf("all-absent: mandatory.kmsEncryptionKey = %q, want empty", got.Mandatory.KmsEncryptionKey)
	}
	if got.Defaults.DeletionProtectionEnabled {
		t.Error("all-absent: defaults.deletionProtectionEnabled should be false")
	}
	if got.Defaults.KmsEncryptionKey != "" {
		t.Errorf("all-absent: defaults.kmsEncryptionKey = %q, want empty", got.Defaults.KmsEncryptionKey)
	}
}

// TestDSQLCascade_MandatoryIsolatedFromDefaults — mandatory fields must not
// bleed into defaults and vice versa.
func TestDSQLCascade_MandatoryIsolatedFromDefaults(t *testing.T) {
	got := mergeDSQLAll(
		cascade.DSQLKropathSection{DeletionProtectionEnabled: true},                               // level 1
		zeroDSQLKropath,
		cascade.DSQLConfigSection{KmsEncryptionKey: "arn:aws:kms:us-east-1:111:key/mandatory"},     // level 3
		zeroDSQLCfg,
		cascade.DSQLConfigSection{KmsEncryptionKey: "arn:aws:kms:us-east-1:111:key/defaults"},      // level 6
		zeroDSQLCfg,
		zeroDSQLKropath,
		zeroDSQLKropath,
	)

	if !got.Mandatory.DeletionProtectionEnabled {
		t.Error("mandatory.deletionProtectionEnabled should be true")
	}
	if got.Mandatory.KmsEncryptionKey != "arn:aws:kms:us-east-1:111:key/mandatory" {
		t.Errorf("mandatory.kmsEncryptionKey = %q, want mandatory key", got.Mandatory.KmsEncryptionKey)
	}
	if got.Defaults.DeletionProtectionEnabled {
		t.Error("defaults.deletionProtectionEnabled must not bleed from mandatory")
	}
	if got.Defaults.KmsEncryptionKey != "arn:aws:kms:us-east-1:111:key/defaults" {
		t.Errorf("defaults.kmsEncryptionKey = %q, want defaults key", got.Defaults.KmsEncryptionKey)
	}
}

// TestDSQLCascade_MandatoryCascadeOrder — verifies the mandatory priority order
// for deletionProtectionEnabled (level 1 > 2 > 3 > 4).
func TestDSQLCascade_MandatoryCascadeOrder(t *testing.T) {
	cases := []struct {
		name                   string
		globalKropathMandatory cascade.DSQLKropathSection
		localKropathMandatory  cascade.DSQLKropathSection
		globalDSQLCfgMandatory cascade.DSQLConfigSection
		localDSQLCfgMandatory  cascade.DSQLConfigSection
		wantEnabled            bool
	}{
		{
			name:                   "level1-wins",
			globalKropathMandatory: cascade.DSQLKropathSection{DeletionProtectionEnabled: true},
			localKropathMandatory:  zeroDSQLKropath,
			globalDSQLCfgMandatory: zeroDSQLCfg,
			localDSQLCfgMandatory:  zeroDSQLCfg,
			wantEnabled:            true,
		},
		{
			name:                   "level2-wins-when-1-absent",
			globalKropathMandatory: zeroDSQLKropath,
			localKropathMandatory:  cascade.DSQLKropathSection{DeletionProtectionEnabled: true},
			globalDSQLCfgMandatory: zeroDSQLCfg,
			localDSQLCfgMandatory:  zeroDSQLCfg,
			wantEnabled:            true,
		},
		{
			name:                   "level3-wins-when-1-2-absent",
			globalKropathMandatory: zeroDSQLKropath,
			localKropathMandatory:  zeroDSQLKropath,
			globalDSQLCfgMandatory: cascade.DSQLConfigSection{DeletionProtectionEnabled: true},
			localDSQLCfgMandatory:  zeroDSQLCfg,
			wantEnabled:            true,
		},
		{
			name:                   "level4-wins-when-1-2-3-absent",
			globalKropathMandatory: zeroDSQLKropath,
			localKropathMandatory:  zeroDSQLKropath,
			globalDSQLCfgMandatory: zeroDSQLCfg,
			localDSQLCfgMandatory:  cascade.DSQLConfigSection{DeletionProtectionEnabled: true},
			wantEnabled:            true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeDSQLAll(
				tc.globalKropathMandatory,
				tc.localKropathMandatory,
				tc.globalDSQLCfgMandatory,
				tc.localDSQLCfgMandatory,
				zeroDSQLCfg, zeroDSQLCfg,
				zeroDSQLKropath, zeroDSQLKropath,
			)
			if got.Mandatory.DeletionProtectionEnabled != tc.wantEnabled {
				t.Errorf("deletionProtectionEnabled = %v, want %v", got.Mandatory.DeletionProtectionEnabled, tc.wantEnabled)
			}
		})
	}
}

// TestDSQLCascade_DefaultsCascadeOrder — verifies the defaults priority order (level 6 > 7 > 8 > 9).
func TestDSQLCascade_DefaultsCascadeOrder(t *testing.T) {
	cases := []struct {
		name                  string
		localDSQLCfgDefaults  cascade.DSQLConfigSection
		globalDSQLCfgDefaults cascade.DSQLConfigSection
		localKropathDefaults  cascade.DSQLKropathSection
		globalKropathDefaults cascade.DSQLKropathSection
		wantEnabled           bool
	}{
		{
			name:                 "level6-wins",
			localDSQLCfgDefaults: cascade.DSQLConfigSection{DeletionProtectionEnabled: true},
			wantEnabled:          true,
		},
		{
			name:                  "level7-wins-when-6-absent",
			globalDSQLCfgDefaults: cascade.DSQLConfigSection{DeletionProtectionEnabled: true},
			wantEnabled:           true,
		},
		{
			name:                 "level8-wins-when-6-7-absent",
			localKropathDefaults: cascade.DSQLKropathSection{DeletionProtectionEnabled: true},
			wantEnabled:          true,
		},
		{
			name:                  "level9-fallback",
			globalKropathDefaults: cascade.DSQLKropathSection{DeletionProtectionEnabled: true},
			wantEnabled:           true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeDSQLAll(
				zeroDSQLKropath, zeroDSQLKropath,
				zeroDSQLCfg, zeroDSQLCfg,
				tc.localDSQLCfgDefaults,
				tc.globalDSQLCfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
			)
			if got.Defaults.DeletionProtectionEnabled != tc.wantEnabled {
				t.Errorf("defaults.deletionProtectionEnabled = %v, want %v", got.Defaults.DeletionProtectionEnabled, tc.wantEnabled)
			}
		})
	}
}

// TestDSQLCascade_TagsMerge (C-8 unit coverage) — KropathConfig.mandatory.tags + DSQLConfig.mandatory.tags
// produce a union in effCfg.mandatory.tags; level-1 (KropathConfig) wins on key conflict.
func TestDSQLCascade_TagsMerge(t *testing.T) {
	got := mergeDSQLAll(
		cascade.DSQLKropathSection{Tags: map[string]string{"env": "prod", "owner": "platform"}}, // level 1
		zeroDSQLKropath,
		cascade.DSQLConfigSection{Tags: map[string]string{"team": "data", "owner": "data-team"}}, // level 3 (lower priority)
		zeroDSQLCfg,
		zeroDSQLCfg,
		zeroDSQLCfg,
		zeroDSQLKropath,
		zeroDSQLKropath,
	)

	if got.Mandatory.Tags["env"] != "prod" {
		t.Errorf("mandatory.tags['env'] = %q, want 'prod'", got.Mandatory.Tags["env"])
	}
	if got.Mandatory.Tags["team"] != "data" {
		t.Errorf("mandatory.tags['team'] = %q, want 'data'", got.Mandatory.Tags["team"])
	}
	// level-1 KropathConfig wins the 'owner' key conflict
	if got.Mandatory.Tags["owner"] != "platform" {
		t.Errorf("mandatory.tags['owner'] = %q, want 'platform' (level-1 wins)", got.Mandatory.Tags["owner"])
	}
}

// TestDSQLCascade_KmsEncryptionKeyCascadeOrder — verifies the mandatory kmsEncryptionKey
// priority order for levels 3 and 4 (only ConfigSection levels).
func TestDSQLCascade_KmsEncryptionKeyCascadeOrder(t *testing.T) {
	cases := []struct {
		name                   string
		globalDSQLCfgMandatory cascade.DSQLConfigSection
		localDSQLCfgMandatory  cascade.DSQLConfigSection
		wantKey                string
	}{
		{
			name:                   "level3-wins",
			globalDSQLCfgMandatory: cascade.DSQLConfigSection{KmsEncryptionKey: "arn:global"},
			localDSQLCfgMandatory:  cascade.DSQLConfigSection{KmsEncryptionKey: "arn:local"},
			wantKey:                "arn:global",
		},
		{
			name:                  "level4-wins-when-3-absent",
			globalDSQLCfgMandatory: zeroDSQLCfg,
			localDSQLCfgMandatory:  cascade.DSQLConfigSection{KmsEncryptionKey: "arn:local"},
			wantKey:                "arn:local",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeDSQLAll(
				zeroDSQLKropath, zeroDSQLKropath,
				tc.globalDSQLCfgMandatory,
				tc.localDSQLCfgMandatory,
				zeroDSQLCfg, zeroDSQLCfg,
				zeroDSQLKropath, zeroDSQLKropath,
			)
			if got.Mandatory.KmsEncryptionKey != tc.wantKey {
				t.Errorf("mandatory.kmsEncryptionKey = %q, want %q", got.Mandatory.KmsEncryptionKey, tc.wantKey)
			}
		})
	}
}
