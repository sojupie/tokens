# Development

Install Go 1.26 or newer. Chrome or Chromium is needed for live crawling.
Node.js 22.18 or newer is needed for CSS generation and the browser-script checks. Serving a saved publication
and running the Go tests do not require Chromium.

For component authoring and state ownership, see [Go/templ components](reference/components-api.md).

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

To check rendered behavior requirements against a running Catalog, run
`go run ./cmd/check-behaviors --catalog http://127.0.0.1:8085`.
Chrome or Chromium is required; use `--chrome-path PATH` when needed.
Use `--path /components/search` to check one component.


To include captured-data checks, build assets for a local publication, then run:

```sh
npm run build:tokens -- --published published-review
TOKEN_ASSETS_DIR="$PWD/published-review/assets" npm run test:tokens
TOKEN_CATALOG_THEME_DB="$PWD/published-review/tokens.sqlite" go test ./catalog/web ./source/dtcg
```

Without these environment variables, optional captured-data checks skip.

Tests use local fixtures. Keep normal checks offline. Live captures and DNS or
Cloudflare checks are separate operational checks.

## Templates and assets

Page templates have named files such as `layout.templ` and `changes.templ`.
Component examples use `specimen_<family>.templ`; `component_specimen.templ` only
dispatches to the selected family. Matching `specimen_*.go` files hold attribute
helpers, while `component_gallery.go` holds the shared preview case model.

Edit templates under `catalog/web/templates/`, then run `npm run generate:web`.
This writes compiled Go to `catalog/web/internal/render/` and browser TypeScript
output to `catalog/web/assets/generated/`. Commit sources and generated output.
`npm run check:web` detects stale output. See [web ownership](../catalog/web/README.md)
for the package boundary and directory map.

The shared site header, navigation menu and footer live in `forge/site/`; the catalog renders those Go templates per request; Material Motion Lab renders them during its static build. `tokens site-shell --out DIRECTORY` exports the same markup and assets for the static carousel lab. Hosts also load the shared `navigation` CSS module and its `runtime/utils.js` and `runtime/navigation.js` scripts. The site uses a primary rail beside a scrolling section drawer on wide windows. Rail destinations reveal their section on hover or keyboard focus; clicking a destination follows its native link. Medium windows overlay the section drawer. Wide pages with no section drawer also reveal it over content without changing the reserved width. Section changes fade, overlay drawers enter and exit with effects/spatial motion roles, and only the drawer gains elevation while the rail or drawer is hovered. Compact windows use one modal drawer with main-menu and section views. The rail is a sibling of the drawer; drawer styles never own rail anatomy. Destinations use native links; opening, dismissal, focus return and swipe handling belong to the shared component runtime. The site shell owns section selection and preserves section-list scroll positions across HTMX swaps. Site CSS and JavaScript are maintained
in `catalog/web/assets/styles/`. Shared component CSS sources live in
`forge/components/styles/`; behavior sources live in `forge/components/behaviors/`.
New browser controllers and shared helpers use TypeScript. The migration currently
covers the asset loader, chips, switches, button groups, segmented buttons and
shared control/navigation helpers. Other controllers and most build/test scripts
still use JavaScript. `forge/short-css-names.ts` and the component asset builder
also use TypeScript. CSS grouping, module bundling, theme subsets and the shared
motion runtime have also moved to TypeScript.

Run `npm run check:types` for strict type checking. Node 22.18+ executes build
TypeScript by stripping types; it does not type-check it. `erasableSyntaxOnly`
rejects syntax that needs further runtime transformation. The existing esbuild
step compiles browser TypeScript to classic JavaScript at the same public paths.
Pure helper imports remain bundled into each entry. There is no browser-side
TypeScript compiler or new runtime dependency.

