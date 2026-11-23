# Crawler history

How TokenCatalog's crawler evolved from the initial commit.

## Initial commit (10 August 2026)

Commit `d538ae5` (Aug 10, 2026) included two parallel crawlers:

- **JavaScript / Playwright (`tools/m3-site-harvester/`):** `crawl.mjs` discovered pages from `robots.txt` and sitemaps, listened for `TOKEN_TABLE...json` network responses, and gzipped fetched payloads into a content-addressed cache. `harvest.mjs` and `normalize.mjs` extracted single-component snapshots (defaulting to Buttons).
- **Go / chromedp (`cmd/m3crawl`, `internal/m3crawl`):** Observed network traffic in headless Chromium using a page catalog, with discovery, caching, refresh, and dry-run flags.

The JavaScript scripts were removed the next day in `7e6a821` (Aug 11, 2026) when the first complete Go snapshot landed.

## Milestones

- **11–12 August 2026 (`7e6a821` and `73e8f56`, Aug 11; `682893a`, Aug 12):** Added versioned snapshots, a curated page catalog, payload details, coverage reports, and SHA-256 checksums.
- **15 August 2026 (`d1b92f3`):** Hardened snapshot and capture validation to track page observations and payload results separately.
- **September 2026:** Split offline publication from crawling so the web server runs read-only without Chromium.
- **7 October 2026 (`503bf65`, `a8ccca6`, `855c183`):** Switched page completion from fixed timeouts to payload + network-quiet detection, added multi-payload capture and failed-page retries, introduced persistent update runs with review/approval and SMTP alerts, and reduced peak verification memory.

For current crawler behavior, see [Material Design 3 crawler](../reference/crawler.md).
