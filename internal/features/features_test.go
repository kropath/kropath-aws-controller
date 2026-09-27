// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package features_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kropath/kropath-controller/api/v1alpha1"
	"github.com/kropath/kropath-controller/internal/features"
	"github.com/kropath/kropath-controller/internal/registry"
)

// helperPackages lists directories under internal/reconciler/ that contain shared
// helpers rather than a Reconciler struct and should not appear in features.All.
var helperPackages = map[string]bool{
	"util": true,
}

// TestRegistryCoversAllPackages ensures that every directory under
// internal/reconciler/ is listed in features.All, and vice versa.
// This is the CI gate that makes "adding a reconciler without registering it"
// a failing test rather than a silent operational gap.
func TestRegistryCoversAllPackages(t *testing.T) {
	entries, err := os.ReadDir("../reconciler")
	if err != nil {
		t.Fatalf("reading internal/reconciler: %v", err)
	}

	dirs := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() && !helperPackages[e.Name()] {
			dirs[e.Name()] = true
		}
	}

	registered := map[string]bool{}
	for _, r := range features.All {
		registered[r.Package] = true
	}

	for pkg := range dirs {
		if !registered[pkg] {
			t.Errorf("package %q exists in internal/reconciler/ but is missing from features.All — add it", pkg)
		}
	}
	for pkg := range registered {
		if !dirs[pkg] {
			t.Errorf("features.All lists package %q but internal/reconciler/%s/ does not exist — remove it or create the package", pkg, pkg)
		}
	}
}

// packagesWithoutOwnCRD lists reconcilers in features.All that do not own a
// primary <package>s.aws.kropath.run CRD. The label operator watches every
// already-registered config kind rather than a kind of its own.
var packagesWithoutOwnCRD = map[string]bool{
	"labeloperator":       true,
	"kropathconfigstatus": true, // watches the existing KropathConfig kind, owns none of its own
	"namespaceplacement":  true, // watches the core Namespace kind, owns no CRD
}

// coreKinds lists built-in Kubernetes kinds this controller watches that are
// served by the API server itself rather than by any kropath-aws CRD.
// kropath-aws is the authority for every CRD kind (see TestWatchedKindsMatchUpstreamCRDs),
// but it has no opinion on core types, so those kinds are exempt from that check.
var coreKinds = map[string]bool{
	"Namespace": true, // core/v1, watched by namespaceplacement
}

// crdNameRE matches the `  name: <crd>` line of a CRD's metadata block.
var crdNameRE = regexp.MustCompile(`(?m)^  name: ([a-z0-9.]+)$`)

// crdFixtureDirs lists the directories that may contain CRD fixtures.
// tests/fixtures/crds/ holds CRDs applied at setup time.
// tests/fixtures/crds-optional/ holds CRDs that dynamic-detection suites need absent
// at operator startup and install themselves on demand.
// Both directories count as "present" for the purposes of the tests below.
var crdFixtureDirs = []string{
	"../../tests/fixtures/crds",
	"../../tests/fixtures/crds-optional",
}

// gatherFixturesFromDirs globs all YAML files from the given directories and
// returns a flat slice of file paths. Exported as a package-level helper so
// that unit tests can exercise it with temporary directories.
func gatherFixturesFromDirs(t *testing.T, dirs []string) []string {
	t.Helper()
	var all []string
	for _, dir := range dirs {
		matches, err := filepath.Glob(dir + "/*.yaml")
		if err != nil {
			t.Fatalf("globbing %s: %v", dir, err)
		}
		all = append(all, matches...)
	}
	return all
}

// gatherCRDFixtures globs all YAML files from both fixture directories and
// returns a flat slice of file paths.
func gatherCRDFixtures(t *testing.T) []string {
	t.Helper()
	return gatherFixturesFromDirs(t, crdFixtureDirs)
}