Run `npm run build:components` to assemble component sources into the existing
served files in `forge/components/generated/`. This also discovers TypeScript entries
in `catalog/web/assets/scripts/` and writes compiled JavaScript to `assets/generated/`. Edit the `.ts`
source when a generated file has that header. `npm run check:components` checks
both types and generated output, including Catalog scripts. Dedicated scripts such as `menus.js` and `fields.js`
are maintained in `forge/components/behaviors/` and compiled into `generated/`. Catalog's demo responses live in
`catalog/web/assets/scripts/component-demos.js`. Both sets of served assets are embedded
in the binary. Rebuild and restart the
server after changing them. Styles come from the published token data through
the [DTCG and CSS build](reference/dtcg.md) when generated assets are present.
`catalog/web/internal/server/theme.go` supports publications without generated assets. Layout
rules are in the static stylesheets.

The [component reference](reference/components.md) and [button reference](reference/buttons.md)
cover generated state bindings, shared interactions, and the local galleries.
The [CSS inspection and override instructions](reference/components.md#inspect-css-sources)
explain how to find manual declarations and change a token value.

## Asset delivery checks

The [delivery and performance reference](reference/delivery-performance.md)
collects the implemented backend/browser optimizations, measurements and limits.

`forge/components/modules.json` assigns each structural CSS source and generated
binding family to one module. `forge/component-bundles.ts` builds those modules
without inspecting selectors. The build rejects missing and duplicate owners.

Renderers request their module with `@useComponentAssets("list")`. Put the request
beside the markup that needs it, including nested controls and conditional
content. For example, the reorder dialog requests `sheet`; a switch row requests
`switch` through `listSelectionControl`. Thumbnails that author their own anatomy
request assets in their rendering branch too. Do not add route or child-family
maps. `@componentAssetScope()` collects requests from its children and inserts
versioned CSS and theme dependency links before their markup. Full HTML responses
hoist and deduplicate those links into the document head. Fragment responses
retain declarations for the client to prepare before swapping. Use a scope inside each HTMX
replacement boundary. Mark each scope container with `data-component-assets`. The layout scopes its
main content and footer; the playground has a
separate scope inside `#component-preview`, so changing a control declares its
current module dependencies. The generated module manifest supplies link order.

`catalog/web/assets/generated/asset-loader.js` selects optional behavior from rendered JSON
declarations. Register Forge ownership in `modules.json` and request its module
beside the rendering primitive; declare Catalog-owned example scripts at their
rendering branch. Keep diagnostic selectors aligned for the CLI comparison, but
do not rely on a selector to supply a missing declaration.

The layout loads shared CSS and core controls. Only modules rendered by the
initial page appear in its head. Before an HTMX swap, the loader adds missing
component styles in manifest order and waits for all of them in parallel.
Already loaded styles remain attached for the document lifetime; the fragment
links are removed before insertion. Failed stylesheet loads leave the current
content in place and can be retried. Version changes to loaded component styles
trigger a full reload, as do changes to core assets. This retains code splitting
without a stylesheet gap when replacing recurring controls. Browser history
keeps previously loaded styles available.
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

`forge/theme-subsets.ts` now derives theme inputs from shared CSS, template
references, family bundles and literal JavaScript reads. It follows CSS aliases
in all appearances. Captured DTCG aliases remain CSS references when their
resolved values agree after the site conversions. Public context exports retain
their existing output.
Dynamic spring property names have explicit dependency roots in `pageThemes`.
Add a root there if behavior introduces another dynamically constructed token
name. Unknown Material or Catalog references fail the build.

The shared theme stays in the head. Remaining module inputs are scoped to
`[data-component-assets]` regions and standalone `#component-preview` fixtures. Their links follow
the rendered content through HTMX navigation. Galleries collect the modules their
thumbnails render. Publications without theme bundles keep the complete theme.
`/theme.css` and public token downloads remain complete.

Open `/theme-aliases.html` to compare resolved variables with a saved
`baseline-theme.tmp` across all six appearances.

Open `/component-modules.html` on the conformance server to compare modules with
the complete stylesheet across all six appearances. It discovers public pages
from the sitemap, including agent instructions, change logs and individual token
sets. It also tests each playground control option from default and interactive
views. Computed styles include state-layer pseudo-elements. Removing card CSS
and card tokens independently must produce differences as a check of the test
itself. Add `?path=/md3/components/search` to rerun one component and its options.
Add `&width=320` or `&width=390` to check a phone-width frame. Inline date
pickers also fail the check if their panel or content overflows horizontally.
This compares delivery with the full output; it does not establish that
the full output implements every Material rule correctly.

The conformance proxy defaults to Catalog at `http://127.0.0.1:8081/md3`.
Pass `--catalog` to use another base URL.

`/list-reorder.html` checks the
placement and layout effect of the reorder indicator. Run the Go render tests
with `go test ./catalog/web`.

For a theme comparison, copy the complete generated `theme.css` to
`forge/conformance/fixtures/baseline-theme.tmp`. Open `/theme-equivalence.html` to compare
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
additional variables scoped to the rail preview. Overview pages in that run
loaded the full set for mixed component thumbnails; current image-backed cards
can omit those live compositions.

For a CSS build comparison, save the previous theme and full component stylesheet
as `forge/conformance/fixtures/baseline-theme.tmp` and `baseline-components.tmp`, then open
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

The script triggers the upstream HTMX 4 preload extension through a custom event.
The extension constructs the request and retains its response for five seconds.
A matching HTMX navigation consumes that response, including an in-flight
request. Expired or network-failed prefetches fall back to normal navigation. Automatic
mousedown and touch preloading are disabled; the hover/focus policy owns when
requests start. Rendered responses retain the revalidation policy below.

To check this in a browser, hover a component link briefly, then for more than
65 ms. In DevTools, check the GET request. Click within five seconds and confirm
that HTMX consumes the response without a second document request. After the
response arrives, the asset loader starts the incoming HTML's script downloads
while preparing its CSS. Scripts execute after the new region is inserted and
its CSS is ready. This does not fetch destination scripts during the hover.

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

HTMX 4.0.1 boosts internal navigation and preserves `#tokens-site-header`.
The official `hx-head` extension merges document metadata. An asset-version guard
forces a full navigation when deployed core assets change; dynamically loaded
component scripts remain mounted. The official 2.x compatibility extension keeps
existing lifecycle listeners working during migration.

Normal and boosted navigation share the document cache. An inert
`<template hx type="partial">` in that document updates only
`#tokens-site-destinations`; the drawer and appearance controls retain their
instances. Component and style navigation comes from the loaded publication and
is cached when the server starts. The section list retains its scroll position
while its current-page link updates. Index search uses `hx-sync="this:replace"` so late responses cannot
replace a newer query.

Playground requests and token-table submissions replace their own regions.
Table submissions retain the latest playground query settings, view preferences
and folder expansion. Table and modal-field styles stay inside their replaceable
asset scopes, so repeated updates remove their previous stylesheet links. Context validation replaces only its fields; Cancel restores the
unapplied values. Requests use htmx synchronization and cancellation, while local
JavaScript owns focus, disclosure state and pending/error feedback.

Download controls declare their request targets and dependent fieldsets from
Go's export-selection model. Structural choices request
`/downloads/export/options`, which returns resolved fields through `outerMorph`
and updates the summary and action through htmx 4 partials in the same response.
Format-only changes request `/downloads/export/summary` directly. Both use the
form's `replace` synchronization scope.

An index built from the loaded publication stores context vocabulary, parsed
value labels and token references. Option resolution visits selected tokens and
their references without reparsing captured JSON or scanning unrelated subjects.
Selection summaries count jobs using the same `exportSelections` rules as the
download. They do not convert tokens, render specimens or gather asset contents.
Exact conversion/dependency/omission counts remain in the on-demand report;
implementation archives are assembled only when downloaded.

`hx-disable` blocks stale dependent fieldsets and a temporary guard around
Download; option checks retain its disabled appearance. A separate
confirmation fieldset represents server-confirmed validity. The loading button
is preserved across partials, keeps the label Download and remains in the docked
footer beside the summary and report link. Local JavaScript preserves disclosure
state, retains draft values from temporarily disabled children during rapid changes, invalidates the action,
and handles opening, cancellation and errors. Morph swaps retain control focus.
The form still submits a native file download. During preparation, its button
uses the button loading primitive and blocks duplicate submissions. A per-request
short-lived response cookie clears that state when attachment headers arrive;
a hidden transfer frame keeps preparation errors from navigating the page.
Acknowledged responses are not cached. No download body passes through JavaScript.
The public JSON options and
summary API remain available. Multi-format and context selections keep their ZIP
behavior.

Run `/htmx-catalog.html` through the conformance server to check region identity,
context cancellation, download selections, metadata, drawer updates and history.
Native links and GET forms remain usable without JavaScript.

## Rendering and motion

Component cards and token-table wrappers use `content-visibility: auto` with
intrinsic placeholder sizes. Browsers can skip offscreen layout and paint,
particularly on mobile. The DOM and styles are still delivered; this is not DOM
virtualization. Printing disables the containment optimization.

The landing page shows a small component/style selection, then a separate
section for change history and agent access. Hero artwork lives in
`catalog/web/assets/scripts/artwork.js`. Its color-block choreography and loop timing
are authored; sampled captured spatial springs drive resize and shape
transitions. Stripe turns use the standard easing and medium2 duration tokens
to settle without overshoot. Grid resizing, directional corner changes and stripe
turns share one scene clock, with a transition every two seconds and no column
delays. Different tiles turn on different beats; held scenes preserve variety.
The pause control, reduced-motion preference, document visibility and viewport
intersection control playback.

The mosaic has an inset inside the frame; complete blocks stay visible throughout
the loop. The default renderer uses Canvas 2D in `catalog/web/assets/scripts/artwork-canvas.js`.
Add `?artwork=dom` to the homepage to compare the DOM/CSS renderer.
Both use the same tracks and spring samples from `artwork.js`. Canvas resolves
palette colors through CSS on appearance changes, resizes its backing buffer with
the viewport, and caps bitmap density at 2x. It sleeps during settled holds and
cancels its frame/timer callbacks while paused, hidden or offscreen. The HTML tiles
remain as a fallback if a 2D context cannot be created. Only the Canvas variant
loads the additional renderer script.

Carousel touch scrolling uses native momentum and snapping. The visual track uses native sticky positioning.
Signed scroll offsets are clamped at both ends so Safari rubber-banding does not
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
`forge/components/motion.ts`, compiled to `forge/components/generated/motion.cjs`
and served as `motion.js` before component code. Both the browser entry and the
CSS compiler import spring sampling from `forge/components/helpers/springs.ts`,
so the compiler does not read a potentially stale generated JavaScript copy.

The [motion contract and browser checks](reference/components.md#motion-schemes-and-lifecycle)
describe assignments, interruption handling, reduced motion and popup fallbacks.
Run `go run ./forge/conformance` alongside Catalog and open
`http://127.0.0.1:8082/motion.html` to repeat the checks.

## Navigation benchmarks

Run Catalog, then measure its real navigation and download controls in Chrome:

```sh
go run ./cmd/bench-navigation --catalog http://127.0.0.1:8085 \
  --chrome-path /usr/bin/chromium --runs 15 --output /tmp/catalog-benchmark
```

The command writes raw samples to `results.json` and prints median and p95 times.
It captures an additional trace per scenario, excluded from those statistics,
and records another 500 ms after readiness to include subsequent paints.
Import a `*.trace.json` file into Chrome DevTools' Performance panel to inspect
main-thread work, layout, paint and screenshots. Use `--traces=false` to omit them.

Navigation cases cover cold destination assets, warmed HTTP assets, reused
prefetches and expired prefetches. Each sample starts in a fresh document; warm
assets do not mean previously initialized component instances survive. The source
page still loads its own dependencies in every case. Server response caching is
independent: `Server-Timing` identifies cache hits and handler rendering on misses.
The command checks that prefetch scenarios actually reuse or expire their response.
The download case measures changing content in an already open dialog until the
new options are confirmed and Download is enabled.
Use `--journey gallery` to measure returning from Cards to the component index.
`--scenario cold|warm|prefetch|expired` limits a run to one cache condition.
Gallery samples include `visibleImages` and an `images-ready` mark after the
visible full-resolution candidates decode. Resource entries distinguish cached
image bytes (`transferBytes: 0`) from network transfers. These marks exclude the
intentional image fade; decoding and finishing that transition are separate.

`dom` means the expected content is present. `ready` means the navigation asset
loader has finished, or the download selection is confirmed. `frame-opportunity`
records two animation-frame callbacks after the content change. None of these
marks proves a paint; use the trace's paint events and screenshots for that.
Activation uses a programmatic click, so these are not physical-input or INP
measurements.

Keep browser, hardware, power settings and viewport consistent between runs.
The default viewport is 1440 × 900; change it with `--width` and `--height`.
Samples record the browser, viewport, resource sizes, protocol and server timings.
Use `--journey cards`, `date` or `download` to limit a run. Measure slower clients
separately with `--cpu-rate 4` or `--latency-ms 150`; the latency option does not
limit bandwidth. Establish repeat-run variance before setting CI thresholds.

### Custom-property name experiment

```sh
go run ./cmd/bench-css-names --catalog http://127.0.0.1:8085 \
  --chrome-path /usr/bin/chromium --runs 15 --output /tmp/css-names
```

This captures rendered Cards, Date Picker and Carousel pages, then compares the
same frozen DOM with original and shortened custom-property names. Application
scripts and animations are removed from both arms. An identical small controller
keeps benchmark theme/contrast controls usable. This measures CSS costs without changing
the site's runtime or its public token names. JavaScript consumers, including
dynamically constructed property names, still need a shared naming contract
before this transform can be used in a live build.

The transform uses PostCSS to visit declarations, `var()` references and relevant
at-rule parameters. It leaves quoted text and other dashed identifiers alone.
Names are ordered by occurrence count, with a lexical tie-break, then assigned
unique identifiers from a 64-character ASCII alphabet. The first 64 names need
one character after `--`; the next 4,096 need two. The mapping is deterministic
for the complete captured corpus, independent of file order. Changing that
corpus can change assignments: it is not a stable public API or a per-file pass.
Escaped and non-ASCII custom-property identifiers are rejected by this experiment.

Each arm uses prebuilt gzip responses, cold browser caches and a 1440 × 900
viewport. Run order alternates each round. Results include raw and gzip CSS
sizes, observed CSS and total page transfers, first contentful paint, load-time
style/layout counters, isolated parsing with `CSSStyleSheet.replaceSync`, and
forced style recalculation while switching themes. Appearance tests run before
the isolated parsing loops, so their stylesheet allocations cannot affect those
tests. These are full-document loads and attribute changes on frozen pages, not
the site's enhanced navigation or component-controller interactions.

The first pair must have identical computed standard styles and geometry across
the rendered elements in all six theme/contrast appearances. Both screenshots are saved for visual comparison, along
with `names.json`, source snapshots and raw measurements in `results.json`.
Use `--names original`, `short` or `both` to select an arm, `--paths` to change the
captured pages, and `--cpu-rate 4` to run a separate slower-client profile.
Run from the repository root with Node dependencies installed.

To combine actual rendered specimens from 14 component families into one page:

```sh
go run ./cmd/bench-css-names --catalog http://127.0.0.1:8085 \
  --chrome-path /usr/bin/chromium --mixed --runs 15 \
  --listen 127.0.0.1:8096 --output /tmp/css-names-mixed
```

The command prints original/short page URLs and keeps them available until
interrupted. Components retain frozen behavior; the page's Theme and Contrast
controls work. `--paths` overrides the captured families. Shared CSS is deduplicated
and optional modules retain manifest order. `--input /tmp/css-names-mixed/original.json`
replays the exact captured corpus for another run or CPU profile instead of
recapturing pages; keep the same `--catalog` for referenced fonts and images.

Each appearance test covers both theme directions at default, medium and high
contrast, and all six directed contrast pairs in both light and dark themes.
One complete transition block warms up; two measured blocks reverse their order.
Raw samples record each transition's synchronous forced style/layout duration and
elapsed time through two animation-frame callbacks. The latter includes frame
scheduling and is not a paint-duration measurement.

## Site and scene layout

Follow [the layout contract](reference/layout.md) for region grids, margins,
spacing, app-bar title ownership and scene framing. Shared measurements live in
`forge/site/site.css`; content-specific spans stay with the page or scene.

## Component illustrations

Catalog cards use generated images of the simple previews. Interactive playgrounds keep their live markup.
Generate thumbnails for each new publication before serving it. Theme rebuilds
preserve images only when their manifest matches the publication. Missing images
show a placeholder; catalog cards never substitute live component markup.

Generate the simple catalog thumbnails after building token assets:

```sh
go run ./cmd/tokens illustrations --published published-review --chrome-path /usr/bin/chromium
```

The command captures 720 × 450 thumbnails in six appearances and publishes
`illustrations/manifest.json` after every capture succeeds. Rendering lives in
`catalog/web/templates/previews/illustrations.templ`; component anatomy and measurements remain in
Forge. The playground stays interactive.

Each capture produces 720px and 360px JPEGs plus a 24 × 15 placeholder embedded
in the manifest. Cards inline the placeholders for immediate blurred artwork and
use native lazy loading with responsive image candidates. The full image fades
in after decoding; reduced motion disables the transition. Aspect ratios reserve
card space, and full-page refresh leaves scroll restoration to the browser.

When exporting the shared shell for a static tool, include the publication to
build the same component and style links without a browser-side navigation fetch:

```sh
go run ./cmd/tokens site-shell --published published-review --out /tmp/site-shell
```

The `--published` argument is optional for consumers that need only the primary
navigation and tool links. Appearance controls stay docked in the mobile drawer;
on desktop the rail's Appearance button opens those same controls in a popover.

### Behavior without JavaScript

The shared site shell provides a native disclosure navigation when scripting is
unavailable. Catalog playgrounds and token filters submit GET forms; their named
select elements also hold the values used by the enhanced selectors. Do not add
an additional hidden input with the same name. The `/download` page uses the
existing export option resolver and download endpoint with ordinary form submits.

Forge supplies native range, date/time entry, filter-chip and list-selection
fallbacks. Tab content stays readable, progress uses native `<progress>`, and the
loading indicator can rotate with CSS. Dialog triggers carry native invoker
commands alongside the enhanced controller hooks; controllers cancel the native
default when they handle an action. Browser support for invoker commands controls
whether basic modal opening works without scripting.

Exact spring/morph motion, draggable sheet anchors, custom calendar/dial behavior,
live search suggestions, reorder/swipe gestures and application commands remain
JavaScript enhancements. Their absence must not remove site navigation, token
content, form choices or downloads. Keep fallback styling in the owning component
module and catalog presentation in the catalog stylesheets.


### CSS names in releases

Production asset builds use shortened CSS custom properties by default.
`deploy/build-assets.sh` passes `--css-names short`; set
`TOKENS_CSS_NAMES=original` to release original names. Local `build:tokens`
builds default to original names and accept `--css-names short` for review.

The delivery pack shares one deterministic name map across site, component and
theme CSS. Names referenced by controllers, inline styles or templates retain
their original spelling. The build scans those sources automatically; there is
no browser translation table. DTCG, exported CSS and component implementation
downloads retain original names and bytes.

The server verifies the pack against its embedded styles and loaded themes at
startup. Profiled stylesheet URLs include both the content hash and name-map ID.
Original stylesheet URLs remain available to static consumers. See the
[delivery details](reference/delivery-performance.md#production-css-delivery).

Full custom-property rewriting and class/ID shortening remain separate benchmark
experiments; the release pack does not rewrite runtime selectors or markup.
