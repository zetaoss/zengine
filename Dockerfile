### Dockerfile
# https://github.com/zetaoss/zbase/pkgs/container/zbase
FROM ghcr.io/zetaoss/zbase:v0.43.602

ARG ZENGINE_VERSION=v0.0.0
ENV ZENGINE_VERSION=${ZENGINE_VERSION}

# https://hub.docker.com/_/composer
COPY --from=composer:2.9.2 /usr/bin/composer /usr/bin/composer

COPY . /app/

RUN set -eux \
    && mv     /var/www/html                     /app/w \
    && cp -a  /app/w/composer.local.json-sample /app/w/composer.local.json \
    && ln -rs /app/mwz/extensions/ZetaExtension /app/w/extensions/ \
    && ln -rs /app/mwz/skins/ZetaSkin           /app/w/skins/ \
    && cd /app/laravel/              && composer install \
    && cd /app/vue/                  && pnpm install --frozen-lockfile && pnpm run build \
    && cd /app/w/                    && composer update \
    && cd /app/w/skins/ZetaSkin/vue/ && pnpm install --frozen-lockfile && pnpm run build \
    && chown www-data:www-data -R /app/*
