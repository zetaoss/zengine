# syntax=docker/dockerfile:1

# Stages:
#   extensions  MediaWiki extensions from mwz/extensions.yaml
#   base        MediaWiki runtime + PHP extensions + extensions (formerly the zbase image)
#   dev         base + development tools (formerly the zdev image); `docker build --target dev`
#   prod        base + application (default target)
#
# dev and prod share the base stage, so MediaWiki, PHP extensions and MediaWiki extensions are
# the same in both.
ARG GO_VERSION=1.26

FROM node:24-trixie-slim AS extensions

RUN apt-get update \
    && apt-get install -y --no-install-recommends git ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy only what the installer reads, so application changes do not re-clone extensions.
# Add a COPY line here when a new override directory is added under mwz/extensions.
WORKDIR /src
COPY hack/package.json hack/pnpm-lock.yaml hack/
RUN corepack enable \
    && corepack prepare pnpm@11 --activate \
    && pnpm -C hack install --frozen-lockfile
COPY hack/extensions.mjs hack/
COPY mwz/extensions.yaml mwz/extensions.lock mwz/
RUN EXTENSIONS_DIR=/extensions node hack/extensions.mjs install

# https://github.com/wikimedia/mediawiki-docker/blob/main/1.43/fpm/Dockerfile
FROM php:8.3.37-fpm AS base

# install-php-extensions comes from a pinned image and is mounted only for this step, so builds are
# reproducible and it does not stay in the image.
RUN --mount=type=bind,from=mlocati/php-extension-installer:2.12.1,source=/usr/bin/install-php-extensions,target=/usr/local/bin/install-php-extensions \
    set -eux; \
    ## system packages
    apt-get update; \
    apt-get install -y --no-install-recommends ca-certificates curl gnupg git librsvg2-bin imagemagick python3; \
    curl -fsSL https://nginx.org/keys/nginx_signing.key | gpg --dearmor -o /usr/share/keyrings/nginx-archive-keyring.gpg; \
    echo 'deb [signed-by=/usr/share/keyrings/nginx-archive-keyring.gpg] https://nginx.org/packages/mainline/debian trixie nginx' > /etc/apt/sources.list.d/nginx.list; \
    apt-get update; \
    apt-get install -y --no-install-recommends nginx=1.31.9-1~trixie; \
    rm -rf /var/lib/apt/lists/*; \
    ## php extensions
    ## redis is built from the GitHub tag instead of PECL (deprecated); the clone includes the
    ## bundled liblzf submodule, which GitHub source tarballs omit.
    git clone --quiet --depth 1 --branch 6.3.0 --recurse-submodules --shallow-submodules \
        https://github.com/phpredis/phpredis.git /tmp/phpredis; \
    install-php-extensions \
        apcu-5.1.28 \
        calendar \
        intl \
        luasandbox-4.1.2 \
        mbstring \
        mysqli \
        opcache \
        pcntl \
        pdo_mysql \
        /tmp/phpredis \
        wikidiff2 \
        zip; \
    rm -rf /tmp/phpredis

RUN { \
        echo 'opcache.memory_consumption=128'; \
        echo 'opcache.interned_strings_buffer=8'; \
        echo 'opcache.max_accelerated_files=4000'; \
        echo 'opcache.revalidate_freq=60'; \
    } > /usr/local/etc/php/conf.d/opcache-recommended.ini

ENV MEDIAWIKI_MAJOR_VERSION=1.43
ENV MEDIAWIKI_VERSION=1.43.12
RUN set -eux; \
    curl -fsSL "https://releases.wikimedia.org/mediawiki/${MEDIAWIKI_MAJOR_VERSION}/mediawiki-${MEDIAWIKI_VERSION}.tar.gz" -o /tmp/mediawiki.tar.gz; \
    echo 'ff2b4361425a0634e8be21c6fa6531b96bb96b7b4c291979908d98b1da59dd90  /tmp/mediawiki.tar.gz' | sha256sum -c -; \
    tar -xzf /tmp/mediawiki.tar.gz -C /var/www/html --strip-components=1; \
    rm /tmp/mediawiki.tar.gz; \
    chown -R www-data:www-data /var/www/html/extensions /var/www/html/skins /var/www/html/cache /var/www/html/images

COPY --from=extensions /extensions/ /var/www/html/extensions/

# PHP dependencies of MediaWiki and the extensions above (composer.local.json merges
# extensions/*/composer.json, e.g. AWS), installed from the committed lock so every build gets the
# same versions. Regenerate the lock with `make composer-lock` (COMPOSER_MODE=update).
# Composer is mounted only for the install and is not retained in the image.
ARG COMPOSER_MODE=install
COPY hack/mediawiki-composer.lock /var/www/html/composer.lock
RUN --mount=type=bind,from=composer:2.10,source=/usr/bin/composer,target=/usr/local/bin/composer \
    set -eux; \
    cd /var/www/html; \
    cp composer.local.json-sample composer.local.json; \
    if [ "$COMPOSER_MODE" = update ]; then rm -f composer.lock; fi; \
    composer "$COMPOSER_MODE" --no-dev --no-scripts --optimize-autoloader

