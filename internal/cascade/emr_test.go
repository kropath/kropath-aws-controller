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

// emptyEMRKropath is the zero-value KropathSection — no enforcement at that level.
var emptyEMRKropath = cascade.EMRKropathSection{}

// emptyEMRCfg is the zero-value EMRConfigSection — no enforcement at that level.
var emptyEMRCfg = cascade.EMRConfigSection{}

// TestMergeEMRCascade_GlobalKropathMandatoryWins verifies that a globalKropathMandatory
// value beats all other mandatory levels (AC-1 / mandatory override path).
func TestMergeEMRCascade_GlobalKropathMandatoryWins(t *testing.T) {
	t.Parallel()
	result := cascade.MergeEMRCascade(
		cascade.EMRKropathSection{ReleaseLabel: "emr-7.0.0"}, // level 1 — wins
		cascade.EMRKropathSection{ReleaseLabel: "emr-6.0.0"}, // level 2
		cascade.EMRConfigSection{ReleaseLabel: "emr-5.0.0"},  // level 3
		cascade.EMRConfigSection{ReleaseLabel: "emr-4.0.0"},  // level 4
		emptyEMRCfg,                                           // level 6
		emptyEMRCfg,                                           // level 7
		emptyEMRKropath,                                       // level 8
		emptyEMRKropath,                                       // level 9
	)
	if got := result.Mandatory.ReleaseLabel; got != "emr-7.0.0" {
		t.Errorf("mandatory.releaseLabel = %q; want %q", got, "emr-7.0.0")
	}
}

// TestMergeEMRCascade_LocalEMRCfgMandatoryFallsThrough verifies that when no
// higher mandatory level is set, the local EMRConfig mandatory value is used.
func TestMergeEMRCascade_LocalEMRCfgMandatoryFallsThrough(t *testing.T) {
	t.Parallel()
	result := cascade.MergeEMRCascade(
		emptyEMRKropath, // level 1 — empty
		emptyEMRKropath, // level 2 — empty
		emptyEMRCfg,     // level 3 — empty
		cascade.EMRConfigSection{ReleaseLabel: "emr-4.0.0"}, // level 4 — only source
		emptyEMRCfg,
		emptyEMRCfg,
		emptyEMRKropath,
		emptyEMRKropath,
	)
	if got := result.Mandatory.ReleaseLabel; got != "emr-4.0.0" {
		t.Errorf("mandatory.releaseLabel = %q; want %q", got, "emr-4.0.0")
	}
}

// TestMergeEMRCascade_DefaultsPriorityOrder verifies the defaults tier (L6 > L7 > L8 > L9).
func TestMergeEMRCascade_DefaultsPriorityOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		l6   cascade.EMRConfigSection
		l7   cascade.EMRConfigSection
		l8   cascade.EMRKropathSection
		l9   cascade.EMRKropathSection
		want string
	}{
		{
			name: "L6 wins",
			l6:   cascade.EMRConfigSection{ReleaseLabel: "emr-local-cfg"},
			l7:   cascade.EMRConfigSection{ReleaseLabel: "emr-global-cfg"},
			l8:   cascade.EMRKropathSection{ReleaseLabel: "emr-local-kpc"},
			l9:   cascade.EMRKropathSection{ReleaseLabel: "emr-global-kpc"},
			want: "emr-local-cfg",
		},
		{
			name: "L7 wins when L6 empty",
			l6:   emptyEMRCfg,
			l7:   cascade.EMRConfigSection{ReleaseLabel: "emr-global-cfg"},
			l8:   cascade.EMRKropathSection{ReleaseLabel: "emr-local-kpc"},
			l9:   cascade.EMRKropathSection{ReleaseLabel: "emr-global-kpc"},
			want: "emr-global-cfg",
		},
		{
			name: "L8 wins when L6,L7 empty",
			l6:   emptyEMRCfg,
			l7:   emptyEMRCfg,
			l8:   cascade.EMRKropathSection{ReleaseLabel: "emr-local-kpc"},
			l9:   cascade.EMRKropathSection{ReleaseLabel: "emr-global-kpc"},
			want: "emr-local-kpc",
		},
		{
			name: "L9 fallback",
			l6:   emptyEMRCfg,
			l7:   emptyEMRCfg,
			l8:   emptyEMRKropath,
			l9:   cascade.EMRKropathSection{ReleaseLabel: "emr-global-kpc"},
			want: "emr-global-kpc",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := cascade.MergeEMRCascade(
				emptyEMRKropath, emptyEMRKropath, emptyEMRCfg, emptyEMRCfg,
				tc.l6, tc.l7, tc.l8, tc.l9,
			)
			if got := result.Defaults.ReleaseLabel; got != tc.want {
				t.Errorf("defaults.releaseLabel = %q; want %q", got, tc.want)
			}
		})
	}
}

