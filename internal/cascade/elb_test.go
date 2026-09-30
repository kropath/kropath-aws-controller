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

// zeroKropathELB is a zero-value ELBKropathSection (absent source).
var zeroKropathELB = cascade.ELBKropathSection{}

// zeroELBCfg is a zero-value ELBConfigSection (absent source).
var zeroELBCfg = cascade.ELBConfigSection{}

// mergeELBAll calls MergeELBCascade with all eight inputs.
func mergeELBAll(
	globalKropathMandatory,
	localKropathMandatory cascade.ELBKropathSection,
	globalELBCfgMandatory,
	localELBCfgMandatory,
	localELBCfgDefaults,
	globalELBCfgDefaults cascade.ELBConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.ELBKropathSection,
) cascade.EffectiveELBConfig {
	return cascade.MergeELBCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalELBCfgMandatory,
		localELBCfgMandatory,
		localELBCfgDefaults,
		globalELBCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeELBCascade_AC1 — globalKropathConfig.mandatory.elb.deletionProtection=true
// at level 1 propagates to effCfg.mandatory.deletionProtection (level 1 wins).
func TestMergeELBCascade_AC1(t *testing.T) {
	got := mergeELBAll(
		cascade.ELBKropathSection{DeletionProtection: true}, // level 1
		zeroKropathELB,
		zeroELBCfg,
		zeroELBCfg,
		zeroELBCfg,
		zeroELBCfg,
		zeroKropathELB,
		zeroKropathELB,
	)

	if !got.Mandatory.DeletionProtection {
		t.Errorf("AC-1: mandatory.deletionProtection = false, want true (level 1 wins)")
	}
	if got.Defaults.DeletionProtection {
		t.Errorf("AC-1: defaults.deletionProtection = true, must not bleed from mandatory")
	}
}

// TestMergeELBCascade_AC2 — globalAWSELBConfig.mandatory.deletionProtection=true
// at level 3 propagates when levels 1-2 are absent.
func TestMergeELBCascade_AC2(t *testing.T) {
	got := mergeELBAll(
		zeroKropathELB,
		zeroKropathELB,
		cascade.ELBConfigSection{DeletionProtection: true}, // level 3
		zeroELBCfg,
		zeroELBCfg,
		zeroELBCfg,
		zeroKropathELB,
		zeroKropathELB,
	)

	if !got.Mandatory.DeletionProtection {
		t.Errorf("AC-2: mandatory.deletionProtection = false, want true (level 3 wins when 1-2 absent)")
	}
}

// TestMergeELBCascade_AC3 — localAWSELBConfig.defaults.deletionProtection=true
// at level 6 propagates; mandatory must remain false.
func TestMergeELBCascade_AC3(t *testing.T) {
	got := mergeELBAll(
		zeroKropathELB,
		zeroKropathELB,
		zeroELBCfg,
		zeroELBCfg,
		cascade.ELBConfigSection{DeletionProtection: true}, // level 6
		zeroELBCfg,
		zeroKropathELB,
		zeroKropathELB,
	)

	if got.Mandatory.DeletionProtection {
		t.Errorf("AC-3: mandatory.deletionProtection = true, must not bleed from defaults")
	}
	if !got.Defaults.DeletionProtection {
		t.Errorf("AC-3: defaults.deletionProtection = false, want true (level 6)")
	}
}

// TestMergeELBCascade_AC4 — globalKropathConfig.mandatory.elb.accessLogsEnabled=true
// at level 1 propagates to effCfg.mandatory.accessLogsEnabled.
func TestMergeELBCascade_AC4(t *testing.T) {
	got := mergeELBAll(
		cascade.ELBKropathSection{AccessLogsEnabled: true}, // level 1
		zeroKropathELB,
		zeroELBCfg,
		zeroELBCfg,
		zeroELBCfg,
		zeroELBCfg,
		zeroKropathELB,
		zeroKropathELB,
	)

	if !got.Mandatory.AccessLogsEnabled {
		t.Errorf("AC-4: mandatory.accessLogsEnabled = false, want true (level 1 wins)")
	}
}

// TestMergeELBCascade_AC5 — globalKropathConfig.mandatory.elb.internalOnly=true
// at level 1 propagates to effCfg.mandatory.internalOnly.
func TestMergeELBCascade_AC5(t *testing.T) {
	got := mergeELBAll(
		cascade.ELBKropathSection{InternalOnly: true}, // level 1
		zeroKropathELB,
		zeroELBCfg,
		zeroELBCfg,
		zeroELBCfg,
		zeroELBCfg,
		zeroKropathELB,
		zeroKropathELB,
	)

	if !got.Mandatory.InternalOnly {
		t.Errorf("AC-5: mandatory.internalOnly = false, want true (level 1 wins)")
	}
}

// TestMergeELBCascade_AC6 — globalAWSELBConfig.mandatory.accessLogsS3Bucket="org-logs-bucket"
// at level 3 propagates. KropathConfig does not carry accessLogsS3Bucket.
func TestMergeELBCascade_AC6(t *testing.T) {
	got := mergeELBAll(
		zeroKropathELB,
		zeroKropathELB,
		cascade.ELBConfigSection{AccessLogsS3Bucket: "org-logs-bucket"}, // level 3
		zeroELBCfg,
		zeroELBCfg,
		zeroELBCfg,
		zeroKropathELB,
		zeroKropathELB,
	)

	if got.Mandatory.AccessLogsS3Bucket != "org-logs-bucket" {
		t.Errorf("AC-6: mandatory.accessLogsS3Bucket = %q, want org-logs-bucket (level 3)", got.Mandatory.AccessLogsS3Bucket)
	}
}

// TestMergeELBCascade_AC7 — globalAWSELBConfig.mandatory.crossZoneEnabled=true
// at level 3 propagates. KropathConfig does not carry crossZoneEnabled.
func TestMergeELBCascade_AC7(t *testing.T) {
	got := mergeELBAll(
		zeroKropathELB,
		zeroKropathELB,
		cascade.ELBConfigSection{CrossZoneEnabled: true}, // level 3
		zeroELBCfg,
		zeroELBCfg,
		zeroELBCfg,
		zeroKropathELB,
		zeroKropathELB,
	)

	if !got.Mandatory.CrossZoneEnabled {
		t.Errorf("AC-7: mandatory.crossZoneEnabled = false, want true (level 3)")
	}
}

// TestMergeELBCascade_AC8 — globalAWSELBConfig.mandatory.idleTimeoutSeconds=120
// at level 3 propagates. KropathConfig does not carry idleTimeoutSeconds.
func TestMergeELBCascade_AC8(t *testing.T) {
	got := mergeELBAll(
		zeroKropathELB,
		zeroKropathELB,
		cascade.ELBConfigSection{IdleTimeoutSeconds: 120}, // level 3
		zeroELBCfg,
		zeroELBCfg,
		zeroELBCfg,
		zeroKropathELB,
		zeroKropathELB,
	)

	if got.Mandatory.IdleTimeoutSeconds != 120 {
		t.Errorf("AC-8: mandatory.idleTimeoutSeconds = %d, want 120 (level 3)", got.Mandatory.IdleTimeoutSeconds)
	}
	if got.Defaults.IdleTimeoutSeconds != 0 {
		t.Errorf("AC-8: defaults.idleTimeoutSeconds = %d, must not bleed from mandatory", got.Defaults.IdleTimeoutSeconds)
	}
}

// TestMergeELBCascade_AC9 — globalAWSELBConfig.defaults.idleTimeoutSeconds=60 (level 7)
// wins over globalKropathDefaults (which has no idleTimeoutSeconds).
func TestMergeELBCascade_AC9(t *testing.T) {
	got := mergeELBAll(
		zeroKropathELB,
		zeroKropathELB,
		zeroELBCfg,
		zeroELBCfg,
		zeroELBCfg,
		cascade.ELBConfigSection{IdleTimeoutSeconds: 60}, // level 7
		zeroKropathELB,
		zeroKropathELB,
	)

	if got.Defaults.IdleTimeoutSeconds != 60 {
		t.Errorf("AC-9: defaults.idleTimeoutSeconds = %d, want 60 (level 7)", got.Defaults.IdleTimeoutSeconds)
	}
}

// TestMergeELBCascade_AC10 — globalAWSELBConfig.mandatory.sslPolicy wins at level 3.
// KropathConfig does not carry sslPolicy.
func TestMergeELBCascade_AC10(t *testing.T) {
	const policy = "ELBSecurityPolicy-TLS13-1-2-2021-06"
	got := mergeELBAll(
		zeroKropathELB,
		zeroKropathELB,
		cascade.ELBConfigSection{SslPolicy: policy}, // level 3
		zeroELBCfg,
		zeroELBCfg,
		zeroELBCfg,
		zeroKropathELB,
		zeroKropathELB,
	)

	if got.Mandatory.SslPolicy != policy {
		t.Errorf("AC-10: mandatory.sslPolicy = %q, want %q (level 3)", got.Mandatory.SslPolicy, policy)
	}
}

// TestMergeELBCascade_AC11 — globalAWSELBConfig.defaults.namingTemplate="{namespace}-{name}"
// at level 7 propagates. KropathConfig does not carry namingTemplate.
func TestMergeELBCascade_AC11(t *testing.T) {
	got := mergeELBAll(
		zeroKropathELB,
		zeroKropathELB,
		zeroELBCfg,
		zeroELBCfg,
		zeroELBCfg,
		cascade.ELBConfigSection{NamingTemplate: "{namespace}-{name}"}, // level 7
		zeroKropathELB,
		zeroKropathELB,
	)

	if got.Defaults.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("AC-11: defaults.namingTemplate = %q, want {namespace}-{name} (level 7)", got.Defaults.NamingTemplate)
	}
}

