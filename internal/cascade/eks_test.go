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

func boolPtrEKS(b bool) *bool { return &b }

func TestMergeEKSCascade(t *testing.T) {
	t.Parallel()

	type inputs struct {
		globalKropathMandatory cascade.EKSKropathSection
		localKropathMandatory  cascade.EKSKropathSection
		globalEKSCfgMandatory  cascade.EKSConfigSection
		localEKSCfgMandatory   cascade.EKSConfigSection
		localEKSCfgDefaults    cascade.EKSConfigSection
		globalEKSCfgDefaults   cascade.EKSConfigSection
		localKropathDefaults   cascade.EKSKropathSection
		globalKropathDefaults  cascade.EKSKropathSection
	}

	tests := []struct {
		name string
		in   inputs
		want cascade.EffectiveEKSConfig
	}{
		// AC-1: globalKropathConfig.mandatory.eks.version="1.31" propagates to
		// effCfg.mandatory.version="1.31" (level 1).
		{
			name: "ac1-global-kpc-version-propagates",
			in: inputs{
				globalKropathMandatory: cascade.EKSKropathSection{Version: "1.31"},
			},
			want: cascade.EffectiveEKSConfig{
				Mandatory: cascade.EffectiveEKSSection{Version: "1.31"},
			},
		},
		// AC-2: globalEKSConfig.mandatory.version="1.30" propagates when
		// KropathConfig.mandatory.eks.version is empty.
		{
			name: "ac2-global-ekscfg-mandatory-version",
			in: inputs{
				globalEKSCfgMandatory: cascade.EKSConfigSection{Version: "1.30"},
			},
			want: cascade.EffectiveEKSConfig{
				Mandatory: cascade.EffectiveEKSSection{Version: "1.30"},
			},
		},
		// AC-3: all mandatory sources empty; localEKSConfig.defaults.version="1.29"
		// propagates to effCfg.defaults.version (level 6).
		{
			name: "ac3-defaults-version-from-local-ekscfg",
			in: inputs{
				localEKSCfgDefaults: cascade.EKSConfigSection{Version: "1.29"},
			},
			want: cascade.EffectiveEKSConfig{
				Defaults: cascade.EffectiveEKSSection{Version: "1.29"},
			},
		},
		// AC-4: globalKropathConfig.mandatory.eks.authenticationMode="API" propagates.
		{
			name: "ac4-global-kpc-authmode-propagates",
			in: inputs{
				globalKropathMandatory: cascade.EKSKropathSection{AuthenticationMode: "API"},
			},
			want: cascade.EffectiveEKSConfig{
				Mandatory: cascade.EffectiveEKSSection{AuthenticationMode: "API"},
			},
		},
		// AC-5: globalEKSConfig.defaults.authenticationMode="API" (level 7) wins over
		// globalKropathConfig.defaults.eks.authenticationMode="API_AND_CONFIG_MAP" (level 9).
		{
			name: "ac5-ekscfg-defaults-authmode-wins-over-kpc-defaults",
			in: inputs{
				globalEKSCfgDefaults:  cascade.EKSConfigSection{AuthenticationMode: "API"},
				globalKropathDefaults: cascade.EKSKropathSection{AuthenticationMode: "API_AND_CONFIG_MAP"},
			},
			want: cascade.EffectiveEKSConfig{
				Defaults: cascade.EffectiveEKSSection{AuthenticationMode: "API"},
			},
		},
		// AC-6: globalKropathConfig.mandatory.eks.loggingTypes=["api","audit"] propagates.
		{
			name: "ac6-global-kpc-loggingtypes-propagates",
			in: inputs{
				globalKropathMandatory: cascade.EKSKropathSection{LoggingTypes: []string{"api", "audit"}},
			},
			want: cascade.EffectiveEKSConfig{
				Mandatory: cascade.EffectiveEKSSection{LoggingTypes: []string{"api", "audit"}},
			},
		},
		// AC-7: globalEKSConfig.mandatory.encryptionKeyArn propagates (levels 3-4 only).
		{
			name: "ac7-global-ekscfg-encryptionkeyarn-propagates",
			in: inputs{
				globalEKSCfgMandatory: cascade.EKSConfigSection{
					EncryptionKeyArn: "arn:aws:kms:us-east-1:123456789012:key/test-key",
				},
			},
			want: cascade.EffectiveEKSConfig{
				Mandatory: cascade.EffectiveEKSSection{
					EncryptionKeyArn: "arn:aws:kms:us-east-1:123456789012:key/test-key",
				},
			},
		},
		// AC-8: globalEKSConfig.defaults.endpointPublicAccess=true propagates.
		{
			name: "ac8-global-ekscfg-defaults-endpoint-public",
			in: inputs{
				globalEKSCfgDefaults: cascade.EKSConfigSection{
					EndpointPublicAccess: boolPtrEKS(true),
				},
			},
			want: cascade.EffectiveEKSConfig{
				Defaults: cascade.EffectiveEKSSection{
					EndpointPublicAccess: boolPtrEKS(true),
				},
			},
		},
		// AC-9: globalEKSConfig.defaults.endpointPrivateAccess=true propagates.
		{
			name: "ac9-global-ekscfg-defaults-endpoint-private",
			in: inputs{
				globalEKSCfgDefaults: cascade.EKSConfigSection{
					EndpointPrivateAccess: boolPtrEKS(true),
				},
			},
			want: cascade.EffectiveEKSConfig{
				Defaults: cascade.EffectiveEKSSection{
					EndpointPrivateAccess: boolPtrEKS(true),
				},
			},
		},
		// AC-10: globalEKSConfig.defaults.supportType="STANDARD" propagates.
		{
			name: "ac10-global-ekscfg-defaults-supporttype",
			in: inputs{
				globalEKSCfgDefaults: cascade.EKSConfigSection{SupportType: "STANDARD"},
			},
			want: cascade.EffectiveEKSConfig{
				Defaults: cascade.EffectiveEKSSection{SupportType: "STANDARD"},
			},
		},
		// AC-11: globalEKSConfig.defaults.namingTemplate="{namespace}-{name}" propagates.
		{
			name: "ac11-global-ekscfg-defaults-namingtemplate",
			in: inputs{
				globalEKSCfgDefaults: cascade.EKSConfigSection{NamingTemplate: "{namespace}-{name}"},
			},
			want: cascade.EffectiveEKSConfig{
				Defaults: cascade.EffectiveEKSSection{NamingTemplate: "{namespace}-{name}"},
			},
		},
		// AC-12: KropathConfig.mandatory.tags and EKSConfig.mandatory.tags are
		// union-merged into effCfg.mandatory.tags.
		{
			name: "ac12-mandatory-tags-union-merge",
			in: inputs{
				globalKropathMandatory: cascade.EKSKropathSection{
					Tags: map[string]string{"cost-centre": "platform"},
				},
				globalEKSCfgMandatory: cascade.EKSConfigSection{
					Tags: map[string]string{"service": "eks"},
				},
			},
			want: cascade.EffectiveEKSConfig{
				Mandatory: cascade.EffectiveEKSSection{
					Tags: map[string]string{
						"cost-centre": "platform",
						"service":     "eks",
					},
				},
			},
		},
		// AC-13: EKSConfig.mandatory.syncedLabels propagates to effCfg.mandatory.syncedLabels.
		{
			name: "ac13-ekscfg-mandatory-synced-labels",
			in: inputs{
				localEKSCfgMandatory: cascade.EKSConfigSection{
					SyncedLabels: map[string]string{"team": "platform"},
				},
			},
			want: cascade.EffectiveEKSConfig{
				Mandatory: cascade.EffectiveEKSSection{
					SyncedLabels: map[string]string{"team": "platform"},
				},
			},
		},
		// AC-14: level-1 KropathConfig.mandatory.eks.version wins over
		// level-3 globalEKSConfig.mandatory.version.
		{
			name: "ac14-level1-kpc-version-wins-over-level3-ekscfg",
			in: inputs{
				globalKropathMandatory: cascade.EKSKropathSection{Version: "1.31"},
				globalEKSCfgMandatory:  cascade.EKSConfigSection{Version: "1.30"},
			},
			want: cascade.EffectiveEKSConfig{
				Mandatory: cascade.EffectiveEKSSection{Version: "1.31"},
			},
		},
		// AC-15: loggingTypes treated as scalar (first-non-empty wins);
		// level-1 loggingTypes=["api"] wins over level-3 loggingTypes=["audit","scheduler"].
		{
			name: "ac15-loggingtypes-scalar-first-wins",
			in: inputs{
				globalKropathMandatory: cascade.EKSKropathSection{LoggingTypes: []string{"api"}},
				globalEKSCfgMandatory:  cascade.EKSConfigSection{LoggingTypes: []string{"audit", "scheduler"}},
			},
			want: cascade.EffectiveEKSConfig{
				Mandatory: cascade.EffectiveEKSSection{LoggingTypes: []string{"api"}},
			},
		},
		// AC-16: when all mandatory sources are empty and globalKropathConfig.defaults.eks.version
		// is set, it propagates to effCfg.defaults.version (level 9, lowest defaults priority).
		{
			name: "ac16-global-kpc-defaults-version-propagates",
			in: inputs{
				globalKropathDefaults: cascade.EKSKropathSection{Version: "1.28"},
			},
			want: cascade.EffectiveEKSConfig{
				Defaults: cascade.EffectiveEKSSection{Version: "1.28"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := cascade.MergeEKSCascade(
				tt.in.globalKropathMandatory,
				tt.in.localKropathMandatory,
				tt.in.globalEKSCfgMandatory,
				tt.in.localEKSCfgMandatory,
				tt.in.localEKSCfgDefaults,
				tt.in.globalEKSCfgDefaults,
				tt.in.localKropathDefaults,
				tt.in.globalKropathDefaults,
			)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeEKSCascade() mismatch\ngot:  %+v\nwant: %+v", got, tt.want)
			}
		})
	}
}

