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

// zeroAthenaKropath is a zero-value AthenaKropathSection (absent source).
var zeroAthenaKropath = cascade.AthenaKropathSection{}

// zeroAthenaCfg is a zero-value AthenaConfigSection (absent source).
var zeroAthenaCfg = cascade.AthenaConfigSection{}

// mergeAthenaAll calls MergeAthenaCascade with all eight inputs.
func mergeAthenaAll(
	globalKropathMandatory,
	localKropathMandatory cascade.AthenaKropathSection,
	globalAthenaCfgMandatory,
	localAthenaCfgMandatory,
	localAthenaCfgDefaults,
	globalAthenaCfgDefaults cascade.AthenaConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.AthenaKropathSection,
) cascade.EffectiveAthenaConfig {
	return cascade.MergeAthenaCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalAthenaCfgMandatory,
		localAthenaCfgMandatory,
		localAthenaCfgDefaults,
		globalAthenaCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeAthenaCascade_KropathConfigMandatory_Override verifies that
// globalKropathConfig.mandatory.athena.enforceWorkGroupConfiguration=true (level 1)
// propagates to effCfg.mandatory.enforceWorkGroupConfiguration.
func TestMergeAthenaCascade_KropathConfigMandatory_Override(t *testing.T) {
	got := mergeAthenaAll(
		cascade.AthenaKropathSection{EnforceWorkGroupConfiguration: true}, // level 1
		zeroAthenaKropath,
		zeroAthenaCfg,
		zeroAthenaCfg,
		zeroAthenaCfg,
		zeroAthenaCfg,
		zeroAthenaKropath,
		zeroAthenaKropath,
	)

	if !got.Mandatory.EnforceWorkGroupConfiguration {
		t.Errorf("mandatory.enforceWorkGroupConfiguration = false, want true (level 1 wins)")
	}
	if got.Defaults.EnforceWorkGroupConfiguration {
		t.Errorf("defaults.enforceWorkGroupConfiguration = true, must not bleed from mandatory")
	}
}

// TestMergeAthenaCascade_KropathConfigDefaults_Passthrough verifies that
// globalKropathConfig.defaults.athena.enforceWorkGroupConfiguration=true (level 9)
// propagates to effCfg.defaults.enforceWorkGroupConfiguration when no higher priority is set.
func TestMergeAthenaCascade_KropathConfigDefaults_Passthrough(t *testing.T) {
	got := mergeAthenaAll(
		zeroAthenaKropath,
		zeroAthenaKropath,
		zeroAthenaCfg,
		zeroAthenaCfg,
		zeroAthenaCfg,
		zeroAthenaCfg,
		zeroAthenaKropath,
		cascade.AthenaKropathSection{EnforceWorkGroupConfiguration: true}, // level 9
	)

	if !got.Defaults.EnforceWorkGroupConfiguration {
		t.Errorf("defaults.enforceWorkGroupConfiguration = false, want true (level 9 passthrough)")
	}
	if got.Mandatory.EnforceWorkGroupConfiguration {
		t.Errorf("mandatory.enforceWorkGroupConfiguration = true, must not bleed from defaults")
	}
}

// TestMergeAthenaCascade_AthenaConfigMandatory_Only verifies that
// AthenaConfig.mandatory-only fields (no KropathConfig equivalent) flow from level 3.
func TestMergeAthenaCascade_AthenaConfigMandatory_Only(t *testing.T) {
	got := mergeAthenaAll(
		zeroAthenaKropath,
		zeroAthenaKropath,
		cascade.AthenaConfigSection{
			BytesScannedCutoffPerQuery: 10737418240, // level 3
			ResultEncryptionOption:     "SSE_KMS",
			ResultOutputLocation:       "s3://my-bucket/output",
			EngineVersion:              "Athena engine version 3",
			RequesterPaysEnabled:       true,
		},
		zeroAthenaCfg,
		zeroAthenaCfg,
		zeroAthenaCfg,
		zeroAthenaKropath,
		zeroAthenaKropath,
	)

	if got.Mandatory.BytesScannedCutoffPerQuery != 10737418240 {
		t.Errorf("mandatory.bytesScannedCutoffPerQuery = %d, want 10737418240", got.Mandatory.BytesScannedCutoffPerQuery)
	}
	if got.Mandatory.ResultEncryptionOption != "SSE_KMS" {
		t.Errorf("mandatory.resultEncryptionOption = %q, want SSE_KMS", got.Mandatory.ResultEncryptionOption)
	}
	if got.Mandatory.ResultOutputLocation != "s3://my-bucket/output" {
		t.Errorf("mandatory.resultOutputLocation = %q, want s3://my-bucket/output", got.Mandatory.ResultOutputLocation)
	}
	if got.Mandatory.EngineVersion != "Athena engine version 3" {
		t.Errorf("mandatory.engineVersion = %q, want Athena engine version 3", got.Mandatory.EngineVersion)
	}
	if !got.Mandatory.RequesterPaysEnabled {
		t.Errorf("mandatory.requesterPaysEnabled = false, want true (level 3)")
	}
	// These fields must not bleed into defaults.
	if got.Defaults.BytesScannedCutoffPerQuery != 0 {
		t.Errorf("defaults.bytesScannedCutoffPerQuery = %d, must not bleed from mandatory", got.Defaults.BytesScannedCutoffPerQuery)
	}
}

// TestMergeAthenaCascade_AthenaConfigDefaults_Only verifies that
// AthenaConfig.defaults-only fields flow from level 6 (localAthenaCfgDefaults).
func TestMergeAthenaCascade_AthenaConfigDefaults_Only(t *testing.T) {
	got := mergeAthenaAll(
		zeroAthenaKropath,
		zeroAthenaKropath,
		zeroAthenaCfg,
		zeroAthenaCfg,
		cascade.AthenaConfigSection{
			NamingTemplate: "{namespace}-{name}", // level 6
			EngineVersion:  "Athena engine version 3",
		},
		zeroAthenaCfg,
		zeroAthenaKropath,
		zeroAthenaKropath,
	)

	if got.Defaults.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("defaults.namingTemplate = %q, want {namespace}-{name} (level 6)", got.Defaults.NamingTemplate)
	}
	if got.Defaults.EngineVersion != "Athena engine version 3" {
		t.Errorf("defaults.engineVersion = %q, want Athena engine version 3 (level 6)", got.Defaults.EngineVersion)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("mandatory.namingTemplate = %q, must not bleed from defaults", got.Mandatory.NamingTemplate)
	}
}

