### Dockerfile
FROM node:24-trixie-slim AS nodebuild

RUN corepack enable pnpm

COPY mwz vue /app/

WORKDIR /app/mwz/skins/ZetaSkin/vue
RUN pnpm install --frozen-lockfile
RUN pnpm run build

WORKDIR /app/vue
RUN pnpm install --frozen-lockfile
RUN pnpm run build

# https://github.com/zetaoss/zbase/pkgs/container/zbase
FROM ghcr.io/zetaoss/zbase:v0.43.611

ARG APP_VERSION=v0.0.0
ENV APP_VERSION=${APP_VERSION}

COPY laravel mwz vue /app/

COPY --from=nodebuild /app/vue/dist                          /app/vue/dist
COPY --from=nodebuild /app/mwz/skins/ZetaSkin/resources/dist /app/mwz/skins/ZetaSkin/resources/dist

RUN set -eux \
    && APP_VERSION_STRIPPED="${APP_VERSION#v}" \
    && sed -i "s/\"version\": \".*\"/\"version\": \"${APP_VERSION_STRIPPED}\"/" /app/mwz/skins/ZetaSkin/skin.json \
    && ln -rs /app/mwz/extensions/ZetaExtension /app/w/extensions/ \
    && ln -rs /app/mwz/skins/ZetaSkin           /app/w/skins/ \
    && cd /app/laravel/ && composer install --no-dev --no-scripts --optimize-autoloader \
    && cd /app/w/       && composer install --no-dev --no-scripts --optimize-autoloader --no-security-blocking \
    && chown www-data:www-data -R /app/*
