SHELL := /bin/bash
MAKEFLAGS += --no-print-directory

.PHONY: version-sync
version-sync:
	node hack/version-sync.mjs

# checks hierarchy
# GROUP            TARGET              CACHE PATHS                    CHECKS
# extensions-check extensions-check    -                              configuration and installation
# check-php        check-extension     ZetaExtension                  lint
#                  check-skin          ZetaSkin                       lint
# check-svelte     check-main-svelte   main svelte                    deps, install, peers, lint, build
#                  check-skin-svelte   skin svelte, shared svelte     install, peers, lint, build
# check-goapp      check-goapp         goapp                          golangci-lint, test, build
# checks runs configuration validation plus PHP, Svelte, and Go checks with caching.
.PHONY: checks
checks:
	node hack/checks.mjs checks

.PHONY: checks-no-cache
checks-no-cache:
	node hack/checks.mjs checks --no-cache

.PHONY: clear
clear:
	node hack/checks.mjs clear

.PHONY: extensions-check
extensions-check:
	node hack/extensions-check.mjs

.PHONY: extensions-sync
extensions-sync:
	node hack/extensions-sync.mjs

.PHONY: extensions-recreate
extensions-recreate:
	node hack/extensions-recreate.mjs

.PHONY: check-php
check-php:
	node hack/checks.mjs check-php

.PHONY: check-extension
check-extension:
	node hack/checks.mjs check-extension

.PHONY: check-skin
check-skin:
	node hack/checks.mjs check-skin

.PHONY: check-svelte
check-svelte:
	node hack/checks.mjs check-svelte

.PHONY: svfmt
svfmt:
	pnpm -C svelte format:fix
	pnpm -C mwz/skins/ZetaSkin/svelte format:fix

.PHONY: svelte-common-deps
svelte-common-deps:
	node hack/svelte-common-deps.mjs

.PHONY: install-main-svelte
install-main-svelte:
	pnpm -C svelte install --frozen-lockfile

.PHONY: install-skin-svelte
install-skin-svelte:
	pnpm -C mwz/skins/ZetaSkin/svelte install --frozen-lockfile

.PHONY: check-main-svelte
check-main-svelte:
	node hack/checks.mjs check-main-svelte

.PHONY: check-skin-svelte
check-skin-svelte:
	node hack/checks.mjs check-skin-svelte

.PHONY: install-goapp-tools
install-goapp-tools:
	$(MAKE) -C goapp install-tools

.PHONY: check-goapp
check-goapp:
	node hack/checks.mjs check-goapp
