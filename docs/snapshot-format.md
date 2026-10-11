# Token Catalog snapshot format

**Status:** Current implementation contract, version 3
**Date:** 2026-08-12

## Purpose

A snapshot is an immutable bundle of Material's fetched token-table source
payloads and the generated token data extracted from them at one point in time.
It provides a stable data source without requiring the
crawl tool, a browser, a live API, or Material's website.

The snapshot preserves source meaning, including unresolved references and
unknown fields.

## Bundle layout

```text
material-token-data-<snapshot-id>/
|-- manifest.json
|-- page-catalog.json             # Material token-source page URL catalog
|-- source-payloads/<hash-prefix>/<sha256>.json.gz
|-- generated-token-graph.json.gz
|-- source-payload-details.json
|-- page-results.json
|-- lifecycle-notices.json
`-- checksums.txt
```

Successful content-addressed payload artifacts are part of the snapshot. A
publisher may provide a separately indexed or compressed distribution, but it
must contain the same verified payload bytes.

## Snapshot identity

The publisher serializes release files deterministically, calculates their
SHA-256 hashes, then writes the manifest. The manifest digest is the integrity
identity used by consumers. Current Token Catalog-generated IDs use the source
family, capture date, and capture time in UTC (`HHMM`):

```text
material-2026-08-11-2348
```

The manifest SHA-256, not the human-readable ID, provides content identity.
Existing immutable snapshots may retain historical `m3-...` IDs and the
historical extractor name `m3source`; they must not be rewritten.

Never replace files under a published snapshot ID. Publish a new snapshot when
content, extraction rules, or source metadata changes.

## Manifest

Version 3 has this logical shape:

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

The final schema must state which timestamps are descriptive and which fields
participate in deterministic identity. Snapshot tooling must reject duplicate
paths, invalid hashes, unsupported schema versions, and missing required files.

## Generated token graph

`generated-token-graph.json.gz` is generated from the source payload files. It is
Material-specific and loss-preserving. It retains:

- canonical and display token names;
- upstream token, value, table, and revision IDs;
- token sets, states, contexts, and tags;
- literals and full reference trees;
- original types and units;
- unknown fields required to diagnose source-schema drift;
- source record IDs that join to source-payload details.

Each source-payload and extracted record also carries the source classification
available at capture time:

- `sourcePayloadURL` — the upstream JSON payload URL;
- `pageURLs` — every reviewed page that observed that payload;
- `sourcePayloadKind` — for example `component`, `elevation`, `motion`,
  `spacing`, `typography`, `shape`, `color`, or `unknown`;
- `sourceArea`, `sourceSubject`, and `sourcePageType` — page-catalog context.

These fields describe what the payload represents. A component payload may
contain references to foundation or style tokens, and one payload may be
observed by more than one page.

The graph preserves themes and contexts rather than flattening them into one
value. The token browser, downloads, and change history use this data.

## Source-payload details

Source-payload details record page and source-payload URLs, the same source
classification fields, source-payload hashes, retrieval metadata, upstream
revision details, and extraction notes. Use stable source IDs so generated
records can refer to source-payload details without duplicating long strings.

## Page results

The page-results file lists each page and whether it observed source payloads,
generated records, warnings, or errors. A missing source payload is visible
evidence, not a silently dropped page. The field is plural because a page may
expose more than one payload.

## Lifecycle notices

`lifecycle-notices.json` preserves explicit Material notices about the
lifecycle or recommendation status of source records. It does not infer that a
component is active or deprecated from a missing payload, a 404, or the absence
of a warning.

The notice scope is separate from capture status:

- `component` applies to a component-level notice;
- `style` applies to a token-set or style group;
- `token` applies to an individual token;
- `record` preserves a notice whose source shape does not map to one of those
  scopes.

`deprecated` and `not-recommended` remain separate states. The raw message and
source record identifiers are retained so readers can trace each notice to its
source.

## Publishing and storage

Publish static files to content-addressable object storage, GitHub Releases, or
both. Static HTTP files are the stable contract. Documentation sites, APIs,
packages, and SQLite databases are derived views.

Git stores code, schemas, the curated page URL catalog, fixtures, and documentation.
Object storage holds full snapshot bundles and optional raw evidence. A
per-snapshot SQLite file may provide local query convenience, but it is generated
from the snapshot and is never canonical.

## Consumer rules

A consumer must:

1. Pin a snapshot ID and manifest SHA-256.
2. Download the immutable manifest and verify its digest.
3. Verify every artifact before parsing it.
4. Reject unsupported schema versions.
5. Treat import output as a review candidate unless its own policy states
   otherwise.

A consumer must not rely on mutable `latest` URLs for reproducible builds.

## Diffing

`snapshot compare` verifies both input directories, then compares their page URL
catalogs, page results, capture completeness, content-addressed source payloads,
generic extracted records, and lifecycle notices. It reports added, removed, and
changed page results. Source payloads are matched by their stable upstream
filename, not by the release directory in the URL. A release-directory change
therefore appears as one changed payload with before/after URLs and hashes,
rather than as a removal and an addition. Record changes are keyed by source
payload, record kind, and upstream record ID. Occurrence order is ignored for
unique records and is used only to distinguish duplicate upstream IDs; the
complete upstream object is compared through the record's loss-preserving
`raw` field. It ignores snapshot IDs, crawl IDs, retrieval timestamps, and
artifact paths when deciding whether capture metadata changed.

The diff also contains a semantic view for human review. It compares resolved
token values and selected component or token-set fields, ignores volatile
internal version identifiers, and deduplicates records repeated in shared
Material payloads. The CLI uses this semantic view by default; structural
record and payload details remain available to programmatic consumers.

## Historical snapshots

Snapshots created before version 3 retain their original names and fields. They
are immutable evidence and are not rewritten to match this contract. Current
verification continues to check their file integrity; new snapshots use the
version-3 schemas and source-payload terminology.
