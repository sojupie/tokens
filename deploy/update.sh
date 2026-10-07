#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
if [ "$#" -eq 0 ]; then
    echo 'Usage: sh deploy/update.sh run|status|inspect RUN_ID|approve RUN_ID|retry RUN_ID|test-notification' >&2
    exit 1
fi
if [ "$1" = run ]; then
    shift
    set -- run --page-catalog /etc/tokens/page-catalog.json \
        --chrome-path /usr/bin/chromium --concurrency 1 --page-timeout 30s --timeout 1h "$@"
fi
docker compose -f deploy/compose.yaml run --rm --no-deps updater \
    update "$@" --state-dir /var/lib/tokens