# Patches to MediaWiki and its extensions (mwz/patches, docs/patches.md). A patch that no longer
# applies, e.g. after a MediaWiki or extension upgrade, fails the build.
RUN --mount=type=bind,source=mwz/patches,target=/tmp/patches \
    set -eux; \
    for p in /tmp/patches/*.patch; do \
        patch -d /var/www/html -p1 --forward --fuzz=0 --no-backup-if-mismatch < "$p"; \
    done

FROM golang:${GO_VERSION}-trixie AS go-devtools

# Pinned: @latest moves on and can require a newer Go than GO_VERSION (gopls v0.24.0 needs Go 1.27).

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    set -eux; \
    GOBIN=/out/bin go install golang.org/x/tools/gopls@v0.23.0; \
    GOBIN=/out/bin go install github.com/air-verse/air@v1.67.4

FROM node:24-bookworm-slim AS cli-tools

ENV PATH=/root/.local/bin:${PATH}

RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends ca-certificates curl; \
    rm -rf /var/lib/apt/lists/*

# pnpm major is pinned to match the prod build (corepack pnpm@11) and CI.
RUN set -eux; \
    npm install --global pnpm@11; \
    curl -fsSL https://antigravity.google/cli/install.sh | bash; \
    curl -fsSL https://chatgpt.com/codex/install.sh | sh; \
    curl -fsSL https://gh.io/copilot-install | bash

FROM base AS dev

ENV MW_INSTALL_PATH=/var/www/html \
    GOPATH=/go \
    PATH=/usr/local/go/bin:/go/bin:/root/.local/bin:/root/.local/share/pnpm/bin:${PATH}

RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
        bash \
        coreutils \
        gcc \
        gh \
        inotify-tools \
        jq \
        mariadb-client \
        libc6-dev \
        openssh-server \
        procps \
        psmisc \
        ripgrep \
        supervisor; \
    rm -rf /var/lib/apt/lists/*; \
    printf '%s\n' \
        'export GOPATH=/go' \
        'export PATH="/usr/local/go/bin:/go/bin:/root/.local/bin:/root/.local/share/pnpm/bin:$PATH"' \
        > /etc/profile.d/zdev.sh

COPY --link --from=go-devtools /usr/local/go/ /usr/local/go/
COPY --link --from=go-devtools /out/bin/ /go/bin/
COPY --link --from=cli-tools /root/.codex/packages/ /root/.codex/packages/
COPY --link --from=cli-tools /root/.local/bin/ /root/.local/bin/
COPY --link --from=cli-tools /usr/local/bin/ /usr/local/bin/
COPY --link --from=cli-tools /usr/local/lib/node_modules/ /usr/local/lib/node_modules/
COPY --from=composer:2.10 /usr/bin/composer /usr/bin/composer

FROM node:24-trixie-slim AS nodebuild

RUN apt-get update \
    && apt-get install -y --no-install-recommends git ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && corepack enable \
    && corepack prepare pnpm@11 --activate

WORKDIR /app
COPY . .
RUN pnpm -C svelte                    install --frozen-lockfile
RUN pnpm -C mwz/skins/ZetaSkin/svelte install --frozen-lockfile
RUN node hack/version-sync.mjs
RUN pnpm -C svelte                    run build
RUN pnpm -C mwz/skins/ZetaSkin/svelte run build
# The runtime needs only the build output, not the build dependencies.
RUN rm -rf svelte/node_modules mwz/skins/ZetaSkin/svelte/node_modules

FROM --platform=$BUILDPLATFORM golang:1.25-trixie AS gobuild

WORKDIR /src/goapp
COPY goapp/go.* ./
RUN go mod download
COPY goapp/ ./
RUN mkdir -p /out/bin \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
        go build -trimpath -ldflags="-s -w" -o /out/bin/ \
        ./cmd/server \
        ./cmd/worker \
        ./cmd/scheduler \
        ./cmd/tool

FROM base AS prod

ENV MW_INSTALL_PATH=/app/w

# Move and chown MediaWiki before copying the application so this layer stays cached until base
# changes. Done after the application COPY, it rewrote MediaWiki and the extensions into a new
# layer on every release.
RUN set -eux \
    && mkdir /app \
    && mv /var/www/html /app/w \
    && chown www-data:www-data -R /app/w

# MediaWiki settings (docs/config.md). The deployment provides SiteSettings.php and SiteExtensionSettings.php.
# Copied after the move, so a settings change does not rebuild the MediaWiki layer.
COPY --chown=www-data:www-data mwz/settings/LocalSettings.php mwz/settings/BaseSettings.php mwz/settings/BaseExtensionSettings.php /app/w/

# Go binaries before the application: goapp changes less often, so this layer is often reused too.
COPY --from=gobuild   --chown=www-data:www-data /out/bin/ /app/bin/
COPY --from=nodebuild --chown=www-data:www-data /app      /app

RUN set -eux \
    && ln -rs /app/mwz/extensions/ZetaExtension /app/w/extensions/ \
    && ln -rs /app/mwz/skins/ZetaSkin           /app/w/skins/
