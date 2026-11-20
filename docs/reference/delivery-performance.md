# Delivery and on-page performance

How Token Catalog handles request processing, CSS/script splitting, browser
navigation, and component behavior lifecycle. For historical benchmark numbers,
see [November 2026 performance benchmarks](../history/performance-benchmarks.md).

## Request and navigation flow

1. The server loads one immutable publication and its verified generated assets
   at startup, then closes SQLite. Ordinary browsing never queries SQLite.
2. Cacheable GET/HEAD requests hit an in-memory LRU or render Go/templ HTML from
   memory.
3. Full documents hoist shared CSS and rendered component modules into `<head>`.
4. Hover or keyboard focus on eligible internal links prefetches destination HTML
   via HTMX's preload extension.
5. On navigation, rendered declarations start optional script preloads and fetch
   missing component CSS before HTMX swaps `#tokens-site-page`. The navigation
   shell stays mounted.
6. Optional controllers evaluate in manifest order after CSS resolves and emit
   `tokens:assets-ready`.

There is no client-side hydration pass; downloads, external links, and
non-document API routes stay native.

## Backend caching and assets

### Startup and response cache
- `publish.Load` reads the publication into `PublishedState` once at startup and
  verifies generated theme, CSS value index, and illustration manifests.
- Decoded preview configs, representative thumbnail cases, module order, and
  script URLs are initialized once.
- An in-memory LRU caches successful cacheable GET/HEAD responses (up to 256
  entries / 32 MiB of keys and bodies), keyed by URL, HTMX request/target,
  accepted content type, and encoding. Boosted document requests share the
  full-document entry.
- HTML responses use `public, no-cache` with representation-specific ETags
  (returning `304` on revalidation). Errors, `no-store`, `Set-Cookie`, range
  requests, static assets, unscoped downloads, health, usage, and MCP bypass the
  cache.
- `Server-Timing` reports cache lookup or handler render time.

### Static transfer and compression
- Asset URLs include a SHA-256 content version and get
  `public, max-age=31536000, immutable`. Bare or stale URLs revalidate.
- Gzipped CSS, JS, and SVG are cached in process by content hash; HTML is
  gzipped on cache miss and cached with separate ETags per encoding.

### Download dialog and usage recording
- The download dialog builds an in-memory export index on first use (token
  names, context labels, references, and convertibility) so option and summary
  checks don't convert documents or build ZIPs.
- Actual DTCG/CSS conversion and ZIP assembly run on form submit (capped at 64
  combinations, cancellation-aware, reusing each converted document for its
  conversion report).
- Usage recording enqueues events into a 4,096-entry buffer flushed to SQLite in
  batches (256 events or 1 s) so telemetry never blocks responses.

## CSS delivery and production name shortening

- `forge/components/modules.json` maps every structural stylesheet, generated
  binding family, and behavior script to one module. Templates register modules
  beside their markup with `@useComponentAssets(...)`, including nested controls.
- The theme build emits shared declarations once and only changed values for the
  other five appearances. `forge/theme-subsets.ts` splits shared root variables
  from per-module variables scoped to `[data-component-assets]` (or
  `#component-preview` in standalone fixtures).
- **Production CSS delivery:** `deploy/build-assets.sh` passes `--css-names short`
  by default (set `TOKENS_CSS_NAMES=original` to disable; local `build:tokens`
  defaults to original). `forge/css-delivery.ts` assigns short custom-property
  names across site, component, and theme CSS while protecting names and
  prefixes referenced by JavaScript or templates, and strips empty rules.
  Class/ID names are never rewritten. DTCG, CSS, and implementation downloads
  always keep original names.
- Profiled stylesheet URLs include both the content hash and dictionary ID;
  version mismatches across deployments trigger a full document reload.

## Browser delivery, prefetch, and lifecycle

### Styles and scripts
- Full pages hoist and deduplicate component stylesheet links into `<head>`.
  Fragment responses declare dependencies so `asset-loader.js` can fetch missing
  styles in manifest order before swapping. Loaded styles stay mounted for the
  document lifetime.
