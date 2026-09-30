# Container Image Build & Push — Design

**Date:** 2026-07-10
**Status:** Approved

## Purpose

kropath-aws-controller currently has no way to produce a container image. This adds a
Dockerfile, a local `make` workflow for building the image, and a CI job that
validates the build on every PR and pushes to GitHub Container Registry (ghcr.io)
on merges to `main`.

## Components

### Dockerfile

Multi-stage build at the repo root:

- **Stage `builder`**: `golang:1.26.5` (matches `go.mod`). `CGO_ENABLED=0 GOOS=linux`
  static build of `./cmd/manager` → `/kropath-operator`.
- **Stage `runtime`**: `gcr.io/distroless/static:nonroot`. No shell, runs as the
  distroless non-root UID. Copies only the compiled binary from `builder`.
- `ENTRYPOINT ["/kropath-operator"]`. `EXPOSE 8080` (metrics) and `8081` (health
  probes), matching the ports documented in CLAUDE.md.

### .dockerignore

New file excluding `bin/`, `test-results/`, `coverage.out`, `coverage.html`,
`.git`, and `docs/` to keep the build context minimal.

### Makefile additions

New pins alongside the existing version-pin block:

```
IMAGE_REGISTRY := ghcr.io/kropath
IMAGE_NAME     := kropath-aws-controller
IMAGE_TAG      ?= $(shell git rev-parse --short HEAD)
```

New targets:

- `docker-build`: `docker build -t $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) -t $(IMAGE_REGISTRY)/$(IMAGE_NAME):latest .`
- `docker-push`: pushes both the SHA tag and `latest`. Intended for CI use, not
  part of the local dev loop or `make security`.

### CI (`.github/workflows/ci.yaml`)

New `image` job, `needs: [lint, unit, security]` (same gate as the existing `e2e`
job; runs in parallel with it, not blocking on it):

- Checkout, then `docker/setup-buildx-action`.
- **On pull requests**: `docker/build-push-action` with `push: false` — builds
  the image to confirm the Dockerfile is valid, does not publish anything, no
  registry auth needed.
- **On push to `main`**: `docker/login-action` against `ghcr.io` using the
  built-in `GITHUB_TOKEN` (no new secrets required), then
  `docker/build-push-action` with `push: true`, tagging
  `ghcr.io/kropath/kropath-aws-controller:latest` and
  `ghcr.io/kropath/kropath-aws-controller:sha-<short-sha>`.

## Out of Scope

- Multi-arch builds (linux/amd64 only for now). *(Superseded 2026-08-13: CI and
  release now publish `linux/amd64` + `linux/arm64` manifests.)*
- Image signing / SBOM attestation.
- Helm chart or Deployment manifest updates to reference the new image (tracked
  separately if/when needed).

## Testing

- `make docker-build` succeeds locally and produces a runnable image
  (`docker run --rm <image> --help` or equivalent smoke check).
- CI `image` job passes on this PR (build-only path, since it's a PR).
- Manual verification after merge: confirm the `main` push job publishes to
  ghcr.io.
