FROM node:24-trixie-slim AS nodebuild

RUN corepack enable && corepack prepare pnpm@11 --activate

WORKDIR /app
COPY . .
RUN pnpm -C svelte                    install --frozen-lockfile
RUN pnpm -C mwz/skins/ZetaSkin/svelte install --frozen-lockfile
RUN node hack/version.mjs
RUN pnpm -C svelte                    run build
RUN pnpm -C mwz/skins/ZetaSkin/svelte run build

FROM --platform=$BUILDPLATFORM golang:1.25-trixie AS gobuild

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src/goapp
COPY goapp/go.* ./
RUN go mod download
COPY goapp/ ./
	RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-$(go env GOARCH)} go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server
	RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-$(go env GOARCH)} go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker
	RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-$(go env GOARCH)} go build -trimpath -ldflags="-s -w" -o /out/scheduler ./cmd/scheduler
	RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-$(go env GOARCH)} go build -trimpath -ldflags="-s -w" -o /out/tool ./cmd/tool

# https://github.com/zetaoss/zbase/pkgs/container/zbase
FROM ghcr.io/zetaoss/zbase:v0.43.900

ENV MW_INSTALL_PATH=/app/w

COPY --from=nodebuild /app           /app
COPY --from=gobuild   /out/server    /app/goapp/server
COPY --from=gobuild   /out/worker    /app/goapp/worker
COPY --from=gobuild   /out/scheduler /app/goapp/scheduler
COPY --from=gobuild   /out/tool      /app/goapp/tool

RUN set -eux \
    && mv /var/www/html                         /app/w \
    && ln -rs /app/mwz/extensions/ZetaExtension /app/w/extensions/ \
    && ln -rs /app/mwz/skins/ZetaSkin           /app/w/skins/ \
    && /app/goapp/tool extensions upgrade \
    && chown www-data:www-data -R /app/*
