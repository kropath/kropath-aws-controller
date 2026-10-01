# Frequent Chainsaw Errors — kropath-aws-controller

Recurring, non-obvious failure modes in this repo's Chainsaw suites, in the same
`What Fails / Why / What Works Instead` shape as kropath-aws
`docs/frequent-rgd-errors.md`. That document remains the reference for kro/ACK-specific
problems; this one covers what is specific to the controller's test setup.

**Environment invariant.** The Chainsaw cluster for this repo runs the kropath-operator
against plain config CRDs. There is **no kro and there are no ACK controllers**, so
these objects carry **no finalizers** and trigger no cloud calls. Several rules in
kropath-aws exist purely to work around ACK finalizers and do not transfer here — see
§3.

---

## 1. A Reconciler Whose CRD Is Missing Kills the Whole Manager Two Minutes In

* **What Fails:** a large, arbitrary-looking subset of suites fails, every one at
  ~35s, while the rest pass. Which suites fail changes between machines and between
  runs. Each failing suite passes in isolation. Nothing in the diff touches them.

    ```
    --- PASS: chainsaw/ctrl-apigwv2-01 (8.65s)
    ... 13 more passes ...
    --- FAIL: chainsaw/ctrl-eks-01 (35.40s)
    --- FAIL: chainsaw/ctrl-efs-01-cascade (35.77s)
    ```

* **Why:** every reconciler in `internal/features.All` starts unconditionally — there
  are no per-feature flags. An informer for a kind the API server does not serve never
  syncs, and controller-runtime fails `mgr.Start()` when its cache-sync timeout
  elapses. That takes down the **entire manager**, not the one controller:

    ```
    ERROR problem running manager {"error": "failed to wait for elbconfig caches to sync
      kind source: *v1alpha1.ELBConfig: timed out waiting for cache to be synced"}
    ```

  The default timeout is two minutes, so the operator behaves perfectly for two
  minutes and then exits. Suites finishing inside that window pass; everything after
  fails on a 30s assert timeout against a dead process. Membership of the failing set
  is decided by machine speed and test ordering, not by anything wrong with those
  families. `make chainsaw-wait` cannot catch it — `/readyz` is green for the whole
  window.

* **Diagnosis:** check liveness *after* the run, before reading any assertion diff:

    ```bash
    P=$(cat /tmp/kropath-aws-controller/pid); kill -0 "$P" && echo ALIVE || echo DEAD
    tail -30 /tmp/kropath-aws-controller/controller.log   # names the offending kind
    ```

* **What Works Instead:** ship a CRD under `tests/fixtures/crds/` for every kind any
  reconciler watches. `TestEveryReconcilerHasCRDFixture` enforces this at unit-test
  time, and `make chainsaw-setup` derives its `kubectl wait` list from that directory
  so it cannot drift. Copy the CRD from `kropath-aws/crds/`.

* **Rule:** any uniform failure duration across unrelated suites is a timeout. Look at
  the clock and the process, not the assertions. Reference:
  `docs/troubleshooting-logs/2026-08-13-elbconfig-missing-crd-manager-crash.md`.

---

## 1b. The Same Crash From a Mis-Cased Kind, With Every In-Repo Check Green

* **What Fails:** the §1 symptom exactly — manager dead two minutes in, later suites
  timing out — but `tests/fixtures/crds/` *does* contain a CRD for the kind, and
  `TestEveryReconcilerHasCRDFixture` passes. The log names a kind that plainly has a
  CRD:

    ```
    ERROR ... {"kind": "ApiGatewayConfig.aws.kropath.run",
      "error": "no matches for kind \"ApiGatewayConfig\" in version \"aws.kropath.run/v1alpha1\""}
    ```

* **Why:** the CRD's *plural* and the CRD's *Kind* are independent. §1's guard keys off
  the metadata name (`apigatewayconfigs.aws.kropath.run`), which is lowercase and
  therefore identical whichever way the Kind is cased. The Kind is what the informer
  matches on: controller-runtime derives it from the Go type name registered in
  `api/v1alpha1/register.go`, and the API server serves whatever `spec.names.kind`
  says. `ApiGatewayConfig` vs `APIGatewayConfig` differ by one letter's case, match
  nothing, and produce a "CRD not installed" error for a CRD that is installed.

  The fixtures are hand-maintained and deliberately **not** verbatim copies of
  `kropath-aws/crds/`, so the registry, the scheme and the fixture can all agree with
  each other and still disagree with the CRD a real cluster serves. Every check in
  this repo stayed green while the operator crash-looped in the integration cluster.

* **Diagnosis:** compare the two spellings directly — they will look identical at a
  glance, so diff them rather than eyeballing:

    ```bash
    kubectl get crd apigatewayconfigs.aws.kropath.run -o jsonpath='{.spec.names.kind}'
    grep -rn 'cascade("' internal/features/features.go | grep -i apigateway
    ```

* **What Works Instead:** `kropath-aws` owns the CRDs, so it is the authority on every
  Kind name. `make crds-verify` fetches them and fails on any watched Kind that is
  missing or mis-cased; CI runs it in the *Feature registry drift gate* job.
  `TestReconcilerKindMatchesFixtureAndScheme` covers the in-repo half offline —
  registry Kind, fixture `spec.names.kind`, and the Kind the scheme registers must
  agree exactly.

* **Rule:** "no matches for kind X" when a CRD for X exists means the *casing*
  differs, not the CRD. Never rename a Kind to match the Go type — rename the Go type
  to match `kropath-aws/crds/`. Reference: KRO-675.

---

## 2. Reusing One Resource Name Across Steps Forces Reset Deletes

