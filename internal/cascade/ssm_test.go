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

// emptySSMKropath returns a zero-value SSMKropathSection (no fields set).
func emptySSMKropath() cascade.SSMKropathSection { return cascade.SSMKropathSection{} }

// emptySSMCfg returns a zero-value SSMConfigSection (no fields set).
func emptySSMCfg() cascade.SSMConfigSection { return cascade.SSMConfigSection{} }

func TestMergeSSMCascade_L1_GlobalKropathMandatory(t *testing.T) {
	globalKPC := cascade.SSMKropathSection{
		ParameterType: "SecureString",
		ParameterTier: "Advanced",
		KeyID:         "alias/my-key",
	}
	eff := cascade.MergeSSMCascade(
		globalKPC, emptySSMKropath(),
		emptySSMCfg(), emptySSMCfg(),
		emptySSMCfg(), emptySSMCfg(),
		emptySSMKropath(), emptySSMKropath(),
	)
	if eff.Mandatory.Type != "SecureString" {
		t.Errorf("mandatory.type: got %q, want %q", eff.Mandatory.Type, "SecureString")
	}
	if eff.Mandatory.Tier != "Advanced" {
		t.Errorf("mandatory.tier: got %q, want %q", eff.Mandatory.Tier, "Advanced")
	}
	if eff.Mandatory.KeyID != "alias/my-key" {
		t.Errorf("mandatory.keyID: got %q, want %q", eff.Mandatory.KeyID, "alias/my-key")
	}
}

func TestMergeSSMCascade_L2_LocalKropathMandatory_OverridesGlobalKropathWhenGlobalAbsent(t *testing.T) {
	localKPC := cascade.SSMKropathSection{ParameterType: "StringList"}
	eff := cascade.MergeSSMCascade(
		emptySSMKropath(), localKPC,
		emptySSMCfg(), emptySSMCfg(),
		emptySSMCfg(), emptySSMCfg(),
		emptySSMKropath(), emptySSMKropath(),
	)
	if eff.Mandatory.Type != "StringList" {
		t.Errorf("mandatory.type: got %q, want %q", eff.Mandatory.Type, "StringList")
	}
}

func TestMergeSSMCascade_L1_Wins_Over_L2(t *testing.T) {
	globalKPC := cascade.SSMKropathSection{ParameterType: "SecureString"}
	localKPC := cascade.SSMKropathSection{ParameterType: "String"}
	eff := cascade.MergeSSMCascade(
		globalKPC, localKPC,
		emptySSMCfg(), emptySSMCfg(),
		emptySSMCfg(), emptySSMCfg(),
		emptySSMKropath(), emptySSMKropath(),
	)
	if eff.Mandatory.Type != "SecureString" {
		t.Errorf("L1 must win over L2: got %q, want %q", eff.Mandatory.Type, "SecureString")
	}
}

func TestMergeSSMCascade_L3_GlobalSSMCfgMandatory(t *testing.T) {
	globalSSMCfg := cascade.SSMConfigSection{
		Type:         "SecureString",
		DocumentType: "Command",
		OperatingSystem: "AMAZON_LINUX_2",
	}
	eff := cascade.MergeSSMCascade(
		emptySSMKropath(), emptySSMKropath(),
		globalSSMCfg, emptySSMCfg(),
		emptySSMCfg(), emptySSMCfg(),
		emptySSMKropath(), emptySSMKropath(),
	)
	if eff.Mandatory.Type != "SecureString" {
		t.Errorf("mandatory.type from L3: got %q, want %q", eff.Mandatory.Type, "SecureString")
	}
	if eff.Mandatory.DocumentType != "Command" {
		t.Errorf("mandatory.documentType from L3: got %q, want %q", eff.Mandatory.DocumentType, "Command")
	}
	if eff.Mandatory.OperatingSystem != "AMAZON_LINUX_2" {
		t.Errorf("mandatory.operatingSystem from L3: got %q, want %q", eff.Mandatory.OperatingSystem, "AMAZON_LINUX_2")
	}
}

