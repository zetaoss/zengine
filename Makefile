SHELL := /bin/bash
MAKEFLAGS += --no-print-directory

CACHE_DIR := /tmp/make-checks
MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
USE_CACHE ?= 0
GOLANGCI_LINT_VERSION ?= v2.12.2

define print_fix
echo; \
echo "💡 This might be fixed with:"; \
echo; \
echo "   $(1)"; \
echo;
endef

define run_pnpm
	@echo "➡️  $(1): pnpm -C $(1) $(2)"
	@pnpm -C $(1) $(2) || { \
		if [ -n "$(strip $(3))" ]; then \
			$(call print_fix,$(3)) \
		fi; \
		exit 1; \
	}
endef

define strip_ignore_cves_null
	@sed -i 's| ignoreCves: null ||' $(1)
endef

define run_mwz_test
	@echo "➡️  $(1): composer test"
	@cd $(1) && composer test || { \
		$(call print_fix,cd $(1) && composer fix) \
		exit 1; \
	}
endef

define run_cached
	@bash -lc 'set -euo pipefail; \
	key="$(1)"; \
	cmd="$(2)"; \
	paths="$(3)"; \
	cache_dir="$(CACHE_DIR)"; \
	mkdir -p "$$cache_dir"; \
	hash_cmd() { \
		if command -v sha256sum >/dev/null 2>&1; then sha256sum "$$@"; else shasum -a 256 "$$@"; fi; \
	}; \
	tmp_hash="$$(mktemp)"; \
	{ \
		echo "cmd=$$cmd"; \
		echo "node=$$(node -v 2>/dev/null || true)"; \
		echo "pnpm=$$(pnpm -v 2>/dev/null || true)"; \
		echo "php=$$(php -v 2>/dev/null | head -n 1 || true)"; \
		if [ -f "$(MAKEFILE_PATH)" ]; then hash_cmd "$(MAKEFILE_PATH)"; fi; \
		for p in $$paths; do \
			if [ -d "$$p" ]; then \
				find "$$p" -type f \
					-not -path "*/node_modules/*" \
					-not -path "*/dist/*" \
					-not -path "*/.svelte-kit/*" \
					-not -path "*/vendor/*" \
					-print | sort | while IFS= read -r f; do hash_cmd "$$f"; done; \
			elif [ -f "$$p" ]; then \
				hash_cmd "$$p"; \
			fi; \
		done; \
	} | hash_cmd | awk "{print \$$1}" > "$$tmp_hash"; \
	hash_file="$$cache_dir/$$key.hash"; \
	mkdir -p "$$cache_dir"; \
	if [ -f "$$hash_file" ] && cmp -s "$$tmp_hash" "$$hash_file"; then \
		echo "⏭️  $$key: no changes, skip"; \
		rm -f "$$tmp_hash"; \
	else \
		echo "➡️  $$key: $$cmd"; \
		eval "$$cmd"; \
		mv "$$tmp_hash" "$$hash_file"; \
	fi'
endef

# checks hierarchy (USE_CACHE=1)
# GROUP          TARGET              CACHEDIR        CHECKS
# check-php      check-extension     ZetaExtension   lint
#                check-skin          ZetaSkin        lint
# check-svelte   check-overrides     -               overrides
#                check-main-svelte   main svelte     install, lint, format, audit, build
#                check-skin-svelte   skin svelte     install, lint, format, audit, build
# check-goapp    check-goapp         goapp           golangci-lint, test, build
.PHONY: checks
checks:
	@$(MAKE) USE_CACHE=1 check-php check-svelte check-goapp
	@echo "✅  All checks passed"

# checks-no-cache runs the same tree with USE_CACHE=0.
.PHONY: checks-no-cache
checks-no-cache:
	@$(MAKE) USE_CACHE=0 check-php check-svelte check-goapp
	@echo "✅  All checks passed (no cache)"

.PHONY: clear
clear:
	@echo "🧹 clear cache: $(CACHE_DIR)"
	rm -rf $(CACHE_DIR)

.PHONY: check-php
check-php:
	@$(MAKE) USE_CACHE=$(USE_CACHE) check-extension check-skin

.PHONY: check-extension
check-extension:
ifeq ($(USE_CACHE),1)
	$(call run_cached,check-extension,$(MAKE) USE_CACHE=0 check-extension,mwz/extensions/ZetaExtension)
else
	$(call run_mwz_test,mwz/extensions/ZetaExtension)
endif

.PHONY: check-skin
check-skin:
ifeq ($(USE_CACHE),1)
	$(call run_cached,check-skin,$(MAKE) USE_CACHE=0 check-skin,mwz/skins/ZetaSkin)
else
	$(call run_mwz_test,mwz/skins/ZetaSkin)
endif

.PHONY: check-svelte
check-svelte:
	@$(MAKE) USE_CACHE=$(USE_CACHE) check-main-svelte check-skin-svelte

.PHONY: svelte-overrides
svelte-overrides:
	@echo "➡️  root: pnpm overrides"
	pnpm overrides

.PHONY: install-main-svelte
install-main-svelte:
	$(call run_pnpm,svelte,install --frozen-lockfile)

.PHONY: install-skin-svelte
install-skin-svelte:
	$(call run_pnpm,mwz/skins/ZetaSkin/svelte,install --frozen-lockfile)

.PHONY: check-main-svelte
check-main-svelte:
ifeq ($(USE_CACHE),1)
	$(call run_cached,check-main-svelte,$(MAKE) USE_CACHE=0 check-main-svelte,svelte pnpm-lock.yaml package.json)
else
		@$(MAKE) svelte-overrides
	$(call run_pnpm,svelte,install --frozen-lockfile)
	$(call run_pnpm,svelte,peers check)
	$(call run_pnpm,svelte,lint)
	$(call run_pnpm,svelte,build)
endif

.PHONY: check-skin-svelte
check-skin-svelte:
ifeq ($(USE_CACHE),1)
	$(call run_cached,check-skin-svelte,$(MAKE) USE_CACHE=0 check-skin-svelte,mwz/skins/ZetaSkin/svelte svelte/src/shared pnpm-lock.yaml package.json)
else
	$(call run_pnpm,mwz/skins/ZetaSkin/svelte,install --frozen-lockfile)
	$(call run_pnpm,mwz/skins/ZetaSkin/svelte,peers check)
	$(call run_pnpm,mwz/skins/ZetaSkin/svelte,lint)
	$(call run_pnpm,mwz/skins/ZetaSkin/svelte,build)
endif

.PHONY: check-goapp
check-goapp:
ifeq ($(USE_CACHE),1)
	$(call run_cached,check-goapp,$(MAKE) USE_CACHE=0 check-goapp,goapp)
else
	@echo "➡️  goapp: golangci-lint run"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		cd goapp && golangci-lint run; \
	else \
		echo "ℹ️  golangci-lint not found, using go run fallback"; \
		cd goapp && go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run; \
	fi
	@echo "➡️  goapp: go test ./..."
	cd goapp && go test ./...
	@echo "➡️  goapp: go build ./..."
	cd goapp && go build ./...
endif
