# TokenCatalog performance inventory

Benchmarks through November 10, 2026, at Catalog revision `71811ed`. The
inventory follows Git history from source capture through browser delivery.
The [delivery reference](../reference/delivery-performance.md) and [development guide](../development.md)
document the current architecture and benchmark commands.

The biggest measured gains came from avoiding work. The crawler can finish
before its deadline, publication builds can reuse history, repeated requests
can reuse rendered responses, and the download dialog can confirm a selection
without converting tokens. Images reduce DOM and style work on the Components
index but can increase desktop transfer. CSS shortening reliably saves bytes.
Its effect on first paint varies by page.

## Method

- **Historical A/B** uses the indicated revisions with the same inputs.
  **Isolated A/B** changes one mechanism or runs historical code in a fixture.
  **Prior measurement** carries forward a result from the existing references.
- New tests ran on Linux, Ryzen 5 5600X, Go 1.27.3 and, where applicable,
  headless Chrome 156. The serving tests used the local `button-color-review`
  publication. Cold browser loads disabled the HTTP cache. The Catalog process
  could remain warm between loads. CPU ×4 means Chrome's fourfold throttling.
- Times are medians unless stated otherwise. Gzip totals sum independently
  compressed response bodies. Resource Timing transfer sizes include Chrome's
  header accounting. FCP means first contentful paint. Loader readiness means
  assets are ready. Several tests overlapped on the workstation, so small
  timing differences are noisy.

## Capture, publication, and deployment

| Change | Comparison | Result |
| --- | --- | --- |
| End capture when payload and network activity settle (`503bf65`) | Historical A/B, two runs per arm of the commit's Chromium fixture, with equal deadlines. | Fast page, one payload, **8.03 → 3.03 s**. Delayed page, two payloads, **8.02 → 5.78 s**. Hanging unrelated iframe, one payload, **8.03 → 3.03 s**. Empty page, zero payloads, **1.02 → 1.02 s**. |
| Stream graph verification (`855c183`) | Historical A/B, three runs per arm on the October 11 snapshot. | **6.37 → 4.05 s** and **2.69 → 0.76 GiB peak RSS**. Both verified the snapshot. |
| Reuse verified graph and history inputs in one publish (`c0db284`) | Historical A/B, two builds per arm with the August 12 and October 11 snapshots. | **38.48 → 24.12 s** and **3.85 → 3.72 GiB peak RSS**. Both built 39 subjects and 6,729 change records. |
| Cache history indexes and semantic comparisons across builds (`8ef8900`) | Same commit, one empty-cache build and two populated-cache builds, using the same two snapshots. | **24.96 → 15.86 s** and **3.72 → 2.69 GiB peak RSS**. Logs confirmed reuse of both an index and a comparison. The cache used about **207 MiB**. |
| Load SQLite into an in-memory publication (`47699aa`) | Current-path probe. Warm prepared SQL query returning 554 Buttons names versus looking up the loaded subject and reading its token count. | Load **1.00 s**, peak **188 MiB**, 42 subjects. SQL query **0.20 ms**, map lookup **18 ns** average. This shows work moved to startup. The two probes do different amounts of work and do not measure a full request. |
| Reuse Chromium for offline thumbnail capture (`94343f0`) | Six real captures in three batches per arm. The old allocation pattern failed on capture two with the current browser stack. A fresh browser per capture provides a working baseline. | **2.40 → 1.12 s** from fresh per capture to one shared browser. PNG hashes matched. Time excludes image encoding. |
| Reuse Go compile cache across image builds (`ab0d1b6`, `d393286`) | Isolated Go equivalent of the BuildKit cache mount. Change embedded CSS while keeping dependencies fixed. Two runs per arm, module cache populated. | `go build -trimpath` **16.19 → 3.14 s**. Chromium package-layer reuse and the larger Docker cache budget have no isolated deployment timing. |