func TestMergeSSMCascade_L4_LocalSSMCfgMandatory(t *testing.T) {
	localSSMCfg := cascade.SSMConfigSection{
		Type:         "String",
		DocumentType: "Session",
	}
	eff := cascade.MergeSSMCascade(
		emptySSMKropath(), emptySSMKropath(),
		emptySSMCfg(), localSSMCfg,
		emptySSMCfg(), emptySSMCfg(),
		emptySSMKropath(), emptySSMKropath(),
	)
	if eff.Mandatory.Type != "String" {
		t.Errorf("mandatory.type from L4: got %q, want %q", eff.Mandatory.Type, "String")
	}
	if eff.Mandatory.DocumentType != "Session" {
		t.Errorf("mandatory.documentType from L4: got %q, want %q", eff.Mandatory.DocumentType, "Session")
	}
}

func TestMergeSSMCascade_L1_Wins_Over_L4(t *testing.T) {
	globalKPC := cascade.SSMKropathSection{ParameterType: "SecureString"}
	localSSMCfg := cascade.SSMConfigSection{Type: "String"}
	eff := cascade.MergeSSMCascade(
		globalKPC, emptySSMKropath(),
		emptySSMCfg(), localSSMCfg,
		emptySSMCfg(), emptySSMCfg(),
		emptySSMKropath(), emptySSMKropath(),
	)
	if eff.Mandatory.Type != "SecureString" {
		t.Errorf("L1 must win over L4: got %q, want %q", eff.Mandatory.Type, "SecureString")
	}
}

func TestMergeSSMCascade_L6_LocalSSMCfgDefaults(t *testing.T) {
	localDefaults := cascade.SSMConfigSection{
		Type: "String",
		Tier: "Standard",
	}
	eff := cascade.MergeSSMCascade(
		emptySSMKropath(), emptySSMKropath(),
		emptySSMCfg(), emptySSMCfg(),
		localDefaults, emptySSMCfg(),
		emptySSMKropath(), emptySSMKropath(),
	)
	if eff.Defaults.Type != "String" {
		t.Errorf("defaults.type from L6: got %q, want %q", eff.Defaults.Type, "String")
	}
	if eff.Defaults.Tier != "Standard" {
		t.Errorf("defaults.tier from L6: got %q, want %q", eff.Defaults.Tier, "Standard")
	}
}

func TestMergeSSMCascade_L7_GlobalSSMCfgDefaults(t *testing.T) {
	globalDefaults := cascade.SSMConfigSection{Tier: "Advanced"}
	eff := cascade.MergeSSMCascade(
		emptySSMKropath(), emptySSMKropath(),
		emptySSMCfg(), emptySSMCfg(),
		emptySSMCfg(), globalDefaults,
		emptySSMKropath(), emptySSMKropath(),
	)
	if eff.Defaults.Tier != "Advanced" {
		t.Errorf("defaults.tier from L7: got %q, want %q", eff.Defaults.Tier, "Advanced")
	}
}

func TestMergeSSMCascade_L6_Wins_Over_L7(t *testing.T) {
	localDefaults := cascade.SSMConfigSection{Tier: "Standard"}
	globalDefaults := cascade.SSMConfigSection{Tier: "Advanced"}
	eff := cascade.MergeSSMCascade(
		emptySSMKropath(), emptySSMKropath(),
		emptySSMCfg(), emptySSMCfg(),
		localDefaults, globalDefaults,
		emptySSMKropath(), emptySSMKropath(),
	)
	if eff.Defaults.Tier != "Standard" {
		t.Errorf("L6 must win over L7: got %q, want %q", eff.Defaults.Tier, "Standard")
	}
}

