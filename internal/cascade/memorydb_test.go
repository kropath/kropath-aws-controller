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

// zeroKropathMDB is a zero-value MemoryDBKropathSection (absent source).
var zeroKropathMDB = cascade.MemoryDBKropathSection{}

// zeroMDBCfg is a zero-value MemoryDBConfigSection (absent source).
var zeroMDBCfg = cascade.MemoryDBConfigSection{}

// mergeMDBAll calls MergeMemoryDBCascade with all eight inputs.
func mergeMDBAll(
	globalKropathMandatory,
	localKropathMandatory cascade.MemoryDBKropathSection,
	globalMDBCfgMandatory,
	localMDBCfgMandatory,
	localMDBCfgDefaults,
	globalMDBCfgDefaults cascade.MemoryDBConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.MemoryDBKropathSection,
) cascade.EffectiveMemoryDBConfig {
	return cascade.MergeMemoryDBCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalMDBCfgMandatory,
		localMDBCfgMandatory,
		localMDBCfgDefaults,
		globalMDBCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeMemoryDBCascade_AC1 — KropathConfig(global).mandatory.memorydb.tlsEnabled=true
// (level 1) wins over MemoryDBConfig.mandatory.tlsEnabled=false (levels 3-4).
func TestMergeMemoryDBCascade_AC1(t *testing.T) {
	got := mergeMDBAll(
		cascade.MemoryDBKropathSection{TLSEnabled: true}, // level 1
		zeroKropathMDB,
		cascade.MemoryDBConfigSection{TLSEnabled: false}, // level 3
		cascade.MemoryDBConfigSection{TLSEnabled: false}, // level 4
		zeroMDBCfg,
		zeroMDBCfg,
		zeroKropathMDB,
		zeroKropathMDB,
	)

	if !got.Mandatory.TLSEnabled {
		t.Errorf("AC-1: mandatory.tlsEnabled = false, want true (level 1 wins)")
	}
	if got.Defaults.TLSEnabled {
		t.Errorf("AC-1: defaults.tlsEnabled = true, must not bleed from mandatory")
	}
}

// TestMergeMemoryDBCascade_AC2 — MemoryDBConfig(local).mandatory.tlsEnabled=true (level 4)
// wins when levels 1-3 are false.
func TestMergeMemoryDBCascade_AC2(t *testing.T) {
	got := mergeMDBAll(
		zeroKropathMDB,
		zeroKropathMDB,
		zeroMDBCfg,
		cascade.MemoryDBConfigSection{TLSEnabled: true}, // level 4
		zeroMDBCfg,
		zeroMDBCfg,
		zeroKropathMDB,
		zeroKropathMDB,
	)

	if !got.Mandatory.TLSEnabled {
		t.Errorf("AC-2: mandatory.tlsEnabled = false, want true (level 4 wins when 1-3 false)")
	}
}

// TestMergeMemoryDBCascade_AC3 — all mandatory tiers false; local MemoryDBConfig.defaults.tlsEnabled=true
// (level 6) propagates; mandatory stays false.
func TestMergeMemoryDBCascade_AC3(t *testing.T) {
	got := mergeMDBAll(
		zeroKropathMDB,
		zeroKropathMDB,
		zeroMDBCfg,
		zeroMDBCfg,
		cascade.MemoryDBConfigSection{TLSEnabled: true}, // level 6
		zeroMDBCfg,
		zeroKropathMDB,
		zeroKropathMDB,
	)

	if got.Mandatory.TLSEnabled {
		t.Errorf("AC-3: mandatory.tlsEnabled = true, want false (defaults must not bleed into mandatory)")
	}
	if !got.Defaults.TLSEnabled {
		t.Errorf("AC-3: defaults.tlsEnabled = false, want true (level 6)")
	}
}

// TestMergeMemoryDBCascade_AC4 — KropathConfig(global).mandatory.memorydb.allowedNodeTypes
// (level 1) wins over MemoryDBConfig.mandatory.allowedNodeTypes (levels 3-4).
func TestMergeMemoryDBCascade_AC4(t *testing.T) {
	got := mergeMDBAll(
		cascade.MemoryDBKropathSection{AllowedNodeTypes: []string{"db.r7g.large"}}, // level 1
		zeroKropathMDB,
		cascade.MemoryDBConfigSection{AllowedNodeTypes: []string{"db.t4g.small"}}, // level 3
		zeroMDBCfg,
		zeroMDBCfg,
		zeroMDBCfg,
		zeroKropathMDB,
		zeroKropathMDB,
	)

	if len(got.Mandatory.AllowedNodeTypes) != 1 || got.Mandatory.AllowedNodeTypes[0] != "db.r7g.large" {
		t.Errorf("AC-4: mandatory.allowedNodeTypes = %v, want [db.r7g.large] (level 1 wins)", got.Mandatory.AllowedNodeTypes)
	}
}

// TestMergeMemoryDBCascade_AC5 — MemoryDBConfig(local).mandatory.allowedNodeTypes (level 4)
// wins when levels 1-3 are absent.
func TestMergeMemoryDBCascade_AC5(t *testing.T) {
	got := mergeMDBAll(
		zeroKropathMDB,
		zeroKropathMDB,
		zeroMDBCfg,
		cascade.MemoryDBConfigSection{AllowedNodeTypes: []string{"db.r7g.large", "db.r7g.xlarge"}}, // level 4
		zeroMDBCfg,
		zeroMDBCfg,
		zeroKropathMDB,
		zeroKropathMDB,
	)

	if len(got.Mandatory.AllowedNodeTypes) != 2 {
		t.Errorf("AC-5: mandatory.allowedNodeTypes = %v, want 2 entries (level 4)", got.Mandatory.AllowedNodeTypes)
	}
}

// TestMergeMemoryDBCascade_AC6 — KropathConfig(global).mandatory.memorydb.snapshotRetentionLimit=14
// (level 1) wins over MemoryDBConfig.mandatory.snapshotRetentionLimit=7 (level 3).
func TestMergeMemoryDBCascade_AC6(t *testing.T) {
	got := mergeMDBAll(
		cascade.MemoryDBKropathSection{SnapshotRetentionLimit: 14}, // level 1
		zeroKropathMDB,
		cascade.MemoryDBConfigSection{SnapshotRetentionLimit: 7}, // level 3
		zeroMDBCfg,
		zeroMDBCfg,
		zeroMDBCfg,
		zeroKropathMDB,
		zeroKropathMDB,
	)

	if got.Mandatory.SnapshotRetentionLimit != 14 {
		t.Errorf("AC-6: mandatory.snapshotRetentionLimit = %d, want 14 (level 1 wins)", got.Mandatory.SnapshotRetentionLimit)
	}
}

// TestMergeMemoryDBCascade_AC7 — MemoryDBConfig(local).mandatory.snapshotRetentionLimit=7 (level 4)
// wins when levels 1-3 are absent.
func TestMergeMemoryDBCascade_AC7(t *testing.T) {
	got := mergeMDBAll(
		zeroKropathMDB,
		zeroKropathMDB,
		zeroMDBCfg,
		cascade.MemoryDBConfigSection{SnapshotRetentionLimit: 7}, // level 4
		zeroMDBCfg,
		zeroMDBCfg,
		zeroKropathMDB,
		zeroKropathMDB,
	)

	if got.Mandatory.SnapshotRetentionLimit != 7 {
		t.Errorf("AC-7: mandatory.snapshotRetentionLimit = %d, want 7 (level 4)", got.Mandatory.SnapshotRetentionLimit)
	}
}

// TestMergeMemoryDBCascade_AC8 — MemoryDBConfig(global).defaults.snapshotRetentionLimit=7 (level 7)
// wins when no mandatory source is set; KropathConfig defaults (levels 8-9) provide additional fallback.
func TestMergeMemoryDBCascade_AC8(t *testing.T) {
	got := mergeMDBAll(
		zeroKropathMDB,
		zeroKropathMDB,
		zeroMDBCfg,
		zeroMDBCfg,
		zeroMDBCfg,
		cascade.MemoryDBConfigSection{SnapshotRetentionLimit: 7}, // level 7
		zeroKropathMDB,
		zeroKropathMDB,
	)

	if got.Mandatory.SnapshotRetentionLimit != 0 {
		t.Errorf("AC-8: mandatory.snapshotRetentionLimit = %d, want 0 (defaults must not bleed into mandatory)", got.Mandatory.SnapshotRetentionLimit)
	}
	if got.Defaults.SnapshotRetentionLimit != 7 {
		t.Errorf("AC-8: defaults.snapshotRetentionLimit = %d, want 7 (level 7)", got.Defaults.SnapshotRetentionLimit)
	}
}

// TestMergeMemoryDBCascade_AC9_NodeTypeValidation — when mandatory.nodeType is set and not in
// mandatory.allowedNodeTypes, ValidateMemoryDBNodeType returns (false, message).
func TestMergeMemoryDBCascade_AC9_NodeTypeValidation(t *testing.T) {
	mandatory := cascade.EffectiveMemoryDBSection{
		NodeType:         "db.r7g.large",
		AllowedNodeTypes: []string{"db.t4g.small", "db.t4g.medium"},
	}
	valid, msg := cascade.ValidateMemoryDBNodeType(mandatory)
	if valid {
		t.Errorf("AC-9: expected validation failure (nodeType not in allowedNodeTypes), got valid=true")
	}
	if msg == "" {
		t.Errorf("AC-9: expected non-empty error message")
	}
}

// TestMergeMemoryDBCascade_AC9_NodeTypeValidation_Valid — nodeType in allowedNodeTypes passes.
func TestMergeMemoryDBCascade_AC9_NodeTypeValidation_Valid(t *testing.T) {
	mandatory := cascade.EffectiveMemoryDBSection{
		NodeType:         "db.r7g.large",
		AllowedNodeTypes: []string{"db.r7g.large", "db.r7g.xlarge"},
	}
	valid, msg := cascade.ValidateMemoryDBNodeType(mandatory)
	if !valid {
		t.Errorf("AC-9 (valid): expected validation pass, got msg=%q", msg)
	}
}

// TestMergeMemoryDBCascade_AC9_NodeTypeValidation_NoAllowedList — when allowedNodeTypes is empty,
// any nodeType is valid (constraint does not apply).
func TestMergeMemoryDBCascade_AC9_NodeTypeValidation_NoAllowedList(t *testing.T) {
	mandatory := cascade.EffectiveMemoryDBSection{
		NodeType:         "db.r7g.large",
		AllowedNodeTypes: nil,
	}
	valid, _ := cascade.ValidateMemoryDBNodeType(mandatory)
	if !valid {
		t.Errorf("AC-9 (no list): expected valid when allowedNodeTypes is empty")
	}
}

// TestMergeMemoryDBCascade_AC10_TagLabelMerge — tags from multiple tiers are union-merged;
// syncedLabels propagates from MemoryDBConfig levels only.
func TestMergeMemoryDBCascade_AC10_TagLabelMerge(t *testing.T) {
	got := mergeMDBAll(
		cascade.MemoryDBKropathSection{Tags: map[string]string{"cost-centre": "infra"}}, // level 1
		zeroKropathMDB,
		cascade.MemoryDBConfigSection{
			Tags:         map[string]string{"service": "memorydb"},
			SyncedLabels: map[string]string{"data-class": "confidential"},
		}, // level 3
		zeroMDBCfg,
		zeroMDBCfg,
		zeroMDBCfg,
		zeroKropathMDB,
		zeroKropathMDB,
	)

	if got.Mandatory.Tags["cost-centre"] != "infra" {
		t.Errorf("AC-10: mandatory.tags[cost-centre] = %q, want infra", got.Mandatory.Tags["cost-centre"])
	}
	if got.Mandatory.Tags["service"] != "memorydb" {
		t.Errorf("AC-10: mandatory.tags[service] = %q, want memorydb", got.Mandatory.Tags["service"])
	}
	if got.Mandatory.SyncedLabels["data-class"] != "confidential" {
		t.Errorf("AC-10: mandatory.syncedLabels[data-class] = %q, want confidential", got.Mandatory.SyncedLabels["data-class"])
	}
}

// TestMergeMemoryDBCascade_AC11_NamingDefaults — MemoryDBConfig(global).defaults.namingTemplate (level 7)
// propagates when no mandatory naming template is set.
func TestMergeMemoryDBCascade_AC11_NamingDefaults(t *testing.T) {
	got := mergeMDBAll(
		zeroKropathMDB,
		zeroKropathMDB,
		zeroMDBCfg,
		zeroMDBCfg,
		zeroMDBCfg,
		cascade.MemoryDBConfigSection{NamingTemplate: "{namespace}-{name}"}, // level 7
		zeroKropathMDB,
		zeroKropathMDB,
	)

	if got.Defaults.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("AC-11: defaults.namingTemplate = %q, want {namespace}-{name} (level 7)", got.Defaults.NamingTemplate)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("AC-11: mandatory.namingTemplate = %q, want empty (no mandatory naming set)", got.Mandatory.NamingTemplate)
	}
}

