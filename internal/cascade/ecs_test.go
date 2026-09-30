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

// zeroKropathECS is a zero-value ECSKropathSection (absent source).
var zeroKropathECS = cascade.ECSKropathSection{}

// zeroECSCfg is a zero-value ECSConfigSection (absent source).
var zeroECSCfg = cascade.ECSConfigSection{}

// mergeECSAll calls MergeECSCascade with all eight inputs.
func mergeECSAll(
	globalKropathMandatory,
	localKropathMandatory cascade.ECSKropathSection,
	globalECSCfgMandatory,
	localECSCfgMandatory,
	localECSCfgDefaults,
	globalECSCfgDefaults cascade.ECSConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.ECSKropathSection,
) cascade.EffectiveECSConfig {
	return cascade.MergeECSCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalECSCfgMandatory,
		localECSCfgMandatory,
		localECSCfgDefaults,
		globalECSCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeECSCascade_AC_C1 — globalKropathConfig.mandatory.ecs.containerInsights=true
// at level 1 propagates to effectiveConfig.mandatory.containerInsights (mirrors spec AC-8).
func TestMergeECSCascade_AC_C1(t *testing.T) {
	got := mergeECSAll(
		cascade.ECSKropathSection{ContainerInsights: true}, // level 1
		zeroKropathECS,
		zeroECSCfg,
		zeroECSCfg,
		zeroECSCfg,
		zeroECSCfg,
		zeroKropathECS,
		zeroKropathECS,
	)

	if !got.Mandatory.ContainerInsights {
		t.Error("AC-C1: mandatory.containerInsights should be true when set at level 1")
	}
	if got.Defaults.ContainerInsights {
		t.Error("AC-C1: defaults.containerInsights must not bleed from mandatory")
	}
}

// TestMergeECSCascade_AC_C2 — globalKropathConfig.mandatory.ecs.defaultLaunchType="FARGATE"
// at level 1 propagates to effectiveConfig.mandatory.defaultLaunchType.
func TestMergeECSCascade_AC_C2(t *testing.T) {
	got := mergeECSAll(
		cascade.ECSKropathSection{DefaultLaunchType: "FARGATE"}, // level 1
		zeroKropathECS,
		zeroECSCfg,
		zeroECSCfg,
		zeroECSCfg,
		zeroECSCfg,
		zeroKropathECS,
		zeroKropathECS,
	)

	if got.Mandatory.DefaultLaunchType != "FARGATE" {
		t.Errorf("AC-C2: mandatory.defaultLaunchType = %q, want FARGATE", got.Mandatory.DefaultLaunchType)
	}
	if got.Defaults.DefaultLaunchType != "" {
		t.Errorf("AC-C2: defaults.defaultLaunchType must not bleed from mandatory, got %q", got.Defaults.DefaultLaunchType)
	}
}

// TestMergeECSCascade_AC_C3 — globalECSConfig.mandatory.containerInsights=true at
// level 3 propagates when levels 1-2 are absent.
func TestMergeECSCascade_AC_C3(t *testing.T) {
	got := mergeECSAll(
		zeroKropathECS,
		zeroKropathECS,
		cascade.ECSConfigSection{ContainerInsights: true}, // level 3
		zeroECSCfg,
		zeroECSCfg,
		zeroECSCfg,
		zeroKropathECS,
		zeroKropathECS,
	)

	if !got.Mandatory.ContainerInsights {
		t.Error("AC-C3: mandatory.containerInsights should be true when set at level 3 and levels 1-2 absent")
	}
}

// TestMergeECSCascade_AC_C4 — globalECSConfig.mandatory.defaultLaunchType="EC2" at
// level 3 propagates when levels 1-2 are absent.
func TestMergeECSCascade_AC_C4(t *testing.T) {
	got := mergeECSAll(
		zeroKropathECS,
		zeroKropathECS,
		cascade.ECSConfigSection{DefaultLaunchType: "EC2"}, // level 3
		zeroECSCfg,
		zeroECSCfg,
		zeroECSCfg,
		zeroKropathECS,
		zeroKropathECS,
	)

	if got.Mandatory.DefaultLaunchType != "EC2" {
		t.Errorf("AC-C4: mandatory.defaultLaunchType = %q, want EC2", got.Mandatory.DefaultLaunchType)
	}
}

// TestMergeECSCascade_AC_C5 — level 1 wins over level 3 for containerInsights.
func TestMergeECSCascade_AC_C5(t *testing.T) {
	// Both level 1 and level 3 set containerInsights; level 1 should win (firstTrue picks first true).
	got := mergeECSAll(
		cascade.ECSKropathSection{ContainerInsights: true},  // level 1
		zeroKropathECS,
		cascade.ECSConfigSection{ContainerInsights: true},   // level 3
		zeroECSCfg,
		zeroECSCfg,
		zeroECSCfg,
		zeroKropathECS,
		zeroKropathECS,
	)

	if !got.Mandatory.ContainerInsights {
		t.Error("AC-C5: mandatory.containerInsights should be true (level 1 and 3 both true)")
	}

	// level 1 wins for defaultLaunchType over level 3 (firstNonEmptyString).
	got2 := mergeECSAll(
		cascade.ECSKropathSection{DefaultLaunchType: "FARGATE"}, // level 1
		zeroKropathECS,
		cascade.ECSConfigSection{DefaultLaunchType: "EC2"},       // level 3
		zeroECSCfg,
		zeroECSCfg,
		zeroECSCfg,
		zeroKropathECS,
		zeroKropathECS,
	)

	if got2.Mandatory.DefaultLaunchType != "FARGATE" {
		t.Errorf("AC-C5: level-1 must win over level-3 for defaultLaunchType; got %q, want FARGATE", got2.Mandatory.DefaultLaunchType)
	}
}

// TestMergeECSCascade_AC_C6 — defaults.containerInsights cascade order: level 6 > 7 > 8 > 9.
func TestMergeECSCascade_AC_C6(t *testing.T) {
	cases := []struct {
		name                  string
		localECSCfgDefaults   cascade.ECSConfigSection
		globalECSCfgDefaults  cascade.ECSConfigSection
		localKropathDefaults  cascade.ECSKropathSection
		globalKropathDefaults cascade.ECSKropathSection
		want                  bool
	}{
		{
			name:                "level6-wins",
			localECSCfgDefaults: cascade.ECSConfigSection{ContainerInsights: true},
			want:                true,
		},
		{
			name:                 "level7-wins-when-6-absent",
			globalECSCfgDefaults: cascade.ECSConfigSection{ContainerInsights: true},
			want:                 true,
		},
		{
			name:                 "level8-wins-when-6-7-absent",
			localKropathDefaults: cascade.ECSKropathSection{ContainerInsights: true},
			want:                 true,
		},
		{
			name:                  "level9-fallback",
			globalKropathDefaults: cascade.ECSKropathSection{ContainerInsights: true},
			want:                  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeECSAll(
				zeroKropathECS, zeroKropathECS,
				zeroECSCfg, zeroECSCfg,
				tc.localECSCfgDefaults,
				tc.globalECSCfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
			)
			if got.Defaults.ContainerInsights != tc.want {
				t.Errorf("defaults.containerInsights = %v, want %v", got.Defaults.ContainerInsights, tc.want)
			}
		})
	}
}

