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

func mqBoolPtr(b bool) *bool { return &b }

func TestMergeMQCascade_CC1_GlobalKropathMandatoryLevel1(t *testing.T) {
	// CC-1: globalKropathConfig.mandatory.mq.publiclyAccessible=false (level 1) propagates.
	// false is a valid enforced value — nil is the only skip sentinel.
	result := cascade.MergeMQCascade(
		cascade.MQKropathSection{PubliclyAccessible: mqBoolPtr(false)}, // level 1 (global kpc mandatory)
		cascade.MQKropathSection{},                                    // level 2
		cascade.MQConfigSection{},                                     // level 3
		cascade.MQConfigSection{},                                     // level 4
		cascade.MQConfigSection{},                                     // level 6
		cascade.MQConfigSection{},                                     // level 7
		cascade.MQKropathSection{},                                    // level 8
		cascade.MQKropathSection{},                                    // level 9
	)
	if result.Mandatory.PubliclyAccessible == nil || *result.Mandatory.PubliclyAccessible != false {
		t.Errorf("CC-1: expected mandatory.publiclyAccessible=false from level 1, got %v", result.Mandatory.PubliclyAccessible)
	}
}

func TestMergeMQCascade_CC2_Level1WinsOverLevel3(t *testing.T) {
	// CC-2: globalKropathConfig.mandatory.mq.publiclyAccessible=false (level 1) wins over
	// globalMQCfgMandatory.publiclyAccessible=true (level 3). Level 1 is stronger.
	result := cascade.MergeMQCascade(
		cascade.MQKropathSection{PubliclyAccessible: mqBoolPtr(false)}, // level 1 wins
		cascade.MQKropathSection{},
		cascade.MQConfigSection{PubliclyAccessible: mqBoolPtr(true)}, // level 3 loses
		cascade.MQConfigSection{},
		cascade.MQConfigSection{},
		cascade.MQConfigSection{},
		cascade.MQKropathSection{},
		cascade.MQKropathSection{},
	)
	if result.Mandatory.PubliclyAccessible == nil || *result.Mandatory.PubliclyAccessible != false {
		t.Errorf("CC-2: expected mandatory.publiclyAccessible=false (level 1 wins), got %v", result.Mandatory.PubliclyAccessible)
	}
}

func TestMergeMQCascade_CC3_GlobalMQCfgMandatoryEngineTypeLevel3(t *testing.T) {
	// CC-3: globalMQCfgMandatory.engineType="ACTIVEMQ" (level 3) propagates when KropathConfig is empty.
	result := cascade.MergeMQCascade(
		cascade.MQKropathSection{},
		cascade.MQKropathSection{},
		cascade.MQConfigSection{EngineType: "ACTIVEMQ"}, // level 3
		cascade.MQConfigSection{},
		cascade.MQConfigSection{},
		cascade.MQConfigSection{},
		cascade.MQKropathSection{},
		cascade.MQKropathSection{},
	)
	if result.Mandatory.EngineType != "ACTIVEMQ" {
		t.Errorf("CC-3: expected mandatory.engineType=ACTIVEMQ from level 3, got %q", result.Mandatory.EngineType)
	}
}

func TestMergeMQCascade_CC4_LocalMQCfgMandatoryHostInstanceTypeLevel4(t *testing.T) {
	// CC-4: localMQCfgMandatory.hostInstanceType="mq.m5.large" (level 4) propagates.
	// hostInstanceType is MQConfig-only — KropathConfig.mq has no such field.
	result := cascade.MergeMQCascade(
		cascade.MQKropathSection{},
		cascade.MQKropathSection{},
		cascade.MQConfigSection{},
		cascade.MQConfigSection{HostInstanceType: "mq.m5.large"}, // level 4
		cascade.MQConfigSection{},
		cascade.MQConfigSection{},
		cascade.MQKropathSection{},
		cascade.MQKropathSection{},
	)
	if result.Mandatory.HostInstanceType != "mq.m5.large" {
		t.Errorf("CC-4: expected mandatory.hostInstanceType=mq.m5.large from level 4, got %q", result.Mandatory.HostInstanceType)
	}
}

