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
	"reflect"
	"testing"

	"github.com/kropath/kropath-aws-controller/internal/cascade"
)

// Zero-value sentinels used across tests.
var (
	zeroOrgKropath = cascade.OrganizationsKropathSection{}
	zeroOrgCfg     = cascade.OrganizationsConfigSection{}
)

// isZeroEffOrgSection returns true if every field of s is empty / nil.
func isZeroEffOrgSection(s cascade.EffectiveOrganizationsSection) bool {
	return s.IamUserAccessToBilling == "" &&
		s.RoleName == "" &&
		s.ParentID == "" &&
		s.NamingTemplate == "" &&
		len(s.SyncedLabels) == 0 &&
		len(s.SyncedAnnotations) == 0 &&
		len(s.Tags) == 0
}

// mergeOrgsAll is a thin wrapper that passes all eight cascade inputs to
// MergeOrganizationsCascade in the canonical order, so individual tests only
// need to override the levels they care about.
func mergeOrgsAll(
	glKpcMand, lcKpcMand cascade.OrganizationsKropathSection,
	glOrgMand, lcOrgMand cascade.OrganizationsConfigSection,
	lcOrgDef, glOrgDef cascade.OrganizationsConfigSection,
	lcKpcDef, glKpcDef cascade.OrganizationsKropathSection,
) cascade.EffectiveOrganizationsConfig {
	return cascade.MergeOrganizationsCascade(
		glKpcMand, lcKpcMand,
		glOrgMand, lcOrgMand,
		lcOrgDef, glOrgDef,
		lcKpcDef, glKpcDef,
	)
}

// TestMergeOrganizationsCascade_AC9_KropathConfigMandatoryIamUserAccessToBilling
// verifies AC-9: KropathConfig.mandatory.organizations.iamUserAccessToBilling="deny"
// propagates to effectiveConfig.mandatory.iamUserAccessToBilling.
func TestMergeOrganizationsCascade_AC9_KropathConfigMandatoryIamUserAccessToBilling(t *testing.T) {
	got := mergeOrgsAll(
		cascade.OrganizationsKropathSection{IamUserAccessToBilling: "deny"}, // L1 global KPC
		zeroOrgKropath, zeroOrgCfg, zeroOrgCfg,
		zeroOrgCfg, zeroOrgCfg,
		zeroOrgKropath, zeroOrgKropath,
	)

	if got.Mandatory.IamUserAccessToBilling != "deny" {
		t.Errorf("Mandatory.IamUserAccessToBilling = %q, want %q",
			got.Mandatory.IamUserAccessToBilling, "deny")
	}
	if !isZeroEffOrgSection(got.Defaults) {
		t.Errorf("Defaults should be zero when only mandatory KPC L1 is set, got %+v", got.Defaults)
	}
}

// TestMergeOrganizationsCascade_AC9_Level1OverridesLevel3 verifies that
// globalKropathMandatory (L1) beats globalOrgCfgMandatory (L3) for scalar fields.
func TestMergeOrganizationsCascade_AC9_Level1OverridesLevel3(t *testing.T) {
	got := mergeOrgsAll(
		cascade.OrganizationsKropathSection{IamUserAccessToBilling: "deny"}, // L1 wins
		zeroOrgKropath,
		cascade.OrganizationsConfigSection{IamUserAccessToBilling: "allow"}, // L3
		zeroOrgCfg,
		zeroOrgCfg, zeroOrgCfg,
		zeroOrgKropath, zeroOrgKropath,
	)

	if got.Mandatory.IamUserAccessToBilling != "deny" {
		t.Errorf("Mandatory.IamUserAccessToBilling = %q, want %q (L1 must win over L3)",
			got.Mandatory.IamUserAccessToBilling, "deny")
	}
}

