# Source Loom data flow

Source Loom has two jobs:

1. crawl Material's pages and save the source data it publishes;
2. create and verify an immutable material-token-data snapshot.

It does not decide which tokens Token Forge supports.

## The flow

```text
inputs/page-catalog.json
        |
        v
    sourceloom crawl
        |
        +--> .cache/material-token-table-responses/
        |      downloaded source-payload files
        |
        +--> crawl-results/<crawl-id>.json
               crawl receipt
        |
        v
    sourceloom snapshot create
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
        +--> sourceloom snapshot verify
        +--> sourceloom snapshot compare
```

To repair a partial crawl, `sourceloom crawl --retry-failed` reads the newest
crawl receipt, opens only its failed pages, and writes another merged crawl
receipt. Successful pages and source payloads are carried forward; retried
pages replace their earlier results.

## What is fetched?

The crawler opens each selected Material token-source page with Chrome. Normal
crawls select pages from the reviewed page catalog. Broad discovery can also
consider component pages ending in `/specs` and style or foundation pages
ending in `/tokens` or `/specs`. The crawler watches each page's network
traffic for uppercase source-payload JSON responses. It then
downloads each unique source payload and stores the compressed response in the
source-payload cache.

The page HTML is observed but is not a snapshot input. The token-table response
is the fetched source data.

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

`snapshot verify` means checking the bundle's file hashes and checksums. It does
not decide whether the Material data is complete or suitable for Forge.

`snapshot compare` means comparing two verified bundles.

`captureStatus` describes whether the crawl obtained all expected data:

- `complete` — all expected pages and source payloads succeeded;
- `partial` — some expected data is missing or failed.

`reviewStatus` describes whether a person has approved the snapshot for Forge:

- `provisional` — evidence only;
- `approved` — accepted as a Forge input;
- `rejected` — not accepted.

Capture status and review status are separate. A snapshot can be successfully
assembled while still being partial and provisional.

## Token Forge boundary

Token Forge consumes a pinned snapshot bundle. It does not run the crawler,
access Material's website, or depend on the Source Loom executable during normal
builds. Token Forge candidates point to the snapshot ID, manifest hash,
source-payload ID, and source record ID. Canonical Token Forge tokens require a
separate review decision.