// TestMergeMemoryDBCascade_AC12_ProviderIdentity — provider identity is tested at the reconciler
// level; at the cascade merge level, verify all-absent results are zero-value.
func TestMergeMemoryDBCascade_AC12_AllAbsent(t *testing.T) {
	got := mergeMDBAll(
		zeroKropathMDB, zeroKropathMDB,
		zeroMDBCfg, zeroMDBCfg, zeroMDBCfg, zeroMDBCfg,
		zeroKropathMDB, zeroKropathMDB,
	)

	if got.Mandatory.TLSEnabled {
		t.Errorf("all-absent: mandatory.tlsEnabled = true, want false")
	}
	if got.Mandatory.KMSKeyARN != "" {
		t.Errorf("all-absent: mandatory.kmsKeyArn = %q, want empty", got.Mandatory.KMSKeyARN)
	}
	if got.Mandatory.NodeType != "" {
		t.Errorf("all-absent: mandatory.nodeType = %q, want empty", got.Mandatory.NodeType)
	}
	if got.Mandatory.EngineVersion != "" {
		t.Errorf("all-absent: mandatory.engineVersion = %q, want empty", got.Mandatory.EngineVersion)
	}
	if len(got.Mandatory.AllowedNodeTypes) != 0 {
		t.Errorf("all-absent: mandatory.allowedNodeTypes = %v, want empty", got.Mandatory.AllowedNodeTypes)
	}
	if got.Mandatory.NumReplicasPerShard != 0 {
		t.Errorf("all-absent: mandatory.numReplicasPerShard = %d, want 0", got.Mandatory.NumReplicasPerShard)
	}
	if got.Mandatory.SnapshotRetentionLimit != 0 {
		t.Errorf("all-absent: mandatory.snapshotRetentionLimit = %d, want 0", got.Mandatory.SnapshotRetentionLimit)
	}
	if got.Mandatory.AutoMinorVersionUpgrade {
		t.Errorf("all-absent: mandatory.autoMinorVersionUpgrade = true, want false")
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

// TestMergeMemoryDBCascade_TLSMandatoryPriorityOrder — level 1 > 2 > 3 > 4 for tlsEnabled.
func TestMergeMemoryDBCascade_TLSMandatoryPriorityOrder(t *testing.T) {
	cases := []struct {
		name                   string
		globalKropathMandatory cascade.MemoryDBKropathSection
		localKropathMandatory  cascade.MemoryDBKropathSection
		globalMDBCfgMandatory  cascade.MemoryDBConfigSection
		localMDBCfgMandatory   cascade.MemoryDBConfigSection
		wantTLS                bool
	}{
		{
			name:                   "level1-wins",
			globalKropathMandatory: cascade.MemoryDBKropathSection{TLSEnabled: true},
			localKropathMandatory:  cascade.MemoryDBKropathSection{TLSEnabled: false},
			globalMDBCfgMandatory:  cascade.MemoryDBConfigSection{TLSEnabled: false},
			localMDBCfgMandatory:   cascade.MemoryDBConfigSection{TLSEnabled: false},
			wantTLS:                true,
		},
		{
			name:                   "level3-wins-when-1-2-false",
			globalKropathMandatory: zeroKropathMDB,
			localKropathMandatory:  zeroKropathMDB,
			globalMDBCfgMandatory:  cascade.MemoryDBConfigSection{TLSEnabled: true},
			localMDBCfgMandatory:   cascade.MemoryDBConfigSection{TLSEnabled: false},
			wantTLS:                true,
		},
		{
			name:                   "level4-only",
			globalKropathMandatory: zeroKropathMDB,
			localKropathMandatory:  zeroKropathMDB,
			globalMDBCfgMandatory:  zeroMDBCfg,
			localMDBCfgMandatory:   cascade.MemoryDBConfigSection{TLSEnabled: true},
			wantTLS:                true,
		},
		{
			name:                   "all-false",
			globalKropathMandatory: zeroKropathMDB,
			localKropathMandatory:  zeroKropathMDB,
			globalMDBCfgMandatory:  zeroMDBCfg,
			localMDBCfgMandatory:   zeroMDBCfg,
			wantTLS:                false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeMDBAll(
				tc.globalKropathMandatory,
				tc.localKropathMandatory,
				tc.globalMDBCfgMandatory,
				tc.localMDBCfgMandatory,
				zeroMDBCfg, zeroMDBCfg,
				zeroKropathMDB, zeroKropathMDB,
			)
			if got.Mandatory.TLSEnabled != tc.wantTLS {
				t.Errorf("mandatory.tlsEnabled = %v, want %v", got.Mandatory.TLSEnabled, tc.wantTLS)
			}
		})
	}
}

// TestMergeMemoryDBCascade_TLSDefaultsPriorityOrder — level 6 > 7 > 8 > 9.
func TestMergeMemoryDBCascade_TLSDefaultsPriorityOrder(t *testing.T) {
	cases := []struct {
		name                  string
		localMDBCfgDefaults   cascade.MemoryDBConfigSection
		globalMDBCfgDefaults  cascade.MemoryDBConfigSection
		localKropathDefaults  cascade.MemoryDBKropathSection
		globalKropathDefaults cascade.MemoryDBKropathSection
		wantTLS               bool
	}{
		{
			name:                  "level6-wins",
			localMDBCfgDefaults:   cascade.MemoryDBConfigSection{TLSEnabled: true},
			globalMDBCfgDefaults:  cascade.MemoryDBConfigSection{TLSEnabled: false},
			localKropathDefaults:  cascade.MemoryDBKropathSection{TLSEnabled: false},
			globalKropathDefaults: cascade.MemoryDBKropathSection{TLSEnabled: false},
			wantTLS:               true,
		},
		{
			name:                  "level7-wins-when-6-absent",
			localMDBCfgDefaults:   zeroMDBCfg,
			globalMDBCfgDefaults:  cascade.MemoryDBConfigSection{TLSEnabled: true},
			localKropathDefaults:  cascade.MemoryDBKropathSection{TLSEnabled: false},
			globalKropathDefaults: cascade.MemoryDBKropathSection{TLSEnabled: false},
			wantTLS:               true,
		},
		{
			name:                  "level8-wins-when-6-7-absent",
			localMDBCfgDefaults:   zeroMDBCfg,
			globalMDBCfgDefaults:  zeroMDBCfg,
			localKropathDefaults:  cascade.MemoryDBKropathSection{TLSEnabled: true},
			globalKropathDefaults: cascade.MemoryDBKropathSection{TLSEnabled: false},
			wantTLS:               true,
		},
		{
			name:                  "level9-fallback",
			localMDBCfgDefaults:   zeroMDBCfg,
			globalMDBCfgDefaults:  zeroMDBCfg,
			localKropathDefaults:  zeroKropathMDB,
			globalKropathDefaults: cascade.MemoryDBKropathSection{TLSEnabled: true},
			wantTLS:               true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeMDBAll(
				zeroKropathMDB, zeroKropathMDB,
				zeroMDBCfg, zeroMDBCfg,
				tc.localMDBCfgDefaults,
				tc.globalMDBCfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
			)
			if got.Defaults.TLSEnabled != tc.wantTLS {
				t.Errorf("defaults.tlsEnabled = %v, want %v", got.Defaults.TLSEnabled, tc.wantTLS)
			}
		})
	}
}

