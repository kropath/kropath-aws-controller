// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"strings"
	"testing"

	"github.com/kropath/kropath-aws-controller/internal/features"
)

// TestFamilyConfigKindsMatchesFeatures keeps the familyConfigKinds literal
// (which internal/metrics cannot derive from features.All at runtime -- see
// family_kinds.go) in sync with the actual cascade registry: a family added
// to features.All without a matching entry here would silently drop out of
// the withheld-config and family-kinds-unavailable metrics.
func TestFamilyConfigKindsMatchesFeatures(t *testing.T) {
	want := map[string]string{} // kind -> family
	for _, r := range features.All {
		if r.Name == "KropathConfig" || !strings.HasSuffix(r.Name, "Config") {
			continue
		}
		want[r.Name] = r.Package
	}

	got := map[string]string{}
	for _, fk := range familyConfigKinds {
		got[fk.kind] = fk.family
	}

	if len(got) != len(want) {
		t.Fatalf("familyConfigKinds has %d entries, features.All has %d cascade entries", len(got), len(want))
	}
	for kind, family := range want {
		gotFamily, ok := got[kind]
		if !ok {
			t.Errorf("familyConfigKinds is missing kind %q (family %q)", kind, family)
			continue
		}
		if gotFamily != family {
			t.Errorf("familyConfigKinds[%q].family = %q, want %q", kind, gotFamily, family)
		}
	}
}

func TestFamilyConfigKindsNewListReturnsDistinctInstances(t *testing.T) {
	for _, fk := range familyConfigKinds {
		a := fk.newList()
		b := fk.newList()
		if a == b {
			t.Errorf("%s: newList() returned the same instance twice", fk.kind)
		}
	}
}
