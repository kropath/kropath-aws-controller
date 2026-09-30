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

// zeroKeyspacesKropath is the zero-value used when a KropathConfig tier is absent.
var zeroKeyspacesKropath = cascade.KeyspacesKropathSection{}

// zeroKeyspacesCfg is the zero-value used when a KeyspacesConfig tier is absent.
var zeroKeyspacesCfg = cascade.KeyspacesConfigSection{}

func TestMergeKeyspacesCascade_MandatoryOrder(t *testing.T) {
	tests := []struct {
		name                        string
		globalKropathMandatory      cascade.KeyspacesKropathSection
		localKropathMandatory       cascade.KeyspacesKropathSection
		globalKeyspacesCfgMandatory cascade.KeyspacesConfigSection
		localKeyspacesCfgMandatory  cascade.KeyspacesConfigSection
		wantReplicationStrategy     string
		wantThroughputMode          string
		wantEncryptionType          string
		wantPointInTimeRecovery     bool
		wantTTLEnabled              bool
	}{
		{
			name: "global KropathConfig mandatory wins (L1 strongest)",
			globalKropathMandatory: cascade.KeyspacesKropathSection{
				ReplicationStrategy: "SINGLE_REGION",
				ThroughputMode:      "PAY_PER_REQUEST",
				EncryptionType:      "AWS_OWNED_KMS_KEY",
				PointInTimeRecovery: true,
				TTLEnabled:          true,
			},
			localKropathMandatory: cascade.KeyspacesKropathSection{
				ReplicationStrategy: "MULTI_REGION",
				ThroughputMode:      "PROVISIONED",
			},
			globalKeyspacesCfgMandatory: cascade.KeyspacesConfigSection{
				ReplicationStrategy: "SINGLE_REGION",
			},
			localKeyspacesCfgMandatory: cascade.KeyspacesConfigSection{
				ReplicationStrategy: "MULTI_REGION",
			},
			wantReplicationStrategy: "SINGLE_REGION",
			wantThroughputMode:      "PAY_PER_REQUEST",
			wantEncryptionType:      "AWS_OWNED_KMS_KEY",
			wantPointInTimeRecovery: true,
			wantTTLEnabled:          true,
		},
		{
			name:                   "local KropathConfig mandatory wins over KeyspacesConfig (L2 > L3)",
			globalKropathMandatory: zeroKeyspacesKropath,
			localKropathMandatory: cascade.KeyspacesKropathSection{
				ReplicationStrategy: "SINGLE_REGION",
				ThroughputMode:      "PAY_PER_REQUEST",
			},
			globalKeyspacesCfgMandatory: cascade.KeyspacesConfigSection{
				ReplicationStrategy: "MULTI_REGION",
			},
			localKeyspacesCfgMandatory: cascade.KeyspacesConfigSection{
				ReplicationStrategy: "MULTI_REGION",
			},
			wantReplicationStrategy: "SINGLE_REGION",
			wantThroughputMode:      "PAY_PER_REQUEST",
		},
		{
			name:                   "global KeyspacesConfig mandatory wins over local (L3 > L4)",
			globalKropathMandatory: zeroKeyspacesKropath,
			localKropathMandatory:  zeroKeyspacesKropath,
			globalKeyspacesCfgMandatory: cascade.KeyspacesConfigSection{
				EncryptionType: "CUSTOMER_MANAGED_KMS_KEY",
			},
			localKeyspacesCfgMandatory: cascade.KeyspacesConfigSection{
				EncryptionType: "AWS_OWNED_KMS_KEY",
			},
			wantEncryptionType: "CUSTOMER_MANAGED_KMS_KEY",
		},
		{
			name:                        "local KeyspacesConfig mandatory wins when all above absent (L4 weakest mandatory)",
			globalKropathMandatory:      zeroKeyspacesKropath,
			localKropathMandatory:       zeroKeyspacesKropath,
			globalKeyspacesCfgMandatory: zeroKeyspacesCfg,
			localKeyspacesCfgMandatory: cascade.KeyspacesConfigSection{
				ReplicationStrategy: "MULTI_REGION",
				PointInTimeRecovery: true,
			},
			wantReplicationStrategy: "MULTI_REGION",
			wantPointInTimeRecovery: true,
		},
		{
			name:                        "booleans use OR semantics — true at any level overrides false",
			globalKropathMandatory:      cascade.KeyspacesKropathSection{TTLEnabled: true},
			localKropathMandatory:       zeroKeyspacesKropath,
			globalKeyspacesCfgMandatory: zeroKeyspacesCfg,
			localKeyspacesCfgMandatory:  zeroKeyspacesCfg,
			wantTTLEnabled:              true,
		},
		{
			name:                        "all absent returns zero values",
			globalKropathMandatory:      zeroKeyspacesKropath,
			localKropathMandatory:       zeroKeyspacesKropath,
			globalKeyspacesCfgMandatory: zeroKeyspacesCfg,
			localKeyspacesCfgMandatory:  zeroKeyspacesCfg,
			wantReplicationStrategy:     "",
			wantThroughputMode:          "",
			wantEncryptionType:          "",
			wantPointInTimeRecovery:     false,
			wantTTLEnabled:              false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cascade.MergeKeyspacesCascade(
				tt.globalKropathMandatory,
				tt.localKropathMandatory,
				tt.globalKeyspacesCfgMandatory,
				tt.localKeyspacesCfgMandatory,
				zeroKeyspacesCfg,
				zeroKeyspacesCfg,
				zeroKeyspacesKropath,
				zeroKeyspacesKropath,
			)
			if got.Mandatory.ReplicationStrategy != tt.wantReplicationStrategy {
				t.Errorf("Mandatory.ReplicationStrategy = %q, want %q", got.Mandatory.ReplicationStrategy, tt.wantReplicationStrategy)
			}
			if got.Mandatory.ThroughputMode != tt.wantThroughputMode {
				t.Errorf("Mandatory.ThroughputMode = %q, want %q", got.Mandatory.ThroughputMode, tt.wantThroughputMode)
			}
			if got.Mandatory.EncryptionType != tt.wantEncryptionType {
				t.Errorf("Mandatory.EncryptionType = %q, want %q", got.Mandatory.EncryptionType, tt.wantEncryptionType)
			}
			if got.Mandatory.PointInTimeRecovery != tt.wantPointInTimeRecovery {
				t.Errorf("Mandatory.PointInTimeRecovery = %v, want %v", got.Mandatory.PointInTimeRecovery, tt.wantPointInTimeRecovery)
			}
			if got.Mandatory.TTLEnabled != tt.wantTTLEnabled {
				t.Errorf("Mandatory.TTLEnabled = %v, want %v", got.Mandatory.TTLEnabled, tt.wantTTLEnabled)
			}
		})
	}
}