// TestMergeMemoryDBCascade_MandatoryIsolatedFromDefaults — mandatory fields must not
// bleed into defaults and vice versa.
func TestMergeMemoryDBCascade_MandatoryIsolatedFromDefaults(t *testing.T) {
	got := mergeMDBAll(
		cascade.MemoryDBKropathSection{TLSEnabled: true},                             // level 1 mandatory
		zeroKropathMDB,
		cascade.MemoryDBConfigSection{NodeType: "db.r7g.large", KMSKeyARN: "arn:x"}, // level 3 mandatory
		zeroMDBCfg,
		cascade.MemoryDBConfigSection{NodeType: "db.t4g.small", AutoMinorVersionUpgrade: true}, // level 6 defaults
		zeroMDBCfg,
		zeroKropathMDB,
		zeroKropathMDB,
	)

	if !got.Mandatory.TLSEnabled {
		t.Errorf("mandatory.tlsEnabled = false, want true")
	}
	if got.Mandatory.NodeType != "db.r7g.large" {
		t.Errorf("mandatory.nodeType = %q, want db.r7g.large", got.Mandatory.NodeType)
	}
	if got.Defaults.TLSEnabled {
		t.Errorf("defaults.tlsEnabled = true, must not bleed from mandatory")
	}
	if got.Defaults.NodeType != "db.t4g.small" {
		t.Errorf("defaults.nodeType = %q, want db.t4g.small", got.Defaults.NodeType)
	}
	if !got.Defaults.AutoMinorVersionUpgrade {
		t.Errorf("defaults.autoMinorVersionUpgrade = false, want true")
	}
	if got.Mandatory.AutoMinorVersionUpgrade {
		t.Errorf("mandatory.autoMinorVersionUpgrade = true, must not bleed from defaults")
	}
}

