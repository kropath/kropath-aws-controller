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

// zeroKropathEC is a zero-value ElastiCacheKropathSection (absent source).
var zeroKropathEC = cascade.ElastiCacheKropathSection{}

// zeroECCfg is a zero-value ElastiCacheConfigSection (absent source).
var zeroECCfg = cascade.ElastiCacheConfigSection{}

// mergeECAll calls MergeElastiCacheCascade with all eight inputs.
func mergeECAll(
	globalKropathMandatory,
	localKropathMandatory cascade.ElastiCacheKropathSection,
	globalECCfgMandatory,
	localECCfgMandatory,
	localECCfgDefaults,
	globalECCfgDefaults cascade.ElastiCacheConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.ElastiCacheKropathSection,
) cascade.EffectiveElastiCacheConfig {
	return cascade.MergeElastiCacheCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalECCfgMandatory,
		localECCfgMandatory,
		localECCfgDefaults,
		globalECCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeElastiCacheCascade_AC1 — KropathConfig(global).mandatory.elasticache.atRestEncryptionEnabled=true
// (level 1) wins over ElastiCacheConfig.mandatory.atRestEncryptionEnabled=false (levels 3-4).
func TestMergeElastiCacheCascade_AC1(t *testing.T) {
	got := mergeECAll(
		cascade.ElastiCacheKropathSection{AtRestEncryptionEnabled: true}, // level 1
		zeroKropathEC,
		cascade.ElastiCacheConfigSection{AtRestEncryptionEnabled: false}, // level 3
		cascade.ElastiCacheConfigSection{AtRestEncryptionEnabled: false}, // level 4
		zeroECCfg,
		zeroECCfg,
		zeroKropathEC,
		zeroKropathEC,
	)

	if !got.Mandatory.AtRestEncryptionEnabled {
		t.Errorf("AC-1: mandatory.atRestEncryptionEnabled = false, want true (level 1 wins)")
	}
	if got.Defaults.AtRestEncryptionEnabled {
		t.Errorf("AC-1: defaults.atRestEncryptionEnabled = true, must not bleed from mandatory")
	}
}

// TestMergeElastiCacheCascade_AC2 — KropathConfig(global).mandatory.elasticache.atRestEncryptionEnabled=false
// (level 1 absent); ElastiCacheConfig(global).mandatory.atRestEncryptionEnabled=true (level 3) wins.
func TestMergeElastiCacheCascade_AC2(t *testing.T) {
	got := mergeECAll(
		zeroKropathEC,
		zeroKropathEC,
		cascade.ElastiCacheConfigSection{AtRestEncryptionEnabled: true}, // level 3
		zeroECCfg,
		zeroECCfg,
		zeroECCfg,
		zeroKropathEC,
		zeroKropathEC,
	)

	if !got.Mandatory.AtRestEncryptionEnabled {
		t.Errorf("AC-2: mandatory.atRestEncryptionEnabled = false, want true (level 3 wins when 1-2 false)")
	}
}

// TestMergeElastiCacheCascade_AC3 — all mandatory tiers false;
// ElastiCacheConfig(local).defaults.atRestEncryptionEnabled=true (level 6) propagates.
func TestMergeElastiCacheCascade_AC3(t *testing.T) {
	got := mergeECAll(
		zeroKropathEC,
		zeroKropathEC,
		zeroECCfg,
		zeroECCfg,
		cascade.ElastiCacheConfigSection{AtRestEncryptionEnabled: true}, // level 6
		zeroECCfg,
		zeroKropathEC,
		zeroKropathEC,
	)

	if got.Mandatory.AtRestEncryptionEnabled {
		t.Errorf("AC-3: mandatory.atRestEncryptionEnabled = true, want false (defaults must not bleed)")
	}
	if !got.Defaults.AtRestEncryptionEnabled {
		t.Errorf("AC-3: defaults.atRestEncryptionEnabled = false, want true (level 6)")
	}
}

// TestMergeElastiCacheCascade_AC4 — KropathConfig(global).mandatory.elasticache.transitEncryptionEnabled=true
// (level 1) wins.
func TestMergeElastiCacheCascade_AC4(t *testing.T) {
	got := mergeECAll(
		cascade.ElastiCacheKropathSection{TransitEncryptionEnabled: true}, // level 1
		zeroKropathEC,
		cascade.ElastiCacheConfigSection{TransitEncryptionEnabled: false}, // level 3
		zeroECCfg,
		zeroECCfg,
		zeroECCfg,
		zeroKropathEC,
		zeroKropathEC,
	)

	if !got.Mandatory.TransitEncryptionEnabled {
		t.Errorf("AC-4: mandatory.transitEncryptionEnabled = false, want true (level 1 wins)")
	}
}

// TestMergeElastiCacheCascade_AC5 — ElastiCacheConfig(global).defaults.transitEncryptionEnabled=true
// (level 7) wins over KropathConfig.defaults.elasticache.transitEncryptionEnabled=false (level 9).
func TestMergeElastiCacheCascade_AC5(t *testing.T) {
	got := mergeECAll(
		zeroKropathEC,
		zeroKropathEC,
		zeroECCfg,
		zeroECCfg,
		zeroECCfg,
		cascade.ElastiCacheConfigSection{TransitEncryptionEnabled: true},    // level 7
		zeroKropathEC,
		cascade.ElastiCacheKropathSection{TransitEncryptionEnabled: false}, // level 9
	)

	if !got.Defaults.TransitEncryptionEnabled {
		t.Errorf("AC-5: defaults.transitEncryptionEnabled = false, want true (level 7 wins over level 9)")
	}
}

// TestMergeElastiCacheCascade_AC6 — ElastiCacheConfig(global).mandatory.automaticFailoverEnabled=true
// (level 3; no KropathConfig.elasticache levels exist for this field).
func TestMergeElastiCacheCascade_AC6(t *testing.T) {
	got := mergeECAll(
		zeroKropathEC,
		zeroKropathEC,
		cascade.ElastiCacheConfigSection{AutomaticFailoverEnabled: true}, // level 3
		zeroECCfg,
		zeroECCfg,
		zeroECCfg,
		zeroKropathEC,
		zeroKropathEC,
	)

	if !got.Mandatory.AutomaticFailoverEnabled {
		t.Errorf("AC-6: mandatory.automaticFailoverEnabled = false, want true (level 3)")
	}
}

// TestMergeElastiCacheCascade_AC7 — ElastiCacheConfig(global).mandatory.multiAZEnabled=true (level 3).
func TestMergeElastiCacheCascade_AC7(t *testing.T) {
	got := mergeECAll(
		zeroKropathEC,
		zeroKropathEC,
		cascade.ElastiCacheConfigSection{MultiAZEnabled: true}, // level 3
		zeroECCfg,
		zeroECCfg,
		zeroECCfg,
		zeroKropathEC,
		zeroKropathEC,
	)

	if !got.Mandatory.MultiAZEnabled {
		t.Errorf("AC-7: mandatory.multiAZEnabled = false, want true (level 3)")
	}
}

// TestMergeElastiCacheCascade_AC8 — ElastiCacheConfig(global).mandatory.snapshotRetentionLimit=7 (level 3).
func TestMergeElastiCacheCascade_AC8(t *testing.T) {
	got := mergeECAll(
		zeroKropathEC,
		zeroKropathEC,
		cascade.ElastiCacheConfigSection{SnapshotRetentionLimit: 7}, // level 3
		zeroECCfg,
		zeroECCfg,
		zeroECCfg,
		zeroKropathEC,
		zeroKropathEC,
	)

	if got.Mandatory.SnapshotRetentionLimit != 7 {
		t.Errorf("AC-8: mandatory.snapshotRetentionLimit = %d, want 7 (level 3)", got.Mandatory.SnapshotRetentionLimit)
	}
}

// TestMergeElastiCacheCascade_AC9 — ElastiCacheConfig(global).mandatory.engine="redis" (level 3).
func TestMergeElastiCacheCascade_AC9(t *testing.T) {
	got := mergeECAll(
		zeroKropathEC,
		zeroKropathEC,
		cascade.ElastiCacheConfigSection{Engine: "redis"}, // level 3
		zeroECCfg,
		zeroECCfg,
		zeroECCfg,
		zeroKropathEC,
		zeroKropathEC,
	)

	if got.Mandatory.Engine != "redis" {
		t.Errorf("AC-9: mandatory.engine = %q, want redis (level 3)", got.Mandatory.Engine)
	}
}

// TestMergeElastiCacheCascade_AC10 — ElastiCacheConfig(global).mandatory.blockNoPasswordUsers=true (level 3).
func TestMergeElastiCacheCascade_AC10(t *testing.T) {
	got := mergeECAll(
		zeroKropathEC,
		zeroKropathEC,
		cascade.ElastiCacheConfigSection{BlockNoPasswordUsers: true}, // level 3
		zeroECCfg,
		zeroECCfg,
		zeroECCfg,
		zeroKropathEC,
		zeroKropathEC,
	)

	if !got.Mandatory.BlockNoPasswordUsers {
		t.Errorf("AC-10: mandatory.blockNoPasswordUsers = false, want true (level 3)")
	}
}

// TestMergeElastiCacheCascade_AC11 — ElastiCacheConfig(global).defaults.namingTemplate="{namespace}-{name}"
// (level 7); no KropathConfig.elasticache.namingTemplate field exists.
func TestMergeElastiCacheCascade_AC11(t *testing.T) {
	got := mergeECAll(
		zeroKropathEC,
		zeroKropathEC,
		zeroECCfg,
		zeroECCfg,
		zeroECCfg,
		cascade.ElastiCacheConfigSection{NamingTemplate: "{namespace}-{name}"}, // level 7
		zeroKropathEC,
		zeroKropathEC,
	)

	if got.Defaults.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("AC-11: defaults.namingTemplate = %q, want {namespace}-{name} (level 7)", got.Defaults.NamingTemplate)
	}
}

