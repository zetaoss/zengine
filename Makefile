SHELL := /bin/bash

SVELTE_DIRS := svelte mwz/skins/ZetaSkin/svelte

define run_pnpm
	@set -e; \
	for d in $(SVELTE_DIRS); do \
		echo "➡️  $$d: pnpm -C $$d $(1)"; \
		pnpm -C $$d $(1); \
	done
endef

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

.PHONY: svelte-format
svelte-format:
	$(call run_pnpm,format)

.PHONY: svelte-build
svelte-build:
	$(call run_pnpm,build)

.PHONY: svelte-audit
svelte-audit:
	$(call run_pnpm,audit)

.PHONY: checks
checks: svelte-overrides svelte-install svelte-lint svelte-format svelte-build svelte-audit
	@echo "✅  All checks passed"

.PHONY: svelte-lint-fix
svelte-lint-fix:
	$(call run_pnpm,lint:fix)

.PHONY: svelte-audit-fix
svelte-audit-fix:
	$(call run_pnpm,audit --fix)
	$(call run_pnpm,install --no-frozen-lockfile)

.PHONY: pnpm-update
pnpm-update:
	@echo "➡️  Updating pnpm to the latest version..."
	corepack prepare pnpm@latest --activate
	@echo "✅  pnpm updated to: $$(pnpm --version)"