// TestMergeMemoryDBCascade_TagUnionMerge — tags from multiple tiers are union-merged;
// higher-priority (lower-level number) source wins on key conflict.
func TestMergeMemoryDBCascade_TagUnionMerge(t *testing.T) {
	got := mergeMDBAll(
		cascade.MemoryDBKropathSection{Tags: map[string]string{"owner": "platform", "cost-centre": "infra"}}, // level 1 wins
		zeroKropathMDB,
		cascade.MemoryDBConfigSection{Tags: map[string]string{"cost-centre": "payments", "env": "prod"}}, // level 3
		zeroMDBCfg,
		zeroMDBCfg,
		zeroMDBCfg,
		zeroKropathMDB,
		zeroKropathMDB,
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

// TestMergeMemoryDBCascade_SyncedLabelsUnionMerge — syncedLabels from MemoryDBConfig levels only;
// level 3 wins over level 4 on key conflict.
func TestMergeMemoryDBCascade_SyncedLabelsUnionMerge(t *testing.T) {
	got := mergeMDBAll(
		zeroKropathMDB,
		zeroKropathMDB,
		cascade.MemoryDBConfigSection{SyncedLabels: map[string]string{"team": "platform", "data-class": "public"}},    // level 3 wins
		cascade.MemoryDBConfigSection{SyncedLabels: map[string]string{"data-class": "internal", "region": "ap-se-2"}}, // level 4
		zeroMDBCfg,
		zeroMDBCfg,
		zeroKropathMDB,
		zeroKropathMDB,
	)

	if got.Mandatory.SyncedLabels["data-class"] != "public" {
		t.Errorf("syncedLabels: mandatory.syncedLabels[data-class] = %q, want public (level 3 wins)", got.Mandatory.SyncedLabels["data-class"])
	}
	if got.Mandatory.SyncedLabels["team"] != "platform" {
		t.Errorf("syncedLabels: mandatory.syncedLabels[team] = %q, want platform", got.Mandatory.SyncedLabels["team"])
	}
	if got.Mandatory.SyncedLabels["region"] != "ap-se-2" {
		t.Errorf("syncedLabels: mandatory.syncedLabels[region] = %q, want ap-se-2 (additive)", got.Mandatory.SyncedLabels["region"])
	}
}

// TestMergeMemoryDBCascade_AllowedNodeTypesPriorityOrder — level 1 > 2 > 3 > 4 for allowedNodeTypes.
func TestMergeMemoryDBCascade_AllowedNodeTypesPriorityOrder(t *testing.T) {
	cases := []struct {
		name                   string
		globalKropathMandatory cascade.MemoryDBKropathSection
		localKropathMandatory  cascade.MemoryDBKropathSection
		globalMDBCfgMandatory  cascade.MemoryDBConfigSection
		localMDBCfgMandatory   cascade.MemoryDBConfigSection
		wantFirst              string
	}{
		{
			name:                   "level1-wins",
			globalKropathMandatory: cascade.MemoryDBKropathSection{AllowedNodeTypes: []string{"db.r7g.large"}},
			localKropathMandatory:  cascade.MemoryDBKropathSection{AllowedNodeTypes: []string{"db.t4g.small"}},
			globalMDBCfgMandatory:  cascade.MemoryDBConfigSection{AllowedNodeTypes: []string{"db.r6g.large"}},
			localMDBCfgMandatory:   cascade.MemoryDBConfigSection{AllowedNodeTypes: []string{"db.t4g.medium"}},
			wantFirst:              "db.r7g.large",
		},
		{
			name:                   "level3-wins-when-1-2-absent",
			globalKropathMandatory: zeroKropathMDB,
			localKropathMandatory:  zeroKropathMDB,
			globalMDBCfgMandatory:  cascade.MemoryDBConfigSection{AllowedNodeTypes: []string{"db.r6g.large"}},
			localMDBCfgMandatory:   cascade.MemoryDBConfigSection{AllowedNodeTypes: []string{"db.t4g.medium"}},
			wantFirst:              "db.r6g.large",
		},
		{
			name:                   "level4-only",
			globalKropathMandatory: zeroKropathMDB,
			localKropathMandatory:  zeroKropathMDB,
			globalMDBCfgMandatory:  zeroMDBCfg,
			localMDBCfgMandatory:   cascade.MemoryDBConfigSection{AllowedNodeTypes: []string{"db.t4g.small"}},
			wantFirst:              "db.t4g.small",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeMDBAll(
				tc.globalKropathMandatory,
				tc.localKropathMandatory,
				tc.globalMDBCfgMandatory,
				tc.localMDBCfgMandatory,
				zeroMDBCfg, zeroMDBCfg,
				zeroKropathMDB, zeroKropathMDB,
			)
			if len(got.Mandatory.AllowedNodeTypes) == 0 || got.Mandatory.AllowedNodeTypes[0] != tc.wantFirst {
				t.Errorf("mandatory.allowedNodeTypes[0] = %q, want %q", func() string {
					if len(got.Mandatory.AllowedNodeTypes) > 0 {
						return got.Mandatory.AllowedNodeTypes[0]
					}
					return "<empty>"
				}(), tc.wantFirst)
			}
		})
	}
}

// TestMergeMemoryDBCascade_SnapshotRetentionDefaultsPriorityOrder — level 6 > 7 > 8 > 9.
func TestMergeMemoryDBCascade_SnapshotRetentionDefaultsPriorityOrder(t *testing.T) {
	cases := []struct {
		name                  string
		localMDBCfgDefaults   cascade.MemoryDBConfigSection
		globalMDBCfgDefaults  cascade.MemoryDBConfigSection
		localKropathDefaults  cascade.MemoryDBKropathSection
		globalKropathDefaults cascade.MemoryDBKropathSection
		wantRetention         int64
	}{
		{
			name:                  "level6-wins",
			localMDBCfgDefaults:   cascade.MemoryDBConfigSection{SnapshotRetentionLimit: 14},
			globalMDBCfgDefaults:  cascade.MemoryDBConfigSection{SnapshotRetentionLimit: 7},
			localKropathDefaults:  cascade.MemoryDBKropathSection{SnapshotRetentionLimit: 3},
			globalKropathDefaults: cascade.MemoryDBKropathSection{SnapshotRetentionLimit: 1},
			wantRetention:         14,
		},
		{
			name:                  "level7-wins-when-6-absent",
			localMDBCfgDefaults:   zeroMDBCfg,
			globalMDBCfgDefaults:  cascade.MemoryDBConfigSection{SnapshotRetentionLimit: 7},
			localKropathDefaults:  cascade.MemoryDBKropathSection{SnapshotRetentionLimit: 3},
			globalKropathDefaults: cascade.MemoryDBKropathSection{SnapshotRetentionLimit: 1},
			wantRetention:         7,
		},
		{
			name:                  "level8-wins-when-6-7-absent",
			localMDBCfgDefaults:   zeroMDBCfg,
			globalMDBCfgDefaults:  zeroMDBCfg,
			localKropathDefaults:  cascade.MemoryDBKropathSection{SnapshotRetentionLimit: 3},
			globalKropathDefaults: cascade.MemoryDBKropathSection{SnapshotRetentionLimit: 1},
			wantRetention:         3,
		},
		{
			name:                  "level9-fallback",
			localMDBCfgDefaults:   zeroMDBCfg,
			globalMDBCfgDefaults:  zeroMDBCfg,
			localKropathDefaults:  zeroKropathMDB,
			globalKropathDefaults: cascade.MemoryDBKropathSection{SnapshotRetentionLimit: 1},
			wantRetention:         1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeMDBAll(
				zeroKropathMDB, zeroKropathMDB,
				zeroMDBCfg, zeroMDBCfg,
				tc.localMDBCfgDefaults,
				tc.globalMDBCfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
			)
			if got.Defaults.SnapshotRetentionLimit != tc.wantRetention {
				t.Errorf("defaults.snapshotRetentionLimit = %d, want %d", got.Defaults.SnapshotRetentionLimit, tc.wantRetention)
			}
		})
	}
}

