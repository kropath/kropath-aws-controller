# Container Image Build & Push Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a Dockerfile for kropath-aws-controller, a `make` workflow to build/push the image locally, and a CI job that build-validates the image on every PR and pushes it to GitHub Container Registry (ghcr.io) on merges to `main`.

**Architecture:** A multi-stage Dockerfile compiles `./cmd/manager` in a `golang:1.26.5` builder stage and copies the static binary into a `gcr.io/distroless/static:nonroot` runtime stage. The Makefile gets two new targets (`docker-build`, `docker-push`) using pinned image-name/tag variables. CI adds an `image` job (parallel to the existing `e2e` job, gated on `lint`/`unit`/`security`) that runs `docker/build-push-action`, pushing only on `push` events to `main`.

**Tech Stack:** Docker (multi-stage build), Make, GitHub Actions (`docker/setup-buildx-action`, `docker/login-action`, `docker/build-push-action`), GitHub Container Registry.

## Global Constraints

- Go version for the builder stage: `1.26.5` (from `go.mod`, must match exactly — see `CLAUDE.md` version-pin convention).
- Runtime base image: `gcr.io/distroless/static:nonroot` (no shell, non-root by default) — per `docs/STANDARDS.md` "Security First: Default to secure."
- Registry: `ghcr.io/kropath`, image name `kropath-aws-controller`.
- Tags: `latest` and `sha-<short-sha>` on push to `main`; no tags pushed on PRs (build-only).
- Auth: built-in `GITHUB_TOKEN` via `docker/login-action` — no new repo secrets.
- Binary entrypoint ports: `8080` (metrics), `8081` (health probes) — from `cmd/manager/main.go` flag defaults (`-metrics-bind-address=:8080`, `-health-probe-bind-address=:8081`).
- Module path: `github.com/kropath/kropath-aws-controller`; build target: `./cmd/manager`.
- Version pins for CI tooling live in both `Makefile` (top block) and `.github/workflows/ci.yaml` (`env:` block) and must stay in sync per the existing repo convention.

---

### Task 1: Dockerfile and .dockerignore

**Files:**
- Create: `Dockerfile`
- Create: `.dockerignore`

**Interfaces:**
- Produces: an image buildable via `docker build -t <tag> .` that runs `/kropath-operator` as entrypoint, listening on `:8080` and `:8081` when passed matching flags. Task 2 (Makefile targets) invokes this Dockerfile by plain `docker build .` — no build args required.

- [ ] **Step 1: Write `.dockerignore`**

```
.git
bin/
test-results/
coverage.out
coverage.html
docs/
*.md
```

- [ ] **Step 2: Write `Dockerfile`**

```dockerfile
# Copyright 2026 kropath Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.26.5 AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY api/ api/
COPY cmd/ cmd/
COPY internal/ internal/

RUN CGO_ENABLED=0 GOOS=linux go build -o /kropath-operator ./cmd/manager

FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY --from=builder /kropath-operator /kropath-operator

USER 65532:65532

EXPOSE 8080 8081

ENTRYPOINT ["/kropath-operator"]
```

- [ ] **Step 3: Build the image locally**

