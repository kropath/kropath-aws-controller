# Standards binding kropath-controller

Scope note: this file carries only the standards with a counterpart in this repo's code or in a
contract this repo publishes. Standards for authoring kro RGDs — required child-resource wiring,
provider deletion policies, CRD directory rules, resource naming templates — are owned by the
provider repos (`kropath-aws`, `kropath-gcp`, `kropath-azure`) and are deliberately not duplicated
here. Issue-tracker conventions are not engineering standards and do not belong in this file.

The sections below document the standards and contracts that govern this controller. Each section
is self-contained so that a contributor can understand the design without access to any external
reference.

## Where this repo sits

**kropath** (kro + golden path) is a multi-cloud golden path platform. `kropath-controller` is the
Go controller: it pre-merges the governance config hierarchy and writes `status.effectiveConfig`
onto namespaced ResourceConfig CRs, so that each kro RGD needs exactly one lookup instead of four.

The provider repos — `kropath-aws` (ACK), `kropath-gcp` (KCC), `kropath-azure` (ASO) — author the
CRDs and RGDs that **consume** `status.effectiveConfig`. They are this controller's only
downstream.

## Governance config hierarchy

Three layers per provider. This controller merges all sources and writes the result to
`status.effectiveConfig` on the namespaced ResourceConfig CR — the single object RGDs read.

| Layer | Kind | Scope |
|---|---|---|
| 1 | `KropathConfig` | Org-wide + namespace |
| 2 | `<ResourceFamily>Config` | Per-resource-family; this controller writes its `status.effectiveConfig` |
| 3 | Resource instance `spec` | Developer overrides |

### Merge priority

Every `internal/cascade/*.go` merger resolves each field independently by taking the **first
non-zero value** down an ordered chain. A source that is absent, nil, or zero-valued is skipped —
it never blanks out a weaker source. The chains run in opposite directions for the two tiers:

| Level | Mandatory source | Direction |
|---|---|---|
| 1 | `KropathConfig` in `kro-system` | **strongest** |
| 2 | `KropathConfig` in resource namespace | |
| 3 | `<Family>Config` in `kro-system` | |
| 4 | `<Family>Config` in resource namespace | weakest |

| Level | Defaults source | Direction |
|---|---|---|
| 6 | `<Family>Config` in resource namespace | **strongest** |
| 7 | `<Family>Config` in `kro-system` | |
| 8 | `KropathConfig` in resource namespace | |
| 9 | `KropathConfig` in `kro-system` | weakest |

Level 5 is the resource instance `spec` — resolved in RGD CEL, not in this controller.

The asymmetry is the whole point and follows Security First: for `mandatory`, the **org-wide**
config wins, so a namespace cannot relax a guardrail. For `defaults`, the **most specific** config
wins, so a team can override a suggestion.

Provider identity (`aws.accountId`, `aws.region`) is copied into `effCfg.aws.*` during the config merge step.

Merge logic requires unit tests. The full controller reconcile path requires Chainsaw integration
tests.

## The `effectiveConfig` contract this controller publishes

RGDs read the merged document through a single `externalRef`:

```yaml
resources:
- id: rsrcCfg
  externalRef:
    apiVersion: aws.kropath.run/v1alpha1
    kind: S3Config
    metadata:
      name: ${schema.spec.configRef}
      namespace: ${schema.metadata.namespace}

- id: ackResource
  template:
    spec:
      tags: ${rsrcCfg.status.effectiveConfig.mandatory.tags + schema.spec.tags + rsrcCfg.status.effectiveConfig.defaults.tags}
```

Consumers access `effCfg.mandatory.*`, `effCfg.defaults.*`, `effCfg.aws.*` — never `effCfg.spec.*`.
Status definitions live under `spec.schema.status`, not `spec.status`.

Changing the shape of `status.effectiveConfig` is a breaking change for every provider repo. Treat
it as a versioned API, not an internal struct.

## ARN inputs this controller consumes

The PolicyDocument reconciler resolves `spec.statements[].principals[].ref` and
`.resources[].ref` to ARNs. It reads, in order:

1. `status.predictedArn` — built by the RGD from the resolved resource name. Never derived from
   `metadata.name`, so this controller must not assume the two match.
2. `status.arn` — fallback for kinds whose ARN cannot be predicted ahead of creation.

A ref that resolves to neither leaves the document unresolved and is counted by
`kropath_policydocument_unresolved_refs` (see `docs/metrics.md`).

## Commit and release convention

PRs are squash-merged, so the PR title becomes the commit subject on `main` and is the only input
release-please parses. Titles must be conventional commits with the ticket id as the scope:

```
feat(scope): add feature-registry metadata
fix(scope): guard nil effectiveConfig on first reconcile
docs(scope): document blocked_by metadata format
```

`feat` cuts a minor release; `fix`, `perf`, and `deps` cut a patch; every other type cuts nothing.
A run of non-releasable commits produces no release PR. `.github/workflows/pr-title.yaml` enforces
this, and the older `[KRO-nnn]: feat: …` form is rejected — it parses with no type and silently
keeps the change out of `CHANGELOG.md`.

## Other standards

- **Security First:** default to secure; mandatory config overrides user input.
- **API group:** `aws.kropath.run` (e.g. `apiVersion: aws.kropath.run/v1alpha1`).
- **Licensing:** Apache 2.0 headers in every script.
- **CEL:** use `${}` for all dynamic values.