// TestMergeAthenaCascade_AllThreeKropathConfigBooleans verifies that all three
// KropathConfig boolean fields (level 1) propagate to mandatory.
func TestMergeAthenaCascade_AllThreeKropathConfigBooleans(t *testing.T) {
	got := mergeAthenaAll(
		cascade.AthenaKropathSection{
			EnforceWorkGroupConfiguration:        true, // level 1
			EnableMinimumEncryptionConfiguration: true,
			PublishCloudWatchMetricsEnabled:      true,
		},
		zeroAthenaKropath,
		zeroAthenaCfg,
		zeroAthenaCfg,
		zeroAthenaCfg,
		zeroAthenaCfg,
		zeroAthenaKropath,
		zeroAthenaKropath,
	)

	if !got.Mandatory.EnforceWorkGroupConfiguration {
		t.Errorf("mandatory.enforceWorkGroupConfiguration = false, want true")
	}
	if !got.Mandatory.EnableMinimumEncryptionConfiguration {
		t.Errorf("mandatory.enableMinimumEncryptionConfiguration = false, want true")
	}
	if !got.Mandatory.PublishCloudWatchMetricsEnabled {
		t.Errorf("mandatory.publishCloudWatchMetricsEnabled = false, want true")
	}
}

// TestMergeAthenaCascade_TagsMergedUnion verifies that tags from KropathConfig
// mandatory (level 1) and AthenaConfig mandatory (level 4) are union-merged.
// Level 1 wins on key conflict.
func TestMergeAthenaCascade_TagsMergedUnion(t *testing.T) {
	got := mergeAthenaAll(
		cascade.AthenaKropathSection{Tags: map[string]string{"cost-centre": "platform", "env": "org-level"}}, // level 1
		zeroAthenaKropath,
		zeroAthenaCfg,
		cascade.AthenaConfigSection{Tags: map[string]string{"team": "analytics", "env": "config-level"}}, // level 4
		zeroAthenaCfg,
		zeroAthenaCfg,
		zeroAthenaKropath,
		zeroAthenaKropath,
	)

	if got.Mandatory.Tags["cost-centre"] != "platform" {
		t.Errorf("mandatory.tags[cost-centre] = %q, want platform (from level 1)", got.Mandatory.Tags["cost-centre"])
	}
	if got.Mandatory.Tags["team"] != "analytics" {
		t.Errorf("mandatory.tags[team] = %q, want analytics (from level 4)", got.Mandatory.Tags["team"])
	}
	// Level 1 wins key conflict.
	if got.Mandatory.Tags["env"] != "org-level" {
		t.Errorf("mandatory.tags[env] = %q, want org-level (level 1 wins over level 4)", got.Mandatory.Tags["env"])
	}
}

