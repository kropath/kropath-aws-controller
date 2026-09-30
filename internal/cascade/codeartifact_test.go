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

// zeroKropathCA is a zero-value CodeArtifactKropathSection (absent source).
var zeroKropathCA = cascade.CodeArtifactKropathSection{}

// zeroCACfg is a zero-value CodeArtifactConfigSection (absent source).
var zeroCACfg = cascade.CodeArtifactConfigSection{}

// mergeCAAll calls MergeCodeArtifactCascade with all eight inputs.
func mergeCAAll(
	globalKropathMandatory,
	localKropathMandatory cascade.CodeArtifactKropathSection,
	globalCACfgMandatory,
	localCACfgMandatory,
	localCACfgDefaults,
	globalCACfgDefaults cascade.CodeArtifactConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.CodeArtifactKropathSection,
) cascade.EffectiveCodeArtifactConfig {
	return cascade.MergeCodeArtifactCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalCACfgMandatory,
		localCACfgMandatory,
		localCACfgDefaults,
		globalCACfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeCodeArtifactCascade_AC4_MandatoryEncryptionKey — globalKropathConfig.mandatory.codeartifact.encryptionKey
// at level 1 propagates to effCfg.mandatory.encryptionKey (AC-4: org-wide KMS enforcement).
func TestMergeCodeArtifactCascade_AC4_MandatoryEncryptionKey(t *testing.T) {
	keyARN := "arn:aws:kms:us-east-1:123456789012:key/org-key"
	got := mergeCAAll(
		cascade.CodeArtifactKropathSection{EncryptionKey: keyARN}, // level 1
		zeroKropathCA,
		zeroCACfg,
		zeroCACfg,
		zeroCACfg,
		zeroCACfg,
		zeroKropathCA,
		zeroKropathCA,
	)

	if got.Mandatory.EncryptionKey != keyARN {
		t.Errorf("AC-4: mandatory.encryptionKey = %q, want %q (level 1 wins)", got.Mandatory.EncryptionKey, keyARN)
	}
	if got.Defaults.EncryptionKey != "" {
		t.Errorf("AC-4: defaults.encryptionKey = %q, must not bleed from mandatory", got.Defaults.EncryptionKey)
	}
}

// TestMergeCodeArtifactCascade_AC4_MandatoryEncryptionKeyFromConfig — globalCodeArtifactConfig.mandatory.encryptionKey
// at level 3 wins when KropathConfig levels are empty (AC-4: per-config KMS enforcement).
func TestMergeCodeArtifactCascade_AC4_MandatoryEncryptionKeyFromConfig(t *testing.T) {
	keyARN := "arn:aws:kms:us-east-1:123456789012:key/cfg-key"
	got := mergeCAAll(
		zeroKropathCA,
		zeroKropathCA,
		cascade.CodeArtifactConfigSection{EncryptionKey: keyARN}, // level 3
		zeroCACfg,
		zeroCACfg,
		zeroCACfg,
		zeroKropathCA,
		zeroKropathCA,
	)

	if got.Mandatory.EncryptionKey != keyARN {
		t.Errorf("AC-4: mandatory.encryptionKey = %q, want %q (level 3 wins when 1-2 empty)", got.Mandatory.EncryptionKey, keyARN)
	}
}

// TestMergeCodeArtifactCascade_AC4_DefaultsEncryptionKey — localCodeArtifactConfig.defaults.encryptionKey
// at level 6 propagates; mandatory stays empty.
func TestMergeCodeArtifactCascade_AC4_DefaultsEncryptionKey(t *testing.T) {
	keyARN := "arn:aws:kms:us-east-1:123456789012:key/default-key"
	got := mergeCAAll(
		zeroKropathCA,
		zeroKropathCA,
		zeroCACfg,
		zeroCACfg,
		cascade.CodeArtifactConfigSection{EncryptionKey: keyARN}, // level 6
		zeroCACfg,
		zeroKropathCA,
		zeroKropathCA,
	)

	if got.Mandatory.EncryptionKey != "" {
		t.Errorf("AC-4: mandatory.encryptionKey = %q, want empty", got.Mandatory.EncryptionKey)
	}
	if got.Defaults.EncryptionKey != keyARN {
		t.Errorf("AC-4: defaults.encryptionKey = %q, want %q (level 6)", got.Defaults.EncryptionKey, keyARN)
	}
}

// TestMergeCodeArtifactCascade_NamingTemplateFromConfig — globalCodeArtifactConfig.defaults.namingTemplate
// at level 7 propagates. KropathConfig.codeartifact has no namingTemplate field.
func TestMergeCodeArtifactCascade_NamingTemplateFromConfig(t *testing.T) {
	tmpl := "{namespace}-{name}"
	got := mergeCAAll(
		zeroKropathCA,
		zeroKropathCA,
		zeroCACfg,
		zeroCACfg,
		zeroCACfg,
		cascade.CodeArtifactConfigSection{NamingTemplate: tmpl}, // level 7
		zeroKropathCA,
		zeroKropathCA,
	)

	if got.Defaults.NamingTemplate != tmpl {
		t.Errorf("namingTemplate: defaults.namingTemplate = %q, want %q (level 7)", got.Defaults.NamingTemplate, tmpl)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("namingTemplate: mandatory.namingTemplate = %q, must be empty", got.Mandatory.NamingTemplate)
	}
}

// TestMergeCodeArtifactCascade_NamingTemplateMandatory — globalCodeArtifactConfig.mandatory.namingTemplate
// at level 3 propagates to effCfg.mandatory.namingTemplate.
func TestMergeCodeArtifactCascade_NamingTemplateMandatory(t *testing.T) {
	tmpl := "corp-{namespace}-{name}"
	got := mergeCAAll(
		zeroKropathCA,
		zeroKropathCA,
		cascade.CodeArtifactConfigSection{NamingTemplate: tmpl}, // level 3
		zeroCACfg,
		zeroCACfg,
		zeroCACfg,
		zeroKropathCA,
		zeroKropathCA,
	)

	if got.Mandatory.NamingTemplate != tmpl {
		t.Errorf("namingTemplate-mandatory: mandatory.namingTemplate = %q, want %q (level 3)", got.Mandatory.NamingTemplate, tmpl)
	}
}

// TestMergeCodeArtifactCascade_AC7_TagsUnion — KropathConfig.mandatory.tags and
// CodeArtifactConfig.mandatory.tags are union-merged into effCfg.mandatory.tags (AC-7).
func TestMergeCodeArtifactCascade_AC7_TagsUnion(t *testing.T) {
	got := mergeCAAll(
		cascade.CodeArtifactKropathSection{Tags: map[string]string{"cost-centre": "platform"}},            // level 1
		zeroKropathCA,
		zeroCACfg,
		cascade.CodeArtifactConfigSection{Tags: map[string]string{"artifact-class": "production"}}, // level 4
		zeroCACfg,
		zeroCACfg,
		zeroKropathCA,
		zeroKropathCA,
	)

	if got.Mandatory.Tags["cost-centre"] != "platform" {
		t.Errorf("AC-7: mandatory.tags[cost-centre] = %q, want platform", got.Mandatory.Tags["cost-centre"])
	}
	if got.Mandatory.Tags["artifact-class"] != "production" {
		t.Errorf("AC-7: mandatory.tags[artifact-class] = %q, want production", got.Mandatory.Tags["artifact-class"])
	}
}

// TestMergeCodeArtifactCascade_AC7_TagsDefaultsUnion — KropathConfig.defaults.tags and
// CodeArtifactConfig.defaults.tags are union-merged into effCfg.defaults.tags.
func TestMergeCodeArtifactCascade_AC7_TagsDefaultsUnion(t *testing.T) {
	got := mergeCAAll(
		zeroKropathCA,
		zeroKropathCA,
		zeroCACfg,
		zeroCACfg,
		cascade.CodeArtifactConfigSection{Tags: map[string]string{"team": "backend"}},   // level 6
		zeroCACfg,
		cascade.CodeArtifactKropathSection{Tags: map[string]string{"env": "staging"}}, // level 8
		zeroKropathCA,
	)

	if got.Defaults.Tags["team"] != "backend" {
		t.Errorf("AC-7: defaults.tags[team] = %q, want backend", got.Defaults.Tags["team"])
	}
	if got.Defaults.Tags["env"] != "staging" {
		t.Errorf("AC-7: defaults.tags[env] = %q, want staging", got.Defaults.Tags["env"])
	}
}

// TestMergeCodeArtifactCascade_AC8_SyncedLabels — CodeArtifactConfig.mandatory.syncedLabels
// at level 3 propagates to effCfg.mandatory.syncedLabels (AC-8).
func TestMergeCodeArtifactCascade_AC8_SyncedLabels(t *testing.T) {
	got := mergeCAAll(
		zeroKropathCA,
		zeroKropathCA,
		cascade.CodeArtifactConfigSection{SyncedLabels: map[string]string{"data-class": "internal"}}, // level 3
		zeroCACfg,
		zeroCACfg,
		zeroCACfg,
		zeroKropathCA,
		zeroKropathCA,
	)

	if got.Mandatory.SyncedLabels["data-class"] != "internal" {
		t.Errorf("AC-8: mandatory.syncedLabels[data-class] = %q, want internal", got.Mandatory.SyncedLabels["data-class"])
	}
}

// TestMergeCodeArtifactCascade_AllAbsent — when all sources are zero, effectiveConfig is
// all-zero (permissive; no governance enforced).
func TestMergeCodeArtifactCascade_AllAbsent(t *testing.T) {
	got := mergeCAAll(
		zeroKropathCA, zeroKropathCA,
		zeroCACfg, zeroCACfg, zeroCACfg, zeroCACfg,
		zeroKropathCA, zeroKropathCA,
	)

	if got.Mandatory.EncryptionKey != "" {
		t.Errorf("all-absent: mandatory.encryptionKey = %q, want empty", got.Mandatory.EncryptionKey)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("all-absent: mandatory.namingTemplate = %q, want empty", got.Mandatory.NamingTemplate)
	}
	if len(got.Mandatory.Tags) != 0 {
		t.Errorf("all-absent: mandatory.tags = %v, want empty", got.Mandatory.Tags)
	}
	if got.Defaults.EncryptionKey != "" {
		t.Errorf("all-absent: defaults.encryptionKey = %q, want empty", got.Defaults.EncryptionKey)
	}
	if len(got.Defaults.Tags) != 0 {
		t.Errorf("all-absent: defaults.tags = %v, want empty", got.Defaults.Tags)
	}
}

// TestMergeCodeArtifactCascade_MandatoryPriorityOrder — verifies mandatory priority order
// for encryptionKey: level 1 > 2 > 3 > 4.
func TestMergeCodeArtifactCascade_MandatoryPriorityOrder(t *testing.T) {
	cases := []struct {
		name                   string
		globalKropathMandatory cascade.CodeArtifactKropathSection
		localKropathMandatory  cascade.CodeArtifactKropathSection
		globalCACfgMandatory   cascade.CodeArtifactConfigSection
		localCACfgMandatory    cascade.CodeArtifactConfigSection
		wantEncryptionKey      string
	}{
		{
			name:                   "level1-wins",
			globalKropathMandatory: cascade.CodeArtifactKropathSection{EncryptionKey: "level1"},
			localKropathMandatory:  cascade.CodeArtifactKropathSection{EncryptionKey: "level2"},
			globalCACfgMandatory:   cascade.CodeArtifactConfigSection{EncryptionKey: "level3"},
			localCACfgMandatory:    cascade.CodeArtifactConfigSection{EncryptionKey: "level4"},
			wantEncryptionKey:      "level1",
		},
		{
			name:                   "level2-wins-when-1-absent",
			globalKropathMandatory: zeroKropathCA,
			localKropathMandatory:  cascade.CodeArtifactKropathSection{EncryptionKey: "level2"},
			globalCACfgMandatory:   cascade.CodeArtifactConfigSection{EncryptionKey: "level3"},
			localCACfgMandatory:    cascade.CodeArtifactConfigSection{EncryptionKey: "level4"},
			wantEncryptionKey:      "level2",
		},
		{
			name:                   "level3-wins-when-1-2-absent",
			globalKropathMandatory: zeroKropathCA,
			localKropathMandatory:  zeroKropathCA,
			globalCACfgMandatory:   cascade.CodeArtifactConfigSection{EncryptionKey: "level3"},
			localCACfgMandatory:    cascade.CodeArtifactConfigSection{EncryptionKey: "level4"},
			wantEncryptionKey:      "level3",
		},
		{
			name:                   "level4-wins-when-1-2-3-absent",
			globalKropathMandatory: zeroKropathCA,
			localKropathMandatory:  zeroKropathCA,
			globalCACfgMandatory:   zeroCACfg,
			localCACfgMandatory:    cascade.CodeArtifactConfigSection{EncryptionKey: "level4"},
			wantEncryptionKey:      "level4",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeCAAll(
				tc.globalKropathMandatory,
				tc.localKropathMandatory,
				tc.globalCACfgMandatory,
				tc.localCACfgMandatory,
				zeroCACfg,
				zeroCACfg,
				zeroKropathCA,
				zeroKropathCA,
			)
			if got.Mandatory.EncryptionKey != tc.wantEncryptionKey {
				t.Errorf("mandatory.encryptionKey = %q, want %q", got.Mandatory.EncryptionKey, tc.wantEncryptionKey)
			}
		})
	}
}

