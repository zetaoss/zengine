# syntax=docker/dockerfile:1

ARG ZBASE_VERSION=0.2.2

FROM node:24-trixie-slim AS nodebuild

RUN corepack enable && corepack prepare pnpm@11 --activate

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
FROM ghcr.io/zetaoss/zbase:${ZBASE_VERSION}

ENV MW_INSTALL_PATH=/app/w

COPY --from=nodebuild /app      /app
COPY --from=gobuild   /out/bin/ /app/bin/

RUN set -eux \
    && mv /var/www/html                         /app/w \
    && ln -rs /app/mwz/extensions/ZetaExtension /app/w/extensions/ \
    && ln -rs /app/mwz/skins/ZetaSkin           /app/w/skins/ \
    && chown www-data:www-data -R /app/*

# extensions-sync: generated; do not edit
RUN --mount=from=composer:2.10,source=/usr/bin/composer,target=/usr/bin/composer \
    set -eux \
    && cd /app/w/extensions \
    && rm -rf 'MailAPI' 'SimpleMermaid' \
    && git clone --depth=1 --branch 'v0.1.1' 'https://github.com/mailapi/mediawiki-extensions-MailAPI.git' 'MailAPI' \
    && git clone --depth=1 --branch 'v0.1.4' 'https://github.com/jmnote/SimpleMermaid.git' 'SimpleMermaid' \
    && cd /app/w \
    && composer update --minimal-changes --no-dev --no-scripts --optimize-autoloader \
    && chown www-data:www-data -R /app/*
# /extensions-sync: generated; do not edit
