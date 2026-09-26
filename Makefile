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

# ─── Version pins — the single source of truth ─────────────────────────────────
#
# The Go version is read from go.mod via setup-go. Every other tool version is
# pinned here and *only* here. ci.yaml installs through the `install-*` targets
# below and reads `print-tool-versions` for the one tool it takes as a prebuilt
# binary, so there is no second copy of these numbers to keep in sync and CI
# cannot silently run a different golangci-lint from a local `make lint`.
#
# gosec and govulncheck previously floated on `@latest` in CI, which meant a
# security scan could start failing on a tree that had not changed. They are
# pinned to what `@latest` resolved to when this pin was introduced, so the
# change is a freeze, not an upgrade.
KIND_VERSION        := v0.25.0
GOLANGCI_VERSION    := v2.11.4
CHAINSAW_VERSION    := v0.2.15
GOSEC_VERSION       := v2.29.0
GOVULNCHECK_VERSION := v1.8.0
GOIMPORTS_VERSION   := v0.50.0

# ─── Paths ─────────────────────────────────────────────────────────────────────
BINARY           := bin/kropath-operator
MAIN_PKG         := ./cmd/manager
IMAGE_REGISTRY   := ghcr.io/kropath
IMAGE_NAME       := kropath-controller
IMAGE_TAG        ?= $(shell git rev-parse --short=7 HEAD)
REPORT_DIR       := test-results
CONTROLLER_LOG   := /tmp/kropath-controller/controller.log
CONTROLLER_PID   := /tmp/kropath-controller/pid

# ─── Version stamping ──────────────────────────────────────────────────────────
VERSION    ?= dev
GIT_COMMIT := $(shell git rev-parse --short=7 HEAD 2>/dev/null || echo none)
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
MODULE     := github.com/kropath/kropath-controller
LDFLAGS    := -s -w \
    -X $(MODULE)/internal/version.Version=$(VERSION) \
    -X $(MODULE)/internal/version.GitCommit=$(GIT_COMMIT) \
    -X "$(MODULE)/internal/version.BuildDate=$(BUILD_DATE)"

# ─── Test config ───────────────────────────────────────────────────────────────
KIND_CLUSTER     := kropath-controller-test
HEALTH_PORT      ?= 18081
METRICS_PORT     ?= 18080
TEST_NAMESPACES  := kro-system payments-prod events-prod network-prod registry-prod data-platform
CHAINSAW         ?= chainsaw
GOLANGCI         ?= golangci-lint

# Git ref of kropath-aws whose CRDs are the authority for crds-verify.
KROPATH_AWS_REF  ?= main
# Optional space-separated list of companion refs (e.g. CRD PR branches not yet
# on main). crds-verify fetches CRDs from every listed ref, so multiple parallel
# controller+CRD PR pairs can be reviewed together before any of them merges.
# Each ref is silently skipped if it does not exist (e.g. after branch deletion).
KROPATH_AWS_COMPANION_REFS ?=

# --skip-delete: Chainsaw otherwise deletes every resource a step created at the
# end of that step, and blocks on it. Since the KRO-1137 namespace-pair migration
# each acceptance criterion creates its own Namespace pair, so a full run spent
# 310 namespace deletions x ~5.3s = 27.4 min of its 32 min waiting on Kubernetes
# namespace termination -- 86% of the job, against 19s of actual assertions
# (KRO-1154). The CRs themselves delete in ~0s; it is purely the namespaces.
#
# Skipping deletion is safe because suite isolation comes from unique-name-per-AC
# (see any suite header), not from cleanup, and `test-chainsaw` builds and throws
# away its kind cluster inside a single run. Iterating locally against a
# long-lived cluster accumulates namespaces across runs; `make kind-down` resets.
CHAINSAW_FLAGS   := --parallel 1 --skip-delete --report-format JUNIT-TEST --report-path $(REPORT_DIR)/