// TestMergeELBCascade_AC12 — KropathConfig.mandatory.tags and AWSELBConfig.mandatory.tags
// are union-merged; KropathConfig keys win on conflict.
func TestMergeELBCascade_AC12(t *testing.T) {
	got := mergeELBAll(
		cascade.ELBKropathSection{Tags: map[string]string{"cost-centre": "platform", "owner": "infra"}}, // level 1 (wins)
		zeroKropathELB,
		zeroELBCfg,
		cascade.ELBConfigSection{Tags: map[string]string{"cost-centre": "payments", "env": "prod"}}, // level 4
		zeroELBCfg,
		zeroELBCfg,
		zeroKropathELB,
		zeroKropathELB,
	)

	if got.Mandatory.Tags["cost-centre"] != "platform" {
		t.Errorf("AC-12: mandatory.tags[cost-centre] = %q, want platform (level 1 wins)", got.Mandatory.Tags["cost-centre"])
	}
	if got.Mandatory.Tags["owner"] != "infra" {
		t.Errorf("AC-12: mandatory.tags[owner] = %q, want infra", got.Mandatory.Tags["owner"])
	}
	if got.Mandatory.Tags["env"] != "prod" {
		t.Errorf("AC-12: mandatory.tags[env] = %q, want prod (additive from level 4)", got.Mandatory.Tags["env"])
	}
}

