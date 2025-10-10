FROM ghcr.io/zetaoss/zbase:latest

COPY --from=composer:2.2.25 /usr/bin/composer /usr/bin/composer

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
