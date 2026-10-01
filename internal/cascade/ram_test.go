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
	zeroRAMKropath = cascade.RAMKropathSection{}
	zeroRAMCfg     = cascade.RAMConfigSection{}
)

// mergeRAMAll calls MergeRAMCascade with all eight inputs.
func mergeRAMAll(
	globalKropathMandatory,
	localKropathMandatory cascade.RAMKropathSection,
	globalRAMCfgMandatory,
	localRAMCfgMandatory,
	localRAMCfgDefaults,
	globalRAMCfgDefaults cascade.RAMConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.RAMKropathSection,
) cascade.EffectiveRAMConfig {
	return cascade.MergeRAMCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalRAMCfgMandatory,
		localRAMCfgMandatory,
		localRAMCfgDefaults,
		globalRAMCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeRAMCascade_MandatoryAllowExternalPrincipalsKropathWins verifies that
// KropathConfig.mandatory.ram.allowExternalPrincipals (L1) beats RAMConfig.mandatory (L4).
func TestMergeRAMCascade_MandatoryAllowExternalPrincipalsKropathWins(t *testing.T) {
	// L1 sets false (block), L4 sets true (allow) — L1 must win.
	got := mergeRAMAll(
		cascade.RAMKropathSection{AllowExternalPrincipals: boolPtr(false)}, // level 1 — wins
		zeroRAMKropath, // level 2
		zeroRAMCfg,     // level 3
		cascade.RAMConfigSection{AllowExternalPrincipals: boolPtr(true)}, // level 4 — overridden
		zeroRAMCfg,     // level 6
		zeroRAMCfg,     // level 7
		zeroRAMKropath, // level 8
		zeroRAMKropath, // level 9
	)

	if got.Mandatory.AllowExternalPrincipals == nil || *got.Mandatory.AllowExternalPrincipals != false {
		t.Errorf("mandatory.allowExternalPrincipals = %v, want false (L1 wins)", got.Mandatory.AllowExternalPrincipals)
	}
	if got.Defaults.AllowExternalPrincipals != nil {
		t.Errorf("mandatory allowExternalPrincipals must not bleed into defaults; got defaults.allowExternalPrincipals = %v", got.Defaults.AllowExternalPrincipals)
	}
}

// TestMergeRAMCascade_MandatoryAllowExternalPrincipalsFallthrough verifies that
// a nil L1 falls through to L4 for allowExternalPrincipals.
func TestMergeRAMCascade_MandatoryAllowExternalPrincipalsFallthrough(t *testing.T) {
	got := mergeRAMAll(
		zeroRAMKropath, // level 1 — nil, falls through
		zeroRAMKropath, // level 2
		zeroRAMCfg,     // level 3
		cascade.RAMConfigSection{AllowExternalPrincipals: boolPtr(false)}, // level 4 — sets value
		zeroRAMCfg,     // level 6
		zeroRAMCfg,     // level 7
		zeroRAMKropath, // level 8
		zeroRAMKropath, // level 9
	)

	if got.Mandatory.AllowExternalPrincipals == nil || *got.Mandatory.AllowExternalPrincipals != false {
		t.Errorf("mandatory.allowExternalPrincipals = %v, want false (L4 fallthrough)", got.Mandatory.AllowExternalPrincipals)
	}
}

// TestMergeRAMCascade_MandatoryAllowedResourceTypesKropathWins verifies that
// KropathConfig mandatory allowedResourceTypes (L1) beats RAMConfig mandatory (L4).
func TestMergeRAMCascade_MandatoryAllowedResourceTypesKropathWins(t *testing.T) {
	got := mergeRAMAll(
		cascade.RAMKropathSection{AllowedResourceTypes: []string{"ec2:Subnet"}}, // level 1 — wins
		zeroRAMKropath, // level 2
		zeroRAMCfg,     // level 3
		cascade.RAMConfigSection{AllowedResourceTypes: []string{"ec2:TransitGateway"}}, // level 4 — overridden
		zeroRAMCfg,     // level 6
		zeroRAMCfg,     // level 7
		zeroRAMKropath, // level 8
		zeroRAMKropath, // level 9
	)

	if len(got.Mandatory.AllowedResourceTypes) != 1 || got.Mandatory.AllowedResourceTypes[0] != "ec2:Subnet" {
		t.Errorf("mandatory.allowedResourceTypes = %v, want [ec2:Subnet] (L1 wins)", got.Mandatory.AllowedResourceTypes)
	}
}

// TestMergeRAMCascade_MandatoryAllowedResourceTypesFallthrough verifies that
// empty L1 falls through to L4 for allowedResourceTypes.
func TestMergeRAMCascade_MandatoryAllowedResourceTypesFallthrough(t *testing.T) {
	got := mergeRAMAll(
		zeroRAMKropath, // level 1 — empty, falls through
		zeroRAMKropath, // level 2
		zeroRAMCfg,     // level 3
		cascade.RAMConfigSection{AllowedResourceTypes: []string{"ec2:Subnet", "ec2:TransitGateway"}}, // level 4
		zeroRAMCfg,     // level 6
		zeroRAMCfg,     // level 7
		zeroRAMKropath, // level 8
		zeroRAMKropath, // level 9
	)

	if len(got.Mandatory.AllowedResourceTypes) != 2 {
		t.Errorf("mandatory.allowedResourceTypes = %v, want [ec2:Subnet ec2:TransitGateway] (L4 fallthrough)", got.Mandatory.AllowedResourceTypes)
	}
}

// TestMergeRAMCascade_MandatoryNamingTemplateRAMConfigOnly verifies that namingTemplate
// in the mandatory tier comes only from RAMConfig levels (L3, L4), not KropathConfig.
func TestMergeRAMCascade_MandatoryNamingTemplateRAMConfigOnly(t *testing.T) {
	got := mergeRAMAll(
		zeroRAMKropath, // level 1
		zeroRAMKropath, // level 2
		cascade.RAMConfigSection{NamingTemplate: "corp-{name}"}, // level 3
		zeroRAMCfg,     // level 4
		zeroRAMCfg,     // level 6
		zeroRAMCfg,     // level 7
		zeroRAMKropath, // level 8
		zeroRAMKropath, // level 9
	)

	if got.Mandatory.NamingTemplate != "corp-{name}" {
		t.Errorf("mandatory.namingTemplate = %q, want %q (L3 provides mandatory namingTemplate)", got.Mandatory.NamingTemplate, "corp-{name}")
	}
	if got.Defaults.NamingTemplate != "" {
		t.Errorf("mandatory namingTemplate must not bleed into defaults; got defaults.namingTemplate = %q", got.Defaults.NamingTemplate)
	}
}

// TestMergeRAMCascade_DefaultsNamingTemplate verifies that
// localRAMCfgDefaults.NamingTemplate at level 6 propagates to effCfg.defaults.namingTemplate.
func TestMergeRAMCascade_DefaultsNamingTemplate(t *testing.T) {
	tmpl := "{namespace}-{name}"
	got := mergeRAMAll(
		zeroRAMKropath, // level 1
		zeroRAMKropath, // level 2
		zeroRAMCfg,     // level 3
		zeroRAMCfg,     // level 4
		cascade.RAMConfigSection{NamingTemplate: tmpl}, // level 6
		zeroRAMCfg,     // level 7
		zeroRAMKropath, // level 8
		zeroRAMKropath, // level 9
	)

	if got.Defaults.NamingTemplate != tmpl {
		t.Errorf("defaults.namingTemplate = %q, want %q (level 6)", got.Defaults.NamingTemplate, tmpl)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("defaults namingTemplate must not bleed into mandatory; got mandatory.namingTemplate = %q", got.Mandatory.NamingTemplate)
	}
}

// TestMergeRAMCascade_MandatoryTagsFromKropathConfig verifies that
// globalKropathConfig.mandatory.tags win over RAMConfig mandatory tags (L1 beats L4).
func TestMergeRAMCascade_MandatoryTagsFromKropathConfig(t *testing.T) {
	got := mergeRAMAll(
		cascade.RAMKropathSection{Tags: map[string]string{"cost-centre": "infra"}}, // level 1
		zeroRAMKropath, // level 2
		zeroRAMCfg,     // level 3
		cascade.RAMConfigSection{Tags: map[string]string{"cost-centre": "team"}}, // level 4 — overridden by L1
		zeroRAMCfg,     // level 6
		zeroRAMCfg,     // level 7
		zeroRAMKropath, // level 8
		zeroRAMKropath, // level 9
	)

	if got.Mandatory.Tags["cost-centre"] != "infra" {
		t.Errorf("mandatory.tags[cost-centre] = %q, want %q (L1 wins)", got.Mandatory.Tags["cost-centre"], "infra")
	}
	if len(got.Defaults.Tags) != 0 {
		t.Errorf("mandatory tags must not bleed into defaults; got defaults.tags = %v", got.Defaults.Tags)
	}
}

// TestMergeRAMCascade_MandatoryTagsMerge verifies additive union merge for tags.
func TestMergeRAMCascade_MandatoryTagsMerge(t *testing.T) {
	got := mergeRAMAll(
		cascade.RAMKropathSection{Tags: map[string]string{"cost-centre": "infra"}}, // level 1
		zeroRAMKropath, // level 2
		zeroRAMCfg,     // level 3
		cascade.RAMConfigSection{Tags: map[string]string{"family": "ram", "cost-centre": "team"}}, // level 4
		cascade.RAMConfigSection{Tags: map[string]string{"team": "platform"}}, // level 6
		zeroRAMCfg,     // level 7
		zeroRAMKropath, // level 8
		zeroRAMKropath, // level 9
	)

	if got.Mandatory.Tags["cost-centre"] != "infra" {
		t.Errorf("mandatory.tags[cost-centre] = %q, want %q (L1 wins on conflict)", got.Mandatory.Tags["cost-centre"], "infra")
	}
	if got.Mandatory.Tags["family"] != "ram" {
		t.Errorf("mandatory.tags[family] = %q, want %q (L4 unique key present)", got.Mandatory.Tags["family"], "ram")
	}
	if got.Mandatory.Tags["team"] != "" {
		t.Errorf("defaults tag must not bleed into mandatory; mandatory.tags[team] = %q", got.Mandatory.Tags["team"])
	}
	if got.Defaults.Tags["team"] != "platform" {
		t.Errorf("defaults.tags[team] = %q, want %q", got.Defaults.Tags["team"], "platform")
	}
}

// TestMergeRAMCascade_DefaultsAllowExternalPrincipals verifies L6 wins over L9 for defaults.
func TestMergeRAMCascade_DefaultsAllowExternalPrincipals(t *testing.T) {
	got := mergeRAMAll(
		zeroRAMKropath, // level 1
		zeroRAMKropath, // level 2
		zeroRAMCfg,     // level 3
		zeroRAMCfg,     // level 4
		cascade.RAMConfigSection{AllowExternalPrincipals: boolPtr(true)}, // level 6 — wins
		cascade.RAMConfigSection{AllowExternalPrincipals: boolPtr(false)}, // level 7 — overridden
		cascade.RAMKropathSection{AllowExternalPrincipals: boolPtr(false)}, // level 8
		cascade.RAMKropathSection{AllowExternalPrincipals: boolPtr(false)}, // level 9
	)

	if got.Defaults.AllowExternalPrincipals == nil || *got.Defaults.AllowExternalPrincipals != true {
		t.Errorf("defaults.allowExternalPrincipals = %v, want true (L6 wins)", got.Defaults.AllowExternalPrincipals)
	}
}

// TestMergeRAMCascade_DefaultsAllowedResourceTypesPriority verifies L6 wins over L7-L9.
func TestMergeRAMCascade_DefaultsAllowedResourceTypesPriority(t *testing.T) {
	got := mergeRAMAll(
		zeroRAMKropath, // level 1
		zeroRAMKropath, // level 2
		zeroRAMCfg,     // level 3
		zeroRAMCfg,     // level 4
		cascade.RAMConfigSection{AllowedResourceTypes: []string{"ec2:Subnet"}}, // level 6 — wins
		cascade.RAMConfigSection{AllowedResourceTypes: []string{"ec2:TransitGateway"}}, // level 7
		cascade.RAMKropathSection{AllowedResourceTypes: []string{"ec2:VPC"}}, // level 8
		cascade.RAMKropathSection{AllowedResourceTypes: []string{"ec2:InternetGateway"}}, // level 9
	)

	if len(got.Defaults.AllowedResourceTypes) != 1 || got.Defaults.AllowedResourceTypes[0] != "ec2:Subnet" {
		t.Errorf("defaults.allowedResourceTypes = %v, want [ec2:Subnet] (L6 wins)", got.Defaults.AllowedResourceTypes)
	}
}

// TestMergeRAMCascade_MandatorySyncedLabels verifies additive union with L1 winning on conflict.
func TestMergeRAMCascade_MandatorySyncedLabels(t *testing.T) {
	got := mergeRAMAll(
		cascade.RAMKropathSection{SyncedLabels: map[string]string{"org": "kropath"}}, // level 1
		zeroRAMKropath, // level 2
		zeroRAMCfg,     // level 3
		cascade.RAMConfigSection{SyncedLabels: map[string]string{"data-class": "internal", "org": "ram-team"}}, // level 4
		zeroRAMCfg,     // level 6
		zeroRAMCfg,     // level 7
		zeroRAMKropath, // level 8
		zeroRAMKropath, // level 9
	)

	if got.Mandatory.SyncedLabels["org"] != "kropath" {
		t.Errorf("mandatory.syncedLabels[org] = %q, want %q (L1 wins)", got.Mandatory.SyncedLabels["org"], "kropath")
	}
	if got.Mandatory.SyncedLabels["data-class"] != "internal" {
		t.Errorf("mandatory.syncedLabels[data-class] = %q, want %q", got.Mandatory.SyncedLabels["data-class"], "internal")
	}
	if len(got.Defaults.SyncedLabels) != 0 {
		t.Errorf("mandatory syncedLabels must not bleed into defaults; got defaults.syncedLabels = %v", got.Defaults.SyncedLabels)
	}
}

// TestMergeRAMCascade_MandatorySyncedAnnotations verifies additive union with L1 winning.
func TestMergeRAMCascade_MandatorySyncedAnnotations(t *testing.T) {
	got := mergeRAMAll(
		cascade.RAMKropathSection{SyncedAnnotations: map[string]string{"compliance": "hipaa"}}, // level 1
		zeroRAMKropath, // level 2
		zeroRAMCfg,     // level 3
		cascade.RAMConfigSection{SyncedAnnotations: map[string]string{"owner": "ram-team", "compliance": "ram"}}, // level 4
		zeroRAMCfg,     // level 6
		zeroRAMCfg,     // level 7
		zeroRAMKropath, // level 8
		zeroRAMKropath, // level 9
	)

	if got.Mandatory.SyncedAnnotations["compliance"] != "hipaa" {
		t.Errorf("mandatory.syncedAnnotations[compliance] = %q, want %q (L1 wins)", got.Mandatory.SyncedAnnotations["compliance"], "hipaa")
	}
	if got.Mandatory.SyncedAnnotations["owner"] != "ram-team" {
		t.Errorf("mandatory.syncedAnnotations[owner] = %q, want %q", got.Mandatory.SyncedAnnotations["owner"], "ram-team")
	}
}

// TestMergeRAMCascade_DefaultsPriorityOrder verifies L6 wins over L7, L8, L9 for tags.
func TestMergeRAMCascade_DefaultsPriorityOrder(t *testing.T) {
	got := mergeRAMAll(
		zeroRAMKropath, // level 1
		zeroRAMKropath, // level 2
		zeroRAMCfg,     // level 3
		zeroRAMCfg,     // level 4
		cascade.RAMConfigSection{Tags: map[string]string{"owner": "local-ram"}}, // level 6 (wins)
		cascade.RAMConfigSection{Tags: map[string]string{"owner": "global-ram"}}, // level 7
		cascade.RAMKropathSection{Tags: map[string]string{"owner": "local-kpc"}}, // level 8
		cascade.RAMKropathSection{Tags: map[string]string{"owner": "global-kpc"}}, // level 9
	)

	if got.Defaults.Tags["owner"] != "local-ram" {
		t.Errorf("defaults priority: defaults.tags[owner] = %q, want %q (L6 wins)", got.Defaults.Tags["owner"], "local-ram")
	}
}

// TestMergeRAMCascade_AllowedResourceTypesNotUnionMerged verifies that allowedResourceTypes
// uses first-non-empty (not array union) semantics.
func TestMergeRAMCascade_AllowedResourceTypesNotUnionMerged(t *testing.T) {
	got := mergeRAMAll(
		cascade.RAMKropathSection{AllowedResourceTypes: []string{"ec2:Subnet"}}, // level 1
		zeroRAMKropath, // level 2
		zeroRAMCfg,     // level 3
		cascade.RAMConfigSection{AllowedResourceTypes: []string{"ec2:TransitGateway"}}, // level 4
		zeroRAMCfg,     // level 6
		zeroRAMCfg,     // level 7
		zeroRAMKropath, // level 8
		zeroRAMKropath, // level 9
	)

	// Must be L1's list only; must NOT union-merge with L4's list.
	if len(got.Mandatory.AllowedResourceTypes) != 1 || got.Mandatory.AllowedResourceTypes[0] != "ec2:Subnet" {
		t.Errorf("allowedResourceTypes must be first-non-empty (not union); got = %v, want [ec2:Subnet]", got.Mandatory.AllowedResourceTypes)
	}
}

// TestMergeRAMCascade_EmptySourcesYieldEmptyResult verifies that all-zero inputs
// produce an empty EffectiveRAMConfig with no panics.
func TestMergeRAMCascade_EmptySourcesYieldEmptyResult(t *testing.T) {
	got := mergeRAMAll(
		zeroRAMKropath, zeroRAMKropath,
		zeroRAMCfg, zeroRAMCfg, zeroRAMCfg, zeroRAMCfg,
		zeroRAMKropath, zeroRAMKropath,
	)

	if got.Mandatory.AllowExternalPrincipals != nil {
		t.Errorf("empty inputs: mandatory.allowExternalPrincipals = %v, want nil", got.Mandatory.AllowExternalPrincipals)
	}
	if len(got.Mandatory.AllowedResourceTypes) != 0 {
		t.Errorf("empty inputs: mandatory.allowedResourceTypes = %v, want nil/empty", got.Mandatory.AllowedResourceTypes)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("empty inputs: mandatory.namingTemplate = %q, want empty", got.Mandatory.NamingTemplate)
	}
	if len(got.Mandatory.Tags) != 0 {
		t.Errorf("empty inputs: mandatory.tags = %v, want nil/empty", got.Mandatory.Tags)
	}
	if got.Defaults.AllowExternalPrincipals != nil {
		t.Errorf("empty inputs: defaults.allowExternalPrincipals = %v, want nil", got.Defaults.AllowExternalPrincipals)
	}
	if len(got.Defaults.AllowedResourceTypes) != 0 {
		t.Errorf("empty inputs: defaults.allowedResourceTypes = %v, want nil/empty", got.Defaults.AllowedResourceTypes)
	}
}