// TestMergeAthenaCascade_DefaultsTagsUnion verifies union merge for defaults tags.
// Level 6 wins over level 9 on key conflict.
func TestMergeAthenaCascade_DefaultsTagsUnion(t *testing.T) {
	got := mergeAthenaAll(
		zeroAthenaKropath,
		zeroAthenaKropath,
		zeroAthenaCfg,
		zeroAthenaCfg,
		cascade.AthenaConfigSection{Tags: map[string]string{"team": "analytics", "env": "local"}}, // level 6
		zeroAthenaCfg,
		zeroAthenaKropath,
		cascade.AthenaKropathSection{Tags: map[string]string{"cost-centre": "platform", "env": "global"}}, // level 9
	)

	if got.Defaults.Tags["team"] != "analytics" {
		t.Errorf("defaults.tags[team] = %q, want analytics (level 6)", got.Defaults.Tags["team"])
	}
	if got.Defaults.Tags["cost-centre"] != "platform" {
		t.Errorf("defaults.tags[cost-centre] = %q, want platform (level 9)", got.Defaults.Tags["cost-centre"])
	}
	// Level 6 wins key conflict.
	if got.Defaults.Tags["env"] != "local" {
		t.Errorf("defaults.tags[env] = %q, want local (level 6 wins over level 9)", got.Defaults.Tags["env"])
	}
}

// TestMergeAthenaCascade_ZeroValuesNotEnforced verifies that zero-value boolean
// and integer fields from KropathConfig are treated as "not enforced" and do not
// overwrite non-zero values from lower-priority levels.
func TestMergeAthenaCascade_ZeroValuesNotEnforced(t *testing.T) {
	got := mergeAthenaAll(
		zeroAthenaKropath, // level 1: all false — not enforced
		zeroAthenaKropath, // level 2
		zeroAthenaCfg,     // level 3
		cascade.AthenaConfigSection{
			BytesScannedCutoffPerQuery: 5368709120, // level 4: non-zero wins
			EnforceWorkGroupConfiguration: true,
		},
		zeroAthenaCfg,
		zeroAthenaCfg,
		zeroAthenaKropath,
		zeroAthenaKropath,
	)

	if got.Mandatory.BytesScannedCutoffPerQuery != 5368709120 {
		t.Errorf("mandatory.bytesScannedCutoffPerQuery = %d, want 5368709120 (level 4 wins when 3 is zero)", got.Mandatory.BytesScannedCutoffPerQuery)
	}
	if !got.Mandatory.EnforceWorkGroupConfiguration {
		t.Errorf("mandatory.enforceWorkGroupConfiguration = false, want true (level 4 wins when 1-3 are zero)")
	}
}

// TestMergeAthenaCascade_AthenaConfigOnlyFields verifies AthenaConfig-only integer
// and string fields that are absent from KropathConfig.
func TestMergeAthenaCascade_AthenaConfigOnlyFields(t *testing.T) {
	got := mergeAthenaAll(
		zeroAthenaKropath,
		zeroAthenaKropath,
		cascade.AthenaConfigSection{
			BytesScannedCutoffPerQuery: 10737418240, // level 3
			ResultEncryptionOption:     "SSE_S3",
		},
		cascade.AthenaConfigSection{
			BytesScannedCutoffPerQuery: 5368709120, // level 4 — level 3 wins
			ResultEncryptionOption:     "CSE_KMS",
		},
		zeroAthenaCfg,
		zeroAthenaCfg,
		zeroAthenaKropath,
		zeroAthenaKropath,
	)

	// Level 3 wins over level 4.
	if got.Mandatory.BytesScannedCutoffPerQuery != 10737418240 {
		t.Errorf("mandatory.bytesScannedCutoffPerQuery = %d, want 10737418240 (level 3 wins over level 4)", got.Mandatory.BytesScannedCutoffPerQuery)
	}
	if got.Mandatory.ResultEncryptionOption != "SSE_S3" {
		t.Errorf("mandatory.resultEncryptionOption = %q, want SSE_S3 (level 3 wins over level 4)", got.Mandatory.ResultEncryptionOption)
	}
}

