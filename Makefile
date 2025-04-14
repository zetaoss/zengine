SHELL=/bin/bash

vue-audit:
	@echo "🔍 Running vue-audit"
	cd mwz/skins/ZetaSkin/vue && npm audit
	cd vue && npm audit

vue-diff:
	@echo "🧮 Running vue-diff"
	node hack/vue-diff.js

checks: vue-audit vue-diff