# Development

Install Go 1.26 or newer. Chrome or Chromium is needed for live crawling.
Node.js 22 or newer is needed for CSS generation and the browser-script checks. Serving a saved publication
and running the Go tests do not require Chromium.

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

On Windows, build `bin/tokens.exe` and invoke it with PowerShell. Use
`crawl --chrome-path PATH` when Chromium cannot be found automatically.
The [CLI reference](reference/cli.md) covers explicit paths and smaller captures.

For development with existing data, run `publish` and `serve` against a verified
snapshot. Rebuilding a publication does not need another crawl. Avoid modifying
snapshot files after creation; their checksums verify the captured evidence.

## Checks

```sh
go test ./...
npm ci
npm run test:ui
npm run test:tokens
```

To include captured-data checks, build assets for a local publication, then run:

```sh
npm run build:tokens -- --published published-review
TOKEN_ASSETS_DIR="$PWD/published-review/assets" npm run test:tokens
TOKEN_CATALOG_THEME_DB="$PWD/published-review/tokens.sqlite" go test ./catalog/web ./source/dtcg
```

Without these environment variables, optional captured-data checks skip.
The [handover](handover-components-and-exports.md) records validation boundaries.

Tests use local fixtures. Keep normal checks offline. Live captures and DNS or
Cloudflare checks are separate operational checks.

## Templates and assets

Edit the affected `catalog/web/*.templ` file, then regenerate the Go files with
the pinned templ version:

```sh
go run github.com/a-h/templ/cmd/templ@v0.3.960 generate -path catalog/web
```

Commit both the template and generated Go file. CSS and JavaScript are maintained
in `catalog/web/static/`. Shared component assets live in
`forge/components/static/`. Both are embedded in the binary. Rebuild and restart the
server after changing them. Styles come from the published token data through
the [DTCG and CSS build](reference/dtcg.md) when generated assets are present.
`catalog/web/theme.go` supports publications without generated assets. Layout
rules are in the static stylesheets.

