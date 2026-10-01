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

// emptyGlueKropath is the zero-value KropathSection — no enforcement at that level.
var emptyGlueKropath = cascade.GlueKropathSection{}

// emptyGlueCfg is the zero-value GlueConfigSection — no enforcement at that level.
var emptyGlueCfg = cascade.GlueConfigSection{}

// TestMergeGlueCascade_GlobalKropathMandatoryWins verifies that a globalKropathMandatory
// value beats all other mandatory levels (string field — glueVersion).
func TestMergeGlueCascade_GlobalKropathMandatoryWins(t *testing.T) {
	t.Parallel()
	result := cascade.MergeGlueCascade(
		cascade.GlueKropathSection{GlueVersion: "4.0"}, // level 1 — wins
		cascade.GlueKropathSection{GlueVersion: "3.0"}, // level 2
		cascade.GlueConfigSection{GlueVersion: "2.0"},  // level 3
		cascade.GlueConfigSection{GlueVersion: "1.0"},  // level 4
		emptyGlueCfg,                                   // level 6
		emptyGlueCfg,                                   // level 7
		emptyGlueKropath,                               // level 8
		emptyGlueKropath,                               // level 9
	)
	if got := result.Mandatory.GlueVersion; got != "4.0" {
		t.Errorf("mandatory.glueVersion = %q; want %q", got, "4.0")
	}
}

// TestMergeGlueCascade_LocalGlueCfgMandatoryFallsThrough verifies that when no
// higher mandatory level is set, the local GlueConfig mandatory value is used.
func TestMergeGlueCascade_LocalGlueCfgMandatoryFallsThrough(t *testing.T) {
	t.Parallel()
	result := cascade.MergeGlueCascade(
		emptyGlueKropath, // level 1 — empty
		emptyGlueKropath, // level 2 — empty
		emptyGlueCfg,     // level 3 — empty
		cascade.GlueConfigSection{GlueVersion: "3.0"}, // level 4 — only source
		emptyGlueCfg,
		emptyGlueCfg,
		emptyGlueKropath,
		emptyGlueKropath,
	)
	if got := result.Mandatory.GlueVersion; got != "3.0" {
		t.Errorf("mandatory.glueVersion = %q; want %q", got, "3.0")
	}
}

// TestMergeGlueCascade_DefaultsPriorityOrder verifies the defaults tier (L6 > L7 > L8 > L9).
func TestMergeGlueCascade_DefaultsPriorityOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		l6   cascade.GlueConfigSection
		l7   cascade.GlueConfigSection
		l8   cascade.GlueKropathSection
		l9   cascade.GlueKropathSection
		want string
	}{
		{
			name: "L6 wins",
			l6:   cascade.GlueConfigSection{GlueVersion: "local-cfg"},
			l7:   cascade.GlueConfigSection{GlueVersion: "global-cfg"},
			l8:   cascade.GlueKropathSection{GlueVersion: "local-kpc"},
			l9:   cascade.GlueKropathSection{GlueVersion: "global-kpc"},
			want: "local-cfg",
		},
		{
			name: "L7 wins when L6 empty",
			l6:   emptyGlueCfg,
			l7:   cascade.GlueConfigSection{GlueVersion: "global-cfg"},
			l8:   cascade.GlueKropathSection{GlueVersion: "local-kpc"},
			l9:   cascade.GlueKropathSection{GlueVersion: "global-kpc"},
			want: "global-cfg",
		},
		{
			name: "L8 wins when L6,L7 empty",
			l6:   emptyGlueCfg,
			l7:   emptyGlueCfg,
			l8:   cascade.GlueKropathSection{GlueVersion: "local-kpc"},
			l9:   cascade.GlueKropathSection{GlueVersion: "global-kpc"},
			want: "local-kpc",
		},
		{
			name: "L9 fallback",
			l6:   emptyGlueCfg,
			l7:   emptyGlueCfg,
			l8:   emptyGlueKropath,
			l9:   cascade.GlueKropathSection{GlueVersion: "global-kpc"},
			want: "global-kpc",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := cascade.MergeGlueCascade(
				emptyGlueKropath, emptyGlueKropath, emptyGlueCfg, emptyGlueCfg,
				tc.l6, tc.l7, tc.l8, tc.l9,
			)
			if got := result.Defaults.GlueVersion; got != tc.want {
				t.Errorf("defaults.glueVersion = %q; want %q", got, tc.want)
			}
		})
	}
}

