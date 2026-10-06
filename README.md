# Tokens

Tokens captures and explains the design-token data published by Material 3.
It discovers reviewed token-source pages, observes their lazy token-table
requests, stores source payloads by content hash, generates a loss-preserving
token graph, and publishes immutable material-token-data snapshots.

Tokens does not decide which values a component library should support.
Token Forge consumes a pinned snapshot, reviews its changes, and promotes accepted
values into canonical DTCG tokens.

## Requirements

- Go 1.26 or newer.
- Chrome or Chromium for pages that lazy-load token tables.

Download dependencies:

```text
go mod download
```

No Node.js, pnpm, Playwright, or `node_modules` directory is required.

## Build and run

```powershell
go build -o .\bin\tokens.exe .\cmd\tokens
.\bin\tokens.exe serve --published published --listen 127.0.0.1:8081
```

For public deployment, add `--public-origin https://tokens.marmenlind.com`.
See [publication](docs/publication.md) for creating the published data and
[search indexing](docs/web-seo.md) for deployment checks.

## Crawl

A normal crawl uses the checked-in page URL catalog. It lists reviewed Material
token-source pages, including component specification pages and foundation or
style token pages:

```text
go run ./cmd/tokens crawl
```

The page URL catalog is the default because it is small, deterministic, and
reviewable.
It avoids repeated sitemap discovery during routine capture.

Audit the live sitemap against the page URL catalog:

```text
go run ./cmd/tokens crawl --discover --dry-run
```

Run the broader inventory scope when reviewing pages outside the current
catalog:

```text
go run ./cmd/tokens crawl --discover --scope token-source-pages --dry-run
```

This scope considers component `/specs` pages and style or foundation token
pages ending in `/tokens`, `/specs`, `-tokens`, or `-scale`. It reports candidates and page-catalog drift;
it does not prove that every Material page publishes a token table.

Update the page URL catalog after reviewing the reported additions and removals:

```text
go run ./cmd/tokens crawl --update-page-catalog --dry-run=false
```

Run a small dry crawl during development:

```text
go run ./cmd/tokens crawl --max-pages 10 --page-timeout 8s --dry-run
```

By default, valid cached source payloads are reused without an HTTP download.
Force a fresh download when needed:

```text
go run ./cmd/tokens crawl --refresh
```

Show Chromium while debugging one page:

```text
go run ./cmd/tokens crawl --headed --max-pages 1 --page-timeout 15s
```

Use `--chrome-path` when Chrome cannot be found automatically.

Retry only the failed pages from the newest crawl receipt:

```text
go run ./cmd/tokens crawl --retry-failed
```

Use `--retry-failed-from <crawl-receipt>` to select a specific receipt. The
repair crawl opens only failed pages, then writes a new receipt containing the
previous successful pages plus the retried results.

To rerun snapshot extraction without opening Chrome or crawling pages again, use
`snapshot create`. It reads the selected crawl receipt and cached source-payload
artifacts. A normal `crawl` still re-observes pages so it can detect new, removed,
or changed source-payload URLs; valid payload bytes are reused unless `--refresh`
is supplied.

## Discovery

The default live discovery scope is `component-specs`. The
`token-source-pages` scope also considers style and foundation pages such as
`/styles/elevation/tokens`, `/styles/motion/overview/specs`,
`/styles/typography/type-scale-tokens`, and `/styles/shape/corner-radius-scale`.

The page URL catalog is the reviewed source of truth for normal crawls. Broad
discovery is an inventory aid. Review its additions before updating the page
catalog.

Live discovery reads `robots.txt`, follows sitemap indexes, filters same-origin
URLs, and falls back to same-origin link discovery only when no sitemap yields
usable pages. The seed page is included when it matches the configured scope.

The crawler does not need manually copied token-table URLs. For each selected
page it:

1. Opens the page through chromedp.
2. Observes matching `TOKEN_TABLE*.json` and `TOKEN_TYPE_*.json` network responses.
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
deliberately packages them. The hash identifies the bytes Tokens ingested;
the upstream URL records where they came from.

If upstream publishes a new URL, a later crawl stores a new artifact and leaves
the old content untouched. If upstream mutates an existing URL, use `--refresh`.
The new response receives a new content hash.

## Material token-data snapshots

Snapshots are immutable, content-addressed bundles of Material token-table
source payloads and generated token data. They contain a manifest, page URL catalog,
source payloads, generated token graph, source-payload details, page results, and
checksums. Lifecycle notices preserve explicit upstream deprecation and
recommendation statements without turning them into Token Forge support decisions.

Token Forge pins a specific manifest digest. It never consumes `latest`, runs this
crawler in its build, or depends on a Tokens service being online.

See the [data flow](docs/data-flow.md), [snapshot format](docs/snapshot-format.md), and the
[Tokens roadmap](docs/plans/2026-08-11-tokens-roadmap.md).

Create and verify a snapshot from the newest completed crawl:

```text
go run ./cmd/tokens snapshot create

go run ./cmd/tokens snapshot verify
```

`snapshot create` selects the newest crawl receipt, derives a UTC timestamped
snapshot ID, and writes to `snapshots/material-token-data/<snapshot-id>`. Use
`--help` on any command for explicit path overrides and validation rules.

Compare the two latest snapshot runs:

```text
go run ./cmd/tokens snapshot compare
```

The output directory must be empty. Snapshot files are immutable after
publication. The command prints the manifest SHA-256 used by Token Forge locks.
Historical snapshot IDs retain their original `m3-` prefix; new IDs should use
the Material source date and hour/minute.

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
override examples. See the [pre-compiler worklist](docs/plans/2026-08-11-tokens-precompiler-worklist.md)
for the remaining Tokens contract and Token Forge policy decisions.