func TestMergeKeyspacesCascade_DefaultsOrder(t *testing.T) {
	tests := []struct {
		name                       string
		localKeyspacesCfgDefaults  cascade.KeyspacesConfigSection
		globalKeyspacesCfgDefaults cascade.KeyspacesConfigSection
		localKropathDefaults       cascade.KeyspacesKropathSection
		globalKropathDefaults      cascade.KeyspacesKropathSection
		wantReplicationStrategy    string
		wantThroughputMode         string
		wantEncryptionType         string
		wantPointInTimeRecovery    bool
		wantTTLEnabled             bool
	}{
		{
			name: "local KeyspacesConfig defaults strongest (L6)",
			localKeyspacesCfgDefaults: cascade.KeyspacesConfigSection{
				ReplicationStrategy: "SINGLE_REGION",
				ThroughputMode:      "PAY_PER_REQUEST",
			},
			globalKeyspacesCfgDefaults: cascade.KeyspacesConfigSection{
				ReplicationStrategy: "MULTI_REGION",
			},
			localKropathDefaults: cascade.KeyspacesKropathSection{
				ReplicationStrategy: "MULTI_REGION",
			},
			globalKropathDefaults: cascade.KeyspacesKropathSection{
				ReplicationStrategy: "MULTI_REGION",
			},
			wantReplicationStrategy: "SINGLE_REGION",
			wantThroughputMode:      "PAY_PER_REQUEST",
		},
		{
			name:                      "global KeyspacesConfig defaults wins over KropathConfig (L7 > L8)",
			localKeyspacesCfgDefaults: zeroKeyspacesCfg,
			globalKeyspacesCfgDefaults: cascade.KeyspacesConfigSection{
				EncryptionType: "CUSTOMER_MANAGED_KMS_KEY",
			},
			localKropathDefaults: cascade.KeyspacesKropathSection{
				EncryptionType: "AWS_OWNED_KMS_KEY",
			},
			globalKropathDefaults: cascade.KeyspacesKropathSection{
				EncryptionType: "AWS_OWNED_KMS_KEY",
			},
			wantEncryptionType: "CUSTOMER_MANAGED_KMS_KEY",
		},
		{
			name:                       "local KropathConfig defaults wins over global (L8 > L9)",
			localKeyspacesCfgDefaults:  zeroKeyspacesCfg,
			globalKeyspacesCfgDefaults: zeroKeyspacesCfg,
			localKropathDefaults: cascade.KeyspacesKropathSection{
				ThroughputMode:      "PAY_PER_REQUEST",
				PointInTimeRecovery: true,
			},
			globalKropathDefaults: cascade.KeyspacesKropathSection{
				ThroughputMode: "PROVISIONED",
			},
			wantThroughputMode:      "PAY_PER_REQUEST",
			wantPointInTimeRecovery: true,
		},
		{
			name:                       "global KropathConfig defaults weakest (L9)",
			localKeyspacesCfgDefaults:  zeroKeyspacesCfg,
			globalKeyspacesCfgDefaults: zeroKeyspacesCfg,
			localKropathDefaults:       zeroKeyspacesKropath,
			globalKropathDefaults: cascade.KeyspacesKropathSection{
				ReplicationStrategy: "SINGLE_REGION",
				TTLEnabled:          true,
			},
			wantReplicationStrategy: "SINGLE_REGION",
			wantTTLEnabled:          true,
		},
		{
			name:                       "all absent returns zero values",
			localKeyspacesCfgDefaults:  zeroKeyspacesCfg,
			globalKeyspacesCfgDefaults: zeroKeyspacesCfg,
			localKropathDefaults:       zeroKeyspacesKropath,
			globalKropathDefaults:      zeroKeyspacesKropath,
			wantReplicationStrategy:    "",
			wantThroughputMode:         "",
			wantEncryptionType:         "",
			wantPointInTimeRecovery:    false,
			wantTTLEnabled:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cascade.MergeKeyspacesCascade(
				zeroKeyspacesKropath,
				zeroKeyspacesKropath,
				zeroKeyspacesCfg,
				zeroKeyspacesCfg,
				tt.localKeyspacesCfgDefaults,
				tt.globalKeyspacesCfgDefaults,
				tt.localKropathDefaults,
				tt.globalKropathDefaults,
			)
			if got.Defaults.ReplicationStrategy != tt.wantReplicationStrategy {
				t.Errorf("Defaults.ReplicationStrategy = %q, want %q", got.Defaults.ReplicationStrategy, tt.wantReplicationStrategy)
			}
			if got.Defaults.ThroughputMode != tt.wantThroughputMode {
				t.Errorf("Defaults.ThroughputMode = %q, want %q", got.Defaults.ThroughputMode, tt.wantThroughputMode)
			}
			if got.Defaults.EncryptionType != tt.wantEncryptionType {
				t.Errorf("Defaults.EncryptionType = %q, want %q", got.Defaults.EncryptionType, tt.wantEncryptionType)
			}
			if got.Defaults.PointInTimeRecovery != tt.wantPointInTimeRecovery {
				t.Errorf("Defaults.PointInTimeRecovery = %v, want %v", got.Defaults.PointInTimeRecovery, tt.wantPointInTimeRecovery)
			}
			if got.Defaults.TTLEnabled != tt.wantTTLEnabled {
				t.Errorf("Defaults.TTLEnabled = %v, want %v", got.Defaults.TTLEnabled, tt.wantTTLEnabled)
			}
		})
	}
}