.PHONY: all build test test-cover vet fmt lint \
        features-gen features-verify crds-verify \
        docker-build docker-push \
        kind-up kind-down \
        chainsaw-setup chainsaw-start chainsaw-wait chainsaw-stop \
        test-iam test-s3 test-kms test-policy test-label-operator \
        test-apigatewayv2 test-autoscaling test-cwl test-dynamodb test-ec2 \
        test-ecr test-ecs test-efs test-eks test-elasticache test-eventbridge test-bedrock \
        test-rds test-secretsmanager test-sns test-sqs test-stepfunctions \
        test-version test-features \
        test-dyn-01 test-dyn-02 test-dyn-03 test-dyn \
        test-organizations \
        test-s3advanced \
        test-chainsaw \
        install-tools install-kind install-chainsaw install-golangci-lint \
        install-goimports install-gosec install-govulncheck \
        print-tool-versions gosec vulncheck security \
        help default

default: help

all: build ## Build the operator binary (default target).

# ─── Build ─────────────────────────────────────────────────────────────────────

build: ## Compile the operator binary → bin/kropath-operator (with version ldflags).
	@mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(MAIN_PKG)

# ─── Feature registry ──────────────────────────────────────────────────────────

features-gen: ## Regenerate docs/features.yaml from internal/features.All.
	go run ./cmd/gen-features

features-verify: ## CI gate: fail if docs/features.yaml is stale (run make features-gen to fix).
	go run ./cmd/gen-features --check

# ─── Cross-repo CRD gate ───────────────────────────────────────────────────────

# kropath-aws owns the CRDs this controller watches (docs/STANDARDS.md), so it is
# the authority on every Kind name. Nothing inside this repo can catch a Kind that
# drifted from it: tests/fixtures/crds/ is hand-maintained, so the registry, the
# scheme and the fixture can all agree with each other and still disagree with the
# CRD a real cluster serves. That is how KRO-675 shipped — the controller watched
# "ApiGatewayConfig" while the CRD served "APIGatewayConfig", the plural matched,
# every in-repo check stayed green, and the manager crash-looped in the
# integration cluster at the 2-minute cache-sync timeout.
#
# Set KROPATH_AWS_CRDS_DIR to check against a local checkout instead of fetching.
crds-verify: ## CI gate: fail if a watched Kind is missing or mis-cased vs kropath-aws CRDs.
	@set -eu; \
	dir="$${KROPATH_AWS_CRDS_DIR:-}"; \
	if [ -n "$$dir" ]; then \
	  echo "Checking watched Kinds against $$dir"; \
	else \
	  dir=$$(mktemp -d); \
	  trap 'rm -rf "$$dir"' EXIT; \
	  echo "Fetching kropath-aws CRDs @ $(KROPATH_AWS_REF)"; \
	  for path in crds crds/policy; do \
	    gh api "repos/kropath/kropath-aws/contents/$$path?ref=$(KROPATH_AWS_REF)" \
	      --jq '.[] | select(.type == "file") | select(.name | endswith(".yaml")) | .name' \
	    | while read -r name; do \
	        _done=0; \
	        for _i in 1 2 3 4 5; do \
	          _tmpf=$$(mktemp); \
	          gh api "repos/kropath/kropath-aws/contents/$$path/$$name?ref=$(KROPATH_AWS_REF)" \
	            --jq .content 2>/dev/null | base64 -d > "$$_tmpf" 2>/dev/null || true; \
	          if test -s "$$_tmpf"; then \
	            mv "$$_tmpf" "$$dir/$$name"; _done=1; break; \
	          fi; \
	          rm -f "$$_tmpf"; \
	          [ $$_i -lt 5 ] && sleep 5 || true; \
	        done; \
	        [ $$_done -eq 1 ]; \
	      done; \
	  done; \
	  for companion_ref in $(KROPATH_AWS_COMPANION_REFS); do \
	    echo "Fetching companion kropath-aws CRDs @ $$companion_ref"; \
	    for path in crds crds/policy; do \
	      gh api "repos/kropath/kropath-aws/contents/$$path?ref=$$companion_ref" \
	        --jq '.[] | select(.type == "file") | select(.name | endswith(".yaml")) | .name' \
	      2>/dev/null \
	      | while read -r name; do \
	          _tmpf=$$(mktemp); \
	          gh api "repos/kropath/kropath-aws/contents/$$path/$$name?ref=$$companion_ref" \
	            --jq .content 2>/dev/null | base64 -d > "$$_tmpf" 2>/dev/null \
	          && test -s "$$_tmpf" && mv "$$_tmpf" "$$dir/$$name" \
	          || rm -f "$$_tmpf"; \
	        done; \
	    done || true; \
	  done; \
	fi; \
	KROPATH_AWS_CRDS_DIR="$$dir" go test ./internal/features/ -run 'TestWatchedKindsMatchUpstreamCRDs|TestSpecDefaultDrift|TestConfigKindSetDrift' -v

