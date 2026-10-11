# syntax=docker/dockerfile:1
FROM golang:1.26-trixie AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,id=tokens-go-mod,target=/go/pkg/mod,sharing=locked go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY schemas ./schemas
RUN --mount=type=cache,id=tokens-go-mod,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,id=tokens-go-build,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -o /out/tokens ./cmd/tokens

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
CMD ["serve", "--published", "/var/lib/tokens/current", "--listen", "0.0.0.0:8080"]
