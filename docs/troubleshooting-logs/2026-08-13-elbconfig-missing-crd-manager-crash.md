# Missing ELBConfig CRD Crashed the Manager Two Minutes Into Every Chainsaw Run

**Date:** 2026-08-13
**Issue:** KRO-635 (PR #35 — retire per-feature flags)
**Scope:** `tests/fixtures/crds/`, `Makefile` (`chainsaw-setup`), `internal/features`

## Symptom

CI reported 8 of 22 Chainsaw suites failing, each at almost exactly 35.4s:

```
--- PASS: chainsaw/ctrl-apigwv2-01 (8.65s)
--- PASS: chainsaw/ctrl-iam-01-cascade (6.90s)
... 12 more passes ...
--- FAIL: chainsaw/ctrl-eks-01 (35.40s)
--- FAIL: chainsaw/ctrl-efs-01-cascade (35.77s)
--- FAIL: chainsaw/ctrl-elasticache-01-cascade (35.42s)
--- FAIL: chainsaw/ctrl-dynamodb-01-cascade (35.72s)
--- FAIL: chainsaw/ctrl-ecr-01-cascade (35.44s)
--- FAIL: chainsaw/ctrl-ec2-01-cascade (35.78s)
--- FAIL: chainsaw/ctrl-cwl-01-cascade (35.52s)
--- FAIL: chainsaw/ctrl-autoscaling-01-cascade (35.41s)
Passed 14, Failed 8
```

The PR touched no Chainsaw fixture and no reconciler logic. Its only test-related
change was deleting the 19 `--enable-*` flags from `chainsaw-start`.

## What made this hard to read

Three properties of the symptom actively pointed away from the cause.

1. **The failing set looked meaningful but wasn't.** Eight named resource families
   failed and fourteen passed, which reads as "something is wrong with those eight
   families." Reproducing locally produced a *different* split — 8 passed and 14
   failed, including `iam`, `kms` and `label-operator`, which had passed in CI. The
   membership of the failing set is decided purely by how many suites finish before
   a fixed wall-clock deadline, so a faster machine "fixes" different suites.

2. **Every failing suite passed in isolation.** `chainsaw test tests/autoscaling/…`
   passed in 10.9s. Only the full run failed, which suggests cross-suite state
   pollution — a plausible and completely wrong hypothesis that the suites' shared
   `kro-system/general-policy` resource names made look even more plausible.

3. **35.4s is not a timeout anyone configured.** It is the 30s `assert` timeout plus
   step overhead, and it is uniform because every failing suite fails on its *first*
   assert, not at some suite-specific point.

## Root cause

`internal/reconciler/elbconfig/` watches `ELBConfig`, and `tests/fixtures/crds/` had
no CRD for it. The reconciler had been dormant, not absent: `chainsaw-start` never
passed an `--enable-elb-cascade` flag, so it was never registered with the manager
and its missing CRD cost nothing. Retiring the flags made every reconciler in
`features.All` start unconditionally, which registered it for the first time.

An informer for a kind the API server does not serve never syncs. controller-runtime
waits out its cache-sync timeout and then fails `mgr.Start()` — killing the whole
manager, not just the one controller:

```
ERROR controller-runtime.source.Kind  if kind is a CRD, it should be installed before calling Start
  {"kind": "ELBConfig.aws.kropath.run",
   "error": "no matches for kind \"ELBConfig\" in version \"aws.kropath.run/v1alpha1\""}
ERROR problem running manager
  {"error": "failed to wait for elbconfig caches to sync kind source: *v1alpha1.ELBConfig:
   timed out waiting for cache to be synced for Kind *v1alpha1.ELBConfig"}
```

The controller log pins it exactly: started `22:34:31`, exited `22:36:47` — 2m16s, the
default cache-sync timeout. So the operator reconciles correctly for two minutes and
then vanishes. Suites that run inside that window pass; every suite after it fails on
a 30s assert timeout against a controller that is no longer running.

`make chainsaw-wait` cannot catch this either: `/readyz` is green for the whole
two-minute window, because the manager genuinely is healthy right up until it isn't.

## Evidence that settled it

The decisive step was checking whether the controller was still alive *after* the run
rather than inspecting the failing suites:

```
$ P=$(cat /tmp/kropath-aws-controller/pid); kill -0 "$P" && echo ALIVE || echo DEAD
DEAD
$ curl -fsS http://127.0.0.1:18081/readyz
curl: (7) Failed to connect to 127.0.0.1 port 18081
```

Once the process was known to be dead, the last 30 lines of `controller.log` named
the kind outright. Everything before that — comparing passing vs failing suites,
mapping namespace usage, auditing shared resource names — was wasted effort spent
inside a false frame.

## Fix

1. **`tests/fixtures/crds/awselbconfig.yaml`** — copied from `kropath-aws/crds/elbconfig.yaml`.

2. **`Makefile`** — the `kubectl wait --for=condition=Established` list is now derived
   from the fixture directory instead of hand-maintained:

   ```make
   CRD_WAIT_TARGETS := $(shell awk '/^  name: [a-z0-9.]+$$/ {print "crd/" $$2}' \
       tests/fixtures/crds/*.yaml | sort -u)
   ```

   The hand-written list had also silently drifted: `apigatewayv2configs`,
   `autoscalingconfigs`, `dynamodbconfigs`, `eventbridgeconfigs`, `rdsconfigs` and
   `secretsmanagerconfigs` were all applied but never waited on.

3. **`TestEveryReconcilerHasCRDFixture`** in `internal/features` — asserts every
   entry in `features.All` has a matching CRD under `tests/fixtures/crds/`. Verified
   it catches the real bug by deleting `awselbconfig.yaml` and confirming failure.
   This converts a two-minute-delayed manager crash into an immediate unit-test
   failure that names the missing file.

## Lessons

- **When per-feature flags are retired, every previously-unflagged reconciler becomes
  a new startup dependency.** Removing a flag does not just change configuration; it
  promotes a dormant component to one that can abort the process. Audit what each
  retired flag was keeping switched off.
- **A missing CRD is fatal to the whole manager, not to one controller.** There is no
  partial-degradation mode. One unregistered kind takes the operator down.
- **A uniform failure duration is a timeout, and a timeout means look at the clock,
  not the assertions.** Failures clustered at 35.4s across unrelated suites cannot be
  eight independent logic bugs.
- **When "which tests fail" changes between machines, the test content is not the
  variable.** Check process liveness and wall-clock ordering first.
- **A health probe proves health now, not for the run.** `/readyz` passing at startup
  says nothing about a cache-sync deadline still ticking in the background.