// TestMergeOrganizationsCascade_Level2OverridesLevel3 verifies localKropathMandatory
// (L2) beats globalOrgCfgMandatory (L3).
func TestMergeOrganizationsCascade_Level2OverridesLevel3(t *testing.T) {
	got := mergeOrgsAll(
		zeroOrgKropath,
		cascade.OrganizationsKropathSection{RoleName: "AWSControlTowerExecution"}, // L2
		cascade.OrganizationsConfigSection{RoleName: "OtherRole"},                 // L3
		zeroOrgCfg,
		zeroOrgCfg, zeroOrgCfg,
		zeroOrgKropath, zeroOrgKropath,
	)

	if got.Mandatory.RoleName != "AWSControlTowerExecution" {
		t.Errorf("Mandatory.RoleName = %q, want %q (L2 must win over L3)",
			got.Mandatory.RoleName, "AWSControlTowerExecution")
	}
}

// TestMergeOrganizationsCascade_Level3OverridesLevel4 verifies globalOrgCfgMandatory
// (L3) beats localOrgCfgMandatory (L4).
func TestMergeOrganizationsCascade_Level3OverridesLevel4(t *testing.T) {
	got := mergeOrgsAll(
		zeroOrgKropath, zeroOrgKropath,
		cascade.OrganizationsConfigSection{ParentID: "r-glob"}, // L3
		cascade.OrganizationsConfigSection{ParentID: "r-local"}, // L4
		zeroOrgCfg, zeroOrgCfg,
		zeroOrgKropath, zeroOrgKropath,
	)

	if got.Mandatory.ParentID != "r-glob" {
		t.Errorf("Mandatory.ParentID = %q, want %q (L3 must win over L4)",
			got.Mandatory.ParentID, "r-glob")
	}
}

// TestMergeOrganizationsCascade_Level4FallsThrough verifies localOrgCfgMandatory (L4)
// propagates when L1–L3 are all empty.
func TestMergeOrganizationsCascade_Level4FallsThrough(t *testing.T) {
	got := mergeOrgsAll(
		zeroOrgKropath, zeroOrgKropath,
		zeroOrgCfg,
		cascade.OrganizationsConfigSection{
			IamUserAccessToBilling: "allow",
			RoleName:               "OrganizationAccountAccessRole",
			ParentID:               "ou-abc-1234567",
		}, // L4
		zeroOrgCfg, zeroOrgCfg,
		zeroOrgKropath, zeroOrgKropath,
	)

	if got.Mandatory.IamUserAccessToBilling != "allow" {
		t.Errorf("Mandatory.IamUserAccessToBilling = %q, want %q", got.Mandatory.IamUserAccessToBilling, "allow")
	}
	if got.Mandatory.RoleName != "OrganizationAccountAccessRole" {
		t.Errorf("Mandatory.RoleName = %q, want %q", got.Mandatory.RoleName, "OrganizationAccountAccessRole")
	}
	if got.Mandatory.ParentID != "ou-abc-1234567" {
		t.Errorf("Mandatory.ParentID = %q, want %q", got.Mandatory.ParentID, "ou-abc-1234567")
	}
}

// TestMergeOrganizationsCascade_DefaultsLevel6WinsOverLevel7 verifies L6
// (localOrgCfgDefaults) beats L7 (globalOrgCfgDefaults) for defaults.
func TestMergeOrganizationsCascade_DefaultsLevel6WinsOverLevel7(t *testing.T) {
	got := mergeOrgsAll(
		zeroOrgKropath, zeroOrgKropath,
		zeroOrgCfg, zeroOrgCfg,
		cascade.OrganizationsConfigSection{RoleName: "LocalDefault"},  // L6
		cascade.OrganizationsConfigSection{RoleName: "GlobalDefault"}, // L7
		zeroOrgKropath, zeroOrgKropath,
	)

	if got.Defaults.RoleName != "LocalDefault" {
		t.Errorf("Defaults.RoleName = %q, want %q (L6 must win over L7)",
			got.Defaults.RoleName, "LocalDefault")
	}
}