// TestValidateMemoryDBNodeType_EdgeCases — edge cases for cross-field validation.
func TestValidateMemoryDBNodeType_EdgeCases(t *testing.T) {
	cases := []struct {
		name      string
		mandatory cascade.EffectiveMemoryDBSection
		wantValid bool
	}{
		{
			name:      "both-empty",
			mandatory: cascade.EffectiveMemoryDBSection{},
			wantValid: true,
		},
		{
			name: "nodeType-set-no-allowedList",
			mandatory: cascade.EffectiveMemoryDBSection{
				NodeType:         "db.r7g.large",
				AllowedNodeTypes: nil,
			},
			wantValid: true,
		},
		{
			name: "allowedList-set-no-nodeType",
			mandatory: cascade.EffectiveMemoryDBSection{
				AllowedNodeTypes: []string{"db.t4g.small"},
			},
			wantValid: true,
		},
		{
			name: "exact-match",
			mandatory: cascade.EffectiveMemoryDBSection{
				NodeType:         "db.t4g.small",
				AllowedNodeTypes: []string{"db.t4g.small"},
			},
			wantValid: true,
		},
		{
			name: "not-in-list",
			mandatory: cascade.EffectiveMemoryDBSection{
				NodeType:         "db.r7g.large",
				AllowedNodeTypes: []string{"db.t4g.small"},
			},
			wantValid: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			valid, _ := cascade.ValidateMemoryDBNodeType(tc.mandatory)
			if valid != tc.wantValid {
				t.Errorf("ValidateMemoryDBNodeType() = %v, want %v", valid, tc.wantValid)
			}
		})
	}
}
