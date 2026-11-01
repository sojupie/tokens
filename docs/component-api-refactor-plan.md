# Catalog component API refactor

Status: implementation complete; device and extended manual checks remain open below.
Baseline: `5e2fb3b`. Update the checklist as work completes. Record the relevant
commit and checks under each completed phase; keep implementation evidence in
its existing location rather than creating another assessment system.

## Outcome and ownership

Every Material component rendered by a Catalog playground must use the Forge
Go/templ API. Catalog chooses its data, configuration and demonstrated states.
It must not maintain another implementation of component anatomy or behavior.
Apply that boundary to the Catalog interface, application scenes and capture
fixtures too. Plain HTML remains appropriate for application content and layout.

| Concern | Owner and extension point |
| --- | --- |
| Component anatomy, variants, geometry, accessibility, lifecycle and asset requirements | Forge API and its controllers |
| Example content, captions, comparison cases and deliberately forced states | Catalog adapters calling the API |
| Playground environment: viewport, backdrop, scale, containment, comparison arrangement, appearance and controls | Separate Catalog environment templates and styles |
| Application actions: routing, saving, filtering, exports, request cancellation and feedback | Catalog application code using component events and native controls |
| Consumer styling and application hooks | Existing `Attrs` / `InputAttrs`; root `class` merges with component classes. Add slot attributes or content slots only for an actual caller requirement. |

The TSX `className` analogy fits `Attrs["class"]` in templ. Keep that extension
point rather than introducing an untyped residual-style configuration or a
parallel theme system. Support root `style` through the existing typed templ
attribute conventions when needed. Prefer normal component props for shared
variants. Consumer styles cover placement and specific presentation; recurring
component fixes belong in Forge.

Separate the playground environment from the specimen in source. Use clearly
named environment templates and `tokens-playground-*` styles for new wrappers.
Mark the environment/specimen boundary in the DOM where useful for inspection;
retain existing IDs and loader hooks during migration. Do not add visible labels
to the page just to explain code ownership. Environment controls can still use
Forge components: their configuration logic and layout belong to Catalog.

Forced hover, focus, pressed and disabled examples remain supported. Pass these
through a clearly identified playground adapter to the API's existing state
hooks. Keep their styles in Forge. Normal API callers must work without
`componentCase`, demo attributes or Catalog scripts.

## Sequence and checklist

### 1. Establish the migration inventory and delivery baseline

- [x] Inventory every maintained Catalog template, Go HTML builder and browser-created component. Classify component markup, content slots and environment markup separately.
- [x] Include playground specimens, site UI, shared shell consumers, six application scenes, older capture fixtures, thumbnails and implementation exports. Exclude generated `*_templ.go` files from the authoring inventory.
- [x] Assign each component occurrence to an existing API or a concrete missing capability. Record the remaining occurrences here as migrations proceed; an import of `ui` does not establish complete adoption.
- [x] Record module names, stylesheet/script URLs, raw and gzip asset bytes, and response variants for representative pages and fragments. Count shared assets once when comparing a navigation session.
- [x] Capture current form names/values, IDs, ARIA references, application hooks and important event behavior before changing their markup.

Representative cases: landing page; component index with image and DOM fallback;
buttons; lists with a nested switch and reorder sheet; navigation rail with FAB;
search; date/time picker; download and table-context dialogs; each application
scene. Include direct requests, HTMX component-only updates and boosted navigation.

### 2. Make the APIs sufficient for real callers

- [x] Fix chip-owned behavior markers and verify selection/removal without Catalog attributes.
- [x] Remove sample decisions from generic fields: third-option divider, fixed option icon and hard-coded validation copy. Let callers supply those where required.
- [x] Consolidate `Select`, `Field`, Catalog selects and chip-style selectors around shared field/listbox anatomy. Preserve label/value separation, native form values, names, IDs and change events.
- [x] Complete list row/reveal/reorder/expansion and navigation container/destination/rail-control APIs where the playground still owns component anatomy. Keep example data and application destinations outside Forge.
- [x] Add only the missing content/attribute slots needed by downloads, context dialogs and table controls. Preserve custom form ownership, dynamic icons, menu links and button loading state.
- [x] Align public component hooks with runtime discovery. App bars, snackbars, chips and other API callers must not depend on playground wrappers to load or initialize their behavior.
- [x] Check attribute merging and semantic references. Consumer classes must preserve base classes; explicit playground overrides must remain separate from normal typed configuration.
- [x] Check standalone API rendering and interaction with nested dependencies and no Catalog imports or sample-state attributes.

### 3. Preserve asset delivery before moving callers