// TestMergeElastiCacheCascade_AC12 — KropathConfig.mandatory.tags and ElastiCacheConfig.mandatory.tags
// are union-merged into effCfg.mandatory.tags; syncedLabels propagates from ElastiCacheConfig.mandatory.
func TestMergeElastiCacheCascade_AC12(t *testing.T) {
	got := mergeECAll(
		cascade.ElastiCacheKropathSection{Tags: map[string]string{"cost-centre": "infra"}}, // level 1
		zeroKropathEC,
		cascade.ElastiCacheConfigSection{
			Tags:         map[string]string{"cache-tier": "shared"},
			SyncedLabels: map[string]string{"data-class": "internal"},
		}, // level 3
		zeroECCfg,
		zeroECCfg,
		zeroECCfg,
		zeroKropathEC,
		zeroKropathEC,
	)

	if got.Mandatory.Tags["cost-centre"] != "infra" {
		t.Errorf("AC-12: mandatory.tags[cost-centre] = %q, want infra", got.Mandatory.Tags["cost-centre"])
	}
	if got.Mandatory.Tags["cache-tier"] != "shared" {
		t.Errorf("AC-12: mandatory.tags[cache-tier] = %q, want shared", got.Mandatory.Tags["cache-tier"])
	}
	if got.Mandatory.SyncedLabels["data-class"] != "internal" {
		t.Errorf("AC-12: mandatory.syncedLabels[data-class] = %q, want internal", got.Mandatory.SyncedLabels["data-class"])
	}
}