// TestMergeAthenaCascade_MandatoryPriorityOrder verifies the full mandatory priority
// order for enforceWorkGroupConfiguration: level 1 > 2 > 3 > 4.
func TestMergeAthenaCascade_MandatoryPriorityOrder(t *testing.T) {
	cases := []struct {
		name                     string
		globalKropathMandatory   cascade.AthenaKropathSection
		localKropathMandatory    cascade.AthenaKropathSection
		globalAthenaCfgMandatory cascade.AthenaConfigSection
		localAthenaCfgMandatory  cascade.AthenaConfigSection
		wantEnforce              bool
	}{
		{
			name:                     "level1-wins",
			globalKropathMandatory:   cascade.AthenaKropathSection{EnforceWorkGroupConfiguration: true},
			localKropathMandatory:    zeroAthenaKropath,
			globalAthenaCfgMandatory: zeroAthenaCfg,
			localAthenaCfgMandatory:  zeroAthenaCfg,
			wantEnforce:              true,
		},
		{
			name:                     "level2-wins-when-1-false",
			globalKropathMandatory:   zeroAthenaKropath,
			localKropathMandatory:    cascade.AthenaKropathSection{EnforceWorkGroupConfiguration: true},
			globalAthenaCfgMandatory: zeroAthenaCfg,
			localAthenaCfgMandatory:  zeroAthenaCfg,
			wantEnforce:              true,
		},
		{
			name:                     "level3-wins-when-1-2-false",
			globalKropathMandatory:   zeroAthenaKropath,
			localKropathMandatory:    zeroAthenaKropath,
			globalAthenaCfgMandatory: cascade.AthenaConfigSection{EnforceWorkGroupConfiguration: true},
			localAthenaCfgMandatory:  zeroAthenaCfg,
			wantEnforce:              true,
		},
		{
			name:                     "level4-wins-when-1-2-3-false",
			globalKropathMandatory:   zeroAthenaKropath,
			localKropathMandatory:    zeroAthenaKropath,
			globalAthenaCfgMandatory: zeroAthenaCfg,
			localAthenaCfgMandatory:  cascade.AthenaConfigSection{EnforceWorkGroupConfiguration: true},
			wantEnforce:              true,
		},
		{
			name:                     "all-false-returns-false",
			globalKropathMandatory:   zeroAthenaKropath,
			localKropathMandatory:    zeroAthenaKropath,
			globalAthenaCfgMandatory: zeroAthenaCfg,
			localAthenaCfgMandatory:  zeroAthenaCfg,
			wantEnforce:              false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeAthenaAll(
				tc.globalKropathMandatory,
				tc.localKropathMandatory,
				tc.globalAthenaCfgMandatory,
				tc.localAthenaCfgMandatory,
				zeroAthenaCfg,
				zeroAthenaCfg,
				zeroAthenaKropath,
				zeroAthenaKropath,
			)
			if got.Mandatory.EnforceWorkGroupConfiguration != tc.wantEnforce {
				t.Errorf("mandatory.enforceWorkGroupConfiguration = %v, want %v", got.Mandatory.EnforceWorkGroupConfiguration, tc.wantEnforce)
			}
		})
	}
}

