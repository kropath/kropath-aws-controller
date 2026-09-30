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

// zeroBedrockKropath is a zero-value BedrockKropathSection.
var zeroBedrockKropath = cascade.BedrockKropathSection{}

// zeroBedrockCfg is a zero-value BedrockConfigSection.
var zeroBedrockCfg = cascade.BedrockConfigSection{}

// mergeBedrock calls MergeBedrockCascade with all eight inputs.
func mergeBedrock(
	globalKropathMandatory, localKropathMandatory cascade.BedrockKropathSection,
	globalBCCfgMandatory, localBCCfgMandatory,
	localBCCfgDefaults, globalBCCfgDefaults cascade.BedrockConfigSection,
	localKropathDefaults, globalKropathDefaults cascade.BedrockKropathSection,
) cascade.EffectiveBedrockConfig {
	return cascade.MergeBedrockCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalBCCfgMandatory,
		localBCCfgMandatory,
		localBCCfgDefaults,
		globalBCCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeBedrockCascade_AC1 — globalKropathConfig.mandatory.bedrock.guardrailIdentifier (level 1)
// and guardrailVersion both propagate to effCfg.mandatory.
func TestMergeBedrockCascade_AC1(t *testing.T) {
	got := mergeBedrock(
		cascade.BedrockKropathSection{GuardrailIdentifier: "abc123", GuardrailVersion: "1"}, // level 1
		zeroBedrockKropath,
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockKropath,
		zeroBedrockKropath,
	)

	if got.Mandatory.GuardrailIdentifier != "abc123" {
		t.Errorf("AC-1: mandatory.guardrailIdentifier = %q, want %q", got.Mandatory.GuardrailIdentifier, "abc123")
	}
	if got.Mandatory.GuardrailVersion != "1" {
		t.Errorf("AC-1: mandatory.guardrailVersion = %q, want %q", got.Mandatory.GuardrailVersion, "1")
	}
	if got.Defaults.GuardrailIdentifier != "" {
		t.Errorf("AC-1: defaults.guardrailIdentifier must not bleed from mandatory, got %q", got.Defaults.GuardrailIdentifier)
	}
}

// TestMergeBedrockCascade_AC2 — level 1 KropathConfig wins over level 3 BedrockConfig.
func TestMergeBedrockCascade_AC2(t *testing.T) {
	got := mergeBedrock(
		cascade.BedrockKropathSection{GuardrailIdentifier: "abc123", GuardrailVersion: "1"}, // level 1
		zeroBedrockKropath,
		cascade.BedrockConfigSection{GuardrailIdentifier: "def456", GuardrailVersion: "2"}, // level 3
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockKropath,
		zeroBedrockKropath,
	)

	if got.Mandatory.GuardrailIdentifier != "abc123" {
		t.Errorf("AC-2: level 1 should win; got %q", got.Mandatory.GuardrailIdentifier)
	}
}

// TestMergeBedrockCascade_AC3 — only globalBedrockConfig.defaults set; mandatory should be empty.
func TestMergeBedrockCascade_AC3(t *testing.T) {
	got := mergeBedrock(
		zeroBedrockKropath,
		zeroBedrockKropath,
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockCfg,
		cascade.BedrockConfigSection{GuardrailIdentifier: "abc123", GuardrailVersion: "DRAFT"}, // level 7
		zeroBedrockKropath,
		zeroBedrockKropath,
	)

	if got.Defaults.GuardrailIdentifier != "abc123" {
		t.Errorf("AC-3: defaults.guardrailIdentifier = %q, want %q", got.Defaults.GuardrailIdentifier, "abc123")
	}
	if got.Defaults.GuardrailVersion != "DRAFT" {
		t.Errorf("AC-3: defaults.guardrailVersion = %q, want %q", got.Defaults.GuardrailVersion, "DRAFT")
	}
	if got.Mandatory.GuardrailIdentifier != "" {
		t.Errorf("AC-3: mandatory.guardrailIdentifier must be empty, got %q", got.Mandatory.GuardrailIdentifier)
	}
}

// TestMergeBedrockCascade_AC4 — global BedrockConfig.mandatory.foundationModel (level 3) wins
// over local BedrockConfig.mandatory.foundationModel (level 4).
func TestMergeBedrockCascade_AC4(t *testing.T) {
	got := mergeBedrock(
		zeroBedrockKropath,
		zeroBedrockKropath,
		cascade.BedrockConfigSection{FoundationModel: "anthropic.claude-3-5-sonnet-20241022-v2:0"}, // level 3
		cascade.BedrockConfigSection{FoundationModel: "amazon.titan-text-premier-v1:0"},             // level 4
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockKropath,
		zeroBedrockKropath,
	)

	if got.Mandatory.FoundationModel != "anthropic.claude-3-5-sonnet-20241022-v2:0" {
		t.Errorf("AC-4: global level 3 should win; got %q", got.Mandatory.FoundationModel)
	}
}

// TestMergeBedrockCascade_AC5 — only globalBedrockConfig.defaults.foundationModel; mandatory empty.
func TestMergeBedrockCascade_AC5(t *testing.T) {
	got := mergeBedrock(
		zeroBedrockKropath,
		zeroBedrockKropath,
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockCfg,
		cascade.BedrockConfigSection{FoundationModel: "anthropic.claude-3-haiku-20240307-v1:0"}, // level 7
		zeroBedrockKropath,
		zeroBedrockKropath,
	)

	if got.Defaults.FoundationModel != "anthropic.claude-3-haiku-20240307-v1:0" {
		t.Errorf("AC-5: defaults.foundationModel = %q, want haiku", got.Defaults.FoundationModel)
	}
	if got.Mandatory.FoundationModel != "" {
		t.Errorf("AC-5: mandatory.foundationModel must be empty, got %q", got.Mandatory.FoundationModel)
	}
}

// TestMergeBedrockCascade_AC6 — globalKropathConfig.mandatory.bedrock.allowedModels (level 1).
func TestMergeBedrockCascade_AC6(t *testing.T) {
	models := []string{"anthropic.claude-3-5-sonnet-20241022-v2:0", "anthropic.claude-3-haiku-20240307-v1:0"}
	got := mergeBedrock(
		cascade.BedrockKropathSection{AllowedModels: models}, // level 1
		zeroBedrockKropath,
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockKropath,
		zeroBedrockKropath,
	)

	if len(got.Mandatory.AllowedModels) != 2 {
		t.Errorf("AC-6: mandatory.allowedModels len = %d, want 2", len(got.Mandatory.AllowedModels))
	}
}

// TestMergeBedrockCascade_AC7 — level 1 KropathConfig allowedModels wins over level 3 BedrockConfig.
func TestMergeBedrockCascade_AC7(t *testing.T) {
	kcModels := []string{"model-a"}
	bcModels := []string{"model-b", "model-c"}
	got := mergeBedrock(
		cascade.BedrockKropathSection{AllowedModels: kcModels}, // level 1
		zeroBedrockKropath,
		cascade.BedrockConfigSection{AllowedModels: bcModels}, // level 3
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockKropath,
		zeroBedrockKropath,
	)

	if len(got.Mandatory.AllowedModels) != 1 || got.Mandatory.AllowedModels[0] != "model-a" {
		t.Errorf("AC-7: level 1 allowedModels should win; got %v", got.Mandatory.AllowedModels)
	}
}

// TestMergeBedrockCascade_AC8 — globalBedrockConfig.mandatory.maxIterations (level 3).
func TestMergeBedrockCascade_AC8(t *testing.T) {
	got := mergeBedrock(
		zeroBedrockKropath,
		zeroBedrockKropath,
		cascade.BedrockConfigSection{MaxIterations: 50}, // level 3
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockKropath,
		zeroBedrockKropath,
	)

	if got.Mandatory.MaxIterations != 50 {
		t.Errorf("AC-8: mandatory.maxIterations = %d, want 50", got.Mandatory.MaxIterations)
	}
}

// TestMergeBedrockCascade_AC9 — only globalBedrockConfig.defaults.maxIterations; mandatory 0.
func TestMergeBedrockCascade_AC9(t *testing.T) {
	got := mergeBedrock(
		zeroBedrockKropath,
		zeroBedrockKropath,
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockCfg,
		cascade.BedrockConfigSection{MaxIterations: 100}, // level 7
		zeroBedrockKropath,
		zeroBedrockKropath,
	)

	if got.Defaults.MaxIterations != 100 {
		t.Errorf("AC-9: defaults.maxIterations = %d, want 100", got.Defaults.MaxIterations)
	}
	if got.Mandatory.MaxIterations != 0 {
		t.Errorf("AC-9: mandatory.maxIterations must be 0, got %d", got.Mandatory.MaxIterations)
	}
}

// TestMergeBedrockCascade_AC10 — globalBedrockConfig.mandatory.maxTokens (level 3).
func TestMergeBedrockCascade_AC10(t *testing.T) {
	got := mergeBedrock(
		zeroBedrockKropath,
		zeroBedrockKropath,
		cascade.BedrockConfigSection{MaxTokens: 50000}, // level 3
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockKropath,
		zeroBedrockKropath,
	)

	if got.Mandatory.MaxTokens != 50000 {
		t.Errorf("AC-10: mandatory.maxTokens = %d, want 50000", got.Mandatory.MaxTokens)
	}
}

// TestMergeBedrockCascade_AC11 — globalBedrockConfig.mandatory.timeoutSeconds (level 3).
func TestMergeBedrockCascade_AC11(t *testing.T) {
	got := mergeBedrock(
		zeroBedrockKropath,
		zeroBedrockKropath,
		cascade.BedrockConfigSection{TimeoutSeconds: 300}, // level 3
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockKropath,
		zeroBedrockKropath,
	)

	if got.Mandatory.TimeoutSeconds != 300 {
		t.Errorf("AC-11: mandatory.timeoutSeconds = %d, want 300", got.Mandatory.TimeoutSeconds)
	}
}

// TestMergeBedrockCascade_AC12 — globalBedrockConfig.mandatory.idleSessionTTLInSeconds (level 3).
func TestMergeBedrockCascade_AC12(t *testing.T) {
	got := mergeBedrock(
		zeroBedrockKropath,
		zeroBedrockKropath,
		cascade.BedrockConfigSection{IdleSessionTTLInSeconds: 1800}, // level 3
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockKropath,
		zeroBedrockKropath,
	)

	if got.Mandatory.IdleSessionTTLInSeconds != 1800 {
		t.Errorf("AC-12: mandatory.idleSessionTTLInSeconds = %d, want 1800", got.Mandatory.IdleSessionTTLInSeconds)
	}
}

// TestValidateBedrockConfig_AC13 — foundationModel not in allowedModels → InvalidModelNotInAllowedList.
func TestValidateBedrockConfig_AC13(t *testing.T) {
	eff := cascade.EffectiveBedrockConfig{
		Mandatory: cascade.EffectiveBedrockSection{
			FoundationModel: "anthropic.claude-v2",
			AllowedModels:   []string{"anthropic.claude-3-sonnet"},
		},
	}
	valid, reason, _ := cascade.ValidateBedrockConfig(eff)
	if valid {
		t.Error("AC-13: expected validation failure")
	}
	if reason != "InvalidModelNotInAllowedList" {
		t.Errorf("AC-13: reason = %q, want InvalidModelNotInAllowedList", reason)
	}
}

// TestValidateBedrockConfig_AC14 — foundationModel in allowedModels → valid.
func TestValidateBedrockConfig_AC14(t *testing.T) {
	eff := cascade.EffectiveBedrockConfig{
		Mandatory: cascade.EffectiveBedrockSection{
			FoundationModel: "anthropic.claude-3-5-sonnet-20241022-v2:0",
			AllowedModels:   []string{"anthropic.claude-3-5-sonnet-20241022-v2:0", "amazon.titan-text-premier-v1:0"},
		},
	}
	valid, _, _ := cascade.ValidateBedrockConfig(eff)
	if !valid {
		t.Error("AC-14: expected validation pass")
	}
}

// TestValidateBedrockConfig_AC15 — foundationModel set, allowedModels empty → valid (no restriction).
func TestValidateBedrockConfig_AC15(t *testing.T) {
	eff := cascade.EffectiveBedrockConfig{
		Mandatory: cascade.EffectiveBedrockSection{
			FoundationModel: "anthropic.claude-3-5-sonnet-20241022-v2:0",
			AllowedModels:   nil,
		},
	}
	valid, _, _ := cascade.ValidateBedrockConfig(eff)
	if !valid {
		t.Error("AC-15: empty allowedModels means no restriction; expected validation pass")
	}
}

// TestValidateBedrockConfig_AC16 — mandatory.guardrailIdentifier set, mandatory.guardrailVersion empty → InvalidGuardrailConfiguration.
func TestValidateBedrockConfig_AC16(t *testing.T) {
	eff := cascade.EffectiveBedrockConfig{
		Mandatory: cascade.EffectiveBedrockSection{
			GuardrailIdentifier: "abc123",
			GuardrailVersion:    "",
		},
	}
	valid, reason, _ := cascade.ValidateBedrockConfig(eff)
	if valid {
		t.Error("AC-16: expected validation failure for guardrail pair mismatch in mandatory")
	}
	if reason != "InvalidGuardrailConfiguration" {
		t.Errorf("AC-16: reason = %q, want InvalidGuardrailConfiguration", reason)
	}
}

// TestValidateBedrockConfig_AC17 — defaults.guardrailVersion set, defaults.guardrailIdentifier empty → InvalidGuardrailConfiguration.
func TestValidateBedrockConfig_AC17(t *testing.T) {
	eff := cascade.EffectiveBedrockConfig{
		Defaults: cascade.EffectiveBedrockSection{
			GuardrailVersion:    "1",
			GuardrailIdentifier: "",
		},
	}
	valid, reason, _ := cascade.ValidateBedrockConfig(eff)
	if valid {
		t.Error("AC-17: expected validation failure for guardrail pair mismatch in defaults")
	}
	if reason != "InvalidGuardrailConfiguration" {
		t.Errorf("AC-17: reason = %q, want InvalidGuardrailConfiguration", reason)
	}
}

// TestValidateBedrockConfig_AC18 — both guardrailIdentifier and guardrailVersion set in mandatory → valid.
func TestValidateBedrockConfig_AC18(t *testing.T) {
	eff := cascade.EffectiveBedrockConfig{
		Mandatory: cascade.EffectiveBedrockSection{
			GuardrailIdentifier: "abc123",
			GuardrailVersion:    "1",
		},
	}
	valid, _, _ := cascade.ValidateBedrockConfig(eff)
	if !valid {
		t.Error("AC-18: both guardrail fields set; expected validation pass")
	}
}

// TestMergeBedrockCascade_AC22 — tags-merge: KropathConfig.mandatory.tags union with BedrockConfig.mandatory.tags.
func TestMergeBedrockCascade_AC23(t *testing.T) {
	got := mergeBedrock(
		cascade.BedrockKropathSection{Tags: map[string]string{"cost-centre": "ai"}}, // level 1
		zeroBedrockKropath,
		cascade.BedrockConfigSection{Tags: map[string]string{"ai-safety": "enforced"}}, // level 3
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockKropath,
		zeroBedrockKropath,
	)

	if got.Mandatory.Tags["cost-centre"] != "ai" {
		t.Errorf("AC-23: mandatory.tags[cost-centre] = %q, want %q", got.Mandatory.Tags["cost-centre"], "ai")
	}
	if got.Mandatory.Tags["ai-safety"] != "enforced" {
		t.Errorf("AC-23: mandatory.tags[ai-safety] = %q, want %q", got.Mandatory.Tags["ai-safety"], "enforced")
	}
}

// TestMergeBedrockCascade_Level1WinsGuardrailOverLevel3 — supplemental priority test.
func TestMergeBedrockCascade_Level1WinsGuardrailOverLevel3(t *testing.T) {
	got := mergeBedrock(
		cascade.BedrockKropathSection{GuardrailIdentifier: "kc-guardrail", GuardrailVersion: "kc-1"}, // level 1
		zeroBedrockKropath,
		cascade.BedrockConfigSection{GuardrailIdentifier: "bc-guardrail", GuardrailVersion: "bc-1"}, // level 3
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockCfg,
		zeroBedrockKropath,
		zeroBedrockKropath,
	)

	if got.Mandatory.GuardrailIdentifier != "kc-guardrail" {
		t.Errorf("level 1 KropathConfig should win; got %q", got.Mandatory.GuardrailIdentifier)
	}
}

// TestMergeBedrockCascade_DefaultsLevel6WinsOverLevel7 — more specific defaults wins.
func TestMergeBedrockCascade_DefaultsLevel6WinsOverLevel7(t *testing.T) {
	got := mergeBedrock(
		zeroBedrockKropath,
		zeroBedrockKropath,
		zeroBedrockCfg,
		zeroBedrockCfg,
		cascade.BedrockConfigSection{GuardrailIdentifier: "local-id", GuardrailVersion: "local-1"}, // level 6
		cascade.BedrockConfigSection{GuardrailIdentifier: "global-id", GuardrailVersion: "global-1"}, // level 7
		zeroBedrockKropath,
		zeroBedrockKropath,
	)

	if got.Defaults.GuardrailIdentifier != "local-id" {
		t.Errorf("level 6 (local BCCfg) should win in defaults; got %q", got.Defaults.GuardrailIdentifier)
	}
}
