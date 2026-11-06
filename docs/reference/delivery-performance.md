# Delivery and on-page performance

This document describes the work that reduces request processing, transferred
assets and browser work in Token Catalog. Implemented behavior is separated from
experiments and remaining limits. Measurements below were recorded on November
5–6, 2026; they are local baselines, not production service targets.

The standards-mode comparison at the end supersedes the earlier short-name
browser timings. The earlier replay omitted the doctype and used quirks mode.
Its payload counts remain valid; its rendering timings do not establish gains
for the standards-mode site.

## Request and browser flow

1. The server loads one immutable publication and its verified generated assets.
2. A cacheable request either reuses a rendered response or renders Go/templ HTML
   from the loaded publication. Ordinary browsing does not query SQLite.
3. Full documents put shared CSS and the rendered components' dependencies in the
   head. The browser can discover those styles before reaching component markup.
4. Eligible hover/focus intent can fetch destination HTML ahead of navigation.
   HTMX owns that response's reuse and expiry.
5. When navigation consumes HTML, its declarations start optional script fetches
   and prepare missing styles. Existing content remains until those styles load.
6. HTMX inserts or morphs the content. The preserved navigation shell remains in
   place; already delivered component styles and scripts remain available.
7. Optional controllers evaluate in order after CSS resolves. Readiness events
   allow controllers to adopt the new markup and release playground controls.

This uses server-rendered HTML and incremental JavaScript activation. It does not
serialize and resume component execution state, and it has no React-style
hydration pass. Downloads, external links and non-document API links stay native.

## Backend work

### Publication and generated assets

`publish.Load` reads the publication into `PublishedState` and closes SQLite.
Server startup resolves a particular publication generation, so later activation
of another generation does not mix its data or downloads into the running server.
Navigation metadata is constructed from this state. Generated themes and CSS
value indexes are loaded with publication and asset-checksum checks. Illustration
manifests pin the publication and validate the listed files and appearances.
Template compilation and CSS generation happen before serving.

Decoded preview configuration files and representative thumbnail cases are
cached on the server. Repeated card rendering and playground updates do not
reparse the same coverage JSON. Module order and the script-source URL inventory
also use one-time initialization.

Optional build-time history caching reuses semantic history work. Entries carry
a processing version, identity and data checksum; corrupt or obsolete entries
are misses. This reduces publication-build work, not browser navigation time.

Sources: [publication loading](../../source/publish/memory.go),
[history cache](../../source/publish/history_cache.go),
[server startup](../../cmd/tokens/serve.go),
[generated assets](../../catalog/web/generated_assets.go),
[preview configurations](../../catalog/web/component_preview.go),
[thumbnail cases](../../catalog/web/component_thumbnails.go).

### Rendered-response cache

The per-handler LRU retains successful cacheable GET/HEAD responses. Its body and
key budget is 32 MiB, with a maximum of 256 entries. Full URLs, HTMX request and
target headers, accepted representation and accepted encoding distinguish cache
entries. Boosted document requests normalize their HTMX variant to reuse the
full-document representation.

A hit avoids template rendering, token-table construction, HTML dependency
hoisting and response compression. HEAD can seed the GET representation without
storing an empty body. Matching ETags return 304. HTML uses `public, no-cache`:
the browser may store it but must revalidate; this is distinct from the server's
in-memory cache and HTMX's prefetch cache.

Errors, `no-store` responses and responses setting cookies are not stored. Range
requests, static/theme assets, unscoped download paths, health, usage and MCP
requests bypass this response cache. Restarting or replacing the handler clears
it. Arbitrary filter/search URLs are bounded by eviction rather than retained
indefinitely. Header and object overhead are outside the stated byte budget.
Concurrent cold misses are not coalesced into one render.

`Server-Timing` reports current cache lookup time on a hit, or handler rendering
time on a miss or a handler result that cannot be stored. Endpoints skipped by
the wrapper are not timed here. Cached responses do not replay an old render
duration. These measurements exclude work outside the timed handler, network
transit and browser processing.

Source: [response cache and timings](../../catalog/web/response_cache.go).

### Static transfer and compression

Asset URLs include a SHA-256 content version. Correctly versioned URLs receive a
one-year immutable cache lifetime. Bare URLs and stale version parameters require
revalidation. Identity and gzip representations have distinct ETags; encoding
negotiation contributes `Vary: Accept-Encoding`.

CSS, JavaScript and SVG gzip bytes are generated once per content identity and
reused in process. HTML is gzipped while rendering a cache miss; its encoded
response can then be reused by the response cache. There is no application Brotli
encoder. HTTP/2, HTTP/3, edge caching and connection behavior depend on the hosting
path; the application does not guarantee a particular transport protocol.

Sources: [static assets](../../catalog/web/static.go),
[encoding](../../catalog/web/encoding.go).

### Download option checks and attachment generation

The download dialog uses a publication-derived index, built once on first use.
It stores token names, context labels, references, whether a value exists and
whether it can be converted. Each distinct raw value is tested for convertibility
once while building that index, rather than reconverted for every option change.

