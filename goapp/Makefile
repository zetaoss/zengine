SHELL := /bin/bash
MAKEFLAGS += --no-print-directory

GOLANGCI_LINT_VERSION ?= v2.12.2
BIN_DIR := $(CURDIR)/bin
GOLANGCI_LINT := $(BIN_DIR)/golangci-lint

.PHONY: dev
dev:
	@fuser -k 8080/tcp || true && go run ./cmd/server --dev=true

.PHONY: run
run:
	@fuser -k 8080/tcp || true && go run ./cmd/server

.PHONY: lint
lint:
	@if [ -x "$(GOLANGCI_LINT)" ]; then \
		"$(GOLANGCI_LINT)" run; \
	elif command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "ℹ️  golangci-lint not found, installing to $(BIN_DIR)"; \
		$(MAKE) install-tools; \
		"$(GOLANGCI_LINT)" run; \
	fi

.PHONY: install-tools
install-tools:
	@echo "➡️  goapp: install golangci-lint $(GOLANGCI_LINT_VERSION)"
	@mkdir -p $(BIN_DIR)
	@GOBIN=$(BIN_DIR) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