// TestMergeELBCascade_AC13 — AWSELBConfig.mandatory.syncedLabels propagates;
// SyncedLabels only from AWSELBConfig levels (no KropathConfig).
func TestMergeELBCascade_AC13(t *testing.T) {
	got := mergeELBAll(
		zeroKropathELB,
		zeroKropathELB,
		cascade.ELBConfigSection{SyncedLabels: map[string]string{"team": "platform", "data-class": "public"}}, // level 3
		cascade.ELBConfigSection{SyncedLabels: map[string]string{"data-class": "internal", "region": "ap"}},  // level 4
		zeroELBCfg,
		zeroELBCfg,
		zeroKropathELB,
		zeroKropathELB,
	)

	if got.Mandatory.SyncedLabels["data-class"] != "public" {
		t.Errorf("AC-13: mandatory.syncedLabels[data-class] = %q, want public (level 3 wins)", got.Mandatory.SyncedLabels["data-class"])
	}
	if got.Mandatory.SyncedLabels["team"] != "platform" {
		t.Errorf("AC-13: mandatory.syncedLabels[team] = %q, want platform", got.Mandatory.SyncedLabels["team"])
	}
	if got.Mandatory.SyncedLabels["region"] != "ap" {
		t.Errorf("AC-13: mandatory.syncedLabels[region] = %q, want ap (additive from level 4)", got.Mandatory.SyncedLabels["region"])
	}
}