// TestMergeElastiCacheCascade_AC13 — provider identity (aws.region, aws.accountId) is tested
// at the reconciler level; at the cascade merge level, verify all-absent results are zero.
// The actual AC-13 assertion is in the Chainsaw test and reconciler.
func TestMergeElastiCacheCascade_AC13_AllAbsent(t *testing.T) {
	got := mergeECAll(
		zeroKropathEC, zeroKropathEC,
		zeroECCfg, zeroECCfg, zeroECCfg, zeroECCfg,
		zeroKropathEC, zeroKropathEC,
	)

	if got.Mandatory.AtRestEncryptionEnabled {
		t.Errorf("all-absent: mandatory.atRestEncryptionEnabled = true, want false")
	}
	if got.Mandatory.TransitEncryptionEnabled {
		t.Errorf("all-absent: mandatory.transitEncryptionEnabled = true, want false")
	}
	if got.Mandatory.AutomaticFailoverEnabled {
		t.Errorf("all-absent: mandatory.automaticFailoverEnabled = true, want false")
	}
	if got.Mandatory.MultiAZEnabled {
		t.Errorf("all-absent: mandatory.multiAZEnabled = true, want false")
	}
	if got.Mandatory.Engine != "" {
		t.Errorf("all-absent: mandatory.engine = %q, want empty", got.Mandatory.Engine)
	}
	if got.Mandatory.BlockNoPasswordUsers {
		t.Errorf("all-absent: mandatory.blockNoPasswordUsers = true, want false")
	}
	if got.Mandatory.SnapshotRetentionLimit != 0 {
		t.Errorf("all-absent: mandatory.snapshotRetentionLimit = %d, want 0", got.Mandatory.SnapshotRetentionLimit)
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

// TestMergeElastiCacheCascade_AtRestMandatoryPriorityOrder — level 1 > 2 > 3 > 4
// for atRestEncryptionEnabled.
func TestMergeElastiCacheCascade_AtRestMandatoryPriorityOrder(t *testing.T) {
	cases := []struct {
		name                   string
		globalKropathMandatory cascade.ElastiCacheKropathSection
		localKropathMandatory  cascade.ElastiCacheKropathSection
		globalECCfgMandatory   cascade.ElastiCacheConfigSection
		localECCfgMandatory    cascade.ElastiCacheConfigSection
		wantAtRest             bool
	}{
		{
			name:                   "level1-wins",
			globalKropathMandatory: cascade.ElastiCacheKropathSection{AtRestEncryptionEnabled: true},
			localKropathMandatory:  cascade.ElastiCacheKropathSection{AtRestEncryptionEnabled: false},
			globalECCfgMandatory:   cascade.ElastiCacheConfigSection{AtRestEncryptionEnabled: false},
			localECCfgMandatory:    cascade.ElastiCacheConfigSection{AtRestEncryptionEnabled: false},
			wantAtRest:             true,
		},
		{
			name:                   "level3-wins-when-1-2-false",
			globalKropathMandatory: zeroKropathEC,
			localKropathMandatory:  zeroKropathEC,
			globalECCfgMandatory:   cascade.ElastiCacheConfigSection{AtRestEncryptionEnabled: true},
			localECCfgMandatory:    cascade.ElastiCacheConfigSection{AtRestEncryptionEnabled: false},
			wantAtRest:             true,
		},
		{
			name:                   "level4-only",
			globalKropathMandatory: zeroKropathEC,
			localKropathMandatory:  zeroKropathEC,
			globalECCfgMandatory:   zeroECCfg,
			localECCfgMandatory:    cascade.ElastiCacheConfigSection{AtRestEncryptionEnabled: true},
			wantAtRest:             true,
		},
		{
			name:                   "all-false",
			globalKropathMandatory: zeroKropathEC,
			localKropathMandatory:  zeroKropathEC,
			globalECCfgMandatory:   zeroECCfg,
			localECCfgMandatory:    zeroECCfg,
			wantAtRest:             false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeECAll(
				tc.globalKropathMandatory,
				tc.localKropathMandatory,
				tc.globalECCfgMandatory,
				tc.localECCfgMandatory,
				zeroECCfg, zeroECCfg,
				zeroKropathEC, zeroKropathEC,
			)
			if got.Mandatory.AtRestEncryptionEnabled != tc.wantAtRest {
				t.Errorf("mandatory.atRestEncryptionEnabled = %v, want %v", got.Mandatory.AtRestEncryptionEnabled, tc.wantAtRest)
			}
		})
	}
}

// TestMergeElastiCacheCascade_AtRestDefaultsPriorityOrder — level 6 > 7 > 8 > 9.
func TestMergeElastiCacheCascade_AtRestDefaultsPriorityOrder(t *testing.T) {
	cases := []struct {
		name                  string
		localECCfgDefaults    cascade.ElastiCacheConfigSection
		globalECCfgDefaults   cascade.ElastiCacheConfigSection
		localKropathDefaults  cascade.ElastiCacheKropathSection
		globalKropathDefaults cascade.ElastiCacheKropathSection
		wantAtRest            bool
	}{
		{
			name:                  "level6-wins",
			localECCfgDefaults:    cascade.ElastiCacheConfigSection{AtRestEncryptionEnabled: true},
			globalECCfgDefaults:   cascade.ElastiCacheConfigSection{AtRestEncryptionEnabled: false},
			localKropathDefaults:  cascade.ElastiCacheKropathSection{AtRestEncryptionEnabled: false},
			globalKropathDefaults: cascade.ElastiCacheKropathSection{AtRestEncryptionEnabled: false},
			wantAtRest:            true,
		},
		{
			name:                  "level7-wins-when-6-absent",
			localECCfgDefaults:    zeroECCfg,
			globalECCfgDefaults:   cascade.ElastiCacheConfigSection{AtRestEncryptionEnabled: true},
			localKropathDefaults:  cascade.ElastiCacheKropathSection{AtRestEncryptionEnabled: false},
			globalKropathDefaults: cascade.ElastiCacheKropathSection{AtRestEncryptionEnabled: false},
			wantAtRest:            true,
		},
		{
			name:                  "level8-wins-when-6-7-absent",
			localECCfgDefaults:    zeroECCfg,
			globalECCfgDefaults:   zeroECCfg,
			localKropathDefaults:  cascade.ElastiCacheKropathSection{AtRestEncryptionEnabled: true},
			globalKropathDefaults: cascade.ElastiCacheKropathSection{AtRestEncryptionEnabled: false},
			wantAtRest:            true,
		},
		{
			name:                  "level9-fallback",
			localECCfgDefaults:    zeroECCfg,
			globalECCfgDefaults:   zeroECCfg,
			localKropathDefaults:  zeroKropathEC,
			globalKropathDefaults: cascade.ElastiCacheKropathSection{AtRestEncryptionEnabled: true},
			wantAtRest:            true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeECAll(
				zeroKropathEC, zeroKropathEC,
				zeroECCfg, zeroECCfg,
				tc.localECCfgDefaults,
				tc.globalECCfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
			)
			if got.Defaults.AtRestEncryptionEnabled != tc.wantAtRest {
				t.Errorf("defaults.atRestEncryptionEnabled = %v, want %v", got.Defaults.AtRestEncryptionEnabled, tc.wantAtRest)
			}
		})
	}
}