Run: `docker build -t kropath-aws-controller:test .`
Expected: build completes successfully, final line is `Successfully tagged kropath-aws-controller:test` (or buildkit's equivalent `naming to docker.io/library/kropath-aws-controller:test done`).

- [ ] **Step 4: Smoke-test the image**

Run: `docker run --rm kropath-aws-controller:test --help`
Expected: prints flag usage (including `-metrics-bind-address`, `-health-probe-bind-address`, `-enable-poldoc`, `-enable-kms-cascade`) and exits with status `0`. Verify exit code with `echo $?` immediately after — expect `0`.

- [ ] **Step 5: Commit**

```bash
git add Dockerfile .dockerignore
git commit -m "feat: add Dockerfile for kropath-aws-controller image"
```

---

### Task 2: Makefile `docker-build` / `docker-push` targets

**Files:**
- Modify: `Makefile` (add to the `─── Paths ───` pin block and add a new `─── Container image ───` section after the `build:` target)

**Interfaces:**
- Consumes: `Dockerfile` at repo root (Task 1).
- Produces: `make docker-build` (builds `$(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)` and `:latest`), `make docker-push` (pushes both tags). Task 3 (CI) does not call these targets directly — it uses `docker/build-push-action` with the same registry/name/tag values, kept in sync by convention documented in Global Constraints.

- [ ] **Step 1: Add image variables to the pin/paths block**

In `Makefile`, after the existing `BINARY`/`MAIN_PKG` lines in the `# ─── Paths ───` section:

```makefile
IMAGE_REGISTRY   := ghcr.io/kropath
IMAGE_NAME       := kropath-aws-controller
IMAGE_TAG        ?= $(shell git rev-parse --short HEAD)
```

- [ ] **Step 2: Add `docker-build` and `docker-push` to `.PHONY`**

Change:

```makefile
.PHONY: all build test test-cover vet fmt lint \
        kind-up kind-down \
        chainsaw-setup chainsaw-start chainsaw-wait chainsaw-stop \
        test-iam test-s3 test-kms test-policy test-chainsaw \
        install-tools gosec vulncheck security \
        help default
```

to:

```makefile
.PHONY: all build test test-cover vet fmt lint \
        docker-build docker-push \
        kind-up kind-down \
        chainsaw-setup chainsaw-start chainsaw-wait chainsaw-stop \
        test-iam test-s3 test-kms test-policy test-chainsaw \
        install-tools gosec vulncheck security \
        help default
```

- [ ] **Step 3: Add the container-image section**

Insert immediately after the existing `# ─── Build ───` section (after the `build:` target's recipe, before `# ─── Unit tests ───`):

```makefile
# ─── Container image ────────────────────────────────────────────────────────

docker-build: ## Build the container image, tagged with the short git SHA and 'latest'.
	docker build -t $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) \
		-t $(IMAGE_REGISTRY)/$(IMAGE_NAME):latest .

docker-push: ## Push the SHA-tagged and 'latest' images (CI use; requires prior registry login).
	docker push $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
	docker push $(IMAGE_REGISTRY)/$(IMAGE_NAME):latest
```

- [ ] **Step 4: Verify `make help` lists the new targets**

Run: `make help`
Expected: output includes lines for `docker-build` and `docker-push` with their descriptions, e.g.:

```
  docker-build             Build the container image, tagged with the short git SHA and 'latest'.
  docker-push              Push the SHA-tagged and 'latest' images (CI use; requires prior registry login).
```

- [ ] **Step 5: Run `make docker-build` and confirm the image tags**

Run: `make docker-build`
Expected: exits `0`. Then run `docker images ghcr.io/kropath/kropath-aws-controller` and confirm two rows, tagged with the current short SHA (`git rev-parse --short HEAD`) and `latest`.

- [ ] **Step 6: Commit**

```bash
git add Makefile
git commit -m "feat: add docker-build and docker-push Makefile targets"
```

---

### Task 3: CI `image` job

**Files:**
- Modify: `.github/workflows/ci.yaml` (add a new `image` job after the existing `security` job, before `e2e`)

**Interfaces:**
- Consumes: `Dockerfile` (Task 1) at repo root; registry/name values from Global Constraints (kept in sync with Task 2's Makefile variables by convention, not by shared code).
- Produces: on `pull_request` events, a build-only CI check named `Build image`. On `push` to `main`, publishes `ghcr.io/kropath/kropath-aws-controller:latest` and `ghcr.io/kropath/kropath-aws-controller:sha-<short-sha>`.

- [ ] **Step 1: Add the `image` job**

In `.github/workflows/ci.yaml`, insert a new job after `security:` and before `e2e:`:

```yaml
  image:
    name: Build image
    runs-on: ubuntu-latest
    needs: [lint, unit, security]
    permissions:
      contents: read
      packages: write

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to GHCR
        if: github.event_name == 'push' && github.ref == 'refs/heads/main'
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Build and push
        uses: docker/build-push-action@v6
        with:
          context: .
          push: ${{ github.event_name == 'push' && github.ref == 'refs/heads/main' }}
          tags: |
            ghcr.io/kropath/kropath-aws-controller:latest
            ghcr.io/kropath/kropath-aws-controller:sha-${{ github.sha }}
```

- [ ] **Step 2: Validate workflow YAML syntax**

Run: `python3 -c "import yaml, sys; yaml.safe_load(open('.github/workflows/ci.yaml'))" && echo OK`
Expected: prints `OK` with no exception.

- [ ] **Step 3: Validate with `actionlint` if available, otherwise GitHub's workflow lint**

Run: `command -v actionlint >/dev/null 2>&1 && actionlint .github/workflows/ci.yaml || echo "actionlint not installed, skipping local lint"`
Expected: either no output (actionlint found zero issues) or the skip message — no YAML/schema errors reported.

- [ ] **Step 4: Commit**

```bash
git add .github/workflows/ci.yaml
git commit -m "ci: build and push container image to ghcr.io"
```

- [ ] **Step 5: Push the branch and open/update the PR to observe the `Build image` check run**

Run: `git push -u origin HEAD`
Expected: push succeeds; on GitHub, the PR's checks list now includes `Build image`, which runs `docker/build-push-action` with `push: false` (since this event is a `pull_request`, not a `push` to `main`) and completes successfully.

---

## Self-Review Notes

- **Spec coverage:** Dockerfile (Task 1) ✅, `.dockerignore` (Task 1) ✅, Makefile targets (Task 2) ✅, CI build-on-PR/push-on-main job (Task 3) ✅, ghcr.io + GITHUB_TOKEN auth (Task 3) ✅, image tags `latest`/sha (Task 2 + Task 3) ✅. Multi-arch, signing, and manifest updates are explicitly out of scope per the spec and are not tasked here.
- **Type/naming consistency:** `IMAGE_REGISTRY=ghcr.io/kropath`, `IMAGE_NAME=kropath-aws-controller` used identically in Task 2 (Makefile) and Task 3 (CI tags), matching the spec's Global Constraints.
- **No placeholders:** all code blocks are complete and copy-pasteable; no TBD/TODO markers.