Atomic activation and generation pinning (`37a04d5`) are consistency changes.
Building and retaining illustrations before deployment (`f1a25c7`) leaves
**zero screenshot or encode jobs per visitor request**. Activation still
validates and syncs the generation.

## Server responses

### Caching and repeated work

| Change | Result |
| --- | --- |
| Rendered-response LRU (`d9059bd`) | Twelve uncached Cards requests, then the same URL through boosted HTMX. Loopback **12.01 → 0.20 ms**, handler **11.74 ms → below the 0.001 ms reporting resolution**. Both sent **30.75 kB gzip**. Revalidation returned **304, zero body bytes, 0.16 ms**. |
| Immutable content-versioned URLs (`1d6370a`) | Chrome requested Common CSS twice. A matching version transferred **21,816 → 0 B**. A stale version transferred **21,816 → 300 B** with revalidation. |
| Gzip and one-time static encoding (`2916f33`) | Common CSS **253,281 B identity → 21,516 B gzip**. The Go helper took **1.02 ms** to encode versus **79 ns** to retrieve memoized bytes. This excludes hashing and transfer. |
| Cache preview coverage and selected thumbnail cases (`0490d9c`, `0f55e0c`) | Forced decode versus cached lookup. Buttons coverage **0.97 ms and 2.08 MB allocated → 11 ns and zero allocations**. General coverage **15.75 ms and 36.97 MB → 11 ns and zero allocations**. Case selection and rendering still take work. |
| Resolve script ownership from one parsed manifest (`264651e`) | Isolated Go test resolving the same module list. Decode and resolve **24.5 µs, 72 allocations** versus resolve with cached manifest **1.98 µs, 9 allocations**. The script URL inventory is also built once. |
| Queue and batch usage writes (`4787613`, `53ca0db`) | Seven runs per arm, 256 events each. Synchronous per-event upserts took **9.87 ms** on the caller versus **0.011 ms** to enqueue. Including close and flush, **10.09 → 4.97 ms**. All 256 events persisted in each arm. |

### Bounded responses and export work

Twelve uncached requests per arm used the same URL and query with or without
the relevant HTMX headers. Times include loopback and body receipt.

| Change | Full → fragment wall time | Full → fragment gzip |
| --- | ---: | ---: |
| Components search for `button` (`126c8b3`) | 6.38 → 1.49 ms | 23.67 → 5.00 kB |
| Buttons playground, hovered state (`126c8b3`) | 31.21 → 10.03 ms | 37.08 → 2.71 kB |
| Buttons all-token table filtered to `container` (`f1070bb`) | 125.95 → 60.75 ms | 59.15 → 39.36 kB |

The public `view=table` route (`5355faf`) makes a reduced representation directly
addressable. For the default Buttons set, the full page versus table-only
response was **37,048 → 26,442 B gzip** for HTML and **2,897 → 1,361 B gzip**
for Markdown.

Historical change-history grouping (`1738753`) reduced the first `/changes`
response from **113,834 → 16,160 raw B** on the same publication
(**4.88 → 4.70 ms** over 12 warm requests per arm). Expanding App Bars fetched
50 records, **109,750 B**, in **2.45 ms**. The later summary page (`cda83f3`)
replaced inline expansion with links to paginated subject history.

Export option checks, historical A/B (`03f568c` → `0f55e0c`), used the same
publication, one valid Web/Light DTCG selection, and 12 cache-busting requests
per arm. The old path needed options and summary requests in sequence.
The new options response includes confirmation.

| Selection | Before, two requests | After, one request | Current, one request |
| --- | ---: | ---: | ---: |
| Buttons | 750 ms | 31.7 ms | 30.8 ms |
| Whole collection | 1,221 ms | 128.8 ms | 141.8 ms |

The whole-collection responses grew from **3.95 → 5.21 kB gzip** at the change.
The timing includes both less server work and one fewer round trip. The
isolation test below separates those effects.

#### Isolating the three export changes