// TestEveryReconcilerHasCRDFixture ensures every reconciler in features.All has its
// CRD in tests/fixtures/crds/ or tests/fixtures/crds-optional/. The former is applied
// by `make chainsaw-setup`; the latter is for CRDs that dynamic-detection suites need
// absent at operator startup.
//
// This guards a failure mode that is expensive to diagnose from its symptom. Since
// per-feature flags were retired (KRO-635), every reconciler starts unconditionally.
// If a reconciler watches a kind whose CRD is absent from the cluster, its informer
// never syncs, and controller-runtime aborts the *entire manager* once the 2-minute
// cache-sync timeout elapses. The operator therefore serves traffic normally for two
// minutes and then exits, so the visible symptom is that every Chainsaw suite which
// happens to run after the ~2-minute mark fails on a 30s assert timeout — while the
// suites that ran before it pass. Which suites fail depends only on machine speed and
// test ordering, not on the resource families involved, which makes the failure look
// unrelated to its actual cause.
func TestEveryReconcilerHasCRDFixture(t *testing.T) {
	fixtures := gatherCRDFixtures(t)
	if len(fixtures) == 0 {
		t.Fatal("no CRD fixtures found under tests/fixtures/crds/ or tests/fixtures/crds-optional/")
	}

	present := map[string]bool{}
	for _, f := range fixtures {
		data, err := os.ReadFile(f) //nolint:gosec // test-only read of a repo-relative fixture path
		if err != nil {
			t.Fatalf("reading %s: %v", f, err)
		}
		for _, m := range crdNameRE.FindAllStringSubmatch(string(data), -1) {
			present[m[1]] = true
		}
	}

	for _, r := range features.All {
		if packagesWithoutOwnCRD[r.Package] {
			continue
		}
		crd := r.Package + "s.aws.kropath.run"
		if !present[crd] {
			t.Errorf("reconciler %q (package %q) watches a kind with no CRD fixture: expected %q under tests/fixtures/crds/ or tests/fixtures/crds-optional/.\n"+
				"Without it the manager exits after the 2-minute cache-sync timeout and later Chainsaw suites fail on assert timeouts.\n"+
				"Copy the CRD from kropath-aws/crds/ into tests/fixtures/crds/ (or crds-optional/ if the suite needs it absent at startup), "+
				"or add %q to packagesWithoutOwnCRD if it genuinely owns no CRD.",
				r.Name, r.Package, crd, r.Package)
		}
	}

	// Every config reconciler also reads KropathConfig, so that CRD must be present too.
	for _, crd := range []string{"kropathconfigs.aws.kropath.run", "kropathconfigs.kropath.run"} {
		if !present[crd] {
			t.Errorf("missing CRD fixture %q under tests/fixtures/crds/ or tests/fixtures/crds-optional/ — every config reconciler watches KropathConfig", crd)
		}
	}
}

// crdKindRE matches the `    kind: <Kind>` line of a CRD's spec.names block.
// Only spec.names.kind sits at this indent — group/names/scope/versions are its
// only siblings, and schema properties are nested far deeper.
var crdKindRE = regexp.MustCompile(`(?m)^    kind: ([A-Za-z0-9]+)$`)