// TestMergeGlueCascade_IntegerMandatoryCascade verifies that integer fields use
// zero-sentinel and respect priority order (numberOfWorkers, timeout, etc.).
func TestMergeGlueCascade_IntegerMandatoryCascade(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		l1              cascade.GlueKropathSection
		l2              cascade.GlueKropathSection
		l3              cascade.GlueConfigSection
		l4              cascade.GlueConfigSection
		wantWorkers     int64
		wantTimeout     int64
		wantMaxRetries  int64
		wantConcurrency int64
		wantNotify      int64
	}{
		{
			name:            "level-1 wins over all",
			l1:              cascade.GlueKropathSection{NumberOfWorkers: 10, Timeout: 120, MaxRetries: 3, MaxConcurrentRuns: 2, NotifyDelayAfter: 30},
			l2:              cascade.GlueKropathSection{NumberOfWorkers: 5, Timeout: 60},
			l3:              cascade.GlueConfigSection{NumberOfWorkers: 3, Timeout: 30},
			l4:              cascade.GlueConfigSection{NumberOfWorkers: 2, Timeout: 15},
			wantWorkers:     10,
			wantTimeout:     120,
			wantMaxRetries:  3,
			wantConcurrency: 2,
			wantNotify:      30,
		},
		{
			name:        "zero skipped — level-3 wins when L1,L2 zero",
			l1:          emptyGlueKropath,
			l2:          emptyGlueKropath,
			l3:          cascade.GlueConfigSection{NumberOfWorkers: 8, Timeout: 480},
			l4:          cascade.GlueConfigSection{NumberOfWorkers: 4, Timeout: 240},
			wantWorkers: 8,
			wantTimeout: 480,
		},
		{
			name:        "level-4 fallthrough when L1-3 zero",
			l1:          emptyGlueKropath,
			l2:          emptyGlueKropath,
			l3:          emptyGlueCfg,
			l4:          cascade.GlueConfigSection{NumberOfWorkers: 2, MaxRetries: 1},
			wantWorkers: 2,
			wantMaxRetries: 1,
		},
		{
			name:        "all zero — all remain zero (not enforced)",
			l1:          emptyGlueKropath,
			l2:          emptyGlueKropath,
			l3:          emptyGlueCfg,
			l4:          emptyGlueCfg,
			wantWorkers: 0,
			wantTimeout: 0,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := cascade.MergeGlueCascade(
				tc.l1, tc.l2, tc.l3, tc.l4,
				emptyGlueCfg, emptyGlueCfg, emptyGlueKropath, emptyGlueKropath,
			)
			if got := result.Mandatory.NumberOfWorkers; got != tc.wantWorkers {
				t.Errorf("mandatory.numberOfWorkers = %d; want %d", got, tc.wantWorkers)
			}
			if got := result.Mandatory.Timeout; got != tc.wantTimeout {
				t.Errorf("mandatory.timeout = %d; want %d", got, tc.wantTimeout)
			}
			if got := result.Mandatory.MaxRetries; got != tc.wantMaxRetries {
				t.Errorf("mandatory.maxRetries = %d; want %d", got, tc.wantMaxRetries)
			}
			if got := result.Mandatory.MaxConcurrentRuns; got != tc.wantConcurrency {
				t.Errorf("mandatory.maxConcurrentRuns = %d; want %d", got, tc.wantConcurrency)
			}
			if got := result.Mandatory.NotifyDelayAfter; got != tc.wantNotify {
				t.Errorf("mandatory.notifyDelayAfter = %d; want %d", got, tc.wantNotify)
			}
		})
	}
}

