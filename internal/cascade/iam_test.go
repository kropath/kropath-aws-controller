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

// zero is a convenience zero-value IAMSection used to represent "absent" sources.
var zero = cascade.IAMSection{}

// mergeAll calls MergeIAMCascade with all eight inputs.
func mergeAll(
	globalKropathMandatory,
	localKropathMandatory,
	globalIAMCfgMandatory,
	localIAMCfgMandatory,
	localIAMCfgDefaults,
	globalIAMCfgDefaults,
	localKropathDefaults,
	globalKropathDefaults cascade.IAMSection,
) cascade.EffectiveIAMConfig {
	return cascade.MergeIAMCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalIAMCfgMandatory,
		localIAMCfgMandatory,
		localIAMCfgDefaults,
		globalIAMCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeIAMCascade_AC1 — globalKropathConfig.mandatory.iam.permissionsBoundaryArn
// set at level 1 propagates to effCfg.mandatory.permissionsBoundaryArn.
func TestMergeIAMCascade_AC1(t *testing.T) {
	globalKropathMandatory := cascade.IAMSection{
		PermissionsBoundaryArn: "arn:aws:iam::123:policy/GlobalBlanket",
	}

	got := mergeAll(globalKropathMandatory, zero, zero, zero, zero, zero, zero, zero)

	if got.Mandatory.PermissionsBoundaryArn != "arn:aws:iam::123:policy/GlobalBlanket" {
		t.Errorf("AC-1: mandatory.permissionsBoundaryArn = %q, want %q",
			got.Mandatory.PermissionsBoundaryArn, "arn:aws:iam::123:policy/GlobalBlanket")
	}
	if got.Defaults.PermissionsBoundaryArn != "" {
		t.Errorf("AC-1: defaults.permissionsBoundaryArn should be empty, got %q",
			got.Defaults.PermissionsBoundaryArn)
	}
}

// TestMergeIAMCascade_AC2 — level-1 (KropathConfig) mandatory wins over
// level-3 (IAMConfig) mandatory when both are set.
func TestMergeIAMCascade_AC2(t *testing.T) {
	globalKropathMandatory := cascade.IAMSection{
		PermissionsBoundaryArn: "arn:aws:iam::123:policy/GlobalBlanket",
	}
	globalIAMCfgMandatory := cascade.IAMSection{
		PermissionsBoundaryArn: "arn:aws:iam::123:policy/IAMCfgBoundary",
	}

	got := mergeAll(globalKropathMandatory, zero, globalIAMCfgMandatory, zero, zero, zero, zero, zero)

	const want = "arn:aws:iam::123:policy/GlobalBlanket"
	if got.Mandatory.PermissionsBoundaryArn != want {
		t.Errorf("AC-2: level-1 must win; got %q, want %q",
			got.Mandatory.PermissionsBoundaryArn, want)
	}
}

// TestMergeIAMCascade_AC3 — globalKropathConfig.mandatory.iam.blockIamUserAccessKeys
// propagates across all profiles and namespaces.
func TestMergeIAMCascade_AC3(t *testing.T) {
	globalKropathMandatory := cascade.IAMSection{
		BlockIamUserAccessKeys: true,
	}

	got := mergeAll(globalKropathMandatory, zero, zero, zero, zero, zero, zero, zero)

	if !got.Mandatory.BlockIamUserAccessKeys {
		t.Error("AC-3: mandatory.blockIamUserAccessKeys should be true")
	}
}

// TestMergeIAMCascade_AC4 — globalKropathConfig level-1 maxSessionDurationSeconds
// wins over globalIAMConfig level-3 when both are set.
func TestMergeIAMCascade_AC4(t *testing.T) {
	globalKropathMandatory := cascade.IAMSection{
		MaxSessionDurationSeconds: 3600,
	}
	globalIAMCfgMandatory := cascade.IAMSection{
		MaxSessionDurationSeconds: 7200,
	}

	got := mergeAll(globalKropathMandatory, zero, globalIAMCfgMandatory, zero, zero, zero, zero, zero)

	if got.Mandatory.MaxSessionDurationSeconds != 3600 {
		t.Errorf("AC-4: mandatory.maxSessionDurationSeconds = %d, want 3600",
			got.Mandatory.MaxSessionDurationSeconds)
	}
}

// TestMergeIAMCascade_AC5 — only globalIAMConfig.defaults.maxSessionDurationSeconds
// set; no mandatory or KropathConfig values. Defaults populated; mandatory zero.
func TestMergeIAMCascade_AC5(t *testing.T) {
	globalIAMCfgDefaults := cascade.IAMSection{
		MaxSessionDurationSeconds: 3600,
	}

	got := mergeAll(zero, zero, zero, zero, zero, globalIAMCfgDefaults, zero, zero)

	if got.Defaults.MaxSessionDurationSeconds != 3600 {
		t.Errorf("AC-5: defaults.maxSessionDurationSeconds = %d, want 3600",
			got.Defaults.MaxSessionDurationSeconds)
	}
	if got.Mandatory.MaxSessionDurationSeconds != 0 {
		t.Errorf("AC-5: mandatory.maxSessionDurationSeconds should be 0, got %d",
			got.Mandatory.MaxSessionDurationSeconds)
	}
}

// TestMergeIAMCascade_AC6 — global mandatory (level 1) wins over local
// mandatory (level 2); org-wide mandatory cannot be overridden by namespace config.
func TestMergeIAMCascade_AC6(t *testing.T) {
	globalKropathMandatory := cascade.IAMSection{
		PermissionsBoundaryArn: "arn:aws:iam::123:policy/GlobalBlanket",
	}
	localKropathMandatory := cascade.IAMSection{
		PermissionsBoundaryArn: "arn:aws:iam::123:policy/NsBlanket",
	}

	got := mergeAll(globalKropathMandatory, localKropathMandatory, zero, zero, zero, zero, zero, zero)

	const want = "arn:aws:iam::123:policy/GlobalBlanket"
	if got.Mandatory.PermissionsBoundaryArn != want {
		t.Errorf("AC-6: level-1 global must win; got %q, want %q",
			got.Mandatory.PermissionsBoundaryArn, want)
	}
}

// TestMergeIAMCascade_AllAbsent — when all sources are zero, effectiveConfig
// fields are all zero (permissive; no governance enforced).
func TestMergeIAMCascade_AllAbsent(t *testing.T) {
	got := mergeAll(zero, zero, zero, zero, zero, zero, zero, zero)

	if got.Mandatory.PermissionsBoundaryArn != "" {
		t.Errorf("all-absent: mandatory.permissionsBoundaryArn should be empty")
	}
	if got.Mandatory.BlockIamUserAccessKeys {
		t.Error("all-absent: mandatory.blockIamUserAccessKeys should be false")
	}
	if got.Mandatory.MaxSessionDurationSeconds != 0 {
		t.Errorf("all-absent: mandatory.maxSessionDurationSeconds should be 0")
	}
	if got.Defaults.PermissionsBoundaryArn != "" {
		t.Errorf("all-absent: defaults.permissionsBoundaryArn should be empty")
	}
	if got.Defaults.BlockIamUserAccessKeys {
		t.Error("all-absent: defaults.blockIamUserAccessKeys should be false")
	}
	if got.Defaults.MaxSessionDurationSeconds != 0 {
		t.Errorf("all-absent: defaults.maxSessionDurationSeconds should be 0")
	}
}

// TestMergeIAMCascade_DefaultsCascadeOrder — verifies the defaults priority order
// (level 6 > 7 > 8 > 9).
func TestMergeIAMCascade_DefaultsCascadeOrder(t *testing.T) {
	cases := []struct {
		name                  string
		localIAMCfgDefaults   cascade.IAMSection
		globalIAMCfgDefaults  cascade.IAMSection
		localKropathDefaults  cascade.IAMSection
		globalKropathDefaults cascade.IAMSection
		wantMaxSession        int64
		wantBoundaryArn       string
	}{
		{
			name:                  "level6-wins",
			localIAMCfgDefaults:   cascade.IAMSection{MaxSessionDurationSeconds: 1800},
			globalIAMCfgDefaults:  cascade.IAMSection{MaxSessionDurationSeconds: 3600},
			localKropathDefaults:  cascade.IAMSection{MaxSessionDurationSeconds: 7200},
			globalKropathDefaults: cascade.IAMSection{MaxSessionDurationSeconds: 43200},
			wantMaxSession:        1800,
		},
		{
			name:                  "level7-wins-when-6-absent",
			localIAMCfgDefaults:   zero,
			globalIAMCfgDefaults:  cascade.IAMSection{MaxSessionDurationSeconds: 3600},
			localKropathDefaults:  cascade.IAMSection{MaxSessionDurationSeconds: 7200},
			globalKropathDefaults: cascade.IAMSection{MaxSessionDurationSeconds: 43200},
			wantMaxSession:        3600,
		},
		{
			name:                  "level8-wins-when-6-7-absent",
			localIAMCfgDefaults:   zero,
			globalIAMCfgDefaults:  zero,
			localKropathDefaults:  cascade.IAMSection{MaxSessionDurationSeconds: 7200},
			globalKropathDefaults: cascade.IAMSection{MaxSessionDurationSeconds: 43200},
			wantMaxSession:        7200,
		},
		{
			name:                  "level9-wins-when-6-7-8-absent",
			localIAMCfgDefaults:   zero,
			globalIAMCfgDefaults:  zero,
			localKropathDefaults:  zero,
			globalKropathDefaults: cascade.IAMSection{MaxSessionDurationSeconds: 43200},
			wantMaxSession:        43200,
		},
		{
			name:                  "boundary-level6-wins",
			localIAMCfgDefaults:   cascade.IAMSection{PermissionsBoundaryArn: "arn:level6"},
			globalIAMCfgDefaults:  cascade.IAMSection{PermissionsBoundaryArn: "arn:level7"},
			localKropathDefaults:  cascade.IAMSection{PermissionsBoundaryArn: "arn:level8"},
			globalKropathDefaults: cascade.IAMSection{PermissionsBoundaryArn: "arn:level9"},
			wantBoundaryArn:       "arn:level6",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeAll(
				zero, zero, zero, zero,
				tc.localIAMCfgDefaults,
				tc.globalIAMCfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
			)
			if tc.wantMaxSession != 0 && got.Defaults.MaxSessionDurationSeconds != tc.wantMaxSession {
				t.Errorf("maxSessionDurationSeconds = %d, want %d",
					got.Defaults.MaxSessionDurationSeconds, tc.wantMaxSession)
			}
			if tc.wantBoundaryArn != "" && got.Defaults.PermissionsBoundaryArn != tc.wantBoundaryArn {
				t.Errorf("permissionsBoundaryArn = %q, want %q",
					got.Defaults.PermissionsBoundaryArn, tc.wantBoundaryArn)
			}
		})
	}
}

