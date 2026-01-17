SHELL := /bin/bash

VUE_DIRS := mwz/skins/ZetaSkin/vue vue

define run_pnpm
	@set -e; \
	for d in $(VUE_DIRS); do \
		echo "➡️  $$d: pnpm -C $$d $(1)"; \
		pnpm -C $$d $(1); \
	done
endef

.PHONY: vue-overrides
vue-overrides:
	@echo "➡️  root: pnpm overrides"
	pnpm overrides

.PHONY: vue-overrides-fix
vue-overrides-fix:
	@echo "➡️  root: pnpm overrides:fix"
	pnpm overrides:fix

.PHONY: vue-install
vue-install:
	$(call run_pnpm,install --frozen-lockfile)

.PHONY: vue-lint
vue-lint:
	$(call run_pnpm,lint)

.PHONY: vue-build
vue-build:
	$(call run_pnpm,build)

.PHONY: vue-audit
vue-audit:
	$(call run_pnpm,audit)

.PHONY: checks
checks: vue-overrides vue-install vue-lint vue-build vue-audit
	@echo "✅  All checks passed"

.PHONY: vue-lint-fix
vue-lint-fix:
	$(call run_pnpm,lint:fix)

.PHONY: vue-audit-fix
vue-audit-fix:
	$(call run_pnpm,audit --fix)
	$(call run_pnpm,install --no-frozen-lockfile)

.PHONY: pnpm-update
pnpm-update:
	@echo "➡️  Updating pnpm to the latest version..."
	corepack prepare pnpm@latest --activate
	@echo "✅  pnpm updated to: $$(pnpm --version)"
