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

FROM --platform=$BUILDPLATFORM golang:1.26-trixie AS gobuild

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src/goapp
COPY goapp/go.* ./
RUN go mod download
COPY goapp/ ./
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-$(go env GOARCH)} go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-$(go env GOARCH)} go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-$(go env GOARCH)} go build -trimpath -ldflags="-s -w" -o /out/ctl ./cmd/ctl

# https://github.com/zetaoss/zbase/pkgs/container/zbase
FROM ghcr.io/zetaoss/zbase:v0.43.800

ARG APP_VERSION=v0.0.0
ENV APP_VERSION=${APP_VERSION}

COPY --from=nodebuild /app          /app
COPY --from=gobuild   /out/server   /app/goapp/server
COPY --from=gobuild   /out/worker   /app/goapp/worker
COPY --from=gobuild   /out/ctl      /app/goapp/ctl

RUN set -eux \
    && mv /var/www/html                         /app/w \
    && ln -rs /app/mwz/extensions/ZetaExtension /app/w/extensions/ \
    && ln -rs /app/mwz/skins/ZetaSkin           /app/w/skins/ \
    && chown www-data:www-data -R /app/*
