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

# Tools compile to static files. Neither OpenCV nor FFmpeg native libraries are
# needed in this build stage or in the production server.
FROM build AS tool-build
USER root
COPY --from=node /usr/local/bin/node /usr/local/bin/node
COPY --from=node /usr/local/lib/node_modules /usr/local/lib/node_modules
RUN ln -s /usr/local/lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm \
    && ln -s /src /tokens
WORKDIR /ui-motion
COPY --from=material_motion /package.json /package-lock.json /go.mod /go.sum /Makefile /tsconfig.json /vite.config.ts /.oxlintrc.json ./
COPY --from=material_motion /web ./web
COPY --from=material_motion /public ./public
COPY --from=material_motion /scripts ./scripts
COPY --from=material_motion /internal/web ./internal/web
COPY --from=material_motion /internal/motion ./internal/motion
COPY --from=material_motion /internal/analysis/example ./internal/analysis/example
COPY --from=material_motion /cmd/web-static ./cmd/web-static
RUN --mount=type=cache,id=tools-npm,target=/root/.npm,sharing=locked npm ci
RUN --mount=type=cache,id=tokens-go-mod,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,id=tokens-go-build,target=/root/.cache/go-build \
    make build-static
WORKDIR /carousel-motion-lab
COPY --from=carousel_playground /package.json ./
COPY --from=carousel_playground /scripts ./scripts
COPY --from=carousel_playground /src ./src
COPY --from=carousel_playground /site ./site
RUN --mount=type=cache,id=tokens-go-mod,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,id=tokens-go-build,target=/root/.cache/go-build \
    npm run build

FROM runtime AS server
COPY --from=tool-build /ui-motion/dist/ /opt/tools/motion/
COPY --from=tool-build /carousel-motion-lab/site/ /opt/tools/motion-lab/
ENV TOKENS_TOOLS_DIR=/opt/tools
EXPOSE 8080
CMD ["serve", "--published", "/var/lib/tokens/current", "--assets", "/var/lib/tokens/site-assets", "--listen", "0.0.0.0:8080"]