// TestMergeCodeArtifactCascade_DefaultsPriorityOrder — verifies defaults priority order
// for encryptionKey: level 6 > 7 > 8 > 9.
func TestMergeCodeArtifactCascade_DefaultsPriorityOrder(t *testing.T) {
	cases := []struct {
		name                  string
		localCACfgDefaults    cascade.CodeArtifactConfigSection
		globalCACfgDefaults   cascade.CodeArtifactConfigSection
		localKropathDefaults  cascade.CodeArtifactKropathSection
		globalKropathDefaults cascade.CodeArtifactKropathSection
		wantEncryptionKey     string
	}{
		{
			name:                  "level6-wins",
			localCACfgDefaults:    cascade.CodeArtifactConfigSection{EncryptionKey: "level6"},
			globalCACfgDefaults:   cascade.CodeArtifactConfigSection{EncryptionKey: "level7"},
			localKropathDefaults:  cascade.CodeArtifactKropathSection{EncryptionKey: "level8"},
			globalKropathDefaults: cascade.CodeArtifactKropathSection{EncryptionKey: "level9"},
			wantEncryptionKey:     "level6",
		},
		{
			name:                  "level7-wins-when-6-absent",
			localCACfgDefaults:    zeroCACfg,
			globalCACfgDefaults:   cascade.CodeArtifactConfigSection{EncryptionKey: "level7"},
			localKropathDefaults:  cascade.CodeArtifactKropathSection{EncryptionKey: "level8"},
			globalKropathDefaults: cascade.CodeArtifactKropathSection{EncryptionKey: "level9"},
			wantEncryptionKey:     "level7",
		},
		{
			name:                  "level8-wins-when-6-7-absent",
			localCACfgDefaults:    zeroCACfg,
			globalCACfgDefaults:   zeroCACfg,
			localKropathDefaults:  cascade.CodeArtifactKropathSection{EncryptionKey: "level8"},
			globalKropathDefaults: cascade.CodeArtifactKropathSection{EncryptionKey: "level9"},
			wantEncryptionKey:     "level8",
		},
		{
			name:                  "level9-wins-when-6-7-8-absent",
			localCACfgDefaults:    zeroCACfg,
			globalCACfgDefaults:   zeroCACfg,
			localKropathDefaults:  zeroKropathCA,
			globalKropathDefaults: cascade.CodeArtifactKropathSection{EncryptionKey: "level9"},
			wantEncryptionKey:     "level9",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeCAAll(
				zeroKropathCA,
				zeroKropathCA,
				zeroCACfg,
				zeroCACfg,
				tc.localCACfgDefaults,
				tc.globalCACfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
			)
			if got.Defaults.EncryptionKey != tc.wantEncryptionKey {
				t.Errorf("defaults.encryptionKey = %q, want %q", got.Defaults.EncryptionKey, tc.wantEncryptionKey)
			}
		})
	}
}

