# kropath-controller metrics

Operator-facing reference for every metric this controller exposes on `/metrics`
(port `8080` by default, `--metrics-bind-address`) plus the alerting rules shipped
in `config/monitoring/rules.yaml`.

Spec: [`controller-observability-metrics.md`](https://github.com/kropath/kropath-core/blob/main/docs/specs/controller-observability-metrics.md)
§1, §2, §4. Design: [`controller-observability-metrics.md`](https://github.com/kropath/kropath-core/blob/main/docs/design/controller-observability-metrics.md)
§8, §9.

## Reading this document

- **Type** follows the Prometheus metric types: `counter` (monotonically
  increasing; use `rate()`/`increase()`), or `gauge` (point-in-time value,
  sampled at scrape time by a collector; use `max by (...)`, never `sum by
  (...)`, across replicas — see [Cardinality and aggregation](#cardinality-and-aggregation)).
- **Closed label values** are the complete, enumerable set a label can take.
  A value outside this list is either a code defect or a stale alert/panel —
  the `internal/monitoring` label-value check (`go test ./internal/monitoring/...`)
  fails CI if `config/monitoring/rules.yaml` or the dashboard names a value not
  in this list.
- **Alert** names the `config/monitoring/rules.yaml` rule that reads this
  metric, or "—" if the metric is dashboard-only (§4.1: every metric with no
  alert has a dashboard panel instead, to satisfy the bidirectional
  metric-name check).

## Cardinality and aggregation

The collector-registering runnable (`internal/metrics.NewCollectorRunnable`)
registers on **every** replica, not only the elected leader, so every replica
reports the same cluster-wide value for every gauge in this document. Alert
expressions and dashboard panels therefore use `max by (...)`, never `sum by
(...)`, for a gauge — `sum` over N replicas reads N times too high. Counters
are per-replica facts, so `sum by (...) (rate(...))` is correct for them.

The one exception is `kropath_registry_reconciler_pending_since_timestamp_seconds`,
which is a **per-process observation** (the instant *this* process first saw
the reconciler as pending), not a cluster-wide fact every replica computes
identically. Its alert uses `min by (package)` instead: a replica that
restarted recently reports a recent timestamp, and `max` would let that fresh
observation silently resolve an alert about a CRD that is still missing on
the cluster. `min` takes the earliest observation, the best available lower
bound on when the CRD actually went missing.

## Shared — placement and config profile

| Metric | Type | Labels | Closed label values | Alert |
|---|---|---|---|---|
| `kropath_placement_resolutions_total` | counter | `reason` | `NamespaceUnreadable`, `TeamAnnotationUnsupported`, `MissingAccountAnnotation`, `InvalidAccountAnnotation`, `MissingRegionAnnotation`, `PlacementResolved`, `GlobalTierInput`, `ResolvedFromNamespace` (contract value; never actually emitted — see note below) | `KropathPlacementResolutionFailing` |
| `kropath_config_profile_resolutions_total` | counter | `family`, `reason` | `family`: the `features.All` package name for every registered family (e.g. `s3config`), never the CRD Kind (`S3Config`); `reason`: `ProfileFound`, `ProfileFallthrough`, `ProfileUnresolved` | dashboard panel: *Config profile fallthrough by family* |

`ResolvedFromNamespace` is a member of the closed set (and therefore passes
the label-value check) but is never emitted: it is the family-layer condition
name for a resolution that `kropath_placement_resolutions_total` has already
counted as `PlacementResolved` one layer down. Counting both would double-count
every successful resolution.

## PolicyDocument

| Metric | Type | Labels | Closed label values | Alert |
|---|---|---|---|---|
| `kropath_policydocument_documents` | gauge (collector, `Ready` condition) | `reason` | `DocumentResolved`, `InvalidDocumentJSON`, `SidConflict`, `SourceNotReady`, `SourceMissing`, `SourcePending`, `MergeFromRawNotSupported` | `KropathPolicyDocumentsNotReady` |
| `kropath_policydocument_unresolved_refs` | gauge (collector) | `kind`, `field` | `kind`: `AWSIAMRole`, `AWSS3Bucket`, `AWSLambdaFunction`, `AWSSQSQueue`, `AWSKMSKey`, `AWSSecretsManagerSecret`, `other` (bucket for any kind outside that set); `field`: `predictedArn`, `arn`, `unsupported` | dashboard panel: *Unresolved policy refs* |
| `kropath_policydocument_ref_resolutions_total` | counter | `kind`, `field`, `outcome` | `kind`/`field` as above; `outcome`: `resolved`, `pending`, `crd_absent`, `error` | `KropathPolicyRefCRDAbsent` (`outcome="crd_absent"` slice) |
| `kropath_policydocument_sid_conflicts_total` | counter | — | — | dashboard panel: *Sid conflict rate* |

`SourceNotReady` (an unresolved *ref*) and `SourcePending` (an unresolved
source *document*) are distinct failures with distinct operator actions —
don't conflate them when triaging `kropath_policydocument_documents`.

`kind` and `field` are bucketed rather than raw because `PolicyRef.Kind` has
no CRD enum — a tenant typo would otherwise be an unbounded label. A
`kind="other"` series is itself the signal that something is referencing a
kind kropath does not watch at all.

## Config cascade

| Metric | Type | Labels | Closed label values | Alert |
|---|---|---|---|---|
| `kropath_cascade_effective_config_withheld` | gauge (collector, `APIReader`) | `family`, `reason` | `family`: `features.All` package name; `reason`: the `PlacementResolved` condition reason withholding the config (any `kropath_placement_resolutions_total` reason value) | `KropathEffectiveConfigWithheld` |
| `kropath_cascade_mapfunc_errors_total` | counter | `family`, `trigger` | `family`: `features.All` package name; `trigger`: `kropathconfig`, `familyconfig`, `namespace` | `KropathCascadeChangeEventsDropped` |

## Label injection

| Metric | Type | Labels | Closed label values | Alert |
|---|---|---|---|---|
| `kropath_labeloperator_watched_kinds` | gauge | `group` | `aws.kropath.run`, `gcp.kropath.run`, `azure.kropath.run` | `KropathLabelInjectionOff` |
| `kropath_labeloperator_group_discovery_total` | counter | `group`, `outcome` | `group` as above; `outcome`: `discovered`, `empty`, `error` | dashboard panel: *Group discovery outcomes* |
| `kropath_labeloperator_patches_total` | counter | `group`, `outcome` | `group` as above; `outcome`: `patched`, `not_found`, `error` | dashboard panel: *Label patch outcomes* |

Unlike `kropath_reconciler_active`, `kropath_labeloperator_watched_kinds` is
explicitly `0` (not simply absent) when discovery ran and found nothing to
watch, so `absent(...) or ... == 0` in `KropathLabelInjectionOff` catches both
"never registered" and "registered with zero kinds."

## KropathConfigStatus

| Metric | Type | Labels | Closed label values | Alert |
|---|---|---|---|---|
| `kropath_kropathconfigstatus_configs` | gauge (collector, `APIReader`) | `reason` | `GlobalTier`, `LocalTier`, `GlobalAndLocalTier`, `Unreferenced` | `KropathConfigUnreferenced` (`reason="Unreferenced"` slice) |
| `kropath_kropathconfigstatus_family_kinds_unavailable` | gauge | — | — | `KropathConfigFamilyKindsUnavailable` |

A non-zero `kropath_kropathconfigstatus_family_kinds_unavailable` means a
`<Family>Config` CRD was absent at the last consumer count, which can
understate consumers and flip a `KropathConfig` to `Unreferenced` even though
it is correctly placed — treat `KropathConfigUnreferenced` with suspicion
while this gauge is non-zero.

## NamespacePlacement

| Metric | Type | Labels | Closed label values | Alert |
|---|---|---|---|---|
| `kropath_namespaceplacement_namespaces` | gauge (collector) | `status` | `ok`, `TeamAnnotationUnsupported`, `MissingAccountAnnotation`, `InvalidAccountAnnotation`, `MissingRegionAnnotation` | dashboard panel: *Placement verdicts* |
| `kropath_namespaceplacement_transitions_total` | counter | `to` | same 5 values as `status` above | dashboard panel: *Placement transitions* |

## Registry and dynamic CRD detection

| Metric | Type | Labels | Closed label values | Alert |
|---|---|---|---|---|
| `kropath_registry_reconciler_pending_since_timestamp_seconds` | gauge | `package` (exempt from the label-value check — see below) | `features.All` package names | `KropathReconcilerPendingTooLong` |
| `kropath_registry_crd_watch_events_total` | counter | `outcome` | `activated`, `store_miss`, `not_servable`, `cast_failed`, `already_active` | dashboard panel: *CRD watch events* |
| `kropath_registry_optional_kinds_attached` | gauge | `package` (exempt) | `features.All` package names | dashboard panel: *Optional kinds attached* |

`package` is the one label the §10.2.1 label-value check does not validate:
it is bounded but code-derived from `features.All`, and pinning its value set
here would fail CI on every new resource family. The bidirectional
metric-name check (every metric in this document appears in a rule or panel,
and vice versa) still covers the metric itself.

`kropath_registry_reconciler_pending_since_timestamp_seconds` has **no
series** for a package once that reconciler activates — an absent series
means "not pending," not zero. See [Cardinality and aggregation](#cardinality-and-aggregation)
for why its alert aggregates with `min by`, not `max by`.

## Collector self-observability

| Metric | Type | Labels | Closed label values | Alert |
|---|---|---|---|---|
| `kropath_metrics_collect_errors_total` | counter | `collector` | `policydocument_documents`, `policydocument_unresolved_refs`, `cascade_effective_config_withheld`, `kropathconfigstatus_configs`, `namespaceplacement_namespaces` | `KropathMetricsCollectorFailing` |

Every scrape-time collector in `internal/metrics` lists under a 2-second
timeout. On any list error it emits **no samples** for its own metric and
increments this counter instead — emitting a zero would assert "no objects
are in a bad state," which is a worse failure than a visible gap.
`KropathMetricsCollectorFailing` exists precisely because an absent series is
otherwise indistinguishable from a healthy one.

## Alerting rules

Canonical source: [`config/monitoring/rules.yaml`](../config/monitoring/rules.yaml).
`config/monitoring/prometheusrule.yaml` is **generated** from it by `make
monitoring-gen` (`cmd/gen-monitoring`) — never hand-edit the generated file.
CI (`make monitoring-verify`) fails if it drifts from `rules.yaml`.

| Alert | Severity | `for` | Fires when |
|---|---|---|---|
| `KropathConfigUnreferenced` | critical | 10m | A `KropathConfig` is resolved by no `<Family>Config` — the KRO-1104 failure mode |
| `KropathConfigFamilyKindsUnavailable` | warning | 15m | A `<Family>Config` CRD is absent, so `KropathConfigUnreferenced` may be understated |
| `KropathEffectiveConfigWithheld` | critical | 10m | A family publishes no `status.effectiveConfig`; RGDs reading it fail in CEL |
| `KropathPlacementResolutionFailing` | warning | 15m | Namespace placement is rejecting at a sustained rate (excludes the two success reasons) |
| `KropathPolicyDocumentsNotReady` | warning | 15m | A `PolicyDocument` is stuck on anything other than `DocumentResolved` |
| `KropathPolicyRefCRDAbsent` | warning | 10m | A `PolicyDocument` references a kind whose CRD is not installed |
| `KropathLabelInjectionOff` | critical | 5m | No label-injection controller is registered for a provider group |
| `KropathCascadeChangeEventsDropped` | warning | 10m | A family's watch map function dropped a batch of change events |
| `KropathReconcilerPendingTooLong` | warning | 5m | A reconciler has been pending (its CRD missing) for over an hour |
| `KropathMetricsCollectorFailing` | warning | 10m | A collector cannot list; its metric's series are silently absent |

Every rule carries `component: kropath-controller` so a consumer can route
every alert from this operator to one receiver without enumerating alert
names. `severity: critical` means tenant workloads are or will be broken;
`severity: warning` means degraded or blind, with no tenant impact yet. This
repo defines the labels routing is written against — not the consumer's
alertmanager receivers, inhibition rules, or escalation policy.

`for:` durations and thresholds are this repo's own defaults, derived from
its requeue intervals (`kropathconfigstatus`'s 15s, `policydocument`'s 10s
`defaultRequeueAfter`). A consumer wanting different thresholds patches
`config/monitoring/kustomization.yaml` in their own overlay rather than
relying on a flag — this repo does not attempt to make them configurable.

## Deploying the rules and dashboard

`config/monitoring/` is its own kustomization and is **not** part of any
default deploy — it depends on the Prometheus Operator's `PrometheusRule`
CRD, which this repo does not otherwise require. Nothing in `cmd/manager`
reads a `PrometheusRule`, and no health or readiness check references one.

```bash
make deploy-monitoring   # kubectl apply -k config/monitoring — requires the
                          # Prometheus Operator CRDs already installed
```

`config/monitoring/dashboards/kropath-controller.json` ships as a plain file
for the consumer's own provisioning (e.g. a Grafana sidecar `ConfigMap`) and
is not part of the kustomization.