// TestMergeECSCascade_AC_C7 — defaults.defaultLaunchType cascade order: level 6 > 7 > 8 > 9.
func TestMergeECSCascade_AC_C7(t *testing.T) {
	cases := []struct {
		name                  string
		localECSCfgDefaults   cascade.ECSConfigSection
		globalECSCfgDefaults  cascade.ECSConfigSection
		localKropathDefaults  cascade.ECSKropathSection
		globalKropathDefaults cascade.ECSKropathSection
		want                  string
	}{
		{
			name:                "level6-wins",
			localECSCfgDefaults: cascade.ECSConfigSection{DefaultLaunchType: "FARGATE"},
			want:                "FARGATE",
		},
		{
			name:                 "level7-wins-when-6-absent",
			globalECSCfgDefaults: cascade.ECSConfigSection{DefaultLaunchType: "EC2"},
			want:                 "EC2",
		},
		{
			name:                 "level8-wins-when-6-7-absent",
			localKropathDefaults: cascade.ECSKropathSection{DefaultLaunchType: "FARGATE"},
			want:                 "FARGATE",
		},
		{
			name:                  "level9-fallback",
			globalKropathDefaults: cascade.ECSKropathSection{DefaultLaunchType: "EC2"},
			want:                  "EC2",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeECSAll(
				zeroKropathECS, zeroKropathECS,
				zeroECSCfg, zeroECSCfg,
				tc.localECSCfgDefaults,
				tc.globalECSCfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
			)
			if got.Defaults.DefaultLaunchType != tc.want {
				t.Errorf("defaults.defaultLaunchType = %q, want %q", got.Defaults.DefaultLaunchType, tc.want)
			}
		})
	}
}