// TestMergeEMRCascade_EMROnlyFields verifies that EMRConfig-only fields
// (not present in KropathConfig) are resolved from levels 3-4 (mandatory)
// and 6-7 (defaults) only.
func TestMergeEMRCascade_EMROnlyFields(t *testing.T) {
	t.Parallel()

	result := cascade.MergeEMRCascade(
		emptyEMRKropath, // level 1
		emptyEMRKropath, // level 2
		cascade.EMRConfigSection{ // level 3 — global EMRConfig mandatory
			AutoStopIdleTimeoutMinutes: 15,
			MaximumCapacityCPU:         "200",
			MaximumCapacityMemory:      "2000g",
			DiskEncryptionKeyARN:       "arn:aws:kms:us-east-1:111111111111:key/global",
			NamingTemplate:             "{namespace}-{name}-global",
		},
		cascade.EMRConfigSection{ // level 4 — local EMRConfig mandatory (lower priority)
			AutoStopIdleTimeoutMinutes: 30,
			MaximumCapacityCPU:         "400",
			MaximumCapacityMemory:      "4000g",
			DiskEncryptionKeyARN:       "arn:aws:kms:us-east-1:111111111111:key/local",
			NamingTemplate:             "{namespace}-{name}-local",
		},
		cascade.EMRConfigSection{ // level 6 — local EMRConfig defaults
			AutoStopIdleTimeoutMinutes: 60,
			MaximumCapacityCPU:         "100",
			MaximumCapacityMemory:      "1000g",
			DiskEncryptionKeyARN:       "arn:aws:kms:us-east-1:111111111111:key/defaults-local",
			NamingTemplate:             "{namespace}-{name}-defaults-local",
		},
		emptyEMRCfg,    // level 7
		emptyEMRKropath, // level 8
		emptyEMRKropath, // level 9
	)

	// Mandatory: level 3 wins over level 4
	if got := result.Mandatory.AutoStopIdleTimeoutMinutes; got != 15 {
		t.Errorf("mandatory.autoStopIdleTimeoutMinutes = %d; want 15", got)
	}
	if got := result.Mandatory.MaximumCapacityCPU; got != "200" {
		t.Errorf("mandatory.maximumCapacityCPU = %q; want %q", got, "200")
	}
	if got := result.Mandatory.MaximumCapacityMemory; got != "2000g" {
		t.Errorf("mandatory.maximumCapacityMemory = %q; want %q", got, "2000g")
	}
	if got := result.Mandatory.DiskEncryptionKeyARN; got != "arn:aws:kms:us-east-1:111111111111:key/global" {
		t.Errorf("mandatory.diskEncryptionKeyARN = %q; want global key", got)
	}
	if got := result.Mandatory.NamingTemplate; got != "{namespace}-{name}-global" {
		t.Errorf("mandatory.namingTemplate = %q; want global template", got)
	}
	// Defaults: level 6 is the only non-zero source
	if got := result.Defaults.AutoStopIdleTimeoutMinutes; got != 60 {
		t.Errorf("defaults.autoStopIdleTimeoutMinutes = %d; want 60", got)
	}
}

// TestMergeEMRCascade_AutoStopZeroIsNotSet verifies the integer sentinel:
// a value of 0 means "not enforced" and is skipped in priority resolution.
func TestMergeEMRCascade_AutoStopZeroIsNotSet(t *testing.T) {
	t.Parallel()

	result := cascade.MergeEMRCascade(
		emptyEMRKropath,
		emptyEMRKropath,
		cascade.EMRConfigSection{AutoStopIdleTimeoutMinutes: 0},  // level 3 — zero = not set
		cascade.EMRConfigSection{AutoStopIdleTimeoutMinutes: 20}, // level 4 — fallback
		emptyEMRCfg,
		emptyEMRCfg,
		emptyEMRKropath,
		emptyEMRKropath,
	)
	if got := result.Mandatory.AutoStopIdleTimeoutMinutes; got != 20 {
		t.Errorf("mandatory.autoStopIdleTimeoutMinutes = %d; want 20 (zero sentinel skipped)", got)
	}
}