// TestMergeAthenaCascade_DefaultsPriorityOrder verifies defaults priority order
// for enforceWorkGroupConfiguration: level 6 > 7 > 8 > 9.
func TestMergeAthenaCascade_DefaultsPriorityOrder(t *testing.T) {
	cases := []struct {
		name                   string
		localAthenaCfgDefaults cascade.AthenaConfigSection
		globalAthenaCfgDefaults cascade.AthenaConfigSection
		localKropathDefaults   cascade.AthenaKropathSection
		globalKropathDefaults  cascade.AthenaKropathSection
		wantEnforce            bool
	}{
		{
			name:                    "level6-wins",
			localAthenaCfgDefaults:  cascade.AthenaConfigSection{EnforceWorkGroupConfiguration: true},
			globalAthenaCfgDefaults: zeroAthenaCfg,
			localKropathDefaults:    zeroAthenaKropath,
			globalKropathDefaults:   zeroAthenaKropath,
			wantEnforce:             true,
		},
		{
			name:                    "level7-wins-when-6-absent",
			localAthenaCfgDefaults:  zeroAthenaCfg,
			globalAthenaCfgDefaults: cascade.AthenaConfigSection{EnforceWorkGroupConfiguration: true},
			localKropathDefaults:    zeroAthenaKropath,
			globalKropathDefaults:   zeroAthenaKropath,
			wantEnforce:             true,
		},
		{
			name:                    "level8-wins-when-6-7-absent",
			localAthenaCfgDefaults:  zeroAthenaCfg,
			globalAthenaCfgDefaults: zeroAthenaCfg,
			localKropathDefaults:    cascade.AthenaKropathSection{EnforceWorkGroupConfiguration: true},
			globalKropathDefaults:   zeroAthenaKropath,
			wantEnforce:             true,
		},
		{
			name:                    "level9-wins-when-6-7-8-absent",
			localAthenaCfgDefaults:  zeroAthenaCfg,
			globalAthenaCfgDefaults: zeroAthenaCfg,
			localKropathDefaults:    zeroAthenaKropath,
			globalKropathDefaults:   cascade.AthenaKropathSection{EnforceWorkGroupConfiguration: true},
			wantEnforce:             true,
		},
		{
			name:                    "all-false-returns-false",
			localAthenaCfgDefaults:  zeroAthenaCfg,
			globalAthenaCfgDefaults: zeroAthenaCfg,
			localKropathDefaults:    zeroAthenaKropath,
			globalKropathDefaults:   zeroAthenaKropath,
			wantEnforce:             false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeAthenaAll(
				zeroAthenaKropath,
				zeroAthenaKropath,
				zeroAthenaCfg,
				zeroAthenaCfg,
				tc.localAthenaCfgDefaults,
				tc.globalAthenaCfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
			)
			if got.Defaults.EnforceWorkGroupConfiguration != tc.wantEnforce {
				t.Errorf("defaults.enforceWorkGroupConfiguration = %v, want %v", got.Defaults.EnforceWorkGroupConfiguration, tc.wantEnforce)
			}
		})
	}
}

// TestMergeAthenaCascade_NamingTemplateKropathNotParticipating verifies that setting
// all KropathSection boolean fields does NOT produce a namingTemplate, since that
// field is absent from KropathConfig.athena entirely.
func TestMergeAthenaCascade_NamingTemplateKropathNotParticipating(t *testing.T) {
	got := mergeAthenaAll(
		cascade.AthenaKropathSection{
			EnforceWorkGroupConfiguration:        true,
			EnableMinimumEncryptionConfiguration: true,
			PublishCloudWatchMetricsEnabled:      true,
		},
		zeroAthenaKropath,
		zeroAthenaCfg,
		zeroAthenaCfg,
		zeroAthenaCfg,
		zeroAthenaCfg,
		zeroAthenaKropath,
		zeroAthenaKropath,
	)

	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("mandatory.namingTemplate = %q, must be empty (KropathConfig.athena has no namingTemplate)", got.Mandatory.NamingTemplate)
	}
}

// TestMergeAthenaCascade_SyncedLabelsUnion verifies that SyncedLabels from
// global (L3) and local (L4) AthenaConfig mandatory tiers are union-merged;
// L3 wins on key conflict. KropathConfig.athena does NOT carry syncedLabels.
func TestMergeAthenaCascade_SyncedLabelsUnion(t *testing.T) {
	got := mergeAthenaAll(
		zeroAthenaKropath,
		zeroAthenaKropath,
		cascade.AthenaConfigSection{SyncedLabels: map[string]string{"tier": "global", "data-class": "internal"}}, // level 3
		cascade.AthenaConfigSection{SyncedLabels: map[string]string{"tier": "local", "env": "prod"}},             // level 4
		zeroAthenaCfg,
		zeroAthenaCfg,
		zeroAthenaKropath,
		zeroAthenaKropath,
	)

	if got.Mandatory.SyncedLabels["data-class"] != "internal" {
		t.Errorf("mandatory.syncedLabels[data-class] = %q, want internal (from L3)", got.Mandatory.SyncedLabels["data-class"])
	}
	if got.Mandatory.SyncedLabels["env"] != "prod" {
		t.Errorf("mandatory.syncedLabels[env] = %q, want prod (from L4)", got.Mandatory.SyncedLabels["env"])
	}
	// L3 wins key conflict.
	if got.Mandatory.SyncedLabels["tier"] != "global" {
		t.Errorf("mandatory.syncedLabels[tier] = %q, want global (L3 wins over L4)", got.Mandatory.SyncedLabels["tier"])
	}
}