func TestMergeMQCascade_CC5_Level3HostInstanceTypeWinsOverLevel4(t *testing.T) {
	// CC-5: globalMQCfgMandatory.hostInstanceType="mq.m5.xlarge" (level 3) wins over
	// localMQCfgMandatory.hostInstanceType="mq.m5.large" (level 4).
	result := cascade.MergeMQCascade(
		cascade.MQKropathSection{},
		cascade.MQKropathSection{},
		cascade.MQConfigSection{HostInstanceType: "mq.m5.xlarge"}, // level 3 wins
		cascade.MQConfigSection{HostInstanceType: "mq.m5.large"},  // level 4 loses
		cascade.MQConfigSection{},
		cascade.MQConfigSection{},
		cascade.MQKropathSection{},
		cascade.MQKropathSection{},
	)
	if result.Mandatory.HostInstanceType != "mq.m5.xlarge" {
		t.Errorf("CC-5: expected mandatory.hostInstanceType=mq.m5.xlarge (level 3 wins), got %q", result.Mandatory.HostInstanceType)
	}
}

func TestMergeMQCascade_CC6_LocalMQCfgDefaultsLevel6(t *testing.T) {
	// CC-6: localMQCfgDefaults.deploymentMode="SINGLE_INSTANCE" (level 6) propagates to defaults
	// when all mandatory levels are empty. Verifies defaults tier resolves correctly.
	result := cascade.MergeMQCascade(
		cascade.MQKropathSection{},
		cascade.MQKropathSection{},
		cascade.MQConfigSection{},
		cascade.MQConfigSection{},
		cascade.MQConfigSection{DeploymentMode: "SINGLE_INSTANCE"}, // level 6
		cascade.MQConfigSection{},
		cascade.MQKropathSection{},
		cascade.MQKropathSection{},
	)
	if result.Mandatory.DeploymentMode != "" {
		t.Errorf("CC-6: expected mandatory.deploymentMode to be empty (only defaults set), got %q", result.Mandatory.DeploymentMode)
	}
	if result.Defaults.DeploymentMode != "SINGLE_INSTANCE" {
		t.Errorf("CC-6: expected defaults.deploymentMode=SINGLE_INSTANCE from level 6, got %q", result.Defaults.DeploymentMode)
	}
}

func TestMergeMQCascade_CC7_Level6WinsOverLevel7Defaults(t *testing.T) {
	// CC-7: localMQCfgDefaults.logsGeneral=true (level 6) wins over
	// globalMQCfgDefaults.logsGeneral=false (level 7).
	result := cascade.MergeMQCascade(
		cascade.MQKropathSection{},
		cascade.MQKropathSection{},
		cascade.MQConfigSection{},
		cascade.MQConfigSection{},
		cascade.MQConfigSection{LogsGeneral: mqBoolPtr(true)},  // level 6 wins
		cascade.MQConfigSection{LogsGeneral: mqBoolPtr(false)}, // level 7 loses
		cascade.MQKropathSection{},
		cascade.MQKropathSection{},
	)
	if result.Defaults.LogsGeneral == nil || *result.Defaults.LogsGeneral != true {
		t.Errorf("CC-7: expected defaults.logsGeneral=true (level 6 wins), got %v", result.Defaults.LogsGeneral)
	}
}

func TestMergeMQCascade_CC8_Level7DefaultsWhenLevel6Empty(t *testing.T) {
	// CC-8: globalMQCfgDefaults.namingTemplate="{namespace}-{name}" (level 7) propagates
	// when localMQCfgDefaults has no namingTemplate. NamingTemplate is MQConfig-only.
	result := cascade.MergeMQCascade(
		cascade.MQKropathSection{},
		cascade.MQKropathSection{},
		cascade.MQConfigSection{},
		cascade.MQConfigSection{},
		cascade.MQConfigSection{},                                  // level 6 empty
		cascade.MQConfigSection{NamingTemplate: "{namespace}-{name}"}, // level 7
		cascade.MQKropathSection{},
		cascade.MQKropathSection{},
	)
	if result.Defaults.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("CC-8: expected defaults.namingTemplate={namespace}-{name} from level 7, got %q", result.Defaults.NamingTemplate)
	}
}

