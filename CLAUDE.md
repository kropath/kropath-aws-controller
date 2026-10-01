@docs/STANDARDS.md

## This Repo

**kropath-aws-controller** — Multi-reconciler `kropath-operator` binary. One `controller-runtime` Manager, one Reconciler struct per feature. Every reconciler registered in `internal/features.All` runs unconditionally — there are no per-feature flags (architecture decision P-1).

### Reconciler 1 — Config merge (ADR-010)

- Watches `KropathConfig` (global + namespace) and `ResourceConfig` (global + namespace) CRs
- Merges all four sources and writes `status.effectiveConfig` onto namespaced ResourceConfig CRs
- Merge priority (mandatory): M-1 globalKropathCfg → M-2 → M-3 → M-4 → M-5 → M-6 localResourceCfg
- Merge priority (defaults): D-1 localResourceCfg → D-2 → D-3 → D-4 → D-5 → D-6 globalKropathCfg
- Also copies provider identity (`aws.accountId`, `aws.region`) into `effCfg.aws.*` (ADR-010 D-3)
- Unit tests required for merge logic; integration tests required for full controller reconcile path

### Reconciler 2 — PolicyDocument

- Watches `PolicyDocument` CRs and `KropathConfig` CRs (account ID / region lookup)
- Resolves `spec.statements[].principals[].ref` and `spec.statements[].resources[].ref` to ARNs using ARN prediction; falls back to `status.arn` for non-predictable kinds
- Merges Statement arrays from `spec.sources` (left-ordered concatenation); detects Sid conflicts
- Serializes resolved document to `status.resolvedDocumentJSON`; sets `Ready`, `SidConflict`, `SourceNotReady` conditions
- Raw `spec.documentJSON` path: validate JSON and pass through; not mergeable as a source
- Exposes Prometheus metrics: `kropath_policydocument_documents`, `kropath_policydocument_unresolved_refs`, `kropath_policydocument_ref_resolutions_total`, `kropath_policydocument_sid_conflicts_total` — see [`docs/metrics.md`](docs/metrics.md). `poldoc` was never a shipped prefix; do not reintroduce it.
- Test suite at `tests/policy/` (three phases: CRD validation, ref resolution, source merge)

### Reconciler 3 — KropathConfigStatus (KRO-1121)

- Watches `KropathConfig` only; does not watch the ~57 family Config kinds (would add every one of them to the §1 missing-CRD-kills-the-manager blast radius for a status field that tolerates being briefly stale)
- On each reconcile, lists every `<ResourceFamily>Config` kind (derived from `features.All`, not hand-maintained) via the manager's *uncached* `APIReader` — a cached `List()` for a kind with no running informer starts one lazily and blocks the calling reconcile forever if the CRD is genuinely absent; a direct read instead returns a same-call error, which is treated as zero consumers for that kind
- Classifies itself as the **global tier** for any namespace whose resolved `global-config-namespace` is this object's namespace, and as the **local tier** when a family config shares its own namespace; publishes a `Reconciled` condition with `Reason` one of `GlobalTier` / `LocalTier` / `GlobalAndLocalTier` / `Unreferenced`
- Re-evaluates every `requeueInterval` (15s) in addition to reacting to the KropathConfig object's own create/update/delete, so a family config appearing or disappearing elsewhere is still picked up
- Test suite at `tests/kropathconfig/ctrl-kropathconfig-01/`

### Planned reconcilers (not yet implemented)

- **Composite RGD sequencer** — apply child CRs in dependency order
- **Config cascade validator** — surface conflicts across `KropathConfig` / `ResourceConfig` namespace hierarchy
- **Cross-RGD status aggregator** — roll up child CR ready conditions to parent status

### Shared operator concerns

- Supports multiple replicas with leader election
- Single Deployment, single RBAC manifest, single `/metrics` endpoint (port 8080)
- Health probes: `/healthz` port 8081 (manager alive), `/readyz` port 8081 (leader lease + watches established)
- `/features` endpoint on `:8080` — returns version, git commit, and the live reconciler list as JSON
- **`config/monitoring/`** — alerting rules and a dashboard for every metric this controller exposes. `rules.yaml` is canonical and is what `promtool check rules` validates directly; `prometheusrule.yaml` is **generated** from it by `make monitoring-gen` (`cmd/gen-monitoring`) and must never be hand-edited — `make monitoring-verify` (CI-gated) fails on drift. It is its own kustomization, opt-in via `make deploy-monitoring`, and is not part of any default deploy. See [`docs/metrics.md`](docs/metrics.md) for every metric, its labels, and its alert.
- **No per-feature flags.** Every reconciler in `internal/features.All` runs unconditionally. To add a reconciler: create its package under `internal/reconciler/<pkg>/`, add an entry to `features.All`, and run `make features-gen`. Missing registrations fail `TestRegistryCoversAllPackages`.
- **Every watched kind must have a CRD in `tests/fixtures/crds/` or `tests/fixtures/crds-optional/`.** A reconciler whose CRD is missing from the test cluster takes down the **whole manager** two minutes after startup, surfacing as unrelated Chainsaw suites timing out. `TestEveryReconcilerHasCRDFixture` catches this at unit-test time by scanning both directories. See `docs/frequent-chainsaw-errors.md` §1.
  - **`tests/fixtures/crds/`** — the default. `make chainsaw-setup` applies everything here and derives its `kubectl wait` list from this directory only.
  - **`tests/fixtures/crds-optional/`** — for CRDs that dynamic-detection suites need **absent at operator startup**. `chainsaw-setup` does NOT apply these. A suite installs one on demand via `make chainsaw-install-optional-crd CRD_FILE=<path>`.
  - When in doubt, put the fixture in `crds/`. Only use `crds-optional/` when a suite explicitly requires the CRD to be missing when the manager starts.

### Chainsaw tests

**Read `docs/frequent-chainsaw-errors.md` before writing or debugging a Chainsaw suite.**
It covers the failure modes that are specific to this repo and expensive to rediscover:
a missing CRD killing the manager mid-run, why suites use a unique resource name per
step, why `skipDelete` does *not* transfer from kropath-aws, why a dropped `expect:`
block masquerades as a rate-limiter error, and order-stable list assertions.

The short version for a new suite: give every acceptance criterion its own resource
name (`ac<N>-<family>-policy`), delete nothing between steps, and use one
`ac<NN>-setup.yaml` plus one `ac<NN>-assert.yaml` per AC. `tests/eks/ctrl-eks-01/` is
the reference layout. The controller pairs a namespaced config with its `kro-system`
counterpart *by name*, so a per-step name is a complete isolation boundary.

Dated write-ups of specific incidents live in `docs/troubleshooting-logs/`.

### Before Creating a PR

Run these and confirm all pass before opening a pull request:

```bash
make lint
make test-cover
make test-chainsaw
make security
```

**The PR title must be a conventional commit with the ticket id as the scope** —
`feat(KRO-637): restore feature-registry metadata`. PRs are squash-merged, so the title becomes the
commit subject on `main` and is the only thing release-please parses. The older
`[KRO-637]: feat: …` form is rejected by `.github/workflows/pr-title.yaml`: the bracketed prefix
breaks the conventional-commit regex, so the commit parses with no type and the change never
reaches `CHANGELOG.md`. `feat` cuts a minor release, `fix`/`perf`/`deps` cut a patch, a trailing
`!` (`feat(KRO-637)!: …`) cuts a major, and
`docs`/`chore`/`ci`/`test`/`refactor`/`build`/`revert` cut nothing. See the Releases section of
`README.md`.
