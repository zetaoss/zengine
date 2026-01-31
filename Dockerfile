### Dockerfile
FROM node:24-trixie-slim AS nodebuild

COPY package.json                                                              /app/
COPY vue/package.json                    vue/pnpm-lock.yaml                    /app/vue/
COPY mwz/skins/ZetaSkin/vue/package.json mwz/skins/ZetaSkin/vue/pnpm-lock.yaml /app/mwz/skins/ZetaSkin/vue/

RUN corepack enable pnpm \
    && cd /app/vue                    && pnpm install --frozen-lockfile \
    && cd /app/mwz/skins/ZetaSkin/vue && pnpm install --frozen-lockfile

COPY . /app
RUN echo run build \
    && cd /app/vue                    && pnpm run build \
    && cd /app/mwz/skins/ZetaSkin/vue && pnpm run build

# https://github.com/zetaoss/zbase/pkgs/container/zbase
FROM ghcr.io/zetaoss/zbase:v0.43.609

ARG APP_VERSION=v0.0.0
ENV APP_VERSION=${APP_VERSION}

# https://hub.docker.com/_/composer
COPY --from=composer:2.9.2 /usr/bin/composer /usr/bin/composer

COPY . /app/

COPY --from=nodebuild /app/vue/dist                          /app/vue/dist
COPY --from=nodebuild /app/mwz/skins/ZetaSkin/resources/dist /app/mwz/skins/ZetaSkin/resources/dist

RUN set -eux \
    && APP_VERSION_STRIPPED="${APP_VERSION#v}" \
    && sed -i "s/\"version\": \".*\"/\"version\": \"${APP_VERSION_STRIPPED}\"/" /app/mwz/skins/ZetaSkin/skin.json \
    && rm -rf /var/www/html \
    && ln -s  /app/w                            /var/www/html \
    && cp -a  /app/w/composer.local.json-sample /app/w/composer.local.json \
    && ln -rs /app/mwz/extensions/ZetaExtension /app/w/extensions/ \
    && ln -rs /app/mwz/skins/ZetaSkin           /app/w/skins/ \
    && cd /app/laravel/ && composer install --no-dev --no-scripts --optimize-autoloader \
    && cd /app/w/       && composer install --no-dev --no-scripts --optimize-autoloader --no-security-blocking \
    && chown www-data:www-data -R /app/*
