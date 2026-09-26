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

package cascade

import (
	"reflect"
	"testing"
)

func TestMergeRoute53Cascade_AllAbsent(t *testing.T) {
	got := MergeRoute53Cascade(
		Route53KropathSection{},
		Route53KropathSection{},
		Route53ConfigSection{},
		Route53ConfigSection{},
		Route53ConfigSection{},
		Route53ConfigSection{},
		Route53KropathSection{},
		Route53KropathSection{},
	)
	want := EffectiveRoute53Config{
		Mandatory: EffectiveRoute53Section{},
		Defaults:  EffectiveRoute53Section{},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("AllAbsent: got %+v, want %+v", got, want)
	}
}

func TestMergeRoute53Cascade_MandatoryPriorityOrder(t *testing.T) {
	tests := []struct {
		name             string
		l1, l2           Route53KropathSection
		l3, l4           Route53ConfigSection
		wantTTL          int64
		wantInterval     int64
		wantThreshold    int64
		wantEndpointType string
	}{
		{
			name:    "level1-wins",
			l1:      Route53KropathSection{DefaultTTL: 60, HealthCheckRequestInterval: 10, HealthCheckFailureThreshold: 2, ResolverEndpointType: "IPV4"},
			l2:      Route53KropathSection{DefaultTTL: 120, HealthCheckRequestInterval: 30, HealthCheckFailureThreshold: 3, ResolverEndpointType: "IPV6"},
			l3:      Route53ConfigSection{DefaultTTL: 180, HealthCheckRequestInterval: 30, HealthCheckFailureThreshold: 5, ResolverEndpointType: "DUALSTACK"},
			l4:      Route53ConfigSection{DefaultTTL: 300, HealthCheckRequestInterval: 30, HealthCheckFailureThreshold: 8, ResolverEndpointType: "IPV6"},
			wantTTL: 60, wantInterval: 10, wantThreshold: 2, wantEndpointType: "IPV4",
		},
		{
			name:    "level2-wins-when-level1-absent",
			l1:      Route53KropathSection{},
			l2:      Route53KropathSection{DefaultTTL: 120, HealthCheckRequestInterval: 30, HealthCheckFailureThreshold: 3, ResolverEndpointType: "IPV6"},
			l3:      Route53ConfigSection{DefaultTTL: 180, HealthCheckRequestInterval: 10, HealthCheckFailureThreshold: 5, ResolverEndpointType: "DUALSTACK"},
			l4:      Route53ConfigSection{DefaultTTL: 300, HealthCheckRequestInterval: 30, HealthCheckFailureThreshold: 8, ResolverEndpointType: "IPV4"},
			wantTTL: 120, wantInterval: 30, wantThreshold: 3, wantEndpointType: "IPV6",
		},
		{
			name:    "level3-wins-when-levels1-2-absent",
			l1:      Route53KropathSection{},
			l2:      Route53KropathSection{},
			l3:      Route53ConfigSection{DefaultTTL: 180, HealthCheckRequestInterval: 10, HealthCheckFailureThreshold: 5, ResolverEndpointType: "DUALSTACK"},
			l4:      Route53ConfigSection{DefaultTTL: 300, HealthCheckRequestInterval: 30, HealthCheckFailureThreshold: 8, ResolverEndpointType: "IPV6"},
			wantTTL: 180, wantInterval: 10, wantThreshold: 5, wantEndpointType: "DUALSTACK",
		},
		{
			name:    "level4-wins-when-levels1-3-absent",
			l1:      Route53KropathSection{},
			l2:      Route53KropathSection{},
			l3:      Route53ConfigSection{},
			l4:      Route53ConfigSection{DefaultTTL: 300, HealthCheckRequestInterval: 30, HealthCheckFailureThreshold: 8, ResolverEndpointType: "IPV6"},
			wantTTL: 300, wantInterval: 30, wantThreshold: 8, wantEndpointType: "IPV6",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeRoute53Cascade(
				tc.l1, tc.l2, tc.l3, tc.l4,
				Route53ConfigSection{},
				Route53ConfigSection{},
				Route53KropathSection{},
				Route53KropathSection{},
			)
			if got.Mandatory.DefaultTTL != tc.wantTTL {
				t.Errorf("DefaultTTL: got %d, want %d", got.Mandatory.DefaultTTL, tc.wantTTL)
			}
			if got.Mandatory.HealthCheckRequestInterval != tc.wantInterval {
				t.Errorf("HealthCheckRequestInterval: got %d, want %d", got.Mandatory.HealthCheckRequestInterval, tc.wantInterval)
			}
			if got.Mandatory.HealthCheckFailureThreshold != tc.wantThreshold {
				t.Errorf("HealthCheckFailureThreshold: got %d, want %d", got.Mandatory.HealthCheckFailureThreshold, tc.wantThreshold)
			}
			if got.Mandatory.ResolverEndpointType != tc.wantEndpointType {
				t.Errorf("ResolverEndpointType: got %q, want %q", got.Mandatory.ResolverEndpointType, tc.wantEndpointType)
			}
		})
	}
}