// TestMergeElastiCacheCascade_MandatoryIsolatedFromDefaults — mandatory fields must not
// bleed into defaults and vice versa.
func TestMergeElastiCacheCascade_MandatoryIsolatedFromDefaults(t *testing.T) {
	got := mergeECAll(
		cascade.ElastiCacheKropathSection{AtRestEncryptionEnabled: true}, // level 1 mandatory
		zeroKropathEC,
		cascade.ElastiCacheConfigSection{Engine: "redis"}, // level 3 mandatory
		zeroECCfg,
		cascade.ElastiCacheConfigSection{Engine: "valkey", AutomaticFailoverEnabled: true}, // level 6 defaults
		zeroECCfg,
		zeroKropathEC,
		zeroKropathEC,
	)

	if !got.Mandatory.AtRestEncryptionEnabled {
		t.Errorf("mandatory.atRestEncryptionEnabled = false, want true")
	}
	if got.Mandatory.Engine != "redis" {
		t.Errorf("mandatory.engine = %q, want redis", got.Mandatory.Engine)
	}
	if got.Defaults.AtRestEncryptionEnabled {
		t.Errorf("defaults.atRestEncryptionEnabled = true, must not bleed from mandatory")
	}
	if got.Defaults.Engine != "valkey" {
		t.Errorf("defaults.engine = %q, want valkey", got.Defaults.Engine)
	}
	if !got.Defaults.AutomaticFailoverEnabled {
		t.Errorf("defaults.automaticFailoverEnabled = false, want true")
	}
	if got.Mandatory.AutomaticFailoverEnabled {
		t.Errorf("mandatory.automaticFailoverEnabled = true, must not bleed from defaults")
	}
}