func TestMergeSSMCascade_L8_LocalKropathDefaults(t *testing.T) {
	localKPCDefaults := cascade.SSMKropathSection{ParameterType: "String"}
	eff := cascade.MergeSSMCascade(
		emptySSMKropath(), emptySSMKropath(),
		emptySSMCfg(), emptySSMCfg(),
		emptySSMCfg(), emptySSMCfg(),
		localKPCDefaults, emptySSMKropath(),
	)
	if eff.Defaults.Type != "String" {
		t.Errorf("defaults.type from L8: got %q, want %q", eff.Defaults.Type, "String")
	}
}

func TestMergeSSMCascade_L9_GlobalKropathDefaults(t *testing.T) {
	globalKPCDefaults := cascade.SSMKropathSection{ParameterType: "SecureString"}
	eff := cascade.MergeSSMCascade(
		emptySSMKropath(), emptySSMKropath(),
		emptySSMCfg(), emptySSMCfg(),
		emptySSMCfg(), emptySSMCfg(),
		emptySSMKropath(), globalKPCDefaults,
	)
	if eff.Defaults.Type != "SecureString" {
		t.Errorf("defaults.type from L9: got %q, want %q", eff.Defaults.Type, "SecureString")
	}
}

func TestMergeSSMCascade_L6_Wins_Over_L9(t *testing.T) {
	localDefaults := cascade.SSMConfigSection{Type: "String"}
	globalKPCDefaults := cascade.SSMKropathSection{ParameterType: "SecureString"}
	eff := cascade.MergeSSMCascade(
		emptySSMKropath(), emptySSMKropath(),
		emptySSMCfg(), emptySSMCfg(),
		localDefaults, emptySSMCfg(),
		emptySSMKropath(), globalKPCDefaults,
	)
	if eff.Defaults.Type != "String" {
		t.Errorf("L6 must win over L9: got %q, want %q", eff.Defaults.Type, "String")
	}
}

// TestMergeSSMCascade_BoolNilSentinel verifies that nil approvedPatchesEnableNonSecurity
// falls through to the next non-nil value, and explicit false propagates correctly.
func TestMergeSSMCascade_BoolNilSentinel(t *testing.T) {
	falseBool := false
	trueBool := true

	tests := []struct {
		name           string
		l3NonSecurity  *bool
		l4NonSecurity  *bool
		wantMandatory  *bool
	}{
		{
			name:          "nil_falls_through_to_L4",
			l3NonSecurity: nil,
			l4NonSecurity: &trueBool,
			wantMandatory: &trueBool,
		},
		{
			name:          "L3_false_wins_over_L4_true",
			l3NonSecurity: &falseBool,
			l4NonSecurity: &trueBool,
			wantMandatory: &falseBool,
		},
		{
			name:          "both_nil_yields_nil",
			l3NonSecurity: nil,
			l4NonSecurity: nil,
			wantMandatory: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			globalSSMCfg := cascade.SSMConfigSection{ApprovedPatchesEnableNonSecurity: tc.l3NonSecurity}
			localSSMCfg := cascade.SSMConfigSection{ApprovedPatchesEnableNonSecurity: tc.l4NonSecurity}
			eff := cascade.MergeSSMCascade(
				emptySSMKropath(), emptySSMKropath(),
				globalSSMCfg, localSSMCfg,
				emptySSMCfg(), emptySSMCfg(),
				emptySSMKropath(), emptySSMKropath(),
			)
			if tc.wantMandatory == nil {
				if eff.Mandatory.ApprovedPatchesEnableNonSecurity != nil {
					t.Errorf("expected nil, got %v", *eff.Mandatory.ApprovedPatchesEnableNonSecurity)
				}
			} else {
				if eff.Mandatory.ApprovedPatchesEnableNonSecurity == nil {
					t.Error("expected non-nil bool pointer, got nil")
				} else if *eff.Mandatory.ApprovedPatchesEnableNonSecurity != *tc.wantMandatory {
					t.Errorf("got %v, want %v", *eff.Mandatory.ApprovedPatchesEnableNonSecurity, *tc.wantMandatory)
				}
			}
		})
	}
}