# ─── Container image ────────────────────────────────────────────────────────────

# Single-architecture (the host's) on purpose: buildx cannot --load a multi-platform
# image into the local daemon. CI publishes linux/amd64 + linux/arm64 manifests.
docker-build: ## Build the container image for the host architecture (version-stamped), tagged with the short git SHA and 'latest'.
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE="$(BUILD_DATE)" \
		-t $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) \
		-t $(IMAGE_REGISTRY)/$(IMAGE_NAME):latest .

docker-push: ## Push the SHA-tagged and 'latest' images (CI use; requires prior registry login).
	docker push $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
	docker push $(IMAGE_REGISTRY)/$(IMAGE_NAME):latest

# ─── Unit tests ────────────────────────────────────────────────────────────────

test: ## Run unit tests with race detection.
	go test -race ./...

test-cover: ## Run unit tests with race detection and produce an HTML coverage report.
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# ─── Code quality ──────────────────────────────────────────────────────────────

vet: ## Run go vet over all packages.
	go vet ./...

fmt: ## Format Go source files with gofmt and goimports.
	gofmt -w .
	@goimports -w . 2>/dev/null || \
		echo "goimports not installed; run: go install golang.org/x/tools/cmd/goimports@latest"

lint: vet features-verify ## Run golangci-lint (required gate before every commit).
	$(GOLANGCI) run ./...

# ─── Kind cluster ──────────────────────────────────────────────────────────────

kind-up: ## Create the kind cluster (no-op if it already exists).
	@command -v kind >/dev/null 2>&1 || \
		{ echo "ERROR: kind not found. Run: make install-tools"; exit 1; }
	@if kind get clusters 2>/dev/null | grep -q "^$(KIND_CLUSTER)$$"; then \
		echo "Kind cluster '$(KIND_CLUSTER)' already exists."; \
	else \
		kind create cluster --name $(KIND_CLUSTER) --wait 120s; \
	fi

kind-down: ## Delete the kind cluster.
	kind delete cluster --name $(KIND_CLUSTER)

# ─── Chainsaw integration-test helpers ────────────────────────────────────────
#
# Typical workflow:
#   make test-chainsaw          # end-to-end: kind-up → build → CRDs → start → run all → stop
#
# To iterate on a single suite without restarting the controller:
#   make chainsaw-start chainsaw-wait
#   make test-kms
#   make chainsaw-stop
#
# On failure mid-run, clean up manually: make chainsaw-stop kind-down

