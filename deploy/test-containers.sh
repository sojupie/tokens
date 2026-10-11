#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
project=tokens-container-check
compose() {
    docker compose -p "$project" -f deploy/compose.yaml -f deploy/compose.test.yaml "$@"
}
cleanup() {
    compose --profile update down --volumes --remove-orphans
}
trap cleanup EXIT
compose --profile update config --quiet
compose build tokens updater
docker volume create --label "com.docker.compose.project=$project" \
    --label com.docker.compose.volume=test-state "${project}_test-state" >/dev/null
docker run --rm --user 0 --entrypoint sh \
    -v "${project}_test-state:/var/lib/tokens" tokens-updater-check:local \
    -c 'install -d -o 10001 -g 10001 -m 0700 /var/lib/tokens && touch /var/lib/tokens/.container-check'
compose run --rm --no-deps --entrypoint /bin/sh updater /fixtures/prepare.sh
document=$(compose run --rm --no-deps --entrypoint /usr/bin/chromium updater \
    --headless=new --disable-gpu --disable-dev-shm-usage --disable-background-networking \
    --dump-dom about:blank)
printf '%s' "$document" | grep -q '<html>'

if compose run --rm --no-deps updater update run --state-dir /var/lib/tokens \
    --receipt /var/lib/tokens/fixture/receipt.json; then
    echo 'Expected baseline approval to be required.' >&2
    exit 1
fi
status=$(compose run --rm --no-deps updater update status --state-dir /var/lib/tokens)
printf '%s' "$status" | grep -q needs_review
run_id=$(printf '%s\n' "$status" | awk 'NR == 1 {print $1}')
compose run --rm --no-deps updater update approve "$run_id" --state-dir /var/lib/tokens
compose up -d --wait tokens
compose exec -T tokens curl --fail --silent --output /dev/null http://127.0.0.1:8080/md3/components/snackbar
robots=$(compose exec -T tokens curl --fail --silent http://127.0.0.1:8080/robots.txt)
printf '%s' "$robots" | grep -q 'Sitemap: https://tokencatalog.dev/md3/sitemap.xml'
compose exec -T tokens sh -c 'test ! -w /var/lib/tokens/current/tokens.sqlite; ! command -v chromium'
compose run --rm --no-deps updater update status --state-dir /var/lib/tokens
echo 'Container checks passed: Chromium, review, activation, serving, public URL and read-only state.'
