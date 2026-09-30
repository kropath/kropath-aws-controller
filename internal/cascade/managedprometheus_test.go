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

var zeroKropathMP = cascade.ManagedPrometheusKropathSection{}
var zeroMPCfg = cascade.ManagedPrometheusConfigSection{}

func mergeMPAll(
	globalKropathMandatory,
	localKropathMandatory cascade.ManagedPrometheusKropathSection,
	globalMPCfgMandatory,
	localMPCfgMandatory,
	localMPCfgDefaults,
	globalMPCfgDefaults cascade.ManagedPrometheusConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.ManagedPrometheusKropathSection,
) cascade.EffectiveManagedPrometheusConfig {
	return cascade.MergeManagedPrometheusCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalMPCfgMandatory,
		localMPCfgMandatory,
		localMPCfgDefaults,
		globalMPCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeManagedPrometheusCascade_AC1 — globalKropathConfig.mandatory.managedprometheus.alias
// at level 1 propagates to effCfg.mandatory.alias.
func TestMergeManagedPrometheusCascade_AC1(t *testing.T) {
	got := mergeMPAll(
		cascade.ManagedPrometheusKropathSection{Alias: "prod-amp"}, // level 1
		zeroKropathMP,
		zeroMPCfg,
		zeroMPCfg,
		zeroMPCfg,
		zeroMPCfg,
		zeroKropathMP,
		zeroKropathMP,
	)

	if got.Mandatory.Alias != "prod-amp" {
		t.Errorf("AC-1: mandatory.alias = %q, want prod-amp (level 1 wins)", got.Mandatory.Alias)
	}
	if got.Defaults.Alias != "" {
		t.Errorf("AC-1: defaults.alias = %q, must not bleed from mandatory", got.Defaults.Alias)
	}
}

// TestMergeManagedPrometheusCascade_AC2 — globalMPConfig.mandatory.logGroupARN at level 3
// wins when levels 1-2 are empty.
func TestMergeManagedPrometheusCascade_AC2(t *testing.T) {
	arn := "arn:aws:logs:us-east-1:123:log-group/amp"
	got := mergeMPAll(
		zeroKropathMP,
		zeroKropathMP,
		cascade.ManagedPrometheusConfigSection{LogGroupARN: arn}, // level 3
		zeroMPCfg,
		zeroMPCfg,
		zeroMPCfg,
		zeroKropathMP,
		zeroKropathMP,
	)

	if got.Mandatory.LogGroupARN != arn {
		t.Errorf("AC-2: mandatory.logGroupARN = %q, want %q (level 3 wins when 1-2 empty)", got.Mandatory.LogGroupARN, arn)
	}
}

// TestMergeManagedPrometheusCascade_AC3 — localMPConfig.defaults.logGroupARN at level 6
// propagates; mandatory stays empty.
func TestMergeManagedPrometheusCascade_AC3(t *testing.T) {
	arn := "arn:aws:logs:us-east-1:123:log-group/default-amp"
	got := mergeMPAll(
		zeroKropathMP,
		zeroKropathMP,
		zeroMPCfg,
		zeroMPCfg,
		cascade.ManagedPrometheusConfigSection{LogGroupARN: arn}, // level 6
		zeroMPCfg,
		zeroKropathMP,
		zeroKropathMP,
	)

	if got.Mandatory.LogGroupARN != "" {
		t.Errorf("AC-3: mandatory.logGroupARN = %q, want empty", got.Mandatory.LogGroupARN)
	}
	if got.Defaults.LogGroupARN != arn {
		t.Errorf("AC-3: defaults.logGroupARN = %q, want %q (level 6)", got.Defaults.LogGroupARN, arn)
	}
}

// TestMergeManagedPrometheusCascade_AC4 — when no tier sets namingTemplate, defaults.namingTemplate
// falls back to "{namespace}-{name}".
func TestMergeManagedPrometheusCascade_AC4(t *testing.T) {
	got := mergeMPAll(
		zeroKropathMP, zeroKropathMP,
		zeroMPCfg, zeroMPCfg, zeroMPCfg, zeroMPCfg,
		zeroKropathMP, zeroKropathMP,
	)

	if got.Defaults.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("AC-4: defaults.namingTemplate = %q, want {namespace}-{name} (fallback)", got.Defaults.NamingTemplate)
	}
}