// TestMergeECSCascade_AC_C8 — KropathConfig.mandatory.tags and ECSConfig.mandatory.tags
// are union-merged into effectiveConfig.mandatory.tags; KropathConfig (level 1) wins on
// key conflicts (mirrors spec AC-11).
func TestMergeECSCascade_AC_C8(t *testing.T) {
	got := mergeECSAll(
		cascade.ECSKropathSection{Tags: map[string]string{"env": "mandatory", "cost-centre": "infra"}}, // level 1
		zeroKropathECS,
		cascade.ECSConfigSection{Tags: map[string]string{"env": "ecscfg", "service-type": "containers"}}, // level 3
		zeroECSCfg,
		zeroECSCfg,
		zeroECSCfg,
		zeroKropathECS,
		zeroKropathECS,
	)

	if got.Mandatory.Tags["env"] != "mandatory" {
		t.Errorf("AC-C8: level-1 KropathConfig must win on key conflict; tags[env] = %q, want mandatory", got.Mandatory.Tags["env"])
	}
	if got.Mandatory.Tags["cost-centre"] != "infra" {
		t.Errorf("AC-C8: tags[cost-centre] = %q, want infra", got.Mandatory.Tags["cost-centre"])
	}
	if got.Mandatory.Tags["service-type"] != "containers" {
		t.Errorf("AC-C8: tags[service-type] = %q, want containers", got.Mandatory.Tags["service-type"])
	}
}

// TestMergeECSCascade_AC_C9 — syncedLabels from ECSConfig mandatory levels only;
// globalECSCfg (level 3) wins on key conflicts in mandatory tier (mirrors spec AC-12).
func TestMergeECSCascade_AC_C9(t *testing.T) {
	got := mergeECSAll(
		// KropathSection SyncedLabels should NOT appear in mandatory.syncedLabels (ECSConfig-only).
		cascade.ECSKropathSection{SyncedLabels: map[string]string{"should-not-appear": "yes"}},
		zeroKropathECS,
		cascade.ECSConfigSection{SyncedLabels: map[string]string{"workload-type": "ecs", "tier": "global"}}, // level 3
		cascade.ECSConfigSection{SyncedLabels: map[string]string{"workload-type": "local", "ns-label": "x"}}, // level 4
		zeroECSCfg,
		zeroECSCfg,
		zeroKropathECS,
		zeroKropathECS,
	)

	if got.Mandatory.SyncedLabels["workload-type"] != "ecs" {
		t.Errorf("AC-C9: level-3 globalECSCfg must win on key conflict; syncedLabels[workload-type] = %q, want ecs", got.Mandatory.SyncedLabels["workload-type"])
	}
	if got.Mandatory.SyncedLabels["tier"] != "global" {
		t.Errorf("AC-C9: syncedLabels[tier] = %q, want global", got.Mandatory.SyncedLabels["tier"])
	}
	if got.Mandatory.SyncedLabels["ns-label"] != "x" {
		t.Errorf("AC-C9: syncedLabels[ns-label] = %q, want x", got.Mandatory.SyncedLabels["ns-label"])
	}
	if _, ok := got.Mandatory.SyncedLabels["should-not-appear"]; ok {
		t.Error("AC-C9: KropathSection SyncedLabels must NOT appear in mandatory.syncedLabels (ECSConfig-only field)")
	}
}

