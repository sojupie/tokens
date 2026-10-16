# Source Loom / Token Forge boundary

Historical plan. See [current status](README.md) before following these instructions.


**Status:** Accepted for initial implementation  
**Date:** 2026-08-11

## Decision

Token Forge and Source Loom are separate repositories with separate release cycles.
Source Loom publishes immutable evidence snapshots. Token Forge pins one snapshot,
reviews its semantic changes, and promotes accepted values into canonical DTCG
tokens.

Token Forge must not depend on:

- the Source Loom executable;
- a live Source Loom API;
- Material's website during builds or tests;
- mutable URLs such as `latest`;
- raw browser-capture formats in component code.

Normal Forge builds use committed canonical token files and work offline.

## Data flow

```text
Material website
    -> Source Loom capture and extraction
    -> immutable source snapshot
    -> pinned Token Forge source lock
    -> candidate import and semantic review
    -> committed canonical DTCG tokens
    -> Forge compiler and packages
```

The import step is explicit. Updating the source lock alone does not change
canonical tokens or published components.

## Source lock

Token Forge records the reviewed upstream input in the existing
`sources/m3-source.lock.json` file when a snapshot is approved. The filename
retains the historical repository/module identity for now; it is not a second
source product name. The lock has this logical shape:

```json
{
  "schemaVersion": 1,
  "snapshotId": "m3-2026-08-11-<digest-prefix>",
  "manifestUrl": "https://example.invalid/source-loom/snapshots/<id>/manifest.json",
  "manifestSha256": "<64 lowercase hexadecimal characters>",
  "sourceSchemaVersion": 2,
  "reviewedAt": "2026-08-11T00:00:00Z"
}
```

The snapshot ID is a label. The manifest SHA-256 is the integrity check. Token Forge
stores the reviewed lock in Git so a release can identify its exact evidence
base.

Until a snapshot is approved, a lock may carry `status: "provisional"` and
`captureComplete: false`. Tooling must refuse to use that lock for canonical
promotion or release compilation.

## Version boundaries

These versions remain independent:

1. Source Loom tool version.
2. Source Loom snapshot schema version.
3. Immutable source snapshot ID and manifest digest.
4. Token Forge canonical token revision.
5. Token Forge compiler version.
6. Token Forge package version.

An upstream token change therefore does not force a Forge release. Forge can
reject it, defer it, reinterpret it for the web, or adopt it in a later release.

## Availability

Source Loom can publish snapshots through static object storage, GitHub Releases,
or both. Static files are the contract. A documentation site, query API,
package, or SQLite index can be added as a derived view. None of those services
becomes required infrastructure for Forge consumers.

## Failure policy

- A missing or unavailable upstream site can fail a live Source Loom audit. It
  cannot fail Token Forge CI.
- A source digest mismatch stops import before parsing.
- An unknown source schema stops automatic candidate generation and produces a
  diagnostic.
- A semantic change requires review. It is never promoted by a scheduled job.