// TestMergeElastiCacheCascade_TagUnionMerge — tags from multiple tiers are union-merged;
// higher-priority (lower-level number) source wins on key conflict.
func TestMergeElastiCacheCascade_TagUnionMerge(t *testing.T) {
	got := mergeECAll(
		cascade.ElastiCacheKropathSection{Tags: map[string]string{"owner": "platform", "cost-centre": "infra"}}, // level 1 wins
		zeroKropathEC,
		cascade.ElastiCacheConfigSection{Tags: map[string]string{"cost-centre": "payments", "env": "prod"}}, // level 3
		zeroECCfg,
		zeroECCfg,
		zeroECCfg,
		zeroKropathEC,
		zeroKropathEC,
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

// TestMergeElastiCacheCascade_SyncedLabelsUnionMerge — syncedLabels from ElastiCacheConfig levels
// only; level 3 wins over level 4 on key conflict.
func TestMergeElastiCacheCascade_SyncedLabelsUnionMerge(t *testing.T) {
	got := mergeECAll(
		zeroKropathEC,
		zeroKropathEC,
		cascade.ElastiCacheConfigSection{SyncedLabels: map[string]string{"team": "platform", "data-class": "public"}},    // level 3 wins
		cascade.ElastiCacheConfigSection{SyncedLabels: map[string]string{"data-class": "internal", "region": "ap-se-2"}}, // level 4
		zeroECCfg,
		zeroECCfg,
		zeroKropathEC,
		zeroKropathEC,
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

// TestMergeElastiCacheCascade_NoKropathLevelsForNonEncryptionFields — fields with no KropathConfig
// equivalents are correctly bounded to levels 3-4 (mandatory) and 6-7 (defaults).
func TestMergeElastiCacheCascade_NoKropathLevelsForNonEncryptionFields(t *testing.T) {
	// Set KropathSection fields to ensure non-encryption fields don't accidentally
	// pick up from a KropathConfig-level source.
	got := mergeECAll(
		cascade.ElastiCacheKropathSection{AtRestEncryptionEnabled: true, TransitEncryptionEnabled: true}, // level 1
		zeroKropathEC,
		cascade.ElastiCacheConfigSection{
			Engine:                 "redis",
			SnapshotRetentionLimit: 7,
			BlockNoPasswordUsers:   true,
			AutomaticFailoverEnabled: true,
			MultiAZEnabled:         true,
		}, // level 3
		zeroECCfg,
		zeroECCfg,
		zeroECCfg,
		zeroKropathEC,
		zeroKropathEC,
	)

	if got.Mandatory.Engine != "redis" {
		t.Errorf("mandatory.engine = %q, want redis", got.Mandatory.Engine)
	}
	if got.Mandatory.SnapshotRetentionLimit != 7 {
		t.Errorf("mandatory.snapshotRetentionLimit = %d, want 7", got.Mandatory.SnapshotRetentionLimit)
	}
	if !got.Mandatory.BlockNoPasswordUsers {
		t.Errorf("mandatory.blockNoPasswordUsers = false, want true")
	}
	if !got.Mandatory.AutomaticFailoverEnabled {
		t.Errorf("mandatory.automaticFailoverEnabled = false, want true")
	}
	if !got.Mandatory.MultiAZEnabled {
		t.Errorf("mandatory.multiAZEnabled = false, want true")
	}
	// atRest and transit also propagate from level 1 KropathConfig.
	if !got.Mandatory.AtRestEncryptionEnabled {
		t.Errorf("mandatory.atRestEncryptionEnabled = false, want true (level 1)")
	}
}
