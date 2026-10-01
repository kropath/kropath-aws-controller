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

// emptyAS is a zero-value AppScalingKropathSection (sentinel: nothing enforced).
var emptyAS = cascade.AppScalingKropathSection{}

// emptyASCfg is a zero-value AppScalingConfigSection.
var emptyASCfg = cascade.AppScalingConfigSection{}

// merge is a convenience wrapper that calls MergeAppScalingCascade with all eight
// inputs explicitly named so that the test cases remain readable.
func mergeAS(
	globalKropathMandatory, localKropathMandatory cascade.AppScalingKropathSection,
	globalASCfgMandatory, localASCfgMandatory cascade.AppScalingConfigSection,
	localASCfgDefaults, globalASCfgDefaults cascade.AppScalingConfigSection,
	localKropathDefaults, globalKropathDefaults cascade.AppScalingKropathSection,
) cascade.EffectiveAppScalingConfig {
	return cascade.MergeAppScalingCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalASCfgMandatory,
		localASCfgMandatory,
		localASCfgDefaults,
		globalASCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeAppScalingCascade_AllEmpty verifies that when all sources are zero-valued
// the result is entirely zero (no phantom values).
func TestMergeAppScalingCascade_AllEmpty(t *testing.T) {
	got := mergeAS(emptyAS, emptyAS, emptyASCfg, emptyASCfg, emptyASCfg, emptyASCfg, emptyAS, emptyAS)
	if got.Mandatory.MinCapacity != 0 {
		t.Errorf("expected Mandatory.MinCapacity=0 got %d", got.Mandatory.MinCapacity)
	}
	if got.Mandatory.MaxCapacity != 0 {
		t.Errorf("expected Mandatory.MaxCapacity=0 got %d", got.Mandatory.MaxCapacity)
	}
	if got.Mandatory.DisableScaleIn {
		t.Error("expected Mandatory.DisableScaleIn=false got true")
	}
	if got.Defaults.MinCapacity != 0 {
		t.Errorf("expected Defaults.MinCapacity=0 got %d", got.Defaults.MinCapacity)
	}
}

// TestMergeAppScalingCascade_GlobalKropathMandatoryWins_AC9 covers AC-9:
// KropathConfig.mandatory.minCapacity overrides a weaker AppScalingConfig value.
func TestMergeAppScalingCascade_GlobalKropathMandatoryWins_AC9(t *testing.T) {
	globalKM := cascade.AppScalingKropathSection{MinCapacity: 5}
	localASCfgM := cascade.AppScalingConfigSection{MinCapacity: 2}

	got := mergeAS(globalKM, emptyAS, emptyASCfg, localASCfgM, emptyASCfg, emptyASCfg, emptyAS, emptyAS)

	if got.Mandatory.MinCapacity != 5 {
		t.Errorf("AC-9: expected Mandatory.MinCapacity=5 (globalKropathMandatory wins), got %d", got.Mandatory.MinCapacity)
	}
}

// TestMergeAppScalingCascade_LocalASCfgMandatoryFallback verifies that when the
// global KropathConfig mandatory field is zero, the local AppScalingConfig mandatory
// value is used (level-4 fallback).
func TestMergeAppScalingCascade_LocalASCfgMandatoryFallback(t *testing.T) {
	localASCfgM := cascade.AppScalingConfigSection{MinCapacity: 3}

	got := mergeAS(emptyAS, emptyAS, emptyASCfg, localASCfgM, emptyASCfg, emptyASCfg, emptyAS, emptyAS)

	if got.Mandatory.MinCapacity != 3 {
		t.Errorf("expected Mandatory.MinCapacity=3 from localASCfgMandatory, got %d", got.Mandatory.MinCapacity)
	}
}

// TestMergeAppScalingCascade_DisableScaleIn_FirstTrueWins verifies that
// firstTrue semantics hold: the highest-priority source that is true wins.
func TestMergeAppScalingCascade_DisableScaleIn_FirstTrueWins(t *testing.T) {
	// Only level-3 (globalASCfgMandatory) sets DisableScaleIn=true.
	globalASCfgM := cascade.AppScalingConfigSection{DisableScaleIn: true}

	got := mergeAS(emptyAS, emptyAS, globalASCfgM, emptyASCfg, emptyASCfg, emptyASCfg, emptyAS, emptyAS)

	if !got.Mandatory.DisableScaleIn {
		t.Error("expected Mandatory.DisableScaleIn=true from globalASCfgMandatory, got false")
	}
}

// TestMergeAppScalingCascade_MaxCapacityDefaults verifies that the defaults chain
// resolution (level 6 → 7 → 8 → 9) picks the highest-priority non-zero value.
func TestMergeAppScalingCascade_MaxCapacityDefaults(t *testing.T) {
	localASCfgD := cascade.AppScalingConfigSection{MaxCapacity: 100}  // level 6 (strongest defaults)
	globalASCfgD := cascade.AppScalingConfigSection{MaxCapacity: 200} // level 7

	got := mergeAS(emptyAS, emptyAS, emptyASCfg, emptyASCfg, localASCfgD, globalASCfgD, emptyAS, emptyAS)

	if got.Defaults.MaxCapacity != 100 {
		t.Errorf("expected Defaults.MaxCapacity=100 (localASCfgDefaults level-6 wins), got %d", got.Defaults.MaxCapacity)
	}
}

// TestMergeAppScalingCascade_TagMerge_MandatoryPriorityOrder verifies that the
// mandatory tags union merge respects level priority: level-1 tags win on key conflict.
func TestMergeAppScalingCascade_TagMerge_MandatoryPriorityOrder(t *testing.T) {
	globalKM := cascade.AppScalingKropathSection{Tags: map[string]string{"env": "prod", "org": "acme"}}
	localASCfgM := cascade.AppScalingConfigSection{Tags: map[string]string{"env": "dev", "team": "platform"}}

	got := mergeAS(globalKM, emptyAS, emptyASCfg, localASCfgM, emptyASCfg, emptyASCfg, emptyAS, emptyAS)

	if got.Mandatory.Tags["env"] != "prod" {
		t.Errorf("expected Tags[env]=prod (globalKropathMandatory level-1 wins), got %q", got.Mandatory.Tags["env"])
	}
	if got.Mandatory.Tags["org"] != "acme" {
		t.Errorf("expected Tags[org]=acme from globalKropathMandatory, got %q", got.Mandatory.Tags["org"])
	}
	if got.Mandatory.Tags["team"] != "platform" {
		t.Errorf("expected Tags[team]=platform from localASCfgMandatory, got %q", got.Mandatory.Tags["team"])
	}
}

// TestMergeAppScalingCascade_TagMerge_DefaultsPriorityOrder verifies defaults tag
// priority: level-6 (localASCfgDefaults) wins on key conflict.
func TestMergeAppScalingCascade_TagMerge_DefaultsPriorityOrder(t *testing.T) {
	localASCfgD := cascade.AppScalingConfigSection{Tags: map[string]string{"env": "staging"}} // level 6
	globalKD := cascade.AppScalingKropathSection{Tags: map[string]string{"env": "shared"}}    // level 9

	got := mergeAS(emptyAS, emptyAS, emptyASCfg, emptyASCfg, localASCfgD, emptyASCfg, emptyAS, globalKD)

	if got.Defaults.Tags["env"] != "staging" {
		t.Errorf("expected Defaults.Tags[env]=staging (localASCfgDefaults level-6 wins), got %q", got.Defaults.Tags["env"])
	}
}

// TestMergeAppScalingCascade_SyncedLabels_MandatoryUnion verifies that SyncedLabels
// from mandatory AppScalingConfig sources are unioned additively (level-3 wins on conflict).
func TestMergeAppScalingCascade_SyncedLabels_MandatoryUnion(t *testing.T) {
	globalASCfgM := cascade.AppScalingConfigSection{SyncedLabels: map[string]string{"tier": "gold", "shared": "yes"}}
	localASCfgM := cascade.AppScalingConfigSection{SyncedLabels: map[string]string{"tier": "silver", "local": "true"}}

	got := mergeAS(emptyAS, emptyAS, globalASCfgM, localASCfgM, emptyASCfg, emptyASCfg, emptyAS, emptyAS)

	if got.Mandatory.SyncedLabels["tier"] != "gold" {
		t.Errorf("expected SyncedLabels[tier]=gold (globalASCfgMandatory level-3 wins), got %q", got.Mandatory.SyncedLabels["tier"])
	}
	if got.Mandatory.SyncedLabels["local"] != "true" {
		t.Errorf("expected SyncedLabels[local]=true from localASCfgMandatory, got %q", got.Mandatory.SyncedLabels["local"])
	}
	if got.Mandatory.SyncedLabels["shared"] != "yes" {
		t.Errorf("expected SyncedLabels[shared]=yes from globalASCfgMandatory, got %q", got.Mandatory.SyncedLabels["shared"])
	}
}

// TestMergeAppScalingCascade_SyncedAnnotations_DefaultsUnion verifies additive union
// for SyncedAnnotations in the defaults chain (level-6 wins on conflict).
func TestMergeAppScalingCascade_SyncedAnnotations_DefaultsUnion(t *testing.T) {
	localASCfgD := cascade.AppScalingConfigSection{SyncedAnnotations: map[string]string{"prov": "team-a"}}
	globalASCfgD := cascade.AppScalingConfigSection{SyncedAnnotations: map[string]string{"prov": "global", "contact": "ops@example.com"}}

	got := mergeAS(emptyAS, emptyAS, emptyASCfg, emptyASCfg, localASCfgD, globalASCfgD, emptyAS, emptyAS)

	if got.Defaults.SyncedAnnotations["prov"] != "team-a" {
		t.Errorf("expected SyncedAnnotations[prov]=team-a (localASCfgDefaults level-6 wins), got %q", got.Defaults.SyncedAnnotations["prov"])
	}
	if got.Defaults.SyncedAnnotations["contact"] != "ops@example.com" {
		t.Errorf("expected SyncedAnnotations[contact]=ops@example.com from globalASCfgDefaults, got %q", got.Defaults.SyncedAnnotations["contact"])
	}
}

// TestMergeAppScalingCascade_Cooldowns_MandatoryChain verifies that cooldown fields
// in the mandatory chain follow the same first-non-zero priority order.
func TestMergeAppScalingCascade_Cooldowns_MandatoryChain(t *testing.T) {
	globalKM := cascade.AppScalingKropathSection{ScaleInCooldown: 120, ScaleOutCooldown: 60}
	localASCfgM := cascade.AppScalingConfigSection{ScaleInCooldown: 30, ScaleOutCooldown: 15}

	got := mergeAS(globalKM, emptyAS, emptyASCfg, localASCfgM, emptyASCfg, emptyASCfg, emptyAS, emptyAS)

	if got.Mandatory.ScaleInCooldown != 120 {
		t.Errorf("expected Mandatory.ScaleInCooldown=120 (globalKropathMandatory level-1 wins), got %d", got.Mandatory.ScaleInCooldown)
	}
	if got.Mandatory.ScaleOutCooldown != 60 {
		t.Errorf("expected Mandatory.ScaleOutCooldown=60 (globalKropathMandatory level-1 wins), got %d", got.Mandatory.ScaleOutCooldown)
	}
}

// TestMergeAppScalingCascade_GlobalKropathDefaults_WeakestDefaults verifies that
// the global KropathConfig defaults (level 9, weakest) are used only when no
// stronger defaults source provides a non-zero value.
func TestMergeAppScalingCascade_GlobalKropathDefaults_WeakestDefaults(t *testing.T) {
	globalKD := cascade.AppScalingKropathSection{MinCapacity: 2} // level 9 (weakest)

	got := mergeAS(emptyAS, emptyAS, emptyASCfg, emptyASCfg, emptyASCfg, emptyASCfg, emptyAS, globalKD)

	if got.Defaults.MinCapacity != 2 {
		t.Errorf("expected Defaults.MinCapacity=2 from globalKropathDefaults (level-9 fallback), got %d", got.Defaults.MinCapacity)
	}
}
