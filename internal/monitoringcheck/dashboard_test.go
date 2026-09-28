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

// Package monitoringcheck validates the config/monitoring/ artifacts this
// repo ships (design §10, spec §4). It lives outside internal/metrics
// because internal/metrics must not import internal/reconciler/util (or any
// reconciler package) -- util imports metrics, and the reverse would cycle
// (see internal/metrics/tier_configs_collector.go's doc comment). This
// package has no such constraint: neither metrics nor util imports it.
//
// This file covers only the dashboard (spec §4.1, KRO-1283/I-6). The
// canonical §10.2.1 label-value check -- a promql/parser-based Go test
// unioning every reason/outcome/group/kind closed set and validating both
// config/monitoring/rules.yaml and this dashboard -- is KRO-1282/I-5's
// deliverable. Per the two tickets' coordination note, whichever of I-5/I-6
// merges second extends the shared check to also cover the other's artifact;
// this file is I-6's half, landed first.
package monitoringcheck

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/kropath/kropath-controller/internal/reconciler/util"
)

const dashboardPath = "../../config/monitoring/dashboards/kropath-controller.json"

type dashboardTarget struct {
	Expr string `json:"expr"`
}

type dashboardPanel struct {
	Title   string            `json:"title"`
	Targets []dashboardTarget `json:"targets"`
}

type dashboard struct {
	Title  string           `json:"title"`
	Panels []dashboardPanel `json:"panels"`
}

// wantPanels is the spec §4.1 table verbatim: title -> exact PromQL
// expression. Every gauge query uses max by per M-15, and the one label
// literal in the set (reason="ProfileFallthrough") is asserted against
// util.ConfigProfileReasons() below rather than hardcoded twice.
var wantPanels = map[string]string{
	"Config profile fallthrough by family": `sum by (family) (rate(kropath_config_profile_resolutions_total{reason="ProfileFallthrough"}[15m]))`,
	"Unresolved policy refs":               `max by (kind, field) (kropath_policydocument_unresolved_refs)`,
	"Sid conflict rate":                    `sum(rate(kropath_policydocument_sid_conflicts_total[15m]))`,
	"Label patch outcomes":                 `sum by (group, outcome) (rate(kropath_labeloperator_patches_total[15m]))`,
	"Group discovery outcomes":             `sum by (group, outcome) (rate(kropath_labeloperator_group_discovery_total[15m]))`,
	"Placement verdicts":                   `max by (status) (kropath_namespaceplacement_namespaces)`,
	"Placement transitions":                `sum by (to) (rate(kropath_namespaceplacement_transitions_total[1h]))`,
	"CRD watch events":                     `sum by (outcome) (rate(kropath_registry_crd_watch_events_total[15m]))`,
	"Optional kinds attached":              `max by (package) (kropath_registry_optional_kinds_attached)`,
}

func loadDashboard(t *testing.T) dashboard {
	t.Helper()
	path, err := filepath.Abs(dashboardPath)
	if err != nil {
		t.Fatalf("resolving dashboard path: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading dashboard JSON: %v", err)
	}
	var d dashboard
	if err := json.Unmarshal(data, &d); err != nil {
		t.Fatalf("dashboard JSON does not parse: %v", err)
	}
	return d
}

func TestDashboard_ValidJSONWithExpectedPanelCount(t *testing.T) {
	d := loadDashboard(t)
	if len(d.Panels) != len(wantPanels) {
		t.Fatalf("panel count = %d, want %d (spec §4.1)", len(d.Panels), len(wantPanels))
	}
}

// TestDashboard_EveryPanelMatchesSpec is the panel-expression name check
// (spec §4.1, §10.2): every panel's query is asserted verbatim against the
// spec table, which catches a metric or label typo the same way the full
// §10.2.1 label-value check would once it also covers this file (I-5).
func TestDashboard_EveryPanelMatchesSpec(t *testing.T) {
	d := loadDashboard(t)

	seen := make(map[string]bool, len(d.Panels))
	for _, p := range d.Panels {
		wantExpr, ok := wantPanels[p.Title]
		if !ok {
			t.Errorf("panel %q is not in the spec §4.1 table", p.Title)
			continue
		}
		seen[p.Title] = true

		if len(p.Targets) != 1 {
			t.Errorf("panel %q has %d targets, want 1", p.Title, len(p.Targets))
			continue
		}
		if got := p.Targets[0].Expr; got != wantExpr {
			t.Errorf("panel %q expr = %q, want %q", p.Title, got, wantExpr)
		}
	}

	for title := range wantPanels {
		if !seen[title] {
			t.Errorf("spec §4.1 panel %q is missing from the dashboard", title)
		}
	}
}

// TestDashboard_LabelLiteralIsInClosedSet is the one label-value literal the
// dashboard carries (spec §2.5, §10.2.1 prerequisite): reason="ProfileFallthrough"
// on the config-profile panel must be a member of util.ConfigProfileReasons(),
// the exported closed set, not a value that has drifted from the code.
func TestDashboard_LabelLiteralIsInClosedSet(t *testing.T) {
	const literal = "ProfileFallthrough"
	if !slices.Contains(util.ConfigProfileReasons(), literal) {
		t.Errorf("dashboard label literal %q is not in util.ConfigProfileReasons() = %v", literal, util.ConfigProfileReasons())
	}
}
