FROM ghcr.io/zetaoss/zbase:latest

COPY --from=composer:2.2.25 /usr/bin/composer /usr/bin/composer

RUN set -eux \
    && cd / \
    && git clone https://github.com/zetaoss/zengine.git app \
    && cd /app/ \
    && mv     /var/www/html                     /app/w \
    && cp -a  /app/w/composer.local.json-sample /app/w/composer.local.json \
    && ln -rs /app/mwz/extensions/ZetaExtension /app/w/extensions/ \
    && ln -rs /app/mwz/skins/ZetaSkin           /app/w/skins/ \
    && cd /app/laravel/              && composer install \
    && cd /app/vue/                  && npm install \
    && cd /app/w/                    && composer update \
    && cd /app/w/skins/ZetaSkin/vue/ && npm install && npm run build \
    && chown www-data:www-data -R /app/*