- Each asset scope emits a `script[type="application/json"]` block with
  `data-component-behaviors`. The loader adds `<link rel="preload">` for known
  scripts to overlap CSS downloads, then evaluates them in manifest order once
  CSS resolves.
- Stateless helpers in `forge/components/helpers/` are bundled into classic
  script entries by esbuild. Shared stateful services (`MaterialMotion`) stay on
  host globals.
- Playgrounds stay inert while loading and show a reload prompt on failure;
  failed loads clear their pending entries so retries work.

### Hover prefetch and persistent shell
- Same-origin document links prefetch after 65 ms of hover or keyboard focus
  (capped at 8 destinations per document; skipped on `Save-Data`, 2G, hidden
  tabs, downloads, and external links).
- HTMX's preload extension holds each prefetched response for 5 seconds.
  Destination scripts start preloading only when navigation consumes the HTML.
- Rail and drawer navigation morph in place via partial updates while keeping
  section scroll positions. Subject search debounces at 250 ms with
  `hx-sync="this:replace"`.

### Static tool consumers
Material Motion Lab renders shared Go/templ markup at build time and packages
matching Forge assets; `tokens site-shell` exports the same shell HTML and
assets for the static Carousel playground. Neither uses Catalog's HTMX loader.

## Behavior prerequisites and lifecycle

| Scripts | Prerequisites and lifecycle |
| --- | --- |
| `motion.js`, `buttons.js`, `runtime/utils.js` | `motion.js` publishes `MaterialMotion` (read by `buttons.js`). `utils.js` keeps compatibility `MaterialComponents` exports. |
| `runtime/fab-menus.js`, `runtime/split-buttons.js` | Bundle keyboard/positioning helpers; delegated listeners; measure on interaction. |
| `runtime/button-groups.js`, `runtime/segmented-buttons.js` | Bundle navigation helper; delegated listeners; no per-node init. |
| `runtime/chips.js`, `runtime/tabs.js`, `runtime/switches.js`, `runtime/app-bars.js` | Delegated listeners (tabs use `MaterialMotion`); no per-node init. |
| `runtime/navigation.js`, `runtime/sliders.js`, `runtime/toolbars.js`, `runtime/snackbars.js` | Bundle helpers; per-node init/observers, asset-ready/swap retry, and cleanup. Navigation has a history-restore hook. |
| `runtime/tooltips.js`, `fields.js`, `menus.js` | Per-node state, geometry on interaction, and cleanup. Fields and menus are core scripts with history-restore hooks; menus need `MaterialMotion`. |
| `sheets.js`, `lists.js`, `carousel.js` | Per-node/gesture state, swap/asset-ready and cleanup hooks. Sheets and lists use `MaterialMotion`. |
| `search.js`, `date-picker.js`, `time-picker.js` | Per-node state and cleanup. Pickers use `MaterialMotion`; search has a history-restore hook. |
| `dialogs.js` | Delegated open/close; binds listeners/observers on open and releases on close (`MaterialDialogs.initialize` / `dispose`). Needs `MaterialMotion`. |
| `progress.js`, `loading.js` | Manage observers and animation loops. `loading.js` fetches geometry asynchronously after script load. |
| `artwork-canvas.js`, `artwork.js` | `artwork-canvas.js` publishes `TokenArtworkCanvas`, required before `artwork.js` initializes. |
| `component-demos.js`, `dialog-demo.js` | Catalog-owned example actions (separate from Forge controllers). |

## Checking behavior declarations

`TokenAssets.audit(root)` compares rendered `data-component-behaviors`
declarations against diagnostic selector discovery. Run the CLI check against a
running Catalog:

```sh
go run ./cmd/check-behaviors --catalog http://127.0.0.1:8085
```

Use `--chrome-path PATH` to select Chromium and `--path /components/search` to
scope to one route. Missing declarations fail the check; extra declarations
(such as broad `common` or `messages` module entries) are reported.
