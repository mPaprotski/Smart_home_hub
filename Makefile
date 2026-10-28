SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c
GO ?= go
GOFMT ?= $(shell $(GO) env GOROOT)/bin/gofmt
PYTHON ?= .venv/bin/python
RUFF ?= .venv/bin/ruff
MYPY ?= .venv/bin/mypy
export GOCACHE := $(CURDIR)/.cache/go-build
export GOMODCACHE := $(CURDIR)/.cache/go-mod
GO_PKGS := ./cmd/... ./internal/... ./telegram/...
.PHONY: help fmt fmt-check lint typecheck test test-go test-python build build-go check run-sim smoke openapi-check bench compose-check
help:
	@printf '%s\n' 'make check: format/lint/types/tests/OpenAPI/build (Core, frontend, Python)' 'make run-sim | smoke | bench | compose-check' 'Bootstrap and pinned tools: docs/development/README.md'
fmt:
	$(GO) fmt $(GO_PKGS)
	$(RUFF) format ml
	scripts/npm.sh run format
fmt-check:
	@files="$$($(GOFMT) -l cmd internal telegram)"; test -z "$$files" || { echo 'Run make fmt'; exit 1; }
	$(RUFF) format --check ml
	scripts/npm.sh run format:check
lint:
	$(GO) vet $(GO_PKGS)
	$(RUFF) check ml
	scripts/npm.sh run lint
typecheck:
	$(MYPY) --config-file ml/pyproject.toml ml/smarthome_ml
	scripts/npm.sh run typecheck
test: test-go test-python
test-go:
	$(GO) test -race $(GO_PKGS)
test-python:
	$(PYTHON) -m pytest ml/tests -q
build-go:
	mkdir -p bin
	$(GO) build -trimpath -o bin/hub ./cmd/hub
	$(GO) build -trimpath -o bin/screen-client ./cmd/screen-client
	$(GO) build -trimpath -o bin/bot ./telegram/cmd/bot
build: build-go
	scripts/npm.sh run build
	$(PYTHON) -m build --no-isolation --outdir bin/python ml
openapi-check:
	$(PYTHON) -m openapi_spec_validator docs/api/openapi.yaml
check: fmt-check lint typecheck test openapi-check build
run-sim: build-go
	./bin/hub --config config.example.yaml
smoke: build-go
	$(PYTHON) scripts/smoke.py
	$(PYTHON) scripts/config_smoke.py
bench:
	$(GO) test ./internal/events -run '^$$' -bench . -benchmem
compose-check:
	docker compose -f deploy/docker/compose.dev.yaml config --quiet

.PHONY: smoke-persistence compose-persistence-check
smoke-persistence: build-go
	$(PYTHON) scripts/persistence_smoke.py
compose-persistence-check:
	docker compose -f deploy/docker/compose.dev.yaml -f deploy/docker/compose.sqlite.yaml config --quiet

.PHONY: smoke-auth compose-auth-check
smoke-auth: build-go
	$(PYTHON) scripts/auth_smoke.py
compose-auth-check:
	docker compose -f deploy/docker/compose.dev.yaml -f deploy/docker/compose.sqlite.yaml -f deploy/docker/compose.auth.yaml config --quiet
