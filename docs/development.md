# Development

Requires Go 1.26+, Node.js 22.18+ (for CSS generation and UI checks), and Chrome
or Chromium (for live crawling and browser checks). Serving a saved publication
and running Go unit tests do not require Chromium.

For Go/templ component authoring and state ownership, see
[components API](reference/components-api.md).

## Build and capture

Run commands from the repository root:

```sh
go build -o ./bin/tokens ./cmd/tokens
./bin/tokens crawl
./bin/tokens snapshot create
./bin/tokens snapshot verify
./bin/tokens publish
./bin/tokens serve --published published --listen 127.0.0.1:8081
```

On Windows, build `bin/tokens.exe` and invoke it from PowerShell. Pass
`--chrome-path PATH` to `crawl` if Chromium isn't auto-detected. See the
[CLI reference](reference/cli.md) for explicit paths and smaller captures.

When iterating on existing data, run `publish` and `serve` against a verified
snapshot—rebuilding a publication does not need another crawl. Never edit
snapshot files after creation; their checksums verify the captured evidence.

### Local HTTP/3 with Caddy

To test HTTPS and HTTP/3 locally, run Caddy in front of the loopback server:

```sh
./bin/tokens serve --published published --listen 127.0.0.1:8086 \
  --public-origin https://localhost:8085/md3
caddy run --config deploy/Caddyfile.dev --adapter caddyfile
```

Trust the local Caddy CA in your test browser and open `https://localhost:8085/md3`,
or verify QUIC with an HTTP/3-capable `curl`:

```sh
curl --http3-only --cacert "$(caddy environ | sed -n 's/^caddy.AppDataDir=//p')/pki/authorities/local/root.crt" \
  -o /dev/null -w '%{http_version}\n' https://localhost:8085/md3/api/v1/health
```

To test from another device, set `TOKENS_DEV_HTTPS_SITE` (URL with port),
`TOKENS_DEV_HTTPS_BIND` (host interface), and a matching `--public-origin`.

## Checks

```sh
go test ./...
npm ci
npm run test:ui
npm run test:tokens
npm run check:components
npm run check:types
```

To verify rendered component behaviors against a running Catalog (requires Chromium):

```sh
go run ./cmd/check-behaviors --catalog http://127.0.0.1:8085
```

Pass `--path /components/search` to check a single component.

To include captured-data checks, build token assets for a local publication first:

```sh
npm run build:tokens -- --published published-review
TOKEN_ASSETS_DIR="$PWD/published-review/assets" npm run test:tokens
TOKEN_CATALOG_THEME_DB="$PWD/published-review/tokens.sqlite" go test ./catalog/web ./source/dtcg
```

Tests run offline against local fixtures.

## Templates, TypeScript, and component assets

- **Catalog templates:** Edit `catalog/web/templates/` and run
  `npm run generate:web` to update `catalog/web/internal/render/` and
  `catalog/web/assets/generated/`. Commit both sources and generated output;
  `npm run check:web` catches stale files. See [web ownership](../catalog/web/README.md).
- **Shared site shell:** `forge/site/` holds the navigation rail, drawer,
  appearance controls, and footer shared by Catalog and the browser tools.
  Export static shell markup and assets with:

  ```sh
  go run ./cmd/tokens site-shell --published published-review --out /tmp/site-shell
  ```

- **Component CSS and behaviors:** Edit CSS in `forge/components/styles/` and
  behaviors in `forge/components/behaviors/`, then run `npm run build:components`
  to assemble `forge/components/generated/` and compile TypeScript entries in
  `catalog/web/assets/scripts/`. Run `npm run check:components` to check types
  and generated output.
- **Token-bound CSS:** Run `npm run build:tokens -- --published <dir>` to
  generate themes and component styles from a publication. See
  [DTCG exports and CSS](reference/dtcg.md).
- **TypeScript:** Node 22.18+ runs build scripts by stripping types
  (`erasableSyntaxOnly` is enforced); esbuild compiles browser `.ts` files to
  classic JavaScript with no extra browser runtime.
- **Release CSS names:** Production builds (`deploy/build-assets.sh`) pass
  `--css-names short` by default; set `TOKENS_CSS_NAMES=original` to keep
  original names. Local `build:tokens` defaults to original names and accepts
  `--css-names short`.

