# Material Design 3 crawler

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

Pages are observed one at a time in a shared browser. `--page-timeout` is a
maximum wait, not a delay applied to every page. Capture finishes early once
the DOM is ready, a token payload has appeared, and relevant requests have
finished and stayed quiet for two seconds. Unrelated iframe traffic does not
extend the wait. Pages with no payload use the full timeout (30 seconds by
default). Capture checks still run before publication; network quiet alone
does not prove that every table was found.

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
deliberately packages them. The hash identifies the bytes Token Catalog ingested;
the upstream URL records where they came from.

If upstream publishes a new URL, a later crawl stores a new artifact and leaves
the old content untouched. If upstream mutates an existing URL, use `--refresh`.
The new response receives a new content hash.


See [snapshot format](snapshot-format.md) for immutable bundles and
[CLI reference](cli.md) for snapshot commands.