// TestMergeECSCascade_AC_C10 — mandatory tier fields must not bleed into defaults and vice versa.
func TestMergeECSCascade_AC_C10(t *testing.T) {
	got := mergeECSAll(
		cascade.ECSKropathSection{ContainerInsights: true, DefaultLaunchType: "FARGATE"},
		zeroKropathECS,
		cascade.ECSConfigSection{
			DefaultPlatformVersion: "1.4.0",
			DefaultNetworkMode:     "awsvpc",
			NamingTemplate:         "{namespace}-{name}-mand",
		},
		zeroECSCfg,
		cascade.ECSConfigSection{
			DefaultPlatformVersion: "LATEST",
			DefaultNetworkMode:     "bridge",
			NamingTemplate:         "{namespace}-{name}-def",
		},
		zeroECSCfg,
		zeroKropathECS,
		zeroKropathECS,
	)

	if !got.Mandatory.ContainerInsights {
		t.Error("AC-C10: mandatory.containerInsights should be true")
	}
	if got.Mandatory.DefaultLaunchType != "FARGATE" {
		t.Errorf("AC-C10: mandatory.defaultLaunchType = %q, want FARGATE", got.Mandatory.DefaultLaunchType)
	}
	if got.Mandatory.DefaultPlatformVersion != "1.4.0" {
		t.Errorf("AC-C10: mandatory.defaultPlatformVersion = %q, want 1.4.0", got.Mandatory.DefaultPlatformVersion)
	}
	if got.Mandatory.NamingTemplate != "{namespace}-{name}-mand" {
		t.Errorf("AC-C10: mandatory.namingTemplate = %q, want {namespace}-{name}-mand", got.Mandatory.NamingTemplate)
	}

	if got.Defaults.ContainerInsights {
		t.Error("AC-C10: defaults.containerInsights must not bleed from mandatory")
	}
	if got.Defaults.DefaultLaunchType != "" {
		t.Errorf("AC-C10: defaults.defaultLaunchType must not bleed from mandatory, got %q", got.Defaults.DefaultLaunchType)
	}
	if got.Defaults.DefaultPlatformVersion != "LATEST" {
		t.Errorf("AC-C10: defaults.defaultPlatformVersion = %q, want LATEST", got.Defaults.DefaultPlatformVersion)
	}
	if got.Defaults.NamingTemplate != "{namespace}-{name}-def" {
		t.Errorf("AC-C10: defaults.namingTemplate = %q, want {namespace}-{name}-def", got.Defaults.NamingTemplate)
	}
}

// TestMergeECSCascade_AllAbsent — when all sources are zero, all effectiveConfig fields
// are zero (permissive; no governance enforced).
func TestMergeECSCascade_AllAbsent(t *testing.T) {
	got := mergeECSAll(
		zeroKropathECS, zeroKropathECS,
		zeroECSCfg, zeroECSCfg, zeroECSCfg, zeroECSCfg,
		zeroKropathECS, zeroKropathECS,
	)

	if got.Mandatory.ContainerInsights {
		t.Error("all-absent: mandatory.containerInsights should be false")
	}
	if got.Mandatory.DefaultLaunchType != "" {
		t.Errorf("all-absent: mandatory.defaultLaunchType = %q, want empty", got.Mandatory.DefaultLaunchType)
	}
	if got.Mandatory.DefaultPlatformVersion != "" {
		t.Errorf("all-absent: mandatory.defaultPlatformVersion = %q, want empty", got.Mandatory.DefaultPlatformVersion)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("all-absent: mandatory.namingTemplate = %q, want empty", got.Mandatory.NamingTemplate)
	}
	if got.Defaults.ContainerInsights {
		t.Error("all-absent: defaults.containerInsights should be false")
	}
	if got.Defaults.DefaultLaunchType != "" {
		t.Errorf("all-absent: defaults.defaultLaunchType = %q, want empty", got.Defaults.DefaultLaunchType)
	}
}

// TestMergeECSCascade_MandatoryCascadeOrder — mandatory containerInsights priority order
// (level 1 > 2 > 3 > 4).
func TestMergeECSCascade_MandatoryCascadeOrder(t *testing.T) {
	cases := []struct {
		name                   string
		globalKropathMandatory cascade.ECSKropathSection
		localKropathMandatory  cascade.ECSKropathSection
		globalECSCfgMandatory  cascade.ECSConfigSection
		localECSCfgMandatory   cascade.ECSConfigSection
		want                   bool
	}{
		{
			name:                   "level1-sets-true",
			globalKropathMandatory: cascade.ECSKropathSection{ContainerInsights: true},
			want:                   true,
		},
		{
			name:                  "level2-sets-true-when-1-absent",
			localKropathMandatory: cascade.ECSKropathSection{ContainerInsights: true},
			want:                  true,
		},
		{
			name:                  "level3-sets-true-when-1-2-absent",
			globalECSCfgMandatory: cascade.ECSConfigSection{ContainerInsights: true},
			want:                  true,
		},
		{
			name:                 "level4-sets-true-when-1-2-3-absent",
			localECSCfgMandatory: cascade.ECSConfigSection{ContainerInsights: true},
			want:                 true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeECSAll(
				tc.globalKropathMandatory,
				tc.localKropathMandatory,
				tc.globalECSCfgMandatory,
				tc.localECSCfgMandatory,
				zeroECSCfg, zeroECSCfg,
				zeroKropathECS, zeroKropathECS,
			)
			if got.Mandatory.ContainerInsights != tc.want {
				t.Errorf("containerInsights = %v, want %v", got.Mandatory.ContainerInsights, tc.want)
			}
		})
	}
}