// TestMergeOrganizationsCascade_DefaultsLevel7OverLevel8 verifies L7 beats L8.
func TestMergeOrganizationsCascade_DefaultsLevel7OverLevel8(t *testing.T) {
	got := mergeOrgsAll(
		zeroOrgKropath, zeroOrgKropath,
		zeroOrgCfg, zeroOrgCfg,
		zeroOrgCfg,
		cascade.OrganizationsConfigSection{IamUserAccessToBilling: "allow"}, // L7
		cascade.OrganizationsKropathSection{IamUserAccessToBilling: "deny"}, // L8
		zeroOrgKropath,
	)

	if got.Defaults.IamUserAccessToBilling != "allow" {
		t.Errorf("Defaults.IamUserAccessToBilling = %q, want %q (L7 must win over L8)",
			got.Defaults.IamUserAccessToBilling, "allow")
	}
}

// TestMergeOrganizationsCascade_DefaultsLevel9FallsThrough verifies globalKropathDefaults
// (L9) propagates when L6–L8 are all empty.
func TestMergeOrganizationsCascade_DefaultsLevel9FallsThrough(t *testing.T) {
	got := mergeOrgsAll(
		zeroOrgKropath, zeroOrgKropath,
		zeroOrgCfg, zeroOrgCfg,
		zeroOrgCfg, zeroOrgCfg,
		zeroOrgKropath,
		cascade.OrganizationsKropathSection{ParentID: "r-globaldefault"}, // L9
	)

	if got.Defaults.ParentID != "r-globaldefault" {
		t.Errorf("Defaults.ParentID = %q, want %q (L9 must propagate)",
			got.Defaults.ParentID, "r-globaldefault")
	}
}

// TestMergeOrganizationsCascade_MandatoryDoesNotLeakToDefaults verifies that a
// mandatory setting at L1 does not appear in the Defaults tier.
func TestMergeOrganizationsCascade_MandatoryDoesNotLeakToDefaults(t *testing.T) {
	got := mergeOrgsAll(
		cascade.OrganizationsKropathSection{
			IamUserAccessToBilling: "deny",
			RoleName:               "CtExecution",
			ParentID:               "r-corp",
		}, // L1
		zeroOrgKropath,
		zeroOrgCfg, zeroOrgCfg,
		zeroOrgCfg, zeroOrgCfg,
		zeroOrgKropath, zeroOrgKropath,
	)

	if !isZeroEffOrgSection(got.Defaults) {
		t.Errorf("Defaults must be zero when only Mandatory L1 is set, got %+v", got.Defaults)
	}
}

// TestMergeOrganizationsCascade_AC10_TagUnionMerge verifies AC-10: tags from
// KropathConfig.mandatory.tags (global) and OrganizationsConfig.mandatory.tags
// (local) are union-merged into effectiveConfig.mandatory.tags.
func TestMergeOrganizationsCascade_AC10_TagUnionMerge(t *testing.T) {
	got := mergeOrgsAll(
		cascade.OrganizationsKropathSection{Tags: map[string]string{"cost-centre": "infra"}}, // L1
		zeroOrgKropath,
		zeroOrgCfg,
		cascade.OrganizationsConfigSection{Tags: map[string]string{"team": "platform"}}, // L4
		zeroOrgCfg, zeroOrgCfg,
		zeroOrgKropath, zeroOrgKropath,
	)

	want := map[string]string{
		"cost-centre": "infra",
		"team":        "platform",
	}
	if !reflect.DeepEqual(got.Mandatory.Tags, want) {
		t.Errorf("Mandatory.Tags = %v, want %v", got.Mandatory.Tags, want)
	}
}

// TestMergeOrganizationsCascade_AC10_TagConflictLevel1Wins verifies that when the
// same key appears at L1 (global KropathConfig) and L4 (localOrgCfg), L1 wins.
func TestMergeOrganizationsCascade_AC10_TagConflictLevel1Wins(t *testing.T) {
	got := mergeOrgsAll(
		cascade.OrganizationsKropathSection{Tags: map[string]string{"env": "prod"}}, // L1
		zeroOrgKropath,
		zeroOrgCfg,
		cascade.OrganizationsConfigSection{Tags: map[string]string{"env": "dev"}}, // L4
		zeroOrgCfg, zeroOrgCfg,
		zeroOrgKropath, zeroOrgKropath,
	)

	if got.Mandatory.Tags["env"] != "prod" {
		t.Errorf("Mandatory.Tags[env] = %q, want %q (L1 must win on key conflict)",
			got.Mandatory.Tags["env"], "prod")
	}
}