// TestMergeGlueCascade_IntegerDefaultsCascade verifies integer field cascade in the defaults tier.
func TestMergeGlueCascade_IntegerDefaultsCascade(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		l6          cascade.GlueConfigSection
		l7          cascade.GlueConfigSection
		l8          cascade.GlueKropathSection
		l9          cascade.GlueKropathSection
		wantWorkers int64
	}{
		{name: "L6 wins", l6: cascade.GlueConfigSection{NumberOfWorkers: 5}, l7: cascade.GlueConfigSection{NumberOfWorkers: 3}, wantWorkers: 5},
		{name: "L7 wins when L6 zero", l6: emptyGlueCfg, l7: cascade.GlueConfigSection{NumberOfWorkers: 3}, l8: cascade.GlueKropathSection{NumberOfWorkers: 2}, wantWorkers: 3},
		{name: "L8 wins when L6,L7 zero", l6: emptyGlueCfg, l7: emptyGlueCfg, l8: cascade.GlueKropathSection{NumberOfWorkers: 2}, l9: cascade.GlueKropathSection{NumberOfWorkers: 1}, wantWorkers: 2},
		{name: "L9 fallback", l6: emptyGlueCfg, l7: emptyGlueCfg, l8: emptyGlueKropath, l9: cascade.GlueKropathSection{NumberOfWorkers: 1}, wantWorkers: 1},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := cascade.MergeGlueCascade(
				emptyGlueKropath, emptyGlueKropath, emptyGlueCfg, emptyGlueCfg,
				tc.l6, tc.l7, tc.l8, tc.l9,
			)
			if got := result.Defaults.NumberOfWorkers; got != tc.wantWorkers {
				t.Errorf("defaults.numberOfWorkers = %d; want %d", got, tc.wantWorkers)
			}
		})
	}
}

// TestMergeGlueCascade_AllNineScalarFieldsMandatory verifies all 9 scalar governance fields
// at the mandatory tier resolve correctly from a single source (level 3).
func TestMergeGlueCascade_AllNineScalarFieldsMandatory(t *testing.T) {
	t.Parallel()
	src := cascade.GlueConfigSection{
		GlueVersion:           "4.0",
		WorkerType:            "G.2X",
		NumberOfWorkers:       10,
		ExecutionClass:        "FLEX",
		Timeout:               480,
		MaxRetries:            3,
		MaxConcurrentRuns:     5,
		SecurityConfiguration: "org-sec",
		NotifyDelayAfter:      15,
	}
	result := cascade.MergeGlueCascade(
		emptyGlueKropath, emptyGlueKropath,
		src,               // level 3
		emptyGlueCfg,     // level 4
		emptyGlueCfg, emptyGlueCfg, emptyGlueKropath, emptyGlueKropath,
	)
	m := result.Mandatory
	if m.GlueVersion != "4.0" {
		t.Errorf("GlueVersion = %q; want %q", m.GlueVersion, "4.0")
	}
	if m.WorkerType != "G.2X" {
		t.Errorf("WorkerType = %q; want %q", m.WorkerType, "G.2X")
	}
	if m.NumberOfWorkers != 10 {
		t.Errorf("NumberOfWorkers = %d; want 10", m.NumberOfWorkers)
	}
	if m.ExecutionClass != "FLEX" {
		t.Errorf("ExecutionClass = %q; want %q", m.ExecutionClass, "FLEX")
	}
	if m.Timeout != 480 {
		t.Errorf("Timeout = %d; want 480", m.Timeout)
	}
	if m.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d; want 3", m.MaxRetries)
	}
	if m.MaxConcurrentRuns != 5 {
		t.Errorf("MaxConcurrentRuns = %d; want 5", m.MaxConcurrentRuns)
	}
	if m.SecurityConfiguration != "org-sec" {
		t.Errorf("SecurityConfiguration = %q; want %q", m.SecurityConfiguration, "org-sec")
	}
	if m.NotifyDelayAfter != 15 {
		t.Errorf("NotifyDelayAfter = %d; want 15", m.NotifyDelayAfter)
	}
}