// TestMergeManagedPrometheusCascade_AC5 — localMPCfgDefaults.namingTemplate (level 6) wins over
// globalMPCfgDefaults.namingTemplate (level 7).
func TestMergeManagedPrometheusCascade_AC5(t *testing.T) {
	got := mergeMPAll(
		zeroKropathMP,
		zeroKropathMP,
		zeroMPCfg,
		zeroMPCfg,
		cascade.ManagedPrometheusConfigSection{NamingTemplate: "team-{namespace}-{name}"}, // level 6
		cascade.ManagedPrometheusConfigSection{NamingTemplate: "{namespace}-{name}"},      // level 7
		zeroKropathMP,
		zeroKropathMP,
	)

	if got.Defaults.NamingTemplate != "team-{namespace}-{name}" {
		t.Errorf("AC-5: defaults.namingTemplate = %q, want team-{namespace}-{name} (level 6 wins over 7)", got.Defaults.NamingTemplate)
	}
}

// TestMergeManagedPrometheusCascade_AC6 — globalMPCfgMandatory.namingTemplate at level 3
// propagates to effCfg.mandatory.namingTemplate.
func TestMergeManagedPrometheusCascade_AC6(t *testing.T) {
	got := mergeMPAll(
		zeroKropathMP,
		zeroKropathMP,
		cascade.ManagedPrometheusConfigSection{NamingTemplate: "corp-{namespace}-{name}"}, // level 3
		zeroMPCfg,
		zeroMPCfg,
		zeroMPCfg,
		zeroKropathMP,
		zeroKropathMP,
	)

	if got.Mandatory.NamingTemplate != "corp-{namespace}-{name}" {
		t.Errorf("AC-6: mandatory.namingTemplate = %q, want corp-{namespace}-{name} (level 3)", got.Mandatory.NamingTemplate)
	}
}

// TestMergeManagedPrometheusCascade_AC7 — KropathConfig.mandatory.tags and
// MPConfig.mandatory.tags are union-merged into effCfg.mandatory.tags.
func TestMergeManagedPrometheusCascade_AC7(t *testing.T) {
	got := mergeMPAll(
		cascade.ManagedPrometheusKropathSection{Tags: map[string]string{"cost-centre": "infra"}},   // level 1
		zeroKropathMP,
		zeroMPCfg,
		cascade.ManagedPrometheusConfigSection{Tags: map[string]string{"env": "prod"}}, // level 4
		zeroMPCfg,
		zeroMPCfg,
		zeroKropathMP,
		zeroKropathMP,
	)

	if got.Mandatory.Tags["cost-centre"] != "infra" {
		t.Errorf("AC-7: mandatory.tags[cost-centre] = %q, want infra", got.Mandatory.Tags["cost-centre"])
	}
	if got.Mandatory.Tags["env"] != "prod" {
		t.Errorf("AC-7: mandatory.tags[env] = %q, want prod", got.Mandatory.Tags["env"])
	}
}

// TestMergeManagedPrometheusCascade_AC8 — MPConfig.mandatory.syncedLabels from global (L3) and
// local (L4) tiers are union-merged; L3 wins on key conflict.
func TestMergeManagedPrometheusCascade_AC8(t *testing.T) {
	got := mergeMPAll(
		zeroKropathMP,
		zeroKropathMP,
		cascade.ManagedPrometheusConfigSection{SyncedLabels: map[string]string{"tier": "global", "data-class": "internal"}}, // level 3
		cascade.ManagedPrometheusConfigSection{SyncedLabels: map[string]string{"tier": "local", "env": "prod"}},             // level 4
		zeroMPCfg,
		zeroMPCfg,
		zeroKropathMP,
		zeroKropathMP,
	)

	if got.Mandatory.SyncedLabels["data-class"] != "internal" {
		t.Errorf("AC-8: mandatory.syncedLabels[data-class] = %q, want internal", got.Mandatory.SyncedLabels["data-class"])
	}
	if got.Mandatory.SyncedLabels["env"] != "prod" {
		t.Errorf("AC-8: mandatory.syncedLabels[env] = %q, want prod", got.Mandatory.SyncedLabels["env"])
	}
	if got.Mandatory.SyncedLabels["tier"] != "global" {
		t.Errorf("AC-8: mandatory.syncedLabels[tier] = %q, want global (L3 wins over L4)", got.Mandatory.SyncedLabels["tier"])
	}
}

