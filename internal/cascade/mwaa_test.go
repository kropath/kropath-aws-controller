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

// emptyMWAAKropath is the zero-value MWAAKropathSection — no enforcement at that level.
var emptyMWAAKropath = cascade.MWAAKropathSection{}

// emptyMWAACfg is the zero-value MWAAConfigSection — no enforcement at that level.
var emptyMWAACfg = cascade.MWAAConfigSection{}

// mergeMWAAAll calls MergeMWAACascade with all eight source levels.
func mergeMWAAAll(
	l1, l2 cascade.MWAAKropathSection,
	l3, l4, l6, l7 cascade.MWAAConfigSection,
	l8, l9 cascade.MWAAKropathSection,
) cascade.EffectiveMWAAConfig {
	return cascade.MergeMWAACascade(l1, l2, l3, l4, l6, l7, l8, l9)
}

// TestMergeMWAACascade_MandatoryPriorityOrder verifies the mandatory tier priority:
// L1 (globalKropathMandatory) beats L2, L3, L4.
func TestMergeMWAACascade_MandatoryPriorityOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		l1   cascade.MWAAKropathSection
		l2   cascade.MWAAKropathSection
		l3   cascade.MWAAConfigSection
		l4   cascade.MWAAConfigSection
		want string
	}{
		{
			name: "L1 wins",
			l1:   cascade.MWAAKropathSection{AirflowVersion: "2.10.0"},
			l2:   cascade.MWAAKropathSection{AirflowVersion: "2.9.0"},
			l3:   cascade.MWAAConfigSection{AirflowVersion: "2.8.0"},
			l4:   cascade.MWAAConfigSection{AirflowVersion: "2.7.0"},
			want: "2.10.0",
		},
		{
			name: "L2 wins when L1 empty",
			l1:   emptyMWAAKropath,
			l2:   cascade.MWAAKropathSection{AirflowVersion: "2.9.0"},
			l3:   cascade.MWAAConfigSection{AirflowVersion: "2.8.0"},
			l4:   cascade.MWAAConfigSection{AirflowVersion: "2.7.0"},
			want: "2.9.0",
		},
		{
			name: "L3 wins when L1,L2 empty",
			l1:   emptyMWAAKropath,
			l2:   emptyMWAAKropath,
			l3:   cascade.MWAAConfigSection{AirflowVersion: "2.8.0"},
			l4:   cascade.MWAAConfigSection{AirflowVersion: "2.7.0"},
			want: "2.8.0",
		},
		{
			name: "L4 fallback",
			l1:   emptyMWAAKropath,
			l2:   emptyMWAAKropath,
			l3:   emptyMWAACfg,
			l4:   cascade.MWAAConfigSection{AirflowVersion: "2.7.0"},
			want: "2.7.0",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := mergeMWAAAll(tc.l1, tc.l2, tc.l3, tc.l4, emptyMWAACfg, emptyMWAACfg, emptyMWAAKropath, emptyMWAAKropath)
			if got.Mandatory.AirflowVersion != tc.want {
				t.Errorf("mandatory.airflowVersion = %q; want %q", got.Mandatory.AirflowVersion, tc.want)
			}
		})
	}
}

