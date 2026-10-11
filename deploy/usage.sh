#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
docker compose -f deploy/compose.yaml --profile usage run --rm --no-deps usage usage "$@"
