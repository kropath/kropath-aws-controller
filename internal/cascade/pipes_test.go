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

var (
	emptyPipesKropath = cascade.PipesKropathSection{}
	emptyPipesCfg     = cascade.PipesConfigSection{}
)

func TestMergePipesCascade_GlobalKropathMandatoryWins(t *testing.T) {
	t.Parallel()

	got := cascade.MergePipesCascade(
		cascade.PipesKropathSection{DesiredState: "STOPPED"}, // L1 wins
		cascade.PipesKropathSection{DesiredState: "RUNNING"}, // L2
		cascade.PipesConfigSection{DesiredState: "RUNNING"},  // L3
		cascade.PipesConfigSection{DesiredState: "RUNNING"},  // L4
		emptyPipesCfg, emptyPipesCfg,
		emptyPipesKropath, emptyPipesKropath,
	)

	if got.Mandatory.DesiredState != "STOPPED" {
		t.Errorf("Mandatory.DesiredState = %q, want %q", got.Mandatory.DesiredState, "STOPPED")
	}
}

func TestMergePipesCascade_LocalPipesCfgMandatoryFallsThrough(t *testing.T) {
	t.Parallel()

	got := cascade.MergePipesCascade(
		emptyPipesKropath,                                    // L1 empty
		emptyPipesKropath,                                    // L2 empty
		emptyPipesCfg,                                        // L3 empty
		cascade.PipesConfigSection{DesiredState: "STOPPED"},  // L4 — fallback mandatory
		emptyPipesCfg, emptyPipesCfg,
		emptyPipesKropath, emptyPipesKropath,
	)

	if got.Mandatory.DesiredState != "STOPPED" {
		t.Errorf("Mandatory.DesiredState = %q, want %q", got.Mandatory.DesiredState, "STOPPED")
	}
	// Defaults should be empty.
	if got.Defaults.DesiredState != "" {
		t.Errorf("Defaults.DesiredState = %q, want empty", got.Defaults.DesiredState)
	}
}

func TestMergePipesCascade_DefaultsPriorityOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		l6   string
		l7   string
		l8   string
		l9   string
		want string
	}{
		{"L6 wins", "STOPPED", "RUNNING", "RUNNING", "RUNNING", "STOPPED"},
		{"L7 when L6 empty", "", "STOPPED", "RUNNING", "RUNNING", "STOPPED"},
		{"L8 when L6 L7 empty", "", "", "STOPPED", "RUNNING", "STOPPED"},
		{"L9 when all above empty", "", "", "", "STOPPED", "STOPPED"},
		{"all empty", "", "", "", "", ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := cascade.MergePipesCascade(
				emptyPipesKropath, emptyPipesKropath,
				emptyPipesCfg, emptyPipesCfg,
				cascade.PipesConfigSection{DesiredState: tt.l6},
				cascade.PipesConfigSection{DesiredState: tt.l7},
				cascade.PipesKropathSection{DesiredState: tt.l8},
				cascade.PipesKropathSection{DesiredState: tt.l9},
			)
			if got.Defaults.DesiredState != tt.want {
				t.Errorf("Defaults.DesiredState = %q, want %q", got.Defaults.DesiredState, tt.want)
			}
		})
	}
}

func TestMergePipesCascade_NamingTemplateNotInKropathConfig(t *testing.T) {
	t.Parallel()

	got := cascade.MergePipesCascade(
		cascade.PipesKropathSection{DesiredState: "RUNNING"}, // L1 — no NamingTemplate field
		emptyPipesKropath,
		cascade.PipesConfigSection{NamingTemplate: "{namespace}-{name}"},  // L3
		emptyPipesCfg,
		emptyPipesCfg, emptyPipesCfg,
		emptyPipesKropath, emptyPipesKropath,
	)

	if got.Mandatory.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("Mandatory.NamingTemplate = %q, want %q", got.Mandatory.NamingTemplate, "{namespace}-{name}")
	}

	got2 := cascade.MergePipesCascade(
		emptyPipesKropath, emptyPipesKropath,
		emptyPipesCfg, emptyPipesCfg,
		cascade.PipesConfigSection{NamingTemplate: "{namespace}-{name}"},  // L6
		emptyPipesCfg,
		cascade.PipesKropathSection{DesiredState: "STOPPED"}, // L8 — no NamingTemplate
		emptyPipesKropath,
	)

	if got2.Defaults.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("Defaults.NamingTemplate = %q, want %q", got2.Defaults.NamingTemplate, "{namespace}-{name}")
	}
}

