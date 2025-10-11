SHELL=/bin/bash

vue-install:
	@echo "Running vue-install"
	cd mwz/skins/ZetaSkin/vue && pnpm install
	cd vue && pnpm install

vue-build:
	@echo "Running vue-build"
	cd mwz/skins/ZetaSkin/vue && pnpm build
	cd vue && pnpm build

vue-diff:
	@echo "Running vue-diff"
	node hack/vue-diff.js

vue-audit:
	@echo "Running vue-audit"
	cd mwz/skins/ZetaSkin/vue && pnpm audit
	cd vue && pnpm audit

checks: vue-install vue-build vue-diff vue-audit
	@echo "✅ All checks passed"