func TestMergeRoute53Cascade_DefaultsPriorityOrder(t *testing.T) {
	tests := []struct {
		name             string
		l6, l7           Route53ConfigSection
		l8, l9           Route53KropathSection
		wantTTL          int64
		wantInterval     int64
		wantThreshold    int64
		wantEndpointType string
	}{
		{
			name:    "level6-wins",
			l6:      Route53ConfigSection{DefaultTTL: 300, HealthCheckRequestInterval: 30, HealthCheckFailureThreshold: 3, ResolverEndpointType: "IPV4"},
			l7:      Route53ConfigSection{DefaultTTL: 600, HealthCheckRequestInterval: 10, HealthCheckFailureThreshold: 5, ResolverEndpointType: "IPV6"},
			l8:      Route53KropathSection{DefaultTTL: 900, HealthCheckRequestInterval: 30, HealthCheckFailureThreshold: 8, ResolverEndpointType: "DUALSTACK"},
			l9:      Route53KropathSection{DefaultTTL: 1200, HealthCheckRequestInterval: 10, HealthCheckFailureThreshold: 10, ResolverEndpointType: "IPV6"},
			wantTTL: 300, wantInterval: 30, wantThreshold: 3, wantEndpointType: "IPV4",
		},
		{
			name:    "level7-wins-when-level6-absent",
			l6:      Route53ConfigSection{},
			l7:      Route53ConfigSection{DefaultTTL: 600, HealthCheckRequestInterval: 10, HealthCheckFailureThreshold: 5, ResolverEndpointType: "IPV6"},
			l8:      Route53KropathSection{DefaultTTL: 900, HealthCheckRequestInterval: 30, HealthCheckFailureThreshold: 8, ResolverEndpointType: "DUALSTACK"},
			l9:      Route53KropathSection{DefaultTTL: 1200, HealthCheckRequestInterval: 30, HealthCheckFailureThreshold: 10, ResolverEndpointType: "IPV4"},
			wantTTL: 600, wantInterval: 10, wantThreshold: 5, wantEndpointType: "IPV6",
		},
		{
			name:    "level8-wins-when-levels6-7-absent",
			l6:      Route53ConfigSection{},
			l7:      Route53ConfigSection{},
			l8:      Route53KropathSection{DefaultTTL: 900, HealthCheckRequestInterval: 30, HealthCheckFailureThreshold: 8, ResolverEndpointType: "DUALSTACK"},
			l9:      Route53KropathSection{DefaultTTL: 1200, HealthCheckRequestInterval: 10, HealthCheckFailureThreshold: 10, ResolverEndpointType: "IPV6"},
			wantTTL: 900, wantInterval: 30, wantThreshold: 8, wantEndpointType: "DUALSTACK",
		},
		{
			name:    "level9-wins-when-levels6-8-absent",
			l6:      Route53ConfigSection{},
			l7:      Route53ConfigSection{},
			l8:      Route53KropathSection{},
			l9:      Route53KropathSection{DefaultTTL: 1200, HealthCheckRequestInterval: 10, HealthCheckFailureThreshold: 10, ResolverEndpointType: "IPV6"},
			wantTTL: 1200, wantInterval: 10, wantThreshold: 10, wantEndpointType: "IPV6",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeRoute53Cascade(
				Route53KropathSection{},
				Route53KropathSection{},
				Route53ConfigSection{},
				Route53ConfigSection{},
				tc.l6, tc.l7, tc.l8, tc.l9,
			)
			if got.Defaults.DefaultTTL != tc.wantTTL {
				t.Errorf("DefaultTTL: got %d, want %d", got.Defaults.DefaultTTL, tc.wantTTL)
			}
			if got.Defaults.HealthCheckRequestInterval != tc.wantInterval {
				t.Errorf("HealthCheckRequestInterval: got %d, want %d", got.Defaults.HealthCheckRequestInterval, tc.wantInterval)
			}
			if got.Defaults.HealthCheckFailureThreshold != tc.wantThreshold {
				t.Errorf("HealthCheckFailureThreshold: got %d, want %d", got.Defaults.HealthCheckFailureThreshold, tc.wantThreshold)
			}
			if got.Defaults.ResolverEndpointType != tc.wantEndpointType {
				t.Errorf("ResolverEndpointType: got %q, want %q", got.Defaults.ResolverEndpointType, tc.wantEndpointType)
			}
		})
	}
}