func TestMergePipesCascade_TagsMergeMandatory(t *testing.T) {
	t.Parallel()

	got := cascade.MergePipesCascade(
		cascade.PipesKropathSection{Tags: map[string]string{"env": "prod", "owner": "platform"}}, // L1
		cascade.PipesKropathSection{Tags: map[string]string{"env": "staging", "team": "infra"}},  // L2
		cascade.PipesConfigSection{Tags: map[string]string{"cost-center": "cc1"}},                // L3
		cascade.PipesConfigSection{Tags: map[string]string{"cost-center": "cc2", "app": "myapp"}}, // L4
		emptyPipesCfg, emptyPipesCfg,
		emptyPipesKropath, emptyPipesKropath,
	)

	// L1 wins on "env" key conflict; all keys should be present.
	if got.Mandatory.Tags["env"] != "prod" {
		t.Errorf("Tags[env] = %q, want prod (L1 wins)", got.Mandatory.Tags["env"])
	}
	if got.Mandatory.Tags["owner"] != "platform" {
		t.Errorf("Tags[owner] = %q, want platform", got.Mandatory.Tags["owner"])
	}
	if got.Mandatory.Tags["team"] != "infra" {
		t.Errorf("Tags[team] = %q, want infra", got.Mandatory.Tags["team"])
	}
	if got.Mandatory.Tags["cost-center"] != "cc1" {
		t.Errorf("Tags[cost-center] = %q, want cc1 (L3 wins over L4)", got.Mandatory.Tags["cost-center"])
	}
	if got.Mandatory.Tags["app"] != "myapp" {
		t.Errorf("Tags[app] = %q, want myapp", got.Mandatory.Tags["app"])
	}
}

func TestMergePipesCascade_TagsMergeDefaults(t *testing.T) {
	t.Parallel()

	got := cascade.MergePipesCascade(
		emptyPipesKropath, emptyPipesKropath,
		emptyPipesCfg, emptyPipesCfg,
		cascade.PipesConfigSection{Tags: map[string]string{"env": "dev", "team": "pipes"}}, // L6
		cascade.PipesConfigSection{Tags: map[string]string{"env": "staging"}},              // L7
		cascade.PipesKropathSection{Tags: map[string]string{"org": "acme"}},                // L8
		cascade.PipesKropathSection{Tags: map[string]string{"org": "global"}},              // L9
	)

	// L6 wins on "env" conflict; L8 wins on "org" conflict over L9.
	if got.Defaults.Tags["env"] != "dev" {
		t.Errorf("Tags[env] = %q, want dev (L6 wins)", got.Defaults.Tags["env"])
	}
	if got.Defaults.Tags["team"] != "pipes" {
		t.Errorf("Tags[team] = %q, want pipes", got.Defaults.Tags["team"])
	}
	if got.Defaults.Tags["org"] != "acme" {
		t.Errorf("Tags[org] = %q, want acme (L8 wins over L9)", got.Defaults.Tags["org"])
	}
}