// TestMergeEKSCascade_TagPriority verifies that lower-level-number tags win on key conflicts.
func TestMergeEKSCascade_TagPriority(t *testing.T) {
	t.Parallel()

	globalKropathMandatory := cascade.EKSKropathSection{Tags: map[string]string{"env": "prod"}}
	globalEKSCfgMandatory := cascade.EKSConfigSection{Tags: map[string]string{"env": "dev", "service": "eks"}}

	got := cascade.MergeEKSCascade(
		globalKropathMandatory,
		cascade.EKSKropathSection{},
		globalEKSCfgMandatory,
		cascade.EKSConfigSection{},
		cascade.EKSConfigSection{},
		cascade.EKSConfigSection{},
		cascade.EKSKropathSection{},
		cascade.EKSKropathSection{},
	)

	// Level 1 (globalKropathMandatory) wins over level 3 (globalEKSCfgMandatory) for "env".
	want := map[string]string{"env": "prod", "service": "eks"}
	if !reflect.DeepEqual(got.Mandatory.Tags, want) {
		t.Errorf("tag priority: got %v, want %v", got.Mandatory.Tags, want)
	}
}

// TestMergeEKSCascade_BoolPtrSentinel verifies nil vs non-nil *bool semantics.
func TestMergeEKSCascade_BoolPtrSentinel(t *testing.T) {
	t.Parallel()

	// false (explicit) must propagate — it is a governance signal, not a zero value.
	falseVal := boolPtrEKS(false)
	got := cascade.MergeEKSCascade(
		cascade.EKSKropathSection{},
		cascade.EKSKropathSection{},
		cascade.EKSConfigSection{EndpointPublicAccess: falseVal},
		cascade.EKSConfigSection{},
		cascade.EKSConfigSection{},
		cascade.EKSConfigSection{},
		cascade.EKSKropathSection{},
		cascade.EKSKropathSection{},
	)

	if got.Mandatory.EndpointPublicAccess == nil || *got.Mandatory.EndpointPublicAccess != false {
		t.Errorf("expected EndpointPublicAccess=false, got %v", got.Mandatory.EndpointPublicAccess)
	}
}

// TestMergeEKSCascade_LoggingTypesDefensiveCopy verifies firstNonEmptyStrings returns a copy.
func TestMergeEKSCascade_LoggingTypesDefensiveCopy(t *testing.T) {
	t.Parallel()

	src := []string{"api", "audit"}
	got := cascade.MergeEKSCascade(
		cascade.EKSKropathSection{LoggingTypes: src},
		cascade.EKSKropathSection{},
		cascade.EKSConfigSection{},
		cascade.EKSConfigSection{},
		cascade.EKSConfigSection{},
		cascade.EKSConfigSection{},
		cascade.EKSKropathSection{},
		cascade.EKSKropathSection{},
	)

	if &got.Mandatory.LoggingTypes[0] == &src[0] {
		t.Error("LoggingTypes slice is aliased — expected a defensive copy")
	}
}