// TestMergeAthenaCascade_AllAbsent verifies that when all sources are zero-value,
// the effective config is all zero (permissive; no governance enforced).
func TestMergeAthenaCascade_AllAbsent(t *testing.T) {
	got := mergeAthenaAll(
		zeroAthenaKropath, zeroAthenaKropath,
		zeroAthenaCfg, zeroAthenaCfg, zeroAthenaCfg, zeroAthenaCfg,
		zeroAthenaKropath, zeroAthenaKropath,
	)

	if got.Mandatory.EnforceWorkGroupConfiguration {
		t.Errorf("all-absent: mandatory.enforceWorkGroupConfiguration = true, want false")
	}
	if got.Mandatory.EnableMinimumEncryptionConfiguration {
		t.Errorf("all-absent: mandatory.enableMinimumEncryptionConfiguration = true, want false")
	}
	if got.Mandatory.PublishCloudWatchMetricsEnabled {
		t.Errorf("all-absent: mandatory.publishCloudWatchMetricsEnabled = true, want false")
	}
	if got.Mandatory.BytesScannedCutoffPerQuery != 0 {
		t.Errorf("all-absent: mandatory.bytesScannedCutoffPerQuery = %d, want 0", got.Mandatory.BytesScannedCutoffPerQuery)
	}
	if got.Mandatory.ResultEncryptionOption != "" {
		t.Errorf("all-absent: mandatory.resultEncryptionOption = %q, want empty", got.Mandatory.ResultEncryptionOption)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("all-absent: mandatory.namingTemplate = %q, want empty", got.Mandatory.NamingTemplate)
	}
	if len(got.Mandatory.Tags) != 0 {
		t.Errorf("all-absent: mandatory.tags = %v, want empty", got.Mandatory.Tags)
	}
	if got.Defaults.EnforceWorkGroupConfiguration {
		t.Errorf("all-absent: defaults.enforceWorkGroupConfiguration = true, want false")
	}
	if got.Defaults.BytesScannedCutoffPerQuery != 0 {
		t.Errorf("all-absent: defaults.bytesScannedCutoffPerQuery = %d, want 0", got.Defaults.BytesScannedCutoffPerQuery)
	}
	if len(got.Defaults.Tags) != 0 {
		t.Errorf("all-absent: defaults.tags = %v, want empty", got.Defaults.Tags)
	}
}

// TestMergeAthenaCascade_BytesScannedCutoffPerQuery_Level3WinsLevel4 verifies
// integer priority: level 3 wins over level 4 for bytesScannedCutoffPerQuery.
func TestMergeAthenaCascade_BytesScannedCutoffPerQuery_Level3WinsLevel4(t *testing.T) {
	got := mergeAthenaAll(
		zeroAthenaKropath,
		zeroAthenaKropath,
		cascade.AthenaConfigSection{BytesScannedCutoffPerQuery: 10737418240}, // level 3
		cascade.AthenaConfigSection{BytesScannedCutoffPerQuery: 5368709120},  // level 4
		zeroAthenaCfg,
		zeroAthenaCfg,
		zeroAthenaKropath,
		zeroAthenaKropath,
	)

	if got.Mandatory.BytesScannedCutoffPerQuery != 10737418240 {
		t.Errorf("mandatory.bytesScannedCutoffPerQuery = %d, want 10737418240 (level 3 wins over level 4)", got.Mandatory.BytesScannedCutoffPerQuery)
	}
}

// TestMergeAthenaCascade_EngineVersion_DefaultsPriority verifies string field
// priority in the defaults chain: level 6 > 7.
func TestMergeAthenaCascade_EngineVersion_DefaultsPriority(t *testing.T) {
	got := mergeAthenaAll(
		zeroAthenaKropath,
		zeroAthenaKropath,
		zeroAthenaCfg,
		zeroAthenaCfg,
		cascade.AthenaConfigSection{EngineVersion: "Athena engine version 3"},      // level 6
		cascade.AthenaConfigSection{EngineVersion: "Athena engine version 2"},      // level 7
		zeroAthenaKropath,
		zeroAthenaKropath,
	)

	if got.Defaults.EngineVersion != "Athena engine version 3" {
		t.Errorf("defaults.engineVersion = %q, want Athena engine version 3 (level 6 wins over level 7)", got.Defaults.EngineVersion)
	}
}