func TestMergeKeyspacesCascade_TagsMerge(t *testing.T) {
	got := cascade.MergeKeyspacesCascade(
		cascade.KeyspacesKropathSection{Tags: map[string]string{"env": "prod", "owner": "platform"}},
		cascade.KeyspacesKropathSection{Tags: map[string]string{"env": "staging"}},
		cascade.KeyspacesConfigSection{Tags: map[string]string{"team": "data"}},
		cascade.KeyspacesConfigSection{Tags: map[string]string{"team": "payments"}},
		cascade.KeyspacesConfigSection{Tags: map[string]string{"cost-center": "cc-001"}},
		cascade.KeyspacesConfigSection{Tags: map[string]string{"cost-center": "cc-002"}},
		cascade.KeyspacesKropathSection{Tags: map[string]string{"app": "myapp"}},
		cascade.KeyspacesKropathSection{Tags: map[string]string{"app": "global"}},
	)

	// Mandatory: lower level wins on conflict; L1 wins env, L1 wins owner, L3 wins team
	if got.Mandatory.Tags["env"] != "prod" {
		t.Errorf("Mandatory Tags[env] = %q, want %q", got.Mandatory.Tags["env"], "prod")
	}
	if got.Mandatory.Tags["owner"] != "platform" {
		t.Errorf("Mandatory Tags[owner] = %q, want %q", got.Mandatory.Tags["owner"], "platform")
	}
	if got.Mandatory.Tags["team"] != "data" {
		t.Errorf("Mandatory Tags[team] = %q, want %q", got.Mandatory.Tags["team"], "data")
	}

	// Defaults: L6 wins on conflict
	if got.Defaults.Tags["cost-center"] != "cc-001" {
		t.Errorf("Defaults Tags[cost-center] = %q, want %q", got.Defaults.Tags["cost-center"], "cc-001")
	}
	if got.Defaults.Tags["app"] != "myapp" {
		t.Errorf("Defaults Tags[app] = %q, want %q", got.Defaults.Tags["app"], "myapp")
	}
}