// TestMergeEMRCascade_TagsAdditiveHigherLevelWins verifies that tags are additively merged
// with the highest priority level winning on key conflicts.
func TestMergeEMRCascade_TagsAdditiveHigherLevelWins(t *testing.T) {
	t.Parallel()

	result := cascade.MergeEMRCascade(
		cascade.EMRKropathSection{Tags: map[string]string{"env": "prod", "source": "global-kpc-m"}},  // level 1
		cascade.EMRKropathSection{Tags: map[string]string{"team": "platform", "source": "local-kpc-m"}}, // level 2
		cascade.EMRConfigSection{Tags: map[string]string{"service": "emr", "source": "global-emr-m"}}, // level 3
		cascade.EMRConfigSection{Tags: map[string]string{"instance": "main", "source": "local-emr-m"}}, // level 4
		emptyEMRCfg,
		emptyEMRCfg,
		emptyEMRKropath,
		emptyEMRKropath,
	)

	// All unique keys should be present
	mandatoryTags := result.Mandatory.Tags
	if mandatoryTags["env"] != "prod" {
		t.Errorf("mandatory.tags.env = %q; want %q", mandatoryTags["env"], "prod")
	}
	if mandatoryTags["team"] != "platform" {
		t.Errorf("mandatory.tags.team = %q; want %q", mandatoryTags["team"], "platform")
	}
	if mandatoryTags["service"] != "emr" {
		t.Errorf("mandatory.tags.service = %q; want %q", mandatoryTags["service"], "emr")
	}
	if mandatoryTags["instance"] != "main" {
		t.Errorf("mandatory.tags.instance = %q; want %q", mandatoryTags["instance"], "main")
	}
	// "source" key conflict: level 1 (globalKropathMandatory) wins
	if mandatoryTags["source"] != "global-kpc-m" {
		t.Errorf("mandatory.tags.source = %q; want %q (L1 wins)", mandatoryTags["source"], "global-kpc-m")
	}
}

// TestMergeEMRCascade_DefaultsTagsAdditiveL6Wins verifies defaults tag merge:
// level 6 (local EMRConfig defaults) wins on key conflicts.
func TestMergeEMRCascade_DefaultsTagsAdditiveL6Wins(t *testing.T) {
	t.Parallel()

	result := cascade.MergeEMRCascade(
		emptyEMRKropath,
		emptyEMRKropath,
		emptyEMRCfg,
		emptyEMRCfg,
		cascade.EMRConfigSection{Tags: map[string]string{"env": "dev", "source": "l6"}},       // level 6 — wins
		cascade.EMRConfigSection{Tags: map[string]string{"env": "staging", "source": "l7"}},    // level 7
		cascade.EMRKropathSection{Tags: map[string]string{"org": "kropath", "source": "l8"}},  // level 8
		cascade.EMRKropathSection{Tags: map[string]string{"global": "true", "source": "l9"}},  // level 9
	)

	defaultsTags := result.Defaults.Tags
	if defaultsTags["source"] != "l6" {
		t.Errorf("defaults.tags.source = %q; want %q (L6 wins)", defaultsTags["source"], "l6")
	}
	if defaultsTags["env"] != "dev" {
		t.Errorf("defaults.tags.env = %q; want %q", defaultsTags["env"], "dev")
	}
	if defaultsTags["org"] != "kropath" {
		t.Errorf("defaults.tags.org = %q; want %q (additive from L8)", defaultsTags["org"], "kropath")
	}
	if defaultsTags["global"] != "true" {
		t.Errorf("defaults.tags.global = %q; want %q (additive from L9)", defaultsTags["global"], "true")
	}
}

