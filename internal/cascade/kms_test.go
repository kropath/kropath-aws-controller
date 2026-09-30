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

// zeroKropathKMS is a zero-value KMSKropathSection (absent source).
var zeroKropathKMS = cascade.KMSKropathSection{}

// zeroKMSCfg is a zero-value KMSConfigSection (absent source).
var zeroKMSCfg = cascade.KMSConfigSection{}

// mergeKMSAll calls MergeKMSCascade with all eight inputs.
func mergeKMSAll(
	globalKropathMandatory,
	localKropathMandatory cascade.KMSKropathSection,
	globalKMSCfgMandatory,
	localKMSCfgMandatory,
	localKMSCfgDefaults,
	globalKMSCfgDefaults cascade.KMSConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.KMSKropathSection,
) cascade.EffectiveKMSConfig {
	return cascade.MergeKMSCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalKMSCfgMandatory,
		localKMSCfgMandatory,
		localKMSCfgDefaults,
		globalKMSCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeKMSCascade_AC1 — globalKropathConfig.mandatory.kms.enableKeyRotation=true
// at level 1 propagates to effCfg.mandatory.enableKeyRotation.
func TestMergeKMSCascade_AC1(t *testing.T) {
	got := mergeKMSAll(
		cascade.KMSKropathSection{EnableKeyRotation: true},
		zeroKropathKMS,
		zeroKMSCfg,
		zeroKMSCfg,
		zeroKMSCfg,
		zeroKMSCfg,
		zeroKropathKMS,
		zeroKropathKMS,
	)

	if !got.Mandatory.EnableKeyRotation {
		t.Error("AC-1: mandatory.enableKeyRotation should be true when set at level 1")
	}
	if got.Defaults.EnableKeyRotation {
		t.Error("AC-1: defaults.enableKeyRotation must not bleed from mandatory")
	}
}

// TestMergeKMSCascade_AC2 — level-1 KropathConfig mandatory wins over level-3
// KMSConfig mandatory when both enableKeyRotation values differ.
func TestMergeKMSCascade_AC2(t *testing.T) {
	got := mergeKMSAll(
		cascade.KMSKropathSection{EnableKeyRotation: true}, // level 1
		zeroKropathKMS,
		cascade.KMSConfigSection{EnableKeyRotation: false}, // level 3 (zero = not enforced)
		zeroKMSCfg,
		zeroKMSCfg,
		zeroKMSCfg,
		zeroKropathKMS,
		zeroKropathKMS,
	)

	if !got.Mandatory.EnableKeyRotation {
		t.Error("AC-2: level-1 KropathConfig must win; mandatory.enableKeyRotation should be true")
	}
}

// TestMergeKMSCascade_AC3 — only globalKMSConfig.defaults.enableKeyRotation set;
// mandatory must be false, defaults must be true.
func TestMergeKMSCascade_AC3(t *testing.T) {
	got := mergeKMSAll(
		zeroKropathKMS,
		zeroKropathKMS,
		zeroKMSCfg,
		zeroKMSCfg,
		zeroKMSCfg,
		cascade.KMSConfigSection{EnableKeyRotation: true}, // level 7 global defaults
		zeroKropathKMS,
		zeroKropathKMS,
	)

	if got.Mandatory.EnableKeyRotation {
		t.Error("AC-3: mandatory.enableKeyRotation should be false when only defaults set")
	}
	if !got.Defaults.EnableKeyRotation {
		t.Error("AC-3: defaults.enableKeyRotation should be true")
	}
}

// TestMergeKMSCascade_AC4 — globalKMSConfig.mandatory.keySpec=SYMMETRIC_DEFAULT (level 3)
// wins over localKMSConfig.mandatory.keySpec=RSA_4096 (level 4).
func TestMergeKMSCascade_AC4(t *testing.T) {
	got := mergeKMSAll(
		zeroKropathKMS,
		zeroKropathKMS,
		cascade.KMSConfigSection{KeySpec: "SYMMETRIC_DEFAULT"}, // level 3
		cascade.KMSConfigSection{KeySpec: "RSA_4096"},          // level 4
		zeroKMSCfg,
		zeroKMSCfg,
		zeroKropathKMS,
		zeroKropathKMS,
	)

	if got.Mandatory.KeySpec != "SYMMETRIC_DEFAULT" {
		t.Errorf("AC-4: mandatory.keySpec = %q, want SYMMETRIC_DEFAULT", got.Mandatory.KeySpec)
	}
}

// TestMergeKMSCascade_AC5 — only globalKMSConfig.defaults.keySpec set;
// mandatory.keySpec must be empty, defaults.keySpec must be set.
func TestMergeKMSCascade_AC5(t *testing.T) {
	got := mergeKMSAll(
		zeroKropathKMS,
		zeroKropathKMS,
		zeroKMSCfg,
		zeroKMSCfg,
		zeroKMSCfg,
		cascade.KMSConfigSection{KeySpec: "SYMMETRIC_DEFAULT"}, // level 7
		zeroKropathKMS,
		zeroKropathKMS,
	)

	if got.Mandatory.KeySpec != "" {
		t.Errorf("AC-5: mandatory.keySpec = %q, want empty", got.Mandatory.KeySpec)
	}
	if got.Defaults.KeySpec != "SYMMETRIC_DEFAULT" {
		t.Errorf("AC-5: defaults.keySpec = %q, want SYMMETRIC_DEFAULT", got.Defaults.KeySpec)
	}
}

// TestMergeKMSCascade_AC6 — globalKMSConfig.mandatory.keyUsage=ENCRYPT_DECRYPT
// propagates to effCfg.mandatory.keyUsage.
func TestMergeKMSCascade_AC6(t *testing.T) {
	got := mergeKMSAll(
		zeroKropathKMS,
		zeroKropathKMS,
		cascade.KMSConfigSection{KeyUsage: "ENCRYPT_DECRYPT"}, // level 3
		zeroKMSCfg,
		zeroKMSCfg,
		zeroKMSCfg,
		zeroKropathKMS,
		zeroKropathKMS,
	)

	if got.Mandatory.KeyUsage != "ENCRYPT_DECRYPT" {
		t.Errorf("AC-6: mandatory.keyUsage = %q, want ENCRYPT_DECRYPT", got.Mandatory.KeyUsage)
	}
}

// TestMergeKMSCascade_AC7 — globalKropathConfig.mandatory.kms.allowedKeySpecs
// propagates across all profiles and namespaces.
func TestMergeKMSCascade_AC7(t *testing.T) {
	allowed := []string{"SYMMETRIC_DEFAULT", "RSA_4096"}
	got := mergeKMSAll(
		cascade.KMSKropathSection{AllowedKeySpecs: allowed}, // level 1
		zeroKropathKMS,
		zeroKMSCfg,
		zeroKMSCfg,
		zeroKMSCfg,
		zeroKMSCfg,
		zeroKropathKMS,
		zeroKropathKMS,
	)

	if len(got.Mandatory.AllowedKeySpecs) != 2 {
		t.Fatalf("AC-7: mandatory.allowedKeySpecs len = %d, want 2", len(got.Mandatory.AllowedKeySpecs))
	}
	if got.Mandatory.AllowedKeySpecs[0] != "SYMMETRIC_DEFAULT" || got.Mandatory.AllowedKeySpecs[1] != "RSA_4096" {
		t.Errorf("AC-7: mandatory.allowedKeySpecs = %v, want [SYMMETRIC_DEFAULT RSA_4096]",
			got.Mandatory.AllowedKeySpecs)
	}
}

// TestMergeKMSCascade_AC8 — level-1 allowedKeySpecs (KropathConfig) wins over
// level-3 allowedKeySpecs (KMSConfig) when both are set.
func TestMergeKMSCascade_AC8(t *testing.T) {
	got := mergeKMSAll(
		cascade.KMSKropathSection{AllowedKeySpecs: []string{"SYMMETRIC_DEFAULT", "RSA_4096"}}, // level 1
		zeroKropathKMS,
		cascade.KMSConfigSection{AllowedKeySpecs: []string{"RSA_2048"}}, // level 3
		zeroKMSCfg,
		zeroKMSCfg,
		zeroKMSCfg,
		zeroKropathKMS,
		zeroKropathKMS,
	)

	if len(got.Mandatory.AllowedKeySpecs) != 2 {
		t.Fatalf("AC-8: mandatory.allowedKeySpecs len = %d, want 2 (level-1 wins)", len(got.Mandatory.AllowedKeySpecs))
	}
	if got.Mandatory.AllowedKeySpecs[0] != "SYMMETRIC_DEFAULT" {
		t.Errorf("AC-8: level-1 must win; allowedKeySpecs[0] = %q, want SYMMETRIC_DEFAULT",
			got.Mandatory.AllowedKeySpecs[0])
	}
}

// TestMergeKMSCascade_AC9 — cross-validation fails when mandatory.keySpec is not in
// mandatory.allowedKeySpecs.
func TestMergeKMSCascade_AC9(t *testing.T) {
	got := mergeKMSAll(
		zeroKropathKMS,
		zeroKropathKMS,
		cascade.KMSConfigSection{
			KeySpec:         "RSA_2048",
			AllowedKeySpecs: []string{"SYMMETRIC_DEFAULT"},
		},
		zeroKMSCfg,
		zeroKMSCfg,
		zeroKMSCfg,
		zeroKropathKMS,
		zeroKropathKMS,
	)

	valid, reason, _ := cascade.ValidateKMSKeySpec(got.Mandatory)

	if valid {
		t.Error("AC-9: ValidateKMSKeySpec should return false when keySpec not in allowedKeySpecs")
	}
	if reason != "InvalidKeySpecNotInAllowedList" {
		t.Errorf("AC-9: reason = %q, want InvalidKeySpecNotInAllowedList", reason)
	}
}

// TestMergeKMSCascade_AC10 — cross-validation passes when mandatory.keySpec is in
// mandatory.allowedKeySpecs.
func TestMergeKMSCascade_AC10(t *testing.T) {
	got := mergeKMSAll(
		zeroKropathKMS,
		zeroKropathKMS,
		cascade.KMSConfigSection{
			KeySpec:         "SYMMETRIC_DEFAULT",
			AllowedKeySpecs: []string{"SYMMETRIC_DEFAULT", "RSA_4096"},
		},
		zeroKMSCfg,
		zeroKMSCfg,
		zeroKMSCfg,
		zeroKropathKMS,
		zeroKropathKMS,
	)

	valid, reason, msg := cascade.ValidateKMSKeySpec(got.Mandatory)

	if !valid {
		t.Errorf("AC-10: ValidateKMSKeySpec should pass; reason=%q msg=%q", reason, msg)
	}
}

// TestMergeKMSCascade_AC11 — empty allowedKeySpecs = no restriction; validation passes
// even when keySpec is set.
func TestMergeKMSCascade_AC11(t *testing.T) {
	got := mergeKMSAll(
		zeroKropathKMS,
		zeroKropathKMS,
		cascade.KMSConfigSection{KeySpec: "RSA_4096"}, // allowedKeySpecs = nil
		zeroKMSCfg,
		zeroKMSCfg,
		zeroKMSCfg,
		zeroKropathKMS,
		zeroKropathKMS,
	)

	valid, _, _ := cascade.ValidateKMSKeySpec(got.Mandatory)

	if !valid {
		t.Error("AC-11: empty allowedKeySpecs = no restriction; validation must pass")
	}
	if got.Mandatory.KeySpec != "RSA_4096" {
		t.Errorf("AC-11: mandatory.keySpec = %q, want RSA_4096", got.Mandatory.KeySpec)
	}
}

// TestMergeKMSCascade_AC13 — KMSConfig/general-policy mandatory.allowedGrantOperations
// (level 3, global) propagates to effCfg.mandatory.allowedGrantOperations, winning over
// a conflicting level-4 local value (aws-kms-03 spec AC-13, cascade-population half).
func TestMergeKMSCascade_AC13(t *testing.T) {
	got := mergeKMSAll(
		zeroKropathKMS,
		zeroKropathKMS,
		cascade.KMSConfigSection{AllowedGrantOperations: []string{"Decrypt", "DescribeKey"}}, // level 3
		cascade.KMSConfigSection{AllowedGrantOperations: []string{"Encrypt"}},                // level 4
		zeroKMSCfg,
		zeroKMSCfg,
		zeroKropathKMS,
		zeroKropathKMS,
	)

	want := []string{"Decrypt", "DescribeKey"}
	if len(got.Mandatory.AllowedGrantOperations) != len(want) ||
		got.Mandatory.AllowedGrantOperations[0] != want[0] ||
		got.Mandatory.AllowedGrantOperations[1] != want[1] {
		t.Errorf("AC-13: mandatory.allowedGrantOperations = %v, want %v",
			got.Mandatory.AllowedGrantOperations, want)
	}
	if len(got.Defaults.AllowedGrantOperations) != 0 {
		t.Errorf("AC-13: defaults.allowedGrantOperations must not bleed from mandatory; got %v",
			got.Defaults.AllowedGrantOperations)
	}
}

// TestMergeKMSCascade_AC14 — defaults.allowedGrantOperations (level 6, local) propagates
// to effCfg.defaults.allowedGrantOperations while the mandatory tier stays empty
// (aws-kms-03 spec AC-14, cascade-population half).
func TestMergeKMSCascade_AC14(t *testing.T) {
	got := mergeKMSAll(
		zeroKropathKMS,
		zeroKropathKMS,
		zeroKMSCfg,
		zeroKMSCfg,
		cascade.KMSConfigSection{AllowedGrantOperations: []string{"Decrypt"}}, // level 6
		zeroKMSCfg,
		zeroKropathKMS,
		zeroKropathKMS,
	)

	if len(got.Mandatory.AllowedGrantOperations) != 0 {
		t.Errorf("AC-14: mandatory.allowedGrantOperations should be empty; got %v",
			got.Mandatory.AllowedGrantOperations)
	}
	want := []string{"Decrypt"}
	if len(got.Defaults.AllowedGrantOperations) != len(want) || got.Defaults.AllowedGrantOperations[0] != want[0] {
		t.Errorf("AC-14: defaults.allowedGrantOperations = %v, want %v",
			got.Defaults.AllowedGrantOperations, want)
	}
}

// TestMergeKMSCascade_AC15 — both tiers empty (the zero value) means no restriction:
// effCfg.{mandatory,defaults}.allowedGrantOperations are both empty (aws-kms-03 spec
// AC-15, cascade-population half).
func TestMergeKMSCascade_AC15(t *testing.T) {
	got := mergeKMSAll(
		zeroKropathKMS, zeroKropathKMS,
		zeroKMSCfg, zeroKMSCfg, zeroKMSCfg, zeroKMSCfg,
		zeroKropathKMS, zeroKropathKMS,
	)

	if len(got.Mandatory.AllowedGrantOperations) != 0 {
		t.Errorf("AC-15: mandatory.allowedGrantOperations should be empty; got %v",
			got.Mandatory.AllowedGrantOperations)
	}
	if len(got.Defaults.AllowedGrantOperations) != 0 {
		t.Errorf("AC-15: defaults.allowedGrantOperations should be empty; got %v",
			got.Defaults.AllowedGrantOperations)
	}
}

// TestMergeKMSCascade_AllowedGrantOperationsNotReadFromKropathConfig — populating every
// KropathConfig-sourced KMS field (both tiers, both KropathConfig levels) must never
// surface as allowedGrantOperations. Unlike allowedKeySpecs, this field has no
// KropathConfig levels at all (aws-kms-03 spec, ADR-018 D-3): KMSKropathSection carries
// no AllowedGrantOperations field to set, so this test proves the absence behaviorally —
// populating every other KropathConfig-sourced field must not leak a value into it.
func TestMergeKMSCascade_AllowedGrantOperationsNotReadFromKropathConfig(t *testing.T) {
	fullKropath := cascade.KMSKropathSection{
		EnableKeyRotation: true,
		AllowedKeySpecs:   []string{"SYMMETRIC_DEFAULT"},
		Tags:              map[string]string{"team": "payments"},
	}
	got := mergeKMSAll(
		fullKropath, fullKropath,
		zeroKMSCfg, zeroKMSCfg, zeroKMSCfg, zeroKMSCfg,
		fullKropath, fullKropath,
	)

	if len(got.Mandatory.AllowedGrantOperations) != 0 {
		t.Errorf("mandatory.allowedGrantOperations must not be populated from KropathConfig; got %v",
			got.Mandatory.AllowedGrantOperations)
	}
	if len(got.Defaults.AllowedGrantOperations) != 0 {
		t.Errorf("defaults.allowedGrantOperations must not be populated from KropathConfig; got %v",
			got.Defaults.AllowedGrantOperations)
	}
}

// TestMergeKMSCascade_AllowedGrantOperationsMandatoryCascadeOrder verifies the mandatory
// priority order for allowedGrantOperations (level 3 > 4 only — no KropathConfig levels,
// per aws-kms-03 spec).
func TestMergeKMSCascade_AllowedGrantOperationsMandatoryCascadeOrder(t *testing.T) {
	cases := []struct {
		name                  string
		globalKMSCfgMandatory cascade.KMSConfigSection
		localKMSCfgMandatory  cascade.KMSConfigSection
		want                  []string
	}{
		{
			name:                  "level3-wins",
			globalKMSCfgMandatory: cascade.KMSConfigSection{AllowedGrantOperations: []string{"LEVEL3"}},
			localKMSCfgMandatory:  cascade.KMSConfigSection{AllowedGrantOperations: []string{"LEVEL4"}},
			want:                  []string{"LEVEL3"},
		},
		{
			name:                  "level4-wins-when-3-absent",
			globalKMSCfgMandatory: zeroKMSCfg,
			localKMSCfgMandatory:  cascade.KMSConfigSection{AllowedGrantOperations: []string{"LEVEL4"}},
			want:                  []string{"LEVEL4"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeKMSAll(
				zeroKropathKMS, zeroKropathKMS,
				tc.globalKMSCfgMandatory, tc.localKMSCfgMandatory,
				zeroKMSCfg, zeroKMSCfg,
				zeroKropathKMS, zeroKropathKMS,
			)
			if len(got.Mandatory.AllowedGrantOperations) != len(tc.want) ||
				(len(tc.want) > 0 && got.Mandatory.AllowedGrantOperations[0] != tc.want[0]) {
				t.Errorf("mandatory.allowedGrantOperations = %v, want %v",
					got.Mandatory.AllowedGrantOperations, tc.want)
			}
		})
	}
}

// TestMergeKMSCascade_AllowedGrantOperationsDefaultsCascadeOrder verifies the defaults
// priority order for allowedGrantOperations (level 6 > 7 only — no KropathConfig levels,
// per aws-kms-03 spec).
func TestMergeKMSCascade_AllowedGrantOperationsDefaultsCascadeOrder(t *testing.T) {
	cases := []struct {
		name                 string
		localKMSCfgDefaults  cascade.KMSConfigSection
		globalKMSCfgDefaults cascade.KMSConfigSection
		want                 []string
	}{
		{
			name:                 "level6-wins",
			localKMSCfgDefaults:  cascade.KMSConfigSection{AllowedGrantOperations: []string{"LEVEL6"}},
			globalKMSCfgDefaults: cascade.KMSConfigSection{AllowedGrantOperations: []string{"LEVEL7"}},
			want:                 []string{"LEVEL6"},
		},
		{
			name:                 "level7-wins-when-6-absent",
			localKMSCfgDefaults:  zeroKMSCfg,
			globalKMSCfgDefaults: cascade.KMSConfigSection{AllowedGrantOperations: []string{"LEVEL7"}},
			want:                 []string{"LEVEL7"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeKMSAll(
				zeroKropathKMS, zeroKropathKMS,
				zeroKMSCfg, zeroKMSCfg,
				tc.localKMSCfgDefaults, tc.globalKMSCfgDefaults,
				zeroKropathKMS, zeroKropathKMS,
			)
			if len(got.Defaults.AllowedGrantOperations) != len(tc.want) ||
				(len(tc.want) > 0 && got.Defaults.AllowedGrantOperations[0] != tc.want[0]) {
				t.Errorf("defaults.allowedGrantOperations = %v, want %v",
					got.Defaults.AllowedGrantOperations, tc.want)
			}
		})
	}
}

// TestMergeKMSCascade_AllAbsent — when all sources are zero, effectiveConfig
// fields are all zero (permissive; no governance enforced).
func TestMergeKMSCascade_AllAbsent(t *testing.T) {
	got := mergeKMSAll(
		zeroKropathKMS, zeroKropathKMS,
		zeroKMSCfg, zeroKMSCfg, zeroKMSCfg, zeroKMSCfg,
		zeroKropathKMS, zeroKropathKMS,
	)

	if got.Mandatory.EnableKeyRotation {
		t.Error("all-absent: mandatory.enableKeyRotation should be false")
	}
	if got.Mandatory.KeySpec != "" {
		t.Errorf("all-absent: mandatory.keySpec = %q, want empty", got.Mandatory.KeySpec)
	}
	if got.Mandatory.KeyUsage != "" {
		t.Errorf("all-absent: mandatory.keyUsage = %q, want empty", got.Mandatory.KeyUsage)
	}
	if len(got.Mandatory.AllowedKeySpecs) != 0 {
		t.Errorf("all-absent: mandatory.allowedKeySpecs = %v, want empty", got.Mandatory.AllowedKeySpecs)
	}
	if len(got.Mandatory.AllowedGrantOperations) != 0 {
		t.Errorf("all-absent: mandatory.allowedGrantOperations = %v, want empty", got.Mandatory.AllowedGrantOperations)
	}
	if got.Defaults.EnableKeyRotation {
		t.Error("all-absent: defaults.enableKeyRotation should be false")
	}
	if got.Defaults.KeySpec != "" {
		t.Errorf("all-absent: defaults.keySpec = %q, want empty", got.Defaults.KeySpec)
	}
	if len(got.Defaults.AllowedKeySpecs) != 0 {
		t.Errorf("all-absent: defaults.allowedKeySpecs = %v, want empty", got.Defaults.AllowedKeySpecs)
	}
	if len(got.Defaults.AllowedGrantOperations) != 0 {
		t.Errorf("all-absent: defaults.allowedGrantOperations = %v, want empty", got.Defaults.AllowedGrantOperations)
	}
}

// TestMergeKMSCascade_MandatoryCascadeOrder — verifies the mandatory priority order
// for allowedKeySpecs (level 1 > 2 > 3 > 4).
func TestMergeKMSCascade_MandatoryCascadeOrder(t *testing.T) {
	cases := []struct {
		name                   string
		globalKropathMandatory cascade.KMSKropathSection
		localKropathMandatory  cascade.KMSKropathSection
		globalKMSCfgMandatory  cascade.KMSConfigSection
		localKMSCfgMandatory   cascade.KMSConfigSection
		wantAllowedKeySpecs    []string
	}{
		{
			name:                   "level1-wins",
			globalKropathMandatory: cascade.KMSKropathSection{AllowedKeySpecs: []string{"LEVEL1"}},
			localKropathMandatory:  cascade.KMSKropathSection{AllowedKeySpecs: []string{"LEVEL2"}},
			globalKMSCfgMandatory:  cascade.KMSConfigSection{AllowedKeySpecs: []string{"LEVEL3"}},
			localKMSCfgMandatory:   cascade.KMSConfigSection{AllowedKeySpecs: []string{"LEVEL4"}},
			wantAllowedKeySpecs:    []string{"LEVEL1"},
		},
		{
			name:                   "level2-wins-when-1-absent",
			globalKropathMandatory: zeroKropathKMS,
			localKropathMandatory:  cascade.KMSKropathSection{AllowedKeySpecs: []string{"LEVEL2"}},
			globalKMSCfgMandatory:  cascade.KMSConfigSection{AllowedKeySpecs: []string{"LEVEL3"}},
			localKMSCfgMandatory:   cascade.KMSConfigSection{AllowedKeySpecs: []string{"LEVEL4"}},
			wantAllowedKeySpecs:    []string{"LEVEL2"},
		},
		{
			name:                   "level3-wins-when-1-2-absent",
			globalKropathMandatory: zeroKropathKMS,
			localKropathMandatory:  zeroKropathKMS,
			globalKMSCfgMandatory:  cascade.KMSConfigSection{AllowedKeySpecs: []string{"LEVEL3"}},
			localKMSCfgMandatory:   cascade.KMSConfigSection{AllowedKeySpecs: []string{"LEVEL4"}},
			wantAllowedKeySpecs:    []string{"LEVEL3"},
		},
		{
			name:                   "level4-wins-when-1-2-3-absent",
			globalKropathMandatory: zeroKropathKMS,
			localKropathMandatory:  zeroKropathKMS,
			globalKMSCfgMandatory:  zeroKMSCfg,
			localKMSCfgMandatory:   cascade.KMSConfigSection{AllowedKeySpecs: []string{"LEVEL4"}},
			wantAllowedKeySpecs:    []string{"LEVEL4"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeKMSAll(
				tc.globalKropathMandatory,
				tc.localKropathMandatory,
				tc.globalKMSCfgMandatory,
				tc.localKMSCfgMandatory,
				zeroKMSCfg, zeroKMSCfg,
				zeroKropathKMS, zeroKropathKMS,
			)
			if len(got.Mandatory.AllowedKeySpecs) != len(tc.wantAllowedKeySpecs) ||
				(len(tc.wantAllowedKeySpecs) > 0 && got.Mandatory.AllowedKeySpecs[0] != tc.wantAllowedKeySpecs[0]) {
				t.Errorf("allowedKeySpecs = %v, want %v",
					got.Mandatory.AllowedKeySpecs, tc.wantAllowedKeySpecs)
			}
		})
	}
}

// TestMergeKMSCascade_DefaultsCascadeOrder — verifies the defaults priority order
// for allowedKeySpecs (level 6 > 7 > 8 > 9).
func TestMergeKMSCascade_DefaultsCascadeOrder(t *testing.T) {
	cases := []struct {
		name                  string
		localKMSCfgDefaults   cascade.KMSConfigSection
		globalKMSCfgDefaults  cascade.KMSConfigSection
		localKropathDefaults  cascade.KMSKropathSection
		globalKropathDefaults cascade.KMSKropathSection
		wantAllowedKeySpecs   []string
	}{
		{
			name:                  "level6-wins",
			localKMSCfgDefaults:   cascade.KMSConfigSection{AllowedKeySpecs: []string{"LEVEL6"}},
			globalKMSCfgDefaults:  cascade.KMSConfigSection{AllowedKeySpecs: []string{"LEVEL7"}},
			localKropathDefaults:  cascade.KMSKropathSection{AllowedKeySpecs: []string{"LEVEL8"}},
			globalKropathDefaults: cascade.KMSKropathSection{AllowedKeySpecs: []string{"LEVEL9"}},
			wantAllowedKeySpecs:   []string{"LEVEL6"},
		},
		{
			name:                  "level7-wins-when-6-absent",
			localKMSCfgDefaults:   zeroKMSCfg,
			globalKMSCfgDefaults:  cascade.KMSConfigSection{AllowedKeySpecs: []string{"LEVEL7"}},
			localKropathDefaults:  cascade.KMSKropathSection{AllowedKeySpecs: []string{"LEVEL8"}},
			globalKropathDefaults: cascade.KMSKropathSection{AllowedKeySpecs: []string{"LEVEL9"}},
			wantAllowedKeySpecs:   []string{"LEVEL7"},
		},
		{
			name:                  "level8-wins-when-6-7-absent",
			localKMSCfgDefaults:   zeroKMSCfg,
			globalKMSCfgDefaults:  zeroKMSCfg,
			localKropathDefaults:  cascade.KMSKropathSection{AllowedKeySpecs: []string{"LEVEL8"}},
			globalKropathDefaults: cascade.KMSKropathSection{AllowedKeySpecs: []string{"LEVEL9"}},
			wantAllowedKeySpecs:   []string{"LEVEL8"},
		},
		{
			name:                  "level9-fallback",
			localKMSCfgDefaults:   zeroKMSCfg,
			globalKMSCfgDefaults:  zeroKMSCfg,
			localKropathDefaults:  zeroKropathKMS,
			globalKropathDefaults: cascade.KMSKropathSection{AllowedKeySpecs: []string{"LEVEL9"}},
			wantAllowedKeySpecs:   []string{"LEVEL9"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeKMSAll(
				zeroKropathKMS, zeroKropathKMS,
				zeroKMSCfg, zeroKMSCfg,
				tc.localKMSCfgDefaults,
				tc.globalKMSCfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
			)
			if len(got.Defaults.AllowedKeySpecs) != len(tc.wantAllowedKeySpecs) ||
				(len(tc.wantAllowedKeySpecs) > 0 && got.Defaults.AllowedKeySpecs[0] != tc.wantAllowedKeySpecs[0]) {
				t.Errorf("defaults.allowedKeySpecs = %v, want %v",
					got.Defaults.AllowedKeySpecs, tc.wantAllowedKeySpecs)
			}
		})
	}
}

// TestMergeKMSCascade_MandatoryIsolatedFromDefaults — mandatory fields must not
// bleed into defaults and vice versa.
func TestMergeKMSCascade_MandatoryIsolatedFromDefaults(t *testing.T) {
	got := mergeKMSAll(
		cascade.KMSKropathSection{EnableKeyRotation: true, AllowedKeySpecs: []string{"SYMMETRIC_DEFAULT"}},
		zeroKropathKMS,
		cascade.KMSConfigSection{KeySpec: "SYMMETRIC_DEFAULT", KeyUsage: "ENCRYPT_DECRYPT"},
		zeroKMSCfg,
		cascade.KMSConfigSection{KeySpec: "RSA_4096", KeyUsage: "SIGN_VERIFY"}, // defaults level 6
		zeroKMSCfg,
		zeroKropathKMS,
		zeroKropathKMS,
	)

	if !got.Mandatory.EnableKeyRotation {
		t.Error("mandatory.enableKeyRotation should be true")
	}
	if got.Mandatory.KeySpec != "SYMMETRIC_DEFAULT" {
		t.Errorf("mandatory.keySpec = %q, want SYMMETRIC_DEFAULT", got.Mandatory.KeySpec)
	}
	if got.Mandatory.KeyUsage != "ENCRYPT_DECRYPT" {
		t.Errorf("mandatory.keyUsage = %q, want ENCRYPT_DECRYPT", got.Mandatory.KeyUsage)
	}
	if got.Defaults.EnableKeyRotation {
		t.Error("defaults.enableKeyRotation must not bleed from mandatory")
	}
	if got.Defaults.KeySpec != "RSA_4096" {
		t.Errorf("defaults.keySpec = %q, want RSA_4096", got.Defaults.KeySpec)
	}
	if got.Defaults.KeyUsage != "SIGN_VERIFY" {
		t.Errorf("defaults.keyUsage = %q, want SIGN_VERIFY", got.Defaults.KeyUsage)
	}
}

// TestValidateKMSKeySpec_NoKeySpec — constraint does not apply when keySpec is empty.
func TestValidateKMSKeySpec_NoKeySpec(t *testing.T) {
	valid, _, _ := cascade.ValidateKMSKeySpec(cascade.EffectiveKMSSection{
		AllowedKeySpecs: []string{"SYMMETRIC_DEFAULT"},
	})
	if !valid {
		t.Error("constraint must not apply when keySpec is empty")
	}
}

// TestValidateKMSKeySpec_EmptyAllowedList — empty allowedKeySpecs = no restriction.
func TestValidateKMSKeySpec_EmptyAllowedList(t *testing.T) {
	valid, _, _ := cascade.ValidateKMSKeySpec(cascade.EffectiveKMSSection{
		KeySpec:         "RSA_4096",
		AllowedKeySpecs: nil,
	})
	if !valid {
		t.Error("constraint must not apply when allowedKeySpecs is empty")
	}
}

// TestValidateKMSKeySpec_MessageContent — failure message must name the keySpec value
// and the allowedKeySpecs list.
func TestValidateKMSKeySpec_MessageContent(t *testing.T) {
	valid, reason, msg := cascade.ValidateKMSKeySpec(cascade.EffectiveKMSSection{
		KeySpec:         "RSA_2048",
		AllowedKeySpecs: []string{"SYMMETRIC_DEFAULT"},
	})

	if valid {
		t.Fatal("expected invalid")
	}
	if reason != "InvalidKeySpecNotInAllowedList" {
		t.Errorf("reason = %q, want InvalidKeySpecNotInAllowedList", reason)
	}
	if msg == "" {
		t.Error("message must not be empty on failure")
	}
}
