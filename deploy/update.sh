#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
if [ "$#" -eq 0 ]; then
    echo 'Usage: sh deploy/update.sh run|status|inspect RUN_ID|approve RUN_ID|retry RUN_ID|test-notification|configure-public-origin URL' >&2
    exit 1
fi
if [ "$1" = configure-public-origin ]; then
    if [ "$#" -ne 2 ]; then
        echo 'Usage: sudo sh deploy/update.sh configure-public-origin https://tokencatalog.dev/md3' >&2
        exit 1
    fi
    public_origin=${2%/}
    case "$public_origin" in
        *[!A-Za-z0-9:./-]*)
            echo 'Public URL must be an HTTPS origin without credentials, query or fragment; an optional clean path is allowed.' >&2
            exit 1
            ;;
    esac
    if ! printf '%s\n' "$public_origin" | LC_ALL=C grep -Eq '^https://[A-Za-z0-9]([A-Za-z0-9.-]*[A-Za-z0-9])?(:[0-9]{1,5})?(/[A-Za-z0-9_-]+)*$'; then
        echo 'Public URL must be an HTTPS origin without credentials, query or fragment; an optional clean path is allowed.' >&2
        exit 1
    fi
    if [ "$(id -u)" != 0 ]; then
        echo 'Run configure-public-origin with sudo.' >&2
        exit 1
    fi
    install -d -m 0755 /etc/tokens
    config_tmp=$(mktemp /etc/tokens/server.env.XXXXXX)
    trap 'rm -f "$config_tmp"' EXIT HUP INT TERM
    printf 'TOKENS_PUBLIC_ORIGIN=%s\n' "$public_origin" > "$config_tmp"
    chmod 0600 "$config_tmp"
    mv -f "$config_tmp" /etc/tokens/server.env
    echo "Public URL saved: $public_origin"
    exit 0
fi
if [ "$1" = run ]; then
    shift
    set -- run --page-catalog /etc/tokens/page-catalog.json \
        --chrome-path /usr/bin/chromium --concurrency 1 --page-timeout 30s --timeout 1h "$@"
fi
docker compose -f deploy/compose.yaml run --rm --no-deps updater \
    update "$@" --state-dir /var/lib/tokens