func TestMergeMQCascade_CC9_KropathDefaultsLevel8Level9(t *testing.T) {
	// CC-9: globalKropathDefaults.mq.encryptionUseAWSOwnedKey=false (level 9) propagates
	// when no MQConfig levels are set. localKropathDefaults wins over global.
	result := cascade.MergeMQCascade(
		cascade.MQKropathSection{},
		cascade.MQKropathSection{},
		cascade.MQConfigSection{},
		cascade.MQConfigSection{},
		cascade.MQConfigSection{},
		cascade.MQConfigSection{},
		cascade.MQKropathSection{},                                            // level 8 empty
		cascade.MQKropathSection{EncryptionUseAWSOwnedKey: mqBoolPtr(false)},   // level 9
	)
	if result.Defaults.EncryptionUseAWSOwnedKey == nil || *result.Defaults.EncryptionUseAWSOwnedKey != false {
		t.Errorf("CC-9: expected defaults.encryptionUseAWSOwnedKey=false from level 9, got %v", result.Defaults.EncryptionUseAWSOwnedKey)
	}
}

func TestMergeMQCascade_CC10_Level8WinsOverLevel9Defaults(t *testing.T) {
	// CC-10: localKropathDefaults.mq.autoMinorVersionUpgrade=true (level 8) wins over
	// globalKropathDefaults.mq.autoMinorVersionUpgrade=false (level 9).
	result := cascade.MergeMQCascade(
		cascade.MQKropathSection{},
		cascade.MQKropathSection{},
		cascade.MQConfigSection{},
		cascade.MQConfigSection{},
		cascade.MQConfigSection{},
		cascade.MQConfigSection{},
		cascade.MQKropathSection{AutoMinorVersionUpgrade: mqBoolPtr(true)},  // level 8 wins
		cascade.MQKropathSection{AutoMinorVersionUpgrade: mqBoolPtr(false)}, // level 9 loses
	)
	if result.Defaults.AutoMinorVersionUpgrade == nil || *result.Defaults.AutoMinorVersionUpgrade != true {
		t.Errorf("CC-10: expected defaults.autoMinorVersionUpgrade=true (level 8 wins), got %v", result.Defaults.AutoMinorVersionUpgrade)
	}
}

func TestMergeMQCascade_CC11_TagUnionMerge(t *testing.T) {
	// CC-11: Tags are union-merged across all mandatory levels.
	// Level 1 wins on key conflict; no level 2/3/4 keys are silently dropped.
	result := cascade.MergeMQCascade(
		cascade.MQKropathSection{Tags: map[string]string{"env": "prod", "owner": "platform"}},   // level 1
		cascade.MQKropathSection{Tags: map[string]string{"env": "staging", "team": "infra"}},   // level 2 (env overridden by L1)
		cascade.MQConfigSection{Tags: map[string]string{"cost-center": "cc-001"}},               // level 3
		cascade.MQConfigSection{Tags: map[string]string{"cost-center": "cc-999", "app": "mq"}}, // level 4 (cost-center overridden by L3)
		cascade.MQConfigSection{},
		cascade.MQConfigSection{},
		cascade.MQKropathSection{},
		cascade.MQKropathSection{},
	)
	tests := []struct {
		key  string
		want string
	}{
		{"env", "prod"},        // L1 wins over L2
		{"owner", "platform"},  // L1 only
		{"team", "infra"},      // L2 only
		{"cost-center", "cc-001"}, // L3 wins over L4
		{"app", "mq"},          // L4 only
	}
	for _, tc := range tests {
		if got := result.Mandatory.Tags[tc.key]; got != tc.want {
			t.Errorf("CC-11: mandatory.tags[%q]=%q, want %q", tc.key, got, tc.want)
		}
	}
}

