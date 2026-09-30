// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package features_test

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/kropath/kropath-aws-controller/internal/features"
)

// awsKropathGroup is the API group this gate is scoped to. Non-AWS groups (e.g.
// gcp.kropath.run) are intentionally ignored — the label operator watches those
// without a per-kind cascade reconciler, which is by design.
const awsKropathGroup = "aws.kropath.run"

// configKindSetAllowlist contains aws.kropath.run *Config kinds that intentionally
// have no cascade reconciler yet. Each entry cites the ticket that will remove it.
// When a reconciler is added for a kind, delete its line here.
var configKindSetAllowlist = map[string]string{}

// extractDefaultsFromSchema recursively walks a JSON-schema node and returns a flat
// map from dot-separated path (e.g. "mandatory.blockPublicAccess") to the value of
// any "default" key at that node. It descends into properties, items, and
// additionalProperties so that nested-object, array, and map-typed fields are all
// covered.
//
// The prefix is the path of the current node; pass "" for the root.
func extractDefaultsFromSchema(node map[string]interface{}, prefix string) map[string]interface{} {
	result := make(map[string]interface{})
	if def, ok := node["default"]; ok {
		result[prefix] = def
	}
	if props, ok := node["properties"].(map[string]interface{}); ok {
		for k, v := range props {
			childPath := k
			if prefix != "" {
				childPath = prefix + "." + k
			}
			if child, ok := v.(map[string]interface{}); ok {
				for p, d := range extractDefaultsFromSchema(child, childPath) {
					result[p] = d
				}
			}
		}
	}
	// array element schema
	if items, ok := node["items"].(map[string]interface{}); ok {
		itemPath := prefix + "[]"
		for p, d := range extractDefaultsFromSchema(items, itemPath) {
			result[p] = d
		}
	}
	// map value schema
	if addProps, ok := node["additionalProperties"].(map[string]interface{}); ok {
		addPath := prefix + "{}"
		for p, d := range extractDefaultsFromSchema(addProps, addPath) {
			result[p] = d
		}
	}
	return result
}

// driftCRD holds the parsed fields needed for the drift checks from one CRD YAML document.
type driftCRD struct {
	metaName string                 // e.g. "s3configs.aws.kropath.run"
	kind     string                 // e.g. "S3Config"
	group    string                 // e.g. "aws.kropath.run"
	spec     map[string]interface{} // openAPIV3Schema.properties.spec from storage version, may be nil
}

// parseDriftCRD reads path and returns the first CRD document in it, or nil if the
// file contains no CRD.
func parseDriftCRD(path string) (*driftCRD, error) {
	data, err := os.ReadFile(path) //nolint:gosec // test-only read of an operator-supplied path
	if err != nil {
		return nil, err
	}
	for _, doc := range strings.Split(string(data), "\n---") {
		if !crdDocRE.MatchString(doc) {
			continue
		}
		var raw map[string]interface{}
		if err := yaml.Unmarshal([]byte(doc), &raw); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}
		meta, _ := raw["metadata"].(map[string]interface{})
		metaName, _ := meta["name"].(string)

		crdSpec, _ := raw["spec"].(map[string]interface{})
		group, _ := crdSpec["group"].(string)
		names, _ := crdSpec["names"].(map[string]interface{})
		kind, _ := names["kind"].(string)

		var specSchema map[string]interface{}
		if versions, ok := crdSpec["versions"].([]interface{}); ok {
			for _, v := range versions {
				ver, _ := v.(map[string]interface{})
				if storage, _ := ver["storage"].(bool); !storage {
					continue
				}
				schema, _ := ver["schema"].(map[string]interface{})
				openAPI, _ := schema["openAPIV3Schema"].(map[string]interface{})
				props, _ := openAPI["properties"].(map[string]interface{})
				specSchema, _ = props["spec"].(map[string]interface{})
				break
			}
		}
		return &driftCRD{metaName: metaName, kind: kind, group: group, spec: specSchema}, nil
	}
	return nil, nil
}

