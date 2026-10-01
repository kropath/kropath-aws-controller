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

// zeroKropathOS is a zero-value OpenSearchKropathSection (absent source).
var zeroKropathOS = cascade.OpenSearchKropathSection{}

// zeroOSCfg is a zero-value OpenSearchConfigSection (absent source).
var zeroOSCfg = cascade.OpenSearchConfigSection{}

// mergeOSAll calls MergeOpenSearchCascade with all eight inputs.
func mergeOSAll(
	globalKropathMandatory,
	localKropathMandatory cascade.OpenSearchKropathSection,
	globalOSCfgMandatory,
	localOSCfgMandatory,
	localOSCfgDefaults,
	globalOSCfgDefaults cascade.OpenSearchConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.OpenSearchKropathSection,
) cascade.EffectiveOpenSearchConfig {
	return cascade.MergeOpenSearchCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalOSCfgMandatory,
		localOSCfgMandatory,
		localOSCfgDefaults,
		globalOSCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeOpenSearchCascade_AC4 — KropathConfig(global).mandatory.opensearch.encryptionAtRestEnabled=true
// (level 1) wins over OpenSearchConfig.mandatory.encryptionAtRestEnabled=false (levels 3-4);
// OpenSearchConfig.defaults.encryptionAtRestEnabled=true (level 6) propagates independently.
func TestMergeOpenSearchCascade_AC4(t *testing.T) {
	got := mergeOSAll(
		cascade.OpenSearchKropathSection{EncryptionAtRestEnabled: true}, // level 1
		zeroKropathOS,
		cascade.OpenSearchConfigSection{EncryptionAtRestEnabled: false}, // level 3 (false, no effect)
		cascade.OpenSearchConfigSection{EncryptionAtRestEnabled: false, // level 4
			TLSSecurityPolicy: ""},
		cascade.OpenSearchConfigSection{EncryptionAtRestEnabled: true}, // level 6
		zeroOSCfg,
		zeroKropathOS,
		zeroKropathOS,
	)

	if !got.Mandatory.EncryptionAtRestEnabled {
		t.Errorf("AC-4: mandatory.encryptionAtRestEnabled = false, want true (level 1 wins)")
	}
	if !got.Defaults.EncryptionAtRestEnabled {
		t.Errorf("AC-4: defaults.encryptionAtRestEnabled = false, want true (level 6)")
	}
}

// TestMergeOpenSearchCascade_AC5 — KropathConfig(global).defaults.opensearch.tlsSecurityPolicy
// is level-9 (weakest); OpenSearchConfig(local).defaults.tlsSecurityPolicy at level-6 wins.
func TestMergeOpenSearchCascade_AC5(t *testing.T) {
	got := mergeOSAll(
		zeroKropathOS,
		zeroKropathOS,
		zeroOSCfg,
		zeroOSCfg,
		cascade.OpenSearchConfigSection{TLSSecurityPolicy: "Policy-Min-TLS-1-2-2019-07"}, // level 6
		zeroOSCfg,
		zeroKropathOS,
		cascade.OpenSearchKropathSection{TLSSecurityPolicy: "Policy-Min-TLS-1-0-2019-10"}, // level 9
	)

	if got.Defaults.TLSSecurityPolicy != "Policy-Min-TLS-1-2-2019-07" {
		t.Errorf("AC-5: defaults.tlsSecurityPolicy = %q, want Policy-Min-TLS-1-2-2019-07 (level 6 wins over level 9)",
			got.Defaults.TLSSecurityPolicy)
	}
}

// TestMergeOpenSearchCascade_AC6 — no KropathConfig.opensearch section;
// OpenSearchConfig(global).mandatory.engineVersion="OpenSearch_2.9" (level 3) propagates.
func TestMergeOpenSearchCascade_AC6(t *testing.T) {
	got := mergeOSAll(
		zeroKropathOS,
		zeroKropathOS,
		cascade.OpenSearchConfigSection{EngineVersion: "OpenSearch_2.9"}, // level 3
		zeroOSCfg,
		zeroOSCfg,
		zeroOSCfg,
		zeroKropathOS,
		zeroKropathOS,
	)

	if got.Mandatory.EngineVersion != "OpenSearch_2.9" {
		t.Errorf("AC-6: mandatory.engineVersion = %q, want OpenSearch_2.9 (level 3)", got.Mandatory.EngineVersion)
	}
}

// TestMergeOpenSearchCascade_AC7 — tags union-merged across mandatory sources;
// KropathConfig.mandatory.tags and OpenSearchConfig.mandatory.tags merge into mandatory.tags;
// OpenSearchConfig.defaults.tags appears in defaults.tags only.
func TestMergeOpenSearchCascade_AC7(t *testing.T) {
	got := mergeOSAll(
		cascade.OpenSearchKropathSection{Tags: map[string]string{"cost-centre": "platform"}}, // level 1
		zeroKropathOS,
		cascade.OpenSearchConfigSection{Tags: map[string]string{"team": "search"}}, // level 3
		zeroOSCfg,
		cascade.OpenSearchConfigSection{Tags: map[string]string{"env": "dev"}}, // level 6
		zeroOSCfg,
		zeroKropathOS,
		zeroKropathOS,
	)

	if got.Mandatory.Tags["cost-centre"] != "platform" {
		t.Errorf("AC-7: mandatory.tags[cost-centre] = %q, want platform (level 1)", got.Mandatory.Tags["cost-centre"])
	}
	if got.Mandatory.Tags["team"] != "search" {
		t.Errorf("AC-7: mandatory.tags[team] = %q, want search (level 3, additive)", got.Mandatory.Tags["team"])
	}
	if got.Defaults.Tags["env"] != "dev" {
		t.Errorf("AC-7: defaults.tags[env] = %q, want dev (level 6)", got.Defaults.Tags["env"])
	}
	if _, ok := got.Mandatory.Tags["env"]; ok {
		t.Errorf("AC-7: mandatory.tags[env] present, want absent (defaults must not bleed into mandatory)")
	}
}

// TestMergeOpenSearchCascade_KropathMandatoryLevel1WinsEncryption — level 1 beats levels 2-4.
func TestMergeOpenSearchCascade_KropathMandatoryLevel1WinsEncryption(t *testing.T) {
	got := mergeOSAll(
		cascade.OpenSearchKropathSection{EncryptionAtRestEnabled: true}, // level 1
		cascade.OpenSearchKropathSection{EncryptionAtRestEnabled: false},
		cascade.OpenSearchConfigSection{EncryptionAtRestEnabled: false},
		cascade.OpenSearchConfigSection{EncryptionAtRestEnabled: false},
		zeroOSCfg, zeroOSCfg,
		zeroKropathOS, zeroKropathOS,
	)
	if !got.Mandatory.EncryptionAtRestEnabled {
		t.Errorf("mandatory.encryptionAtRestEnabled = false, want true (level 1 wins)")
	}
}

// TestMergeOpenSearchCascade_Level3WinsWhen12Absent — OSCfg global mandatory (level 3)
// propagates when KropathConfig levels are absent.
func TestMergeOpenSearchCascade_Level3WinsWhen12Absent(t *testing.T) {
	got := mergeOSAll(
		zeroKropathOS, zeroKropathOS,
		cascade.OpenSearchConfigSection{EncryptionAtRestEnabled: true}, // level 3
		zeroOSCfg,
		zeroOSCfg, zeroOSCfg,
		zeroKropathOS, zeroKropathOS,
	)
	if !got.Mandatory.EncryptionAtRestEnabled {
		t.Errorf("mandatory.encryptionAtRestEnabled = false, want true (level 3 wins when 1-2 false)")
	}
}

// TestMergeOpenSearchCascade_Level4Only — local OSCfg mandatory (level 4) propagates alone.
func TestMergeOpenSearchCascade_Level4Only(t *testing.T) {
	got := mergeOSAll(
		zeroKropathOS, zeroKropathOS,
		zeroOSCfg,
		cascade.OpenSearchConfigSection{NodeToNodeEncryptionEnabled: true}, // level 4
		zeroOSCfg, zeroOSCfg,
		zeroKropathOS, zeroKropathOS,
	)
	if !got.Mandatory.NodeToNodeEncryptionEnabled {
		t.Errorf("mandatory.nodeToNodeEncryptionEnabled = false, want true (level 4)")
	}
}

// TestMergeOpenSearchCascade_DefaultsLevel6WinsOverLevel7 — local OSCfg defaults (level 6)
// wins over global OSCfg defaults (level 7) for enforceHTTPS.
func TestMergeOpenSearchCascade_DefaultsLevel6WinsOverLevel7(t *testing.T) {
	got := mergeOSAll(
		zeroKropathOS, zeroKropathOS,
		zeroOSCfg, zeroOSCfg,
		cascade.OpenSearchConfigSection{EnforceHTTPS: true},  // level 6
		cascade.OpenSearchConfigSection{EnforceHTTPS: false}, // level 7
		zeroKropathOS, zeroKropathOS,
	)
	if !got.Defaults.EnforceHTTPS {
		t.Errorf("defaults.enforceHTTPS = false, want true (level 6 wins over level 7)")
	}
}

// TestMergeOpenSearchCascade_DefaultsLevel8WinsOverLevel9 — local KropathConfig defaults (level 8)
// wins over global KropathConfig defaults (level 9) for tlsSecurityPolicy.
func TestMergeOpenSearchCascade_DefaultsLevel8WinsOverLevel9(t *testing.T) {
	got := mergeOSAll(
		zeroKropathOS, zeroKropathOS,
		zeroOSCfg, zeroOSCfg,
		zeroOSCfg, zeroOSCfg,
		cascade.OpenSearchKropathSection{TLSSecurityPolicy: "Policy-Min-TLS-1-2-2019-07"},  // level 8
		cascade.OpenSearchKropathSection{TLSSecurityPolicy: "Policy-Min-TLS-1-0-2019-10"},  // level 9
	)
	if got.Defaults.TLSSecurityPolicy != "Policy-Min-TLS-1-2-2019-07" {
		t.Errorf("defaults.tlsSecurityPolicy = %q, want Policy-Min-TLS-1-2-2019-07 (level 8 wins over level 9)",
			got.Defaults.TLSSecurityPolicy)
	}
}

// TestMergeOpenSearchCascade_MandatoryIsolatedFromDefaults — mandatory fields must not
// bleed into defaults and vice versa.
func TestMergeOpenSearchCascade_MandatoryIsolatedFromDefaults(t *testing.T) {
	got := mergeOSAll(
		cascade.OpenSearchKropathSection{EncryptionAtRestEnabled: true}, // level 1 mandatory
		zeroKropathOS,
		cascade.OpenSearchConfigSection{EngineVersion: "OpenSearch_2.9"}, // level 3 mandatory
		zeroOSCfg,
		cascade.OpenSearchConfigSection{EngineVersion: "OpenSearch_2.11", AdvancedSecurityEnabled: true}, // level 6 defaults
		zeroOSCfg,
		zeroKropathOS,
		zeroKropathOS,
	)

	if !got.Mandatory.EncryptionAtRestEnabled {
		t.Errorf("mandatory.encryptionAtRestEnabled = false, want true")
	}
	if got.Mandatory.EngineVersion != "OpenSearch_2.9" {
		t.Errorf("mandatory.engineVersion = %q, want OpenSearch_2.9", got.Mandatory.EngineVersion)
	}
	// defaults must not pick up mandatory values
	if got.Defaults.EncryptionAtRestEnabled {
		t.Errorf("defaults.encryptionAtRestEnabled = true, must not bleed from mandatory")
	}
	if got.Defaults.EngineVersion != "OpenSearch_2.11" {
		t.Errorf("defaults.engineVersion = %q, want OpenSearch_2.11 (level 6)", got.Defaults.EngineVersion)
	}
	if !got.Defaults.AdvancedSecurityEnabled {
		t.Errorf("defaults.advancedSecurityEnabled = false, want true (level 6)")
	}
	// mandatory must not pick up defaults values
	if got.Mandatory.AdvancedSecurityEnabled {
		t.Errorf("mandatory.advancedSecurityEnabled = true, must not bleed from defaults")
	}
}

// TestMergeOpenSearchCascade_AllAbsent — all-zero inputs produce all-zero output.
func TestMergeOpenSearchCascade_AllAbsent(t *testing.T) {
	got := mergeOSAll(
		zeroKropathOS, zeroKropathOS,
		zeroOSCfg, zeroOSCfg, zeroOSCfg, zeroOSCfg,
		zeroKropathOS, zeroKropathOS,
	)

	if got.Mandatory.EncryptionAtRestEnabled {
		t.Errorf("all-absent: mandatory.encryptionAtRestEnabled = true, want false")
	}
	if got.Mandatory.NodeToNodeEncryptionEnabled {
		t.Errorf("all-absent: mandatory.nodeToNodeEncryptionEnabled = true, want false")
	}
	if got.Mandatory.EnforceHTTPS {
		t.Errorf("all-absent: mandatory.enforceHTTPS = true, want false")
	}
	if got.Mandatory.TLSSecurityPolicy != "" {
		t.Errorf("all-absent: mandatory.tlsSecurityPolicy = %q, want empty", got.Mandatory.TLSSecurityPolicy)
	}
	if got.Mandatory.AdvancedSecurityEnabled {
		t.Errorf("all-absent: mandatory.advancedSecurityEnabled = true, want false")
	}
	if got.Mandatory.EngineVersion != "" {
		t.Errorf("all-absent: mandatory.engineVersion = %q, want empty", got.Mandatory.EngineVersion)
	}
	if got.Mandatory.AutoTuneDesiredState != "" {
		t.Errorf("all-absent: mandatory.autoTuneDesiredState = %q, want empty", got.Mandatory.AutoTuneDesiredState)
	}
	if got.Mandatory.StandbyReplicas != "" {
		t.Errorf("all-absent: mandatory.standbyReplicas = %q, want empty", got.Mandatory.StandbyReplicas)
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

// TestMergeOpenSearchCascade_TagUnionMerge — tags from multiple mandatory sources are
// union-merged; higher-priority source wins on key conflict.
func TestMergeOpenSearchCascade_TagUnionMerge(t *testing.T) {
	got := mergeOSAll(
		cascade.OpenSearchKropathSection{Tags: map[string]string{"owner": "platform", "cost-centre": "infra"}}, // level 1 wins
		zeroKropathOS,
		cascade.OpenSearchConfigSection{Tags: map[string]string{"cost-centre": "payments", "env": "prod"}}, // level 3
		zeroOSCfg,
		zeroOSCfg, zeroOSCfg,
		zeroKropathOS, zeroKropathOS,
	)

	if got.Mandatory.Tags["cost-centre"] != "infra" {
		t.Errorf("tags: mandatory.tags[cost-centre] = %q, want infra (level 1 wins)", got.Mandatory.Tags["cost-centre"])
	}
	if got.Mandatory.Tags["owner"] != "platform" {
		t.Errorf("tags: mandatory.tags[owner] = %q, want platform", got.Mandatory.Tags["owner"])
	}
	if got.Mandatory.Tags["env"] != "prod" {
		t.Errorf("tags: mandatory.tags[env] = %q, want prod (additive)", got.Mandatory.Tags["env"])
	}
}

// TestMergeOpenSearchCascade_SyncedLabelsUnionMerge — syncedLabels from OSCfg levels only;
// level 3 wins over level 4 on key conflict.
func TestMergeOpenSearchCascade_SyncedLabelsUnionMerge(t *testing.T) {
	got := mergeOSAll(
		zeroKropathOS, zeroKropathOS,
		cascade.OpenSearchConfigSection{SyncedLabels: map[string]string{"team": "platform", "data-class": "public"}},    // level 3 wins
		cascade.OpenSearchConfigSection{SyncedLabels: map[string]string{"data-class": "internal", "region": "ap-se-2"}}, // level 4
		zeroOSCfg, zeroOSCfg,
		zeroKropathOS, zeroKropathOS,
	)

	if got.Mandatory.SyncedLabels["data-class"] != "public" {
		t.Errorf("syncedLabels: mandatory.syncedLabels[data-class] = %q, want public (level 3 wins)",
			got.Mandatory.SyncedLabels["data-class"])
	}
	if got.Mandatory.SyncedLabels["team"] != "platform" {
		t.Errorf("syncedLabels: mandatory.syncedLabels[team] = %q, want platform", got.Mandatory.SyncedLabels["team"])
	}
	if got.Mandatory.SyncedLabels["region"] != "ap-se-2" {
		t.Errorf("syncedLabels: mandatory.syncedLabels[region] = %q, want ap-se-2 (additive)", got.Mandatory.SyncedLabels["region"])
	}
}

// TestMergeOpenSearchCascade_NamingTemplate — NamingTemplate is OSCfg-only; KropathConfig
// levels are irrelevant. Levels 3-4 govern mandatory, 6-7 govern defaults.
func TestMergeOpenSearchCascade_NamingTemplate(t *testing.T) {
	got := mergeOSAll(
		zeroKropathOS, zeroKropathOS,
		zeroOSCfg,
		cascade.OpenSearchConfigSection{NamingTemplate: "{namespace}-{name}"}, // level 4
		zeroOSCfg,
		cascade.OpenSearchConfigSection{NamingTemplate: "{namespace}-{name}-default"}, // level 7
		zeroKropathOS, zeroKropathOS,
	)

	if got.Mandatory.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("mandatory.namingTemplate = %q, want {namespace}-{name} (level 4)", got.Mandatory.NamingTemplate)
	}
	if got.Defaults.NamingTemplate != "{namespace}-{name}-default" {
		t.Errorf("defaults.namingTemplate = %q, want {namespace}-{name}-default (level 7)", got.Defaults.NamingTemplate)
	}
}

// TestMergeOpenSearchCascade_EncryptionMandatoryPriorityOrder — full priority chain 1 > 2 > 3 > 4.
func TestMergeOpenSearchCascade_EncryptionMandatoryPriorityOrder(t *testing.T) {
	cases := []struct {
		name                   string
		globalKropathMandatory cascade.OpenSearchKropathSection
		localKropathMandatory  cascade.OpenSearchKropathSection
		globalOSCfgMandatory   cascade.OpenSearchConfigSection
		localOSCfgMandatory    cascade.OpenSearchConfigSection
		wantEncryption         bool
	}{
		{
			name:                   "level1-wins",
			globalKropathMandatory: cascade.OpenSearchKropathSection{EncryptionAtRestEnabled: true},
			localKropathMandatory:  cascade.OpenSearchKropathSection{EncryptionAtRestEnabled: false},
			globalOSCfgMandatory:   cascade.OpenSearchConfigSection{EncryptionAtRestEnabled: false},
			localOSCfgMandatory:    cascade.OpenSearchConfigSection{EncryptionAtRestEnabled: false},
			wantEncryption:         true,
		},
		{
			name:                   "level2-wins-when-1-false",
			globalKropathMandatory: zeroKropathOS,
			localKropathMandatory:  cascade.OpenSearchKropathSection{EncryptionAtRestEnabled: true},
			globalOSCfgMandatory:   cascade.OpenSearchConfigSection{EncryptionAtRestEnabled: false},
			localOSCfgMandatory:    cascade.OpenSearchConfigSection{EncryptionAtRestEnabled: false},
			wantEncryption:         true,
		},
		{
			name:                   "level3-wins-when-1-2-absent",
			globalKropathMandatory: zeroKropathOS,
			localKropathMandatory:  zeroKropathOS,
			globalOSCfgMandatory:   cascade.OpenSearchConfigSection{EncryptionAtRestEnabled: true},
			localOSCfgMandatory:    cascade.OpenSearchConfigSection{EncryptionAtRestEnabled: false},
			wantEncryption:         true,
		},
		{
			name:                   "level4-only",
			globalKropathMandatory: zeroKropathOS,
			localKropathMandatory:  zeroKropathOS,
			globalOSCfgMandatory:   zeroOSCfg,
			localOSCfgMandatory:    cascade.OpenSearchConfigSection{EncryptionAtRestEnabled: true},
			wantEncryption:         true,
		},
		{
			name:                   "all-false",
			globalKropathMandatory: zeroKropathOS,
			localKropathMandatory:  zeroKropathOS,
			globalOSCfgMandatory:   zeroOSCfg,
			localOSCfgMandatory:    zeroOSCfg,
			wantEncryption:         false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeOSAll(
				tc.globalKropathMandatory,
				tc.localKropathMandatory,
				tc.globalOSCfgMandatory,
				tc.localOSCfgMandatory,
				zeroOSCfg, zeroOSCfg,
				zeroKropathOS, zeroKropathOS,
			)
			if got.Mandatory.EncryptionAtRestEnabled != tc.wantEncryption {
				t.Errorf("mandatory.encryptionAtRestEnabled = %v, want %v", got.Mandatory.EncryptionAtRestEnabled, tc.wantEncryption)
			}
		})
	}
}

// TestMergeOpenSearchCascade_EncryptionDefaultsPriorityOrder — full defaults chain 6 > 7 > 8 > 9.
func TestMergeOpenSearchCascade_EncryptionDefaultsPriorityOrder(t *testing.T) {
	cases := []struct {
		name                  string
		localOSCfgDefaults    cascade.OpenSearchConfigSection
		globalOSCfgDefaults   cascade.OpenSearchConfigSection
		localKropathDefaults  cascade.OpenSearchKropathSection
		globalKropathDefaults cascade.OpenSearchKropathSection
		wantEncryption        bool
	}{
		{
			name:                  "level6-wins",
			localOSCfgDefaults:    cascade.OpenSearchConfigSection{EncryptionAtRestEnabled: true},
			globalOSCfgDefaults:   cascade.OpenSearchConfigSection{EncryptionAtRestEnabled: false},
			localKropathDefaults:  cascade.OpenSearchKropathSection{EncryptionAtRestEnabled: false},
			globalKropathDefaults: cascade.OpenSearchKropathSection{EncryptionAtRestEnabled: false},
			wantEncryption:        true,
		},
		{
			name:                  "level7-wins-when-6-absent",
			localOSCfgDefaults:    zeroOSCfg,
			globalOSCfgDefaults:   cascade.OpenSearchConfigSection{EncryptionAtRestEnabled: true},
			localKropathDefaults:  cascade.OpenSearchKropathSection{EncryptionAtRestEnabled: false},
			globalKropathDefaults: cascade.OpenSearchKropathSection{EncryptionAtRestEnabled: false},
			wantEncryption:        true,
		},
		{
			name:                  "level8-wins-when-6-7-absent",
			localOSCfgDefaults:    zeroOSCfg,
			globalOSCfgDefaults:   zeroOSCfg,
			localKropathDefaults:  cascade.OpenSearchKropathSection{EncryptionAtRestEnabled: true},
			globalKropathDefaults: cascade.OpenSearchKropathSection{EncryptionAtRestEnabled: false},
			wantEncryption:        true,
		},
		{
			name:                  "level9-fallback",
			localOSCfgDefaults:    zeroOSCfg,
			globalOSCfgDefaults:   zeroOSCfg,
			localKropathDefaults:  zeroKropathOS,
			globalKropathDefaults: cascade.OpenSearchKropathSection{EncryptionAtRestEnabled: true},
			wantEncryption:        true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeOSAll(
				zeroKropathOS, zeroKropathOS,
				zeroOSCfg, zeroOSCfg,
				tc.localOSCfgDefaults,
				tc.globalOSCfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
			)
			if got.Defaults.EncryptionAtRestEnabled != tc.wantEncryption {
				t.Errorf("defaults.encryptionAtRestEnabled = %v, want %v", got.Defaults.EncryptionAtRestEnabled, tc.wantEncryption)
			}
		})
	}
}

