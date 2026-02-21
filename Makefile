SHELL := /bin/bash

SVELTE_DIRS := svelte mwz/skins/ZetaSkin/svelte
LARAVEL_DIR := laravel

define run_pnpm
	@set -e; \
	for d in $(SVELTE_DIRS); do \
		echo "➡️  $$d: pnpm -C $$d $(1)"; \
		pnpm -C $$d $(1); \
	done
endef

.PHONY: laravel-format
laravel-format:
	@echo "➡️  $(LARAVEL_DIR): vendor/bin/pint --test"
	cd $(LARAVEL_DIR) && vendor/bin/pint --test

.PHONY: laravel-format-fix
laravel-format-fix:
	@echo "➡️  $(LARAVEL_DIR): vendor/bin/pint"
	cd $(LARAVEL_DIR) && vendor/bin/pint

.PHONY: laravel-test
laravel-test:
	@echo "➡️  $(LARAVEL_DIR): php artisan test"
	cd $(LARAVEL_DIR) && php artisan test

.PHONY: svelte-overrides
svelte-overrides:
	@echo "➡️  root: pnpm overrides"
	pnpm overrides

.PHONY: svelte-overrides-fix
svelte-overrides-fix:
	@echo "➡️  root: pnpm overrides:fix"
	pnpm overrides:fix

.PHONY: svelte-install
svelte-install:
	$(call run_pnpm,install --frozen-lockfile)

.PHONY: svelte-lint
svelte-lint:
	$(call run_pnpm,lint)

.PHONY: svelte-lint-fix
svelte-lint-fix:
	$(call run_pnpm,lint:fix)

.PHONY: svelte-format
svelte-format:
	$(call run_pnpm,format)

.PHONY: svelte-format-fix
svelte-format-fix:
	$(call run_pnpm,format:fix)

.PHONY: svelte-build
svelte-build:
	$(call run_pnpm,build)

.PHONY: svelte-audit
svelte-audit:
	$(call run_pnpm,audit --ignore-unfixable --ignore-registry-errors)

.PHONY: svelte-audit-fix
svelte-audit-fix:
	$(call run_pnpm,audit --fix --ignore-unfixable)
	$(call run_pnpm,install --no-frozen-lockfile)

.PHONY: checks
checks: laravel-format laravel-test svelte-overrides svelte-install svelte-lint svelte-format svelte-build svelte-audit
	@echo "✅  All checks passed"

.PHONY: pnpm-update
pnpm-update:
	@echo "➡️  Updating pnpm to the latest version..."
	corepack prepare pnpm@latest --activate
	@echo "✅  pnpm updated to: $$(pnpm --version)"