// TestMergeMWAACascade_DefaultsPriorityOrder verifies the defaults tier priority:
// L6 (localMWAACfgDefaults) beats L7, L8, L9.
func TestMergeMWAACascade_DefaultsPriorityOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		l6   cascade.MWAAConfigSection
		l7   cascade.MWAAConfigSection
		l8   cascade.MWAAKropathSection
		l9   cascade.MWAAKropathSection
		want string
	}{
		{
			name: "L6 wins",
			l6:   cascade.MWAAConfigSection{EnvironmentClass: "mw1.small"},
			l7:   cascade.MWAAConfigSection{EnvironmentClass: "mw1.medium"},
			l8:   cascade.MWAAKropathSection{EnvironmentClass: "mw1.large"},
			l9:   cascade.MWAAKropathSection{EnvironmentClass: "mw1.xlarge"},
			want: "mw1.small",
		},
		{
			name: "L7 wins when L6 empty",
			l6:   emptyMWAACfg,
			l7:   cascade.MWAAConfigSection{EnvironmentClass: "mw1.medium"},
			l8:   cascade.MWAAKropathSection{EnvironmentClass: "mw1.large"},
			l9:   cascade.MWAAKropathSection{EnvironmentClass: "mw1.xlarge"},
			want: "mw1.medium",
		},
		{
			name: "L8 wins when L6,L7 empty",
			l6:   emptyMWAACfg,
			l7:   emptyMWAACfg,
			l8:   cascade.MWAAKropathSection{EnvironmentClass: "mw1.large"},
			l9:   cascade.MWAAKropathSection{EnvironmentClass: "mw1.xlarge"},
			want: "mw1.large",
		},
		{
			name: "L9 fallback",
			l6:   emptyMWAACfg,
			l7:   emptyMWAACfg,
			l8:   emptyMWAAKropath,
			l9:   cascade.MWAAKropathSection{EnvironmentClass: "mw1.xlarge"},
			want: "mw1.xlarge",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := mergeMWAAAll(emptyMWAAKropath, emptyMWAAKropath, emptyMWAACfg, emptyMWAACfg, tc.l6, tc.l7, tc.l8, tc.l9)
			if got.Defaults.EnvironmentClass != tc.want {
				t.Errorf("defaults.environmentClass = %q; want %q", got.Defaults.EnvironmentClass, tc.want)
			}
		})
	}
}

// TestMergeMWAACascade_BoolNilFallthrough verifies that a nil *bool at a higher-priority level
// falls through to a non-nil value at a lower-priority level.
func TestMergeMWAACascade_BoolNilFallthrough(t *testing.T) {
	t.Parallel()

	t.Run("mandatory nil L1 does not override non-nil L4", func(t *testing.T) {
		t.Parallel()
		enabled := true
		got := mergeMWAAAll(
			emptyMWAAKropath, // L1 — nil DagProcessingLogsEnabled
			emptyMWAAKropath, // L2 — nil
			emptyMWAACfg,     // L3 — nil
			cascade.MWAAConfigSection{DagProcessingLogsEnabled: &enabled}, // L4
			emptyMWAACfg, emptyMWAACfg, emptyMWAAKropath, emptyMWAAKropath,
		)
		if !boolPtrEq(got.Mandatory.DagProcessingLogsEnabled, boolPtr(true)) {
			t.Errorf("mandatory.dagProcessingLogsEnabled = %v; want true (L4 fallback)", got.Mandatory.DagProcessingLogsEnabled)
		}
	})

	t.Run("defaults nil L6 does not override non-nil L9", func(t *testing.T) {
		t.Parallel()
		enabled := false
		got := mergeMWAAAll(
			emptyMWAAKropath, emptyMWAAKropath, emptyMWAACfg, emptyMWAACfg,
			emptyMWAACfg, // L6 — nil SchedulerLogsEnabled
			emptyMWAACfg, // L7 — nil
			emptyMWAAKropath, // L8 — nil
			cascade.MWAAKropathSection{SchedulerLogsEnabled: &enabled}, // L9
		)
		if !boolPtrEq(got.Defaults.SchedulerLogsEnabled, boolPtr(false)) {
			t.Errorf("defaults.schedulerLogsEnabled = %v; want false (L9 fallback)", got.Defaults.SchedulerLogsEnabled)
		}
	})
}