// TestMergeSSMCascade_AllowedDocumentTypes verifies firstNonEmptyStrings across levels.
func TestMergeSSMCascade_AllowedDocumentTypes(t *testing.T) {
	globalKPC := cascade.SSMKropathSection{AllowedDocumentTypes: []string{"Command", "Session"}}
	localSSMCfg := cascade.SSMConfigSection{AllowedDocumentTypes: []string{"Package"}}

	eff := cascade.MergeSSMCascade(
		globalKPC, emptySSMKropath(),
		emptySSMCfg(), localSSMCfg,
		emptySSMCfg(), emptySSMCfg(),
		emptySSMKropath(), emptySSMKropath(),
	)
	// L1 wins (non-empty) over L4
	if len(eff.Mandatory.AllowedDocumentTypes) != 2 {
		t.Errorf("expected 2 allowedDocumentTypes from L1, got %v", eff.Mandatory.AllowedDocumentTypes)
	}
	if eff.Mandatory.AllowedDocumentTypes[0] != "Command" {
		t.Errorf("expected first element %q, got %q", "Command", eff.Mandatory.AllowedDocumentTypes[0])
	}
}

// TestMergeSSMCascade_TagsAdditiveUnion verifies that tags from all mandatory levels
// are merged with higher-priority levels overwriting lower-priority on key conflict.
func TestMergeSSMCascade_TagsAdditiveUnion(t *testing.T) {
	globalKPC := cascade.SSMKropathSection{Tags: map[string]string{"env": "prod", "owner": "platform"}}
	globalSSMCfg := cascade.SSMConfigSection{Tags: map[string]string{"env": "staging", "team": "infra"}}

	eff := cascade.MergeSSMCascade(
		globalKPC, emptySSMKropath(),
		globalSSMCfg, emptySSMCfg(),
		emptySSMCfg(), emptySSMCfg(),
		emptySSMKropath(), emptySSMKropath(),
	)
	// L1 (globalKPC) wins on "env" key conflict; all non-conflicting keys merged
	if eff.Mandatory.Tags["env"] != "prod" {
		t.Errorf("tags.env: got %q, want %q (L1 should win)", eff.Mandatory.Tags["env"], "prod")
	}
	if eff.Mandatory.Tags["owner"] != "platform" {
		t.Errorf("tags.owner: got %q, want %q", eff.Mandatory.Tags["owner"], "platform")
	}
	if eff.Mandatory.Tags["team"] != "infra" {
		t.Errorf("tags.team: got %q, want %q", eff.Mandatory.Tags["team"], "infra")
	}
}

// TestMergeSSMCascade_SyncedLabelsSSMConfigOnly verifies syncedLabels only comes
// from SSMConfig levels (not KropathConfig).
func TestMergeSSMCascade_SyncedLabelsSSMConfigOnly(t *testing.T) {
	globalSSMCfg := cascade.SSMConfigSection{SyncedLabels: map[string]string{"label-l3": "val3"}}
	localSSMCfg := cascade.SSMConfigSection{SyncedLabels: map[string]string{"label-l4": "val4", "label-l3": "val4-override"}}

	eff := cascade.MergeSSMCascade(
		emptySSMKropath(), emptySSMKropath(),
		globalSSMCfg, localSSMCfg,
		emptySSMCfg(), emptySSMCfg(),
		emptySSMKropath(), emptySSMKropath(),
	)
	// L3 wins on key conflict, but L4-only key is additive
	if eff.Mandatory.SyncedLabels["label-l3"] != "val3" {
		t.Errorf("syncedLabels[label-l3]: got %q, want %q (L3 wins)", eff.Mandatory.SyncedLabels["label-l3"], "val3")
	}
	if eff.Mandatory.SyncedLabels["label-l4"] != "val4" {
		t.Errorf("syncedLabels[label-l4]: got %q, want %q", eff.Mandatory.SyncedLabels["label-l4"], "val4")
	}
}