Using `0f55e0c` as the common base, the benchmark restored the parent's
context lookup and summary conversion independently. A test-only request
header made the options handler return before confirming the selection,
so the client could call the summary endpoint separately. All eight arms
used the same templates, the same publication, and seven uncached samples
after one warmup. All confirmed the same counts, 554 Buttons tokens and
6,418 collection tokens.

| Context lookup | Summary | Buttons, two → one request | Collection, two → one request |
| --- | --- | ---: | ---: |
| Scan | Convert | 707 → 598 ms | 1,143 → 886 ms |
| Scan | Count | 269 → 164 ms | 693 → 447 ms |
| Index | Convert | 475 → 455 ms | 618 → 556 ms |
| Index | Count | 45 → 29 ms | 179 → 121 ms |

Starting from the shipped combination, reverting the index alone raised
collection confirmation from **121 → 447 ms**. Restoring conversion raised it
to **556 ms**. Splitting the requests raised it to **179 ms**. The effects depend
on the other settings. For example, the separate summary handler resolves the
options again, so indexing also changes the cost of the two-request path.
The measured split cost includes duplicate server work and loopback transfer.
It excludes the delay of an additional round trip over a remote connection.

For batch exports (`d9059bd`), an isolated Buttons DTCG-plus-report test took
**908 ms and 407 MB allocated** when explicitly reconverting for the report
versus **471 ms and 207 MB** when reusing the converted document. This is two
runs per arm, one job, excluding ZIP construction. Different jobs can still
convert separately. Loading indicators and stale-action guards cover feedback
and correctness while this work runs.

## Browser startup and navigation

| Change | Result |
| --- | --- |
| Inline theme initialization and defer control behavior (`850bb39`) | Historical scripts in a fixture with a 100 ms external script delay, eight loads per arm. FCP **120 → 32 ms** with saved dark theme in both arms. The same commit replaced the 250 px logo **99,022 B PNG → 6,926 B WebP**. |
| Hoist module CSS, wait before swapping, and retain split styles and shell DOM (`2916f33`, `2ee6b92`, `126c8b3`) | Actual warm Buttons → Carousel → Buttons journey. **16 of 16** original component stylesheet links remained attached, two were added, and there were **zero** split stylesheet refetches. The site header retained DOM identity. |
| Fetch optional scripts in parallel, evaluate in order (`adabb12`) | Historical loaders in a fixture with 120 ms CSS and three 90 ms script responses, eight loads per arm. Readiness **397 → 122 ms**, with the same three script requests and execution order. |
| Declare required behaviors in rendered regions (`264651e`, `d4a7a62`, `d9059bd`) | Isolated selector discovery versus reading declarations on real pages, 12 blocks of 100 calls per arm. Components **0.130 → 0.007 ms** per call, Buttons **0.428 → 0.018 ms**, Carousel **0.340 → 0.014 ms**. The methods return different lists because declarations also cover nested dependencies. This excludes fetching and activation. The earlier asset split reduced `/changes` script requests **22 → 12**. |
| Bundle stateless helpers at build time (`ed1b5fc`, `5096a1c`) | Fourteen changed runtime files total **19,380 → 20,655 B gzip**. Existing script entry points remain. Bundling duplicates some helper bytes and needs no new helper requests. |
| Bind dialog presentation on open and release it on close (`369e8a7`) | Fixture with 64 closed dialogs, eight loads per arm. Startup **64 → 0 ResizeObservers** and **192 → 0 direct listeners**. Opening one creates one observer, and closing releases it. Script-load-plus-setup **1.7 → 1.3 ms**. |
| Prefetch on hover and focus (`c9808f8`, `e974ffa`, `7faec2c`, `cae985e`) | Prior local warm vs reused prefetch **70.2 → 70.1 ms** on Cards, **82.8 → 81.2 ms** on Date Picker. With 100 ms added latency in a later run, Cards readiness was **252.7 ms warm**, **69.3 ms prefetched**, **167.9 ms with expired prefetch**. This measures the policy as a whole. |
| Skip repeated HTMX anchor scans and preserve docked drawer state (`d9059bd`) | Prior Cards run, five loads per arm. FCP **44 → 44 ms**, style recalculation **40.3 → 36.8 ms**. CPU ×4 FCP **216 → 212 ms**, style recalculation **179.9 → 161.3 ms**. Warm readiness varied. |
| Render date-picker row count before initialization (`d7889c1`) | Computed-size check on the real docked picker. Six-row fallback **456 px**. Rendering the row count removes a potential **96 px** correction for four rows and **48 px** for five. This measures panel height. |
| Skip offscreen token-table layout with `content-visibility` (`0490d9c`) | Buttons all-token page at 390 px, seven loads per arm. Layout **7.27 → 6.26 ms**, CPU ×4 **39.24 → 33.34 ms**. FCP **52 → 52 ms** and **260 → 260 ms**. The table starts about 2,007 px down the page. DOM count stays at 801 in its wrapper. |
| Canvas hero and idle/visibility gates (`244b570`, `fc692cb`) | Prior Chromium 155 profile. Active renderer CPU **16.1–19.1% → 2.81% of one core**. Both had frame interval p95 about **20.9 ms**. Canvas draw submission p95 **0.3 ms**, with no draws in paused or offscreen samples. |

