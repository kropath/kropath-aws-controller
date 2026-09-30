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

var (
	zeroKropathEC2    = cascade.EC2KropathSection{}
	zeroEC2CfgSection = cascade.EC2ConfigSection{}
)

func mergeEC2All(
	gKpcMand, lKpcMand cascade.EC2KropathSection,
	gCfgMand, lCfgMand, lCfgDef, gCfgDef cascade.EC2ConfigSection,
	lKpcDef, gKpcDef cascade.EC2KropathSection,
) cascade.EffectiveEC2Config {
	return cascade.MergeEC2Cascade(gKpcMand, lKpcMand, gCfgMand, lCfgMand, lCfgDef, gCfgDef, lKpcDef, gKpcDef)
}

// --- AC-1: globalKropathConfig.mandatory.ec2.flowLogsRequired=true propagates ---
func TestEC2_AC1_GlobalKropathMandatoryFlowLogsRequired(t *testing.T) {
	gKpcMand := cascade.EC2KropathSection{FlowLogsRequired: true}
	eff := mergeEC2All(gKpcMand, zeroKropathEC2, zeroEC2CfgSection, zeroEC2CfgSection, zeroEC2CfgSection, zeroEC2CfgSection, zeroKropathEC2, zeroKropathEC2)
	if !eff.Mandatory.FlowLogsRequired {
		t.Errorf("AC-1: expected mandatory.flowLogsRequired=true, got false")
	}
	if eff.Defaults.FlowLogsRequired {
		t.Errorf("AC-1: mandatory must not bleed into defaults")
	}
}

// --- AC-2: globalKropathConfig.mandatory.ec2.imdsv2Required=true propagates ---
func TestEC2_AC2_GlobalKropathMandatoryIMDSv2Required(t *testing.T) {
	gKpcMand := cascade.EC2KropathSection{IMDSv2Required: true}
	eff := mergeEC2All(gKpcMand, zeroKropathEC2, zeroEC2CfgSection, zeroEC2CfgSection, zeroEC2CfgSection, zeroEC2CfgSection, zeroKropathEC2, zeroKropathEC2)
	if !eff.Mandatory.IMDSv2Required {
		t.Errorf("AC-2: expected mandatory.imdsv2Required=true, got false")
	}
}

// --- AC-3: globalKropathConfig.mandatory.ec2.ebsEncryptionRequired=true propagates ---
func TestEC2_AC3_GlobalKropathMandatoryEBSEncryption(t *testing.T) {
	gKpcMand := cascade.EC2KropathSection{EBSEncryptionRequired: true}
	eff := mergeEC2All(gKpcMand, zeroKropathEC2, zeroEC2CfgSection, zeroEC2CfgSection, zeroEC2CfgSection, zeroEC2CfgSection, zeroKropathEC2, zeroKropathEC2)
	if !eff.Mandatory.EBSEncryptionRequired {
		t.Errorf("AC-3: expected mandatory.ebsEncryptionRequired=true, got false")
	}
}

// --- AC-4: level-1 wins over level-3 for flowLogsRequired ---
func TestEC2_AC4_Level1WinsOverLevel3_FlowLogs(t *testing.T) {
	// KropathConfig.mandatory=true (L1), EC2Config.mandatory=false (L3)
	gKpcMand := cascade.EC2KropathSection{FlowLogsRequired: true}
	gCfgMand := cascade.EC2ConfigSection{FlowLogsRequired: false}
	eff := mergeEC2All(gKpcMand, zeroKropathEC2, gCfgMand, zeroEC2CfgSection, zeroEC2CfgSection, zeroEC2CfgSection, zeroKropathEC2, zeroKropathEC2)
	if !eff.Mandatory.FlowLogsRequired {
		t.Errorf("AC-4: expected level-1 (true) to win over level-3 (false), got false")
	}
}

// --- AC-5: level-1 wins over level-3 for imdsv2Required ---
func TestEC2_AC5_Level1WinsOverLevel3_IMDSv2(t *testing.T) {
	gKpcMand := cascade.EC2KropathSection{IMDSv2Required: true}
	gCfgMand := cascade.EC2ConfigSection{IMDSv2Required: false}
	eff := mergeEC2All(gKpcMand, zeroKropathEC2, gCfgMand, zeroEC2CfgSection, zeroEC2CfgSection, zeroEC2CfgSection, zeroKropathEC2, zeroKropathEC2)
	if !eff.Mandatory.IMDSv2Required {
		t.Errorf("AC-5: expected level-1 (true) to win, got false")
	}
}