// TestMergeOpenSearchCascade_NoKropathLevelsForNonSecurityFields — fields without KropathConfig
// equivalents are correctly bounded to OSCfg levels only.
func TestMergeOpenSearchCascade_NoKropathLevelsForNonSecurityFields(t *testing.T) {
	got := mergeOSAll(
		// KropathConfig levels set for security fields only
		cascade.OpenSearchKropathSection{EncryptionAtRestEnabled: true, NodeToNodeEncryptionEnabled: true},
		zeroKropathOS,
		// OSCfg mandatory sets the non-security fields
		cascade.OpenSearchConfigSection{
			AdvancedSecurityEnabled: true,
			EngineVersion:           "OpenSearch_2.9",
			AutoTuneDesiredState:    "ENABLED",
			StandbyReplicas:         "ENABLED",
		},
		zeroOSCfg,
		zeroOSCfg, zeroOSCfg,
		zeroKropathOS, zeroKropathOS,
	)

	if !got.Mandatory.AdvancedSecurityEnabled {
		t.Errorf("mandatory.advancedSecurityEnabled = false, want true")
	}
	if got.Mandatory.EngineVersion != "OpenSearch_2.9" {
		t.Errorf("mandatory.engineVersion = %q, want OpenSearch_2.9", got.Mandatory.EngineVersion)
	}
	if got.Mandatory.AutoTuneDesiredState != "ENABLED" {
		t.Errorf("mandatory.autoTuneDesiredState = %q, want ENABLED", got.Mandatory.AutoTuneDesiredState)
	}
	if got.Mandatory.StandbyReplicas != "ENABLED" {
		t.Errorf("mandatory.standbyReplicas = %q, want ENABLED", got.Mandatory.StandbyReplicas)
	}
	// Security fields from KropathConfig also propagate
	if !got.Mandatory.EncryptionAtRestEnabled {
		t.Errorf("mandatory.encryptionAtRestEnabled = false, want true (level 1)")
	}
}