### Isolating the startup changes

The `d9059bd` startup changes were reversed independently on `71811ed`.
Each arm used the same Cards page, assets, and 1440 × 900 viewport. Chrome
ran eight cold loads per arm and CPU setting in shuffled order. The server
was warm. Style recalculation includes work through the asset barrier and
the following 250 ms.

| Link processing | Docked drawer | FCP, normal / CPU ×4 | Style recalculation, normal / CPU ×4 |
| --- | --- | ---: | ---: |
| Scan descendants | Close and reopen | 44 / 222 ms | 40.26 / 183.01 ms |
| Scan descendants | Preserve | 44 / 232 ms | 37.26 / 168.90 ms |
| Process anchors only | Close and reopen | 44 / 228 ms | 40.72 / 181.96 ms |
| Process anchors only | Preserve | 44 / 224 ms | 36.91 / 166.40 ms |

Diagnostic loads confirmed that the anchor change removed 16 descendant
queries in this startup. Preserving the drawer removed one close and one
reopen. The drawer change saved about **3–4 ms of style recalculation**
at normal speed and **14–16 ms** at CPU ×4 regardless of the link setting.
The link change had little effect on the Cards timings. FCP remained
at 44 ms at normal speed and varied at CPU ×4.

HTTP/3 is enabled at the production edge and the Caddy dev config (`bcccad7`)
supports local QUIC testing. Seven fresh curl processes per protocol requested
the same 99 B production health response. Median connection time was
**10.2 ms HTTP/2** and **9.3 ms HTTP/3**. TTFB was **29.8 ms HTTP/2** and
**38.3 ms HTTP/3**. Both returned 200. This spot check excludes multiplexing,
packet loss, and page navigation.

## Images and gallery work

`05c890f` replaced live gallery previews with images or a generic placeholder.
`f1070bb` added responsive images and inline blur placeholders. `f4f28a5` fixed
responsive image URLs under the deployment prefix.

The isolated A/B used the current Components page and the same code with the
historical inert preview fallback restored. There were 36 pictures or 36 live
specimens. Each arm had seven cold loads per viewport, observed through 500 ms
after load. The local publication still contains the older 24 × 15 JPEG
placeholders.

| Width, DPR, CPU | FCP, live → images | Style calculation, live → images | JPEG requests | JPEG payload |
| --- | ---: | ---: | ---: | ---: |
| 1440 px, 1×, normal | 192 → 48 ms | 141.6 → 45.6 ms | 32 | 178.0 kB |
| 390 px, 1×, normal | 56 → 40 ms | 123.1 → 36.8 ms | 6 | 35.0 kB |
| 390 px, 2×, CPU ×4 | 348 → 164 ms | 595.6 → 188.7 ms | 6 | 104.3 kB |

