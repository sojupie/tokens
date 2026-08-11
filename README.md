# M3 Source

M3 Source captures and explains the design-token data published by Material 3.
It discovers component specification pages, observes their lazy token-table
requests, stores source payloads by content hash, generates a loss-preserving
token graph, and publishes immutable material-token-data snapshots.

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

A normal crawl uses the checked-in page URL catalog. It lists Material
component specification pages under `/components/<component>/specs`:

```text
go run ./cmd/m3source crawl
```

The page URL catalog is the default because it is small, deterministic, and
reviewable.
It avoids repeated sitemap discovery during routine capture.

Audit the live sitemap against the page URL catalog:

```text
go run ./cmd/m3source crawl --discover --dry-run
```

Update the page URL catalog after reviewing the reported additions and removals:

```text
go run ./cmd/m3source crawl --update-page-catalog --dry-run=false
```

Run a small dry crawl during development:

```text
go run ./cmd/m3source crawl --max-pages 10 --page-timeout 8s --dry-run
```

Refresh payloads even when their URL already exists in the cache:

```text
go run ./cmd/m3source crawl --refresh
```

Show Chromium while debugging one page:

```text
go run ./cmd/m3source crawl --headed --max-pages 1 --page-timeout 15s
```

Use `--chrome-path` when Chrome cannot be found automatically.

Retry only the failed pages from the newest crawl receipt:

```text
go run ./cmd/m3source crawl --retry-failed
```

Use `--retry-failed-from <crawl-receipt>` to select a specific receipt. The
repair crawl opens only failed pages, then writes a new receipt containing the
previous successful pages plus the retried results.

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
6. Writes a crawl receipt with pages, observations, URLs, hashes, and metadata.

## Local storage

```text
.cache/material-token-table-responses/
  index.json
  source-payloads/<hash-prefix>/<sha256>.json.gz
crawl-results/<crawl-id>.json
```

The cache contains downloaded Material token-table responses. Crawl receipts
record what one crawl job did. Both are local evidence until a snapshot command
deliberately packages them. The hash identifies the bytes M3 Source ingested;
the upstream URL records where they came from.

If upstream publishes a new URL, a later crawl stores a new artifact and leaves
the old content untouched. If upstream mutates an existing URL, use `--refresh`.
The new response receives a new content hash.

## Material token-data snapshots

Snapshots are immutable, content-addressed bundles of Material token-table
source payloads and generated token data. They contain a manifest, page URL catalog,
source payloads, generated token graph, source-payload details, page results, and
checksums.

M3 Forge pins a specific manifest digest. It never consumes `latest`, runs this
crawler in its build, or depends on an M3 Source service being online.

See the [data flow](docs/data-flow.md), [snapshot format](docs/snapshot-format.md), and the
[M3 Source roadmap](docs/plans/2026-08-11-m3-source-roadmap.md).

Create and verify a snapshot from the newest completed crawl:

```text
go run ./cmd/m3source snapshot create

go run ./cmd/m3source snapshot verify
```

`snapshot create` selects the newest crawl receipt, derives a UTC timestamped
snapshot ID, and writes to `snapshots/material-token-data/<snapshot-id>`. Use
`--help` on any command for explicit path overrides and validation rules.

Compare the two latest snapshot runs:

```text
go run ./cmd/m3source snapshot compare
```

The output directory must be empty. Snapshot files are immutable after
publication. The command prints the manifest SHA-256 used by Forge locks. New
snapshot IDs should include the UTC capture date and hour/minute, for example
`m3-2026-08-11-2348-curated-specs`.

## Current limits

- Snapshot records preserve the source payload objects while exposing generic
  indexes for token, value, component, and reference-tree records.
- Snapshot comparison reports source-payload and generated-record additions, removals, and
  changes; it does not publish snapshots.
- Crawl receipts contain volatile timestamps and are not deterministic
  release manifests.
- Live crawls require network access and can fail when Material changes its site.

Crawl jobs now fail when no page succeeds, record page and source-payload
completeness in the crawl receipt, and snapshot creation rejects crawls with no
successful pages or source payloads. A failed or incomplete crawl is evidence
for diagnosis, not an approved snapshot.

Normal tests must use committed fixtures and remain offline.

See the [CLI guide](docs/cli.md) for command help, defaults, and explicit
override examples.
