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

// zeroKropathSFN is a zero-value StepFunctionsKropathSection (absent source).
var zeroKropathSFN = cascade.StepFunctionsKropathSection{}

// zeroSFNCfg is a zero-value StepFunctionsConfigSection (absent source).
var zeroSFNCfg = cascade.StepFunctionsConfigSection{}

// mergeSFNAll calls MergeStepFunctionsCascade with all eight inputs.
func mergeSFNAll(
	globalKropathMandatory,
	localKropathMandatory cascade.StepFunctionsKropathSection,
	globalSFNCfgMandatory,
	localSFNCfgMandatory,
	localSFNCfgDefaults,
	globalSFNCfgDefaults cascade.StepFunctionsConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.StepFunctionsKropathSection,
) cascade.EffectiveStepFunctionsConfig {
	return cascade.MergeStepFunctionsCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalSFNCfgMandatory,
		localSFNCfgMandatory,
		localSFNCfgDefaults,
		globalSFNCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// boolPtrSFN returns a pointer to the given bool value.
func boolPtrSFN(b bool) *bool { return &b }

// --- loggingLevel cascade tests ---

// TestMergeStepFunctionsCascade_AC1 — globalKropathConfig.mandatory.stepfunctions.loggingLevel
// at level 1 propagates to effectiveConfig.mandatory.loggingLevel (L1 highest priority).
func TestMergeStepFunctionsCascade_AC1(t *testing.T) {
	got := mergeSFNAll(
		cascade.StepFunctionsKropathSection{LoggingLevel: "ALL"}, // level 1
		zeroKropathSFN,
		zeroSFNCfg,
		zeroSFNCfg,
		zeroSFNCfg,
		zeroSFNCfg,
		zeroKropathSFN,
		zeroKropathSFN,
	)

	if got.Mandatory.LoggingLevel != "ALL" {
		t.Errorf("AC-1: mandatory.loggingLevel = %q, want %q", got.Mandatory.LoggingLevel, "ALL")
	}
	if got.Defaults.LoggingLevel != "" {
		t.Errorf("AC-1: defaults.loggingLevel = %q, must not bleed from mandatory", got.Defaults.LoggingLevel)
	}
}

// TestMergeStepFunctionsCascade_AC2 — localKropathConfig.mandatory.stepfunctions.loggingLevel
// at level 2 propagates when level 1 is absent.
func TestMergeStepFunctionsCascade_AC2(t *testing.T) {
	got := mergeSFNAll(
		zeroKropathSFN,
		cascade.StepFunctionsKropathSection{LoggingLevel: "ERROR"}, // level 2
		zeroSFNCfg,
		zeroSFNCfg,
		zeroSFNCfg,
		zeroSFNCfg,
		zeroKropathSFN,
		zeroKropathSFN,
	)

	if got.Mandatory.LoggingLevel != "ERROR" {
		t.Errorf("AC-2: mandatory.loggingLevel = %q, want %q (level 2 when level 1 absent)", got.Mandatory.LoggingLevel, "ERROR")
	}
}

// TestMergeStepFunctionsCascade_AC3 — L1 beats L2 for loggingLevel in mandatory tier.
func TestMergeStepFunctionsCascade_AC3(t *testing.T) {
	got := mergeSFNAll(
		cascade.StepFunctionsKropathSection{LoggingLevel: "ALL"},   // level 1
		cascade.StepFunctionsKropathSection{LoggingLevel: "ERROR"}, // level 2
		zeroSFNCfg,
		zeroSFNCfg,
		zeroSFNCfg,
		zeroSFNCfg,
		zeroKropathSFN,
		zeroKropathSFN,
	)

	if got.Mandatory.LoggingLevel != "ALL" {
		t.Errorf("AC-3: mandatory.loggingLevel = %q, want %q (level 1 must beat level 2)", got.Mandatory.LoggingLevel, "ALL")
	}
}

// TestMergeStepFunctionsCascade_AC4 — globalStepFunctionsConfig.mandatory.loggingLevel
// at level 3 propagates when levels 1-2 are absent.
func TestMergeStepFunctionsCascade_AC4(t *testing.T) {
	got := mergeSFNAll(
		zeroKropathSFN,
		zeroKropathSFN,
		cascade.StepFunctionsConfigSection{LoggingLevel: "FATAL"}, // level 3
		zeroSFNCfg,
		zeroSFNCfg,
		zeroSFNCfg,
		zeroKropathSFN,
		zeroKropathSFN,
	)

	if got.Mandatory.LoggingLevel != "FATAL" {
		t.Errorf("AC-4: mandatory.loggingLevel = %q, want %q (level 3 when 1-2 absent)", got.Mandatory.LoggingLevel, "FATAL")
	}
}

// TestMergeStepFunctionsCascade_AC5 — L1 beats L3 for loggingLevel in mandatory tier.
func TestMergeStepFunctionsCascade_AC5(t *testing.T) {
	got := mergeSFNAll(
		cascade.StepFunctionsKropathSection{LoggingLevel: "ALL"},  // level 1
		zeroKropathSFN,
		cascade.StepFunctionsConfigSection{LoggingLevel: "ERROR"}, // level 3
		zeroSFNCfg,
		zeroSFNCfg,
		zeroSFNCfg,
		zeroKropathSFN,
		zeroKropathSFN,
	)

	if got.Mandatory.LoggingLevel != "ALL" {
		t.Errorf("AC-5: mandatory.loggingLevel = %q, want %q (level 1 must beat level 3)", got.Mandatory.LoggingLevel, "ALL")
	}
}

// TestMergeStepFunctionsCascade_AC6 — localStepFunctionsConfig.mandatory.loggingLevel
// at level 4 propagates when levels 1-3 are absent.
func TestMergeStepFunctionsCascade_AC6(t *testing.T) {
	got := mergeSFNAll(
		zeroKropathSFN,
		zeroKropathSFN,
		zeroSFNCfg,
		cascade.StepFunctionsConfigSection{LoggingLevel: "OFF"}, // level 4
		zeroSFNCfg,
		zeroSFNCfg,
		zeroKropathSFN,
		zeroKropathSFN,
	)

	if got.Mandatory.LoggingLevel != "OFF" {
		t.Errorf("AC-6: mandatory.loggingLevel = %q, want %q (level 4 when 1-3 absent)", got.Mandatory.LoggingLevel, "OFF")
	}
}

// TestMergeStepFunctionsCascade_AC7 — globalStepFunctionsConfig.defaults.loggingLevel
// at level 7 propagates to effectiveConfig.defaults when levels 6 absent.
func TestMergeStepFunctionsCascade_AC7(t *testing.T) {
	got := mergeSFNAll(
		zeroKropathSFN,
		zeroKropathSFN,
		zeroSFNCfg,
		zeroSFNCfg,
		zeroSFNCfg,
		cascade.StepFunctionsConfigSection{LoggingLevel: "OFF"}, // level 7
		zeroKropathSFN,
		zeroKropathSFN,
	)

	if got.Defaults.LoggingLevel != "OFF" {
		t.Errorf("AC-7: defaults.loggingLevel = %q, want %q (level 7 when 6 absent)", got.Defaults.LoggingLevel, "OFF")
	}
	if got.Mandatory.LoggingLevel != "" {
		t.Errorf("AC-7: mandatory.loggingLevel = %q, must not bleed from defaults", got.Mandatory.LoggingLevel)
	}
}

// TestMergeStepFunctionsCascade_AC8 — localStepFunctionsConfig.defaults.loggingLevel (L6)
// beats globalStepFunctionsConfig.defaults.loggingLevel (L7).
func TestMergeStepFunctionsCascade_AC8(t *testing.T) {
	got := mergeSFNAll(
		zeroKropathSFN,
		zeroKropathSFN,
		zeroSFNCfg,
		zeroSFNCfg,
		cascade.StepFunctionsConfigSection{LoggingLevel: "ERROR"}, // level 6
		cascade.StepFunctionsConfigSection{LoggingLevel: "OFF"},   // level 7
		zeroKropathSFN,
		zeroKropathSFN,
	)

	if got.Defaults.LoggingLevel != "ERROR" {
		t.Errorf("AC-8: defaults.loggingLevel = %q, want %q (L6 must beat L7)", got.Defaults.LoggingLevel, "ERROR")
	}
}

// TestMergeStepFunctionsCascade_AC9 — L6 beats L8 for loggingLevel in defaults tier.
func TestMergeStepFunctionsCascade_AC9(t *testing.T) {
	got := mergeSFNAll(
		zeroKropathSFN,
		zeroKropathSFN,
		zeroSFNCfg,
		zeroSFNCfg,
		cascade.StepFunctionsConfigSection{LoggingLevel: "ERROR"},  // level 6
		zeroSFNCfg,
		cascade.StepFunctionsKropathSection{LoggingLevel: "FATAL"}, // level 8
		zeroKropathSFN,
	)

	if got.Defaults.LoggingLevel != "ERROR" {
		t.Errorf("AC-9: defaults.loggingLevel = %q, want %q (L6 must beat L8)", got.Defaults.LoggingLevel, "ERROR")
	}
}

// TestMergeStepFunctionsCascade_AC10 — localKropathConfig.defaults.stepfunctions.loggingLevel
// at level 8 propagates when levels 6-7 are absent.
func TestMergeStepFunctionsCascade_AC10(t *testing.T) {
	got := mergeSFNAll(
		zeroKropathSFN,
		zeroKropathSFN,
		zeroSFNCfg,
		zeroSFNCfg,
		zeroSFNCfg,
		zeroSFNCfg,
		cascade.StepFunctionsKropathSection{LoggingLevel: "ALL"}, // level 8
		zeroKropathSFN,
	)

	if got.Defaults.LoggingLevel != "ALL" {
		t.Errorf("AC-10: defaults.loggingLevel = %q, want %q (level 8 when 6-7 absent)", got.Defaults.LoggingLevel, "ALL")
	}
}

// TestMergeStepFunctionsCascade_AC11 — L8 beats L9 for loggingLevel in defaults tier.
func TestMergeStepFunctionsCascade_AC11(t *testing.T) {
	got := mergeSFNAll(
		zeroKropathSFN,
		zeroKropathSFN,
		zeroSFNCfg,
		zeroSFNCfg,
		zeroSFNCfg,
		zeroSFNCfg,
		cascade.StepFunctionsKropathSection{LoggingLevel: "ALL"},   // level 8
		cascade.StepFunctionsKropathSection{LoggingLevel: "ERROR"}, // level 9
	)

	if got.Defaults.LoggingLevel != "ALL" {
		t.Errorf("AC-11: defaults.loggingLevel = %q, want %q (L8 must beat L9)", got.Defaults.LoggingLevel, "ALL")
	}
}

// TestMergeStepFunctionsCascade_AC12 — globalKropathConfig.defaults.stepfunctions.loggingLevel
// at level 9 propagates when levels 6-8 are absent.
func TestMergeStepFunctionsCascade_AC12(t *testing.T) {
	got := mergeSFNAll(
		zeroKropathSFN,
		zeroKropathSFN,
		zeroSFNCfg,
		zeroSFNCfg,
		zeroSFNCfg,
		zeroSFNCfg,
		zeroKropathSFN,
		cascade.StepFunctionsKropathSection{LoggingLevel: "OFF"}, // level 9
	)

	if got.Defaults.LoggingLevel != "OFF" {
		t.Errorf("AC-12: defaults.loggingLevel = %q, want %q (level 9 when 6-8 absent)", got.Defaults.LoggingLevel, "OFF")
	}
}

// --- tracingEnabled (*bool) cascade tests ---

// TestMergeStepFunctionsCascade_AC13 — globalStepFunctionsConfig.mandatory.tracingEnabled=true
// at level 3 propagates to effectiveConfig.mandatory.tracingEnabled.
func TestMergeStepFunctionsCascade_AC13(t *testing.T) {
	got := mergeSFNAll(
		zeroKropathSFN,
		zeroKropathSFN,
		cascade.StepFunctionsConfigSection{TracingEnabled: boolPtrSFN(true)}, // level 3
		zeroSFNCfg,
		zeroSFNCfg,
		zeroSFNCfg,
		zeroKropathSFN,
		zeroKropathSFN,
	)

	if got.Mandatory.TracingEnabled == nil || !*got.Mandatory.TracingEnabled {
		t.Errorf("AC-13: mandatory.tracingEnabled = %v, want *true", got.Mandatory.TracingEnabled)
	}
	if got.Defaults.TracingEnabled != nil {
		t.Errorf("AC-13: defaults.tracingEnabled = %v, must not bleed from mandatory", got.Defaults.TracingEnabled)
	}
}

// TestMergeStepFunctionsCascade_AC14 — tracingEnabled=false at level 3 is explicitly
// enforced (false != nil — must propagate as *false, not be skipped).
func TestMergeStepFunctionsCascade_AC14(t *testing.T) {
	got := mergeSFNAll(
		zeroKropathSFN,
		zeroKropathSFN,
		cascade.StepFunctionsConfigSection{TracingEnabled: boolPtrSFN(false)}, // level 3
		zeroSFNCfg,
		zeroSFNCfg,
		zeroSFNCfg,
		zeroKropathSFN,
		zeroKropathSFN,
	)

	if got.Mandatory.TracingEnabled == nil || *got.Mandatory.TracingEnabled {
		t.Errorf("AC-14: mandatory.tracingEnabled = %v, want *false (explicitly disabled)", got.Mandatory.TracingEnabled)
	}
}

// TestMergeStepFunctionsCascade_AC15 — L3 wins over L4 for tracingEnabled in mandatory tier.
func TestMergeStepFunctionsCascade_AC15(t *testing.T) {
	got := mergeSFNAll(
		zeroKropathSFN,
		zeroKropathSFN,
		cascade.StepFunctionsConfigSection{TracingEnabled: boolPtrSFN(true)},  // level 3
		cascade.StepFunctionsConfigSection{TracingEnabled: boolPtrSFN(false)}, // level 4
		zeroSFNCfg,
		zeroSFNCfg,
		zeroKropathSFN,
		zeroKropathSFN,
	)

	if got.Mandatory.TracingEnabled == nil || !*got.Mandatory.TracingEnabled {
		t.Errorf("AC-15: mandatory.tracingEnabled = %v, want *true (L3 must beat L4)", got.Mandatory.TracingEnabled)
	}
}

// TestMergeStepFunctionsCascade_AC16 — localStepFunctionsConfig.mandatory.tracingEnabled
// at level 4 propagates when level 3 is absent.
func TestMergeStepFunctionsCascade_AC16(t *testing.T) {
	got := mergeSFNAll(
		zeroKropathSFN,
		zeroKropathSFN,
		zeroSFNCfg,
		cascade.StepFunctionsConfigSection{TracingEnabled: boolPtrSFN(true)}, // level 4
		zeroSFNCfg,
		zeroSFNCfg,
		zeroKropathSFN,
		zeroKropathSFN,
	)

	if got.Mandatory.TracingEnabled == nil || !*got.Mandatory.TracingEnabled {
		t.Errorf("AC-16: mandatory.tracingEnabled = %v, want *true (level 4 when level 3 absent)", got.Mandatory.TracingEnabled)
	}
}

// TestMergeStepFunctionsCascade_AC17 — localStepFunctionsConfig.defaults.tracingEnabled
// at level 6 propagates to effectiveConfig.defaults.
func TestMergeStepFunctionsCascade_AC17(t *testing.T) {
	got := mergeSFNAll(
		zeroKropathSFN,
		zeroKropathSFN,
		zeroSFNCfg,
		zeroSFNCfg,
		cascade.StepFunctionsConfigSection{TracingEnabled: boolPtrSFN(false)}, // level 6
		zeroSFNCfg,
		zeroKropathSFN,
		zeroKropathSFN,
	)

	if got.Defaults.TracingEnabled == nil || *got.Defaults.TracingEnabled {
		t.Errorf("AC-17: defaults.tracingEnabled = %v, want *false", got.Defaults.TracingEnabled)
	}
	if got.Mandatory.TracingEnabled != nil {
		t.Errorf("AC-17: mandatory.tracingEnabled = %v, must not bleed from defaults", got.Mandatory.TracingEnabled)
	}
}

// TestMergeStepFunctionsCascade_AC18 — L6 wins over L7 for tracingEnabled in defaults tier.
func TestMergeStepFunctionsCascade_AC18(t *testing.T) {
	got := mergeSFNAll(
		zeroKropathSFN,
		zeroKropathSFN,
		zeroSFNCfg,
		zeroSFNCfg,
		cascade.StepFunctionsConfigSection{TracingEnabled: boolPtrSFN(true)},  // level 6
		cascade.StepFunctionsConfigSection{TracingEnabled: boolPtrSFN(false)}, // level 7
		zeroKropathSFN,
		zeroKropathSFN,
	)

	if got.Defaults.TracingEnabled == nil || !*got.Defaults.TracingEnabled {
		t.Errorf("AC-18: defaults.tracingEnabled = %v, want *true (L6 must beat L7)", got.Defaults.TracingEnabled)
	}
}

// --- includeExecutionData (*bool) cascade tests ---

// TestMergeStepFunctionsCascade_AC19 — globalStepFunctionsConfig.mandatory.includeExecutionData=true
// at level 3 propagates to effectiveConfig.mandatory.includeExecutionData.
func TestMergeStepFunctionsCascade_AC19(t *testing.T) {
	got := mergeSFNAll(
		zeroKropathSFN,
		zeroKropathSFN,
		cascade.StepFunctionsConfigSection{IncludeExecutionData: boolPtrSFN(true)}, // level 3
		zeroSFNCfg,
		zeroSFNCfg,
		zeroSFNCfg,
		zeroKropathSFN,
		zeroKropathSFN,
	)

	if got.Mandatory.IncludeExecutionData == nil || !*got.Mandatory.IncludeExecutionData {
		t.Errorf("AC-19: mandatory.includeExecutionData = %v, want *true", got.Mandatory.IncludeExecutionData)
	}
}

// TestMergeStepFunctionsCascade_AC20 — L3 wins over L4 for includeExecutionData.
func TestMergeStepFunctionsCascade_AC20(t *testing.T) {
	got := mergeSFNAll(
		zeroKropathSFN,
		zeroKropathSFN,
		cascade.StepFunctionsConfigSection{IncludeExecutionData: boolPtrSFN(false)}, // level 3
		cascade.StepFunctionsConfigSection{IncludeExecutionData: boolPtrSFN(true)},  // level 4
		zeroSFNCfg,
		zeroSFNCfg,
		zeroKropathSFN,
		zeroKropathSFN,
	)

	if got.Mandatory.IncludeExecutionData == nil || *got.Mandatory.IncludeExecutionData {
		t.Errorf("AC-20: mandatory.includeExecutionData = %v, want *false (L3 must beat L4)", got.Mandatory.IncludeExecutionData)
	}
}

// TestMergeStepFunctionsCascade_AC21 — globalStepFunctionsConfig.defaults.includeExecutionData
// at level 7 propagates to effectiveConfig.defaults when level 6 is absent.
func TestMergeStepFunctionsCascade_AC21(t *testing.T) {
	got := mergeSFNAll(
		zeroKropathSFN,
		zeroKropathSFN,
		zeroSFNCfg,
		zeroSFNCfg,
		zeroSFNCfg,
		cascade.StepFunctionsConfigSection{IncludeExecutionData: boolPtrSFN(false)}, // level 7
		zeroKropathSFN,
		zeroKropathSFN,
	)

	if got.Defaults.IncludeExecutionData == nil || *got.Defaults.IncludeExecutionData {
		t.Errorf("AC-21: defaults.includeExecutionData = %v, want *false (level 7 when level 6 absent)", got.Defaults.IncludeExecutionData)
	}
}

// --- tags cascade tests ---

// TestMergeStepFunctionsCascade_AC22 — mandatory tags are an additive union of all four
// mandatory sources; globalKropathConfig (L1) wins on key conflict.
func TestMergeStepFunctionsCascade_AC22(t *testing.T) {
	got := mergeSFNAll(
		cascade.StepFunctionsKropathSection{Tags: map[string]string{"cost-centre": "platform", "env": "org"}},  // level 1
		cascade.StepFunctionsKropathSection{Tags: map[string]string{"team": "infra", "env": "local-kropath"}},  // level 2
		cascade.StepFunctionsConfigSection{Tags: map[string]string{"owner": "workflows", "env": "global-sfn"}}, // level 3
		cascade.StepFunctionsConfigSection{Tags: map[string]string{"app": "myapp", "env": "local-sfn"}},        // level 4
		zeroSFNCfg,
		zeroSFNCfg,
		zeroKropathSFN,
		zeroKropathSFN,
	)

	cases := map[string]string{
		"cost-centre": "platform",  // from L1
		"env":         "org",       // L1 wins
		"team":        "infra",     // from L2
		"owner":       "workflows", // from L3
		"app":         "myapp",     // from L4
	}
	for k, want := range cases {
		if got := got.Mandatory.Tags[k]; got != want {
			t.Errorf("AC-22: mandatory.tags[%q] = %q, want %q", k, got, want)
		}
	}
	if len(got.Defaults.Tags) != 0 {
		t.Errorf("AC-22: defaults.tags must be empty, got %v", got.Defaults.Tags)
	}
}

// TestMergeStepFunctionsCascade_AC23 — defaults tags are an additive union of all four
// defaults sources; localStepFunctionsConfig (L6) wins on key conflict.
func TestMergeStepFunctionsCascade_AC23(t *testing.T) {
	got := mergeSFNAll(
		zeroKropathSFN,
		zeroKropathSFN,
		zeroSFNCfg,
		zeroSFNCfg,
		cascade.StepFunctionsConfigSection{Tags: map[string]string{"env": "local-sfn", "app": "myapp"}}, // level 6
		cascade.StepFunctionsConfigSection{Tags: map[string]string{"env": "global-sfn", "team": "sre"}}, // level 7
		cascade.StepFunctionsKropathSection{Tags: map[string]string{"cost-centre": "core"}},              // level 8
		cascade.StepFunctionsKropathSection{Tags: map[string]string{"org": "acme"}},                      // level 9
	)

	cases := map[string]string{
		"env":         "local-sfn", // L6 wins
		"app":         "myapp",     // from L6
		"team":        "sre",       // from L7
		"cost-centre": "core",      // from L8
		"org":         "acme",      // from L9
	}
	for k, want := range cases {
		if got := got.Defaults.Tags[k]; got != want {
			t.Errorf("AC-23: defaults.tags[%q] = %q, want %q", k, got, want)
		}
	}
}

// --- syncedLabels / syncedAnnotations cascade tests ---

// TestMergeStepFunctionsCascade_AC24 — syncedLabels additive union in mandatory tier from
// StepFunctionsConfig levels only; global (L3) wins on key conflict.
func TestMergeStepFunctionsCascade_AC24(t *testing.T) {
	got := mergeSFNAll(
		cascade.StepFunctionsKropathSection{Tags: map[string]string{"ignored": "yes"}}, // L1 — no syncedLabels in KropathConfig
		zeroKropathSFN,
		cascade.StepFunctionsConfigSection{SyncedLabels: map[string]string{"env": "prod", "team": "workflows"}}, // L3
		cascade.StepFunctionsConfigSection{SyncedLabels: map[string]string{"env": "staging", "app": "sfn"}},     // L4
		zeroSFNCfg,
		zeroSFNCfg,
		zeroKropathSFN,
		zeroKropathSFN,
	)

	cases := map[string]string{
		"env":  "prod",      // L3 wins
		"team": "workflows", // from L3
		"app":  "sfn",       // from L4
	}
	for k, want := range cases {
		if got := got.Mandatory.SyncedLabels[k]; got != want {
			t.Errorf("AC-24: mandatory.syncedLabels[%q] = %q, want %q", k, got, want)
		}
	}
}

// TestMergeStepFunctionsCascade_AC25 — syncedAnnotations additive union in mandatory tier
// from StepFunctionsConfig levels only; global (L3) wins on key conflict.
func TestMergeStepFunctionsCascade_AC25(t *testing.T) {
	got := mergeSFNAll(
		zeroKropathSFN,
		zeroKropathSFN,
		cascade.StepFunctionsConfigSection{SyncedAnnotations: map[string]string{"team": "platform", "env": "prod"}}, // L3
		cascade.StepFunctionsConfigSection{SyncedAnnotations: map[string]string{"team": "infra", "owner": "sre"}},   // L4
		zeroSFNCfg,
		zeroSFNCfg,
		zeroKropathSFN,
		zeroKropathSFN,
	)

	cases := map[string]string{
		"team":  "platform", // L3 wins
		"env":   "prod",     // from L3
		"owner": "sre",      // from L4
	}
	for k, want := range cases {
		if got := got.Mandatory.SyncedAnnotations[k]; got != want {
			t.Errorf("AC-25: mandatory.syncedAnnotations[%q] = %q, want %q", k, got, want)
		}
	}
}

// TestMergeStepFunctionsCascade_AC26 — syncedLabels additive union in defaults tier from
// StepFunctionsConfig levels only; local (L6) wins on key conflict.
func TestMergeStepFunctionsCascade_AC26(t *testing.T) {
	got := mergeSFNAll(
		zeroKropathSFN,
		zeroKropathSFN,
		zeroSFNCfg,
		zeroSFNCfg,
		cascade.StepFunctionsConfigSection{SyncedLabels: map[string]string{"tier": "standard", "env": "local"}}, // L6
		cascade.StepFunctionsConfigSection{SyncedLabels: map[string]string{"tier": "global", "team": "sre"}},    // L7
		zeroKropathSFN,
		zeroKropathSFN,
	)

	cases := map[string]string{
		"tier": "standard", // L6 wins
		"env":  "local",    // from L6
		"team": "sre",      // from L7
	}
	for k, want := range cases {
		if got := got.Defaults.SyncedLabels[k]; got != want {
			t.Errorf("AC-26: defaults.syncedLabels[%q] = %q, want %q", k, got, want)
		}
	}
}

// --- namingTemplate cascade tests ---

// TestMergeStepFunctionsCascade_AC27 — namingTemplate at level 3 propagates to mandatory only.
func TestMergeStepFunctionsCascade_AC27(t *testing.T) {
	const tmpl = "corp-{namespace}-{name}"
	got := mergeSFNAll(
		zeroKropathSFN,
		zeroKropathSFN,
		cascade.StepFunctionsConfigSection{NamingTemplate: tmpl}, // level 3
		zeroSFNCfg,
		zeroSFNCfg,
		zeroSFNCfg,
		zeroKropathSFN,
		zeroKropathSFN,
	)

	if got.Mandatory.NamingTemplate != tmpl {
		t.Errorf("AC-27: mandatory.namingTemplate = %q, want %q", got.Mandatory.NamingTemplate, tmpl)
	}
	if got.Defaults.NamingTemplate != "" {
		t.Errorf("AC-27: defaults.namingTemplate = %q, must not bleed from mandatory", got.Defaults.NamingTemplate)
	}
}

// TestMergeStepFunctionsCascade_AC28 — L3 wins over L4 for namingTemplate in mandatory tier.
func TestMergeStepFunctionsCascade_AC28(t *testing.T) {
	const globalTmpl = "corp-{namespace}-{name}"
	const localTmpl = "{namespace}-{name}"
	got := mergeSFNAll(
		zeroKropathSFN,
		zeroKropathSFN,
		cascade.StepFunctionsConfigSection{NamingTemplate: globalTmpl}, // level 3
		cascade.StepFunctionsConfigSection{NamingTemplate: localTmpl},  // level 4
		zeroSFNCfg,
		zeroSFNCfg,
		zeroKropathSFN,
		zeroKropathSFN,
	)

	if got.Mandatory.NamingTemplate != globalTmpl {
		t.Errorf("AC-28: mandatory.namingTemplate = %q, want %q (L3 must beat L4)", got.Mandatory.NamingTemplate, globalTmpl)
	}
}

// TestMergeStepFunctionsCascade_AC29 — namingTemplate at level 6 propagates to defaults only.
func TestMergeStepFunctionsCascade_AC29(t *testing.T) {
	const tmpl = "{namespace}-{name}"
	got := mergeSFNAll(
		zeroKropathSFN,
		zeroKropathSFN,
		zeroSFNCfg,
		zeroSFNCfg,
		cascade.StepFunctionsConfigSection{NamingTemplate: tmpl}, // level 6
		zeroSFNCfg,
		zeroKropathSFN,
		zeroKropathSFN,
	)

	if got.Defaults.NamingTemplate != tmpl {
		t.Errorf("AC-29: defaults.namingTemplate = %q, want %q", got.Defaults.NamingTemplate, tmpl)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("AC-29: mandatory.namingTemplate = %q, must not bleed from defaults", got.Mandatory.NamingTemplate)
	}
}

// --- tier isolation and all-zero tests ---

// TestMergeStepFunctionsCascade_AC30 — mandatory and defaults tiers do not bleed into
// each other when both are populated.
func TestMergeStepFunctionsCascade_AC30(t *testing.T) {
	got := mergeSFNAll(
		cascade.StepFunctionsKropathSection{LoggingLevel: "ALL"}, // level 1 — mandatory only
		zeroKropathSFN,
		zeroSFNCfg,
		zeroSFNCfg,
		zeroSFNCfg,
		zeroSFNCfg,
		zeroKropathSFN,
		cascade.StepFunctionsKropathSection{LoggingLevel: "OFF"}, // level 9 — defaults only
	)

	if got.Mandatory.LoggingLevel != "ALL" {
		t.Errorf("AC-30: mandatory.loggingLevel = %q, want %q", got.Mandatory.LoggingLevel, "ALL")
	}
	if got.Defaults.LoggingLevel != "OFF" {
		t.Errorf("AC-30: defaults.loggingLevel = %q, want %q", got.Defaults.LoggingLevel, "OFF")
	}
}

// TestMergeStepFunctionsCascade_AC31 — all-zero inputs yield zero effective config.
func TestMergeStepFunctionsCascade_AC31(t *testing.T) {
	got := mergeSFNAll(
		zeroKropathSFN, zeroKropathSFN,
		zeroSFNCfg, zeroSFNCfg, zeroSFNCfg, zeroSFNCfg,
		zeroKropathSFN, zeroKropathSFN,
	)

	if got.Mandatory.LoggingLevel != "" {
		t.Errorf("AC-31: mandatory.loggingLevel = %q, want empty", got.Mandatory.LoggingLevel)
	}
	if got.Mandatory.TracingEnabled != nil {
		t.Errorf("AC-31: mandatory.tracingEnabled = %v, want nil", got.Mandatory.TracingEnabled)
	}
	if got.Mandatory.IncludeExecutionData != nil {
		t.Errorf("AC-31: mandatory.includeExecutionData = %v, want nil", got.Mandatory.IncludeExecutionData)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("AC-31: mandatory.namingTemplate = %q, want empty", got.Mandatory.NamingTemplate)
	}
	if got.Defaults.LoggingLevel != "" {
		t.Errorf("AC-31: defaults.loggingLevel = %q, want empty", got.Defaults.LoggingLevel)
	}
	if got.Defaults.TracingEnabled != nil {
		t.Errorf("AC-31: defaults.tracingEnabled = %v, want nil", got.Defaults.TracingEnabled)
	}
	if got.Defaults.IncludeExecutionData != nil {
		t.Errorf("AC-31: defaults.includeExecutionData = %v, want nil", got.Defaults.IncludeExecutionData)
	}
	if got.Defaults.NamingTemplate != "" {
		t.Errorf("AC-31: defaults.namingTemplate = %q, want empty", got.Defaults.NamingTemplate)
	}
}