// configKindsFromCRDDir walks dir recursively and returns a map from kind name to
// metadata.name for every aws.kropath.run *Config CRD found. It handles both the
// flat layout produced by `make crds-verify` (all files in one directory) and a local
// kropath-aws/crds/ tree with subdirectories like policy/.
func configKindsFromCRDDir(dir string) (map[string]string, error) {
	result := make(map[string]string)
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		if d.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		crd, err := parseDriftCRD(path)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}
		if crd == nil || crd.group != awsKropathGroup || !strings.HasSuffix(crd.kind, "Config") {
			return nil
		}
		result[crd.kind] = crd.metaName
		return nil
	})
	return result, err
}

// TestSpecDefaultDrift verifies that every aws.kropath.run *Config CRD present in
// both the controller fixture directories and kropath-aws has identical "default"
// values at every recursive path in its spec schema (including items and
// additionalProperties). A mismatch means the fixture drifted from the authoritative
// upstream CRD, which silently changes kube-apiserver defaulting behaviour in the
// test cluster while leaving the integration cluster unaffected.
//
// The test is skipped unless KROPATH_AWS_CRDS_DIR is set; `make crds-verify`
// fetches the upstream CRDs and sets it.
func TestSpecDefaultDrift(t *testing.T) {
	dir := os.Getenv("KROPATH_AWS_CRDS_DIR")
	if dir == "" {
		t.Skip("KROPATH_AWS_CRDS_DIR not set — run `make crds-verify` to check against kropath-aws")
	}

	// Index upstream CRDs by metadata.name.
	upstreamByMeta := map[string]*driftCRD{}
	if err := filepath.WalkDir(dir, func(path string, d os.DirEntry, werr error) error {
		if werr != nil || d.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return werr
		}
		crd, err := parseDriftCRD(path)
		if err != nil {
			return fmt.Errorf("parsing upstream %s: %w", path, err)
		}
		if crd != nil && crd.metaName != "" {
			upstreamByMeta[crd.metaName] = crd
		}
		return nil
	}); err != nil {
		t.Fatalf("walking upstream CRD dir %s: %v", dir, err)
	}
	if len(upstreamByMeta) == 0 {
		t.Fatalf("no CRDs found under KROPATH_AWS_CRDS_DIR=%s — is the path correct?", dir)
	}

	fixtures := gatherCRDFixtures(t)
	for _, fixPath := range fixtures {
		fix, err := parseDriftCRD(fixPath)
		if err != nil {
			t.Fatalf("parsing fixture %s: %v", fixPath, err)
		}
		if fix == nil || fix.group != awsKropathGroup || !strings.HasSuffix(fix.kind, "Config") {
			continue
		}
		upstream, ok := upstreamByMeta[fix.metaName]
		if !ok {
			// No upstream counterpart found — TestWatchedKindsMatchUpstreamCRDs covers this.
			continue
		}
		if fix.spec == nil || upstream.spec == nil {
			continue
		}

		fixDefs := extractDefaultsFromSchema(fix.spec, "")
		upDefs := extractDefaultsFromSchema(upstream.spec, "")

		// Report paths present in upstream but absent or different in fixture.
		for path, upVal := range upDefs {
			fixVal, present := fixDefs[path]
			if !present {
				t.Errorf("%s: path %q has default %v in kropath-aws but is absent from fixture.\n"+
					"Run `make crds-verify` locally to see diffs, then copy the CRD from kropath-aws/crds/.",
					fix.kind, path, upVal)
				continue
			}
			if !reflect.DeepEqual(fixVal, upVal) {
				t.Errorf("%s: default mismatch at path %q — fixture=%v upstream=%v.\n"+
					"The fixture CRD drifted from the authoritative CRD in kropath-aws. "+
					"Copy the updated CRD into tests/fixtures/crds/ to fix.",
					fix.kind, path, fixVal, upVal)
			}
		}
		// Report paths in fixture not present in upstream (fixture added a default that upstream lacks).
		for path, fixVal := range fixDefs {
			if _, present := upDefs[path]; !present {
				t.Errorf("%s: path %q has default %v in fixture but is absent from kropath-aws.\n"+
					"The fixture CRD has a default that was removed or never existed upstream.",
					fix.kind, path, fixVal)
			}
		}
	}
}