// TestMergeCodeArtifactCascade_TagsKeyConflict — on key conflict, lower level number wins.
// Level 1 (globalKropathMandatory) wins over level 4 (localCACfgMandatory).
func TestMergeCodeArtifactCascade_TagsKeyConflict(t *testing.T) {
	got := mergeCAAll(
		cascade.CodeArtifactKropathSection{Tags: map[string]string{"env": "org-level"}},        // level 1
		zeroKropathCA,
		zeroCACfg,
		cascade.CodeArtifactConfigSection{Tags: map[string]string{"env": "config-level"}}, // level 4
		zeroCACfg,
		zeroCACfg,
		zeroKropathCA,
		zeroKropathCA,
	)

	if got.Mandatory.Tags["env"] != "org-level" {
		t.Errorf("tag-conflict: mandatory.tags[env] = %q, want org-level (level 1 wins over level 4)", got.Mandatory.Tags["env"])
	}
}

// TestMergeCodeArtifactCascade_SyncedLabelsUnion — SyncedLabels from global (L3) and
// local (L4) CodeArtifactConfig mandatory tiers are union-merged; L3 wins on key conflict.
func TestMergeCodeArtifactCascade_SyncedLabelsUnion(t *testing.T) {
	got := mergeCAAll(
		zeroKropathCA,
		zeroKropathCA,
		cascade.CodeArtifactConfigSection{SyncedLabels: map[string]string{"tier": "global", "data-class": "internal"}}, // level 3
		cascade.CodeArtifactConfigSection{SyncedLabels: map[string]string{"tier": "local", "env": "prod"}},             // level 4
		zeroCACfg,
		zeroCACfg,
		zeroKropathCA,
		zeroKropathCA,
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
