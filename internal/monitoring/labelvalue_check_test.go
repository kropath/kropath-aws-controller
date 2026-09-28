// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package monitoring

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot walks up from the package directory to the repository root, so the
// test works whether `go test` runs from the package directory or the module
// root, and does not depend on the caller's working directory.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repository root (no go.mod found)")
		}
		dir = parent
	}
}

// TestLabelValueCheck_ShippedArtifactsPass is the §10.2.1 CI gate itself: every
// label matcher literal in the real config/monitoring/rules.yaml and dashboard
// JSON must resolve to a member of the closed-set union. A reason, outcome,
// trigger, field or kind rename that forgets to update rules.yaml or the
// dashboard fails here.
func TestLabelValueCheck_ShippedArtifactsPass(t *testing.T) {
	root := repoRoot(t)
	rulesPath := filepath.Join(root, "config", "monitoring", "rules.yaml")
	dashboardPath := filepath.Join(root, "config", "monitoring", "dashboards", "kropath-controller.json")

	violations, err := CheckAll(rulesPath, dashboardPath)
	if err != nil {
		t.Fatalf("CheckAll: %v", err)
	}
	for _, v := range violations {
		t.Errorf("%s", v)
	}
}

// TestClosedLabelValues_NonEmpty guards against a wiring mistake (e.g. a typo
// in a package import) silently producing an empty union, which would make
// TestLabelValueCheck_ShippedArtifactsPass pass for the wrong reason — every
// literal trivially "resolves" against nothing meaningful.
func TestClosedLabelValues_NonEmpty(t *testing.T) {
	closed := ClosedLabelValues()
	// 8 PlacementReasons + 3 ConfigProfileReasons + 4 TierReasons + 7 ReadyReasons
	// + the internal/metrics outcome/kind/field sets is comfortably more than 20
	// distinct values; this is a floor, not an exact count.
	const minExpected = 20
	if len(closed) < minExpected {
		t.Fatalf("ClosedLabelValues() returned %d values, want at least %d — a closed set is probably not wired in", len(closed), minExpected)
	}
}

// TestCheckExpr_DetectsUnknownLiteral is the unit-test half of AC-18's
// "renaming any reason const in a scratch commit must make it fail": it
// proves the check actually rejects a literal outside the closed set, using a
// synthetic expression rather than mutating the real rules.yaml.
func TestCheckExpr_DetectsUnknownLiteral(t *testing.T) {
	closed := map[string]bool{"Unreferenced": true}

	violations, err := CheckExpr("test", `max(kropath_kropathconfigstatus_configs{reason="TotallyRenamed"}) > 0`, closed)
	if err != nil {
		t.Fatalf("CheckExpr: %v", err)
	}
	if len(violations) != 1 {
		t.Fatalf("got %d violations, want 1: %v", len(violations), violations)
	}
	if violations[0].Label != "reason" || violations[0].Value != "TotallyRenamed" {
		t.Fatalf("unexpected violation: %+v", violations[0])
	}
}

// TestCheckExpr_KnownLiteralPasses is the mirror case: a literal that is a
// member of the closed set produces no violation.
func TestCheckExpr_KnownLiteralPasses(t *testing.T) {
	closed := map[string]bool{"Unreferenced": true}

	violations, err := CheckExpr("test", `max(kropath_kropathconfigstatus_configs{reason="Unreferenced"}) > 0`, closed)
	if err != nil {
		t.Fatalf("CheckExpr: %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("got %d violations, want 0: %v", len(violations), violations)
	}
}

// TestCheckExpr_RegexAlternatesEachChecked mirrors the real
// KropathPlacementResolutionFailing rule's `reason!~"A|B"` shape: every
// alternate is checked independently, and one bad alternate among several
// good ones is still reported.
func TestCheckExpr_RegexAlternatesEachChecked(t *testing.T) {
	closed := map[string]bool{"PlacementResolved": true, "ResolvedFromNamespace": true}

	violations, err := CheckExpr("test", `sum(rate(kropath_placement_resolutions_total{reason!~"PlacementResolved|ResolvedFromNamespace"}[15m])) > 0`, closed)
	if err != nil {
		t.Fatalf("CheckExpr: %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("got %d violations for an all-good regex, want 0: %v", len(violations), violations)
	}

	violations, err = CheckExpr("test", `sum(rate(kropath_placement_resolutions_total{reason!~"PlacementResolved|Renamed"}[15m])) > 0`, closed)
	if err != nil {
		t.Fatalf("CheckExpr: %v", err)
	}
	if len(violations) != 1 || violations[0].Value != "Renamed" {
		t.Fatalf("got violations %v, want exactly one flagging \"Renamed\"", violations)
	}
}

// TestCheckExpr_UndecomposableRegexFails is the fail-closed case spec §10.2.1
// requires: a regex matcher that is not plain-literal alternation must be
// reported as a violation rather than silently passed.
func TestCheckExpr_UndecomposableRegexFails(t *testing.T) {
	closed := map[string]bool{"DocumentResolved": true}

	violations, err := CheckExpr("test", `max(kropath_policydocument_documents{reason=~"Document.*"}) > 0`, closed)
	if err != nil {
		t.Fatalf("CheckExpr: %v", err)
	}
	if len(violations) != 1 {
		t.Fatalf("got %d violations, want 1 for an undecomposable regex: %v", len(violations), violations)
	}
}

// TestViolation_String verifies the human-readable form used when
// TestLabelValueCheck_ShippedArtifactsPass reports a real violation via t.Errorf.
func TestViolation_String(t *testing.T) {
	v := Violation{Source: "config/monitoring/rules.yaml: alert Foo", Metric: "kropath_x", Label: "reason", Value: "Bogus"}
	got := v.String()
	for _, want := range []string{"config/monitoring/rules.yaml: alert Foo", "kropath_x", "reason", "Bogus"} {
		if !strings.Contains(got, want) {
			t.Errorf("String() = %q, want it to contain %q", got, want)
		}
	}
}

// TestCheckExpr_ExemptPackageLabelSkipped guards the one documented exemption
// (design §10.2.1): `package` on kropath_registry_* is code-derived from
// features.All and is never checked against a closed set.
func TestCheckExpr_ExemptPackageLabelSkipped(t *testing.T) {
	closed := map[string]bool{}

	violations, err := CheckExpr("test", `min by (package) (kropath_registry_reconciler_pending_since_timestamp_seconds{package="s3config"})`, closed)
	if err != nil {
		t.Fatalf("CheckExpr: %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("got %d violations, want 0 (package is exempt): %v", len(violations), violations)
	}
}