// TestConfigKindSetDrift verifies that every aws.kropath.run *Config kind in
// kropath-aws has a controller origin in features.All (i.e. appears in at least one
// reconciler's Kinds list), and vice versa. Intentional gaps are listed in
// configKindSetAllowlist with a tracking ticket.
//
// The test is skipped unless KROPATH_AWS_CRDS_DIR is set.
func TestConfigKindSetDrift(t *testing.T) {
	dir := os.Getenv("KROPATH_AWS_CRDS_DIR")
	if dir == "" {
		t.Skip("KROPATH_AWS_CRDS_DIR not set — run `make crds-verify` to check against kropath-aws")
	}

	upstreamKinds, err := configKindsFromCRDDir(dir)
	if err != nil {
		t.Fatalf("scanning upstream CRD dir %s: %v", dir, err)
	}
	if len(upstreamKinds) == 0 {
		t.Fatalf("no aws.kropath.run *Config CRDs found under %s — is KROPATH_AWS_CRDS_DIR correct?", dir)
	}

	// controllerKinds is the union of all Kinds entries in features.All that end with "Config".
	// KropathConfig is included here because every cascade reconciler watches it as a secondary kind.
	controllerKinds := map[string]bool{}
	for _, r := range features.All {
		for _, k := range r.Kinds {
			if strings.HasSuffix(k, "Config") {
				controllerKinds[k] = true
			}
		}
	}

	// Check 1: upstream aws.kropath.run *Config kind has no controller origin and is not allowlisted.
	for kind := range upstreamKinds {
		if controllerKinds[kind] {
			continue
		}
		if ticket, allowed := configKindSetAllowlist[kind]; allowed {
			t.Logf("NOTE: %s exists in kropath-aws with no controller origin (tracking %s)", kind, ticket)
			continue
		}
		t.Errorf("%q is in kropath-aws (aws.kropath.run) but has no controller origin in features.All "+
			"and is absent from configKindSetAllowlist.\n"+
			"Add a cascade reconciler for it, or add it to configKindSetAllowlist with a tracking ticket "+
			"in internal/features/crds_drift_test.go.",
			kind)
	}

	// Check 2: controller-known Config kind in aws.kropath.run has no upstream CRD.
	// (This overlaps with TestWatchedKindsMatchUpstreamCRDs but gives a targeted message.)
	for kind := range controllerKinds {
		if upstreamKinds[kind] != "" {
			continue
		}
		t.Errorf("features.All lists Kind %q (in aws.kropath.run) but no matching CRD was found "+
			"in kropath-aws — the controller is watching a kind that does not exist upstream.\n"+
			"Either add the CRD to kropath-aws or remove the reconciler entry from features.All.",
			kind)
	}
}

// ── Unit tests ───────────────────────────────────────────────────────────────────
//
// These tests run without a cluster or KROPATH_AWS_CRDS_DIR and verify the core
// comparison helpers directly.