// TestMergeMWAACascade_BoolFalseIsExplicit verifies that a false *bool at a higher-priority
// level wins over a non-nil true at a lower-priority level (false != nil).
func TestMergeMWAACascade_BoolFalseIsExplicit(t *testing.T) {
	t.Parallel()

	t.Run("mandatory L1 false beats L4 true", func(t *testing.T) {
		t.Parallel()
		enabled := true
		got := mergeMWAAAll(
			cascade.MWAAKropathSection{WorkerLogsEnabled: boolPtr(false)}, // L1 — explicit false
			emptyMWAAKropath,
			emptyMWAACfg,
			cascade.MWAAConfigSection{WorkerLogsEnabled: &enabled}, // L4 — true, should lose
			emptyMWAACfg, emptyMWAACfg, emptyMWAAKropath, emptyMWAAKropath,
		)
		if !boolPtrEq(got.Mandatory.WorkerLogsEnabled, boolPtr(false)) {
			t.Errorf("mandatory.workerLogsEnabled = %v; want false (L1 explicit false wins)", got.Mandatory.WorkerLogsEnabled)
		}
	})

	t.Run("defaults L6 false beats L9 true", func(t *testing.T) {
		t.Parallel()
		got := mergeMWAAAll(
			emptyMWAAKropath, emptyMWAAKropath, emptyMWAACfg, emptyMWAACfg,
			cascade.MWAAConfigSection{TaskLogsEnabled: boolPtr(false)},    // L6 — explicit false
			emptyMWAACfg,
			emptyMWAAKropath,
			cascade.MWAAKropathSection{TaskLogsEnabled: boolPtr(true)},    // L9 — should lose
		)
		if !boolPtrEq(got.Defaults.TaskLogsEnabled, boolPtr(false)) {
			t.Errorf("defaults.taskLogsEnabled = %v; want false (L6 explicit false wins)", got.Defaults.TaskLogsEnabled)
		}
	})
}

// TestMergeMWAACascade_Int64ZeroFallthrough verifies that a 0 int64 at a higher-priority level
// is treated as "not set" and falls through to a non-zero value at a lower-priority level.
func TestMergeMWAACascade_Int64ZeroFallthrough(t *testing.T) {
	t.Parallel()

	t.Run("mandatory L1=0 falls through to L4 non-zero", func(t *testing.T) {
		t.Parallel()
		got := mergeMWAAAll(
			cascade.MWAAKropathSection{MaxWorkers: 0}, // L1 — zero, skipped
			emptyMWAAKropath,
			emptyMWAACfg,
			cascade.MWAAConfigSection{MaxWorkers: 10}, // L4 — used
			emptyMWAACfg, emptyMWAACfg, emptyMWAAKropath, emptyMWAAKropath,
		)
		if got.Mandatory.MaxWorkers != 10 {
			t.Errorf("mandatory.maxWorkers = %d; want 10 (L4 fallback)", got.Mandatory.MaxWorkers)
		}
	})

	t.Run("defaults L6=0 falls through to L9 non-zero", func(t *testing.T) {
		t.Parallel()
		got := mergeMWAAAll(
			emptyMWAAKropath, emptyMWAAKropath, emptyMWAACfg, emptyMWAACfg,
			cascade.MWAAConfigSection{MinWorkers: 0}, // L6 — zero, skipped
			emptyMWAACfg,
			emptyMWAAKropath,
			cascade.MWAAKropathSection{MinWorkers: 3}, // L9 — used
		)
		if got.Defaults.MinWorkers != 3 {
			t.Errorf("defaults.minWorkers = %d; want 3 (L9 fallback)", got.Defaults.MinWorkers)
		}
	})

	t.Run("mandatory L1 non-zero wins over L4 non-zero", func(t *testing.T) {
		t.Parallel()
		got := mergeMWAAAll(
			cascade.MWAAKropathSection{Schedulers: 5}, // L1 — wins
			emptyMWAAKropath,
			emptyMWAACfg,
			cascade.MWAAConfigSection{Schedulers: 2}, // L4 — loses
			emptyMWAACfg, emptyMWAACfg, emptyMWAAKropath, emptyMWAAKropath,
		)
		if got.Mandatory.Schedulers != 5 {
			t.Errorf("mandatory.schedulers = %d; want 5 (L1 wins)", got.Mandatory.Schedulers)
		}
	})
}