func TestMergeKeyspacesCascade_NamingTemplate(t *testing.T) {
	// NamingTemplate is KeyspacesConfig-only — KropathConfig does not have this field.
	got := cascade.MergeKeyspacesCascade(
		zeroKeyspacesKropath,
		zeroKeyspacesKropath,
		cascade.KeyspacesConfigSection{NamingTemplate: "global-ks-{{.Name}}"},
		cascade.KeyspacesConfigSection{NamingTemplate: "local-ks-{{.Name}}"},
		cascade.KeyspacesConfigSection{NamingTemplate: "default-ks-{{.Name}}"},
		cascade.KeyspacesConfigSection{NamingTemplate: "global-default-ks-{{.Name}}"},
		zeroKeyspacesKropath,
		zeroKeyspacesKropath,
	)

	// Mandatory: global KsCfg (L3) wins over local KsCfg (L4)
	if got.Mandatory.NamingTemplate != "global-ks-{{.Name}}" {
		t.Errorf("Mandatory.NamingTemplate = %q, want %q", got.Mandatory.NamingTemplate, "global-ks-{{.Name}}")
	}
	// Defaults: local KsCfg (L6) wins over global KsCfg (L7)
	if got.Defaults.NamingTemplate != "default-ks-{{.Name}}" {
		t.Errorf("Defaults.NamingTemplate = %q, want %q", got.Defaults.NamingTemplate, "default-ks-{{.Name}}")
	}
}

func TestMergeKeyspacesCascade_MandatoryDefaultsIsolation(t *testing.T) {
	// Values set in defaults must not bleed into mandatory tier and vice versa.
	got := cascade.MergeKeyspacesCascade(
		zeroKeyspacesKropath,
		zeroKeyspacesKropath,
		zeroKeyspacesCfg,
		zeroKeyspacesCfg,
		cascade.KeyspacesConfigSection{
			ReplicationStrategy: "SINGLE_REGION",
			PointInTimeRecovery: true,
		},
		zeroKeyspacesCfg,
		zeroKeyspacesKropath,
		zeroKeyspacesKropath,
	)

	if got.Mandatory.ReplicationStrategy != "" {
		t.Errorf("Mandatory.ReplicationStrategy should be empty when only set in defaults, got %q", got.Mandatory.ReplicationStrategy)
	}
	if got.Mandatory.PointInTimeRecovery {
		t.Error("Mandatory.PointInTimeRecovery should be false when only set in defaults")
	}
	if got.Defaults.ReplicationStrategy != "SINGLE_REGION" {
		t.Errorf("Defaults.ReplicationStrategy = %q, want %q", got.Defaults.ReplicationStrategy, "SINGLE_REGION")
	}
	if !got.Defaults.PointInTimeRecovery {
		t.Error("Defaults.PointInTimeRecovery should be true")
	}
}
