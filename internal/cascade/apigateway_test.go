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

var zeroKropathApigw = cascade.ApiGatewayKropathSection{}
var zeroApigwCfg = cascade.ApiGatewayConfigSection{}

func mergeApigwAll(
	globalKropathMandatory,
	localKropathMandatory cascade.ApiGatewayKropathSection,
	globalApigwCfgMandatory,
	localApigwCfgMandatory,
	localApigwCfgDefaults,
	globalApigwCfgDefaults cascade.ApiGatewayConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.ApiGatewayKropathSection,
) cascade.EffectiveApiGatewayConfig {
	return cascade.MergeApiGatewayCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalApigwCfgMandatory,
		localApigwCfgMandatory,
		localApigwCfgDefaults,
		globalApigwCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeApiGatewayCascade_AC1 — globalKropathConfig.mandatory.apigateway.endpointType
// at level 1 propagates to effectiveConfig.mandatory.endpointType.
func TestMergeApiGatewayCascade_AC1(t *testing.T) {
	got := mergeApigwAll(
		cascade.ApiGatewayKropathSection{EndpointType: "REGIONAL"}, // level 1
		zeroKropathApigw,
		zeroApigwCfg,
		zeroApigwCfg,
		zeroApigwCfg,
		zeroApigwCfg,
		zeroKropathApigw,
		zeroKropathApigw,
	)

	if got.Mandatory.EndpointType != "REGIONAL" {
		t.Errorf("AC-1: mandatory.endpointType = %q, want REGIONAL (level 1)", got.Mandatory.EndpointType)
	}
	if got.Defaults.EndpointType != "" {
		t.Errorf("AC-1: defaults.endpointType must not bleed from mandatory")
	}
}

// TestMergeApiGatewayCascade_AC2 — globalApiGatewayConfig.mandatory.endpointType at
// level 3 propagates when L1-L2 absent.
func TestMergeApiGatewayCascade_AC2(t *testing.T) {
	got := mergeApigwAll(
		zeroKropathApigw,
		zeroKropathApigw,
		cascade.ApiGatewayConfigSection{EndpointType: "EDGE"}, // level 3
		zeroApigwCfg,
		zeroApigwCfg,
		zeroApigwCfg,
		zeroKropathApigw,
		zeroKropathApigw,
	)

	if got.Mandatory.EndpointType != "EDGE" {
		t.Errorf("AC-2: mandatory.endpointType = %q, want EDGE (level 3 when L1-L2 absent)", got.Mandatory.EndpointType)
	}
}

// TestMergeApiGatewayCascade_AC3 — level 1 wins over level 3 for endpointType.
func TestMergeApiGatewayCascade_AC3(t *testing.T) {
	got := mergeApigwAll(
		cascade.ApiGatewayKropathSection{EndpointType: "REGIONAL"},  // level 1
		zeroKropathApigw,
		cascade.ApiGatewayConfigSection{EndpointType: "EDGE"},       // level 3
		zeroApigwCfg,
		zeroApigwCfg,
		zeroApigwCfg,
		zeroKropathApigw,
		zeroKropathApigw,
	)

	if got.Mandatory.EndpointType != "REGIONAL" {
		t.Errorf("AC-3: mandatory.endpointType = %q, want REGIONAL (level 1 wins over level 3)", got.Mandatory.EndpointType)
	}
}

// TestMergeApiGatewayCascade_AC4 — localApiGatewayConfig.defaults.endpointType propagates (level 6).
func TestMergeApiGatewayCascade_AC4(t *testing.T) {
	got := mergeApigwAll(
		zeroKropathApigw,
		zeroKropathApigw,
		zeroApigwCfg,
		zeroApigwCfg,
		cascade.ApiGatewayConfigSection{EndpointType: "REGIONAL"}, // level 6
		zeroApigwCfg,
		zeroKropathApigw,
		zeroKropathApigw,
	)

	if got.Defaults.EndpointType != "REGIONAL" {
		t.Errorf("AC-4: defaults.endpointType = %q, want REGIONAL (level 6)", got.Defaults.EndpointType)
	}
	if got.Mandatory.EndpointType != "" {
		t.Errorf("AC-4: mandatory.endpointType must not bleed from defaults")
	}
}

// TestMergeApiGatewayCascade_AC5 — Config-only mandatory field: apiKeySource from level 3.
func TestMergeApiGatewayCascade_AC5(t *testing.T) {
	got := mergeApigwAll(
		zeroKropathApigw,
		zeroKropathApigw,
		cascade.ApiGatewayConfigSection{ApiKeySource: "HEADER"}, // level 3
		zeroApigwCfg,
		zeroApigwCfg,
		zeroApigwCfg,
		zeroKropathApigw,
		zeroKropathApigw,
	)

	if got.Mandatory.ApiKeySource != "HEADER" {
		t.Errorf("AC-5: mandatory.apiKeySource = %q, want HEADER (level 3)", got.Mandatory.ApiKeySource)
	}
}

// TestMergeApiGatewayCascade_AC6 — disableExecuteApiEndpoint firstTrue from Config levels only.
func TestMergeApiGatewayCascade_AC6(t *testing.T) {
	got := mergeApigwAll(
		zeroKropathApigw,
		zeroKropathApigw,
		cascade.ApiGatewayConfigSection{DisableExecuteApiEndpoint: true}, // level 3
		zeroApigwCfg,
		zeroApigwCfg,
		zeroApigwCfg,
		zeroKropathApigw,
		zeroKropathApigw,
	)

	if !got.Mandatory.DisableExecuteApiEndpoint {
		t.Errorf("AC-6: mandatory.disableExecuteApiEndpoint = false, want true (level 3)")
	}
	if got.Defaults.DisableExecuteApiEndpoint {
		t.Errorf("AC-6: defaults.disableExecuteApiEndpoint must not bleed from mandatory")
	}
}

// TestMergeApiGatewayCascade_AC7 — tags union from mandatory levels; level 1 wins on conflict.
func TestMergeApiGatewayCascade_AC7(t *testing.T) {
	got := mergeApigwAll(
		cascade.ApiGatewayKropathSection{Tags: map[string]string{"env": "prod", "shared": "from-l1"}}, // level 1
		zeroKropathApigw,
		cascade.ApiGatewayConfigSection{Tags: map[string]string{"team": "platform", "shared": "from-l3"}}, // level 3
		zeroApigwCfg,
		zeroApigwCfg,
		zeroApigwCfg,
		zeroKropathApigw,
		zeroKropathApigw,
	)

	tags := got.Mandatory.Tags
	if tags["env"] != "prod" {
		t.Errorf("AC-7: tags[env] = %q, want prod", tags["env"])
	}
	if tags["team"] != "platform" {
		t.Errorf("AC-7: tags[team] = %q, want platform", tags["team"])
	}
	if tags["shared"] != "from-l1" {
		t.Errorf("AC-7: tags[shared] = %q, want from-l1 (level 1 wins)", tags["shared"])
	}
}

// TestMergeApiGatewayCascade_AC8 — syncedLabels come from Config levels only (not KropathConfig).
func TestMergeApiGatewayCascade_AC8(t *testing.T) {
	got := mergeApigwAll(
		zeroKropathApigw,
		zeroKropathApigw,
		cascade.ApiGatewayConfigSection{SyncedLabels: map[string]string{"data-class": "internal"}}, // level 3
		cascade.ApiGatewayConfigSection{SyncedLabels: map[string]string{"team": "payments"}},        // level 4
		zeroApigwCfg,
		zeroApigwCfg,
		zeroKropathApigw,
		zeroKropathApigw,
	)

	labels := got.Mandatory.SyncedLabels
	if labels["data-class"] != "internal" {
		t.Errorf("AC-8: syncedLabels[data-class] = %q, want internal", labels["data-class"])
	}
	if labels["team"] != "payments" {
		t.Errorf("AC-8: syncedLabels[team] = %q, want payments", labels["team"])
	}
}

// TestMergeApiGatewayCascade_AC9 — globalKropathDefaults.endpointType at level 9 is weakest defaults source.
func TestMergeApiGatewayCascade_AC9(t *testing.T) {
	got := mergeApigwAll(
		zeroKropathApigw,
		zeroKropathApigw,
		zeroApigwCfg,
		zeroApigwCfg,
		zeroApigwCfg,
		zeroApigwCfg,
		zeroKropathApigw,
		cascade.ApiGatewayKropathSection{EndpointType: "EDGE"}, // level 9
	)

	if got.Defaults.EndpointType != "EDGE" {
		t.Errorf("AC-9: defaults.endpointType = %q, want EDGE (level 9 when 6-8 absent)", got.Defaults.EndpointType)
	}
}

// TestMergeApiGatewayCascade_AC10 — level 6 wins over level 9 for defaults endpointType.
func TestMergeApiGatewayCascade_AC10(t *testing.T) {
	got := mergeApigwAll(
		zeroKropathApigw,
		zeroKropathApigw,
		zeroApigwCfg,
		zeroApigwCfg,
		cascade.ApiGatewayConfigSection{EndpointType: "REGIONAL"}, // level 6
		zeroApigwCfg,
		zeroKropathApigw,
		cascade.ApiGatewayKropathSection{EndpointType: "EDGE"}, // level 9
	)

	if got.Defaults.EndpointType != "REGIONAL" {
		t.Errorf("AC-10: defaults.endpointType = %q, want REGIONAL (level 6 wins over level 9)", got.Defaults.EndpointType)
	}
}

// TestMergeApiGatewayCascade_AC11 — zero inputs produce zero effective config.
func TestMergeApiGatewayCascade_AC11(t *testing.T) {
	got := mergeApigwAll(
		zeroKropathApigw, zeroKropathApigw,
		zeroApigwCfg, zeroApigwCfg, zeroApigwCfg, zeroApigwCfg,
		zeroKropathApigw, zeroKropathApigw,
	)

	if got.Mandatory.EndpointType != "" || got.Mandatory.DisableExecuteApiEndpoint || got.Mandatory.MinimumTlsVersion != "" {
		t.Errorf("AC-11: expected zero mandatory, got %+v", got.Mandatory)
	}
	if got.Defaults.EndpointType != "" || got.Defaults.DisableExecuteApiEndpoint {
		t.Errorf("AC-11: expected zero defaults, got %+v", got.Defaults)
	}
}

// TestMergeApiGatewayCascade_AC12 — namingTemplate is Config-only; KropathConfig cannot provide it.
func TestMergeApiGatewayCascade_AC12(t *testing.T) {
	got := mergeApigwAll(
		zeroKropathApigw,
		zeroKropathApigw,
		cascade.ApiGatewayConfigSection{NamingTemplate: "{namespace}-{name}-api"}, // level 3
		zeroApigwCfg,
		zeroApigwCfg,
		cascade.ApiGatewayConfigSection{NamingTemplate: "{name}-default"},          // level 7
		zeroKropathApigw,
		zeroKropathApigw,
	)

	if got.Mandatory.NamingTemplate != "{namespace}-{name}-api" {
		t.Errorf("AC-12: mandatory.namingTemplate = %q, want {namespace}-{name}-api", got.Mandatory.NamingTemplate)
	}
	if got.Defaults.NamingTemplate != "{name}-default" {
		t.Errorf("AC-12: defaults.namingTemplate = %q, want {name}-default", got.Defaults.NamingTemplate)
	}
}
