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
npm run check:components
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

Page templates have named files such as `layout.templ` and `changes.templ`.
Component examples use `specimen_<family>.templ`; `component_specimen.templ` only
dispatches to the selected family. Matching `specimen_*.go` files hold attribute
helpers, while `component_gallery.go` holds the shared preview case model.

Edit the affected `catalog/web/*.templ` file, then regenerate the Go files with
the pinned templ version:

```sh
go run github.com/a-h/templ/cmd/templ@v0.3.960 generate -path .
```

Commit both the template and generated Go file. Site CSS and JavaScript are maintained
in `catalog/web/static/`. Shared component CSS sources live in
`forge/components/styles/`; behavior sources live in `forge/components/behaviors/`.
Run `npm run build:components` to assemble them into the existing served files in
`forge/components/static/`. Dedicated scripts such as `menus.js` and `fields.js`
remain maintained in `static/`. Catalog's demo responses live in
`catalog/web/static/component-demos.js`. Both sets of served assets are embedded
in the binary. Rebuild and restart the
server after changing them. Styles come from the published token data through
the [DTCG and CSS build](reference/dtcg.md) when generated assets are present.
`catalog/web/theme.go` supports publications without generated assets. Layout
rules are in the static stylesheets.

The [component reference](reference/components.md) and [button reference](reference/buttons.md)
cover generated state bindings, shared interactions, and the local galleries.
The [CSS inspection and override instructions](reference/components.md#inspect-css-sources)
explain how to find manual declarations and change a token value.

## Asset delivery checks

`forge/css-output.mjs` factors the site theme and compacts generated CSS.
`forge/component-bundles.mjs` selects shared controls and optional component
families. Add child families there when a specimen composes other components.
`catalog/web/component_assets.go` selects the stylesheet for each page.
`catalog/web/static/asset-loader.js` selects optional behavior from the rendered
markup. Keep these mappings aligned when adding a component.

The layout loads shared CSS and core controls. Page-specific stylesheets live
inside the replaced page content, so HTMX removes them when leaving a playground.
The loader fetches each optional script once and emits `tokens:assets-ready` after
styles and behavior are ready. Initializers must handle that event and remain
idempotent on revisits. A playground stays inert while its assets load. A failed
load shows a reload instruction. Inert catalog thumbnails skip optional behavior,
except the progress script needed to paint static SVG paths.

Run the conformance proxy described under motion checks below. `/motion.html`
checks direct loads, HTMX navigation, lazy initialization and revisits.
`/asset-performance.html` measures the actual theme and contrast controls, reporting
synchronous style/layout time and time through two animation frames. Close other
animation fixtures before measuring. Compare the same browser, viewport and
publication; these timings are machine-specific.

A Chromium 155 measurement on Linux on 2026-10-22 reduced `/changes` CSS from
3.00 MB to 1.24 MB before compression (220 KB to 101 KB with gzip level 6).
Script requests fell from 22 to 12. In 24 measured appearance changes after four
warmups, synchronous style/layout median fell from 49.5 ms to 42.6 ms on
`/changes`, and from 51.0 ms to 42.2 ms in the navigation-rail playground.
Time through two animation frames did not improve in that first pass.
These figures describe this local browser run, not a cross-browser performance guarantee.

`forge/theme-subsets.mjs` now derives theme inputs from shared CSS, template
references, family bundles and literal JavaScript reads. It follows CSS aliases
in all appearances; Style Dictionary resolves DTCG aliases before this step.
Dynamic spring property names have explicit dependency roots in `pageThemes`.
Add a root there if behavior introduces another dynamically constructed token
name. Unknown Material or Catalog references fail the build.

The shared theme stays in the head. A playground's remaining inputs are scoped
to `#component-preview`; its theme link follows the page stylesheet through HTMX
navigation. Catalog overview pages use the optional `all` theme globally because
their thumbnails sit outside the playground. Existing manifests without theme
bundles keep the complete theme. `/theme.css` and public token downloads remain
complete.

For a theme comparison, copy the complete generated `theme.css` to
`forge/conformance/baseline-theme.tmp`. Open `/theme-equivalence.html` to compare
layout, typography, colors and surfaces across the gallery in all six appearances.
`/theme-scoping.html` alternates the complete and page themes in the same document,
with warmups in each block. It reports time in the control handlers (including any
synchronous style work), the following forced style/layout flush, and two animation
frames. The last measure includes scheduling and rendering; it does not isolate
paint duration. Close other animation fixtures while measuring.

In a same-document A/B run on Chromium 155/Linux on 2026-10-22, with 32 measured
samples per mode after block warmups, the median through two animation frames
fell from 68.3 ms to 51.1 ms on `/changes` and from 68.4 ms to 51.2 ms on the rail.
The forced style/layout flush medians fell from 17.3 ms to 10.2 ms and from
15.7 ms to 10.2 ms respectively. Frame cadence differs between runs; use the
within-run comparison. The root set fell from 4,115 to 2,017 variables, with 153
additional variables scoped to the rail preview. Overview pages still load the
full set for their mixed component thumbnails.

For a CSS build comparison, save the previous theme and full component stylesheet
as `forge/conformance/baseline-theme.tmp` and `baseline-components.tmp`, then open
`/asset-equivalence.html`. It compares retained root variables in all six contexts and
computed component properties across the gallery. These temporary baseline files
are ignored by Git. The comparison complements interaction checks; it does not
replace them.

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
Bare URLs and stale hashes revalidate. Static gzip bodies are cached by content hash within the server process.
HTML, Markdown and static CSS/JS negotiate
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
to settle without overshoot. Grid resizing, directional corner changes and stripe
turns share one scene clock, with a transition every two seconds and no column
delays. Different tiles turn on different beats; held scenes preserve variety.
The pause control, reduced-motion preference, document visibility and viewport
intersection control playback.

The mosaic has an inset inside the frame; complete blocks stay visible throughout
the loop. The default renderer uses Canvas 2D in `catalog/web/static/artwork-canvas.js`.
Add `?artwork=dom` to the homepage to compare the DOM/CSS renderer.
Both use the same tracks and spring samples from `artwork.js`. Canvas resolves
palette colors through CSS on appearance changes, resizes its backing buffer with
the viewport, and caps bitmap density at 2x. It sleeps during settled holds and
cancels its frame/timer callbacks while paused, hidden or offscreen. The HTML tiles
remain as a fallback if a 2D context cannot be created. Only the Canvas variant
loads the additional renderer script.

Carousel touch scrolling uses native momentum and snapping. The sticky visual track isolates descendant layout and carries a transform
compositing hint. Signed scroll offsets are clamped at both ends so Safari rubber-banding does not
fold into item movement. Its masked item geometry
updates directly from scroll events, without another queued animation frame. Button,
keyboard and mouse-release settling use fractional transforms sampled from a Web
Animations clock with the standard easing and medium2 duration tokens. The native
scroll position is synchronized at completion or interruption, rather than on each
visual frame. Destinations come from the laid-out snap elements and are resolved
through the browser’s scroll precision before animating. Artwork uses explicit
fractional transforms instead of percentage centering inside a changing mask.
Unchanged scroll positions skip geometry writes. Mouse release
projects recent drag velocity to choose the destination. Pointer and wheel input
interrupt settling; reduced motion skips it. `/carousel.html` checks all four variants
at desktop and phone widths, RTL, endpoints, focus and interruption behavior.

Open `/hero.html` through the conformance proxy to check full-loop frame fit,
DOM/Canvas geometry and corner parity, six appearances, responsive sizing and
playback controls. Its profile alternates DOM/Canvas/Canvas/DOM on the real
homepage, using two complete active loops per sample and separate paused and
offscreen samples. It reports frame intervals, long animation frames and Canvas
draw-command submission time. Submission time does not include completed GPU
rasterization.

Optional Linux CPU counters come from explicitly selected browser renderer and GPU
processes. Start the conformance proxy with `--profile-pids` followed by their
comma-separated numeric PIDs. `/profile/cpu` reads their cumulative process CPU
counters; the fixture computes deltas for each sample. Percentages use one core
as 100 percent, and GPU-process CPU time is not GPU-device utilization. Pause other
animated pages and avoid browser interaction during sampling. Without configured
PIDs, the profile still reports browser timing metrics. These checks run locally;
the profiling endpoint is not included in Catalog.

A Chromium 155/Linux run on 2026-10-22 used a 381 × 416 CSS-pixel hero at DPR 1,
with DOM/Canvas/Canvas/DOM order. Each active sample lasted 16 seconds; paused and
offscreen samples lasted eight seconds. The hero iframe's renderer process used
16.1–19.1 percent of one core for DOM and 2.81 percent for Canvas. Idle samples
were approximately 1.6–3.0 percent, including the timing harness. GPU-process CPU
was 23.2–23.9 percent for DOM and 20.0–20.5 percent for Canvas, against an idle
baseline around 15.1–15.6 percent. This shared GPU-process counter includes the
browser application and is not a measure of GPU hardware load.

Both renderers had a frame-interval p95 around 20.9 ms and no animation frames over
50 ms. Canvas draw-command submission p95 was 0.3 ms, with no draws during paused
or offscreen samples. These results support lower renderer CPU cost for this scene;
they do not establish battery savings or results on other browsers, sizes or DPRs.



Menus use shared Forge behavior. Browsers with overlay transitions retain the
native exit surface. Other browsers receive an inert, ARIA-hidden visual copy
that follows the spatial/effects spring curves, then is removed. FAB exits
mirror their item animations and delays. The copy never reopens a
popover. Touch select options commit on pointer release and defer focusout dismissal
until that release; canceled gestures do not change the value. Compatibility
clicks do not repeat the change event. Reduced motion skips the fallback. Check
real Firefox interactions as well as the unit fallback checks before declaring
browser conformance.

Rail transitions measure label heights after changing variants and before
capturing animation destinations. The resize observer maintains those heights
afterward; it must not change the base indicator transform during a transition.
All destinations are measured before starting parent animations. Icons and
indicators move with positional offsets rather than animated transforms. Labels
crossfade on the same default spatial progress as the menu icon rotation and
rail geometry. Following AndroidX NavigationItem, outgoing opacity is
`max(0, 1 - 2p)²` and incoming opacity is `max(0, min(1, 2p - 1))²`.
Both destination-label layouts are invisible at `p = 0.5`, when the menu icon
has rotated 45°. Section headings remain at their final size and position,
using only the default effects opacity fade.
Inert outgoing text fades in place. Incoming collapsed labels also stay fixed;
expanded labels approach horizontally from near the icon, covering most of
that distance while transparent. Positional offsets cancel the moving parent
indicator. Outgoing text is removed on completion or interruption. Reduced
motion changes layouts immediately. Rail FAB label styles exclude the ripple
span, which stays absolutely positioned and visible in both rail states. FAB
press feedback retains the source shape and uses its ripple/state layer. The directory
keeps its final layout width during motion, and its position follows the same
spatial spring as the rail. The modal scrim retains zero opacity after its
faster effects spring settles, until spatial cleanup hides it. Horizontal rail
labels use the indicator’s on-container icon foreground. Vertical labels remain
outside the indicator and retain the captured label color.

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