// TestMergeEMRCascade_SyncedLabelsAndAnnotations verifies that syncedLabels and syncedAnnotations
// merge additively within EMRConfig levels (no KropathConfig source for these fields).
func TestMergeEMRCascade_SyncedLabelsAndAnnotations(t *testing.T) {
	t.Parallel()

	result := cascade.MergeEMRCascade(
		emptyEMRKropath,
		emptyEMRKropath,
		cascade.EMRConfigSection{
			SyncedLabels:      map[string]string{"k8s-label": "global", "shared": "global-emr"},
			SyncedAnnotations: map[string]string{"k8s-ann": "global"},
		}, // level 3
		cascade.EMRConfigSection{
			SyncedLabels:      map[string]string{"local-label": "yes", "shared": "local-emr"},
			SyncedAnnotations: map[string]string{"local-ann": "yes"},
		}, // level 4 — lower priority
		cascade.EMRConfigSection{
			SyncedLabels:      map[string]string{"def-label": "yes", "shared": "local-def"},
			SyncedAnnotations: map[string]string{"def-ann": "yes"},
		}, // level 6
		cascade.EMRConfigSection{
			SyncedLabels:      map[string]string{"global-def-label": "yes", "shared": "global-def"},
			SyncedAnnotations: map[string]string{"global-def-ann": "yes"},
		}, // level 7
		emptyEMRKropath,
		emptyEMRKropath,
	)

	// Mandatory syncedLabels: L3 wins on "shared" key
	if got := result.Mandatory.SyncedLabels["shared"]; got != "global-emr" {
		t.Errorf("mandatory.syncedLabels.shared = %q; want %q (L3 wins)", got, "global-emr")
	}
	if _, ok := result.Mandatory.SyncedLabels["k8s-label"]; !ok {
		t.Error("mandatory.syncedLabels missing k8s-label from L3")
	}
	if _, ok := result.Mandatory.SyncedLabels["local-label"]; !ok {
		t.Error("mandatory.syncedLabels missing local-label from L4 (additive)")
	}

	// Defaults syncedLabels: L6 wins on "shared" key
	if got := result.Defaults.SyncedLabels["shared"]; got != "local-def" {
		t.Errorf("defaults.syncedLabels.shared = %q; want %q (L6 wins)", got, "local-def")
	}
	if _, ok := result.Defaults.SyncedLabels["global-def-label"]; !ok {
		t.Error("defaults.syncedLabels missing global-def-label from L7 (additive)")
	}
}

// TestMergeEMRCascade_AllZeroProducesZero verifies that when all inputs are zero-valued,
// the output is also zero-valued (no panics, no unexpected defaults).
func TestMergeEMRCascade_AllZeroProducesZero(t *testing.T) {
	t.Parallel()
	result := cascade.MergeEMRCascade(
		emptyEMRKropath, emptyEMRKropath, emptyEMRCfg, emptyEMRCfg,
		emptyEMRCfg, emptyEMRCfg, emptyEMRKropath, emptyEMRKropath,
	)
	if result.Mandatory.ReleaseLabel != "" {
		t.Errorf("mandatory.releaseLabel = %q; want empty", result.Mandatory.ReleaseLabel)
	}
	if result.Defaults.ReleaseLabel != "" {
		t.Errorf("defaults.releaseLabel = %q; want empty", result.Defaults.ReleaseLabel)
	}
	if result.Mandatory.AutoStopIdleTimeoutMinutes != 0 {
		t.Errorf("mandatory.autoStopIdleTimeoutMinutes = %d; want 0", result.Mandatory.AutoStopIdleTimeoutMinutes)
	}
	if result.Mandatory.Tags != nil {
		t.Errorf("mandatory.tags should be nil when all inputs are empty; got %v", result.Mandatory.Tags)
	}
}

// TestMergeEMRCascade_ArchitectureIndependentOfReleaseLabel verifies that architecture
// and releaseLabel are resolved independently by their own priority chains.
func TestMergeEMRCascade_ArchitectureIndependentOfReleaseLabel(t *testing.T) {
	t.Parallel()

	result := cascade.MergeEMRCascade(
		cascade.EMRKropathSection{Architecture: "ARM64"},         // level 1: sets architecture
		emptyEMRKropath,                                           // level 2
		cascade.EMRConfigSection{ReleaseLabel: "emr-7.1.0"},     // level 3: sets releaseLabel
		emptyEMRCfg,                                              // level 4
		emptyEMRCfg,
		emptyEMRCfg,
		emptyEMRKropath,
		emptyEMRKropath,
	)

	if got := result.Mandatory.Architecture; got != "ARM64" {
		t.Errorf("mandatory.architecture = %q; want %q", got, "ARM64")
	}
	if got := result.Mandatory.ReleaseLabel; got != "emr-7.1.0" {
		t.Errorf("mandatory.releaseLabel = %q; want %q", got, "emr-7.1.0")
	}
}