// TestMergeGlueCascade_NamingTemplateNotInKropathConfig verifies that namingTemplate
// appears only at levels 3–4 (mandatory) and 6–7 (defaults).
func TestMergeGlueCascade_NamingTemplateNotInKropathConfig(t *testing.T) {
	t.Parallel()

	t.Run("mandatory level-3 wins", func(t *testing.T) {
		t.Parallel()
		result := cascade.MergeGlueCascade(
			emptyGlueKropath, emptyGlueKropath,
			cascade.GlueConfigSection{NamingTemplate: "global-{name}"},  // level 3
			cascade.GlueConfigSection{NamingTemplate: "local-{name}"},   // level 4
			emptyGlueCfg, emptyGlueCfg, emptyGlueKropath, emptyGlueKropath,
		)
		if got := result.Mandatory.NamingTemplate; got != "global-{name}" {
			t.Errorf("mandatory.namingTemplate = %q; want %q", got, "global-{name}")
		}
	})

	t.Run("defaults level-6 wins over level-7", func(t *testing.T) {
		t.Parallel()
		result := cascade.MergeGlueCascade(
			emptyGlueKropath, emptyGlueKropath, emptyGlueCfg, emptyGlueCfg,
			cascade.GlueConfigSection{NamingTemplate: "{namespace}-{name}"},  // level 6
			cascade.GlueConfigSection{NamingTemplate: "global-template"},     // level 7
			emptyGlueKropath, emptyGlueKropath,
		)
		if got := result.Defaults.NamingTemplate; got != "{namespace}-{name}" {
			t.Errorf("defaults.namingTemplate = %q; want %q", got, "{namespace}-{name}")
		}
	})
}

// TestMergeGlueCascade_TagsMergeMandatory verifies additive tag merge with priority
// for mandatory tier: level 1 wins on key conflicts.
func TestMergeGlueCascade_TagsMergeMandatory(t *testing.T) {
	t.Parallel()
	result := cascade.MergeGlueCascade(
		cascade.GlueKropathSection{Tags: map[string]string{"env": "prod", "kpc-only": "yes"}}, // level 1
		cascade.GlueKropathSection{Tags: map[string]string{"env": "staging"}},                  // level 2 — loses on "env"
		cascade.GlueConfigSection{Tags: map[string]string{"team": "data", "env": "dev"}},       // level 3 — loses on "env"
		cascade.GlueConfigSection{Tags: map[string]string{"project": "etl"}},                   // level 4
		emptyGlueCfg, emptyGlueCfg, emptyGlueKropath, emptyGlueKropath,
	)
	tags := result.Mandatory.Tags
	if tags["env"] != "prod" {
		t.Errorf("tags[env] = %q; want %q", tags["env"], "prod")
	}
	if tags["kpc-only"] != "yes" {
		t.Errorf("tags[kpc-only] = %q; want %q", tags["kpc-only"], "yes")
	}
	if tags["team"] != "data" {
		t.Errorf("tags[team] = %q; want %q", tags["team"], "data")
	}
	if tags["project"] != "etl" {
		t.Errorf("tags[project] = %q; want %q", tags["project"], "etl")
	}
}

// TestMergeGlueCascade_TagsMergeDefaults verifies additive tag merge for the defaults tier:
// level 6 wins on key conflicts.
func TestMergeGlueCascade_TagsMergeDefaults(t *testing.T) {
	t.Parallel()
	result := cascade.MergeGlueCascade(
		emptyGlueKropath, emptyGlueKropath, emptyGlueCfg, emptyGlueCfg,
		cascade.GlueConfigSection{Tags: map[string]string{"env": "local", "team": "glue"}}, // level 6 — wins
		cascade.GlueConfigSection{Tags: map[string]string{"env": "global"}},                 // level 7 — loses on "env"
		cascade.GlueKropathSection{Tags: map[string]string{"org": "acme"}},                  // level 8
		cascade.GlueKropathSection{Tags: map[string]string{"cost-center": "analytics"}},     // level 9
	)
	tags := result.Defaults.Tags
	if tags["env"] != "local" {
		t.Errorf("tags[env] = %q; want %q", tags["env"], "local")
	}
	if tags["team"] != "glue" {
		t.Errorf("tags[team] = %q; want %q", tags["team"], "glue")
	}
	if tags["org"] != "acme" {
		t.Errorf("tags[org] = %q; want %q", tags["org"], "acme")
	}
	if tags["cost-center"] != "analytics" {
		t.Errorf("tags[cost-center] = %q; want %q", tags["cost-center"], "analytics")
	}
}

