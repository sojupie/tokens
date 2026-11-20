# Material Design 3 crawler

## Running a crawl

Normal crawls use the checked-in, reviewed page catalog
(`inputs/page-catalog.json`):

```sh
go run ./cmd/tokens crawl
```

Common flags:

```sh
# Audit the live sitemap against inputs/page-catalog.json (JSON report, no browser)
go run ./cmd/tokens crawl --pages-only --scope token-source-pages

# Observe candidate pages in Chromium without writing payloads
go run ./cmd/tokens crawl --discover --scope token-source-pages --dry-run

# Write reviewed catalog updates and crawl
go run ./cmd/tokens crawl --update-page-catalog --dry-run=false

# Quick dev dry run or headed single-page debug
go run ./cmd/tokens crawl --max-pages 10 --page-timeout 8s --dry-run
go run ./cmd/tokens crawl --headed --max-pages 1 --page-timeout 15s

# Force re-downloading cached payloads
go run ./cmd/tokens crawl --refresh

# Retry only failed pages from the newest (or a specific) crawl receipt
go run ./cmd/tokens crawl --retry-failed
go run ./cmd/tokens crawl --retry-failed-from crawl-results/<receipt>.json
```

Pass `--chrome-path PATH` if Chromium isn't auto-detected.

### Page timing and caching
- Pages open sequentially in a shared Chromium instance. `--page-timeout` (30 s
  default) is a ceiling; a page finishes early once DOM is ready, a token
  payload appears, and relevant requests stay quiet for 2 s (unrelated iframe
  traffic is ignored).
- Valid cached payloads are reused without re-downloading unless `--refresh` is
  set.
- `--retry-failed` opens only previously failed pages and writes a merged
  receipt carrying forward prior successes.
- To rerun snapshot extraction without opening Chromium, run `tokens snapshot
  create` against an existing crawl receipt.

## Discovery and capture

Live discovery reads `robots.txt` and sitemaps (falling back to same-origin
links if no sitemap yields pages). `component-specs` is the default scope;
`token-source-pages` also includes style/foundation pages ending in `/tokens`,
`/specs`, `-tokens`, or `-scale` (plus `/styles/color/static/baseline`).

For each selected page, the crawler:
1. Opens the page via chromedp.
2. Watches network responses matching `TOKEN_TABLE*.json` and `TOKEN_TYPE_*.json`.
3. Fetches each unique payload once, hashes uncompressed bytes (SHA-256), and
   stores a gzipped content-addressed artifact.
4. Writes a crawl receipt recording pages, observed URLs, hashes, and errors.

## Local storage

```text
.cache/material-token-table-responses/
  index.json
  source-payloads/<hash-prefix>/<sha256>.json.gz
crawl-results/<crawl-id>.json
```

New upstream URLs create new cache artifacts without overwriting older bytes.
Use `--refresh` if upstream mutates an existing URL in place.

See [snapshot format](snapshot-format.md) and [CLI reference](cli.md).
