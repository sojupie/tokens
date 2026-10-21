#!/bin/sh
set -eu
state_dir=${1:-/var/lib/tokens}
# current remains immutable; each server pins one completed asset directory.
output=$(mktemp -d "$state_dir/.site-assets-XXXXXXXX")
trap 'rm -rf "$output"' EXIT HUP INT TERM
node forge/build-tokens.mjs --published "$state_dir/current" --assets "$output/assets" --check-bindings --tokens-binary /usr/local/bin/tokens
ln -s "$output/assets" "$output/link"
mv -Tf "$output/link" "$state_dir/site-assets"
# Retain old directories: an existing server may still be serving their downloads.
trap - EXIT HUP INT TERM