// TestMergeOrganizationsCascade_AC11_SyncedLabels verifies AC-11:
// OrganizationsConfig.mandatory.syncedLabels propagates to
// effectiveConfig.mandatory.syncedLabels.
func TestMergeOrganizationsCascade_AC11_SyncedLabels(t *testing.T) {
	got := mergeOrgsAll(
		zeroOrgKropath, zeroOrgKropath,
		zeroOrgCfg,
		cascade.OrganizationsConfigSection{
			SyncedLabels: map[string]string{"data-class": "restricted"},
		}, // L4
		zeroOrgCfg, zeroOrgCfg,
		zeroOrgKropath, zeroOrgKropath,
	)

	want := map[string]string{"data-class": "restricted"}
	if !reflect.DeepEqual(got.Mandatory.SyncedLabels, want) {
		t.Errorf("Mandatory.SyncedLabels = %v, want %v", got.Mandatory.SyncedLabels, want)
	}
}

// TestMergeOrganizationsCascade_AC11_SyncedLabelsAdditiveUnion verifies that
// SyncedLabels from L3 and L4 are union-merged (L3 wins on key conflict).
func TestMergeOrganizationsCascade_AC11_SyncedLabelsAdditiveUnion(t *testing.T) {
	got := mergeOrgsAll(
		zeroOrgKropath, zeroOrgKropath,
		cascade.OrganizationsConfigSection{
			SyncedLabels: map[string]string{"governance": "strict", "env": "prod"},
		}, // L3 global OrgCfg mandatory
		cascade.OrganizationsConfigSection{
			SyncedLabels: map[string]string{"team": "platform", "env": "dev"},
		}, // L4 local OrgCfg mandatory
		zeroOrgCfg, zeroOrgCfg,
		zeroOrgKropath, zeroOrgKropath,
	)

	// L3 wins on "env" conflict; "governance" from L3 and "team" from L4 both appear.
	want := map[string]string{
		"governance": "strict",
		"env":        "prod",
		"team":       "platform",
	}
	if !reflect.DeepEqual(got.Mandatory.SyncedLabels, want) {
		t.Errorf("Mandatory.SyncedLabels = %v, want %v", got.Mandatory.SyncedLabels, want)
	}
}

// TestMergeOrganizationsCascade_SyncedAnnotationsAdditiveUnion verifies that
// SyncedAnnotations from L3 and L4 are union-merged.
func TestMergeOrganizationsCascade_SyncedAnnotationsAdditiveUnion(t *testing.T) {
	got := mergeOrgsAll(
		zeroOrgKropath, zeroOrgKropath,
		cascade.OrganizationsConfigSection{
			SyncedAnnotations: map[string]string{"iam.kropath.run/boundary": "corp"},
		}, // L3
		cascade.OrganizationsConfigSection{
			SyncedAnnotations: map[string]string{"team": "platform"},
		}, // L4
		zeroOrgCfg, zeroOrgCfg,
		zeroOrgKropath, zeroOrgKropath,
	)

	want := map[string]string{
		"iam.kropath.run/boundary": "corp",
		"team":                     "platform",
	}
	if !reflect.DeepEqual(got.Mandatory.SyncedAnnotations, want) {
		t.Errorf("Mandatory.SyncedAnnotations = %v, want %v", got.Mandatory.SyncedAnnotations, want)
	}
}