// TestReconcilerKindMatchesFixtureAndScheme asserts, for every reconciler, that
// three names agree exactly: the Kind in features.All, the Kind its CRD fixture
// serves, and the Kind the runtime scheme registers.
//
// TestEveryReconcilerHasCRDFixture only checks that a CRD with the right
// *metadata name* (the plural, e.g. apigatewayconfigs.aws.kropath.run) exists.
// That plural is case-insensitive by construction, so it stays green even when
// the CRD serves a differently-cased Kind than the one the controller watches.
//
// The Kind is what actually matters. controller-runtime derives the GVK it
// watches from the Go type name registered via scheme.AddKnownTypes, and the API
// server serves the Kind spelled in spec.names.kind. If those differ by so much
// as one letter's case, the informer matches nothing, never syncs, and takes the
// *entire manager* down at the 2-minute cache-sync timeout — the failure mode
// documented in docs/frequent-chainsaw-errors.md §1.
//
// KRO-675: the controller watched "ApiGatewayConfig" while the authoritative CRD
// in kropath-aws serves "APIGatewayConfig". The plural matched, so every in-repo
// check passed, and the mismatch only surfaced as the operator crash-looping in
// the integration cluster.
func TestReconcilerKindMatchesFixtureAndScheme(t *testing.T) {
	fixtures := gatherCRDFixtures(t)
	if len(fixtures) == 0 {
		t.Fatal("no CRD fixtures found under tests/fixtures/crds/ or tests/fixtures/crds-optional/")
	}

	// Map each CRD's metadata name (the plural) to the Kind it serves.
	kindByCRD := map[string]string{}
	for _, f := range fixtures {
		data, err := os.ReadFile(f) //nolint:gosec // test-only read of a repo-relative fixture path
		if err != nil {
			t.Fatalf("reading %s: %v", f, err)
		}
		for _, doc := range strings.Split(string(data), "\n---") {
			name := crdNameRE.FindStringSubmatch(doc)
			kind := crdKindRE.FindStringSubmatch(doc)
			if name != nil && kind != nil {
				kindByCRD[name[1]] = kind[1]
			}
		}
	}

	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("adding api/v1alpha1 to scheme: %v", err)
	}
	schemeKinds := map[string]bool{}
	for gvk := range scheme.AllKnownTypes() {
		if gvk.Group == v1alpha1.GroupVersion.Group && gvk.Version == v1alpha1.GroupVersion.Version {
			schemeKinds[gvk.Kind] = true
		}
	}

	for _, r := range features.All {
		if packagesWithoutOwnCRD[r.Package] {
			continue
		}
		crd := r.Package + "s.aws.kropath.run"

		if got, ok := kindByCRD[crd]; ok && got != r.Name {
			t.Errorf("reconciler %q: features.All calls its Kind %q but CRD fixture %s serves Kind %q.\n"+
				"The plural matches either way, so TestEveryReconcilerHasCRDFixture cannot see this. "+
				"An informer for a Kind the API server does not serve never syncs and kills the whole manager "+
				"at the 2-minute cache-sync timeout.\n"+
				"Fix the casing so both match the authoritative CRD in kropath-aws/crds/.",
				r.Name, r.Name, crd, got)
		}

		if !schemeKinds[r.Name] {
			t.Errorf("reconciler %q: features.All calls its Kind %q, but api/v1alpha1 registers no such Kind in %s.\n"+
				"controller-runtime derives the watched GVK from the registered Go type name, so the informer "+
				"would watch a Kind nothing serves.\n"+
				"Rename the Go type in api/v1alpha1 to match, or correct the Name in features.All.",
				r.Name, r.Name, v1alpha1.GroupVersion.String())
		}
	}
}

// ── Two-directory fixture tests (KRO-860) ────────────────────────────────────

// TestCRDFixtureDirsCoversOptionalDir verifies that crdFixtureDirs includes both
// the standard crds/ directory and the crds-optional/ directory introduced in KRO-860.
func TestCRDFixtureDirsCoversOptionalDir(t *testing.T) {
	hasCrds, hasOptional := false, false
	for _, d := range crdFixtureDirs {
		if strings.HasSuffix(d, "/crds") {
			hasCrds = true
		}
		if strings.HasSuffix(d, "/crds-optional") {
			hasOptional = true
		}
	}
	if !hasCrds {
		t.Error("crdFixtureDirs is missing tests/fixtures/crds")
	}
	if !hasOptional {
		t.Error("crdFixtureDirs is missing tests/fixtures/crds-optional — add it so dynamic-detection suites can use it")
	}
}

