# M3 Source snapshot format

**Status:** Proposed version 1  
**Date:** 2026-08-11

## Purpose

A snapshot is an immutable account of Material's published token data at one
point in time. It gives M3 Forge and other consumers a stable input without
requiring the crawler, a browser, a live API, or Material's website.

The snapshot preserves source meaning. It does not claim that every upstream
value is correct for a consumer or approved for a design-system release.

## Bundle layout

```text
m3-source-<snapshot-id>/
|-- manifest.json
|-- catalog.json
|-- material.graph.json.gz
|-- provenance.json
|-- coverage.json
`-- checksums.txt
```

An optional evidence bundle can contain raw content-addressed payloads. Keep it
separate so ordinary consumers do not download large browser captures.

## Snapshot identity

The publisher serializes release files deterministically, calculates their
SHA-256 hashes, then writes the manifest. The manifest digest is the integrity
identity used by consumers. A human-readable snapshot ID contains the source
family, capture date, and a digest prefix:

```text
m3-2026-08-11-a1b2c3d4e5f6
```

Never replace files under a published snapshot ID. Publish a new snapshot when
content, extraction rules, or source metadata changes.

## Manifest

Version 1 has this logical shape:

```json
{
  "schemaVersion": 1,
  "snapshotId": "m3-2026-08-11-a1b2c3d4e5f6",
  "createdAt": "2026-08-11T00:00:00Z",
  "source": {
    "system": "material-3",
    "catalogSha256": "<sha256>",
    "pageCount": 36
  },
  "capture": {
    "runId": "20260811T000000.000000000Z",
    "runManifest": "20260811T000000.000000000Z.json",
    "pagesAttempted": 36,
    "pagesSucceeded": 36,
    "payloadsObserved": 36,
    "payloadsFetched": 36,
    "complete": true
  },
  "extractor": {
    "name": "m3source",
    "version": "0.1.0",
    "schemaVersion": 1
  },
  "artifacts": [
    {
      "path": "material.graph.json.gz",
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

## Extracted graph

The graph is Material-specific and loss-preserving. It retains:

- canonical and display token names;
- upstream token, value, table, and revision IDs;
- token sets, states, contexts, and tags;
- literals and full reference trees;
- original types and units;
- unknown fields required to diagnose source-schema drift;
- source record IDs that join to provenance.

It is not DTCG source and does not flatten themes or contexts into one value.
Consumers may generate DTCG candidates, platform inputs, documentation, or
queries from it.

## Provenance

Provenance records page and payload URLs, payload hashes, retrieval metadata,
upstream revision details, and extraction notes. Use stable source IDs so graph
records can refer to provenance without duplicating long strings.

## Coverage

The coverage report lists each catalog page and whether it produced a token
payload, extracted records, warnings, or errors. A missing token table is visible
evidence, not a silently dropped page.

## Publishing and storage

Publish static files to content-addressable object storage, GitHub Releases, or
both. Static HTTP files are the stable contract. Documentation sites, APIs,
packages, and SQLite databases are derived views.

Git stores code, schemas, the curated catalog, fixtures, and small lock records.
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
