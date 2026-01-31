### Dockerfile
FROM node:24-trixie-slim AS nodebuild

ARG APP_VERSION=v0.0.0

RUN corepack enable pnpm

COPY . /app/
RUN cd /app/mwz/skins/ZetaSkin/vue/ && pnpm install --frozen-lockfile
RUN cd /app/vue/                    && pnpm install --frozen-lockfile
RUN cd /app/mwz/skins/ZetaSkin/vue/ && pnpm run build
RUN cd /app/vue/                    && pnpm run build

RUN APP_VERSION_STRIPPED="${APP_VERSION#v}" && sed -i "s/\"version\": \".*\"/\"version\": \"${APP_VERSION_STRIPPED}\"/" /app/mwz/skins/ZetaSkin/skin.json

# https://github.com/zetaoss/zbase/pkgs/container/zbase
FROM ghcr.io/zetaoss/zbase:v0.43.611

ARG APP_VERSION=v0.0.0
ENV APP_VERSION=${APP_VERSION}

COPY .                                        /app/
COPY --from=nodebuild /app/mwz/skins/ZetaSkin /app/w/skins/ZetaSkin
COPY --from=nodebuild /app/vue                /app/vue

RUN set -eux \
    && mv /app/mwz/extensions/ZetaExtension /app/w/extensions/ZetaExtension \
    && cd /app/laravel/ && composer install --no-dev --no-scripts --optimize-autoloader \
    && cd /app/w/       && composer install --no-dev --no-scripts --optimize-autoloader --no-security-blocking \
    && chown www-data:www-data -R /app/*