* **What Fails:** a suite needs interleaved `delete` steps between acceptance criteria
  to stop the previous step's values leaking into the next. Symptoms when a reset is
  missing or mis-ordered: a level-1 value shadowing the level-3 value a later step
  means to test, or an SSA-PATCH retry loop when Chainsaw re-applies to an object it
  created earlier with a different field set.

* **Why:** the controller resolves a config's global counterpart **by name** —
  `payments-prod/X` merges only `kro-system/X`. Sharing one name (`general-policy`)
  across every step makes all steps collide on one object. Chainsaw also defers
  `cleanup:` blocks to the **end of the file** in LIFO order, not after each step, so
  any mid-suite reset has to be an inline `try:` step.

* **What Works Instead:** give each step its own name (`ac<N>-<family>-policy`) and
  delete nothing between steps. A per-step name is a complete isolation boundary in
  both directions, regardless of order. Layout: one `ac<NN>-setup.yaml` holding that
  step's inputs as a multi-document YAML, plus one `ac<NN>-assert.yaml`.

* **Payoff:** `tests/eks` went 7.8s → 3.1s cold and lost 16 inline deletes; across the
  14 converted suites the full run went 205s → 179s and 439 per-step deletes were
  removed. Reference:
  `docs/troubleshooting-logs/2026-08-13-chainsaw-unique-name-rollout.md`.

---

## 3. `skipDelete: true` Does Not Transfer From kropath-aws

* **What Fails:** copying `spec.skipDelete: true` from the kropath-aws canonical
  pattern. Later suites in the run start failing with
  `client rate limiter Wait returned an error: would exceed context deadline`.

* **Why:** in kropath-aws, `skipDelete` avoids cleanup-phase timeouts caused by ACK
  finalizers that no controller ever removes. **This repo has no finalizers**, so
  there is no deletion queue to back up and nothing to avoid. What it does instead is
  leave every suite's objects in the shared kind cluster for the whole run, and the
  resulting steady reconcile load on the API server is enough to starve the test
  client later on.

* **What Works Instead:** let Chainsaw do its normal end-of-test cleanup. It is one
  bulk pass over finalizer-free objects and costs very little.

* **Rule:** before importing a rule from `frequent-rgd-errors.md`, check whether it
  depends on ACK finalizers or kro. Most of the delete-related ones do.

---

## 4. Dropping an `expect:` Block Surfaces as a Rate-Limiter Error

* **What Fails:** a negative-path step — one applying a deliberately invalid manifest
  to prove the API server rejects it — reports

    ```
    client rate limiter Wait returned an error: rate: Wait(n=1) would exceed context deadline
    ```

  which reads as an infrastructure or throughput problem and is nothing of the kind.

* **Why:** the apply is *meant* to fail. Its `expect:` block is what converts the
  rejection into a pass:

    ```yaml
    - apply:
        file: ac07-setup-1.yaml
        expect:
          - match: {apiVersion: aws.kropath.run/v1alpha1, kind: KropathConfig}
            check: {($error != null): true}
    ```

  Lose that block — easily done when merging several applies into one setup file,
  since `expect:` is per-manifest — and the step fails, Chainsaw retries the apply,
  and the retry loop exhausts the operation deadline. The rate limiter is the last
  thing to complain, so it is the only thing you see. The real validation error only
  appears if the deadline is long enough for it to win the race.

* **What Works Instead:** never merge an apply that carries options beyond `file:`.
  Give it its own `ac<NN>-setup-<i>.yaml`.

* **Rule:** treat `client rate limiter ... context deadline exceeded` as "some
  operation is retrying in a loop", not as "the cluster is too slow". Find the
  operation. Raising `--apply-timeout` to make it go away is a false fix — it was
  tried, fixed nothing, and doubled the run to 414s.

---

## 5. A Reset Script Must Be Deleted, Never Renamed

* **What Fails:** after a rename-based refactor, a step asserts a value the cascade
  should produce and gets the zero value instead — e.g. `flowLogsRequired: false`
  where `true` was expected, with the input config plainly present in the setup file.

* **Why:** the step contained a reset script, and the rename rewrote its target:

    ```yaml
    # before — clears the previous step's leftovers
    content: kubectl delete kropathconfigs.aws.kropath.run general-policy -n kro-system --ignore-not-found=true
    # after — deletes the config this step just applied
    content: kubectl delete kropathconfigs.aws.kropath.run ac11-general-policy -n kro-system --ignore-not-found=true
    ```

* **What Works Instead:** drop reset scripts entirely, like inline `delete:` steps —
  unique names make them unnecessary. A `script:` whose every non-comment line starts
  with `kubectl delete` is a reset, never an assertion; anything else (a `jq` check, a
  metrics poll, an apply-and-check) is real and must be kept, with its names renamed
  to match.

---

## 6. Order-Stable Assertions

**Never assert a list/array field by exact position when its element order is not
guaranteed deterministic.**

Any list built by iterating a Go map has iteration order that is not stable across
runs, and a positional Chainsaw `assert` on one is a latent flake — it can pass locally
and fail in CI purely from map-order nondeterminism.

This repo's `effectiveConfig.mandatory.tags` / `.defaults.tags` / `syncedLabels` /
`syncedAnnotations` are all **maps**, so today's asserts are order-stable and need no
change. `PolicyDocument` statement merges are stable too — `spec.sources` are
concatenated in source order, not ranged over a map.

If a future change asserts a **list built from a map**, note that per-item `(?...)`
matching is *also* positional in Chainsaw and does not fix this, and `.exists()` is not
implemented by its assertion engine. Use a `- script:` with `kubectl ... -o json | jq`,
which is genuinely order-independent. See `frequent-rgd-errors.md` §6 "Flaky List/Array
Asserts" for the full writeup and the two false fixes to avoid.