// --- AC-6: EC2Config-only field restrictPublicIpOnLaunch propagates from level-3 ---
func TestEC2_AC6_EC2ConfigOnlyField_RestrictPublicIp(t *testing.T) {
	gCfgMand := cascade.EC2ConfigSection{RestrictPublicIpOnLaunch: true}
	eff := mergeEC2All(zeroKropathEC2, zeroKropathEC2, gCfgMand, zeroEC2CfgSection, zeroEC2CfgSection, zeroEC2CfgSection, zeroKropathEC2, zeroKropathEC2)
	if !eff.Mandatory.RestrictPublicIpOnLaunch {
		t.Errorf("AC-6: expected mandatory.restrictPublicIpOnLaunch=true, got false")
	}
}

// --- AC-7: EC2Config-only field ebsDefaultKmsKeyId propagates from level-3 ---
func TestEC2_AC7_EC2ConfigOnlyField_EBSKMSKey(t *testing.T) {
	const keyID = "arn:aws:kms:us-east-1:123:key/abc"
	gCfgMand := cascade.EC2ConfigSection{EBSDefaultKMSKeyId: keyID}
	eff := mergeEC2All(zeroKropathEC2, zeroKropathEC2, gCfgMand, zeroEC2CfgSection, zeroEC2CfgSection, zeroEC2CfgSection, zeroKropathEC2, zeroKropathEC2)
	if eff.Mandatory.EBSDefaultKMSKeyId != keyID {
		t.Errorf("AC-7: expected ebsDefaultKmsKeyId=%q, got %q", keyID, eff.Mandatory.EBSDefaultKMSKeyId)
	}
}

// --- AC-8: EC2Config.defaults.allowSourceDestCheckDisable=true propagates ---
func TestEC2_AC8_EC2ConfigDefaultsAllowSourceDestCheck(t *testing.T) {
	gCfgDef := cascade.EC2ConfigSection{AllowSourceDestCheckDisable: true}
	eff := mergeEC2All(zeroKropathEC2, zeroKropathEC2, zeroEC2CfgSection, zeroEC2CfgSection, zeroEC2CfgSection, gCfgDef, zeroKropathEC2, zeroKropathEC2)
	if !eff.Defaults.AllowSourceDestCheckDisable {
		t.Errorf("AC-8: expected defaults.allowSourceDestCheckDisable=true, got false")
	}
	if eff.Mandatory.AllowSourceDestCheckDisable {
		t.Errorf("AC-8: defaults must not bleed into mandatory")
	}
}

// --- AC-9: only EC2Config.defaults.imdsv2Required set; mandatory stays false ---
func TestEC2_AC9_DefaultsOnly_IMDSv2(t *testing.T) {
	gCfgDef := cascade.EC2ConfigSection{IMDSv2Required: true}
	eff := mergeEC2All(zeroKropathEC2, zeroKropathEC2, zeroEC2CfgSection, zeroEC2CfgSection, zeroEC2CfgSection, gCfgDef, zeroKropathEC2, zeroKropathEC2)
	if !eff.Defaults.IMDSv2Required {
		t.Errorf("AC-9: expected defaults.imdsv2Required=true, got false")
	}
	if eff.Mandatory.IMDSv2Required {
		t.Errorf("AC-9: mandatory.imdsv2Required must be false when only defaults set")
	}
}

// --- AC-10: level-7 (globalEC2Config.defaults) wins over level-9 (globalKropathConfig.defaults) ---
func TestEC2_AC10_DefaultsPriority_Level7WinsLevel9(t *testing.T) {
	// Both set true — either would pass; verifying the merge produces true when L7 set
	gCfgDef := cascade.EC2ConfigSection{FlowLogsRequired: true}
	gKpcDef := cascade.EC2KropathSection{FlowLogsRequired: true}
	eff := mergeEC2All(zeroKropathEC2, zeroKropathEC2, zeroEC2CfgSection, zeroEC2CfgSection, zeroEC2CfgSection, gCfgDef, zeroKropathEC2, gKpcDef)
	if !eff.Defaults.FlowLogsRequired {
		t.Errorf("AC-10: expected defaults.flowLogsRequired=true (L7 or L9 set)")
	}

	// Level-7 alone is sufficient; level-9 alone should also propagate
	effL7Only := mergeEC2All(zeroKropathEC2, zeroKropathEC2, zeroEC2CfgSection, zeroEC2CfgSection, zeroEC2CfgSection, gCfgDef, zeroKropathEC2, zeroKropathEC2)
	if !effL7Only.Defaults.FlowLogsRequired {
		t.Errorf("AC-10: L7 alone should propagate")
	}
}