// TestMergeECSCascade_DefaultsTagUnionMerge — defaults tags union from KropathConfig and
// ECSConfig defaults levels; level 6 wins on key conflicts.
func TestMergeECSCascade_DefaultsTagUnionMerge(t *testing.T) {
	got := mergeECSAll(
		zeroKropathECS,
		zeroKropathECS,
		zeroECSCfg,
		zeroECSCfg,
		cascade.ECSConfigSection{Tags: map[string]string{"env": "dev", "team": "platform"}},    // level 6
		cascade.ECSConfigSection{Tags: map[string]string{"env": "staging", "region": "us-east-1"}}, // level 7
		cascade.ECSKropathSection{Tags: map[string]string{"env": "org-default", "cost-centre": "platform"}}, // level 8
		zeroKropathECS,
	)

	if got.Defaults.Tags["env"] != "dev" {
		t.Errorf("defaults tag union: level-6 must win on key conflict; tags[env] = %q, want dev", got.Defaults.Tags["env"])
	}
	if got.Defaults.Tags["team"] != "platform" {
		t.Errorf("defaults tag union: tags[team] = %q, want platform", got.Defaults.Tags["team"])
	}
	if got.Defaults.Tags["region"] != "us-east-1" {
		t.Errorf("defaults tag union: tags[region] = %q, want us-east-1", got.Defaults.Tags["region"])
	}
	if got.Defaults.Tags["cost-centre"] != "platform" {
		t.Errorf("defaults tag union: tags[cost-centre] = %q, want platform", got.Defaults.Tags["cost-centre"])
	}
}

// TestMergeECSCascade_ECSConfigOnlyFields — defaultPlatformVersion, defaultNetworkMode,
// defaultTaskCPU, defaultTaskMemory, and namingTemplate appear only from ECSConfig levels.
func TestMergeECSCascade_ECSConfigOnlyFields(t *testing.T) {
	got := mergeECSAll(
		zeroKropathECS,
		zeroKropathECS,
		cascade.ECSConfigSection{
			DefaultPlatformVersion: "1.4.0",
			DefaultNetworkMode:     "awsvpc",
			DefaultTaskCPU:         "512",
			DefaultTaskMemory:      "1024",
			NamingTemplate:         "{namespace}-{name}",
		},
		zeroECSCfg,
		zeroECSCfg,
		cascade.ECSConfigSection{
			DefaultPlatformVersion: "LATEST",
			DefaultNetworkMode:     "bridge",
			DefaultTaskCPU:         "256",
			DefaultTaskMemory:      "512",
			NamingTemplate:         "{namespace}-{name}-default",
		},
		zeroKropathECS,
		zeroKropathECS,
	)

	if got.Mandatory.DefaultPlatformVersion != "1.4.0" {
		t.Errorf("mandatory.defaultPlatformVersion = %q, want 1.4.0", got.Mandatory.DefaultPlatformVersion)
	}
	if got.Mandatory.DefaultNetworkMode != "awsvpc" {
		t.Errorf("mandatory.defaultNetworkMode = %q, want awsvpc", got.Mandatory.DefaultNetworkMode)
	}
	if got.Mandatory.DefaultTaskCPU != "512" {
		t.Errorf("mandatory.defaultTaskCPU = %q, want 512", got.Mandatory.DefaultTaskCPU)
	}
	if got.Mandatory.DefaultTaskMemory != "1024" {
		t.Errorf("mandatory.defaultTaskMemory = %q, want 1024", got.Mandatory.DefaultTaskMemory)
	}
	if got.Mandatory.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("mandatory.namingTemplate = %q, want {namespace}-{name}", got.Mandatory.NamingTemplate)
	}

	if got.Defaults.DefaultPlatformVersion != "LATEST" {
		t.Errorf("defaults.defaultPlatformVersion = %q, want LATEST", got.Defaults.DefaultPlatformVersion)
	}
	if got.Defaults.DefaultNetworkMode != "bridge" {
		t.Errorf("defaults.defaultNetworkMode = %q, want bridge", got.Defaults.DefaultNetworkMode)
	}
	if got.Defaults.DefaultTaskCPU != "256" {
		t.Errorf("defaults.defaultTaskCPU = %q, want 256", got.Defaults.DefaultTaskCPU)
	}
	if got.Defaults.DefaultTaskMemory != "512" {
		t.Errorf("defaults.defaultTaskMemory = %q, want 512", got.Defaults.DefaultTaskMemory)
	}
	if got.Defaults.NamingTemplate != "{namespace}-{name}-default" {
		t.Errorf("defaults.namingTemplate = %q, want {namespace}-{name}-default", got.Defaults.NamingTemplate)
	}
}