// TestMergeMWAACascade_AirflowConfigOptionsMapUnion verifies the key-priority map merge for
// airflowConfigurationOptions: higher-priority level wins on key conflict; non-conflicting
// keys from all levels are included.
func TestMergeMWAACascade_AirflowConfigOptionsMapUnion(t *testing.T) {
	t.Parallel()

	t.Run("mandatory: L1 key wins over L4 on conflict; unique keys included", func(t *testing.T) {
		t.Parallel()
		got := mergeMWAAAll(
			cascade.MWAAKropathSection{AirflowConfigurationOptions: map[string]string{
				"core.parallelism": "32", // L1 — wins on conflict
				"core.dag_concurrency": "16",
			}},
			emptyMWAAKropath,
			emptyMWAACfg,
			cascade.MWAAConfigSection{AirflowConfigurationOptions: map[string]string{
				"core.parallelism":    "8",  // L4 — loses on conflict
				"webserver.workers":   "4",  // L4 — unique, included
			}},
			emptyMWAACfg, emptyMWAACfg, emptyMWAAKropath, emptyMWAAKropath,
		)
		opts := got.Mandatory.AirflowConfigurationOptions
		if opts["core.parallelism"] != "32" {
			t.Errorf("mandatory.airflowConfigurationOptions[core.parallelism] = %q; want 32", opts["core.parallelism"])
		}
		if opts["core.dag_concurrency"] != "16" {
			t.Errorf("mandatory.airflowConfigurationOptions[core.dag_concurrency] = %q; want 16", opts["core.dag_concurrency"])
		}
		if opts["webserver.workers"] != "4" {
			t.Errorf("mandatory.airflowConfigurationOptions[webserver.workers] = %q; want 4", opts["webserver.workers"])
		}
	})

	t.Run("defaults: L6 key wins over L9 on conflict", func(t *testing.T) {
		t.Parallel()
		got := mergeMWAAAll(
			emptyMWAAKropath, emptyMWAAKropath, emptyMWAACfg, emptyMWAACfg,
			cascade.MWAAConfigSection{AirflowConfigurationOptions: map[string]string{
				"scheduler.max_dagruns_to_create_per_loop": "10", // L6 — wins
			}},
			emptyMWAACfg,
			emptyMWAAKropath,
			cascade.MWAAKropathSection{AirflowConfigurationOptions: map[string]string{
				"scheduler.max_dagruns_to_create_per_loop": "5",  // L9 — loses
				"smtp.smtp_mail_from": "airflow@example.com",     // L9 unique
			}},
		)
		opts := got.Defaults.AirflowConfigurationOptions
		if opts["scheduler.max_dagruns_to_create_per_loop"] != "10" {
			t.Errorf("defaults.airflowConfigurationOptions[scheduler.max_dagruns_to_create_per_loop] = %q; want 10", opts["scheduler.max_dagruns_to_create_per_loop"])
		}
		if opts["smtp.smtp_mail_from"] != "airflow@example.com" {
			t.Errorf("defaults.airflowConfigurationOptions[smtp.smtp_mail_from] = %q; want airflow@example.com", opts["smtp.smtp_mail_from"])
		}
	})
}

// TestMergeMWAACascade_TagsUnion verifies that tags from all mandatory levels are merged;
// higher-priority level wins on key conflict.
func TestMergeMWAACascade_TagsUnion(t *testing.T) {
	t.Parallel()

	t.Run("mandatory tags union across all four levels", func(t *testing.T) {
		t.Parallel()
		got := mergeMWAAAll(
			cascade.MWAAKropathSection{Tags: map[string]string{"env": "prod", "owner": "platform"}},
			cascade.MWAAKropathSection{Tags: map[string]string{"team": "infra"}},
			cascade.MWAAConfigSection{Tags: map[string]string{"project": "airflow"}},
			cascade.MWAAConfigSection{Tags: map[string]string{"env": "dev", "cost-center": "eng"}}, // "env" loses to L1
			emptyMWAACfg, emptyMWAACfg, emptyMWAAKropath, emptyMWAAKropath,
		)
		tags := got.Mandatory.Tags
		if tags["env"] != "prod" {
			t.Errorf("mandatory.tags[env] = %q; want prod (L1 wins)", tags["env"])
		}
		if tags["owner"] != "platform" {
			t.Errorf("mandatory.tags[owner] = %q; want platform", tags["owner"])
		}
		if tags["team"] != "infra" {
			t.Errorf("mandatory.tags[team] = %q; want infra", tags["team"])
		}
		if tags["project"] != "airflow" {
			t.Errorf("mandatory.tags[project] = %q; want airflow", tags["project"])
		}
		if tags["cost-center"] != "eng" {
			t.Errorf("mandatory.tags[cost-center] = %q; want eng", tags["cost-center"])
		}
	})
}