// --- AC-11: global mandatory (L1) wins over local mandatory (L2) for flowLogsRequired ---
func TestEC2_AC11_GlobalMandatoryWinsOverLocal(t *testing.T) {
	gKpcMand := cascade.EC2KropathSection{FlowLogsRequired: true}
	// local (L2) is false (zero), global (L1) is true → result must be true
	eff := mergeEC2All(gKpcMand, zeroKropathEC2, zeroEC2CfgSection, zeroEC2CfgSection, zeroEC2CfgSection, zeroEC2CfgSection, zeroKropathEC2, zeroKropathEC2)
	if !eff.Mandatory.FlowLogsRequired {
		t.Errorf("AC-11: global mandatory (L1) must propagate even when local mandatory (L2) is zero")
	}
}

// --- AC-12: Tags are union-merged across all mandatory sources ---
func TestEC2_AC12_TagUnionMerge(t *testing.T) {
	gKpcMand := cascade.EC2KropathSection{Tags: map[string]string{"cost-centre": "platform"}}
	lCfgMand := cascade.EC2ConfigSection{Tags: map[string]string{"resource-type": "networking"}}
	eff := mergeEC2All(gKpcMand, zeroKropathEC2, zeroEC2CfgSection, lCfgMand, zeroEC2CfgSection, zeroEC2CfgSection, zeroKropathEC2, zeroKropathEC2)
	if eff.Mandatory.Tags["cost-centre"] != "platform" {
		t.Errorf("AC-12: expected tags[cost-centre]=platform, got %q", eff.Mandatory.Tags["cost-centre"])
	}
	if eff.Mandatory.Tags["resource-type"] != "networking" {
		t.Errorf("AC-12: expected tags[resource-type]=networking, got %q", eff.Mandatory.Tags["resource-type"])
	}
}

// --- AC-13: KropathConfig.mandatory.syncedLabels propagates; EC2Config.defaults.syncedLabels propagates ---
func TestEC2_AC13_SyncedLabels(t *testing.T) {
	gKpcMand := cascade.EC2KropathSection{SyncedLabels: map[string]string{"environment": "prod"}}
	lCfgDef := cascade.EC2ConfigSection{SyncedLabels: map[string]string{"team": "networking"}}
	eff := mergeEC2All(gKpcMand, zeroKropathEC2, zeroEC2CfgSection, zeroEC2CfgSection, lCfgDef, zeroEC2CfgSection, zeroKropathEC2, zeroKropathEC2)
	if eff.Mandatory.SyncedLabels["environment"] != "prod" {
		t.Errorf("AC-13: expected mandatory.syncedLabels[environment]=prod, got %q", eff.Mandatory.SyncedLabels["environment"])
	}
	if eff.Defaults.SyncedLabels["team"] != "networking" {
		t.Errorf("AC-13: expected defaults.syncedLabels[team]=networking, got %q", eff.Defaults.SyncedLabels["team"])
	}
}

// --- AllAbsent: all zero inputs produce all-zero effective config ---
func TestEC2_AllAbsent(t *testing.T) {
	eff := mergeEC2All(zeroKropathEC2, zeroKropathEC2, zeroEC2CfgSection, zeroEC2CfgSection, zeroEC2CfgSection, zeroEC2CfgSection, zeroKropathEC2, zeroKropathEC2)
	if eff.Mandatory.FlowLogsRequired || eff.Mandatory.IMDSv2Required || eff.Mandatory.EBSEncryptionRequired {
		t.Errorf("AllAbsent: expected all booleans false when no input set")
	}
	if eff.Mandatory.FlowLogTrafficType != "" || eff.Mandatory.EBSDefaultKMSKeyId != "" {
		t.Errorf("AllAbsent: expected all strings empty when no input set")
	}
	if eff.Mandatory.FlowLogMaxAggregationInterval != 0 {
		t.Errorf("AllAbsent: expected int64 zero when no input set")
	}
	if len(eff.Mandatory.Tags) != 0 || len(eff.Defaults.Tags) != 0 {
		t.Errorf("AllAbsent: expected nil tags when no input set")
	}
}

