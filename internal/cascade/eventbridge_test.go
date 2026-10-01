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

// zeroKropathEB is a zero-value EventBridgeKropathSection (absent source).
var zeroKropathEB = cascade.EventBridgeKropathSection{}

// zeroEBCfg is a zero-value EventBridgeConfigSection (absent source).
var zeroEBCfg = cascade.EventBridgeConfigSection{}

// mergeEBAll calls MergeEventBridgeCascade with all eight inputs.
func mergeEBAll(
	globalKropathMandatory,
	localKropathMandatory cascade.EventBridgeKropathSection,
	globalEBCfgMandatory,
	localEBCfgMandatory,
	localEBCfgDefaults,
	globalEBCfgDefaults cascade.EventBridgeConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.EventBridgeKropathSection,
) cascade.EffectiveEventBridgeConfig {
	return cascade.MergeEventBridgeCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalEBCfgMandatory,
		localEBCfgMandatory,
		localEBCfgDefaults,
		globalEBCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeEventBridgeCascade_AC1 — globalKropathConfig.mandatory.eventbridge.archiveRetentionDays
// at level 1 propagates to effCfg.mandatory.archiveRetentionDays (level 1 wins).
func TestMergeEventBridgeCascade_AC1(t *testing.T) {
	got := mergeEBAll(
		cascade.EventBridgeKropathSection{ArchiveRetentionDays: 90}, // level 1
		zeroKropathEB,
		zeroEBCfg,
		zeroEBCfg,
		zeroEBCfg,
		zeroEBCfg,
		zeroKropathEB,
		zeroKropathEB,
	)

	if got.Mandatory.ArchiveRetentionDays != 90 {
		t.Errorf("AC-1: mandatory.archiveRetentionDays = %d, want 90", got.Mandatory.ArchiveRetentionDays)
	}
	if got.Defaults.ArchiveRetentionDays != 0 {
		t.Errorf("AC-1: defaults.archiveRetentionDays = %d, must not bleed from mandatory", got.Defaults.ArchiveRetentionDays)
	}
}

// TestMergeEventBridgeCascade_AC2 — globalEventBridgeConfig.mandatory.archiveRetentionDays at
// level 3 propagates when levels 1-2 are zero.
func TestMergeEventBridgeCascade_AC2(t *testing.T) {
	got := mergeEBAll(
		zeroKropathEB,
		zeroKropathEB,
		cascade.EventBridgeConfigSection{ArchiveRetentionDays: 30}, // level 3
		zeroEBCfg,
		zeroEBCfg,
		zeroEBCfg,
		zeroKropathEB,
		zeroKropathEB,
	)

	if got.Mandatory.ArchiveRetentionDays != 30 {
		t.Errorf("AC-2: mandatory.archiveRetentionDays = %d, want 30 (level 3 wins when 1-2 absent)", got.Mandatory.ArchiveRetentionDays)
	}
}

// TestMergeEventBridgeCascade_AC3 — level 1 wins over level 3 for archiveRetentionDays.
func TestMergeEventBridgeCascade_AC3(t *testing.T) {
	got := mergeEBAll(
		cascade.EventBridgeKropathSection{ArchiveRetentionDays: 90}, // level 1
		zeroKropathEB,
		cascade.EventBridgeConfigSection{ArchiveRetentionDays: 30}, // level 3
		zeroEBCfg,
		zeroEBCfg,
		zeroEBCfg,
		zeroKropathEB,
		zeroKropathEB,
	)

	if got.Mandatory.ArchiveRetentionDays != 90 {
		t.Errorf("AC-3: mandatory.archiveRetentionDays = %d, want 90 (level 1 beats level 3)", got.Mandatory.ArchiveRetentionDays)
	}
}

// TestMergeEventBridgeCascade_AC4 — localEventBridgeConfig.defaults.archiveRetentionDays at
// level 6 propagates to effCfg.defaults.archiveRetentionDays; mandatory stays zero.
func TestMergeEventBridgeCascade_AC4(t *testing.T) {
	got := mergeEBAll(
		zeroKropathEB,
		zeroKropathEB,
		zeroEBCfg,
		zeroEBCfg,
		cascade.EventBridgeConfigSection{ArchiveRetentionDays: 14}, // level 6
		zeroEBCfg,
		zeroKropathEB,
		zeroKropathEB,
	)

	if got.Defaults.ArchiveRetentionDays != 14 {
		t.Errorf("AC-4: defaults.archiveRetentionDays = %d, want 14", got.Defaults.ArchiveRetentionDays)
	}
	if got.Mandatory.ArchiveRetentionDays != 0 {
		t.Errorf("AC-4: mandatory.archiveRetentionDays = %d, must not bleed from defaults", got.Mandatory.ArchiveRetentionDays)
	}
}

// TestMergeEventBridgeCascade_AC5 — level 6 wins over level 8 for defaults.archiveRetentionDays.
func TestMergeEventBridgeCascade_AC5(t *testing.T) {
	got := mergeEBAll(
		zeroKropathEB,
		zeroKropathEB,
		zeroEBCfg,
		zeroEBCfg,
		cascade.EventBridgeConfigSection{ArchiveRetentionDays: 14}, // level 6
		zeroEBCfg,
		cascade.EventBridgeKropathSection{ArchiveRetentionDays: 7}, // level 8
		zeroKropathEB,
	)

	if got.Defaults.ArchiveRetentionDays != 14 {
		t.Errorf("AC-5: defaults.archiveRetentionDays = %d, want 14 (level 6 beats level 8)", got.Defaults.ArchiveRetentionDays)
	}
}

// TestMergeEventBridgeCascade_AC6 — globalKropathConfig.defaults.archiveRetentionDays at
// level 9 propagates when levels 6-8 are zero.
func TestMergeEventBridgeCascade_AC6(t *testing.T) {
	got := mergeEBAll(
		zeroKropathEB,
		zeroKropathEB,
		zeroEBCfg,
		zeroEBCfg,
		zeroEBCfg,
		zeroEBCfg,
		zeroKropathEB,
		cascade.EventBridgeKropathSection{ArchiveRetentionDays: 7}, // level 9
	)

	if got.Defaults.ArchiveRetentionDays != 7 {
		t.Errorf("AC-6: defaults.archiveRetentionDays = %d, want 7 (level 9 when 6-8 absent)", got.Defaults.ArchiveRetentionDays)
	}
}

// TestMergeEventBridgeCascade_AC7 — namingTemplate at level 3 propagates to mandatory only.
func TestMergeEventBridgeCascade_AC7(t *testing.T) {
	const tmpl = "corp-{namespace}-{name}"
	got := mergeEBAll(
		zeroKropathEB,
		zeroKropathEB,
		cascade.EventBridgeConfigSection{NamingTemplate: tmpl}, // level 3
		zeroEBCfg,
		zeroEBCfg,
		zeroEBCfg,
		zeroKropathEB,
		zeroKropathEB,
	)

	if got.Mandatory.NamingTemplate != tmpl {
		t.Errorf("AC-7: mandatory.namingTemplate = %q, want %q", got.Mandatory.NamingTemplate, tmpl)
	}
	if got.Defaults.NamingTemplate != "" {
		t.Errorf("AC-7: defaults.namingTemplate = %q, must not bleed from mandatory", got.Defaults.NamingTemplate)
	}
}

// TestMergeEventBridgeCascade_AC8 — level 3 wins over level 4 for namingTemplate.
func TestMergeEventBridgeCascade_AC8(t *testing.T) {
	const globalTmpl = "corp-{namespace}-{name}"
	const localTmpl = "{namespace}-{name}"
	got := mergeEBAll(
		zeroKropathEB,
		zeroKropathEB,
		cascade.EventBridgeConfigSection{NamingTemplate: globalTmpl}, // level 3
		cascade.EventBridgeConfigSection{NamingTemplate: localTmpl},  // level 4
		zeroEBCfg,
		zeroEBCfg,
		zeroKropathEB,
		zeroKropathEB,
	)

	if got.Mandatory.NamingTemplate != globalTmpl {
		t.Errorf("AC-8: mandatory.namingTemplate = %q, want %q (level 3 beats level 4)", got.Mandatory.NamingTemplate, globalTmpl)
	}
}

// TestMergeEventBridgeCascade_AC9 — namingTemplate at level 6 propagates to defaults only.
func TestMergeEventBridgeCascade_AC9(t *testing.T) {
	const tmpl = "{namespace}-{name}"
	got := mergeEBAll(
		zeroKropathEB,
		zeroKropathEB,
		zeroEBCfg,
		zeroEBCfg,
		cascade.EventBridgeConfigSection{NamingTemplate: tmpl}, // level 6
		zeroEBCfg,
		zeroKropathEB,
		zeroKropathEB,
	)

	if got.Defaults.NamingTemplate != tmpl {
		t.Errorf("AC-9: defaults.namingTemplate = %q, want %q", got.Defaults.NamingTemplate, tmpl)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("AC-9: mandatory.namingTemplate = %q, must not bleed from defaults", got.Mandatory.NamingTemplate)
	}
}

// TestMergeEventBridgeCascade_AC10 — mandatory tags additive union: all four mandatory sources
// merged; KropathConfig global (L1) wins on key conflict.
func TestMergeEventBridgeCascade_AC10(t *testing.T) {
	got := mergeEBAll(
		cascade.EventBridgeKropathSection{Tags: map[string]string{"cost-centre": "platform", "env": "org"}},  // level 1
		cascade.EventBridgeKropathSection{Tags: map[string]string{"team": "infra", "env": "local-kropath"}},  // level 2
		cascade.EventBridgeConfigSection{Tags: map[string]string{"owner": "events", "env": "global-eb"}},     // level 3
		cascade.EventBridgeConfigSection{Tags: map[string]string{"app": "myapp", "env": "local-eb"}},         // level 4
		zeroEBCfg,
		zeroEBCfg,
		zeroKropathEB,
		zeroKropathEB,
	)

	cases := map[string]string{
		"cost-centre": "platform", // from L1
		"env":         "org",      // L1 wins
		"team":        "infra",    // from L2
		"owner":       "events",   // from L3
		"app":         "myapp",    // from L4
	}
	for k, want := range cases {
		if got := got.Mandatory.Tags[k]; got != want {
			t.Errorf("AC-10: mandatory.tags[%q] = %q, want %q", k, got, want)
		}
	}
	if len(got.Defaults.Tags) != 0 {
		t.Errorf("AC-10: defaults.tags must be empty, got %v", got.Defaults.Tags)
	}
}

// TestMergeEventBridgeCascade_AC11 — defaults tags additive union: all four defaults sources
// merged; local EBConfig (L6) wins on key conflict.
func TestMergeEventBridgeCascade_AC11(t *testing.T) {
	got := mergeEBAll(
		zeroKropathEB,
		zeroKropathEB,
		zeroEBCfg,
		zeroEBCfg,
		cascade.EventBridgeConfigSection{Tags: map[string]string{"env": "local-eb", "app": "myapp"}}, // level 6
		cascade.EventBridgeConfigSection{Tags: map[string]string{"env": "global-eb", "team": "sre"}}, // level 7
		cascade.EventBridgeKropathSection{Tags: map[string]string{"cost-centre": "core"}},             // level 8
		cascade.EventBridgeKropathSection{Tags: map[string]string{"org": "acme"}},                     // level 9
	)

	cases := map[string]string{
		"env":         "local-eb", // L6 wins
		"app":         "myapp",    // from L6
		"team":        "sre",      // from L7
		"cost-centre": "core",     // from L8
		"org":         "acme",     // from L9
	}
	for k, want := range cases {
		if got := got.Defaults.Tags[k]; got != want {
			t.Errorf("AC-11: defaults.tags[%q] = %q, want %q", k, got, want)
		}
	}
}

// TestMergeEventBridgeCascade_AC12 — syncedLabels additive union in mandatory tier from
// EventBridgeConfig only; global (L3) wins on key conflict.
func TestMergeEventBridgeCascade_AC12(t *testing.T) {
	got := mergeEBAll(
		cascade.EventBridgeKropathSection{Tags: map[string]string{"ignored": "yes"}}, // L1 tags only — no syncedLabels
		zeroKropathEB,
		cascade.EventBridgeConfigSection{SyncedLabels: map[string]string{"env": "prod", "team": "events"}}, // L3
		cascade.EventBridgeConfigSection{SyncedLabels: map[string]string{"env": "staging", "app": "bus"}},  // L4
		zeroEBCfg,
		zeroEBCfg,
		zeroKropathEB,
		zeroKropathEB,
	)

	cases := map[string]string{
		"env":  "prod",   // L3 wins
		"team": "events", // from L3
		"app":  "bus",    // from L4
	}
	for k, want := range cases {
		if got := got.Mandatory.SyncedLabels[k]; got != want {
			t.Errorf("AC-12: mandatory.syncedLabels[%q] = %q, want %q", k, got, want)
		}
	}
}

// TestMergeEventBridgeCascade_AC13 — syncedAnnotations additive union in mandatory tier from
// EventBridgeConfig only; global (L3) wins on key conflict.
func TestMergeEventBridgeCascade_AC13(t *testing.T) {
	got := mergeEBAll(
		zeroKropathEB,
		zeroKropathEB,
		cascade.EventBridgeConfigSection{SyncedAnnotations: map[string]string{"team": "platform", "env": "prod"}}, // L3
		cascade.EventBridgeConfigSection{SyncedAnnotations: map[string]string{"team": "infra", "owner": "sre"}},   // L4
		zeroEBCfg,
		zeroEBCfg,
		zeroKropathEB,
		zeroKropathEB,
	)

	cases := map[string]string{
		"team":  "platform", // L3 wins
		"env":   "prod",     // from L3
		"owner": "sre",      // from L4
	}
	for k, want := range cases {
		if got := got.Mandatory.SyncedAnnotations[k]; got != want {
			t.Errorf("AC-13: mandatory.syncedAnnotations[%q] = %q, want %q", k, got, want)
		}
	}
}

// TestMergeEventBridgeCascade_AC14 — syncedLabels additive union in defaults tier from
// EventBridgeConfig only; local (L6) wins on key conflict.
func TestMergeEventBridgeCascade_AC14(t *testing.T) {
	got := mergeEBAll(
		zeroKropathEB,
		zeroKropathEB,
		zeroEBCfg,
		zeroEBCfg,
		cascade.EventBridgeConfigSection{SyncedLabels: map[string]string{"tier": "standard", "env": "local"}}, // L6
		cascade.EventBridgeConfigSection{SyncedLabels: map[string]string{"tier": "global", "team": "sre"}},   // L7
		zeroKropathEB,
		zeroKropathEB,
	)

	cases := map[string]string{
		"tier": "standard", // L6 wins
		"env":  "local",    // from L6
		"team": "sre",      // from L7
	}
	for k, want := range cases {
		if got := got.Defaults.SyncedLabels[k]; got != want {
			t.Errorf("AC-14: defaults.syncedLabels[%q] = %q, want %q", k, got, want)
		}
	}
}

// TestMergeEventBridgeCascade_AC15 — all-zero inputs yield zero effective config.
func TestMergeEventBridgeCascade_AC15(t *testing.T) {
	got := mergeEBAll(
		zeroKropathEB, zeroKropathEB,
		zeroEBCfg, zeroEBCfg, zeroEBCfg, zeroEBCfg,
		zeroKropathEB, zeroKropathEB,
	)

	if got.Mandatory.ArchiveRetentionDays != 0 {
		t.Errorf("AC-15: mandatory.archiveRetentionDays = %d, want 0", got.Mandatory.ArchiveRetentionDays)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("AC-15: mandatory.namingTemplate = %q, want empty", got.Mandatory.NamingTemplate)
	}
	if got.Defaults.ArchiveRetentionDays != 0 {
		t.Errorf("AC-15: defaults.archiveRetentionDays = %d, want 0", got.Defaults.ArchiveRetentionDays)
	}
	if got.Defaults.NamingTemplate != "" {
		t.Errorf("AC-15: defaults.namingTemplate = %q, want empty", got.Defaults.NamingTemplate)
	}
}