// TestMergeELBCascade_AllAbsent — all sources zero: effective config is all-zero (permissive).
func TestMergeELBCascade_AllAbsent(t *testing.T) {
	got := mergeELBAll(
		zeroKropathELB, zeroKropathELB,
		zeroELBCfg, zeroELBCfg, zeroELBCfg, zeroELBCfg,
		zeroKropathELB, zeroKropathELB,
	)

	if got.Mandatory.DeletionProtection {
		t.Errorf("all-absent: mandatory.deletionProtection = true, want false")
	}
	if got.Mandatory.AccessLogsEnabled {
		t.Errorf("all-absent: mandatory.accessLogsEnabled = true, want false")
	}
	if got.Mandatory.AccessLogsS3Bucket != "" {
		t.Errorf("all-absent: mandatory.accessLogsS3Bucket = %q, want empty", got.Mandatory.AccessLogsS3Bucket)
	}
	if got.Mandatory.CrossZoneEnabled {
		t.Errorf("all-absent: mandatory.crossZoneEnabled = true, want false")
	}
	if got.Mandatory.InternalOnly {
		t.Errorf("all-absent: mandatory.internalOnly = true, want false")
	}
	if got.Mandatory.IdleTimeoutSeconds != 0 {
		t.Errorf("all-absent: mandatory.idleTimeoutSeconds = %d, want 0", got.Mandatory.IdleTimeoutSeconds)
	}
	if got.Mandatory.SslPolicy != "" {
		t.Errorf("all-absent: mandatory.sslPolicy = %q, want empty", got.Mandatory.SslPolicy)
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

// TestMergeELBCascade_MandatoryPriorityOrder — verifies mandatory priority order
// for deletionProtection (level 1 > 2 > 3 > 4).
func TestMergeELBCascade_MandatoryPriorityOrder(t *testing.T) {
	cases := []struct {
		name                   string
		globalKropathMandatory cascade.ELBKropathSection
		localKropathMandatory  cascade.ELBKropathSection
		globalELBCfgMandatory  cascade.ELBConfigSection
		localELBCfgMandatory   cascade.ELBConfigSection
		wantDeletionProtection bool
	}{
		{
			// level 1 sets true; all others false — level 1 wins
			name:                   "level1-wins",
			globalKropathMandatory: cascade.ELBKropathSection{DeletionProtection: true},
			wantDeletionProtection: true,
		},
		{
			// level 1 absent; level 2 sets true
			name:                  "level2-wins-when-1-absent",
			localKropathMandatory: cascade.ELBKropathSection{DeletionProtection: true},
			wantDeletionProtection: true,
		},
		{
			// levels 1-2 absent; level 3 sets true
			name:                  "level3-wins-when-1-2-absent",
			globalELBCfgMandatory: cascade.ELBConfigSection{DeletionProtection: true},
			wantDeletionProtection: true,
		},
		{
			// levels 1-3 absent; level 4 sets true
			name:                 "level4-wins-when-1-2-3-absent",
			localELBCfgMandatory: cascade.ELBConfigSection{DeletionProtection: true},
			wantDeletionProtection: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeELBAll(
				tc.globalKropathMandatory,
				tc.localKropathMandatory,
				tc.globalELBCfgMandatory,
				tc.localELBCfgMandatory,
				zeroELBCfg, zeroELBCfg,
				zeroKropathELB, zeroKropathELB,
			)
			if got.Mandatory.DeletionProtection != tc.wantDeletionProtection {
				t.Errorf("mandatory.deletionProtection = %v, want %v", got.Mandatory.DeletionProtection, tc.wantDeletionProtection)
			}
		})
	}
}

// TestMergeELBCascade_DefaultsPriorityOrder — verifies defaults priority order
// for deletionProtection (level 6 > 7 > 8 > 9).
func TestMergeELBCascade_DefaultsPriorityOrder(t *testing.T) {
	cases := []struct {
		name                  string
		localELBCfgDefaults   cascade.ELBConfigSection
		globalELBCfgDefaults  cascade.ELBConfigSection
		localKropathDefaults  cascade.ELBKropathSection
		globalKropathDefaults cascade.ELBKropathSection
		wantDeletion          bool
	}{
		{
			name:                "level6-wins",
			localELBCfgDefaults: cascade.ELBConfigSection{DeletionProtection: true},
			wantDeletion:        true,
		},
		{
			name:                 "level7-wins-when-6-absent",
			globalELBCfgDefaults: cascade.ELBConfigSection{DeletionProtection: true},
			wantDeletion:         true,
		},
		{
			name:                 "level8-wins-when-6-7-absent",
			localKropathDefaults: cascade.ELBKropathSection{DeletionProtection: true},
			wantDeletion:         true,
		},
		{
			name:                  "level9-fallback",
			globalKropathDefaults: cascade.ELBKropathSection{DeletionProtection: true},
			wantDeletion:          true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeELBAll(
				zeroKropathELB, zeroKropathELB,
				zeroELBCfg, zeroELBCfg,
				tc.localELBCfgDefaults,
				tc.globalELBCfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
			)
			if got.Defaults.DeletionProtection != tc.wantDeletion {
				t.Errorf("defaults.deletionProtection = %v, want %v", got.Defaults.DeletionProtection, tc.wantDeletion)
			}
		})
	}
}