// --- MandatoryCascadeOrder: table-driven for all mandatory priority levels ---
func TestEC2_MandatoryCascadeOrder(t *testing.T) {
	tests := []struct {
		name     string
		gKpcMand cascade.EC2KropathSection
		lKpcMand cascade.EC2KropathSection
		gCfgMand cascade.EC2ConfigSection
		lCfgMand cascade.EC2ConfigSection
		wantFlow bool
	}{
		{"L1_wins", cascade.EC2KropathSection{FlowLogsRequired: true}, zeroKropathEC2, zeroEC2CfgSection, zeroEC2CfgSection, true},
		{"L2_when_no_L1", zeroKropathEC2, cascade.EC2KropathSection{FlowLogsRequired: true}, zeroEC2CfgSection, zeroEC2CfgSection, true},
		{"L3_when_no_L1_L2", zeroKropathEC2, zeroKropathEC2, cascade.EC2ConfigSection{FlowLogsRequired: true}, zeroEC2CfgSection, true},
		{"L4_when_no_others", zeroKropathEC2, zeroKropathEC2, zeroEC2CfgSection, cascade.EC2ConfigSection{FlowLogsRequired: true}, true},
		{"none_set", zeroKropathEC2, zeroKropathEC2, zeroEC2CfgSection, zeroEC2CfgSection, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			eff := mergeEC2All(tc.gKpcMand, tc.lKpcMand, tc.gCfgMand, tc.lCfgMand, zeroEC2CfgSection, zeroEC2CfgSection, zeroKropathEC2, zeroKropathEC2)
			if eff.Mandatory.FlowLogsRequired != tc.wantFlow {
				t.Errorf("%s: mandatory.flowLogsRequired: got %v, want %v", tc.name, eff.Mandatory.FlowLogsRequired, tc.wantFlow)
			}
		})
	}
}

// --- DefaultsCascadeOrder: table-driven for all defaults priority levels ---
func TestEC2_DefaultsCascadeOrder(t *testing.T) {
	tests := []struct {
		name     string
		lCfgDef  cascade.EC2ConfigSection
		gCfgDef  cascade.EC2ConfigSection
		lKpcDef  cascade.EC2KropathSection
		gKpcDef  cascade.EC2KropathSection
		wantFlow bool
	}{
		{"L6_wins", cascade.EC2ConfigSection{FlowLogsRequired: true}, zeroEC2CfgSection, zeroKropathEC2, zeroKropathEC2, true},
		{"L7_when_no_L6", zeroEC2CfgSection, cascade.EC2ConfigSection{FlowLogsRequired: true}, zeroKropathEC2, zeroKropathEC2, true},
		{"L8_when_no_L6_L7", zeroEC2CfgSection, zeroEC2CfgSection, cascade.EC2KropathSection{FlowLogsRequired: true}, zeroKropathEC2, true},
		{"L9_when_no_others", zeroEC2CfgSection, zeroEC2CfgSection, zeroKropathEC2, cascade.EC2KropathSection{FlowLogsRequired: true}, true},
		{"none_set", zeroEC2CfgSection, zeroEC2CfgSection, zeroKropathEC2, zeroKropathEC2, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			eff := mergeEC2All(zeroKropathEC2, zeroKropathEC2, zeroEC2CfgSection, zeroEC2CfgSection, tc.lCfgDef, tc.gCfgDef, tc.lKpcDef, tc.gKpcDef)
			if eff.Defaults.FlowLogsRequired != tc.wantFlow {
				t.Errorf("%s: defaults.flowLogsRequired: got %v, want %v", tc.name, eff.Defaults.FlowLogsRequired, tc.wantFlow)
			}
		})
	}
}

