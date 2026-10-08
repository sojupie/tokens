# Crawler history

This note records how crawling changed from the first commit to the current implementation. It is based on the repository history and source code, not on assumptions about earlier uncommitted experiments.

## The first commit: two implementations

The initial commit, `d538ae5` (10 August 2026), included both Go and JavaScript crawler paths. The JavaScript tools used Playwright to launch Chromium; the Go crawler used chromedp. The JavaScript path was removed in `7e6a821` the next day, when the first complete source snapshot was added. So the Go crawler was not a later rewrite of a JavaScript-only project: both existed at the start, and Go became the maintained path almost immediately.

### JavaScript / Playwright path

The scripts lived in `tools/m3-site-harvester/` and had two related jobs:

- `crawl.mjs` found pages from `robots.txt`/sitemaps, then followed same-origin links under `/components/` when needed. It opened pages in parallel (default concurrency 2), listened for network responses whose URL matched `TOKEN_TABLE...json`, and recorded each page's status, title and discovered payload URL. It fetched each unique JSON payload outside the browser, gzipped it into a content-addressed cache, and wrote a run manifest. Cached payloads could be reused unless `--refresh` was set.
- `harvest.mjs` handled a single component page (Buttons by default). It opened that page, waited for a matching `TOKEN_TABLE` response, fetched the JSON, and converted it to a narrower snapshot. It kept active button token sets, omitted deprecated sets, and extracted direct values plus contextual reference trees.
- `normalize.mjs` flattened the harvested snapshot into a simpler token list.

The browser's role was mainly to reveal the data endpoint loaded by the site. The scripts then fetched the JSON directly. This path had a small, targeted output and basic cache/run-manifest support, but no full-capture completeness gate or resumable review workflow.

### Go path at the start

The same initial commit had `cmd/m3crawl` and `internal/m3crawl`. It already used Chromium through chromedp to discover the JSON endpoint from page network responses, allowed parallel page observation, and supported a page catalog, discovery, timeout, cache, refresh and dry-run options. The first JavaScript scripts were therefore an alternate/earlier workflow, not the only crawler.

## How the Go crawler grew

- **11–12 August — capture became a snapshot pipeline.** Commits `7e6a821`, `73e8f56` and `682893a` replaced the JavaScript scripts and expanded the Go capture/reporting flow. The project added versioned snapshots, a curated page catalog, source payload details, coverage reports, checksums and captured data. Crawling was no longer just “find a JSON URL”; the captured material could be checked and turned into a reproducible snapshot.
- **15 August — validation and provenance.** `d1b92f3` hardened snapshot and capture validation. The capture began to distinguish what pages and source payloads were observed and to validate the resulting snapshot rather than treating any successful HTTP response as sufficient.
- **September — publication separated from crawling.** The crawler's outputs became inputs to a separate offline publication pipeline: verified snapshots are transformed into a read-only site and token downloads. This keeps serving independent of Chromium and live upstream availability.
- **October — durable, reviewable updates.** The current `internal/crawl` flow uses a page catalog by default, can compare/update it through discovery, and records page observations and payload provenance in crawl receipts. It captures every matching source payload seen on a page (not just one), deduplicates payload URLs, downloads the source JSON, validates cached compressed artifacts, hashes the raw bytes, and preserves results for inspection. Failed pages can be retried from a prior receipt while successful observations are carried forward.
- **October — capture completion and operations.** `503bf65` changed page completion from waiting out a fixed timeout to waiting for a source payload, relevant requests to settle, DOM readiness and a short network-quiet interval, with the timeout retained as a fallback. The update-run workflow added persistent run state, candidate review/approval, targeted retries, locking and failure/review notifications (`a8ccca6`). Later work reduced peak memory during snapshot verification and added SMTP email alerts (`855c183`).

## Current flow

1. Select component-spec pages from the checked-in page catalog, or explicitly run discovery to find catalog drift.
2. Open those pages in a shared headless Chromium instance, with configurable parallel page contexts.
3. Observe the page's main-frame network activity and collect all matching source-payload URLs. A page is considered settled after payload activity and relevant requests quiet down; a per-page timeout still bounds the wait.
4. Fetch each unique payload, validate/reuse the content-addressed gzip cache when possible, and write a receipt with page status, payload provenance, hashes, sizes and failures.
5. Verify the captured set and build a candidate snapshot/publication. Incomplete or suspicious results are not silently activated: an administrator can inspect, retry failed pages, and approve a candidate.
6. Switch the served publication only after approval; the public server itself does not run the crawler.

## Main differences

| Concern | Initial JavaScript path | Current Go path |
| --- | --- | --- |
| Browser | Playwright Chromium | chromedp / Chromium DevTools Protocol |
| Scope | Single-component harvest or broad page crawl; component harvester focused on Buttons | Catalog-driven coverage across component token-source pages |
| Payload capture | Detected `TOKEN_TABLE` response URLs, then fetched JSON | Captures multiple matching source payloads per page, with source context and deduplication |
| Saved output | Narrow transformed snapshot plus basic gzip cache and run manifest | Raw payload cache, detailed receipt, validated snapshot and publication generations |
| Failure handling | Per-page/payload error in a run manifest | Completeness checks, retry failed pages, carry forward successful observations, review/approval and notifications |
| Runtime role | Scraper and normalizer scripts in Node plus Playwright | One Go CLI and Chromium; crawl, verification, publication and serving are separate stages |

The current design is more than a language change. It treats crawl coverage, source provenance, retries and approval as part of the data pipeline, while keeping the website's serving path read-only and independent of live crawling.
