# M3 Source

M3 Source captures and explains the design-token data published by Material 3.
It discovers component specification pages, observes their lazy token-table
requests, stores payloads by content hash, extracts a loss-preserving graph, and
publishes immutable evidence snapshots.

M3 Source does not decide which values a component library should support. M3
Forge consumes a pinned snapshot, reviews its changes, and promotes accepted
values into canonical DTCG tokens.

## Requirements

- Go 1.26 or newer.
- Chrome or Chromium for pages that lazy-load token tables.

Download dependencies:

```text
go mod download
```

No Node.js, pnpm, Playwright, or `node_modules` directory is required.

## Crawl

A normal crawl uses the checked-in catalog of component `/specs` pages:

```text
go run ./cmd/m3source
```

The catalog is the default because it is small, deterministic, and reviewable.
It avoids repeated sitemap discovery during routine capture.

Audit the live sitemap against the catalog:

```text
go run ./cmd/m3source --discover --dry-run
```

Update the catalog after reviewing the reported additions and removals:

```text
go run ./cmd/m3source --update-catalog --dry-run=false
```

Run a small dry crawl during development:

```text
go run ./cmd/m3source --max-pages 10 --page-timeout 8s --dry-run
```

Refresh payloads even when their URL already exists in the cache:

```text
go run ./cmd/m3source --refresh
```

Show Chromium while debugging one page:

```text
go run ./cmd/m3source --headed --max-pages 1 --page-timeout 15s
```

Use `--chrome-path` when Chrome cannot be found automatically.

## Discovery

Material currently exposes token tables on
`/components/<component>/specs`. M3 Source therefore filters live discovery to
specification pages before launching Chromium.

Live discovery reads `robots.txt`, follows sitemap indexes, filters same-origin
URLs, and falls back to same-origin link discovery only when no sitemap yields
usable pages. The seed page is included when it matches the configured scope.

The crawler does not need manually copied token-table URLs. For each selected
page it:

1. Opens the page through chromedp.
2. Observes matching `TOKEN_TABLE*.json` network responses.
3. Fetches each unique payload once.
4. Hashes the uncompressed response with SHA-256.
5. Stores a gzip-compressed content-addressed artifact.
6. Writes a run manifest with pages, observations, URLs, hashes, and metadata.

## Local storage

```text
.cache/m3/
  index.json
  payloads/<hash-prefix>/<sha256>.json.gz
sources/runs/m3/<run-id>.json
```

The cache and ordinary run manifests are ignored by Git. They are local evidence
until a snapshot command deliberately packages them. The hash identifies the
bytes that M3 Source ingested; the upstream URL records where they came from.

If upstream publishes a new URL, a later crawl stores a new artifact and leaves
the old content untouched. If upstream mutates an existing URL, use `--refresh`.
The new response receives a new content hash.

## Snapshot contract

Snapshots are immutable, content-addressed release artifacts. They contain a
manifest, curated catalog, loss-preserving extracted graph, provenance, coverage
report, and checksums. Raw payloads can be distributed as a separate evidence
bundle when licensing and storage policy permit it.

M3 Forge pins a specific manifest digest. It never consumes `latest`, runs this
crawler in its build, or depends on an M3 Source service being online.

See [Snapshot format](docs/snapshot-format.md) and the
[M3 Source roadmap](docs/plans/2026-08-11-m3-source-roadmap.md).

Build and verify a snapshot from a completed run:

```text
go run ./cmd/m3source snapshot build \
  --run-manifest sources/runs/m3/<run-id>.json \
  --output sources/snapshots/m3-2026-08-11-example \
  --id m3-2026-08-11-example

go run ./cmd/m3source snapshot verify \
  --dir sources/snapshots/m3-2026-08-11-example
```

The output directory must be empty. Snapshot files are immutable after
publication. The command prints the manifest SHA-256 used by Forge locks.

## Current limits

- There is no generic loss-preserving extractor yet.
- There is no semantic snapshot diff or snapshot publisher yet.
- Run manifests contain volatile timestamps and are not yet deterministic
  release manifests.
- Live crawls require network access and can fail when Material changes its site.

Capture runs now fail when no page succeeds, record page and payload completeness
in the run manifest, and snapshot builds reject runs with no successful pages or
payloads. A failed or incomplete capture is evidence for diagnosis, not a
publishable snapshot.

Normal tests must use committed fixtures and remain offline.