// TestMergeMWAACascade_NamingTemplateOnlyFromMWAAConfigLevels verifies that namingTemplate is
// governed only by MWAAConfig levels (L3/L4 for mandatory, L6/L7 for defaults).
// KropathConfig sections do NOT carry namingTemplate.
func TestMergeMWAACascade_NamingTemplateOnlyFromMWAAConfigLevels(t *testing.T) {
	t.Parallel()

	t.Run("mandatory namingTemplate comes from L3, not KropathConfig levels", func(t *testing.T) {
		t.Parallel()
		got := mergeMWAAAll(
			emptyMWAAKropath,
			emptyMWAAKropath,
			cascade.MWAAConfigSection{NamingTemplate: "mwaa-{namespace}-{name}"}, // L3
			cascade.MWAAConfigSection{NamingTemplate: "mwaa-local-{name}"},       // L4
			emptyMWAACfg, emptyMWAACfg, emptyMWAAKropath, emptyMWAAKropath,
		)
		if got.Mandatory.NamingTemplate != "mwaa-{namespace}-{name}" {
			t.Errorf("mandatory.namingTemplate = %q; want mwaa-{namespace}-{name} (L3 wins)", got.Mandatory.NamingTemplate)
		}
	})

	t.Run("defaults namingTemplate: L6 wins over L7", func(t *testing.T) {
		t.Parallel()
		got := mergeMWAAAll(
			emptyMWAAKropath, emptyMWAAKropath, emptyMWAACfg, emptyMWAACfg,
			cascade.MWAAConfigSection{NamingTemplate: "mwaa-ns-{namespace}"},   // L6
			cascade.MWAAConfigSection{NamingTemplate: "mwaa-global-{name}"},    // L7
			emptyMWAAKropath, emptyMWAAKropath,
		)
		if got.Defaults.NamingTemplate != "mwaa-ns-{namespace}" {
			t.Errorf("defaults.namingTemplate = %q; want mwaa-ns-{namespace} (L6 wins)", got.Defaults.NamingTemplate)
		}
	})
}

// TestMergeMWAACascade_SyncedLabelsOnlyFromMWAAConfigLevels verifies that syncedLabels and
// syncedAnnotations are merged only from MWAAConfig levels.
func TestMergeMWAACascade_SyncedLabelsOnlyFromMWAAConfigLevels(t *testing.T) {
	t.Parallel()

	t.Run("mandatory syncedLabels: L3 wins over L4 on conflict", func(t *testing.T) {
		t.Parallel()
		got := mergeMWAAAll(
			emptyMWAAKropath,
			emptyMWAAKropath,
			cascade.MWAAConfigSection{SyncedLabels: map[string]string{"app": "airflow", "tier": "data"}}, // L3
			cascade.MWAAConfigSection{SyncedLabels: map[string]string{"app": "override", "ns": "prod"}},  // L4, "app" loses
			emptyMWAACfg, emptyMWAACfg, emptyMWAAKropath, emptyMWAAKropath,
		)
		labels := got.Mandatory.SyncedLabels
		if labels["app"] != "airflow" {
			t.Errorf("mandatory.syncedLabels[app] = %q; want airflow (L3 wins)", labels["app"])
		}
		if labels["tier"] != "data" {
			t.Errorf("mandatory.syncedLabels[tier] = %q; want data", labels["tier"])
		}
		if labels["ns"] != "prod" {
			t.Errorf("mandatory.syncedLabels[ns] = %q; want prod (L4 unique key included)", labels["ns"])
		}
	})

	t.Run("defaults syncedAnnotations: L6 wins over L7 on conflict", func(t *testing.T) {
		t.Parallel()
		got := mergeMWAAAll(
			emptyMWAAKropath, emptyMWAAKropath, emptyMWAACfg, emptyMWAACfg,
			cascade.MWAAConfigSection{SyncedAnnotations: map[string]string{"iam.role": "arn:aws:iam::123:role/a"}}, // L6
			cascade.MWAAConfigSection{SyncedAnnotations: map[string]string{"iam.role": "arn:aws:iam::123:role/b"}}, // L7, loses
			emptyMWAAKropath, emptyMWAAKropath,
		)
		annotations := got.Defaults.SyncedAnnotations
		if annotations["iam.role"] != "arn:aws:iam::123:role/a" {
			t.Errorf("defaults.syncedAnnotations[iam.role] = %q; want arn:aws:iam::123:role/a (L6 wins)", annotations["iam.role"])
		}
	})
}

