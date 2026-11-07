#!/bin/sh
set -eu
state_dir=${1:-/var/lib/tokens}
# current remains immutable; each server pins one completed asset directory.
output=$(mktemp -d "$state_dir/.site-assets-XXXXXXXX")
trap 'rm -rf "$output"' EXIT HUP INT TERM
node forge/build-tokens.mjs --published "$state_dir/current" --assets "$output/assets" --check-bindings --css-names "${TOKENS_CSS_NAMES:-short}" --tokens-binary /usr/local/bin/tokens
# Keep content-addressed image URLs usable for pages open during a release.
if [ -d "$state_dir/site-assets/illustrations" ]; then
    mkdir -p "$output/assets/illustrations"
    for image in "$state_dir"/site-assets/illustrations/*.jpg; do
        [ -f "$image" ] || continue
        cp "$image" "$output/assets/illustrations/"
    done
fi
/usr/local/bin/tokens illustrations --published "$state_dir/current" --assets "$output/assets" --chrome-path /usr/bin/chromium
ln -s "$output/assets" "$output/link"
mv -Tf "$output/link" "$state_dir/site-assets"
# Retain old directories: an existing server may still be serving their downloads.
trap - EXIT HUP INT TERM