// TestMergeELBCascade_MandatoryIsolatedFromDefaults — mandatory fields must not bleed
// into defaults and vice versa.
func TestMergeELBCascade_MandatoryIsolatedFromDefaults(t *testing.T) {
	got := mergeELBAll(
		cascade.ELBKropathSection{DeletionProtection: true, AccessLogsEnabled: true},
		zeroKropathELB,
		cascade.ELBConfigSection{IdleTimeoutSeconds: 120, SslPolicy: "ELBSecurityPolicy-TLS13-1-2-2021-06"},
		zeroELBCfg,
		cascade.ELBConfigSection{DeletionProtection: true, IdleTimeoutSeconds: 60, NamingTemplate: "{namespace}-{name}"}, // defaults level 6
		zeroELBCfg,
		zeroKropathELB,
		zeroKropathELB,
	)

	// Mandatory tier
	if !got.Mandatory.DeletionProtection {
		t.Errorf("mandatory.deletionProtection = false, want true")
	}
	if !got.Mandatory.AccessLogsEnabled {
		t.Errorf("mandatory.accessLogsEnabled = false, want true")
	}
	if got.Mandatory.IdleTimeoutSeconds != 120 {
		t.Errorf("mandatory.idleTimeoutSeconds = %d, want 120", got.Mandatory.IdleTimeoutSeconds)
	}
	if got.Mandatory.SslPolicy != "ELBSecurityPolicy-TLS13-1-2-2021-06" {
		t.Errorf("mandatory.sslPolicy = %q, want ELBSecurityPolicy-TLS13-1-2-2021-06", got.Mandatory.SslPolicy)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("mandatory.namingTemplate = %q, must not bleed from defaults", got.Mandatory.NamingTemplate)
	}

	// Defaults tier
	if got.Defaults.AccessLogsEnabled {
		t.Errorf("defaults.accessLogsEnabled = true, must not bleed from mandatory")
	}
	if got.Defaults.IdleTimeoutSeconds != 60 {
		t.Errorf("defaults.idleTimeoutSeconds = %d, want 60", got.Defaults.IdleTimeoutSeconds)
	}
	if got.Defaults.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("defaults.namingTemplate = %q, want {namespace}-{name}", got.Defaults.NamingTemplate)
	}
}

