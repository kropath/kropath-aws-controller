// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package monitoring

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRulesFile_MissingFile(t *testing.T) {
	if _, err := LoadRulesFile("/nonexistent/rules.yaml"); err == nil {
		t.Fatal("expected an error for a missing file, got nil")
	}
}

func TestLoadRulesFile_InvalidYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rules.yaml")
	if err := os.WriteFile(path, []byte("groups: [not: valid: yaml"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if _, err := LoadRulesFile(path); err == nil {
		t.Fatal("expected a parse error for invalid YAML, got nil")
	}
}

func TestLoadRulesFile_ParsesGroupsAndRules(t *testing.T) {
	root := repoRoot(t)
	rf, err := LoadRulesFile(filepath.Join(root, "config", "monitoring", "rules.yaml"))
	if err != nil {
		t.Fatalf("LoadRulesFile: %v", err)
	}
	// The 7 groups from spec §4: governance, placement, policydocument,
	// labelinjection, cascade, registry, meta.
	if len(rf.Groups) != 7 {
		t.Fatalf("got %d groups, want 7", len(rf.Groups))
	}
	for _, g := range rf.Groups {
		if len(g.Rules) == 0 {
			t.Errorf("group %q has no rules", g.Name)
		}
		for _, r := range g.Rules {
			if r.Labels.Severity == "" || r.Labels.Component != "kropath-controller" {
				t.Errorf("rule %q missing required labels: %+v", r.Alert, r.Labels)
			}
		}
	}
}