func TestMergePipesCascade_SyncedLabelsAndAnnotations(t *testing.T) {
	t.Parallel()

	got := cascade.MergePipesCascade(
		emptyPipesKropath, emptyPipesKropath,
		cascade.PipesConfigSection{
			SyncedLabels:      map[string]string{"label": "global"},
			SyncedAnnotations: map[string]string{"ann": "global"},
		}, // L3
		cascade.PipesConfigSection{
			SyncedLabels:      map[string]string{"label": "local", "extra": "yes"},
			SyncedAnnotations: map[string]string{"ann": "local", "extra": "ann"},
		}, // L4
		emptyPipesCfg, emptyPipesCfg,
		emptyPipesKropath, emptyPipesKropath,
	)

	// L3 wins on "label" conflict (lower level number wins).
	if got.Mandatory.SyncedLabels["label"] != "global" {
		t.Errorf("SyncedLabels[label] = %q, want global (L3 wins)", got.Mandatory.SyncedLabels["label"])
	}
	if got.Mandatory.SyncedLabels["extra"] != "yes" {
		t.Errorf("SyncedLabels[extra] = %q, want yes", got.Mandatory.SyncedLabels["extra"])
	}
	if got.Mandatory.SyncedAnnotations["ann"] != "global" {
		t.Errorf("SyncedAnnotations[ann] = %q, want global (L3 wins)", got.Mandatory.SyncedAnnotations["ann"])
	}
}

func TestMergePipesCascade_KropathMandatorySyncedLabelsWin(t *testing.T) {
	t.Parallel()

	got := cascade.MergePipesCascade(
		cascade.PipesKropathSection{SyncedLabels: map[string]string{"env": "prod", "owner": "platform"}}, // L1
		cascade.PipesKropathSection{SyncedLabels: map[string]string{"env": "staging", "team": "infra"}},  // L2
		cascade.PipesConfigSection{SyncedLabels: map[string]string{"cost-center": "cc1"}},                // L3
		cascade.PipesConfigSection{SyncedLabels: map[string]string{"cost-center": "cc2", "app": "myapp"}}, // L4
		emptyPipesCfg, emptyPipesCfg,
		emptyPipesKropath, emptyPipesKropath,
	)

	// L1 wins on "env"; all unique keys should be present.
	if got.Mandatory.SyncedLabels["env"] != "prod" {
		t.Errorf("SyncedLabels[env] = %q, want prod (L1 wins)", got.Mandatory.SyncedLabels["env"])
	}
	if got.Mandatory.SyncedLabels["owner"] != "platform" {
		t.Errorf("SyncedLabels[owner] = %q, want platform", got.Mandatory.SyncedLabels["owner"])
	}
	if got.Mandatory.SyncedLabels["team"] != "infra" {
		t.Errorf("SyncedLabels[team] = %q, want infra", got.Mandatory.SyncedLabels["team"])
	}
	if got.Mandatory.SyncedLabels["cost-center"] != "cc1" {
		t.Errorf("SyncedLabels[cost-center] = %q, want cc1 (L3 wins over L4)", got.Mandatory.SyncedLabels["cost-center"])
	}
	if got.Mandatory.SyncedLabels["app"] != "myapp" {
		t.Errorf("SyncedLabels[app] = %q, want myapp", got.Mandatory.SyncedLabels["app"])
	}
}

func TestMergePipesCascade_KropathMandatorySyncedAnnotationsWin(t *testing.T) {
	t.Parallel()

	got := cascade.MergePipesCascade(
		cascade.PipesKropathSection{SyncedAnnotations: map[string]string{"iam.role": "arn:global", "owner": "platform"}}, // L1
		cascade.PipesKropathSection{SyncedAnnotations: map[string]string{"iam.role": "arn:local"}},                       // L2
		cascade.PipesConfigSection{SyncedAnnotations: map[string]string{"cost-center": "cc1"}},                           // L3
		emptyPipesCfg,
		emptyPipesCfg, emptyPipesCfg,
		emptyPipesKropath, emptyPipesKropath,
	)

	// L1 wins on "iam.role" conflict.
	if got.Mandatory.SyncedAnnotations["iam.role"] != "arn:global" {
		t.Errorf("SyncedAnnotations[iam.role] = %q, want arn:global (L1 wins)", got.Mandatory.SyncedAnnotations["iam.role"])
	}
	if got.Mandatory.SyncedAnnotations["owner"] != "platform" {
		t.Errorf("SyncedAnnotations[owner] = %q, want platform", got.Mandatory.SyncedAnnotations["owner"])
	}
	if got.Mandatory.SyncedAnnotations["cost-center"] != "cc1" {
		t.Errorf("SyncedAnnotations[cost-center] = %q, want cc1", got.Mandatory.SyncedAnnotations["cost-center"])
	}
}