Both Forge and Catalog served assets are embedded in the Go binary; rebuild and
restart the server after changing them.

For backend caching, HTMX navigation, module/theme splitting, hover prefetch,
and browser benchmarks, see
[delivery and performance](reference/delivery-performance.md). For layout grids
and scene framing, see [layout](reference/layout.md).

## Component illustrations

Catalog index cards display generated thumbnails of simple previews while
playgrounds stay interactive. Generate thumbnails after building token assets
(requires `cwebp` and Chromium):

```sh
go run ./cmd/tokens illustrations --published published-review --chrome-path /usr/bin/chromium
```

This captures 720 × 450 and 360px JPEGs plus a 40 × 25 inline WebP placeholder
across all six appearances and writes `illustrations/manifest.json`. See
[preview scenes](reference/preview-scenes.md) for framing rules.

## Conformance server and benchmarks

Start the local conformance proxy alongside Catalog (defaults to
`http://127.0.0.1:8081/md3`; pass `--catalog URL` to override):

```sh
go run ./forge/conformance
```

Key fixtures on `http://127.0.0.1:8082/`:

- `/motion.html` — motion contracts, direct loads, HTMX navigation, and revisits
- `/component-modules.html` — module CSS parity against the complete stylesheet across all six appearances (`?path=/md3/components/search&width=390` to scope)
- `/htmx-catalog.html` — region swaps, context cancellation, download selections, and history
- `/hero.html` — hero artwork DOM/Canvas parity and timing (pass `--profile-pids` for Linux renderer/GPU CPU counters)
- `/carousel.html` — carousel variants, RTL, endpoints, and interruption
- `/list-reorder.html` — list reorder indicator placement
- `/theme-aliases.html`, `/theme-equivalence.html`, `/theme-scoping.html`, `/asset-equivalence.html` — theme/CSS parity and switching checks against temporary baselines in `forge/conformance/fixtures/`

Benchmark CLIs (require a running Catalog and Chromium):

```sh
go run ./cmd/bench-navigation --catalog http://127.0.0.1:8085 --chrome-path /usr/bin/chromium --runs 15 --output /tmp/catalog-benchmark
go run ./cmd/bench-css-names --catalog http://127.0.0.1:8085 --chrome-path /usr/bin/chromium --runs 15 --output /tmp/css-names
```

See [delivery and performance](reference/delivery-performance.md) for benchmark
methodology and flags.

## Progressive enhancement (no-JS baseline)

Site navigation, token tables, filters, and downloads must work without
JavaScript:

- Navigation uses a native disclosure fallback; playgrounds and filters submit
  ordinary GET forms (keep named `<select>` controls—don't add duplicate hidden
  inputs).
- Forge provides native fallbacks for range, date/time, filter-chip,
  list-selection, `<progress>`, and invoker-command dialog triggers.
- Spring/morph motion, draggable sheets, custom calendars/dials, live search
  suggestions, and swipe/reorder gestures are JavaScript enhancements.

## Code layout

| Path | Responsibility |
| --- | --- |
| `cmd/tokens/` | CLI commands and server startup |
| `source/crawl/` | Browser capture, discovery, payload cache, and receipts |
| `source/tokensnapshot/` | Snapshot assembly, verification, and comparison |
| `source/dtcg/` | Material conversion, context selection, and DTCG exports |
| `source/publish/` | Token projection, retained history, SQLite, and JSON exports |
| `source/update/` | Scheduled runs, review, activation, retry, and notifications |
| `forge/` | Style Dictionary CSS generation, Material mappings, browser behaviors, and conformance |
| `catalog/web/` | HTML pages, filters, downloads, API, and agent interfaces |
| `catalog/usage/` | Request totals, retention, and weekly reports |
| `skills/` | Installable skill and embedded hosted guide |
| `schemas/` | Snapshot JSON schemas |
| `inputs/page-catalog.json` | Reviewed Material Design 3 source pages |
| `deploy/` | Docker, systemd, and deployment scripts |

## Adding another design system

The current capture, extraction, publication mapping, and site theme assume
Material Design 3. Adding another collection requires its own source catalog,
capture/extraction rules, publication mapping, theme, and licensing notes while
preserving source provenance and completeness checks. See
[collection boundaries](architecture.md#collection-boundaries).
