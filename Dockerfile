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

# Pin the builder to the *build* platform and let Go cross-compile to the target.
# Go cross-compiles natively, so this is far faster than emulating the target
# platform under QEMU.
FROM --platform=$BUILDPLATFORM golang:1.27.1 AS builder

# Supplied automatically by BuildKit; defaults keep a plain `docker build` working.
ARG TARGETOS=linux
ARG TARGETARCH

ARG VERSION=dev
ARG GIT_COMMIT=none
ARG BUILD_DATE=unknown

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY api/ api/
COPY cmd/ cmd/
COPY internal/ internal/

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags "-s -w \
        -X github.com/kropath/kropath-aws-controller/internal/version.Version=${VERSION} \
        -X github.com/kropath/kropath-aws-controller/internal/version.GitCommit=${GIT_COMMIT} \
        -X github.com/kropath/kropath-aws-controller/internal/version.BuildDate=${BUILD_DATE}" \
    -o /kropath-operator ./cmd/manager

FROM gcr.io/distroless/static:nonroot

# Re-declare build-args so they are in scope for LABEL in this stage.
ARG VERSION=dev
ARG GIT_COMMIT=none
ARG BUILD_DATE=unknown

LABEL org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${GIT_COMMIT}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.source="https://github.com/kropath/kropath-aws-controller"

WORKDIR /

COPY --from=builder /kropath-operator /kropath-operator
COPY docs/features.yaml /features.yaml

USER 65532:65532

EXPOSE 8080 8081

ENTRYPOINT ["/kropath-operator"]
