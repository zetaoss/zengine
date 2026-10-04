# syntax=docker/dockerfile:1

# Stages:
#   extensions  MediaWiki extensions from hack/extensions.yaml
#   base        MediaWiki runtime + PHP extensions + extensions (formerly the zbase image)
#   dev         base + development tools (formerly the zdev image); `docker build --target dev`
#   prod        application image (default target)
#
# prod still builds FROM the published zbase image. Switching it to `FROM base` changes the
# production runtime and is done as a separate, deliberate step.
ARG ZBASE_VERSION=0.2.2
ARG GO_VERSION=1.26

FROM node:24-trixie-slim AS extensions

RUN apt-get update \
    && apt-get install -y --no-install-recommends git ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy only what the installer reads, so application changes do not re-clone extensions.
# Add a COPY line here when a new override directory is added under mwz/extensions.
WORKDIR /src
COPY hack/extensions.yaml hack/extensions.mjs hack/
COPY mwz/extensions/MsUpload mwz/extensions/MsUpload
RUN EXTENSIONS_DIR=/extensions node hack/extensions.mjs

# https://hub.docker.com/_/mediawiki
FROM mediawiki:1.43.9-fpm AS base

RUN set -eux; \
    ## system packages
    apt-get update; \
    apt-get install -y --no-install-recommends nginx; \
    rm -rf /var/lib/apt/lists/*; \
    ## php extensions
    curl -sSLf -o /usr/local/bin/install-php-extensions \
        https://github.com/mlocati/docker-php-extension-installer/releases/latest/download/install-php-extensions; \
    chmod +x /usr/local/bin/install-php-extensions; \
    install-php-extensions \
        pcntl \
        pdo_mysql \
        redis \
        wikidiff2 \
        zip

COPY --from=extensions /extensions/ /var/www/html/extensions/

# Composer is mounted only for dependency installation and is not retained in the image.
# composer.local.json merges extensions/*/composer.json (e.g. AWS).
RUN --mount=type=bind,from=composer:2.10,source=/usr/bin/composer,target=/usr/local/bin/composer \
    set -eux; \
    cd /var/www/html; \
    cp composer.local.json-sample composer.local.json; \
    composer update --no-dev --no-scripts --optimize-autoloader

FROM golang:${GO_VERSION}-trixie AS go-devtools

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    set -eux; \
    GOBIN=/out/bin go install golang.org/x/tools/gopls@latest; \
    GOBIN=/out/bin go install github.com/air-verse/air@latest

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

ENV GOPATH=/go \
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

# https://github.com/zetaoss/zbase/pkgs/container/zbase
FROM ghcr.io/zetaoss/zbase:${ZBASE_VERSION} AS prod

ENV MW_INSTALL_PATH=/app/w

COPY --from=nodebuild /app      /app
COPY --from=gobuild   /out/bin/ /app/bin/

RUN set -eux \
    && mv /var/www/html                         /app/w \
    && ln -rs /app/mwz/extensions/ZetaExtension /app/w/extensions/ \
    && ln -rs /app/mwz/skins/ZetaSkin           /app/w/skins/ \
    && chown www-data:www-data -R /app/*

# Same overlay as before the stage split: zbase provides the other extensions, so only these are
# replaced. Remove this block (and use `FROM base AS prod`) when prod moves to the base stage.
RUN --mount=type=bind,from=extensions,source=/extensions,target=/tmp/extensions \
    set -eux; \
    for name in CrawlerProtection MailAPI MsUpload SimpleMaps SimpleMarkdown SimpleMathJax SimpleMermaid; do \
        cp -a "/tmp/extensions/$name" /app/w/extensions/; \
        chown -R www-data:www-data "/app/w/extensions/$name"; \
    done