// TestExtractSpecDefaultsDetectsInjectedDefault verifies that extractDefaultsFromSchema
// returns distinct values for two schema nodes where one has an injected non-empty
// default. This is the unit-level check for AC-1.
func TestExtractSpecDefaultsDetectsInjectedDefault(t *testing.T) {
	// fixtureSchema has blockPublicAccess default "true" — a deliberate injection.
	const fixtureSchemaYAML = `
type: object
properties:
  mandatory:
    type: object
    properties:
      blockPublicAccess:
        type: boolean
        default: true
      encryptionAlgorithm:
        type: string
        default: "AES256"
`
	// upstreamSchema has the correct default values.
	const upstreamSchemaYAML = `
type: object
properties:
  mandatory:
    type: object
    properties:
      blockPublicAccess:
        type: boolean
        default: false
      encryptionAlgorithm:
        type: string
        default: ""
`
	parseNode := func(t *testing.T, src string) map[string]interface{} {
		t.Helper()
		var m map[string]interface{}
		if err := yaml.Unmarshal([]byte(src), &m); err != nil {
			t.Fatalf("parsing schema: %v", err)
		}
		return m
	}

	fixture := parseNode(t, fixtureSchemaYAML)
	upstream := parseNode(t, upstreamSchemaYAML)

	fixDefs := extractDefaultsFromSchema(fixture, "")
	upDefs := extractDefaultsFromSchema(upstream, "")

	cases := []struct {
		path            string
		wantFixture     interface{}
		wantUpstream    interface{}
		expectDifferent bool
	}{
		{"mandatory.blockPublicAccess", true, false, true},
		{"mandatory.encryptionAlgorithm", "AES256", "", true},
	}
	for _, tc := range cases {
		fixVal, fixOK := fixDefs[tc.path]
		upVal, upOK := upDefs[tc.path]
		if !fixOK {
			t.Errorf("fixture: default at path %q not found", tc.path)
			continue
		}
		if !upOK {
			t.Errorf("upstream: default at path %q not found", tc.path)
			continue
		}
		if !reflect.DeepEqual(fixVal, tc.wantFixture) {
			t.Errorf("fixture default at %q = %v, want %v", tc.path, fixVal, tc.wantFixture)
		}
		if !reflect.DeepEqual(upVal, tc.wantUpstream) {
			t.Errorf("upstream default at %q = %v, want %v", tc.path, upVal, tc.wantUpstream)
		}
		isDifferent := !reflect.DeepEqual(fixVal, upVal)
		if isDifferent != tc.expectDifferent {
			t.Errorf("path %q: isDifferent=%v, want %v (fixture=%v, upstream=%v)",
				tc.path, isDifferent, tc.expectDifferent, fixVal, upVal)
		}
	}
}

// TestConfigKindSetDriftDetectsMissingKind verifies that configKindsFromCRDDir
// correctly identifies a *Config kind that exists in a synthetic upstream directory
// but has no controller origin in features.All and is not in the allowlist.
// This is the unit-level check for AC-2.
func TestConfigKindSetDriftDetectsMissingKind(t *testing.T) {
	dir := t.TempDir()

	const fakeCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: newfakeconfigs.aws.kropath.run
spec:
  group: aws.kropath.run
  names:
    kind: NewFakeConfig
    plural: newfakeconfigs
  versions:
    - name: v1alpha1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
`
	if err := os.WriteFile(filepath.Join(dir, "newfakeconfig.yaml"), []byte(fakeCRD), 0o600); err != nil {
		t.Fatal(err)
	}

	upstreamKinds, err := configKindsFromCRDDir(dir)
	if err != nil {
		t.Fatalf("configKindsFromCRDDir: %v", err)
	}
	if _, ok := upstreamKinds["NewFakeConfig"]; !ok {
		t.Fatal("configKindsFromCRDDir did not detect NewFakeConfig — kind-set scanning is broken")
	}

	// Verify the detection path: NewFakeConfig must not be in features.All or the allowlist.
	controllerKinds := map[string]bool{}
	for _, r := range features.All {
		for _, k := range r.Kinds {
			controllerKinds[k] = true
		}
	}
	if controllerKinds["NewFakeConfig"] {
		t.Skip("NewFakeConfig unexpectedly found in features.All — unit test is no longer valid")
	}
	if _, allowed := configKindSetAllowlist["NewFakeConfig"]; allowed {
		t.Skip("NewFakeConfig is in configKindSetAllowlist — unit test is no longer valid")
	}

	// Simulate the kind-set drift check: NewFakeConfig should be flagged as a gap.
	driftDetected := false
	for kind := range upstreamKinds {
		if controllerKinds[kind] {
			continue
		}
		if _, allowed := configKindSetAllowlist[kind]; allowed {
			continue
		}
		driftDetected = true
	}
	if !driftDetected {
		t.Error("kind-set drift not detected for NewFakeConfig — " +
			"TestConfigKindSetDrift would silently pass on a real gap")
	}
}