func TestMergeRoute53Cascade_TagsUnion(t *testing.T) {
	got := MergeRoute53Cascade(
		Route53KropathSection{Tags: map[string]string{"env": "prod", "org": "kropath"}},  // L1
		Route53KropathSection{Tags: map[string]string{"env": "staging"}},                 // L2 — L1 wins on "env"
		Route53ConfigSection{Tags: map[string]string{"team": "platform"}},                // L3
		Route53ConfigSection{Tags: map[string]string{"team": "sre", "app": "api"}},       // L4 — L3 wins on "team"
		Route53ConfigSection{Tags: map[string]string{"cost-center": "ops"}},              // L6
		Route53ConfigSection{Tags: map[string]string{"cost-center": "infra"}},            // L7 — L6 wins on "cost-center"
		Route53KropathSection{Tags: map[string]string{"owner": "platform"}},              // L8
		Route53KropathSection{Tags: map[string]string{"owner": "sre", "region": "apac"}}, // L9 — L8 wins on "owner"
	)

	wantMandatoryTags := map[string]string{
		"env":  "prod",     // L1 wins over L2
		"org":  "kropath",  // L1 only
		"team": "platform", // L3 wins over L4
		"app":  "api",      // L4 only
	}
	if !reflect.DeepEqual(got.Mandatory.Tags, wantMandatoryTags) {
		t.Errorf("Mandatory.Tags: got %v, want %v", got.Mandatory.Tags, wantMandatoryTags)
	}

	wantDefaultsTags := map[string]string{
		"cost-center": "ops",      // L6 wins over L7
		"owner":       "platform", // L8 wins over L9
		"region":      "apac",     // L9 only
	}
	if !reflect.DeepEqual(got.Defaults.Tags, wantDefaultsTags) {
		t.Errorf("Defaults.Tags: got %v, want %v", got.Defaults.Tags, wantDefaultsTags)
	}
}

func TestMergeRoute53Cascade_SyncedLabelsUnion(t *testing.T) {
	got := MergeRoute53Cascade(
		Route53KropathSection{},
		Route53KropathSection{},
		Route53ConfigSection{SyncedLabels: map[string]string{"tier": "global", "env": "prod"}},  // L3
		Route53ConfigSection{SyncedLabels: map[string]string{"tier": "local", "app": "api"}},    // L4 — L3 wins on "tier"
		Route53ConfigSection{SyncedLabels: map[string]string{"owner": "sre"}},                   // L6
		Route53ConfigSection{SyncedLabels: map[string]string{"owner": "platform", "cost": "a"}}, // L7 — L6 wins on "owner"
		Route53KropathSection{},
		Route53KropathSection{},
	)

	wantMandatoryLabels := map[string]string{
		"tier": "global", // L3 wins over L4
		"env":  "prod",   // L3 only
		"app":  "api",    // L4 only
	}
	if !reflect.DeepEqual(got.Mandatory.SyncedLabels, wantMandatoryLabels) {
		t.Errorf("Mandatory.SyncedLabels: got %v, want %v", got.Mandatory.SyncedLabels, wantMandatoryLabels)
	}

	wantDefaultsLabels := map[string]string{
		"owner": "sre", // L6 wins over L7
		"cost":  "a",   // L7 only
	}
	if !reflect.DeepEqual(got.Defaults.SyncedLabels, wantDefaultsLabels) {
		t.Errorf("Defaults.SyncedLabels: got %v, want %v", got.Defaults.SyncedLabels, wantDefaultsLabels)
	}
}

