FROM golang:1.26-trixie AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY schemas ./schemas
RUN CGO_ENABLED=0 go build -trimpath -o /out/tokens ./cmd/tokens

FROM debian:trixie-slim AS runtime
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --gid 10001 tokens \
    && useradd --uid 10001 --gid tokens --home-dir /tmp --no-create-home tokens
COPY --from=build /out/tokens /usr/local/bin/tokens
COPY inputs/page-catalog.json /etc/tokens/page-catalog.json
WORKDIR /var/lib/tokens
ENV HOME=/tmp
USER 10001:10001
ENTRYPOINT ["/usr/local/bin/tokens"]

FROM runtime AS updater
USER root
RUN apt-get update && apt-get install -y --no-install-recommends chromium fonts-liberation \
    && rm -rf /var/lib/apt/lists/*
USER 10001:10001

FROM updater AS container-test
COPY deploy/testdata /fixtures

FROM runtime AS server
EXPOSE 8080
CMD ["serve", "--published", "/var/lib/tokens/current", "--listen", "0.0.0.0:8080", "--public-origin", "https://tokens.marmenlind.com"]
