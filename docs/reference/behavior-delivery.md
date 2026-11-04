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
- [x] Start known optional script downloads together while preserving execution
  order and waiting for CSS before controller evaluation.
- [ ] Test direct loads, repeated initialization, nested compositions, HTMX
  swaps, history restoration, superseded navigation and failed asset retries.
- [ ] Verify exported implementations and build-time tool consumers against the
  same contract.
- [ ] Measure cold and warm navigation with the actual delivery protocol before
  changing bundle boundaries or hidden-region activation policy.

Closed dialogs, visible regions and offscreen regions share ownership
declarations. Dialog presentation now creates listeners and observers on opening
and releases them on close. This defers instance work, not the script download.
The ownership manifest does not decide when behavior should activate.

The Catalog loader adds native `preload` links for scripts discovered in the
current document. These links start downloads without evaluating the scripts.
After component CSS resolves, the loader evaluates scripts in its existing order
and checks the DOM again before each script. This preserves requirements added
by earlier controllers. Loaded scripts are reused across swaps. Failed scripts
release their pending entry and preload link so a later activation can retry.
Only the latest activation can update preview controls or report an asset error.
Browsers that ignore preload links retain the ordered loading path.

Run `npm run check:components`, `npm run test:tokens`, `npm run test:ui`, and
`go test ./forge/components ./catalog/web ./cmd/tokens` for this initial stage.

## Prerequisite and lifecycle audit

The table distinguishes execution prerequisites from node initialization and
cleanup. The loader does not consume it as an execution graph.

| Scripts | Prerequisites and lifecycle |
| --- | --- |
| `motion.js`, `buttons.js`, `runtime/utils.js` | Motion publishes `MaterialMotion`; buttons reads it during evaluation. Utilities retain the compatibility `MaterialComponents` export; current Forge consumers import their helpers instead. Motion remains a host prerequisite. |
| `runtime/fab-menus.js`, `runtime/split-buttons.js` | Include imported keyboard and positioning helpers and register delegated handlers. Layout measurements occur on interaction. |
| `runtime/button-groups.js`, `runtime/segmented-buttons.js` | Register delegated handlers and include the imported navigation helper in their own compiled files. They do not need the global helper provider or a per-node initialization pass. |
| `runtime/chips.js`, `runtime/tabs.js`, `runtime/switches.js`, `runtime/app-bars.js` | Register delegated handlers. Tabs use motion during selection. They do not need a per-node initialization pass. |
| `runtime/navigation.js`, `runtime/sliders.js`, `runtime/toolbars.js`, `runtime/snackbars.js` | Include imported helpers. Initialize per-node state and/or observers, retry on asset readiness and swaps, and clean up removed nodes. Navigation also has a direct history-restoration hook. |
| `runtime/tooltips.js`, `fields.js`, `menus.js` | Initialize per-node state and measure geometry during interaction. Cleanup hooks exist. Tooltips retry on asset readiness; fields and menus are core scripts with direct history-restoration hooks. Menu motion needs `MaterialMotion`. |
| `sheets.js`, `lists.js`, `carousel.js` | Maintain per-node or gesture state. Sheets and lists use motion; lists include the imported keyboard helper. Carousel initializes measured layout and observers. All have swap/asset-ready and cleanup hooks. |
| `search.js`, `date-picker.js`, `time-picker.js` | Initialize per-node state and clean up removed instances. Measurements can occur during initialization or opening. Pickers use motion; search has a direct history-restoration hook. Other restore paths currently rely on the loader's asset-ready event. |
| `dialogs.js` | Delegates opening and closing. Binds per-dialog listeners and observers on opening; releases them on close. `MaterialDialogs.initialize(root)` adopts open markup without resetting existing instances; `dispose(root)` releases removed instances. Existing HTMX, history and asset hooks use these methods. Motion remains a prerequisite for animations. |
| `progress.js`, `loading.js` | Manage observers and animation loops. Loading fetches geometry asynchronously; script `load` completion does not mean that geometry has resolved. Preserve that distinction when defining readiness. |
| `artwork-canvas.js`, `artwork.js` | Canvas publishes `TokenArtworkCanvas`. Canvas artwork needs that provider before initializing. Artwork also owns visibility, resize and animation work. |
| `component-demos.js`, `dialog-demo.js` | Catalog owns example actions. Component examples use motion; dialog examples invoke `MaterialDialogs` and attach per-panel listeners. They are separate from the Forge controllers. |

The current safe boundary is still the existing host script order and CSS-before-
initialization barrier. Parallel fetching must not imply parallel evaluation or
skip initialization after styles resolve. An enforceable source-import contract
is now used for keyboard navigation, slider values, toolbar spacing, split-menu
positioning and snackbar timing. Each entry bundles only its imported helpers. Runtime services still use
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