- [x] Keep `ui.WithAssets` attached to the existing root, nested region and export collectors. Each nested component registers its own module.
- [x] Keep deterministic module ordering, common assets, navigation lifetime and per-component theme/style subsets.
- [x] Keep deferred styles for thumbnail fallback and blocking styles for interactive detail components.
- [x] Keep lazy script loading, load-once behavior, ordering, retry handling and initialization after swaps/history restoration. Discover production component hooks as well as compatibility hooks during migration.
- [x] Update token-reference discovery to cover shared API sources where needed. Do not rely on Catalog's former copies to retain required tokens.
- [x] Keep preview readiness, inert/error handling, image fallback, blur and measured scale behavior. Do not generate illustration images during this refactor.
- [x] Run nested-component asset checks through full pages, HTMX fragments and standalone implementation downloads.

### 4. Finish playground adoption and separate its environment

- [x] Migrate remaining list and navigation structure, utility controls and nested sheets to the completed APIs.
- [x] Audit every other specimen, including its example actions and nested components, for direct component anatomy. Replace it with API calls.
- [x] Extract environment wrappers and controls from specimen composition. Keep captions, size constraints, simulated window, backdrop and scaling with the environment.
- [x] Keep `componentCase` and family adapters as Catalog configuration code; remove duplicated component structure and attribute construction after their callers migrate.
- [x] Preserve all supported comparison modes, forced states, query parameters, meaningful configurations and interactive behavior.
- [x] Verify the simple thumbnail and implementation-export render paths still consume these adapters without bringing along playground environment markup or controls.

### 5. Migrate Catalog download and context interfaces

- [x] Replace header split button, download menu and footer action with shared APIs. Preserve labels and preset selections.
- [x] Replace download and table-context dialog surfaces, actions, help tooltips and controls with shared APIs.
- [x] Replace radio and checkbox-chip anatomy with shared native form-capable components. Keep multi-select rules and platform-dependent contexts in Catalog.
- [x] Preserve disclosure animation and icon-only expansion behavior, status/error regions, scroll behavior, busy button transition and focus return.
- [x] Preserve export request cancellation, stale-response protection, selected scopes/contexts, filenames, ZIP contents and required token references.

### 6. Migrate remaining Catalog controls

- [x] Migrate outlined selects, filter pills, token-set selectors and their options using the consolidated APIs.
- [x] Migrate the four table icon controls and tooltips. Preserve mobile visibility, same-row search layout, dynamic icons/labels and existing tooltip behavior.
- [x] Migrate reference-popover close/copy controls, including Go-built and browser-created markup. Prefer one server-rendered reusable template where practical; do not add a second JavaScript component library.
- [x] Migrate index search, hero/page actions, ordinary Material cards, appearance controls and remaining site controls.
- [x] Preserve custom component/tool tiles, their image compositions, keyboard interaction and layout. Do not replace them with generic cards merely to increase API adoption.
- [x] Preserve shared shell behavior for Catalog and tools, base paths, URL handling, native links and existing SEO copy/metadata.

### 7. Finish scenes, capture fixtures and implementation exports

- [x] Audit all six application scenes, shared scene helpers and document-screen controls. Replace remaining component internals and actions with APIs; retain application panes, content and state.
- [x] Audit older scene fixtures and simple thumbnails. Keep their compositions and configurations; remove duplicated component implementations.
- [x] Verify implementation exports contain API-rendered specimen markup and the matching dependency closure. Exclude Catalog environment, sample validation policy and unrelated application scripts.
- [x] Keep existing generated images and manifests unchanged. Image regeneration and scene redesign remain separate work.

### 8. Verify, remove compatibility duplication and finish

- [x] Re-run the occurrence inventory. Every remaining direct component class/anatomy has a named reason: shared API internals, supplied content slot, custom site design or playground environment. Do not count typography classes alone as component duplication.
- [x] Remove superseded helpers/styles and unnecessary explicit asset declarations only after verifying the replacement collects the same required assets.
- [x] Compare delivery against phase 1. Account for any asset growth by the components actually rendered; do not hide missing dependencies to achieve a smaller bundle. Preserve lazy boundaries rather than switching to global component bundles.
- [x] Preserve bounded response caching, variant keys, ETags/304, gzip, versioned immutable URLs, decoded coverage reuse, thumbnail configuration cache and connection-aware prefetch.
- [x] Run scoped Go tests, UI tests, component assembly checks, token-build tests and affected generated theme builds. Re-run guidance validation only when curated checks change; keep source guidance and implementation assessments separate.
- [ ] Complete the extended manual matrix: real iPadOS/Pencil, touch scrolling, reduced motion, throttled loading and failed optional assets. Desktop browser checks completed for selection, keyboard menus, HTMX, history, export/context modals, scene widths and appearance; automated delivery and archive checks pass.
- [x] Update the API/architecture docs to reflect final ownership and remaining intentional custom surfaces. Give a concrete completion summary with checks and limits.
- [x] Commit coherent concern-based changes as phases complete. Leave the local review server available. Do not push or deploy without a later request.

## Completion rule

The refactor is complete when Catalog owns composition, data and environment;
Forge owns each shared component implementation; and the verified delivery paths
retain their optimizations. API imports, successful compilation and image capture
alone do not establish completion or visual fidelity.

## Apple Pencil menu investigation

