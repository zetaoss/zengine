SHELL=/bin/bash

vue-audit:
	@echo "Running vue-audit"
	cd mwz/skins/ZetaSkin/vue && pnpm audit
	cd vue && pnpm audit

vue-build:
	@echo "Running vue-build"
	cd mwz/skins/ZetaSkin/vue && pnpm build
	cd vue && pnpm build

vue-diff:
	@echo "Running vue-diff"
	node hack/vue-diff.js

checks: vue-audit vue-build vue-diff