The [component reference](reference/components.md) and [button reference](reference/buttons.md)
cover generated state bindings, shared interactions, and the local galleries.
The [CSS inspection and override instructions](reference/components.md#inspect-css-sources)
explain how to find manual declarations and change a token value.

## Code layout

| Path | Responsibility |
| --- | --- |
| `cmd/tokens/` | CLI commands and server startup |
| `source/crawl/` | Browser capture, source discovery, payload cache and receipts |
| `source/tokensnapshot/` | Snapshot assembly, verification and semantic comparison |
| `source/dtcg/` | Material conversion, context selection and DTCG exports |
| `forge/` | Style Dictionary CSS generation, Material component mappings, browser behavior and conformance checks |
| `source/publish/` | Token projection, retained history, SQLite and JSON exports |
| `source/update/` | Saved runs, review, activation, retry and notifications |
| `catalog/web/` | HTML, filters, downloads, API and agent interfaces |
| `catalog/usage/` | Request totals, retention and reports |
| `skills/` | Installable skill and embedded hosted guide |
| `schemas/` | Snapshot document schemas |
| `inputs/page-catalog.json` | Reviewed Material Design 3 source pages |
| `deploy/` | Docker, systemd and deployment scripts |

## Other design systems

The current source implementation expects Material Design 3 pages and payloads.
A further collection needs its own source catalog, capture and extraction rules,
publication mapping, and licensing notes. Some names, paths, token relationships
and theme generation also assume Material Design 3. Treat those as implementation
work when adding a source; a new hostname or URL prefix alone is insufficient.

Keep source provenance and completeness checks for any new collection. Its
values should be traceable to the source independently of the rendered site.

## Hover prefetch

Internal page links prefetch after 65 ms of mouse hover or keyboard focus.
Leaving the link before the delay cancels the pending work. Downloads, API
routes, external links and HTMX controls are excluded. Each document prefetches
at most eight distinct destinations, and data-saving or 2G connections skip it.
Hover and focus listeners run during capture so descendant handlers cannot hide
events by stopping propagation. Non-element event targets are ignored.

The script uses browser speculation rules where supported, then native document
prefetch hints. Other browsers keep normal navigation. Rendered responses use the revalidation policy described
below. The inner
prefetch handler can set a 30-second private lifetime, but the outer response
cache replaces it with `public, no-cache` for successful cached responses.

To check this in a browser, hover a component link briefly, then for more than
65 ms. In DevTools, check the prefetch request or the speculative loading panel.
Click the link and confirm the navigation reuses the prefetched response.

## Caching and navigation

The server loads one publication and its generated assets at startup. Rendered
HTML, HTMX fragments, Markdown and eligible API responses share an in-memory
LRU cache, bounded by 256 entries and 32 MiB of bodies and keys. Keys include
the URL/query, HTMX request/target, accepted content type and content encoding.
Boosted page navigation uses the full-document variant. Only successful
responses without cookies or `no-store` enter the cache. Health, usage, MCP,
range requests, static assets and ordinary downloads bypass it.

Cached responses use `public, no-cache` and representation-specific ETags.
Browsers revalidate; unchanged responses return 304. There is no age-based
server expiry: restarting after new data or a code build creates an empty cache.
Coverage reports and fixed thumbnail choices are also cached for that server.
Do not replace assets under a running server's pinned directory.

Versioned static URLs use content hashes and a one-year immutable lifetime.
Bare URLs and stale hashes revalidate. HTML, Markdown and static CSS/JS negotiate
gzip; compressed and uncompressed representations have separate ETags.
DTCG downloads revalidate and remain tied to the loaded publication.

HTMX boosts internal navigation and preserves `#tokens-site-header`, keeping the
navbar in place while page content changes. Native links remain usable without
JavaScript. Preview filters submit directly; their responses reuse the server
cache. The context popup uses shared controls and the surface color role. Its
request checks available context combinations without flashing a loading message;
errors remain visible.

## Rendering and motion

Component cards and token-table wrappers use `content-visibility: auto` with
intrinsic placeholder sizes. Browsers can skip offscreen layout and paint,
particularly on mobile. The DOM and styles are still delivered; this is not DOM
virtualization. Printing disables the containment optimization.

The landing page shows a small component/style selection, then a separate
section for change history and agent access. Hero artwork lives in
`catalog/web/static/artwork.js`. Its color-block choreography and loop timing
are authored; sampled captured spatial springs drive resize and shape
transitions. Stripe turns use the standard easing and medium2 duration tokens
to settle without overshoot. Corner changes share their block's resize clock and direction. Internal stripe
turns start on that clock too, with different tiles turning on different beats.
The pause control, reduced-motion preference, document visibility and viewport
intersection control playback.

Menus use shared Forge behavior. Browsers with overlay transitions retain the
native exit surface. Other browsers receive an inert, ARIA-hidden visual copy
that follows the spatial/effects spring curves, then is removed. FAB exits
mirror their item animations and delays. The copy never reopens a
popover. Reduced motion skips the fallback. Check real Firefox interactions as
well as the unit fallback checks before declaring browser conformance.

## Component motion

The site motion scheme is selected independently of public download contexts in
`forge/material/site-theme.json`. It currently uses Expressive. Rebuild tokens,
then rebuild and restart the server after changing the scheme or shared motion
code. The build derives CSS `linear()` curves and settling durations from the
captured damping/stiffness values; it does not approximate them with arbitrary
short/medium duration tokens. The shared source is
`forge/components/static/motion.cjs`, served as `motion.js` before component code.

The [motion contract and browser checks](reference/components.md#motion-schemes-and-lifecycle)
describe assignments, interruption handling, reduced motion and popup fallbacks.
Run `go run ./forge/conformance` alongside Catalog and open
`http://127.0.0.1:8082/motion.html` to repeat the checks.
