# Material Pain 3

Greenfield design-system research workspace.

The first experiment is a token-ingestion pipeline for Material Design 3. It
is intentionally separate from the future compiler and canonical token source:
upstream documentation is evidence, not something we silently compile into
production output.

## Token ingestion feasibility spike

The live Material button specification loads a public structured `TOKEN_TABLE`
JSON payload. The harvester discovers that payload URL from the page, fetches it
directly, and records the button token graph. This avoids manually operating the
lazy-loaded token tables and keeps the raw upstream references intact.

Each token record includes:

- token-set and upstream revision metadata;
- display name and canonical `md.*` token name;
- value kind (length, color, shape, typography, and so on);
- direct literal values when present;
- context-tagged reference trees, including referenced token/value IDs;
- source page, payload URL, and retrieval time.

The original single-page prototype is still available for comparison. It uses
Playwright and a manually selected page; it is not the automatic crawler:

```text
pnpm install
pnpm exec playwright install chromium
pnpm harvest:m3 -- --output sources/raw/m3-material/buttons.json
pnpm normalize:m3 -- sources/raw/m3-material/buttons.json --output sources/normalized/buttons.json
pnpm test
```

## Crawler

The default run uses a checked-in catalog of known component specification
pages. The crawler can also discover pages from live `robots.txt`/sitemaps,
fall back to same-origin link crawling, use Chrome DevTools Protocol via
`chromedp` to observe token-table requests, and deduplicate payloads into a
local content-addressed cache. It does not require Node dependencies or
Playwright.

Material currently exposes these token tables on component specification pages
(`/components/<component>/specs`), so specs-only discovery is enabled by
default. Other component pages are filtered out before Chromium is launched.

### Setup

Install Go 1.26 or newer, then download the Go modules once:

```text
go version
go mod download
```

The `pnpm` command is only a convenient script wrapper around `go run`; it does
not install or invoke JavaScript dependencies. A normal crawl uses the checked-in
catalog:

```text
pnpm crawl:m3
```

The equivalent direct command is:

```text
go run ./cmd/m3crawl
```

The catalog is intentionally the default so normal ingestion is deterministic.
Live discovery is an explicit audit/update operation.

### Crawl workflow

For each run, the crawler:

1. Loads the curated page catalog by default, or performs live discovery when
   `--discover`, `--update-catalog`, or `--no-catalog` is supplied.
2. When discovering live, fetches the configured sitemap or discovers one from
   `robots.txt`, recursively follows sitemap indexes, and filters to same-origin
   `/specs` pages.
3. Falls back to ordinary same-origin links if live sitemap discovery produces
   no usable pages.
4. Opens candidate pages in Chromium through chromedp.
5. Records matching `TOKEN_TABLE*.json` network responses.
6. Fetches each unique payload once and stores it by SHA-256 hash.
7. Writes a run manifest describing the pages, discovery mode, network
   observations, payload
   URLs, hashes, HTTP metadata, and token counts.

No token-table URL needs to be copied from DevTools manually.

### Common commands

Run a small catalog-backed dry run while developing:

```text
pnpm crawl:m3 -- --max-pages 10 --page-timeout 8s --dry-run
```

Audit the live sitemap against the catalog and crawl the live result:

```text
pnpm crawl:m3 -- --discover --dry-run
```

Update the catalog explicitly after reviewing the reported drift:

```text
pnpm crawl:m3 -- --update-catalog --dry-run
```

Run a fully automatic crawl without reading or comparing the catalog:

```text
pnpm crawl:m3 -- --no-catalog --max-pages 10 --dry-run
```

Run only a single live seed page and save the run elsewhere:

```text
pnpm crawl:m3 -- --no-catalog --seed https://m3.material.io/components/buttons/specs --max-pages 1 --run-dir .cache/runs/buttons
```

Limit discovery to another section of the site:

```text
pnpm crawl:m3 -- --path-prefix /components/buttons/ --max-pages 50
```

Inspect every page under the path prefix only when investigating a future site
change:

```text
pnpm crawl:m3 -- --specs-only=false --max-pages 50 --dry-run
```

Use a particular sitemap instead of `robots.txt` discovery:

```text
pnpm crawl:m3 -- --sitemap https://m3.material.io/sitemap.xml
```

Refresh payloads even when their URL is already present in the cache:

```text
pnpm crawl:m3 -- --refresh
```

Show Chromium while debugging a page that behaves differently when rendered:

```text
pnpm crawl:m3 -- --headed --max-pages 1 --page-timeout 15s
```

If Chromium is not found automatically, provide its executable path:

```text
pnpm crawl:m3 -- --chrome-path "C:\Program Files\Google\Chrome\Application\chrome.exe"
```

### Command-line options