// TestMergeGlueCascade_SyncedLabelsAndAnnotations verifies that syncedLabels and
// syncedAnnotations appear only at levels 3–4 (mandatory) and 6–7 (defaults).
func TestMergeGlueCascade_SyncedLabelsAndAnnotations(t *testing.T) {
	t.Parallel()

	t.Run("mandatory synced fields from levels 3 and 4", func(t *testing.T) {
		t.Parallel()
		result := cascade.MergeGlueCascade(
			emptyGlueKropath, emptyGlueKropath,
			cascade.GlueConfigSection{
				SyncedLabels:      map[string]string{"team": "platform"},
				SyncedAnnotations: map[string]string{"purpose": "etl"},
			},
			cascade.GlueConfigSection{
				SyncedLabels:      map[string]string{"service": "glue", "team": "local-override"},
				SyncedAnnotations: map[string]string{"owner": "data-eng"},
			},
			emptyGlueCfg, emptyGlueCfg, emptyGlueKropath, emptyGlueKropath,
		)
		// level 3 wins on "team" key
		if got := result.Mandatory.SyncedLabels["team"]; got != "platform" {
			t.Errorf("mandatory.syncedLabels[team] = %q; want %q", got, "platform")
		}
		if got := result.Mandatory.SyncedLabels["service"]; got != "glue" {
			t.Errorf("mandatory.syncedLabels[service] = %q; want %q", got, "glue")
		}
		if got := result.Mandatory.SyncedAnnotations["purpose"]; got != "etl" {
			t.Errorf("mandatory.syncedAnnotations[purpose] = %q; want %q", got, "etl")
		}
		if got := result.Mandatory.SyncedAnnotations["owner"]; got != "data-eng" {
			t.Errorf("mandatory.syncedAnnotations[owner] = %q; want %q", got, "data-eng")
		}
	})

	t.Run("defaults synced fields from levels 6 and 7", func(t *testing.T) {
		t.Parallel()
		result := cascade.MergeGlueCascade(
			emptyGlueKropath, emptyGlueKropath, emptyGlueCfg, emptyGlueCfg,
			cascade.GlueConfigSection{SyncedLabels: map[string]string{"env": "prod"}},  // level 6 wins
			cascade.GlueConfigSection{SyncedLabels: map[string]string{"env": "global", "region": "us-east-1"}}, // level 7
			emptyGlueKropath, emptyGlueKropath,
		)
		if got := result.Defaults.SyncedLabels["env"]; got != "prod" {
			t.Errorf("defaults.syncedLabels[env] = %q; want %q", got, "prod")
		}
		if got := result.Defaults.SyncedLabels["region"]; got != "us-east-1" {
			t.Errorf("defaults.syncedLabels[region] = %q; want %q", got, "us-east-1")
		}
	})
}

// TestMergeGlueCascade_AllFieldsEmpty verifies that when all sources are zero, all
// output fields are zero (no spurious defaults inserted by the merge function).
func TestMergeGlueCascade_AllFieldsEmpty(t *testing.T) {
	t.Parallel()
	result := cascade.MergeGlueCascade(
		emptyGlueKropath, emptyGlueKropath, emptyGlueCfg, emptyGlueCfg,
		emptyGlueCfg, emptyGlueCfg, emptyGlueKropath, emptyGlueKropath,
	)
	m := result.Mandatory
	if m.GlueVersion != "" || m.WorkerType != "" || m.NumberOfWorkers != 0 ||
		m.ExecutionClass != "" || m.Timeout != 0 || m.MaxRetries != 0 ||
		m.MaxConcurrentRuns != 0 || m.SecurityConfiguration != "" || m.NotifyDelayAfter != 0 ||
		m.NamingTemplate != "" || len(m.Tags) != 0 || len(m.SyncedLabels) != 0 || len(m.SyncedAnnotations) != 0 {
		t.Errorf("mandatory: expected all-zero result, got %+v", m)
	}
	d := result.Defaults
	if d.GlueVersion != "" || d.WorkerType != "" || d.NumberOfWorkers != 0 ||
		d.ExecutionClass != "" || d.Timeout != 0 || d.MaxRetries != 0 ||
		d.MaxConcurrentRuns != 0 || d.SecurityConfiguration != "" || d.NotifyDelayAfter != 0 ||
		d.NamingTemplate != "" || len(d.Tags) != 0 || len(d.SyncedLabels) != 0 || len(d.SyncedAnnotations) != 0 {
		t.Errorf("defaults: expected all-zero result, got %+v", d)
	}
}