The grid went from **1,269 → 469 elements**, CSS **152.6 → 82.5 kB gzip**,
JavaScript **167.6 → 126.2 kB gzip**, and HTML **36.1 → 40.5 kB gzip**.
Counting these payloads and the loaded JPEGs, desktop went from **356 → 427 kB**.
At 390 px and DPR 1, it went from **356 → 284 kB**.

Across 252 local image/appearance pairs, the median 720 px JPEG is **16.59 kB**
and the 360 px JPEG is **5.94 kB**. Chrome chose 360 px at DPR 1 and 720 px
at DPR 2 in these samples. The Components page embeds six placeholders per
card. In the local fixture, the JPEG was **630 B median** before base64,
and the 216 declarations added **12.66 kB** to the gzipped HTML.

The current production page served 216 WebP placeholders at 40 × 25 px,
**116 B median** before base64. Recompressing that HTML with Go gzip before
and after stripping the declarations gives **49.18 → 27.87 kB**. The WebP
files are smaller individually, while the older JPEGs share more repeated
bytes that gzip can reuse across a page. A same-encoder size probe with the
production WebPs saved **2.48 kB gzip** when baking in a blur scaled to
a 360 px card. This probe re-encodes already compressed images. It does
not match image quality or measure blur CPU cost or time to a sharp preview.

## CSS delivery and naming

| Change | Comparison and result |
| --- | --- |
| Component modules and on-demand delivery (`9655bd9`, `2916f33`, `86f525e`) | Prior October Chrome run on `/changes`. CSS **3.00 → 1.24 MB raw**, **220 → 101 kB gzip**. Synchronous appearance style/layout **49.5 → 42.6 ms** on `/changes` and **51.0 → 42.2 ms** on the rail playground. |
| Shared and scoped theme inputs (`b3ba36c`) | Prior same-document A/B, 32 samples per arm. Root variables **4,115 → 2,017**, plus 153 scoped to the rail. Forced style/layout **17.3 → 10.2 ms** on `/changes` and **15.7 → 10.2 ms** on the rail. Two-frame medians **68.3 → 51.1 ms** and **68.4 → 51.2 ms**. |
| Appearance deltas (`b3ba36c`, `d9059bd`) | Repeated inherited declarations in each of six shared-theme blocks. Expanded versus delta output **1,055,336 → 155,943 B raw**, **74,957 → 11,760 B gzip**. Both keep the base block. |
| Preserve compatible token aliases (`9655bd9`, `b3ba36c`) | Flattening 9,408 simple aliases in that expanded theme used **66,225 B gzip** versus **74,957 B** with aliases. Aliases preserve token relationships and override behavior. The byte comparison favors literals. |
| Group binding selectors and compact whitespace (`9655bd9`, `d9059bd`) | Isolated output probes on the aggregate bindings. Ungrouped versus grouped **2,352 → 2,201 rules**, **81,581 → 80,912 B gzip**. Formatted versus compact **82,389 → 80,912 B gzip**. |
| Drop empty rules from delivery CSS (`bcccad7`) | Isolated on the original-name pack at `71811ed`. Removed **154 rules in 27 stylesheets**, saving **14,317 B raw** and **1,504 B gzip** independently of renaming. |
| Shorten release custom-property names (`623e78b`, `bcccad7`) | Same 61 stylesheets with original, previous, and current transforms. **3,540,301 → 2,994,260 → 1,684,361 B raw**, **358,438 → 329,800 → 277,773 B gzip**. Current saves **22.5% gzip** against original. Downloads use the original assets. |

A direct live-page A/B used the same revision, publication, and generated inputs.
Six cold Chrome loads per arm, page and CPU setting. CSS bytes are the encoded
bodies observed by the browser. Style time includes the 500 ms after load.