The dialog's options and summary endpoints inspect this metadata. They do not
generate DTCG documents, component implementations or ZIP archives just to enable
a button. Context choices follow selected values and their reference dependencies;
omission warnings reuse the indexed metadata and deduplicate equivalent selections
across formats. This is not an O(1) operation overall: selection and dependency
traversal still scale with the relevant tokens and values.

Actual export generation starts on attachment submission. A batch reuses each
selection's converted document and report when writing its file and conversion
metadata. It does not convert again to produce the accompanying report. Reports
are included in token ZIPs and scoped DTCG metadata rather than requiring a
separate pre-download report action. Combinations are limited to 64 and batch
work checks request cancellation. Archives are still built in memory. Legacy
public export requests using `summary=1` can still run `exportDocument`; they are
not the dialog's cheap summary endpoint.

Sources: [index](../../catalog/web/export_index.go),
[context selection](../../catalog/web/export_context.go),
[dialog checks](../../catalog/web/export_dialog.go),
[batch generation](../../catalog/web/export_batch.go),
[export rendering](../../catalog/web/export_selection.go).

### Usage recording

When enabled, usage recording enqueues events without waiting for disk. A worker
batches SQLite updates in transactions, flushing at 256 events or once per
second. The queue holds 4,096 events; overflow is counted as lost usage rather
than blocking page delivery. This trades complete telemetry under overload for
bounded request-side work.

Source: [usage store](../../catalog/usage/store.go).

## CSS delivery and style calculation

The CSS build now separates component modules instead of requiring the complete
component stylesheet on every page. Ownership of structural CSS, generated
bindings and behavior scripts lives in `forge/components/modules.json`; build
checks reject missing and duplicate owners. Rendered primitives request their
own modules, including nested and conditional controls. Page routes do not
maintain a second component dependency list.

The theme build preserves compatible captured aliases as CSS references. It
emits a base appearance and only changed declarations for the other five
appearances. Repeated binding selectors are grouped only when doing so preserves
cascade order. Generated CSS has compact whitespace.

Theme subsets follow references transitively through every appearance. Shared
site/control inputs stay on the root. Optional module inputs are scoped to
`[data-component-assets]` or standalone `#component-preview`, reducing the inherited
variable set outside those regions. Literal runtime reads are included; dynamically
assembled spring names have explicit dependency roots. Missing Material/Catalog
dependencies fail the build. Public theme/token exports retain their full data.

The initial head contains the shared theme, site/application/table/button/common
styles, navigation styles and the optional modules actually rendered. These are
separate files; putting their links in the head does not merge them into one
download. Gallery fallback previews can defer their optional styles; navigation
remains blocking. Publications without theme bundles retain the complete-theme
fallback.

Sources: [CSS output](../../forge/css-output.ts),
[module builder](../../forge/component-bundles.ts),
[theme subsets](../../forge/theme-subsets.ts),
[rendered requirements](../../catalog/web/component_assets.go),
[initial head](../../catalog/web/layout.templ).

## Static tool consumers

Material Motion Lab renders shared Go/templ markup during its build and packages
matching component assets. It does not need a Catalog request or the Catalog
loader to construct that UI. `tokens site-shell` similarly exports shared shell
HTML and site assets for static consumers; their builds supply the matching Forge
dependencies. Stateless source imports are bundled into those same public script
entries, so consumers do not acquire a separate browser import waterfall.

The static tools retain their own delivery and initialization paths. Catalog's
HTMX scheduling changes do not automatically change their readiness contract.
Their shared-code and asset checks remain necessary when updating the host or
converting runtime globals.