// TestMergeSSMCascade_DocumentTypeSSMConfigOnly verifies documentType only comes
// from SSMConfig levels (not KropathConfig levels).
func TestMergeSSMCascade_DocumentTypeSSMConfigOnly(t *testing.T) {
	// Even if we set ParameterType on KropathConfig, documentType stays zero
	// since KropathConfig has no documentType field.
	globalSSMCfg := cascade.SSMConfigSection{DocumentType: "Command"}
	eff := cascade.MergeSSMCascade(
		emptySSMKropath(), emptySSMKropath(),
		globalSSMCfg, emptySSMCfg(),
		emptySSMCfg(), emptySSMCfg(),
		emptySSMKropath(), emptySSMKropath(),
	)
	if eff.Mandatory.DocumentType != "Command" {
		t.Errorf("mandatory.documentType: got %q, want %q", eff.Mandatory.DocumentType, "Command")
	}
}

// TestMergeSSMCascade_ZeroValues verifies that all sources absent yields zero-value effective config.
func TestMergeSSMCascade_ZeroValues(t *testing.T) {
	eff := cascade.MergeSSMCascade(
		emptySSMKropath(), emptySSMKropath(),
		emptySSMCfg(), emptySSMCfg(),
		emptySSMCfg(), emptySSMCfg(),
		emptySSMKropath(), emptySSMKropath(),
	)
	if eff.Mandatory.Type != "" {
		t.Errorf("expected empty type, got %q", eff.Mandatory.Type)
	}
	if eff.Mandatory.ApprovedPatchesEnableNonSecurity != nil {
		t.Error("expected nil ApprovedPatchesEnableNonSecurity")
	}
	if len(eff.Mandatory.Tags) != 0 {
		t.Errorf("expected empty tags, got %v", eff.Mandatory.Tags)
	}
}

// TestValidateSSMDocumentType_Valid covers constraint-not-applicable and member-found cases.
func TestValidateSSMDocumentType_Valid(t *testing.T) {
	tests := []struct {
		name     string
		section  cascade.EffectiveSSMSection
		wantOK   bool
	}{
		{
			name:    "both_empty_no_constraint",
			section: cascade.EffectiveSSMSection{},
			wantOK:  true,
		},
		{
			name:    "documentType_empty_no_constraint",
			section: cascade.EffectiveSSMSection{AllowedDocumentTypes: []string{"Command"}},
			wantOK:  true,
		},
		{
			name:    "allowedDocumentTypes_empty_no_constraint",
			section: cascade.EffectiveSSMSection{DocumentType: "Command"},
			wantOK:  true,
		},
		{
			name: "documentType_in_allowedDocumentTypes",
			section: cascade.EffectiveSSMSection{
				DocumentType:         "Command",
				AllowedDocumentTypes: []string{"Session", "Command", "Package"},
			},
			wantOK: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ok, msg := cascade.ValidateSSMDocumentType(tc.section)
			if ok != tc.wantOK {
				t.Errorf("valid=%v msg=%q, want valid=%v", ok, msg, tc.wantOK)
			}
		})
	}
}

// TestValidateSSMDocumentType_Invalid verifies rejection when documentType not in allowedDocumentTypes.
func TestValidateSSMDocumentType_Invalid(t *testing.T) {
	section := cascade.EffectiveSSMSection{
		DocumentType:         "Automation",
		AllowedDocumentTypes: []string{"Command", "Session"},
	}
	ok, msg := cascade.ValidateSSMDocumentType(section)
	if ok {
		t.Error("expected valid=false, got valid=true")
	}
	if msg == "" {
		t.Error("expected non-empty message for invalid constraint")
	}
}
