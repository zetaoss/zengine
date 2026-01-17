### Dockerfile
# https://github.com/zetaoss/zbase/pkgs/container/zbase
FROM ghcr.io/zetaoss/zbase:v0.43.602

ARG ZENGINE_VERSION=v0.0.0
ENV ZENGINE_VERSION=${ZENGINE_VERSION}

# https://hub.docker.com/_/composer
COPY --from=composer:2.9.2 /usr/bin/composer /usr/bin/composer

# dependencies
COPY laravel/composer.json laravel/composer.lock                               /app/laravel/
COPY vue/package.json vue/pnpm-lock.yaml                                       /app/vue/
COPY mwz/skins/ZetaSkin/vue/package.json mwz/skins/ZetaSkin/vue/pnpm-lock.yaml /app/w/skins/ZetaSkin/vue/

RUN cd /app/laravel/                 && composer install --no-scripts --no-autoloader \
    && cd /app/vue/                  && pnpm install --frozen-lockfile \
    && cd /app/w/skins/ZetaSkin/vue/ && pnpm install --frozen-lockfile

COPY . /app/

RUN set -eux \
    && mv     /var/www/html                     /app/w \
    && cp -a  /app/w/composer.local.json-sample /app/w/composer.local.json \
    && ln -rs /app/mwz/extensions/ZetaExtension /app/w/extensions/ \
    && ln -rs /app/mwz/skins/ZetaSkin           /app/w/skins/ \
    && cd /app/laravel/              && composer dump-autoload --optimize \
    && cd /app/vue/                  && pnpm run build \
    && cd /app/w/                    && composer update \
    && cd /app/w/skins/ZetaSkin/vue/ && pnpm run build \
    && chown www-data:www-data -R /app/*