func TestMergeMQCascade_CC12_SyncedLabelsAndAnnotationsFromMQConfigOnly(t *testing.T) {
	// CC-12: SyncedLabels and SyncedAnnotations are additive-merged from MQConfig levels only.
	// KropathConfig.mq has no syncedLabels/syncedAnnotations fields.
	// Level 3 wins over level 4 for mandatory; level 6 wins over level 7 for defaults.
	result := cascade.MergeMQCascade(
		cascade.MQKropathSection{},
		cascade.MQKropathSection{},
		cascade.MQConfigSection{ // level 3 — mandatory global MQConfig
			SyncedLabels:      map[string]string{"app.kubernetes.io/managed-by": "kropath"},
			SyncedAnnotations: map[string]string{"cost-center": "cc-global"},
		},
		cascade.MQConfigSection{ // level 4 — mandatory local MQConfig
			SyncedLabels:      map[string]string{"app.kubernetes.io/managed-by": "local", "tier": "backend"},
			SyncedAnnotations: map[string]string{"cost-center": "cc-local", "app": "broker"},
		},
		cascade.MQConfigSection{ // level 6 — defaults local MQConfig
			SyncedLabels:      map[string]string{"environment": "prod"},
			SyncedAnnotations: map[string]string{"team": "platform"},
		},
		cascade.MQConfigSection{ // level 7 — defaults global MQConfig
			SyncedLabels:      map[string]string{"environment": "staging"},
			SyncedAnnotations: map[string]string{"team": "infra", "managed-by": "ops"},
		},
		cascade.MQKropathSection{},
		cascade.MQKropathSection{},
	)

	// Mandatory SyncedLabels: L3 wins on key conflict, L4 fills unique keys
	mandatoryLabelTests := []struct{ key, want string }{
		{"app.kubernetes.io/managed-by", "kropath"}, // L3 wins over L4
		{"tier", "backend"},                          // L4 only
	}
	for _, tc := range mandatoryLabelTests {
		if got := result.Mandatory.SyncedLabels[tc.key]; got != tc.want {
			t.Errorf("CC-12: mandatory.syncedLabels[%q]=%q, want %q", tc.key, got, tc.want)
		}
	}

	// Mandatory SyncedAnnotations: L3 wins on key conflict, L4 fills unique keys
	if got := result.Mandatory.SyncedAnnotations["cost-center"]; got != "cc-global" {
		t.Errorf("CC-12: mandatory.syncedAnnotations[cost-center]=%q, want cc-global", got)
	}
	if got := result.Mandatory.SyncedAnnotations["app"]; got != "broker" {
		t.Errorf("CC-12: mandatory.syncedAnnotations[app]=%q, want broker", got)
	}

	// Defaults SyncedLabels: L6 wins over L7 on key conflict
	if got := result.Defaults.SyncedLabels["environment"]; got != "prod" {
		t.Errorf("CC-12: defaults.syncedLabels[environment]=%q, want prod (L6 wins)", got)
	}

	// Defaults SyncedAnnotations: L6 wins over L7 on key conflict, L7 fills unique keys
	if got := result.Defaults.SyncedAnnotations["team"]; got != "platform" {
		t.Errorf("CC-12: defaults.syncedAnnotations[team]=%q, want platform (L6 wins)", got)
	}
	if got := result.Defaults.SyncedAnnotations["managed-by"]; got != "ops" {
		t.Errorf("CC-12: defaults.syncedAnnotations[managed-by]=%q, want ops (L7 only)", got)
	}
}

func TestMergeMQCascade_NilPassthrough(t *testing.T) {
	// All inputs empty — all outputs should be nil/*bool or empty string.
	result := cascade.MergeMQCascade(
		cascade.MQKropathSection{},
		cascade.MQKropathSection{},
		cascade.MQConfigSection{},
		cascade.MQConfigSection{},
		cascade.MQConfigSection{},
		cascade.MQConfigSection{},
		cascade.MQKropathSection{},
		cascade.MQKropathSection{},
	)
	if result.Mandatory.PubliclyAccessible != nil {
		t.Error("empty cascade should produce nil mandatory.publiclyAccessible")
	}
	if result.Defaults.EncryptionUseAWSOwnedKey != nil {
		t.Error("empty cascade should produce nil defaults.encryptionUseAWSOwnedKey")
	}
	if result.Mandatory.EngineType != "" {
		t.Errorf("empty cascade should produce empty mandatory.engineType, got %q", result.Mandatory.EngineType)
	}
}