// TestMergeELBCascade_TagUnionMerge — tags from multiple tiers are union-merged;
// higher-priority (lower level number) source wins on key conflict.
func TestMergeELBCascade_TagUnionMerge(t *testing.T) {
	got := mergeELBAll(
		cascade.ELBKropathSection{Tags: map[string]string{"owner": "platform", "cost-centre": "infra"}}, // level 1 (wins)
		zeroKropathELB,
		cascade.ELBConfigSection{Tags: map[string]string{"cost-centre": "payments", "env": "prod"}}, // level 3
		zeroELBCfg,
		zeroELBCfg,
		zeroELBCfg,
		zeroKropathELB,
		zeroKropathELB,
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

// TestMergeELBCascade_DefaultsTagUnionMerge — defaults tags are union-merged;
// level 6 wins over level 9 on key conflict.
func TestMergeELBCascade_DefaultsTagUnionMerge(t *testing.T) {
	got := mergeELBAll(
		zeroKropathELB,
		zeroKropathELB,
		zeroELBCfg,
		zeroELBCfg,
		cascade.ELBConfigSection{Tags: map[string]string{"owner": "team", "cost-centre": "dev"}},   // level 6 (wins)
		zeroELBCfg,
		zeroKropathELB,
		cascade.ELBKropathSection{Tags: map[string]string{"cost-centre": "org", "region": "ap"}}, // level 9
	)

	if got.Defaults.Tags["cost-centre"] != "dev" {
		t.Errorf("defaults tag union: tags[cost-centre] = %q, want dev (level 6 wins)", got.Defaults.Tags["cost-centre"])
	}
	if got.Defaults.Tags["owner"] != "team" {
		t.Errorf("defaults tag union: tags[owner] = %q, want team", got.Defaults.Tags["owner"])
	}
	if got.Defaults.Tags["region"] != "ap" {
		t.Errorf("defaults tag union: tags[region] = %q, want ap (additive from level 9)", got.Defaults.Tags["region"])
	}
}

// TestMergeELBCascade_KropathOnlyFieldsAbsentFromELBConfigOnly —
// fields not in KropathConfig (crossZoneEnabled, idleTimeoutSeconds, sslPolicy,
// namingTemplate, accessLogsS3Bucket) must be zero when only KropathConfig sources are set.
func TestMergeELBCascade_KropathOnlyFieldsAbsentFromELBConfigOnly(t *testing.T) {
	got := mergeELBAll(
		cascade.ELBKropathSection{DeletionProtection: true, AccessLogsEnabled: true, InternalOnly: true},
		zeroKropathELB,
		zeroELBCfg,
		zeroELBCfg,
		zeroELBCfg,
		zeroELBCfg,
		zeroKropathELB,
		zeroKropathELB,
	)

	if got.Mandatory.CrossZoneEnabled {
		t.Errorf("mandatory.crossZoneEnabled must be zero when no AWSELBConfig sets it")
	}
	if got.Mandatory.IdleTimeoutSeconds != 0 {
		t.Errorf("mandatory.idleTimeoutSeconds = %d, must be zero when no AWSELBConfig sets it", got.Mandatory.IdleTimeoutSeconds)
	}
	if got.Mandatory.SslPolicy != "" {
		t.Errorf("mandatory.sslPolicy = %q, must be empty when no AWSELBConfig sets it", got.Mandatory.SslPolicy)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("mandatory.namingTemplate = %q, must be empty when no AWSELBConfig sets it", got.Mandatory.NamingTemplate)
	}
	if got.Mandatory.AccessLogsS3Bucket != "" {
		t.Errorf("mandatory.accessLogsS3Bucket = %q, must be empty when no AWSELBConfig sets it", got.Mandatory.AccessLogsS3Bucket)
	}
}
