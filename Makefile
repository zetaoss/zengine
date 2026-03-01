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
# GROUP           TARGET               CACHEDIR        CHECKS
# checks-php      checks-laravel       laravel         format, test
#                 checks-extension     ZetaExtension   format
#                 checks-skin          ZetaSkin        format
# checks-svelte   checks-overrides     -               overrides
#                 checks-main-svelte   main svelte     install, lint, format, audit, build
#                 checks-skin-svelte   skin svelte     install, lint, format, audit, build
.PHONY: checks
checks:
	@$(MAKE) USE_CACHE=1 checks-php checks-svelte
	@echo "✅  All checks passed"

# fix runs auto-fixable checks with cache.
.PHONY: fix
fix:
	@$(MAKE) USE_CACHE=1 fix-laravel-format fix-extension fix-skin
	@$(MAKE) svelte-overrides
	@$(MAKE) USE_CACHE=1 fix-main-svelte fix-skin-svelte
	@echo "✅  Fix completed"

# checks-no-cache runs the same tree with USE_CACHE=0.
.PHONY: checks-no-cache
checks-no-cache:
	@$(MAKE) USE_CACHE=0 checks-php checks-svelte
	@echo "✅  All checks passed (no cache)"

.PHONY: clear
clear:
	@echo "🧹 clear cache: $(CACHE_DIR)"
	rm -rf $(CACHE_DIR)

.PHONY: checks-php
checks-php:
	@$(MAKE) USE_CACHE=$(USE_CACHE) checks-laravel checks-extension checks-skin

.PHONY: checks-laravel
checks-laravel:
ifeq ($(USE_CACHE),1)
	$(call run_cached,checks-laravel,$(MAKE) USE_CACHE=0 checks-laravel,laravel pint.json)
else
	$(call run_pint,laravel,laravel/vendor/bin/pint laravel)
	@echo "➡️  laravel: php artisan test"
	cd laravel && php artisan test
endif

.PHONY: checks-laravel-format
checks-laravel-format:
ifeq ($(USE_CACHE),1)
	$(call run_cached,checks-laravel-format,$(MAKE) USE_CACHE=0 checks-laravel-format,laravel pint.json)
else
	$(call run_pint,laravel,laravel/vendor/bin/pint laravel)
endif

.PHONY: fix-laravel-format
fix-laravel-format:
ifeq ($(USE_CACHE),1)
	$(call run_cached,fix-laravel-format,$(MAKE) USE_CACHE=0 fix-laravel-format,laravel pint.json)
else
	$(call run_pint_fix,laravel)
endif

.PHONY: checks-extension
checks-extension:
ifeq ($(USE_CACHE),1)
	$(call run_cached,checks-extension,$(MAKE) USE_CACHE=0 checks-extension,mwz/extensions/ZetaExtension pint.json)
else
	$(call run_pint,mwz/extensions/ZetaExtension,laravel/vendor/bin/pint mwz/extensions/ZetaExtension)
endif

.PHONY: fix-extension
fix-extension:
ifeq ($(USE_CACHE),1)
	$(call run_cached,fix-extension,$(MAKE) USE_CACHE=0 fix-extension,mwz/extensions/ZetaExtension pint.json)
else
	$(call run_pint_fix,mwz/extensions/ZetaExtension)
endif

.PHONY: checks-skin
checks-skin:
ifeq ($(USE_CACHE),1)
	$(call run_cached,checks-skin,$(MAKE) USE_CACHE=0 checks-skin,mwz/skins/ZetaSkin pint.json)
else
	$(call run_pint,mwz/skins/ZetaSkin,laravel/vendor/bin/pint mwz/skins/ZetaSkin)
endif

.PHONY: fix-skin
fix-skin:
ifeq ($(USE_CACHE),1)
	$(call run_cached,fix-skin,$(MAKE) USE_CACHE=0 fix-skin,mwz/skins/ZetaSkin pint.json)
else
	$(call run_pint_fix,mwz/skins/ZetaSkin)
endif

.PHONY: checks-svelte
checks-svelte:
	@$(MAKE) svelte-overrides
	@$(MAKE) USE_CACHE=$(USE_CACHE) checks-main-svelte checks-skin-svelte

.PHONY: svelte-overrides
svelte-overrides:
	@echo "➡️  root: pnpm overrides"
	pnpm overrides

.PHONY: checks-main-svelte
checks-main-svelte:
ifeq ($(USE_CACHE),1)
	$(call run_cached,checks-main-svelte,$(MAKE) USE_CACHE=0 checks-main-svelte,svelte)
else
	$(call run_pnpm,svelte,install --frozen-lockfile)
	$(call run_pnpm,svelte,lint,pnpm -C svelte lint:fix)
	$(call run_pnpm,svelte,format,pnpm -C svelte format:fix)
	$(call run_pnpm,svelte,audit --ignore-unfixable --ignore-registry-errors,pnpm -C svelte audit --fix --ignore-unfixable && pnpm -C svelte install --no-frozen-lockfile)
	$(call run_pnpm,svelte,build)
endif

.PHONY: fix-main-svelte
fix-main-svelte:
ifeq ($(USE_CACHE),1)
	$(call run_cached,fix-main-svelte,$(MAKE) USE_CACHE=0 fix-main-svelte,svelte)
else
	$(call run_pnpm,svelte,install --frozen-lockfile)
	$(call run_pnpm,svelte,lint:fix)
	$(call run_pnpm,svelte,format:fix)
endif

.PHONY: checks-skin-svelte
checks-skin-svelte:
ifeq ($(USE_CACHE),1)
	$(call run_cached,checks-skin-svelte,$(MAKE) USE_CACHE=0 checks-skin-svelte,mwz/skins/ZetaSkin/svelte svelte/src/shared)
else
	$(call run_pnpm,mwz/skins/ZetaSkin/svelte,install --frozen-lockfile)
	$(call run_pnpm,mwz/skins/ZetaSkin/svelte,lint,pnpm -C mwz/skins/ZetaSkin/svelte lint:fix)
	$(call run_pnpm,mwz/skins/ZetaSkin/svelte,format,pnpm -C mwz/skins/ZetaSkin/svelte format:fix)
	$(call run_pnpm,mwz/skins/ZetaSkin/svelte,audit --ignore-unfixable --ignore-registry-errors,pnpm -C mwz/skins/ZetaSkin/svelte audit --fix --ignore-unfixable && pnpm -C mwz/skins/ZetaSkin/svelte install --no-frozen-lockfile)
	$(call run_pnpm,mwz/skins/ZetaSkin/svelte,build)
endif

.PHONY: fix-skin-svelte
fix-skin-svelte:
ifeq ($(USE_CACHE),1)
	$(call run_cached,fix-skin-svelte,$(MAKE) USE_CACHE=0 fix-skin-svelte,mwz/skins/ZetaSkin/svelte svelte/src/shared)
else
	$(call run_pnpm,mwz/skins/ZetaSkin/svelte,install --frozen-lockfile)
	$(call run_pnpm,mwz/skins/ZetaSkin/svelte,lint:fix)
	$(call run_pnpm,mwz/skins/ZetaSkin/svelte,format:fix)
endif
