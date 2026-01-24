### Dockerfile
# https://github.com/zetaoss/zbase/pkgs/container/zbase
FROM ghcr.io/zetaoss/zbase:v0.43.607

ARG APP_VERSION=v0.0.0
ENV APP_VERSION=${APP_VERSION}

# https://hub.docker.com/_/composer
COPY --from=composer:2.9.2 /usr/bin/composer /usr/bin/composer

COPY . /app/

RUN set -eux \
    && APP_VERSION_STRIPPED="${APP_VERSION#v}" \
    && sed -i "s/\"version\": \".*\"/\"version\": \"${APP_VERSION_STRIPPED}\"/" /app/mwz/skins/ZetaSkin/skin.json \
    && mv     /var/www/html                     /app/w \
    && cp -a  /app/w/composer.local.json-sample /app/w/composer.local.json \
    && ln -rs /app/mwz/extensions/ZetaExtension /app/w/extensions/ \
    && ln -rs /app/mwz/skins/ZetaSkin           /app/w/skins/ \
    && cd /app/laravel/              && composer install --no-scripts --optimize-autoloader \
    && cd /app/vue/                  && pnpm install --frozen-lockfile && pnpm run build \
    && cd /app/w/                    && composer update \
    && cd /app/w/skins/ZetaSkin/vue/ && pnpm install --frozen-lockfile && pnpm run build \
    && chown www-data:www-data -R /app/*