// TestMergeIAMCascade_MandatoryCascadeOrder — verifies the mandatory priority order
// (level 1 > 2 > 3 > 4).
func TestMergeIAMCascade_MandatoryCascadeOrder(t *testing.T) {
	cases := []struct {
		name                   string
		globalKropathMandatory cascade.IAMSection
		localKropathMandatory  cascade.IAMSection
		globalIAMCfgMandatory  cascade.IAMSection
		localIAMCfgMandatory   cascade.IAMSection
		wantMaxSession         int64
	}{
		{
			name:                   "level1-wins",
			globalKropathMandatory: cascade.IAMSection{MaxSessionDurationSeconds: 1800},
			localKropathMandatory:  cascade.IAMSection{MaxSessionDurationSeconds: 3600},
			globalIAMCfgMandatory:  cascade.IAMSection{MaxSessionDurationSeconds: 7200},
			localIAMCfgMandatory:   cascade.IAMSection{MaxSessionDurationSeconds: 43200},
			wantMaxSession:         1800,
		},
		{
			name:                   "level2-wins-when-1-absent",
			globalKropathMandatory: zero,
			localKropathMandatory:  cascade.IAMSection{MaxSessionDurationSeconds: 3600},
			globalIAMCfgMandatory:  cascade.IAMSection{MaxSessionDurationSeconds: 7200},
			localIAMCfgMandatory:   cascade.IAMSection{MaxSessionDurationSeconds: 43200},
			wantMaxSession:         3600,
		},
		{
			name:                   "level3-wins-when-1-2-absent",
			globalKropathMandatory: zero,
			localKropathMandatory:  zero,
			globalIAMCfgMandatory:  cascade.IAMSection{MaxSessionDurationSeconds: 7200},
			localIAMCfgMandatory:   cascade.IAMSection{MaxSessionDurationSeconds: 43200},
			wantMaxSession:         7200,
		},
		{
			name:                   "level4-wins-when-1-2-3-absent",
			globalKropathMandatory: zero,
			localKropathMandatory:  zero,
			globalIAMCfgMandatory:  zero,
			localIAMCfgMandatory:   cascade.IAMSection{MaxSessionDurationSeconds: 43200},
			wantMaxSession:         43200,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeAll(
				tc.globalKropathMandatory,
				tc.localKropathMandatory,
				tc.globalIAMCfgMandatory,
				tc.localIAMCfgMandatory,
				zero, zero, zero, zero,
			)
			if got.Mandatory.MaxSessionDurationSeconds != tc.wantMaxSession {
				t.Errorf("maxSessionDurationSeconds = %d, want %d",
					got.Mandatory.MaxSessionDurationSeconds, tc.wantMaxSession)
			}
		})
	}
}

