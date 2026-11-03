# Component behavior delivery

This work preserves incremental asset delivery. Changes are staged so dependency
ownership, scheduling and component lifecycle can be verified separately.

## Initial work

- [x] Resolve Forge script ownership from `forge/components/modules.json` through
  one Go resolver used by Catalog delivery and implementation exports.
- [x] Validate authored runtime outputs and standalone component scripts against
  the manifest during component builds and checks. Reject unowned, multiply
  owned and missing scripts.
- [x] Check that manifest scripts are embedded and have versioned delivery URLs.
- [x] Preserve the browser's selector discovery, download sequence and activation
  behavior during this step.

The inventory comes from `behavior-sources.json` and standalone JavaScript files
in `forge/components/static/`. The shared `buttons.js` runtime and legacy
`components.js` aggregate are excluded explicitly. Catalog example scripts remain
Catalog-owned; they are not dependencies of exported Forge components.

`ScriptsFor` preserves the supplied module order. It resolves ownership, not a
prerequisite graph. Existing scripts are IIFEs that use globals; their global
dependencies cannot be reliably inferred from filenames or DOM selectors.

## Source imports

Behavior entry points can import stateless functions from `forge/components/helpers/`.
The component build uses esbuild to resolve and bundle those imports into the
existing classic-script files. Missing modules and exports fail compilation.
Imports add no browser requests and do not require another runtime dependency
manifest. The legacy aggregate is assembled from the same compiled entries.

Button groups and segmented controls import `navigationIndex` directly. Their
delivered scripts no longer require `MaterialComponents` for keyboard navigation.
The utilities entry imports the same function and retains its existing global
and guarded Node exports for other consumers.

This approach is limited to stateless functions. Independently bundling a module
that registers listeners or owns shared mutable state would create separate
instances. Those services need a shared runtime contract before conversion.
Layout measurements still require CSS readiness; source imports do not determine
when a DOM-dependent controller should initialize.

## Next stages

- [x] Audit script execution prerequisites, CSS measurements, initialization and
  cleanup entry points. Distinguish fetching from evaluation and initialization.
- [ ] Choose an enforceable prerequisite contract. Prefer dependencies expressed
  in source imports where practical; avoid maintaining a parallel hand-written
  graph of dependencies already represented in code.
- [x] Carry diagnostic behavior requirements in each rendered asset scope, including nested
  components, Catalog examples and independently replaced fragments. Preserve
  the current eager activation policy initially.
- [ ] Compare those declarations against current browser discovery across the
  generated component configurations before switching the loader's input.
- [ ] Fetch independent requirements concurrently while preserving necessary
  evaluation order and waiting for CSS before geometry-dependent initialization.
- [ ] Test direct loads, repeated initialization, nested compositions, HTMX
  swaps, history restoration, superseded navigation and failed asset retries.
- [ ] Verify exported implementations and build-time tool consumers against the
  same contract.
- [ ] Measure cold and warm navigation with the actual delivery protocol before
  changing bundle boundaries or hidden-region activation policy.

Closed dialogs, visible regions and offscreen regions currently share ownership
declarations. Deferring a closed dialog requires a separate activation contract;
the ownership manifest does not decide when its behavior should become ready.

Run `npm run check:components`, `npm run test:tokens`, `npm run test:ui`, and
`go test ./forge/components ./catalog/web ./cmd/tokens` for this initial stage.

## Prerequisite and lifecycle audit

The table distinguishes execution prerequisites from node initialization and
cleanup. The loader does not consume it as an execution graph.

| Scripts | Prerequisites and lifecycle |
| --- | --- |
| `motion.js`, `buttons.js`, `runtime/utils.js` | Motion publishes `MaterialMotion`; buttons reads it during evaluation. Utilities publish `MaterialComponents`. These are host prerequisites today. |
| `runtime/fab-menus.js`, `runtime/split-buttons.js` | Read utilities during evaluation and register delegated handlers. Layout measurements occur on interaction. |
| `runtime/button-groups.js`, `runtime/segmented-buttons.js` | Register delegated handlers and include the imported navigation helper in their own compiled files. They do not need the global helper provider or a per-node initialization pass. |
| `runtime/chips.js`, `runtime/tabs.js`, `runtime/switches.js`, `runtime/app-bars.js` | Register delegated handlers. Tabs use motion during selection. They do not need a per-node initialization pass. |
| `runtime/navigation.js`, `runtime/sliders.js`, `runtime/toolbars.js`, `runtime/snackbars.js` | Read utilities during evaluation. Initialize per-node state and/or observers, retry on asset readiness and swaps, and clean up removed nodes. Navigation also has a direct history-restoration hook. |
| `runtime/tooltips.js`, `fields.js`, `menus.js` | Initialize per-node state and measure geometry during interaction. Cleanup hooks exist. Tooltips retry on asset readiness; fields and menus are core scripts with direct history-restoration hooks. Menu motion needs `MaterialMotion`. |
| `sheets.js`, `lists.js`, `carousel.js` | Maintain per-node or gesture state. Sheets and lists use motion; lists also read utilities during keyboard interaction. Carousel initializes measured layout and observers. All have swap/asset-ready and cleanup hooks. |
| `search.js`, `date-picker.js`, `time-picker.js`, `dialogs.js` | Initialize per-node state and clean up removed instances. Measurements can occur during initialization or opening. Pickers and dialogs use motion; search and dialogs have direct history-restoration hooks. Other restore paths currently rely on the loader's asset-ready event. |
| `progress.js`, `loading.js` | Manage observers and animation loops. Loading fetches geometry asynchronously; script `load` completion does not mean that geometry has resolved. Preserve that distinction when defining readiness. |
| `artwork-canvas.js`, `artwork.js` | Canvas publishes `TokenArtworkCanvas`. Canvas artwork needs that provider before initializing. Artwork also owns visibility, resize and animation work. |
| `component-demos.js`, `dialog-demo.js` | Catalog owns example actions. Component examples use motion; dialog examples invoke `MaterialDialogs` and attach per-panel listeners. They are separate from the Forge controllers. |

The current safe boundary is still the existing host script order and CSS-before-
initialization barrier. Parallel fetching must not imply parallel evaluation or
skip initialization after styles resolve. An enforceable source-import contract
is now used for the stateless navigation helper. Runtime services still use
globals; converting them requires preserving standalone exports, the legacy
aggregate and tool consumers.

## Diagnostic comparison

Each asset scope now emits a non-executable `script[type="application/json"]`
with `data-component-behaviors`. Forge requirements come from rendered primitives
and the ownership manifest. Catalog example requirements are recorded where their
markup is rendered. The normal selector-based loader remains authoritative.

`TokenAssets.audit(root)` compares declarations with the same discovery predicate
used by that loader. It excludes already-delivered host scripts and reports
`undeclared` and `declarationOnly` separately. It does not load or initialize the
declared scripts. Inert-thumbnail behavior and the existing progress exception
are preserved.

The initial browser comparison checked 1,026 views with no undeclared behavior.
Declaration-only scripts were `runtime/chips.js` from the common module, `runtime/snackbars.js`
from the messages module, and `progress.js` on loading-indicator views. These are
reported rather than silently removed. This covers exposed individual options,
not every generated configuration combination, live lifecycle behavior or the
decision to defer a hidden region.