// TestMergeMWAACascade_AllZero verifies that when all sources are zero-value the result
// is also a zero-value EffectiveMWAAConfig (no panics, no unexpected defaults).
func TestMergeMWAACascade_AllZero(t *testing.T) {
	t.Parallel()
	got := mergeMWAAAll(
		emptyMWAAKropath, emptyMWAAKropath,
		emptyMWAACfg, emptyMWAACfg, emptyMWAACfg, emptyMWAACfg,
		emptyMWAAKropath, emptyMWAAKropath,
	)
	if got.Mandatory.AirflowVersion != "" {
		t.Errorf("mandatory.airflowVersion = %q; want empty", got.Mandatory.AirflowVersion)
	}
	if got.Mandatory.MaxWorkers != 0 {
		t.Errorf("mandatory.maxWorkers = %d; want 0", got.Mandatory.MaxWorkers)
	}
	if got.Mandatory.DagProcessingLogsEnabled != nil {
		t.Errorf("mandatory.dagProcessingLogsEnabled = %v; want nil", got.Mandatory.DagProcessingLogsEnabled)
	}
	if got.Defaults.AirflowVersion != "" {
		t.Errorf("defaults.airflowVersion = %q; want empty", got.Defaults.AirflowVersion)
	}
	if got.Defaults.MinWorkers != 0 {
		t.Errorf("defaults.minWorkers = %d; want 0", got.Defaults.MinWorkers)
	}
	if got.Defaults.WorkerLogsEnabled != nil {
		t.Errorf("defaults.workerLogsEnabled = %v; want nil", got.Defaults.WorkerLogsEnabled)
	}
}

// TestMergeMWAACascade_MandatoryDefaultsIndependent verifies that setting a *bool field in
// mandatory does not bleed into defaults and vice versa.
func TestMergeMWAACascade_MandatoryDefaultsIndependent(t *testing.T) {
	t.Parallel()

	t.Run("mandatory value does not bleed into defaults", func(t *testing.T) {
		t.Parallel()
		got := mergeMWAAAll(
			cascade.MWAAKropathSection{WebserverLogsEnabled: boolPtr(true)}, // L1 mandatory
			emptyMWAAKropath, emptyMWAACfg, emptyMWAACfg,
			emptyMWAACfg, emptyMWAACfg, emptyMWAAKropath, emptyMWAAKropath,
		)
		if !boolPtrEq(got.Mandatory.WebserverLogsEnabled, boolPtr(true)) {
			t.Errorf("mandatory.webserverLogsEnabled = %v; want true", got.Mandatory.WebserverLogsEnabled)
		}
		if got.Defaults.WebserverLogsEnabled != nil {
			t.Errorf("defaults.webserverLogsEnabled = %v; must not bleed from mandatory", got.Defaults.WebserverLogsEnabled)
		}
	})
}