| Page | CSS gzip, original → short | FCP, normal | FCP, CPU ×4 | Style calculation, CPU ×4 |
| --- | ---: | ---: | ---: | ---: |
| Cards | 87.03 → 70.19 kB | 76 → 104 ms | 232 → 232 ms | 179.2 → 165.5 ms |
| Date Pickers | 92.48 → 74.20 kB | 48 → 44 ms | 210 → 196 ms | 324.7 → 298.4 ms |
| Carousel | 88.02 → 71.36 kB | 52 → 50 ms | 212 → 212 ms | 411.9 → 377.5 ms |

The earlier frozen mixed-component experiment measured **37.7 → 32.1 ms**
isolated CSS parsing and **109.2 → 85.2 kB gzip** for full property rewriting.
That transform changed more names than the release transform, which protects
names used by runtime code. The live-page test shows a clear transfer reduction
and mixed first-paint results.

### Isolating the release CSS changes

The `bcccad7` transform was run with all eight combinations of test-source
exclusion, exact-name protection, and empty-rule pruning. All arms used the
same 61 inputs from `71811ed`, including the later filter margin edit.
Every output file in the fully enabled arm matched the release-profile build's
hash.

| Sources scanned | Protection for names found in source | Names shortened | Gzip with empty rules | Gzip after pruning |
| --- | --- | ---: | ---: | ---: |
| Include tests | Treat every name as a prefix | 1,252 | 329,789 B | 328,264 B |
| Include tests | Exact names and dynamic prefixes | 1,265 | 328,172 B | 326,667 B |
| Exclude tests | Treat every name as a prefix | 5,300 | 281,445 B | 279,760 B |
| Exclude tests | Exact names and dynamic prefixes | 5,312 | 279,459 B | 277,756 B |

With the other release settings fixed, skipping tests accounts for
**48.91 kB** of gzip savings. Narrowing protection to exact names where
possible accounts for **2.00 kB**, and pruning empty rules for **1.70 kB**.

## Experiments that stayed out of the release

| Experiment | Evidence and decision |
| --- | --- |
| Class and ID shortening | The frozen 14-component trial saved **3.85 kB gzip**. In the broader class-only test, classes were **11.32% of CSS characters** and **13.43% of HTML characters**. Shortening saved **10.63 kB gzip across the CSS pack**, **367.05 kB across 261 pages**, and **1.26 kB on the median page**. The class-only paint trial's paired intervals crossed zero. Keep existing names. |
| Private properties only | Frozen comparison saved **4.27 kB gzip** with no clear first-paint change. The release uses the broader CSS-only mapping. |
| Split more out of Common | Cards prototype saved **4.35 kB gzip**, with **67.35 kB** raw reduction. Other consumers still need the omitted rules. Common ownership stays as it was. |
| Zero HTMX settle delay | Prior warm readiness **74.0 → 68.2 ms** locally, **309.7 → 308.7 ms** at CPU ×4, and **251.7 → 252.8 ms** with added latency. The default was already 1 ms. The trial was removed. |

## Reproduction

- Build historical revisions in detached worktrees. Use `./cmd/tokens` in
  later revisions and `./cmd/sourceloom` for the earlier publication and
  history pairs. Use the same snapshot directory and separate output directories.
  For history caching, publish once with an empty `--history-cache` directory,
  then reuse it. Record wall time and RSS with `/usr/bin/time -f '%e %M'`.
- For the crawler, use `TestObservePageBrowser` at `503bf65`. Run the same
  fixture against its parent after removing the early-exit assertion.
- For HTTP checks, request gzip. Vary an ignored query parameter for cache
  misses and reuse the exact URL for hits. Send the returned ETag as
  `If-None-Match` for revalidation. For fragments set `HX-Request: true` and
  target `component-preview`, `subject-list`, or `tokens-token-browser`.
- Use `cmd/bench-navigation` and `cmd/bench-css-names` for navigation and frozen
  CSS experiments. For release CSS, serve matching assets with and without the
  `delivery-css` pack. Confirm the `css` profile parameter in served links.
  Alternate arms, disable Chrome's HTTP cache, and record FCP and CSS encoded
  body size separately.
