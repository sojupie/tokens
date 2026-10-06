#!/bin/sh
set -eu
directory=/var/lib/tokens/fixture
mkdir -p "$directory"
gzip -c /fixtures/source.json > "$directory/payload.json.gz"
hash=$(sha256sum /fixtures/source.json | cut -d ' ' -f 1)
bytes=$(wc -c < /fixtures/source.json)
printf '%s\n' '{"schemaVersion":2,"source":"fixture","scope":"component specs","generatedAt":"2026-06-01T12:00:00Z","pages":[{"url":"https://m3.material.io/components/snackbar/specs","area":"components","subject":"snackbar","pageType":"specs"}]}' > "$directory/catalog.json"
printf '%s\n' "{\"schemaVersion\":2,\"crawlId\":\"container-fixture\",\"startedAt\":\"2026-09-01T00:00:00Z\",\"completedAt\":\"2026-09-01T00:00:00Z\",\"options\":{\"pageCatalogPath\":\"$directory/catalog.json\",\"cacheDir\":\"$directory\"},\"pages\":[\"https://m3.material.io/components/snackbar/specs\"],\"observations\":[{\"pageURL\":\"https://m3.material.io/components/snackbar/specs\",\"status\":200,\"sourcePayloadURLs\":[\"https://m3.material.io/_dsm/data/dsdb-m3/TOKEN_TABLE.fixture.json\"]}],\"sourcePayloads\":[{\"pageURLs\":[\"https://m3.material.io/components/snackbar/specs\"],\"sourcePayloadURL\":\"https://m3.material.io/_dsm/data/dsdb-m3/TOKEN_TABLE.fixture.json\",\"sourcePayloadKind\":\"component\",\"sourceArea\":\"components\",\"sourceSubject\":\"snackbar\",\"sourcePageType\":\"specs\",\"hash\":\"$hash\",\"bytes\":$bytes,\"artifact\":\"payload.json.gz\"}]}" > "$directory/receipt.json"