func TestMergeRoute53Cascade_MandatoryDoesNotPollutDefaults(t *testing.T) {
	got := MergeRoute53Cascade(
		Route53KropathSection{DefaultTTL: 60, ResolverEndpointType: "IPV4"},
		Route53KropathSection{},
		Route53ConfigSection{},
		Route53ConfigSection{},
		Route53ConfigSection{},
		Route53ConfigSection{},
		Route53KropathSection{},
		Route53KropathSection{},
	)
	if got.Defaults.DefaultTTL != 0 {
		t.Errorf("Defaults.DefaultTTL should be 0, got %d", got.Defaults.DefaultTTL)
	}
	if got.Defaults.ResolverEndpointType != "" {
		t.Errorf("Defaults.ResolverEndpointType should be empty, got %q", got.Defaults.ResolverEndpointType)
	}
}

func TestMergeRoute53Cascade_DefaultsDoNotPolluteMandatory(t *testing.T) {
	got := MergeRoute53Cascade(
		Route53KropathSection{},
		Route53KropathSection{},
		Route53ConfigSection{},
		Route53ConfigSection{},
		Route53ConfigSection{DefaultTTL: 300, ResolverEndpointType: "DUALSTACK"},
		Route53ConfigSection{},
		Route53KropathSection{},
		Route53KropathSection{},
	)
	if got.Mandatory.DefaultTTL != 0 {
		t.Errorf("Mandatory.DefaultTTL should be 0, got %d", got.Mandatory.DefaultTTL)
	}
	if got.Mandatory.ResolverEndpointType != "" {
		t.Errorf("Mandatory.ResolverEndpointType should be empty, got %q", got.Mandatory.ResolverEndpointType)
	}
}

// TestMergeRoute53Cascade_NamingTemplateMandatoryPriority — global Route53Config
// (level 3) wins over local Route53Config (level 4); KropathConfig cannot set
// namingTemplate at all (Route53Config-only field).
func TestMergeRoute53Cascade_NamingTemplateMandatoryPriority(t *testing.T) {
	got := MergeRoute53Cascade(
		Route53KropathSection{},
		Route53KropathSection{},
		Route53ConfigSection{NamingTemplate: "{namespace}-{name}-{tag.team}"}, // level 3
		Route53ConfigSection{NamingTemplate: "{name}-override"},               // level 4
		Route53ConfigSection{},
		Route53ConfigSection{},
		Route53KropathSection{},
		Route53KropathSection{},
	)
	if got.Mandatory.NamingTemplate != "{namespace}-{name}-{tag.team}" {
		t.Errorf("Mandatory.NamingTemplate = %q, want global Route53Config value", got.Mandatory.NamingTemplate)
	}
}

// TestMergeRoute53Cascade_NamingTemplateDefaultsPriority — local Route53Config
// (level 6) wins over global Route53Config (level 7).
func TestMergeRoute53Cascade_NamingTemplateDefaultsPriority(t *testing.T) {
	got := MergeRoute53Cascade(
		Route53KropathSection{},
		Route53KropathSection{},
		Route53ConfigSection{},
		Route53ConfigSection{},
		Route53ConfigSection{NamingTemplate: "{namespace}-{name}"}, // level 6
		Route53ConfigSection{NamingTemplate: "{name}-{tag.env}"},   // level 7
		Route53KropathSection{},
		Route53KropathSection{},
	)
	if got.Defaults.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("Defaults.NamingTemplate = %q, want local Route53Config value", got.Defaults.NamingTemplate)
	}
}