func TestMergePipesCascade_KropathDefaultsSyncedLabelsFallthrough(t *testing.T) {
	t.Parallel()

	got := cascade.MergePipesCascade(
		emptyPipesKropath, emptyPipesKropath,
		emptyPipesCfg, emptyPipesCfg,
		cascade.PipesConfigSection{SyncedLabels: map[string]string{"env": "dev", "team": "pipes"}}, // L6
		cascade.PipesConfigSection{SyncedLabels: map[string]string{"env": "staging"}},              // L7
		cascade.PipesKropathSection{SyncedLabels: map[string]string{"org": "acme"}},                // L8
		cascade.PipesKropathSection{SyncedLabels: map[string]string{"org": "global", "tier": "free"}}, // L9
	)

	// L6 wins on "env"; L8 wins on "org"; L9 contributes unique "tier".
	if got.Defaults.SyncedLabels["env"] != "dev" {
		t.Errorf("SyncedLabels[env] = %q, want dev (L6 wins)", got.Defaults.SyncedLabels["env"])
	}
	if got.Defaults.SyncedLabels["team"] != "pipes" {
		t.Errorf("SyncedLabels[team] = %q, want pipes", got.Defaults.SyncedLabels["team"])
	}
	if got.Defaults.SyncedLabels["org"] != "acme" {
		t.Errorf("SyncedLabels[org] = %q, want acme (L8 wins over L9)", got.Defaults.SyncedLabels["org"])
	}
	if got.Defaults.SyncedLabels["tier"] != "free" {
		t.Errorf("SyncedLabels[tier] = %q, want free (L9 contributes)", got.Defaults.SyncedLabels["tier"])
	}
}

func TestMergePipesCascade_KropathDefaultsSyncedAnnotationsFallthrough(t *testing.T) {
	t.Parallel()

	got := cascade.MergePipesCascade(
		emptyPipesKropath, emptyPipesKropath,
		emptyPipesCfg, emptyPipesCfg,
		emptyPipesCfg,
		emptyPipesCfg,
		cascade.PipesKropathSection{SyncedAnnotations: map[string]string{"iam.role": "arn:ns"}},      // L8
		cascade.PipesKropathSection{SyncedAnnotations: map[string]string{"iam.role": "arn:global"}},  // L9
	)

	// L8 wins over L9 on "iam.role".
	if got.Defaults.SyncedAnnotations["iam.role"] != "arn:ns" {
		t.Errorf("SyncedAnnotations[iam.role] = %q, want arn:ns (L8 wins over L9)", got.Defaults.SyncedAnnotations["iam.role"])
	}
}

func TestMergePipesCascade_AllFieldsEmpty(t *testing.T) {
	t.Parallel()

	got := cascade.MergePipesCascade(
		emptyPipesKropath, emptyPipesKropath,
		emptyPipesCfg, emptyPipesCfg,
		emptyPipesCfg, emptyPipesCfg,
		emptyPipesKropath, emptyPipesKropath,
	)

	if got.Mandatory.DesiredState != "" {
		t.Errorf("Mandatory.DesiredState = %q, want empty", got.Mandatory.DesiredState)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("Mandatory.NamingTemplate = %q, want empty", got.Mandatory.NamingTemplate)
	}
	if len(got.Mandatory.Tags) != 0 {
		t.Errorf("Mandatory.Tags = %v, want empty", got.Mandatory.Tags)
	}
	if got.Defaults.DesiredState != "" {
		t.Errorf("Defaults.DesiredState = %q, want empty", got.Defaults.DesiredState)
	}
	if len(got.Defaults.Tags) != 0 {
		t.Errorf("Defaults.Tags = %v, want empty", got.Defaults.Tags)
	}
}
