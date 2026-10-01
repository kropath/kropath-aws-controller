# Contributing to kropath-aws-controller

Thanks for your interest in kropath-aws-controller. This project is **experimental** and moving
quickly — please read the status notice in the [README](README.md) before you build anything on
top of it.

## How to Contribute

**Bugs & small fixes → Open a PR!**

Incorrect merge precedence, nil-guard crashes, wrong or flaky Chainsaw asserts, missing CRD
fixtures, typos, and doc corrections don't need any prior discussion. Fork, fix, and open a pull
request directly.

**New features / architecture → Start a GitHub Issue**

Feature requests are very welcome — open an
[issue](https://github.com/kropath/kropath-aws-controller/issues). Please note that **we are not
accepting pull requests for features yet.** Accepted requests are added to the development
roadmap and implemented by the maintainers; the issue is where the design gets agreed and where
you can follow progress.

Open an issue rather than a PR when your change would:

- add a new reconciler or a new `<Service>Config` CRD type
- change the config-merge model (the `mandatory` M-1…M-6 / `defaults` D-1…D-6 precedence chains,
  or the `status.effectiveConfig` shape)
- change `KropathConfig`, or any published CRD schema
- alter the feature registry contract, the `/features` payload, or the metrics surface
- change the test harness, the CI workflows, or the release automation

`status.effectiveConfig` is a published API — the kro RGDs in
[kropath-aws](https://github.com/kropath/kropath-aws) read it through CEL, and a field that moves
breaks every RGD that referenced it. Agreeing on the shape first saves a rewrite on both sides.

## Before You Start

Two documents will save you hours. Read them first:

1. **[`CLAUDE.md`](CLAUDE.md)** — the working conventions for this repo (written for both humans
   and coding agents): the reconciler contracts, the merge precedence chains, and the PR gate.
2. **[`docs/frequent-chainsaw-errors.md`](docs/frequent-chainsaw-errors.md)** — the catalog of
   integration-test failure modes specific to this repo. Almost every symptom you hit is already
   written down there, with the fix.

[`docs/STANDARDS.md`](docs/STANDARDS.md) carries the engineering standards, and
[`docs/troubleshooting-logs/`](docs/troubleshooting-logs/) has dated write-ups of specific
incidents.

## Development Setup

No AWS account or credentials are needed — the integration flow runs the operator
out-of-cluster against a local kind cluster, with CRDs applied from `tests/fixtures/crds/`.

```bash
# Prerequisites: Go 1.26.6, docker, kind, chainsaw, golangci-lint
make install-tools   # installs the pinned kind / chainsaw / golangci-lint / goimports

make build           # compile bin/kropath-operator
make test            # unit tests with the race detector
make test-chainsaw   # kind-up → build → apply CRDs → start operator → all suites → stop
```

To iterate on one suite without restarting the operator:

```bash
make chainsaw-start chainsaw-wait
make test-iam        # one target per service — `make help` lists them all
make chainsaw-stop
```

On a failed run, clean up manually with `make chainsaw-stop kind-down`.

## Adding a Reconciler

There are no per-feature flags (architecture decision P-1) — every reconciler registered in
`internal/features.All` runs unconditionally. To add one:

1. Create its package under `internal/reconciler/<pkg>/`.
2. Add its CRD Go type to `api/v1alpha1/` and register the type **and** its `List` in
   `api/v1alpha1/register.go`.
3. Add the cascade helper to `internal/cascade/<service>.go` if it merges config.
4. Add an entry to `internal/features.All`.
5. Add the CRD manifest to `tests/fixtures/crds/`.
6. Run `make features-gen` to regenerate `docs/features.yaml`.

Steps 4–6 are enforced, not optional:

- `TestRegistryCoversAllPackages` fails on a missing registry entry.
- `TestEveryReconcilerHasCRDFixture` fails on a missing CRD fixture. This one matters: with the
  flags retired, a reconciler whose CRD is absent from the cluster **takes down the whole
  manager** about two minutes after startup, which surfaces as unrelated Chainsaw suites timing
  out. See `docs/frequent-chainsaw-errors.md` §1.
- The **Feature registry drift gate** CI job fails if `docs/features.yaml` doesn't match the
  registry. Never hand-edit that file.

## Writing Tests

Every new feature needs **both** layers: unit tests for the merge/resolution logic in isolation,
and a Chainsaw suite exercising the full reconcile path.

Suites live in `tests/<service>/<suite-id>/` and run via `make test-<service>`. Non-negotiable
rules, all learned from real failures:

- **A unique resource name per step.** Give every acceptance criterion its own resource name
  (`ac<N>-<family>-policy`) and delete **nothing** between steps. The controller pairs a
  namespaced config with its `kro-system` counterpart *by name*, so a per-step name is a complete
  isolation boundary. `tests/eks/ctrl-eks-01/` is the reference layout.
- **One `ac<NN>-setup.yaml` plus one `ac<NN>-assert.yaml` per acceptance criterion.**
- **`skipDelete: true` does not transfer from kropath-aws.** That repo's global default exists
  because its cluster has no ACK controllers; the reasoning does not apply here. See
  `docs/frequent-chainsaw-errors.md` §3.
- **Never drop an `expect:` block** — a missing one surfaces as a confusing rate-limiter error
  rather than an assertion failure (§4).
- **Assert lists order-stably** (§6). Map-derived iteration order is not guaranteed.
- **A reset script must be deleted, never renamed** (§5).

Log any non-obvious fix in `docs/troubleshooting-logs/<YYYY-MM-DD>-<slug>.md` so the next
contributor doesn't rediscover it.

## Pull Requests

Run all four gates and confirm they pass before opening a PR:

```bash
make lint
make test-cover
make test-chainsaw
make security
```

Keep the change scoped to one reconciler or one concern where possible. In the PR description,
state what changed, which suites you ran, and the result.

CI runs lint, unit tests, the feature-registry drift gate, security scans, the image build, and —
on pull requests only — the Chainsaw integration suites. The `Unit tests` and `Chainsaw
integration tests` contexts are required by the `main` ruleset and must be green before merge.

### PR titles are load-bearing

PRs are squash-merged, so **the PR title becomes the commit subject on `main`** and is the only
input release-please parses. It must be a conventional commit with the ticket id as the scope:

```
feat(scope): add feature-registry metadata
fix(scope): guard nil effectiveConfig on first reconcile
docs(scope): document blocked_by metadata format
```

`.github/workflows/pr-title.yaml` enforces this. The older `[TICKET-ID]: feat: …` form is
**rejected** — the bracketed prefix breaks the conventional-commit header regex, so the commit
parses with no type, never bumps the version, and never reaches `CHANGELOG.md`.

| Type | Release effect |
|---|---|
| `feat` | minor bump |
| `fix`, `perf`, `deps` | patch bump |
| a trailing `!` (`feat(scope)!: …`) | major bump |
| `refactor`, `docs`, `test`, `build`, `ci`, `chore`, `revert` | **no release** |

## Code of Conduct

Be respectful and constructive. Assume good faith, keep reviews about the code, and prefer
questions over assertions when something looks wrong.

## License

By contributing, you agree that your contributions will be licensed under the
[Apache License 2.0](LICENSE).
