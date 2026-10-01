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

var zeroKropathApigwv2 = cascade.ApiGatewayV2KropathSection{}
var zeroApigwv2Cfg = cascade.ApiGatewayV2ConfigSection{}

func mergeApigwv2All(
	globalKropathMandatory,
	localKropathMandatory cascade.ApiGatewayV2KropathSection,
	globalApigwv2CfgMandatory,
	localApigwv2CfgMandatory,
	localApigwv2CfgDefaults,
	globalApigwv2CfgDefaults cascade.ApiGatewayV2ConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.ApiGatewayV2KropathSection,
) cascade.EffectiveApiGatewayV2Config {
	return cascade.MergeApiGatewayV2Cascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalApigwv2CfgMandatory,
		localApigwv2CfgMandatory,
		localApigwv2CfgDefaults,
		globalApigwv2CfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeApiGatewayV2Cascade_AC1 — globalKropathConfig.mandatory.apigatewayv2.corsEnabled
// at level 1 propagates to effectiveConfig.mandatory.corsEnabled.
func TestMergeApiGatewayV2Cascade_AC1(t *testing.T) {
	got := mergeApigwv2All(
		cascade.ApiGatewayV2KropathSection{CorsEnabled: true}, // level 1
		zeroKropathApigwv2,
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		zeroKropathApigwv2,
		zeroKropathApigwv2,
	)

	if !got.Mandatory.CorsEnabled {
		t.Errorf("AC-1: mandatory.corsEnabled = false, want true (level 1)")
	}
	if got.Defaults.CorsEnabled {
		t.Errorf("AC-1: defaults.corsEnabled must not bleed from mandatory")
	}
}

// TestMergeApiGatewayV2Cascade_AC2 — globalApiGatewayV2Config.mandatory.corsEnabled at
// level 3 propagates when L1-L2 absent.
func TestMergeApiGatewayV2Cascade_AC2(t *testing.T) {
	got := mergeApigwv2All(
		zeroKropathApigwv2,
		zeroKropathApigwv2,
		cascade.ApiGatewayV2ConfigSection{CorsEnabled: true}, // level 3
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		zeroKropathApigwv2,
		zeroKropathApigwv2,
	)

	if !got.Mandatory.CorsEnabled {
		t.Errorf("AC-2: mandatory.corsEnabled = false, want true (level 3 when L1-L2 absent)")
	}
}

// TestMergeApiGatewayV2Cascade_AC3 — level 1 wins over level 3 for corsEnabled.
func TestMergeApiGatewayV2Cascade_AC3(t *testing.T) {
	got := mergeApigwv2All(
		cascade.ApiGatewayV2KropathSection{CorsEnabled: true},  // level 1
		zeroKropathApigwv2,
		cascade.ApiGatewayV2ConfigSection{CorsEnabled: false},  // level 3 (false)
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		zeroKropathApigwv2,
		zeroKropathApigwv2,
	)

	if !got.Mandatory.CorsEnabled {
		t.Errorf("AC-3: mandatory.corsEnabled = false, want true (level 1 beats level 3)")
	}
}

// TestMergeApiGatewayV2Cascade_AC4 — globalKropathConfig.mandatory.apigatewayv2.disableExecuteApiEndpoint
// at level 1 propagates; defaults tier stays false.
func TestMergeApiGatewayV2Cascade_AC4(t *testing.T) {
	got := mergeApigwv2All(
		cascade.ApiGatewayV2KropathSection{DisableExecuteApiEndpoint: true}, // level 1
		zeroKropathApigwv2,
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		zeroKropathApigwv2,
		zeroKropathApigwv2,
	)

	if !got.Mandatory.DisableExecuteApiEndpoint {
		t.Errorf("AC-4: mandatory.disableExecuteApiEndpoint = false, want true")
	}
	if got.Defaults.DisableExecuteApiEndpoint {
		t.Errorf("AC-4: defaults.disableExecuteApiEndpoint must not bleed from mandatory")
	}
}

// TestMergeApiGatewayV2Cascade_AC5 — globalKropathConfig.mandatory.apigatewayv2.minimumTlsVersion
// at level 1 propagates.
func TestMergeApiGatewayV2Cascade_AC5(t *testing.T) {
	got := mergeApigwv2All(
		cascade.ApiGatewayV2KropathSection{MinimumTlsVersion: "TLS_1_2"}, // level 1
		zeroKropathApigwv2,
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		zeroKropathApigwv2,
		zeroKropathApigwv2,
	)

	if got.Mandatory.MinimumTlsVersion != "TLS_1_2" {
		t.Errorf("AC-5: mandatory.minimumTlsVersion = %q, want TLS_1_2", got.Mandatory.MinimumTlsVersion)
	}
	if got.Defaults.MinimumTlsVersion != "" {
		t.Errorf("AC-5: defaults.minimumTlsVersion must not bleed from mandatory")
	}
}

// TestMergeApiGatewayV2Cascade_AC6 — localApiGatewayV2Config.defaults.minimumTlsVersion at
// level 6 propagates to effectiveConfig.defaults.
func TestMergeApiGatewayV2Cascade_AC6(t *testing.T) {
	got := mergeApigwv2All(
		zeroKropathApigwv2,
		zeroKropathApigwv2,
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		cascade.ApiGatewayV2ConfigSection{MinimumTlsVersion: "TLS_1_2"}, // level 6
		zeroApigwv2Cfg,
		zeroKropathApigwv2,
		zeroKropathApigwv2,
	)

	if got.Defaults.MinimumTlsVersion != "TLS_1_2" {
		t.Errorf("AC-6: defaults.minimumTlsVersion = %q, want TLS_1_2", got.Defaults.MinimumTlsVersion)
	}
	if got.Mandatory.MinimumTlsVersion != "" {
		t.Errorf("AC-6: mandatory.minimumTlsVersion must not bleed from defaults")
	}
}

// TestMergeApiGatewayV2Cascade_AC7 — level 6 wins over level 7 for defaults.minimumTlsVersion.
func TestMergeApiGatewayV2Cascade_AC7(t *testing.T) {
	got := mergeApigwv2All(
		zeroKropathApigwv2,
		zeroKropathApigwv2,
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		cascade.ApiGatewayV2ConfigSection{MinimumTlsVersion: "TLS_1_2"}, // level 6
		cascade.ApiGatewayV2ConfigSection{MinimumTlsVersion: "TLS_1_0"}, // level 7
		zeroKropathApigwv2,
		zeroKropathApigwv2,
	)

	if got.Defaults.MinimumTlsVersion != "TLS_1_2" {
		t.Errorf("AC-7: defaults.minimumTlsVersion = %q, want TLS_1_2 (level 6 beats level 7)", got.Defaults.MinimumTlsVersion)
	}
}

// TestMergeApiGatewayV2Cascade_AC8 — globalApiGatewayV2Config.mandatory.defaultThrottlingBurstLimit
// at level 3 propagates; localApiGatewayV2Config mandatory (L4) at lower priority.
func TestMergeApiGatewayV2Cascade_AC8(t *testing.T) {
	got := mergeApigwv2All(
		zeroKropathApigwv2,
		zeroKropathApigwv2,
		cascade.ApiGatewayV2ConfigSection{DefaultThrottlingBurstLimit: 1000}, // level 3
		cascade.ApiGatewayV2ConfigSection{DefaultThrottlingBurstLimit: 500},  // level 4
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		zeroKropathApigwv2,
		zeroKropathApigwv2,
	)

	if got.Mandatory.DefaultThrottlingBurstLimit != 1000 {
		t.Errorf("AC-8: mandatory.defaultThrottlingBurstLimit = %d, want 1000 (level 3 beats level 4)", got.Mandatory.DefaultThrottlingBurstLimit)
	}
}

// TestMergeApiGatewayV2Cascade_AC9 — globalApiGatewayV2Config.defaults.defaultThrottlingRateLimit
// at level 7 propagates when level 6 is zero.
func TestMergeApiGatewayV2Cascade_AC9(t *testing.T) {
	got := mergeApigwv2All(
		zeroKropathApigwv2,
		zeroKropathApigwv2,
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		cascade.ApiGatewayV2ConfigSection{DefaultThrottlingRateLimit: 500.0}, // level 7
		zeroKropathApigwv2,
		zeroKropathApigwv2,
	)

	if got.Defaults.DefaultThrottlingRateLimit != 500.0 {
		t.Errorf("AC-9: defaults.defaultThrottlingRateLimit = %f, want 500.0 (level 7 when L6 absent)", got.Defaults.DefaultThrottlingRateLimit)
	}
}

// TestMergeApiGatewayV2Cascade_AC10 — level 6 wins over level 7 for defaultThrottlingRateLimit.
func TestMergeApiGatewayV2Cascade_AC10(t *testing.T) {
	got := mergeApigwv2All(
		zeroKropathApigwv2,
		zeroKropathApigwv2,
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		cascade.ApiGatewayV2ConfigSection{DefaultThrottlingRateLimit: 200.0}, // level 6
		cascade.ApiGatewayV2ConfigSection{DefaultThrottlingRateLimit: 500.0}, // level 7
		zeroKropathApigwv2,
		zeroKropathApigwv2,
	)

	if got.Defaults.DefaultThrottlingRateLimit != 200.0 {
		t.Errorf("AC-10: defaults.defaultThrottlingRateLimit = %f, want 200.0 (level 6 beats level 7)", got.Defaults.DefaultThrottlingRateLimit)
	}
}

// TestMergeApiGatewayV2Cascade_AC11 — accessLogDestinationArn at level 3 mandatory propagates.
func TestMergeApiGatewayV2Cascade_AC11(t *testing.T) {
	const arn = "arn:aws:logs:us-east-1:123456789012:log-group:/kropath/apigw"
	got := mergeApigwv2All(
		zeroKropathApigwv2,
		zeroKropathApigwv2,
		cascade.ApiGatewayV2ConfigSection{AccessLogDestinationArn: arn}, // level 3
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		zeroKropathApigwv2,
		zeroKropathApigwv2,
	)

	if got.Mandatory.AccessLogDestinationArn != arn {
		t.Errorf("AC-11: mandatory.accessLogDestinationArn = %q, want %q", got.Mandatory.AccessLogDestinationArn, arn)
	}
	if got.Defaults.AccessLogDestinationArn != "" {
		t.Errorf("AC-11: defaults.accessLogDestinationArn must not bleed from mandatory")
	}
}

// TestMergeApiGatewayV2Cascade_AC12 — namingTemplate at level 6 defaults propagates.
func TestMergeApiGatewayV2Cascade_AC12(t *testing.T) {
	const tmpl = "{namespace}-{name}"
	got := mergeApigwv2All(
		zeroKropathApigwv2,
		zeroKropathApigwv2,
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		cascade.ApiGatewayV2ConfigSection{NamingTemplate: tmpl}, // level 6
		zeroApigwv2Cfg,
		zeroKropathApigwv2,
		zeroKropathApigwv2,
	)

	if got.Defaults.NamingTemplate != tmpl {
		t.Errorf("AC-12: defaults.namingTemplate = %q, want %q", got.Defaults.NamingTemplate, tmpl)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("AC-12: mandatory.namingTemplate must not bleed from defaults")
	}
}

// TestMergeApiGatewayV2Cascade_AC13 — level 3 wins over level 4 for namingTemplate.
func TestMergeApiGatewayV2Cascade_AC13(t *testing.T) {
	const globalTmpl = "corp-{namespace}-{name}"
	const localTmpl = "{namespace}-{name}"
	got := mergeApigwv2All(
		zeroKropathApigwv2,
		zeroKropathApigwv2,
		cascade.ApiGatewayV2ConfigSection{NamingTemplate: globalTmpl}, // level 3
		cascade.ApiGatewayV2ConfigSection{NamingTemplate: localTmpl},  // level 4
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		zeroKropathApigwv2,
		zeroKropathApigwv2,
	)

	if got.Mandatory.NamingTemplate != globalTmpl {
		t.Errorf("AC-13: mandatory.namingTemplate = %q, want %q (level 3 beats level 4)", got.Mandatory.NamingTemplate, globalTmpl)
	}
}

// TestMergeApiGatewayV2Cascade_AC14 — mandatory tags additive union from all four mandatory
// sources; globalKropathConfig (L1) wins on key conflict.
func TestMergeApiGatewayV2Cascade_AC14(t *testing.T) {
	got := mergeApigwv2All(
		cascade.ApiGatewayV2KropathSection{Tags: map[string]string{"env": "org", "cost-centre": "platform"}},        // level 1
		cascade.ApiGatewayV2KropathSection{Tags: map[string]string{"team": "infra", "env": "local-kropath"}},        // level 2
		cascade.ApiGatewayV2ConfigSection{Tags: map[string]string{"owner": "apigw", "env": "global-apigwv2cfg"}},    // level 3
		cascade.ApiGatewayV2ConfigSection{Tags: map[string]string{"app": "myapi", "env": "local-apigwv2cfg"}},       // level 4
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		zeroKropathApigwv2,
		zeroKropathApigwv2,
	)

	cases := map[string]string{
		"env":         "org",      // L1 wins
		"cost-centre": "platform", // from L1
		"team":        "infra",    // from L2
		"owner":       "apigw",    // from L3
		"app":         "myapi",    // from L4
	}
	for k, want := range cases {
		if v := got.Mandatory.Tags[k]; v != want {
			t.Errorf("AC-14: mandatory.tags[%q] = %q, want %q", k, v, want)
		}
	}
	if len(got.Defaults.Tags) != 0 {
		t.Errorf("AC-14: defaults.tags must be empty, got %v", got.Defaults.Tags)
	}
}

// TestMergeApiGatewayV2Cascade_AC15 — defaults tags additive union from all four defaults
// sources; localApiGatewayV2Config (L6) wins on key conflict.
func TestMergeApiGatewayV2Cascade_AC15(t *testing.T) {
	got := mergeApigwv2All(
		zeroKropathApigwv2,
		zeroKropathApigwv2,
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		cascade.ApiGatewayV2ConfigSection{Tags: map[string]string{"env": "local-apigwv2cfg", "app": "myapi"}}, // level 6
		cascade.ApiGatewayV2ConfigSection{Tags: map[string]string{"env": "global-apigwv2cfg", "team": "sre"}}, // level 7
		cascade.ApiGatewayV2KropathSection{Tags: map[string]string{"cost-centre": "core"}},                    // level 8
		cascade.ApiGatewayV2KropathSection{Tags: map[string]string{"org": "acme"}},                            // level 9
	)

	cases := map[string]string{
		"env":         "local-apigwv2cfg", // L6 wins
		"app":         "myapi",            // from L6
		"team":        "sre",              // from L7
		"cost-centre": "core",             // from L8
		"org":         "acme",             // from L9
	}
	for k, want := range cases {
		if v := got.Defaults.Tags[k]; v != want {
			t.Errorf("AC-15: defaults.tags[%q] = %q, want %q", k, v, want)
		}
	}
}

// TestMergeApiGatewayV2Cascade_AC16 — syncedLabels additive union in mandatory tier from
// ApiGatewayV2Config only; global (L3) wins on key conflict.
func TestMergeApiGatewayV2Cascade_AC16(t *testing.T) {
	got := mergeApigwv2All(
		cascade.ApiGatewayV2KropathSection{Tags: map[string]string{"ignored": "yes"}}, // L1 tags only
		zeroKropathApigwv2,
		cascade.ApiGatewayV2ConfigSection{SyncedLabels: map[string]string{"env": "prod", "team": "apigw"}}, // L3
		cascade.ApiGatewayV2ConfigSection{SyncedLabels: map[string]string{"env": "staging", "app": "api"}}, // L4
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		zeroKropathApigwv2,
		zeroKropathApigwv2,
	)

	cases := map[string]string{
		"env":  "prod",  // L3 wins
		"team": "apigw", // from L3
		"app":  "api",   // from L4
	}
	for k, want := range cases {
		if v := got.Mandatory.SyncedLabels[k]; v != want {
			t.Errorf("AC-16: mandatory.syncedLabels[%q] = %q, want %q", k, v, want)
		}
	}
}

// TestMergeApiGatewayV2Cascade_AC17 — syncedAnnotations additive union in defaults tier from
// ApiGatewayV2Config only; local (L6) wins on key conflict.
func TestMergeApiGatewayV2Cascade_AC17(t *testing.T) {
	got := mergeApigwv2All(
		zeroKropathApigwv2,
		zeroKropathApigwv2,
		zeroApigwv2Cfg,
		zeroApigwv2Cfg,
		cascade.ApiGatewayV2ConfigSection{SyncedAnnotations: map[string]string{"tier": "standard", "env": "local"}}, // L6
		cascade.ApiGatewayV2ConfigSection{SyncedAnnotations: map[string]string{"tier": "global", "team": "sre"}},    // L7
		zeroKropathApigwv2,
		zeroKropathApigwv2,
	)

	cases := map[string]string{
		"tier": "standard", // L6 wins
		"env":  "local",    // from L6
		"team": "sre",      // from L7
	}
	for k, want := range cases {
		if v := got.Defaults.SyncedAnnotations[k]; v != want {
			t.Errorf("AC-17: defaults.syncedAnnotations[%q] = %q, want %q", k, v, want)
		}
	}
}

// TestMergeApiGatewayV2Cascade_AC18 — all-zero inputs yield zero effective config.
func TestMergeApiGatewayV2Cascade_AC18(t *testing.T) {
	got := mergeApigwv2All(
		zeroKropathApigwv2, zeroKropathApigwv2,
		zeroApigwv2Cfg, zeroApigwv2Cfg, zeroApigwv2Cfg, zeroApigwv2Cfg,
		zeroKropathApigwv2, zeroKropathApigwv2,
	)

	if got.Mandatory.CorsEnabled {
		t.Errorf("AC-18: mandatory.corsEnabled should be false for all-zero inputs")
	}
	if got.Mandatory.MinimumTlsVersion != "" {
		t.Errorf("AC-18: mandatory.minimumTlsVersion = %q, want empty", got.Mandatory.MinimumTlsVersion)
	}
	if got.Mandatory.DefaultThrottlingBurstLimit != 0 {
		t.Errorf("AC-18: mandatory.defaultThrottlingBurstLimit = %d, want 0", got.Mandatory.DefaultThrottlingBurstLimit)
	}
	if got.Mandatory.DefaultThrottlingRateLimit != 0.0 {
		t.Errorf("AC-18: mandatory.defaultThrottlingRateLimit = %f, want 0.0", got.Mandatory.DefaultThrottlingRateLimit)
	}
	if got.Defaults.NamingTemplate != "" {
		t.Errorf("AC-18: defaults.namingTemplate = %q, want empty", got.Defaults.NamingTemplate)
	}
	if got.Defaults.CorsEnabled {
		t.Errorf("AC-18: defaults.corsEnabled should be false for all-zero inputs")
	}
}
