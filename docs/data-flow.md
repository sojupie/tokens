# Token Catalog data flow

Token Catalog has two jobs:

1. crawl Material's pages and save the source data it publishes;
2. create and verify an immutable material-token-data snapshot.

## The flow

```text
inputs/page-catalog.json
        |
        v
    tokens crawl
        |
        +--> .cache/material-token-table-responses/
        |      downloaded source-payload files
        |
        +--> crawl-results/<crawl-id>.json
               crawl receipt
        |
        v
    tokens snapshot create
        |
        v
snapshots/material-token-data/<snapshot-id>/
        |
        +--> manifest.json
        +--> page-catalog.json
        +--> source-payloads/
        +--> generated-token-graph.json.gz
        +--> source-payload-details.json
        +--> page-results.json
        +--> lifecycle-notices.json
        +--> checksums.txt
        |
        +--> tokens snapshot verify
        +--> tokens snapshot compare
```

Verified snapshots can then be published for read-only browsing:

```text
verified snapshot(s)
        |
        v
    tokens publish
        |
        +--> published/tokens.sqlite
        +--> published/publication.json
        +--> published/downloads/
        |
        v
    tokens serve
        |
        +--> in-memory published views
        +--> Templ HTML pages
        +--> HTMX fragments
        +--> JSON API
        +--> downloads
```

The generated SQLite file and downloads are derived publication data. They can
be deleted and rebuilt from verified snapshots. The web server is read-only and
does not crawl Material or modify snapshots.

To repair a partial crawl, `tokens crawl --retry-failed` reads the newest
crawl receipt, opens only its failed pages, and writes another merged crawl
receipt. Successful pages and source payloads are carried forward; retried
pages replace their earlier results.

## What is fetched?

The crawler opens each selected Material token-source page with Chrome. Normal
crawls select pages from the reviewed page URL catalog. Broad discovery can
also consider component pages ending in `/specs` and style or foundation pages
ending in `/tokens`, `/specs`, `-tokens`, or `-scale`. The crawler watches each
page's network traffic for uppercase source-payload JSON responses. It then
downloads each unique source payload and stores the compressed response in the
source-payload cache.

The page HTML is observed but is not a snapshot input. The matching
source-payload response is the fetched source data. Different Material areas
use different payload filename stems, so Token Catalog does not assume that
every source payload is named `TOKEN_TABLE`.

## What is generated?

The crawl generates a crawl receipt. It records the pages opened, source-payload
URLs observed, downloaded file hashes, retrieval metadata, and errors.

Snapshot creation reads the crawl receipt and source-payload files. It
generates:

- `generated-token-graph.json.gz`: generic Material records extracted from the
  source payloads;
- `source-payload-details.json`: source URLs, hashes, retrieval details, and
  errors;
- `page-results.json`: page and extraction results;
- `manifest.json`: snapshot identity, inputs, generator version, and file list;
- `checksums.txt`: integrity records for every snapshot file.

The original successful source-payload files are copied into the snapshot.
They are not regenerated from the graph.

## Names and states

`crawl` means opening pages and downloading source payloads.

`snapshot create` means assembling source payloads and generated token data into
one immutable bundle.

`snapshot verify` checks file hashes, checksums, JSON schemas, and bundled source
payloads. Capture completeness and publication approval are separate checks.

`snapshot compare` means comparing two verified bundles.

`captureStatus` describes whether the crawl obtained all expected data:

- `complete` — all expected pages and source payloads succeeded;
- `partial` — some expected data is missing or failed.

`reviewStatus` records the snapshot's review state:

- `provisional` — evidence only;
- `approved` — accepted after review;
- `rejected` — not accepted.

Capture status and review status are separate. A snapshot can be successfully
assembled while still being partial and provisional.
