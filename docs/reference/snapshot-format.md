# Snapshot format (v3)

A snapshot is an immutable bundle of Material's fetched token-table payloads and
the extracted token graph at one point in time, preserving unresolved references
and unknown fields so downstream publishing and diffing work offline.

## Bundle layout

```text
material-token-data-<snapshot-id>/
|-- manifest.json
|-- page-catalog.json
|-- source-payloads/<hash-prefix>/<sha256>.json.gz
|-- generated-token-graph.json.gz
|-- source-payload-details.json
|-- page-results.json
|-- lifecycle-notices.json
`-- checksums.txt
```

## Identity and manifest

Snapshot IDs use `<family>-YYYY-MM-DD-HHMM` in UTC (e.g.
`material-2026-08-11-2348`). The manifest SHA-256—not the human-readable
ID—provides content identity. Never overwrite files under an existing snapshot
ID; create a new snapshot instead. Historical pre-v3 snapshots (`m3-...` IDs,
`m3source` extractor) remain valid immutable evidence and are not rewritten.

```json
{
  "type": "material-token-data-snapshot",
  "schemaVersion": 3,
  "snapshotId": "material-2026-08-11-2348",
  "createdAt": "2026-08-11T00:00:00Z",
  "reviewStatus": "provisional",
  "source": {
    "system": "material-3",
    "pageCatalogSha256": "<sha256>",
    "pageCount": 36
  },
  "capture": {
    "crawlId": "20260811T000000.000000000Z",
    "crawlReceipt": "20260811T000000.000000000Z.json",
    "pagesAttempted": 36,
    "pagesSucceeded": 36,
    "payloadsObserved": 36,
    "payloadsFetched": 36,
    "complete": true
  },
  "extractor": {
    "name": "tokens",
    "version": "0.1.0",
    "schemaVersion": 3
  },
  "artifacts": [
    {
      "path": "generated-token-graph.json.gz",
      "role": "generated-token-data",
      "mediaType": "application/json",
      "contentEncoding": "gzip",
      "sha256": "<sha256>",
      "bytes": 12345
    }
  ]
}
```

`tokens snapshot verify` rejects duplicate paths, invalid hashes, unsupported
schema versions, and missing files.

## Artifacts

- **`generated-token-graph.json.gz`:** Loss-preserving Material records
  (canonical/display names, upstream IDs, token sets, states, contexts, literals,
  reference trees, units, unknown fields, and source classification:
  `sourcePayloadURL`, `pageURLs`, `sourcePayloadKind`, `sourceArea`,
  `sourceSubject`, `sourcePageType`).
- **`source-payload-details.json`:** Payload URLs, classification, hashes,
  retrieval metadata, upstream revision details, and extraction notes keyed by
  stable source IDs.
- **`page-results.json`:** Per-page observed payloads, extracted record counts,
  warnings, and errors so missing payloads stay visible.
- **`lifecycle-notices.json`:** Explicit upstream `deprecated` and
  `not-recommended` notices scoped to `component`, `style`, `token`, or `record`,
  with raw message and source record IDs.

## Diffing

`tokens snapshot compare` verifies both snapshots and diffs their page catalogs,
page results, capture completeness, payloads, extracted records, and lifecycle
notices:

- Payloads match by stable upstream filename (ignoring release directories in
  URLs), so a release-directory bump shows as one updated payload rather than a
  delete + add.
- Records match by `(sourcePayload, recordKind, upstreamRecordID)` and compare
  the full `raw` object, ignoring snapshot/crawl IDs, timestamps, and local
  paths.
- The CLI prints a semantic view by default (resolved values and component/set
  fields, deduplicated across shared payloads, ignoring internal version IDs).

## Consumer rules

Pin a snapshot ID and manifest SHA-256, verify the manifest and artifact hashes
before parsing, and reject unsupported schema versions.
