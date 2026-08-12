# Source Loom snapshot format

**Status:** Proposed version 2
**Date:** 2026-08-11

## Purpose

A snapshot is an immutable bundle of Material's fetched token-table source
payloads and the generated token data extracted from them at one point in time.
It gives Token Forge and other consumers a stable input without requiring the
crawl tool, a browser, a live API, or Material's website.

The snapshot preserves source meaning. It does not claim that every upstream
value is correct for a consumer or approved for a design-system release.

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
identity used by consumers. A human-readable snapshot ID contains the source
family, capture date, capture time in UTC (`HHMM`), and a digest prefix:

```text
material-2026-08-11-2348-a1b2c3d4e5f6
```

Never replace files under a published snapshot ID. Publish a new snapshot when
content, extraction rules, or source metadata changes.

## Manifest

Version 2 has this logical shape:

```json
{
  "type": "material-token-data-snapshot",
  "schemaVersion": 2,
  "snapshotId": "material-2026-08-11-a1b2c3d4e5f6",
  "createdAt": "2026-08-11T00:00:00Z",
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
    "name": "sourceloom",
    "version": "0.1.0",
    "schemaVersion": 2
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

It is not DTCG source and does not flatten themes or contexts into one value.
Consumers may generate DTCG candidates, platform inputs, documentation, or
queries from it.

## Source-payload details

Source-payload details record page and source-payload URLs, source-payload
hashes, retrieval metadata, upstream revision details, and extraction notes. Use
stable source IDs so generated records can refer to source-payload details
without duplicating long strings.

## Page results

The page-results file lists each page and whether it produced source payloads,
generated records, warnings, or errors. A missing token table is visible
evidence, not a silently dropped page.

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
source record identifiers are retained so Token Forge can make a support
decision without treating Source Loom's evidence as policy.

## Publishing and storage

Publish static files to content-addressable object storage, GitHub Releases, or
both. Static HTTP files are the stable contract. Documentation sites, APIs,
packages, and SQLite databases are derived views.

Git stores code, schemas, the curated page URL catalog, fixtures, and small lock records.
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

`snapshot compare` verifies both input directories, then compares their page
page URL catalogs, page results, capture completeness, content-addressed source payloads, and
generic extracted records and lifecycle notices. Record changes are keyed by source payload, record
kind, upstream record ID, and occurrence ordinal; the complete upstream object
is compared through the record's loss-preserving `raw` field. It ignores
snapshot IDs, crawl IDs,
retrieval timestamps, and artifact paths when deciding whether capture metadata
changed.