Sources: [shared site/tool architecture](../architecture.md#shared-site-and-tools),
[shell export](../../cmd/tokens/site_shell.go),
[behavior ownership and consumers](behavior-delivery.md).

## Browser delivery and reuse

### Styles before insertion

Each independently replaced region declares its dependencies. Full-document
rendering hoists and deduplicates component stylesheet links into the head.
Fragment responses retain the declarations for the client to consume.

Before a swap, the loader inserts missing stylesheet links in manifest order and
waits for them together. Loaded styles stay in the head for the document lifetime,
so returning to a component avoids another stylesheet gap. Fragment links are
removed before insertion. Failed styles leave current content intact and permit
a retry. A changed version of an already loaded module, or changed core assets,
forces a document navigation instead of mixing deployment versions.

This improves recurring controls without eagerly loading every component on
every initial page. The trade-off is that visited modules accumulate in the
current document's stylesheet set.

### Declared behavior and ordered evaluation

Rendered JSON declarations select optional behavior downloads. Selector rules
remain diagnostic checks, not a second delivery authority. Catalog example
scripts are declared separately from exported Forge component requirements.

Known scripts receive native preload links together. Fetching overlaps CSS
preparation, while script evaluation remains ordered after CSS readiness.
Successful scripts and in-flight promises are reused across swaps. Failed loads
release their pending entries and preload links for retry. A generation check
prevents superseded activations from enabling controls or reporting stale errors.
Playgrounds remain inert while required assets load, with a failure message if
activation cannot finish.

Stateless helpers use source imports bundled by esbuild into the existing classic
script entries. Missing imports fail compilation, and helpers add no browser
requests or global execution-order dependency. Stateful shared services, including
`MaterialMotion`, still use host-provided globals and ordering. Independent helper
bundling is not a safe way to duplicate those services.

`TokenAssets.ready` and `tokens:assets-ready` mark the loader's CSS/script barrier.
They do not certify that every asynchronous controller task has finished or that
pixels have painted. Loading-indicator geometry is an explicit remaining example.

Sources: [loader](../../catalog/web/static/asset-loader.js),
[navigation integration](../../catalog/web/static/site-ui.js),
[behavior compilation](../../forge/components/build-assets.ts).
The [behavior delivery contract](behavior-delivery.md) has the lifecycle inventory.

### Navigation intent and persistent UI

Eligible same-origin document links prefetch after 65 ms of mouse/pen hover or
keyboard focus. Save-Data, known slow-2G/2G connections, hidden documents,
downloads, external targets and fragment links are excluded. The policy is capped
at eight distinct destinations per document lifetime. The upstream HTMX extension
owns its single-use response cache and five-second expiry. The destination remains
in the local seen set after expiry; it is not repeatedly prefetched.

This fetches HTML ahead of a click. Destination script preloads start when
navigation consumes that HTML, not during the hover request. There is no general
prediction of future component interaction or fetch-on-first-click controller
scheme.

The navigation header is preserved across body updates. Rail and drawer
destinations morph within it, preserving recurring DOM rather than recreating the
whole shell. Section scroll positions and gallery offsets are retained. Search
and playground/table controls request their replacement regions rather than a
new full document; subject search is debounced by 250 ms and replaces superseded
requests.

Sources: [prefetch policy](../../catalog/web/static/prefetch.js),
[HTMX extension](../../catalog/web/static/hx-preload.min.js),
[site shell](../../forge/site/site.templ),
[subject search](../../catalog/web/subject_list.templ).

## Work after the page arrives

Generated thumbnail images replace live component compositions where the
publication supplies them. They have fixed dimensions, lazy loading and async
decoding. Missing thumbnails fall back to inert HTML/CSS previews. Those scopes
omit optional interaction scripts, except `progress.js`, which draws static SVG
paths. The server also caches generated thumbnail case descriptions. Interactive
playgrounds keep their live components.

Several controls use delegated listeners instead of binding every node. Per-node
controllers have family-specific initialization and cleanup hooks; dialogs bind
their presentation listeners and resize observer on opening and release them on
close. Closed dialog markup can still declare and download its scripts. Deferring
an instance's work is separate from deferring its bytes.

Loading/progress and artwork controllers use visibility, reduced-motion and
document-visibility checks to limit animation work. Cleanup releases tracked
nodes, observers and animations when markup is removed. These are component-level
policies, not a universal activation scheduler for every offscreen element.

Download option changes disable dependent fields through HTMX request attributes,
not the entire interface. Replace synchronization cancels superseded checks; only
the server-confirmed selection enables Download. Disclosure state and confirmed
checked properties are applied before morphing to avoid transient collapsed or
stale states. A separate button-loading state covers attachment preparation.
Native forms remain usable without JavaScript.

Sources: [thumbnails](../../catalog/web/subject_list.templ),
[image markup](../../catalog/web/illustrations.templ),
[dialogs](../../forge/components/static/dialogs.js),
[loading](../../forge/components/static/loading.js),
[progress](../../forge/components/static/progress.js),
[artwork](../../catalog/web/static/artwork.js),
[download client](../../catalog/web/static/export.js).

## Measurements

### Navigation baseline

November 5: Linux, Headless Chrome 157.0.8103.38, 1440 × 900, publication
`material-2026-10-15-review`, local HTTP/1.1, no CPU/network throttling. Each case
has ten untraced samples. Trace samples are separate because tracing adds work.

| Interaction | Median loader/selection readiness | p95 |
| --- | ---: | ---: |
| Cards, cold destination assets | 78.8 ms | 90.2 ms |
| Cards, warm HTTP assets | 70.2 ms | 91.6 ms |
| Cards, reused HTML prefetch | 70.1 ms | 72.9 ms |
| Date Picker, warm HTTP assets | 82.8 ms | 85.1 ms |
| Date Picker, reused HTML prefetch | 81.2 ms | 84.9 ms |
| Download option change, warm assets | 13.3 ms | 14.2 ms |

Warm navigation requests took about 1.2 ms. An inspected Cards trace spent about
24 ms updating styles, 9 ms in layout and 11 ms parsing HTML. The application also
parses incoming HTML before HTMX parses/processes it; that application parse was
about 4.4 ms in the warm Cards samples. These observations identify browser work
worth investigating; they do not establish that all remaining time has one cause.
An uncached Cards request spot-check reported 8.098 ms of handler rendering,
followed by a 0.001 ms cache lookup on the repeated request.

Each sample starts in a fresh document. Warm means HTTP assets were primed, not
that existing component instances survived a prior navigation. Activation uses a
programmatic click, not physical input. DOM presence, loader completion and two
animation-frame callbacks are not paint metrics or INP. Extended filmstrips show
the destination appearing after loader readiness; traces continue for 500 ms
after that barrier. The benchmark verifies actual reuse/expiry in prefetch cases.

### Short-name experiment

**Experimental; not enabled in the live Catalog build.** The generator rewrites
custom-property declarations, references, relevant at-rule parameters and inline
styles across one closed corpus. Frequency order with a lexical tie-break assigns
the shortest available identifiers. All 3,427 names in the three-page corpus fit
into 64 one-character and 3,363 two-character names after `--`. Mapping is
collision-free and deterministic for those inputs, independent of file order.
Adding inputs can change assignments. Public token names remain unchanged.
The leading `--` is required by the [CSS custom-property syntax](https://www.w3.org/TR/css-variables-1/#defining-variables).
Only the identifier after that prefix is shortened; bare `--` is reserved.

The experiment captures rendered pages, then removes scripts and animations from
both versions. It serves prebuilt gzip with cold browser caches and alternates
version order. Parsing uses fresh `CSSStyleSheet.replaceSync` calls after page
load measurements. Table values are medians of per-load parse medians, with 25
parse iterations per load; those iterations are not independent page-load samples.
There were 15 page loads per version, using the browser/publication above.

| Page | Raw CSS, original → short | Gzip CSS payload | Isolated parsing |
| --- | ---: | ---: | ---: |
| Cards | 772,724 → 353,898 B | 83,429 → 65,390 B | 29.5 → 25.0 ms |
| Date Picker | 833,510 → 382,993 B | 88,878 → 69,376 B | 31.7 → 27.1 ms |
| Carousel | 771,462 → 358,062 B | 84,418 → 66,502 B | 29.8 → 25.6 ms |

Raw CSS fell about 54%; gzip payload fell 21–22%; isolated parsing improved
14–15%. Median observed whole-page Resource Timing transfers were:

| Page | Original | Short |
| --- | ---: | ---: |
| Cards | 374,228 B | 356,176 B |
| Date Picker | 166,102 B | 146,575 B |
| Carousel | 1,198,022 B | 1,180,042 B |

Resource Timing transfer sizes include the browser's header accounting; they are
not packet-level wire measurements. Raw/gzip CSS figures cover the loaded styles,
not every possible component or appearance export in the repository.

Initial style calculation improved by less than 1 ms and layout barely changed.
Median first contentful paint was 112 → 116 ms on Cards, 84 → 76 ms on Date Picker
and 116 → 108 ms on Carousel. The size and isolated parsing gains did not produce
a consistent first-paint improvement across all three pages.

A separate Cards-only corpus, seven loads per version with 4× CPU throttling,
reduced per-load median parsing from 123.5 to 106.4 ms and median first contentful
paint from 212 to 200 ms. It generated its own mapping; it is not a replay of the
three-page artifacts. Both experiments matched computed standard styles and
geometry; all three desktop screenshot pairs were byte-identical. Regenerating
the three-page transform produced identical artifacts.

Live adoption still needs CSS, inline styles, exports and JavaScript reads/writes
to share the map, including dynamically assembled names. The experiment rejects
escaped/non-ASCII custom-property names. It adds no runtime translation wrapper,
and its results do not measure interactive application behavior.

### Mixed-component appearance benchmark

The mixed fixture combines rendered specimens from 14 component pages: buttons,
button groups, cards, carousel, checkbox, chips, lists, progress indicators,
radio buttons, segmented buttons, sliders, switch, tabs and text fields. It uses
the same frozen component markup as the individual-page experiments. Shared
stylesheets are deduplicated and component modules retain manifest order. The
result has 958 elements. Its native Theme and Contrast controls change root
attributes through one identical small controller in both variants; component
behavior remains frozen.

Both theme directions are measured at all three contrast levels. Contrast tests
cover every directed pair of default, medium and high in both light and dark
mode. Each fresh page warms all 18 transitions, then measures two blocks with
reversed order: 12 theme changes and 24 contrast changes. Timing includes the
attribute changes and a forced style/layout flush. A second duration continues
through two animation-frame callbacks, which includes scheduling rather than
isolating paint. The isolated parse benchmark runs afterward to keep its
stylesheet allocations out of the appearance measurements.

The first pair compares computed standard styles and geometry in all six
appearances. The capture can be replayed with `--input` to keep source HTML,
CSS and the generated mapping fixed between profiles. `--listen` keeps original
and short-name pages available for review after the run; measurements use
loopback even when that review server listens on a LAN address.

The November 5 mixed run used the same Chrome 157/Linux, publication and
1440 × 900 viewport as above, with no throttling and 15 loads per variant. It
replayed one captured corpus for the final measurement order. There were 180
measured theme changes and 360 contrast changes per variant. Table timings use
the median of each page load's median, so changes within one load are not treated
as independent page samples.

| Measure | Original names | Short names |
| --- | ---: | ---: |
| Raw CSS | 969,100 B | 444,703 B |
| Gzip CSS payload | 109,199 B | 85,163 B |
| Observed CSS transfer | 121,799 B | 97,763 B |
| Observed whole-page transfer | 1,428,978 B | 1,404,847 B |
| Isolated parsing | 37.7 ms | 32.1 ms |
| Initial style calculation | 33.42 ms | 32.84 ms |
| Initial layout | 18.42 ms | 18.48 ms |
| Theme change, forced style/layout | 13.05 ms | 12.60 ms |
| Contrast change, forced style/layout | 13.55 ms | 13.25 ms |
| Theme change, through two frame callbacks | 33.30 ms | 33.35 ms |
| Contrast change, through two frame callbacks | 33.30 ms | 33.30 ms |

The 3,963-name map used 64 one-character and 3,899 two-character identifiers after
`--`. Gzip payload fell 22.0% and isolated parsing improved 14.9%. Appearance
changes saved about 0.3–0.45 ms of synchronous work, with no meaningful improvement
through the frame callbacks. Median first contentful paint was 108 → 104 ms;
that small difference is not evidence of a dependable first-paint speedup.
Computed styles and geometry matched in all six appearances, and the default
screenshots were byte-identical.

Keep shortening as a candidate build-time optimization for its measured byte
and parsing savings. It is not yet a live Catalog build option. Production
integration must rewrite runtime references through the same map; shipping a
large original-to-short-name dictionary to the browser could consume the payload
saving this experiment is intended to preserve.

### Earlier CSS measurements

The [development guide](../development.md#asset-delivery-checks) retains the
October 22 Chromium 155/Linux measurements for module splitting and theme
scoping. The recorded `/changes` CSS reduction was 3.00 → 1.24 MB raw and
220 → 101 kB gzip; script requests fell from 22 to 12. A later same-document
scoping comparison reduced root variables from 4,115 to 2,017, with 153 additional
variables scoped to the rail preview. Those historical comparisons were not
rerun for the November baseline and must not be combined into one speedup figure.

## Reproduction and checks

Run from the repository root with Catalog, Chrome and the root Node dependencies
available. [Development commands](../development.md#navigation-benchmarks) explain
each benchmark's scenarios, flags and output files.

```sh
go run ./cmd/bench-navigation --catalog http://127.0.0.1:8085 --chrome-path /path/to/chrome --runs 15 --output /tmp/catalog-navigation
go run ./cmd/bench-css-names --catalog http://127.0.0.1:8085 --chrome-path /path/to/chrome --runs 15 --output /tmp/catalog-css-names
go run ./cmd/check-behaviors --catalog http://127.0.0.1:8085 --chrome-path /path/to/chrome
```

The short-name CLI supports `--names original|short|both`, custom `--paths` and
`--cpu-rate`. It writes the name map, input/output snapshots, screenshots and raw
samples. The navigation CLI records resource sizes/protocols, browser version,
viewport, server timings and phase marks, plus separate DevTools traces with
screenshots. Retain these artifacts when comparing builds; this document's
summary tables do not replace the raw samples.

The behavior CLI checked 1,209 exposed views with no undeclared behavior. It
discovers routes and individual control options, parses detached documents and
uses a negative control. It does not prove every option combination or live
lifecycle path. Broader-than-needed declarations remain reported explicitly.
Build and render tests check ownership, dependency links, cache variants,
conversion behavior and CSS transformations. Use the linked lifecycle and
conformance guides for interaction, history and static-tool checks.

## Remaining limits

- Incoming HTML is still parsed and processed more than once. The measured warm
  navigation cost is mainly browser work on this local connection; changing
  prefetch alone does not remove it.
- Core scripts remain eager. Closed/hidden components can still download declared
  behavior. There is no general interaction prediction or resumability system.
- Shared runtime services still rely on globals. Loader readiness does not include
  every asynchronous component dependency or certify paint.
- Per-node lifecycle handling is component-specific. The full direct-load,
  repeated-init, history, cancellation, failure/retry and static-consumer matrix
  remains broader than the declaration audit.
- Response caches are local to a process. Cold renders, index construction and
  static compression still cost work; attachment generation still costs memory
  and CPU. No distributed response cache or cold-miss coalescing is implemented.
- Full custom-property shortening remains a benchmark experiment.
  Private-only delivery and class/ID shortening have been removed.
- Current measurements cover local Chromium, HTTP/1.1 and defined workloads.
  They do not establish Firefox/Safari behavior, HTTP/3 gains, remote-user latency
  or CI regression thresholds.

### Removed naming experiments

Class/ID shortening and private-only CSS delivery were removed after the
standards-mode comparison below. Class/ID shortening required HTML/CSS parsing
and would also need controller selectors, anchors and ID references rewritten
before live use. Private-only delivery preserved public and runtime-facing
properties but added separate assets, source scanning, fingerprints and startup
checks for a small payload reduction.

Their measured results remain below. The transforms, optional delivery route,
manifest fields, build flag and experiment-only dependencies are no longer part
of the implementation. Full custom-property shortening remains available in
`cmd/bench-css-names` as a frozen-page experiment.

### Standards-mode comparison and recommendation

The replay now restores the Catalog's HTML doctype, records `document.compatMode`,
and rejects non-standard rendering. This comparison uses Chrome
157.0.8103.38/Linux on an AMD Ryzen 5 5600X, 1440 × 900, and the same frozen
14-family page with 958 elements. Application controllers and animations are
removed; the theme/contrast controller remains unchanged. It measures naming
costs, not controller initialization, navigation or server rendering.

Each pair uses a fresh browser with a cleared browser cache and alternates which
version runs first. Normal CPU comparisons use 12 pairs per transform. The
synthetic 4× CPU slowdown uses 12 pairs for private properties and six pairs each
for all custom properties and classes/IDs. The slowdown is a sensitivity check,
not a measured mobile device. Requests use loopback HTTP/1.1 without network
throttling, and both versions serve precompressed fixtures.

For repeated measurements, reduce each page load to its median before comparing
loads. Paired differences compare the two versions within a round. Bootstrap
intervals resample those paired load differences, rather than treating repeated
measurements inside a browser as independent loads. They describe this local
sample; they do not establish cross-browser or production performance.

All versions start with 969,100 B of CSS and 113,777 B of rendered HTML.
Gzip payloads below are summed per file using the benchmark's Go gzip encoder.
The baseline is 109,199 B of CSS and 11,352 B of HTML; observed whole-page
transfer is 1,428,989 B, including the other unchanged resources.

| Transform | Raw CSS after transform | Gzip CSS after transform | Gzip HTML after transform | Combined gzip saving | Whole-page transfer reduction |
| --- | ---: | ---: | ---: | ---: | ---: |
| CSS-only private properties | 879,824 B | 104,933 B | 11,352 B | 4,266 B | 0.30% |
| All custom properties | 444,703 B | 85,163 B | 11,262 B | 24,126 B | 1.69% |
| Classes and IDs | 911,006 B | 105,869 B | 10,833 B | 3,849 B | 0.27% |

The private-property map saves 3,255 B on the current Cards and Carousel pages
and 3,641 B on the interactive Date Picker page. These counts compare the live
server's gzip-encoded component CSS responses with canonical responses. They
exclude unchanged stylesheets, fonts, images and HTML. None of these transforms
removes a request. A cached stylesheet has no transfer saving until it changes
or the cache is lost.

The following values are medians across page loads. Each arrow compares the
baseline with the transformed version in its own paired run; baselines can vary
between runs. Parsing creates detached `CSSStyleSheet` objects repeatedly. It
isolates parsing work and is not a direct measurement of parser time on the
initial critical path.

| Transform | CPU | First contentful paint | Isolated CSS parsing | Initial style calculation | Theme change, forced style/layout | Contrast change, forced style/layout |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| Private properties | Normal | 108 → 108 ms | 37.1 → 35.7 ms | 33.04 → 32.60 ms | 12.65 → 12.62 ms | 13.15 → 13.20 ms |
| All custom properties | Normal | 106 → 104 ms | 37.1 → 32.1 ms | 32.77 → 32.70 ms | 12.65 → 12.38 ms | 13.10 → 12.90 ms |
| Classes/IDs | Normal | 108 → 102 ms | 37.2 → 36.5 ms | 32.92 → 32.85 ms | 12.60 → 12.47 ms | 13.15 → 12.95 ms |
| Private properties | 4× slowdown | 344 → 348 ms | 156.3 → 150.8 ms | 131.15 → 130.97 ms | 54.22 → 53.42 ms | 55.73 → 55.33 ms |
| All custom properties | 4× slowdown | 344 → 330 ms | 156.3 → 133.5 ms | 129.61 → 128.13 ms | 54.50 → 52.58 ms | 55.22 → 54.75 ms |
| Classes/IDs | 4× slowdown | 342 → 340 ms | 155.6 → 153.2 ms | 131.90 → 128.55 ms | 56.10 → 53.65 ms | 55.95 → 55.13 ms |

Private-property parsing improved by a paired median of 1.30 ms normally and
5.25 ms with slowdown. Its first-paint paired difference was zero in both tests;
95% bootstrap intervals included improvements and regressions (−8 to +2 ms
normally, −6 to +8 ms with slowdown). The 4 ms increase between the slowed
first-paint medians is not evidence of a regression.

Full custom-property shortening improved parsing by a paired median of 5.10 ms
normally and 22.30 ms with slowdown. The corresponding 95% intervals were
−5.25 to −4.55 ms and −23.85 to −20.70 ms. Its slowed first-paint paired median
improved by 10 ms, but the interval reached zero (−26 to 0 ms). This supports a
parsing improvement and a possible paint benefit, not a reliable paint target.
Class/ID shortening improved normal first paint by a paired median of 6 ms;
the slowed result was uncertain (−14 to +8 ms). Theme/contrast changes showed
small differences and no consistent practical gain across conditions.

Computed styles and geometry matched in all six theme/contrast combinations for
every comparison. All six baseline/transformed default screenshot pairs were
pixel-identical. This establishes frozen appearance equivalence, not interactive
compatibility for renamed public properties, selectors or anchors.

A separate seven-pass private-delivery build probe, run after browser tests,
measured the source scan and transform without full token generation. Median
scan time was 27 ms for 335 files / 2,352,338 B; transformation of 29 stylesheets
took 268 ms, for 296 ms total. It produced 2,105,354 B of alternate CSS. These
are warm-filesystem, in-process step timings, excluding module startup, file
writes and other build stages. The server verifies and retains those alternate
bytes at startup. Request handling serves prebuilt bytes and uses the existing
gzip cache; it performs no source scan or renaming.

| Work | Recommendation | Reason |
| --- | --- | --- |
| Private-only production delivery | Removed | About 3–4 KB saved, with no clear paint or interaction benefit. Separate assets, source scanning, fingerprints, startup verification and rebuild dependencies are too much maintenance for this gain. |
| Full custom-property transform | Keep as an experiment worth pursuing | About 24 KB less gzip HTML/CSS and repeatable parsing gains. Live delivery still needs one correct mapping across CSS and all consumers while preserving canonical Forge exports. Do not enable the frozen transform on interactive pages. |
| Class/ID shortening | Removed | About 3.8 KB saved and small rendering differences do not justify rewriting controller selectors, public anchors and ID references. Data-attribute shortening is not measured here. |
| TypeScript migration | Keep | Its purpose is compile-time checking and explicit dependencies. These tests provide no reason to treat it as a browser-speed optimization or reverse it. |
| CLI benchmark and equivalence checks | Keep | They expose payload gains, timing uncertainty and visual changes without adding client assets or request-time processing. |

These recommendations concern name shortening and its delivery machinery. They
do not call for removing existing module splitting, response/compression caches,
CSS barriers or prefetch; this comparison did not rebenchmark those features.
The private-only delivery path and class/ID experiment have been removed. The
full custom-property transform, TypeScript work and benchmark checks remain.

Capture the baseline from a readable-name build, then reuse its `original.json`
for each transform. For example:

```sh
go run ./cmd/bench-css-names \
  --catalog http://127.0.0.1:8085 \
  --input /path/to/original.json --runs 12 \
  --chrome-path /path/to/chrome --output /tmp/custom-property-normal
```

This tests all custom properties. Add `--cpu-rate 4` for the slowdown comparison.
The removed experiments are recorded as historical results, not current CLI modes.
Run tests serially. Raw results retain each load, repeated measurements,
appearance fingerprints and screenshots; compare load-level medians and paired
rounds rather than pooling all repetitions as independent observations.

## First-paint and navigation priorities after removing naming paths

A November 6 exploratory profile used the readable-name review server, the same
Chrome 157 executable, 1440 × 900 and normal CPU. Five fresh-browser Cards loads
painted at 48, 48, 44, 44 and 44 ms: 44 ms median FCP. A minimal standards-mode
HTML page served locally painted at 32 ms median. These are different workloads
from the 14-family frozen page's 108 ms; none establishes a universal browser
floor. FCP measures the first content, not completion of a component or its
interactive behavior. See the [FCP definition](https://web.dev/articles/fcp).

In one Cards trace, the HTML response ended around 2.6 ms, component stylesheets
finished around 17 ms, and first content was presented around 48 ms. Before that
paint, traced style calculation took 6.9 ms and layout 9.8 ms. After it, the trace
recorded another 25.8 ms of style calculation, 12.7 ms of layout and 29.4 ms of
script evaluation. These event categories can nest and must not be added as
exclusive wall-clock phases. The page had 2,093 elements and 27 CSSOM stylesheets,
including inline styles. Its common component bundle alone was 252,292 B raw.

Three-run navigation probes found the following medians. Readiness is the current
`tokens:assets-ready` marker, not a guarantee that every asynchronous component
resource is complete.

| Journey | New DOM observed | Current loader readiness |
| --- | ---: | ---: |
| Cold destination assets | 46.3 ms | 84.5 ms |
| Warm destination assets | 38.8 ms | 70.2 ms |
| Prefetched destination | 38.5 ms | 70.2 ms |

Warm navigation reached `htmx:before:settle` at 41.8 ms,
`htmx:after:settle` at 59.7 ms, and `htmx:after:swap` at 63.5 ms. The loader began
initialization around 63.3 ms. The [HTMX 4 lifecycle](https://four.htmx.org/reference/methods/htmx-swap/)
places settle before `after:swap`; waiting for that event affects activation time,
even when CSS and scripts are cached. This identifies a scheduling experiment,
not a measured fix or a reason to remove settling everywhere.

Priorities:

1. Separate interaction binding from final-layout measurement. Test activating
   event-only behavior after insertion while keeping geometry-dependent work
   after settling. Check animation continuity, cancellation, repeated swaps and
   history restoration before changing the shared lifecycle.
2. Profile the post-paint style/layout work by initializer. Batch writes and
   measurements where traces identify repeated forced layout, and avoid doing
   measurements for closed controls until needed. Preserve immediate opening by
   preparing dependencies before the interaction when practical.
3. Revisit the scope of the common module and hidden-dialog requirements. Every
   `ui.Require` currently includes common, which contains several primitive
   families. Use renderer-owned dependency declarations to reduce unconditional
   delivery; avoid a second hand-maintained selector inventory or blindly
   dropping rules that an interactive state will need later.
4. Measure deployed latency and slower devices before more server micro-tuning.
   This local trace does not show server response time dominating. The local
   prefetch/warm difference is negligible; remote latency may change that result.

These are follow-up proposals. This cleanup removes the rejected naming paths;
it does not change activation timing, CSS module ownership or dialog delivery.

## Navigation and startup experiments, November 6

The vendored HTMX 4 sets `defaultSettleDelay` to **1 ms**. The observed gap
between insertion and controller initialization is not HTMX 2's former 20 ms
settle timer. It includes browser work and the settlement lifecycle. A trial
using `settle:0ms` for boosted document navigation did not give a consistent
improvement, so it was removed. Activation still uses the existing lifecycle;
interaction binding and geometry measurement have not been split.

Two changes remain:

- The `htmx:beforeProcessNode` link-policy hook now handles the processed anchor
  itself. It no longer rescans all descendant links whenever HTMX processes a
  container. The initial document scan remains.
- Site navigation preserves an already-open drawer when its native dialog mode
  is correct. The inline layout already opens the docked drawer. Closing and
  reopening it during startup repeated style and focus work. Switching between
  modal and docked modes still closes and reopens it where required.

Five fresh Cards loads per condition used Chrome 157, a 1440 × 900 viewport and
normal or fourfold CPU slowdown. The table gives medians in milliseconds; style,
layout and script counters cover startup, including work after first paint.

| Measurement | Before | After | Before, CPU ×4 | After, CPU ×4 |
| --- | ---: | ---: | ---: | ---: |
| FCP | 44 | 44 | 216 | 212 |
| Script duration | 17.3 | 17.2 | 70.5 | 69.4 |
| Style recalculation | 40.3 | 36.8 | 179.9 | 161.3 |
| Layout | 15.2 | 15.1 | 46.9 | 45.8 |
| Task duration | 110.0 | 107.7 | 451.1 | 428.4 |

The retained changes reduce startup style work. First paint is essentially
unchanged. Single traces also show less inclusive evaluation time for site
navigation, consistent with avoiding the drawer's forced style work. These are
small exploratory samples, not confidence bounds or evidence of a universal
paint improvement.

The navigation comparison ran three samples for each combination of cold,
warm, prefetched and expired-prefetch navigation, normal or slowed CPU, and
zero or 100 ms added latency: 48 samples per implementation, 144 total across
the baseline, zero-settle trial and retained changes. Readiness medians include:

| Condition and journey | Baseline | Zero-settle trial | Retained changes |
| --- | ---: | ---: | ---: |
| Local, warm | 74.0 | 68.2 | 68.0 |
| CPU ×4, warm | 309.7 | 308.7 | 314.8 |
| Added latency, warm | 251.7 | 252.8 | 252.7 |
| CPU ×4 and added latency, warm | 396.7 | 394.2 | 396.3 |
| Added latency, prefetched | 68.6 | 68.9 | 69.3 |
| Added latency, expired prefetch | 168.6 | 167.9 | 167.9 |

Navigation results are mixed; they do not support a general navigation speedup.
The latency comparison does support retaining the existing prefetch behavior.
CPU throttling approximates slower processing, not a particular device. Added
latency does not reproduce a deployed connection's bandwidth, packet loss or
server load. The readiness marker is still `tokens:assets-ready`, not a measure
of first successful user interaction or every asynchronous component resource.

### Common CSS boundary audit

The current common bundle is 252,292 B raw and 21,165 B with the server's Go gzip
settings. A Cards-only prototype omitted the generated field and menu-icon
bindings and structural text-field rules. It measured 184,947 B raw and
16,816 B gzip: 67,345 B less source but only **4,349 B less transfer**.

This prototype is not a valid replacement for common across all pages. Other
renderers and static consumers still need those rules; any split must retain
those dependencies and interactive states. Theme CSS savings and browser gains
were not measured. Common ownership stays unchanged because the measured saving
alone does not justify new modules and dependency changes. No rules were removed
from delivered CSS.

Regression coverage checks the per-node anchor policy and preserving the docked
drawer while retaining modal transitions. The web UI suite passed 234 tests;
component generation/type checks and Go tests for catalog web, Forge site and
components also passed. Real-page checks at 390, 900 and 1440 px exercised download
options, playground navigation, page navigation, browser history and mobile
navigation, and checked behavior declarations and uncaught exceptions. Default
screenshots matched the baseline at all three widths.

Computed-style and geometry fingerprints matched all six theme/contrast
combinations at mobile and wide widths. The first medium-width comparison had a
fingerprint mismatch despite identical screenshots. A subsequent property-by-
property comparison and repeat of all six combinations matched; its original
cause was not isolated. This check is evidence for unchanged appearance, not a
guarantee against every transient state.