// TestMergeOrganizationsCascade_NamingTemplateOrgCfgOnly verifies that
// NamingTemplate is governed at OrganizationsConfig levels only (L3/L4 mandatory,
// L6/L7 defaults) and L3 beats L4.
func TestMergeOrganizationsCascade_NamingTemplateOrgCfgOnly(t *testing.T) {
	got := mergeOrgsAll(
		zeroOrgKropath, zeroOrgKropath,
		cascade.OrganizationsConfigSection{NamingTemplate: "{namespace}-{name}-global"}, // L3
		cascade.OrganizationsConfigSection{NamingTemplate: "{namespace}-{name}-local"},  // L4
		zeroOrgCfg, zeroOrgCfg,
		zeroOrgKropath, zeroOrgKropath,
	)

	if got.Mandatory.NamingTemplate != "{namespace}-{name}-global" {
		t.Errorf("Mandatory.NamingTemplate = %q, want %q (L3 must win over L4)",
			got.Mandatory.NamingTemplate, "{namespace}-{name}-global")
	}
}

// TestMergeOrganizationsCascade_AllZero verifies that all-zero input produces
// all-zero output.
func TestMergeOrganizationsCascade_AllZero(t *testing.T) {
	got := mergeOrgsAll(
		zeroOrgKropath, zeroOrgKropath,
		zeroOrgCfg, zeroOrgCfg,
		zeroOrgCfg, zeroOrgCfg,
		zeroOrgKropath, zeroOrgKropath,
	)

	if !isZeroEffOrgSection(got.Mandatory) {
		t.Errorf("Mandatory should be zero when all inputs are zero, got %+v", got.Mandatory)
	}
	if !isZeroEffOrgSection(got.Defaults) {
		t.Errorf("Defaults should be zero when all inputs are zero, got %+v", got.Defaults)
	}
}

// TestMergeOrganizationsCascade_DefaultsTagsUnionMerge verifies that tags from
// all four defaults sources (L6, L7, L8, L9) are union-merged with L6 winning
// on key conflict.
func TestMergeOrganizationsCascade_DefaultsTagsUnionMerge(t *testing.T) {
	got := mergeOrgsAll(
		zeroOrgKropath, zeroOrgKropath,
		zeroOrgCfg, zeroOrgCfg,
		cascade.OrganizationsConfigSection{Tags: map[string]string{"env": "prod", "from": "L6"}}, // L6
		cascade.OrganizationsConfigSection{Tags: map[string]string{"env": "staging", "from": "L7"}}, // L7
		cascade.OrganizationsKropathSection{Tags: map[string]string{"cost-centre": "L8"}}, // L8
		cascade.OrganizationsKropathSection{Tags: map[string]string{"cost-centre": "L9"}}, // L9
	)

	want := map[string]string{
		"env":         "prod",     // L6 wins over L7 on key conflict
		"from":        "L6",       // L6 wins over L7
		"cost-centre": "L8",       // L8 wins over L9
	}
	if !reflect.DeepEqual(got.Defaults.Tags, want) {
		t.Errorf("Defaults.Tags = %v, want %v", got.Defaults.Tags, want)
	}
}

// TestMergeOrganizationsCascade_DefaultsSyncedLabelsL6OverL7 verifies that
// defaults SyncedLabels from L6 win over L7.
func TestMergeOrganizationsCascade_DefaultsSyncedLabelsL6OverL7(t *testing.T) {
	got := mergeOrgsAll(
		zeroOrgKropath, zeroOrgKropath,
		zeroOrgCfg, zeroOrgCfg,
		cascade.OrganizationsConfigSection{
			SyncedLabels: map[string]string{"tier": "local", "extra": "local-only"},
		}, // L6
		cascade.OrganizationsConfigSection{
			SyncedLabels: map[string]string{"tier": "global", "global-only": "yes"},
		}, // L7
		zeroOrgKropath, zeroOrgKropath,
	)

	want := map[string]string{
		"tier":        "local",
		"extra":       "local-only",
		"global-only": "yes",
	}
	if !reflect.DeepEqual(got.Defaults.SyncedLabels, want) {
		t.Errorf("Defaults.SyncedLabels = %v, want %v", got.Defaults.SyncedLabels, want)
	}
}