// TestMergeManagedPrometheusCascade_AllAbsent — when all sources are zero, effectiveConfig is
// all-zero except defaults.namingTemplate which falls back to "{namespace}-{name}".
func TestMergeManagedPrometheusCascade_AllAbsent(t *testing.T) {
	got := mergeMPAll(
		zeroKropathMP, zeroKropathMP,
		zeroMPCfg, zeroMPCfg, zeroMPCfg, zeroMPCfg,
		zeroKropathMP, zeroKropathMP,
	)

	if got.Mandatory.Alias != "" {
		t.Errorf("all-absent: mandatory.alias = %q, want empty", got.Mandatory.Alias)
	}
	if got.Mandatory.LogGroupARN != "" {
		t.Errorf("all-absent: mandatory.logGroupARN = %q, want empty", got.Mandatory.LogGroupARN)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("all-absent: mandatory.namingTemplate = %q, want empty", got.Mandatory.NamingTemplate)
	}
	if len(got.Mandatory.Tags) != 0 {
		t.Errorf("all-absent: mandatory.tags = %v, want empty", got.Mandatory.Tags)
	}
	if got.Defaults.Alias != "" {
		t.Errorf("all-absent: defaults.alias = %q, want empty", got.Defaults.Alias)
	}
	if got.Defaults.LogGroupARN != "" {
		t.Errorf("all-absent: defaults.logGroupARN = %q, want empty", got.Defaults.LogGroupARN)
	}
	if got.Defaults.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("all-absent: defaults.namingTemplate = %q, want {namespace}-{name} (fallback)", got.Defaults.NamingTemplate)
	}
	if len(got.Defaults.Tags) != 0 {
		t.Errorf("all-absent: defaults.tags = %v, want empty", got.Defaults.Tags)
	}
}

// TestMergeManagedPrometheusCascade_MandatoryPriorityOrder — verifies mandatory priority order
// for alias: level 1 > 2 > 3 > 4.
func TestMergeManagedPrometheusCascade_MandatoryPriorityOrder(t *testing.T) {
	cases := []struct {
		name                   string
		globalKropathMandatory cascade.ManagedPrometheusKropathSection
		localKropathMandatory  cascade.ManagedPrometheusKropathSection
		globalMPCfgMandatory   cascade.ManagedPrometheusConfigSection
		localMPCfgMandatory    cascade.ManagedPrometheusConfigSection
		wantAlias              string
	}{
		{
			name:                   "level1-wins",
			globalKropathMandatory: cascade.ManagedPrometheusKropathSection{Alias: "level1"},
			localKropathMandatory:  cascade.ManagedPrometheusKropathSection{Alias: "level2"},
			globalMPCfgMandatory:   cascade.ManagedPrometheusConfigSection{Alias: "level3"},
			localMPCfgMandatory:    cascade.ManagedPrometheusConfigSection{Alias: "level4"},
			wantAlias:              "level1",
		},
		{
			name:                   "level2-wins-when-1-absent",
			globalKropathMandatory: zeroKropathMP,
			localKropathMandatory:  cascade.ManagedPrometheusKropathSection{Alias: "level2"},
			globalMPCfgMandatory:   cascade.ManagedPrometheusConfigSection{Alias: "level3"},
			localMPCfgMandatory:    cascade.ManagedPrometheusConfigSection{Alias: "level4"},
			wantAlias:              "level2",
		},
		{
			name:                   "level3-wins-when-1-2-absent",
			globalKropathMandatory: zeroKropathMP,
			localKropathMandatory:  zeroKropathMP,
			globalMPCfgMandatory:   cascade.ManagedPrometheusConfigSection{Alias: "level3"},
			localMPCfgMandatory:    cascade.ManagedPrometheusConfigSection{Alias: "level4"},
			wantAlias:              "level3",
		},
		{
			name:                   "level4-wins-when-1-2-3-absent",
			globalKropathMandatory: zeroKropathMP,
			localKropathMandatory:  zeroKropathMP,
			globalMPCfgMandatory:   zeroMPCfg,
			localMPCfgMandatory:    cascade.ManagedPrometheusConfigSection{Alias: "level4"},
			wantAlias:              "level4",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeMPAll(
				tc.globalKropathMandatory,
				tc.localKropathMandatory,
				tc.globalMPCfgMandatory,
				tc.localMPCfgMandatory,
				zeroMPCfg,
				zeroMPCfg,
				zeroKropathMP,
				zeroKropathMP,
			)
			if got.Mandatory.Alias != tc.wantAlias {
				t.Errorf("mandatory.alias = %q, want %q", got.Mandatory.Alias, tc.wantAlias)
			}
		})
	}
}