// TestMergeIAMCascade_MandatoryIsolatedFromDefaults — mandatory fields must not
// bleed into defaults and vice versa.
func TestMergeIAMCascade_MandatoryIsolatedFromDefaults(t *testing.T) {
	globalKropathMandatory := cascade.IAMSection{
		PermissionsBoundaryArn:    "arn:mandatory",
		BlockIamUserAccessKeys:    true,
		MaxSessionDurationSeconds: 3600,
	}
	globalIAMCfgDefaults := cascade.IAMSection{
		PermissionsBoundaryArn:    "arn:defaults",
		MaxSessionDurationSeconds: 7200,
	}

	got := mergeAll(globalKropathMandatory, zero, zero, zero, zero, globalIAMCfgDefaults, zero, zero)

	if got.Mandatory.PermissionsBoundaryArn != "arn:mandatory" {
		t.Errorf("mandatory.boundary = %q, want arn:mandatory", got.Mandatory.PermissionsBoundaryArn)
	}
	if got.Mandatory.MaxSessionDurationSeconds != 3600 {
		t.Errorf("mandatory.maxSession = %d, want 3600", got.Mandatory.MaxSessionDurationSeconds)
	}
	if got.Defaults.PermissionsBoundaryArn != "arn:defaults" {
		t.Errorf("defaults.boundary = %q, want arn:defaults", got.Defaults.PermissionsBoundaryArn)
	}
	if got.Defaults.MaxSessionDurationSeconds != 7200 {
		t.Errorf("defaults.maxSession = %d, want 7200", got.Defaults.MaxSessionDurationSeconds)
	}
}