# Every CRD shipped under tests/fixtures/crds/, as `crd/<metadata.name>` arguments.
# Derived from the fixture files rather than hand-listed: a kind that a reconciler
# watches but whose CRD is absent from the cluster makes controller-runtime's informer
# never sync, and the manager exits with a fatal error after the 2-minute cache-sync
# timeout. That surfaces as every Chainsaw suite after the ~2-minute mark timing out,
# which is confusing to diagnose and has nothing to do with those suites. Deriving the
# list keeps it from drifting out of sync with the fixtures again (KRO-635).
#
# tests/fixtures/crds-optional/ is intentionally excluded: those CRDs must be absent
# at operator startup for dynamic-detection suites. Suites install them on demand via
# chainsaw-install-optional-crd.
CRD_WAIT_TARGETS := $(shell awk '/^  name: [a-z0-9.]+$$/ {print "crd/" $$2}' tests/fixtures/crds/*.yaml | sort -u)

chainsaw-setup: build kind-up ## Create kind cluster, apply CRDs, create test namespaces.
	kubectl apply -f tests/fixtures/crds/
	kubectl wait --for=condition=Established --timeout=60s $(CRD_WAIT_TARGETS)
	@for ns in $(TEST_NAMESPACES); do \
		kubectl create namespace "$$ns" --dry-run=client -o yaml | kubectl apply -f -; \
	done
	# kro-system is the shared global-tier namespace itself, so it carries no
	# global-config-namespace annotation. Every other shared namespace is a
	# resource namespace pointing at it, and since KRO-1128 that role also
	# requires the ACK-native account/region annotations util.ResolvePlacement
	# reads (spec §4.4) -- without them these namespaces are RoleGovernanceOnly
	# and every <ResourceFamily>Config inside them gets no effectiveConfig.
	@for ns in $(filter-out kro-system,$(TEST_NAMESPACES)); do \
		kubectl annotate namespace "$$ns" --overwrite \
			aws.kropath.run/global-config-namespace=kro-system \
			services.k8s.aws/owner-account-id=123456789012 \
			services.k8s.aws/default-region=us-east-1; \
	done

# Install one optional CRD (from tests/fixtures/crds-optional/) and wait for Established.
# Usage: make chainsaw-install-optional-crd CRD_FILE=tests/fixtures/crds-optional/<name>.yaml
chainsaw-install-optional-crd: ## Install a single optional CRD and wait for Established (used by dynamic-detection suites).
	@if [ -z "$(CRD_FILE)" ]; then echo "Usage: make chainsaw-install-optional-crd CRD_FILE=<path>"; exit 1; fi
	kubectl apply -f $(CRD_FILE)
	kubectl wait --for=condition=Established --timeout=60s \
		$$(awk '/^  name: [a-z0-9.]+$$/ {print "crd/" $$2}' $(CRD_FILE) | sort -u)

chainsaw-start: chainsaw-setup ## Build, set up CRDs, and start the operator in the background.
	@mkdir -p /tmp/kropath-controller
	@if [ -f $(CONTROLLER_PID) ] && kill -0 "$$(cat $(CONTROLLER_PID))" 2>/dev/null; then \
		echo "Controller already running (PID $$(cat $(CONTROLLER_PID)))."; \
	else \
		KUBECONFIG="$${HOME}/.kube/config" $(BINARY) \
			--health-probe-bind-address=127.0.0.1:$(HEALTH_PORT) \
			--metrics-bind-address=127.0.0.1:$(METRICS_PORT) \
			> $(CONTROLLER_LOG) 2>&1 & echo $$! > $(CONTROLLER_PID); \
		echo "Controller started (PID $$(cat $(CONTROLLER_PID)))."; \
	fi

chainsaw-wait: ## Block until the controller /readyz endpoint responds stably (up to 60 s).
	@echo "Waiting for /readyz on :$(HEALTH_PORT)..."
	@for i in $$(seq 1 30); do \
		if curl -fsS http://127.0.0.1:$(HEALTH_PORT)/readyz >/dev/null 2>&1; then \
			sleep 2; \
			if curl -fsS http://127.0.0.1:$(HEALTH_PORT)/readyz >/dev/null 2>&1; then \
				echo "Controller is ready."; exit 0; \
			fi; \
		fi; \
		sleep 2; \
	done; \
	echo "ERROR: controller did not become ready. Log:"; \
	cat $(CONTROLLER_LOG); \
	exit 1

chainsaw-stop: ## Stop the background controller process.
	@if [ -f $(CONTROLLER_PID) ]; then \
		kill "$$(cat $(CONTROLLER_PID))" 2>/dev/null && echo "Controller stopped." || true; \
		rm -f $(CONTROLLER_PID); \
	else \
		echo "No controller PID file found (already stopped?)."; \
	fi

# ─── Chainsaw individual suites ────────────────────────────────────────────────
# These targets assume the controller is already running (chainsaw-start + chainsaw-wait).

test-iam: ## Run IAM cascade Chainsaw suite (ctrl-iam-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/iam/ctrl-iam-01/ $(CHAINSAW_FLAGS)

test-s3: ## Run S3 cascade Chainsaw suite (ctrl-s3-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/s3/ctrl-s3-01/ $(CHAINSAW_FLAGS)

test-kms: ## Run KMS cascade Chainsaw suites (ctrl-kms-01, ctrl-kms-02).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/kms/ctrl-kms-01/ tests/kms/ctrl-kms-02/ $(CHAINSAW_FLAGS)

test-policy: ## Run policy document Chainsaw suites (phase2-refs + phase3-merge).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/policy/phase2-refs/ tests/policy/phase3-merge/ $(CHAINSAW_FLAGS)

test-label-operator: ## Run label-operator Chainsaw suite (ctrl-label-op-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/label-operator/ctrl-label-op-01/ $(CHAINSAW_FLAGS)

test-apigatewayv2: ## Run API Gateway v2 cascade Chainsaw suite (ctrl-apigwv2-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/apigatewayv2/ctrl-apigwv2-01/ $(CHAINSAW_FLAGS)

test-autoscaling: ## Run Auto Scaling cascade Chainsaw suite (ctrl-autoscaling-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/autoscaling/ctrl-autoscaling-01/ $(CHAINSAW_FLAGS)

test-cwl: ## Run CloudWatch Logs cascade Chainsaw suite (ctrl-cwl-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/cloudwatchlogs/ctrl-cwl-01/ $(CHAINSAW_FLAGS)

test-dynamodb: ## Run DynamoDB cascade Chainsaw suite (ctrl-dynamodb-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/dynamodb/ctrl-dynamodb-01/ $(CHAINSAW_FLAGS)

test-ec2: ## Run EC2 cascade Chainsaw suite (ctrl-ec2-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/ec2/ctrl-ec2-01/ $(CHAINSAW_FLAGS)

test-ecr: ## Run ECR cascade Chainsaw suite (ctrl-ecr-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/ecr/ctrl-ecr-01/ $(CHAINSAW_FLAGS)

test-ecs: ## Run ECS cascade Chainsaw suite (ctrl-ecs-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/ecs/ctrl-ecs-01/ $(CHAINSAW_FLAGS)

test-efs: ## Run EFS cascade Chainsaw suite (ctrl-efs-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/efs/ctrl-efs-01/ $(CHAINSAW_FLAGS)

test-bedrock: ## Run Bedrock cascade Chainsaw suite (ctrl-bedrock-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/bedrock/ctrl-bedrock-01/ $(CHAINSAW_FLAGS)

test-eks: ## Run EKS cascade Chainsaw suite (ctrl-eks-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/eks/ctrl-eks-01/ $(CHAINSAW_FLAGS)

test-organizations: ## Run Organizations cascade Chainsaw suite (ctrl-org-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/organizations/controller/ctrl-org-01/ $(CHAINSAW_FLAGS)

test-s3advanced: ## Run S3 Advanced cascade Chainsaw suite (ctrl-s3advanced-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/s3advanced/ctrl-s3advanced-01/ $(CHAINSAW_FLAGS)

test-elasticache: ## Run ElastiCache cascade Chainsaw suite (ctrl-elasticache-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/elasticache/ctrl-elasticache-01/ $(CHAINSAW_FLAGS)

test-eventbridge: ## Run EventBridge cascade Chainsaw suite (ctrl-eventbridge-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/eventbridge/ctrl-eventbridge-01/ $(CHAINSAW_FLAGS)

test-rds: ## Run RDS cascade Chainsaw suite (ctrl-rds-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/rds/ctrl-rds-01/ $(CHAINSAW_FLAGS)

test-secretsmanager: ## Run Secrets Manager cascade Chainsaw suite (ctrl-secretsmanager-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/secretsmanager/ctrl-secretsmanager-01/ $(CHAINSAW_FLAGS)

test-sns: ## Run SNS cascade Chainsaw suite (ctrl-sns-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/sns/ctrl-sns-01/ $(CHAINSAW_FLAGS)

test-sqs: ## Run SQS cascade Chainsaw suite (ctrl-sqs-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/sqs/ctrl-sqs-01/ $(CHAINSAW_FLAGS)

test-stepfunctions: ## Run Step Functions cascade Chainsaw suite (ctrl-sfn-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/stepfunctions/ctrl-sfn-01/ $(CHAINSAW_FLAGS)

test-ram: ## Run RAM cascade Chainsaw suite (ctrl-ram-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/ram/ctrl-ram-01/ $(CHAINSAW_FLAGS)

test-ses: ## Run SES cascade Chainsaw suite (ctrl-ses-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/ses/ctrl-ses-01/ $(CHAINSAW_FLAGS)

test-codeartifact: ## Run CodeArtifact cascade Chainsaw suite (ctrl-codeartifact-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/codeartifact/ctrl-codeartifact-01/ $(CHAINSAW_FLAGS)

test-appscaling: ## Run AppScaling cascade Chainsaw suite (ctrl-appscaling-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/appscaling/ctrl-appscaling-01/ $(CHAINSAW_FLAGS)

test-documentdb: ## Run DocumentDB cascade Chainsaw suite (ctrl-docdb-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/documentdb/ctrl-docdb-01/ $(CHAINSAW_FLAGS)

test-version: ## Run build-info and feature-enabled metrics Chainsaw suite (ctrl-version-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/version/ctrl-version-01/ $(CHAINSAW_FLAGS)

test-features: ## Run /features endpoint Chainsaw suite (ctrl-features-01).
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/features/ctrl-features-01/ $(CHAINSAW_FLAGS)

test-dyn-01: ## Run dynamic CRD detection suite 01 — operator starts with ELBConfig CRD absent.
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/ctrl-dyn-01/ $(CHAINSAW_FLAGS)

test-dyn-02: ## Run dynamic CRD detection suite 02 — install ELBConfig CRD, reconciler activates.
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/ctrl-dyn-02/ $(CHAINSAW_FLAGS)

test-dyn-03: ## Run dynamic CRD detection suite 03 — delete and reinstall ELBConfig CRD.
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/ctrl-dyn-03/ $(CHAINSAW_FLAGS)

test-dyn: ## Run dynamic CRD detection suites 01 → 02 → 03 in the required order.
	@mkdir -p $(REPORT_DIR)
	$(CHAINSAW) test tests/ctrl-dyn-01/ $(CHAINSAW_FLAGS)
	$(CHAINSAW) test tests/ctrl-dyn-02/ $(CHAINSAW_FLAGS)
	$(CHAINSAW) test tests/ctrl-dyn-03/ $(CHAINSAW_FLAGS)

test-chainsaw: chainsaw-stop chainsaw-start chainsaw-wait ## Stop any stale controller, start fresh, run ALL Chainsaw suites, then stop it.
	@mkdir -p $(REPORT_DIR)
	# ctrl-dyn suites have strict ordering: 01 → 02 → 03.
	# ctrl-dyn-01 requires ELBConfig CRD absent (pending state).
	# ctrl-dyn-02 installs the CRD; ctrl-dyn-03 requires it already present.
	# Run each as a separate chainsaw invocation to guarantee FIFO execution.
	$(CHAINSAW) test tests/ctrl-dyn-01/ $(CHAINSAW_FLAGS)
	$(CHAINSAW) test tests/ctrl-dyn-02/ $(CHAINSAW_FLAGS)
	$(CHAINSAW) test tests/ctrl-dyn-03/ $(CHAINSAW_FLAGS)
	# All remaining suites are order-independent.
	$(CHAINSAW) test \
		tests/acm/ tests/apigateway/ tests/apigatewayv2/ tests/athena/ tests/autoscaling/ \
		tests/cloudwatch/ tests/cloudwatchlogs/ tests/dynamodb/ tests/ec2/ \
		tests/ecr/ tests/ecs/ tests/efs/ tests/eks/ tests/elasticache/ \
		tests/emr/ tests/eventbridge/ tests/features/ tests/glue/ tests/iam/ tests/kms/ \
		tests/kropathconfig/ tests/label-operator/ tests/memorydb/ tests/msk/ tests/policy/ \
		tests/rds/ tests/s3/ tests/s3advanced/ tests/secretsmanager/ tests/sns/ tests/sqs/ \
		tests/stepfunctions/ tests/version/ tests/waf/ \
		tests/organizations/ \
		$(CHAINSAW_FLAGS)
	$(MAKE) chainsaw-stop

# ─── Tool installation ─────────────────────────────────────────────────────────

# Each tool gets its own target so that a CI job installs exactly what it needs
# at exactly the pinned version, without ci.yaml ever naming a version itself.

install-kind: ## Install the pinned kind binary.
	go install sigs.k8s.io/kind@$(KIND_VERSION)

install-chainsaw: ## Install the pinned chainsaw binary.
	go install github.com/kyverno/chainsaw@$(CHAINSAW_VERSION)

install-golangci-lint: ## Install the pinned golangci-lint binary.
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)

install-goimports: ## Install the pinned goimports binary.
	go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)

install-gosec: ## Install the pinned gosec binary.
	go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION)

install-govulncheck: ## Install the pinned govulncheck binary.
	go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)

install-tools: install-kind install-chainsaw install-golangci-lint install-goimports \
	install-gosec install-govulncheck ## Install every pinned tool locally.

# Emitted as KEY=value lines so a CI step can redirect it straight into
# $$GITHUB_OUTPUT. Used for tools installed from a prebuilt release rather than
# through `go install`, which is the only case ci.yaml cannot cover with an
# `install-*` target.
print-tool-versions: ## Print the pinned tool versions as KEY=value lines.
	@echo "KIND_VERSION=$(KIND_VERSION)"
	@echo "CHAINSAW_VERSION=$(CHAINSAW_VERSION)"
	@echo "GOLANGCI_VERSION=$(GOLANGCI_VERSION)"
	@echo "GOIMPORTS_VERSION=$(GOIMPORTS_VERSION)"
	@echo "GOSEC_VERSION=$(GOSEC_VERSION)"
	@echo "GOVULNCHECK_VERSION=$(GOVULNCHECK_VERSION)"

# ─── Security scans ────────────────────────────────────────────────────────────
# Run only when implementation is complete. Do not run during active development.

gosec: ## Run gosec SAST scanner.
	gosec ./...

vulncheck: ## Run govulncheck dependency CVE scanner.
	govulncheck ./...

security: gosec vulncheck ## Run all security scans (gosec + govulncheck).

# ─── Help ──────────────────────────────────────────────────────────────────────

help: ## Display this help message.
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-24s\033[0m %s\n", $$1, $$2}'