// TestMergeManagedPrometheusCascade_DefaultsPriorityOrder — verifies defaults priority order
// for logGroupARN: level 6 > 7 > 8 > 9.
func TestMergeManagedPrometheusCascade_DefaultsPriorityOrder(t *testing.T) {
	cases := []struct {
		name                  string
		localMPCfgDefaults    cascade.ManagedPrometheusConfigSection
		globalMPCfgDefaults   cascade.ManagedPrometheusConfigSection
		localKropathDefaults  cascade.ManagedPrometheusKropathSection
		globalKropathDefaults cascade.ManagedPrometheusKropathSection
		wantLogGroupARN       string
	}{
		{
			name:                  "level6-wins",
			localMPCfgDefaults:    cascade.ManagedPrometheusConfigSection{LogGroupARN: "level6-arn"},
			globalMPCfgDefaults:   cascade.ManagedPrometheusConfigSection{LogGroupARN: "level7-arn"},
			localKropathDefaults:  cascade.ManagedPrometheusKropathSection{LogGroupARN: "level8-arn"},
			globalKropathDefaults: cascade.ManagedPrometheusKropathSection{LogGroupARN: "level9-arn"},
			wantLogGroupARN:       "level6-arn",
		},
		{
			name:                  "level7-wins-when-6-absent",
			localMPCfgDefaults:    zeroMPCfg,
			globalMPCfgDefaults:   cascade.ManagedPrometheusConfigSection{LogGroupARN: "level7-arn"},
			localKropathDefaults:  cascade.ManagedPrometheusKropathSection{LogGroupARN: "level8-arn"},
			globalKropathDefaults: cascade.ManagedPrometheusKropathSection{LogGroupARN: "level9-arn"},
			wantLogGroupARN:       "level7-arn",
		},
		{
			name:                  "level8-wins-when-6-7-absent",
			localMPCfgDefaults:    zeroMPCfg,
			globalMPCfgDefaults:   zeroMPCfg,
			localKropathDefaults:  cascade.ManagedPrometheusKropathSection{LogGroupARN: "level8-arn"},
			globalKropathDefaults: cascade.ManagedPrometheusKropathSection{LogGroupARN: "level9-arn"},
			wantLogGroupARN:       "level8-arn",
		},
		{
			name:                  "level9-wins-when-6-7-8-absent",
			localMPCfgDefaults:    zeroMPCfg,
			globalMPCfgDefaults:   zeroMPCfg,
			localKropathDefaults:  zeroKropathMP,
			globalKropathDefaults: cascade.ManagedPrometheusKropathSection{LogGroupARN: "level9-arn"},
			wantLogGroupARN:       "level9-arn",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeMPAll(
				zeroKropathMP,
				zeroKropathMP,
				zeroMPCfg,
				zeroMPCfg,
				tc.localMPCfgDefaults,
				tc.globalMPCfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
			)
			if got.Defaults.LogGroupARN != tc.wantLogGroupARN {
				t.Errorf("defaults.logGroupARN = %q, want %q", got.Defaults.LogGroupARN, tc.wantLogGroupARN)
			}
		})
	}
}

// TestMergeManagedPrometheusCascade_TagsKeyConflict — on key conflict, lower level number wins.
// Level 1 (globalKropathMandatory) wins over level 4 (localMPCfgMandatory).
func TestMergeManagedPrometheusCascade_TagsKeyConflict(t *testing.T) {
	got := mergeMPAll(
		cascade.ManagedPrometheusKropathSection{Tags: map[string]string{"env": "org-level"}},  // level 1
		zeroKropathMP,
		zeroMPCfg,
		cascade.ManagedPrometheusConfigSection{Tags: map[string]string{"env": "config-level"}}, // level 4
		zeroMPCfg,
		zeroMPCfg,
		zeroKropathMP,
		zeroKropathMP,
	)

	if got.Mandatory.Tags["env"] != "org-level" {
		t.Errorf("tag-conflict: mandatory.tags[env] = %q, want org-level (level 1 wins over level 4)", got.Mandatory.Tags["env"])
	}
}

// TestMergeManagedPrometheusCascade_NamingTemplateFallbackWhenMPCfgSet — when MPConfig level 6
// has a namingTemplate, fallback "{namespace}-{name}" is NOT used.
func TestMergeManagedPrometheusCascade_NamingTemplateFallbackWhenMPCfgSet(t *testing.T) {
	got := mergeMPAll(
		zeroKropathMP,
		zeroKropathMP,
		zeroMPCfg,
		zeroMPCfg,
		cascade.ManagedPrometheusConfigSection{NamingTemplate: "custom-{namespace}-{name}"}, // level 6
		zeroMPCfg,
		zeroKropathMP,
		zeroKropathMP,
	)

	if got.Defaults.NamingTemplate != "custom-{namespace}-{name}" {
		t.Errorf("naming-fallback: defaults.namingTemplate = %q, want custom-{namespace}-{name}", got.Defaults.NamingTemplate)
	}
}