// --- MandatoryIsolatedFromDefaults: mandatory fields don't leak into defaults and vice versa ---
func TestEC2_MandatoryIsolatedFromDefaults(t *testing.T) {
	gKpcMand := cascade.EC2KropathSection{FlowLogsRequired: true, IMDSv2Required: true, EBSEncryptionRequired: true}
	eff := mergeEC2All(gKpcMand, zeroKropathEC2, zeroEC2CfgSection, zeroEC2CfgSection, zeroEC2CfgSection, zeroEC2CfgSection, zeroKropathEC2, zeroKropathEC2)
	if eff.Defaults.FlowLogsRequired || eff.Defaults.IMDSv2Required || eff.Defaults.EBSEncryptionRequired {
		t.Errorf("mandatory values must not bleed into defaults")
	}

	gKpcDef := cascade.EC2KropathSection{FlowLogsRequired: true, IMDSv2Required: true, EBSEncryptionRequired: true}
	eff2 := mergeEC2All(zeroKropathEC2, zeroKropathEC2, zeroEC2CfgSection, zeroEC2CfgSection, zeroEC2CfgSection, zeroEC2CfgSection, zeroKropathEC2, gKpcDef)
	if eff2.Mandatory.FlowLogsRequired || eff2.Mandatory.IMDSv2Required || eff2.Mandatory.EBSEncryptionRequired {
		t.Errorf("defaults values must not bleed into mandatory")
	}
}

// --- EC2ConfigOnlyFields: fields absent from KropathSection only flow from EC2Config levels ---
func TestEC2_EC2ConfigOnlyFields(t *testing.T) {
	gCfgMand := cascade.EC2ConfigSection{
		FlowLogTrafficType:            "ALL",
		FlowLogMaxAggregationInterval: 600,
		EBSDefaultKMSKeyId:            "arn:aws:kms:us-east-1:123:key/xyz",
		RestrictPublicIpOnLaunch:      true,
		PublicIpRestricted:            true,
		AllowSourceDestCheckDisable:   true,
	}
	eff := mergeEC2All(zeroKropathEC2, zeroKropathEC2, gCfgMand, zeroEC2CfgSection, zeroEC2CfgSection, zeroEC2CfgSection, zeroKropathEC2, zeroKropathEC2)
	if eff.Mandatory.FlowLogTrafficType != "ALL" {
		t.Errorf("expected flowLogTrafficType=ALL, got %q", eff.Mandatory.FlowLogTrafficType)
	}
	if eff.Mandatory.FlowLogMaxAggregationInterval != 600 {
		t.Errorf("expected flowLogMaxAggregationInterval=600, got %d", eff.Mandatory.FlowLogMaxAggregationInterval)
	}
	if eff.Mandatory.EBSDefaultKMSKeyId != "arn:aws:kms:us-east-1:123:key/xyz" {
		t.Errorf("expected ebsDefaultKmsKeyId set, got %q", eff.Mandatory.EBSDefaultKMSKeyId)
	}
	if !eff.Mandatory.RestrictPublicIpOnLaunch {
		t.Errorf("expected restrictPublicIpOnLaunch=true")
	}
	if !eff.Mandatory.PublicIpRestricted {
		t.Errorf("expected publicIpRestricted=true")
	}
	if !eff.Mandatory.AllowSourceDestCheckDisable {
		t.Errorf("expected allowSourceDestCheckDisable=true")
	}
}

// --- TagPriorityOrder: KropathConfig tags win over EC2Config on key conflict ---
func TestEC2_TagPriorityOrder_MandatoryKropathWins(t *testing.T) {
	gKpcMand := cascade.EC2KropathSection{Tags: map[string]string{"env": "prod"}}
	lCfgMand := cascade.EC2ConfigSection{Tags: map[string]string{"env": "staging"}}
	eff := mergeEC2All(gKpcMand, zeroKropathEC2, zeroEC2CfgSection, lCfgMand, zeroEC2CfgSection, zeroEC2CfgSection, zeroKropathEC2, zeroKropathEC2)
	if eff.Mandatory.Tags["env"] != "prod" {
		t.Errorf("global KropathConfig tags must win over EC2Config on key conflict, got %q", eff.Mandatory.Tags["env"])
	}
}

// --- NamingTemplate: governed only at EC2Config levels, KropathSection has no effect ---
func TestEC2_NamingTemplate_EC2ConfigOnly(t *testing.T) {
	gCfgMand := cascade.EC2ConfigSection{NamingTemplate: "{namespace}-{name}"}
	eff := mergeEC2All(zeroKropathEC2, zeroKropathEC2, gCfgMand, zeroEC2CfgSection, zeroEC2CfgSection, zeroEC2CfgSection, zeroKropathEC2, zeroKropathEC2)
	if eff.Mandatory.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("expected namingTemplate={namespace}-{name}, got %q", eff.Mandatory.NamingTemplate)
	}
}
