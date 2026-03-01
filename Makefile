SHELL := /bin/bash
MAKEFLAGS += --no-print-directory

CACHE_DIR := /tmp/make-checks
MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
USE_CACHE ?= 0

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

define run_pint
	@echo "➡️  laravel/vendor/bin/pint --test $(1)"
	@laravel/vendor/bin/pint --test $(1) || { \
		$(call print_fix,$(2)) \
		exit 1; \
	}
endef

define run_pint_fix
	@echo "➡️  laravel/vendor/bin/pint $(1)"
	@laravel/vendor/bin/pint $(1)
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
# check-php      check-laravel       laravel         format, test
#                check-extension     ZetaExtension   format
#                check-skin          ZetaSkin        format
# check-svelte   check-overrides     -               overrides
#                check-main-svelte   main svelte     install, lint, format, audit, build
#                check-skin-svelte   skin svelte     install, lint, format, audit, build
.PHONY: checks
checks:
	@$(MAKE) USE_CACHE=1 check-php check-svelte
	@echo "✅  All checks passed"

# checks-no-cache runs the same tree with USE_CACHE=0.
.PHONY: checks-no-cache
checks-no-cache:
	@$(MAKE) USE_CACHE=0 check-php check-svelte
	@echo "✅  All checks passed (no cache)"

# quickcheck runs non-mutating checks that correspond to `fix`.
.PHONY: quickcheck
quickcheck:
	@$(MAKE) USE_CACHE=1 check-laravel-format check-extension check-skin
	@$(MAKE) USE_CACHE=1 quickcheck-main-svelte quickcheck-skin-svelte
	@echo "✅  Quickcheck passed"

.PHONY: clear
clear:
	@echo "🧹 clear cache: $(CACHE_DIR)"
	rm -rf $(CACHE_DIR)

.PHONY: check-php
check-php:
	@$(MAKE) USE_CACHE=$(USE_CACHE) check-laravel check-extension check-skin

.PHONY: check-laravel
check-laravel:
ifeq ($(USE_CACHE),1)
	$(call run_cached,check-laravel,$(MAKE) USE_CACHE=0 check-laravel,laravel pint.json)
else
	$(call run_pint,laravel,laravel/vendor/bin/pint laravel)
	@echo "➡️  laravel: php artisan test"
	cd laravel && php artisan test
endif

.PHONY: check-laravel-format
check-laravel-format:
ifeq ($(USE_CACHE),1)
	$(call run_cached,check-laravel-format,$(MAKE) USE_CACHE=0 check-laravel-format,laravel pint.json)
else
	$(call run_pint,laravel,laravel/vendor/bin/pint laravel)
endif

.PHONY: check-extension
check-extension:
ifeq ($(USE_CACHE),1)
	$(call run_cached,check-extension,$(MAKE) USE_CACHE=0 check-extension,mwz/extensions/ZetaExtension pint.json)
else
	$(call run_pint,mwz/extensions/ZetaExtension,laravel/vendor/bin/pint mwz/extensions/ZetaExtension)
endif

.PHONY: check-skin
check-skin:
ifeq ($(USE_CACHE),1)
	$(call run_cached,check-skin,$(MAKE) USE_CACHE=0 check-skin,mwz/skins/ZetaSkin pint.json)
else
	$(call run_pint,mwz/skins/ZetaSkin,laravel/vendor/bin/pint mwz/skins/ZetaSkin)
endif

.PHONY: check-svelte
check-svelte:
	@$(MAKE) svelte-overrides
	@$(MAKE) USE_CACHE=$(USE_CACHE) check-main-svelte check-skin-svelte

.PHONY: svelte-overrides
svelte-overrides:
	@echo "➡️  root: pnpm overrides"
	pnpm overrides

.PHONY: check-main-svelte
check-main-svelte:
ifeq ($(USE_CACHE),1)
	$(call run_cached,check-main-svelte,$(MAKE) USE_CACHE=0 check-main-svelte,svelte)
else
	$(call run_pnpm,svelte,install --frozen-lockfile)
	$(call run_pnpm,svelte,lint,pnpm -C svelte lint)
	$(call run_pnpm,svelte,format,pnpm -C svelte format)
	$(call run_pnpm,svelte,audit --ignore-unfixable --ignore-registry-errors,pnpm -C svelte audit --fix --ignore-unfixable && pnpm -C svelte install --no-frozen-lockfile)
	$(call run_pnpm,svelte,build)
endif

.PHONY: quickcheck-main-svelte
quickcheck-main-svelte:
ifeq ($(USE_CACHE),1)
	$(call run_cached,quickcheck-main-svelte,$(MAKE) USE_CACHE=0 quickcheck-main-svelte,svelte)
else
	$(call run_pnpm,svelte,install --frozen-lockfile)
	$(call run_pnpm,svelte,lint,pnpm -C svelte lint)
	$(call run_pnpm,svelte,format,pnpm -C svelte format)
endif

.PHONY: check-skin-svelte
check-skin-svelte:
ifeq ($(USE_CACHE),1)
	$(call run_cached,check-skin-svelte,$(MAKE) USE_CACHE=0 check-skin-svelte,mwz/skins/ZetaSkin/svelte svelte/src/shared)
else
	$(call run_pnpm,mwz/skins/ZetaSkin/svelte,install --frozen-lockfile)
	$(call run_pnpm,mwz/skins/ZetaSkin/svelte,lint,pnpm -C mwz/skins/ZetaSkin/svelte lint)
	$(call run_pnpm,mwz/skins/ZetaSkin/svelte,format,pnpm -C mwz/skins/ZetaSkin/svelte format)
	$(call run_pnpm,mwz/skins/ZetaSkin/svelte,audit --ignore-unfixable --ignore-registry-errors,pnpm -C mwz/skins/ZetaSkin/svelte audit --fix --ignore-unfixable && pnpm -C mwz/skins/ZetaSkin/svelte install --no-frozen-lockfile)
	$(call run_pnpm,mwz/skins/ZetaSkin/svelte,build)
endif

.PHONY: quickcheck-skin-svelte
quickcheck-skin-svelte:
ifeq ($(USE_CACHE),1)
	$(call run_cached,quickcheck-skin-svelte,$(MAKE) USE_CACHE=0 quickcheck-skin-svelte,mwz/skins/ZetaSkin/svelte svelte/src/shared)
else
	$(call run_pnpm,mwz/skins/ZetaSkin/svelte,install --frozen-lockfile)
	$(call run_pnpm,mwz/skins/ZetaSkin/svelte,lint,pnpm -C mwz/skins/ZetaSkin/svelte lint)
	$(call run_pnpm,mwz/skins/ZetaSkin/svelte,format,pnpm -C mwz/skins/ZetaSkin/svelte format)
endif