- [ ] Investigate iPadOS Apple Pencil menu taps: hover/pressed styling and ripple appear and the menu closes, but selection and the primary click/HTMX action do not run. Finger and mouse taps work; Pencil works on buttons outside menus.
- [ ] Trace pen pointerdown/up/cancel, compatibility click, focusout, popover light-dismiss and menu cleanup ordering across selects, action menus and split-button menus.
- [x] Preserve native activation and HTMX delivery; avoid synthetic double activation. Cover pen, touch scroll/cancel, mouse and keyboard paths.
- [ ] Verify with real iPadOS/Pencil when available; simulated pointer checks alone cannot establish Safari correctness.

## Approval

Approved by the user on 2026-11-01. Push and production deployment remain outside this request.


## Implementation evidence · 2026-11-01

- `2517372`: shared native fields, form chips, dialog/form slots, menu links,
  compound card/list/navigation/sheet APIs, runtime discovery and pointer checks.
- `e597bc3`: Catalog controls, playground specimens, nested content, thumbnails
  and scenes call those APIs. `playground_environment.templ` owns simulated
  viewports and arrangement. DOM boundaries use `data-playground-environment`
  and `data-playground-specimens`.
- `8d9caa4`: corrected malformed chip/group conditionals found during browser
  review. `6158ad9`: token discovery includes authored shared API sources.
- `e013325`: native chip styling is self-contained and field labels follow their
  input IDs. `04742f9`: menu/badge consumer classes merge with the base classes.

### Intentional remaining markup

| Surface | Reason |
| --- | --- |
| Component/tool tiles | Custom site design; imagery and keyboard behavior retained |
| `md-*-demo` wrappers in playground environment | Existing containment, scroll and gesture compatibility hooks |
| Card typography and application content inside API slots | Consumer content; component roots/media/body/action slots come from Forge |
| Native change-history `details` and code `pre` blocks | Native content elements; shared card appearance supplied through props |
| Table reference SVG symbols | One shared sprite; API-rendered close/copy controls are cloned from a server template |
| Range input controlling search composition width | Exterior playground environment control |
| Review page width/appearance selectors | Local review harness controls |
| Explicit asset declarations in compatibility adapters | Retained where attribute-level state adapters still require their dependency |

No literal button or dialog element remains in maintained Catalog templ sources.
Old fixture compositions retain the same 31 scene branches. Shared UI assets
remain collected per rendered region, including nested controls and downloads.

### Checks and limits

Scoped Go suites pass for Catalog, Forge, CLI, Source, schemas and skills. UI
tests: 185 passed. Token-build tests: 47 passed, 4 skipped. Component assembly
check and the six-appearance token build pass. The Motion Lab static-web consumer
build passes with `CGO_ENABLED=0` and `-tags staticweb`; its native test build
requires unavailable OpenCV pkg-config files.

Real-page checks covered header component preset, multi-format/multi-theme
exports and busy recovery, context save, HTMX Compare changes, keyboard token-set
navigation, history restoration, list selection and nested reorder actions.
All six application frames loaded; scene checks included 1440px, 840px and the
412px museum composition, plus dark/high contrast. These checks preserve the
existing scenes; they do not constitute a new scene-design acceptance.

The Pencil investigation identified a focusout/release hazard in field options
and mouse-only submenu hover assumptions. Pen contact now shares the guarded
release/cancel path with touch; its following compatibility click is deduplicated,
including zero-detail clicks. Only mouse hover changes submenus. Native input
and change events remain the select contract; generic command clicks are not
synthesized. Automated pointer checks pass. The exact iPadOS action-menu failure
has not been reproduced on hardware, so it remains open above.

Existing illustration files were preserved; no image captures were generated.
The local review service is available at `http://100.103.168.9:8085`. No push or
production deployment was performed.

### Delivery comparison

Compared with `5e2fb3b` on the same publication. Counts below are gzip estimates
using identical local compression settings, not measured network transfer sizes.
Version hashes were ignored when comparing module membership.

| Response | Before | After | Change |
| --- | ---: | ---: | ---: |
| Landing HTML | 12,670 B | 12,802 B | +132 B |
| Button page HTML | 26,153 B | 26,610 B | +457 B |
| Interactive list HTML | 19,277 B | 19,724 B | +447 B |
| Search page HTML | 25,337 B | 25,782 B | +445 B |
| Time-picker page HTML | 31,954 B | 32,445 B | +491 B |
| List HTMX fragment | 2,814 B | 2,865 B | +51 B |

All nine sampled full pages retained their stylesheet/script module membership.
Across the distinct stylesheets collected by these pages, raw CSS grew by 270 B;
the combined gzip estimate grew by 59 B. Shared control/semantic attributes account
for modest HTML growth. No global component bundle was substituted. Actual gzip
responses and ETag conditional requests were also checked; the latter returned
304. Response caching, immutable URL handling and prefetch code were unchanged.
All 1,585 existing illustration files match their baseline hashes.
