# syntax=docker/dockerfile:1
FROM node:22-bookworm-slim AS node

FROM golang:1.26-trixie AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,id=tokens-go-mod,target=/go/pkg/mod,sharing=locked go mod download
COPY cmd ./cmd
COPY source ./source
COPY catalog ./catalog
COPY forge ./forge
COPY skills ./skills
COPY schemas ./schemas
RUN --mount=type=cache,id=tokens-go-mod,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,id=tokens-go-build,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -o /out/tokens ./cmd/tokens

# Derive site assets without changing approved publication generations.
FROM node AS asset-builder
WORKDIR /src
COPY package.json package-lock.json ./
RUN npm ci
COPY forge ./forge
COPY catalog/web/static ./catalog/web/static
COPY --from=build /out/tokens /usr/local/bin/tokens
COPY deploy/build-assets.sh ./deploy/build-assets.sh
ENV HOME=/tmp
USER 10001:10001
ENTRYPOINT ["sh", "deploy/build-assets.sh"]

FROM debian:trixie-slim AS runtime-base
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --gid 10001 tokens \
    && useradd --uid 10001 --gid tokens --home-dir /tmp --no-create-home tokens
WORKDIR /var/lib/tokens
ENV HOME=/tmp
USER 10001:10001
ENTRYPOINT ["/usr/local/bin/tokens"]

FROM runtime-base AS runtime
COPY --from=build /out/tokens /usr/local/bin/tokens
COPY inputs/page-catalog.json /etc/tokens/page-catalog.json

# Install Chromium before adding application files so code changes reuse it.
FROM runtime-base AS updater
USER root
RUN apt-get update && apt-get install -y --no-install-recommends chromium fonts-liberation \
    && rm -rf /var/lib/apt/lists/*
USER 10001:10001
COPY --from=build /out/tokens /usr/local/bin/tokens
COPY inputs/page-catalog.json /etc/tokens/page-catalog.json

FROM updater AS container-test
COPY deploy/testdata /fixtures

FROM runtime AS server
EXPOSE 8080
CMD ["serve", "--published", "/var/lib/tokens/current", "--assets", "/var/lib/tokens/site-assets", "--listen", "0.0.0.0:8080"]
