### Dockerfile
FROM node:24-trixie-slim AS nodebuild

ARG APP_VERSION=v0.0.0

RUN corepack enable && corepack prepare pnpm@10 --activate

WORKDIR /app
COPY . .
RUN pnpm -C svelte                    install --frozen-lockfile
RUN pnpm -C mwz/skins/ZetaSkin/svelte install --frozen-lockfile
RUN pnpm -C svelte                    run build
RUN pnpm -C mwz/skins/ZetaSkin/svelte run build

RUN APP_VERSION_NORMALIZED="${APP_VERSION#v}" \
    && sed -i "s/\"version\": \".*\"/\"version\": \"${APP_VERSION_NORMALIZED}\"/" /app/mwz/skins/ZetaSkin/skin.json \
    && sed -i "s/\"version\": \".*\"/\"version\": \"${APP_VERSION_NORMALIZED}\"/" /app/mwz/extensions/ZetaExtension/extension.json \
    && echo ok

# https://github.com/zetaoss/zbase/pkgs/container/zbase
FROM ghcr.io/zetaoss/zbase:v0.43.621

ARG APP_VERSION=v0.0.0
ENV APP_VERSION=${APP_VERSION}

COPY --from=nodebuild /app /app

RUN set -eux \
    && mv /var/www/html                         /app/w \
    && ln -rs /app/mwz/extensions/ZetaExtension /app/w/extensions/ \
    && ln -rs /app/mwz/skins/ZetaSkin           /app/w/skins/ \
    && cd /app/laravel/ && composer install --no-dev -o \
    && chown www-data:www-data -R /app/*