| Option | Default | Purpose |
| --- | --- | --- |
| `--seed` | Material button specs page | Starting page and site origin |
| `--sitemap` | discovered from `robots.txt` | Sitemap or sitemap-index override |
| `--path-prefix` | `/components/` | Restrict pages to a same-origin path |
| `--specs-only` | `true` | Only open paths ending in `/specs` |
| `--catalog` | `sources/catalog/m3-spec-pages.json` | Curated page catalog used by default |
| `--discover` | off | Use live discovery and report catalog drift |
| `--update-catalog` | off | Use live discovery and replace the catalog (not with `--dry-run`) |
| `--no-catalog` | off | Use live discovery without comparing or updating the catalog |
| `--max-pages` | `500` | Maximum pages opened by Chromium |
| `--concurrency` | `2` | Number of browser pages observed in parallel |
| `--page-timeout` | `10s` | How long each page is allowed to load and emit requests |
| `--cache-dir` | `.cache/m3` | Content-addressed payload cache |
| `--run-dir` | `sources/runs/m3` | JSON run-manifest directory |
| `--refresh` | off | Ignore cached payload records and refetch |
| `--dry-run` | off | Discover and observe without downloading payload bodies |
| `--headed` | off | Run Chromium with a visible window |
| `--chrome-path` | automatic | Explicit Chrome/Chromium executable |
| `--all` | off | Legacy alias for `--no-catalog` |

`--page-timeout` controls observation time, not the HTTP payload download
timeout. Payload downloads use a separate 60-second HTTP timeout.

### Curated page catalog

The default catalog lives at:

```text
sources/catalog/m3-spec-pages.json
```

It is intentionally small and reviewable: one entry per known component specs
page, plus the sitemap source and catalog schema version. Normal crawls use it
without fetching the sitemap. Use `--discover --dry-run` to compare the live
sitemap with the catalog. The manifest records `added` and `removed` pages in
its `discovery` field. Use `--update-catalog` only after reviewing that drift;
the command writes the newly discovered set and then crawls it. `--no-catalog`
is for temporary investigations where the live result should not be compared or
persisted.

The crawler stores payloads in a local content-addressed cache:

```text
.cache/m3/
  index.json
  payloads/<hash-prefix>/<sha256>.json.gz
sources/runs/m3/<run-id>.json
```

`index.json` maps observed payload URLs to their cached artifact and retrieval
metadata. Payload files are gzip-compressed JSON named by their SHA-256 hash;
the hash is calculated from the uncompressed response bytes. Repeated runs can
therefore reuse the same artifact even when multiple pages reference it.

The cache and run directories are ignored by Git. Keep a run manifest when a
crawl needs to be reproducible; it records the exact page list, observed
revisioned payload URLs, response hashes, timestamps, and crawler options.

### Versioning and token updates

The repository versions the crawler, schema, tests, documentation, and curated
page catalog in Git. Generated upstream evidence is intentionally not committed
by default: raw payloads, normalized output, run manifests, and the
content-addressed cache are ignored. This keeps routine crawls from turning the
repository into a large binary-data archive. The page catalog is different: it
is small, text-based, and represents an intentional reviewable input to
ingestion.

When a new component appears, run `--discover --dry-run`, review the reported
drift, then run `--update-catalog` and commit the catalog change. Token payloads
remain content-addressed local evidence until a later compiler/review step
decides that a token snapshot should be curated and versioned.

There are two version identifiers for an observed payload:

1. The upstream token-table URL, which normally contains the upstream revision
   or release timestamp.
2. The SHA-256 hash of the exact uncompressed JSON response.

The hash is authoritative for the bytes we actually ingested. The crawler uses
the URL as the cache-index key and the hash as the artifact identity.

If upstream publishes a new revisioned URL, the next crawl records a new URL,
downloads a new hash-named artifact, and leaves the old artifact untouched. If
upstream mutates an existing URL, normal runs reuse the cached record for that
URL; use `--refresh` to fetch it again. A changed response receives a new hash
and artifact, while the cache index is updated to point that URL at the latest
record. Existing manifests retain the older URL/hash pair, so they remain useful
as historical evidence if they were preserved outside the ignored run folder.

The crawler does not yet perform a semantic token diff or automatically promote
an upstream update into canonical design-system tokens. That boundary is
intentional: ingestion records evidence first; review and normalization decide
whether a changed value is meaningful, deprecated, expressive-only, or suitable
for a future release.

### Interpreting a run manifest

The manifest contains these top-level fields:

- `pages`: the canonical page URLs selected by discovery;
- `observations`: page title/status and the token-table URL observed by
  Chromium, if any;
- `payloads`: downloaded or reused artifacts, their hashes, byte counts, HTTP
  metadata, and basic button token counts;
- `options`: the effective crawl configuration, including human-readable
  `pageTimeout`.

An observation with no `tokenTableUrl` is not necessarily a failure: some pages
may not load token data until a particular component state is rendered. The
manifest preserves those pages so the discovery result can be inspected rather
than silently discarded.

### Troubleshooting

- **No pages discovered:** verify `--seed`, `--path-prefix`, and network access.
  The seed is always included when it is a valid same-origin page.
- **Pages load but no token payload is observed:** increase
  `--page-timeout`, try `--headed`, and inspect the page in Chromium. The site
  may defer the request until client-side content finishes rendering.
- **Chromium cannot start:** install Chrome/Chromium or pass `--chrome-path`.
- **A stale payload is being reused:** pass `--refresh`, or remove only the
  relevant cache entry from `.cache/m3`.
- **The crawl is too expensive:** lower `--max-pages` and
  `--concurrency`; use `--dry-run` before downloading large payloads.

The upstream table intentionally contains references and context tags rather
than one flattened theme value. Normalization preserves that graph so light,
dark, expressive, and other upstream contexts are not accidentally collapsed
into a single value.

The generated files under `sources/raw` and `sources/normalized` are local
evidence/candidates. A future review step will reconcile them with AndroidX
token sources before anything becomes canonical.