// TestFixtureInOptionalDirSatisfiesPresenceCheck verifies that a CRD fixture placed
// in crds-optional/ is treated as present by the same logic
// TestEveryReconcilerHasCRDFixture uses. This is the unit test for AC-U1: a fixture
// in neither directory still fails the check (AC-U2 is covered by the existing
// negative branch of TestEveryReconcilerHasCRDFixture).
func TestFixtureInOptionalDirSatisfiesPresenceCheck(t *testing.T) {
	dir := t.TempDir()
	const fakeCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: fakeresources.test.kropath.run
spec:
  names:
    kind: FakeResource
    plural: fakeresources
`
	if err := os.WriteFile(filepath.Join(dir, "fake.yaml"), []byte(fakeCRD), 0o600); err != nil {
		t.Fatal(err)
	}

	// Simulate placing the fixture in crds-optional/ by passing dir as the optional dir.
	// The crds/ dir is real (has known CRDs); the optional dir is the temp dir with our fake.
	fixtures := gatherFixturesFromDirs(t, []string{"../../tests/fixtures/crds", dir})
	if len(fixtures) == 0 {
		t.Fatal("gatherFixturesFromDirs returned no fixtures")
	}

	present := map[string]bool{}
	for _, f := range fixtures {
		data, err := os.ReadFile(f) //nolint:gosec // test-only read of a repo-relative or temp fixture path
		if err != nil {
			t.Fatalf("reading %s: %v", f, err)
		}
		for _, m := range crdNameRE.FindAllStringSubmatch(string(data), -1) {
			present[m[1]] = true
		}
	}

	// The fake CRD placed in the "optional" dir must appear in the presence map.
	if !present["fakeresources.test.kropath.run"] {
		t.Error("CRD placed in crds-optional/ was not found in the presence map — TestEveryReconcilerHasCRDFixture would incorrectly fail for a reconciler backed by an optional-dir fixture")
	}
}

// TestGatherFixturesFromDirsExcludesNonYAML verifies that gatherFixturesFromDirs
// returns only .yaml files (e.g. README.md in crds-optional/ must be ignored).
func TestGatherFixturesFromDirsExcludesNonYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# readme"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "crd.yaml"), []byte("apiVersion: v1\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	fixtures := gatherFixturesFromDirs(t, []string{dir})
	if len(fixtures) != 1 {
		t.Errorf("expected exactly 1 YAML file; README.md must be excluded — got %d file(s): %v", len(fixtures), fixtures)
	}
}

// TestKindCasingTestReadsBothDirs verifies that TestReconcilerKindMatchesFixtureAndScheme
// (the Kind-casing test added for KRO-675) also reads fixtures from crds-optional/.
// We confirm this indirectly: gatherCRDFixtures (which both tests call) must return
// at least the fixtures from crds/, confirming the same code path is shared.
func TestKindCasingTestReadsBothDirs(t *testing.T) {
	fixtures := gatherCRDFixtures(t)
	hasCrdsFile := false
	for _, f := range fixtures {
		// Any file under crds/ (not crds-optional/) counts as "from crds/".
		if strings.Contains(f, "/crds/") && !strings.Contains(f, "/crds-optional/") {
			hasCrdsFile = true
			break
		}
	}
	if !hasCrdsFile {
		t.Error("gatherCRDFixtures returned no files from tests/fixtures/crds/ — the Kind-casing test would have no fixtures to check")
	}
	// Verify the optional dir path is also in the scan (even if currently empty of .yaml files).
	if len(crdFixtureDirs) < 2 {
		t.Errorf("expected at least 2 fixture dirs; got %v", crdFixtureDirs)
	}
}

// crdDocRE identifies a YAML document that is itself a CustomResourceDefinition,
// as opposed to a CR instance or an example that merely mentions one.
var crdDocRE = regexp.MustCompile(`(?m)^kind: CustomResourceDefinition$`)

// TestWatchedKindsMatchUpstreamCRDs verifies every Kind this controller watches is
// actually served by an authoritative CRD in kropath-aws.
//
// This is the cross-repo half of the guard. The fixtures under tests/fixtures/crds/
// are hand-maintained and deliberately not verbatim copies of kropath-aws/crds/, so
// TestReconcilerKindMatchesFixtureAndScheme can only prove this repo is
// self-consistent — all three names can agree with each other and still disagree
// with the CRD the real cluster serves. That is exactly how KRO-675 shipped.
//
// kropath-aws owns the CRDs (see docs/STANDARDS.md), so it is the authority. The
// test is skipped unless KROPATH_AWS_CRDS_DIR points at a checkout of them; `make
// crds-verify` fetches them and sets it, and CI runs that target.
func TestWatchedKindsMatchUpstreamCRDs(t *testing.T) {
	dir := os.Getenv("KROPATH_AWS_CRDS_DIR")
	if dir == "" {
		t.Skip("KROPATH_AWS_CRDS_DIR not set — run `make crds-verify` to check against kropath-aws")
	}

	upstream := map[string]bool{}
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".yaml" {
			return nil
		}
		data, err := os.ReadFile(path) //nolint:gosec // test-only read of an operator-supplied CRD directory
		if err != nil {
			return err
		}
		for _, doc := range strings.Split(string(data), "\n---") {
			if crdDocRE.MatchString(doc) {
				if kind := crdKindRE.FindStringSubmatch(doc); kind != nil {
					upstream[kind[1]] = true
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", dir, err)
	}
	if len(upstream) == 0 {
		t.Fatalf("no CustomResourceDefinitions found under %s — is KROPATH_AWS_CRDS_DIR correct?", dir)
	}

	// Compare case-insensitively too, so a casing mismatch is reported as such
	// rather than as a wholly absent Kind.
	lower := map[string]string{}
	for k := range upstream {
		lower[strings.ToLower(k)] = k
	}

	watched := map[string]bool{}
	for _, r := range features.All {
		for _, k := range r.Kinds {
			if coreKinds[k] {
				continue
			}
			watched[k] = true
		}
	}

	for kind := range watched {
		if upstream[kind] {
			continue
		}
		if authoritative, ok := lower[strings.ToLower(kind)]; ok {
			t.Errorf("this controller watches Kind %q but kropath-aws serves %q — a casing mismatch.\n"+
				"The plural is identical, so every in-repo check passes while the informer matches nothing, "+
				"never syncs, and kills the whole manager at the 2-minute cache-sync timeout.\n"+
				"kropath-aws owns the CRDs, so rename the Go type in api/v1alpha1 and the entry in features.All to %[2]q.",
				kind, authoritative)
			continue
		}
		t.Errorf("this controller watches Kind %q, but no CRD in kropath-aws serves it.\n"+
			"Either the CRD has not been authored upstream yet, or the reconciler is watching the wrong Kind. "+
			"Until a cluster serves it, the manager exits at the 2-minute cache-sync timeout.", kind)
	}
}

// --- KRO-637 acceptance criteria -------------------------------------------
//
// The tests below restore guarantees specified by KRO-637 that were not
// carried over when KRO-635 landed the simpler static-slice registry.

// TestNoDuplicateEntries asserts that no Name or Package appears twice in
// features.All (KRO-637 AC-6).
//
// KRO-637's map-backed registry got this for free: Register panicked on a
// duplicate key. A static slice has no such guard, so a copy-paste mistake
// would otherwise be invisible — TestRegistryCoversAllPackages compares
// against a map and therefore collapses duplicates silently, and the
// duplicate would surface only as a reconciler started twice at runtime.
func TestNoDuplicateEntries(t *testing.T) {
	seenName := map[string]bool{}
	seenPkg := map[string]bool{}
	for _, r := range features.All {
		if seenName[r.Name] {
			t.Errorf("duplicate Name %q in features.All", r.Name)
		}
		if seenPkg[r.Package] {
			t.Errorf("duplicate Package %q in features.All", r.Package)
		}
		seenName[r.Name] = true
		seenPkg[r.Package] = true
	}
}

// TestPackageIsLowercasedName pins the invariant the observability metrics'
// `family` label derivation depends on (spec §2.3): a call site holding only
// a schema.GroupVersionKind derives the label with
// strings.ToLower(gvk.Kind) and needs no lookup table, which is only correct
// if every entry's Package equals its lowercased Name. Verified for all of
// features.All -- a strict superset of the spec's "for all 57 cascade(name,
// pkg) entries" claim, since the invariant holds for the two
// non-cascade-helper entries (PolicyDocument, LabelOperator) too. A future
// family whose Package is not its lowercased Name would otherwise emit a
// `family` metric label value that matches no other metric.
func TestPackageIsLowercasedName(t *testing.T) {
	for _, r := range features.All {
		if want := strings.ToLower(r.Name); r.Package != want {
			t.Errorf("entry %q: Package = %q, want %q (strings.ToLower(Name))", r.Name, r.Package, want)
		}
	}
}

// TestAllOrderIsDeterministic asserts that repeated reads of features.All
// observe the same order (KRO-637 AC-5).
//
// AC-5 called for sorting by Name because KRO-637 backed the registry with a
// map, and any slice derived from Go map iteration has unstable order (see
// "Chainsaw Test Assertion Stability" in CLAUDE.md). The shipped registry is a
// package-level slice literal, so declaration order already gives the
// stability AC-5 exists to guarantee — docs/features.yaml and the /features
// JSON are byte-identical across runs. This test pins that invariant so a
// future change to a map-backed or lazily-built registry fails here rather
// than as an intermittent diff in generated output.
func TestAllOrderIsDeterministic(t *testing.T) {
	// Compared by Package, the unique stable key: features.Reconciler contains a
	// slice field and so is not comparable with ==.
	first := make([]string, len(features.All))
	for i, r := range features.All {
		first[i] = r.Package
	}
	for i := 0; i < 3; i++ {
		for j, r := range features.All {
			if r.Package != first[j] {
				t.Fatalf("features.All order changed between reads: index %d was %q, now %q", j, first[j], r.Package)
			}
		}
	}
}

// validStability is the closed set of Stability values from KRO-637's design.
var validStability = map[string]bool{"alpha": true, "beta": true, "stable": true}

// TestEveryReconcilerHasCompleteMetadata asserts every entry carries the
// descriptive fields KRO-637 specified, so /features and docs/features.yaml
// stay self-describing as reconcilers are added.
func TestEveryReconcilerHasCompleteMetadata(t *testing.T) {
	for _, r := range features.All {
		if r.Description == "" {
			t.Errorf("reconciler %q has an empty Description", r.Name)
		}
		if r.SinceVersion == "" {
			t.Errorf("reconciler %q has an empty SinceVersion", r.Name)
		}
		if !validStability[r.Stability] {
			t.Errorf("reconciler %q has Stability %q — want one of alpha/beta/stable", r.Name, r.Stability)
		}
	}
}

// wiredRE matches the two shapes used to wire a reconciler — either in
// cmd/manager/main.go (old direct style) or internal/registry/entries.go
// (new data-driven style):
//   - `&<pkg>.Reconciler{`  — struct-based reconcilers
//   - `<pkg>.Setup(`        — labeloperator (and any future multi-controller entries)
func wiredRE(pkg string) *regexp.Regexp {
	return regexp.MustCompile(`\b` + regexp.QuoteMeta(pkg) + `\.(Reconciler\{|Setup\()`)
}

// wiringSources are the files that constitute the "wiring" of a reconciler.
// KRO-848 moved wiring from inline calls in main.go to the data-driven
// internal/registry/entries.go; both files are scanned so that the check
// works before and after the migration.
var wiringSources = []string{
	"../../cmd/manager/main.go",
	"../../internal/registry/entries.go",
}

// TestEveryRegisteredReconcilerIsWired asserts that every entry in features.All
// is actually started by the operator binary.
//
// Wiring may live in cmd/manager/main.go (direct SetupWithManager calls, pre-KRO-848)
// or in internal/registry/entries.go (data-driven Build closures, KRO-848+). Either
// location is sufficient — a package reference in either file means the reconciler runs.
func TestEveryRegisteredReconcilerIsWired(t *testing.T) {
	combined := strings.Builder{}
	for _, path := range wiringSources {
		src, err := os.ReadFile(path)
		if err != nil {
			// Non-fatal: entries.go may not exist on pre-KRO-848 branches.
			continue
		}
		combined.Write(src)
	}
	wiring := combined.String()
	if wiring == "" {
		t.Fatal("could not read any wiring source file")
	}

	for _, r := range features.All {
		if !wiredRE(r.Package).MatchString(wiring) {
			t.Errorf("reconciler %q (package %q) is listed in features.All but never started.\n"+
				"Add its wiring to internal/registry/entries.go (or cmd/manager/main.go for the old style).\n"+
				"/features and docs/features.yaml would otherwise advertise a reconciler the binary does not run.",
				r.Name, r.Package)
		}
	}
}

// TestPolicyDocumentKindsCoverRegistryGVKs verifies that features.All for the
// policydocument reconciler declares every Required kind the registry entry
// watches. Optional GVKs are conditional watches — the manager operates without
// them and they may not exist as upstream CRDs yet — so they are intentionally
// excluded from features.All.Kinds and from this check.
//
// Without this guard, removing a Required kind from registry/entries.go would
// leave features.All silently under-declaring the manager's hard dependencies —
// the regression that prompted KRO-851.
func TestPolicyDocumentKindsCoverRegistryGVKs(t *testing.T) {
	const pkg = "policydocument"

	var pdKinds map[string]bool
	for _, r := range features.All {
		if r.Package == pkg {
			pdKinds = make(map[string]bool, len(r.Kinds))
			for _, k := range r.Kinds {
				pdKinds[k] = true
			}
			break
		}
	}
	if pdKinds == nil {
		t.Fatalf("features.All has no entry for package %q", pkg)
	}

	var found bool
	for _, e := range registry.All() {
		if e.Package != pkg {
			continue
		}
		found = true
		for _, gvk := range e.Required {
			if !pdKinds[gvk.Kind] {
				t.Errorf("registry requires kind %q for %s but features.All does not declare it; "+
					"add it to the Kinds slice in internal/features/features.go and run make features-gen",
					gvk.Kind, pkg)
			}
		}
	}
	if !found {
		t.Fatalf("registry.All has no entry for package %q", pkg)
	}
}
