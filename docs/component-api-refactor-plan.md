# Catalog component API refactor

Status: approved; implementation in progress.
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

- [ ] Inventory every maintained Catalog template, Go HTML builder and browser-created component. Classify component markup, content slots and environment markup separately.
- [ ] Include playground specimens, site UI, shared shell consumers, six application scenes, older capture fixtures, thumbnails and implementation exports. Exclude generated `*_templ.go` files from the authoring inventory.
- [ ] Assign each component occurrence to an existing API or a concrete missing capability. Record the remaining occurrences here as migrations proceed; an import of `ui` does not establish complete adoption.
- [ ] Record module names, stylesheet/script URLs, raw and gzip asset bytes, and response variants for representative pages and fragments. Count shared assets once when comparing a navigation session.
- [ ] Capture current form names/values, IDs, ARIA references, application hooks and important event behavior before changing their markup.

Representative cases: landing page; component index with image and DOM fallback;
buttons; lists with a nested switch and reorder sheet; navigation rail with FAB;
search; date/time picker; download and table-context dialogs; each application
scene. Include direct requests, HTMX component-only updates and boosted navigation.

### 2. Make the APIs sufficient for real callers

- [ ] Fix chip-owned behavior markers and verify selection/removal without Catalog attributes.
- [ ] Remove sample decisions from generic fields: third-option divider, fixed option icon and hard-coded validation copy. Let callers supply those where required.
- [ ] Consolidate `Select`, `Field`, Catalog selects and chip-style selectors around shared field/listbox anatomy. Preserve label/value separation, native form values, names, IDs and change events.
- [ ] Complete list row/reveal/reorder/expansion and navigation container/destination/rail-control APIs where the playground still owns component anatomy. Keep example data and application destinations outside Forge.
- [ ] Add only the missing content/attribute slots needed by downloads, context dialogs and table controls. Preserve custom form ownership, dynamic icons, menu links and button loading state.
- [ ] Align public component hooks with runtime discovery. App bars, snackbars, chips and other API callers must not depend on playground wrappers to load or initialize their behavior.
- [ ] Check attribute merging and semantic references. Consumer classes must preserve base classes; explicit playground overrides must remain separate from normal typed configuration.
- [ ] Check standalone API rendering and interaction with nested dependencies and no Catalog imports or sample-state attributes.

### 3. Preserve asset delivery before moving callers

- [ ] Keep `ui.WithAssets` attached to the existing root, nested region and export collectors. Each nested component registers its own module.
- [ ] Keep deterministic module ordering, common assets, navigation lifetime and per-component theme/style subsets.
- [ ] Keep deferred styles for thumbnail fallback and blocking styles for interactive detail components.
- [ ] Keep lazy script loading, load-once behavior, ordering, retry handling and initialization after swaps/history restoration. Discover production component hooks as well as compatibility hooks during migration.
- [ ] Update token-reference discovery to cover shared API sources where needed. Do not rely on Catalog's former copies to retain required tokens.
- [ ] Keep preview readiness, inert/error handling, image fallback, blur and measured scale behavior. Do not generate illustration images during this refactor.
- [ ] Run nested-component asset checks through full pages, HTMX fragments and standalone implementation downloads.

### 4. Finish playground adoption and separate its environment

- [ ] Migrate remaining list and navigation structure, utility controls and nested sheets to the completed APIs.
- [ ] Audit every other specimen, including its example actions and nested components, for direct component anatomy. Replace it with API calls.
- [ ] Extract environment wrappers and controls from specimen composition. Keep captions, size constraints, simulated window, backdrop and scaling with the environment.
- [ ] Keep `componentCase` and family adapters as Catalog configuration code; remove duplicated component structure and attribute construction after their callers migrate.
- [ ] Preserve all supported comparison modes, forced states, query parameters, meaningful configurations and interactive behavior.
- [ ] Verify the simple thumbnail and implementation-export render paths still consume these adapters without bringing along playground environment markup or controls.

### 5. Migrate Catalog download and context interfaces

- [ ] Replace header split button, download menu and footer action with shared APIs. Preserve labels and preset selections.
- [ ] Replace download and table-context dialog surfaces, actions, help tooltips and controls with shared APIs.
- [ ] Replace radio and checkbox-chip anatomy with shared native form-capable components. Keep multi-select rules and platform-dependent contexts in Catalog.
- [ ] Preserve disclosure animation and icon-only expansion behavior, status/error regions, scroll behavior, busy button transition and focus return.
- [ ] Preserve export request cancellation, stale-response protection, selected scopes/contexts, filenames, ZIP contents and required token references.

### 6. Migrate remaining Catalog controls

- [ ] Migrate outlined selects, filter pills, token-set selectors and their options using the consolidated APIs.
- [ ] Migrate the four table icon controls and tooltips. Preserve mobile visibility, same-row search layout, dynamic icons/labels and existing tooltip behavior.
- [ ] Migrate reference-popover close/copy controls, including Go-built and browser-created markup. Prefer one server-rendered reusable template where practical; do not add a second JavaScript component library.
- [ ] Migrate index search, hero/page actions, ordinary Material cards, appearance controls and remaining site controls.
- [ ] Preserve custom component/tool tiles, their image compositions, keyboard interaction and layout. Do not replace them with generic cards merely to increase API adoption.
- [ ] Preserve shared shell behavior for Catalog and tools, base paths, URL handling, native links and existing SEO copy/metadata.

### 7. Finish scenes, capture fixtures and implementation exports

- [ ] Audit all six application scenes, shared scene helpers and document-screen controls. Replace remaining component internals and actions with APIs; retain application panes, content and state.
- [ ] Audit older scene fixtures and simple thumbnails. Keep their compositions and configurations; remove duplicated component implementations.
- [ ] Verify implementation exports contain API-rendered specimen markup and the matching dependency closure. Exclude Catalog environment, sample validation policy and unrelated application scripts.
- [ ] Keep existing generated images and manifests unchanged. Image regeneration and scene redesign remain separate work.

### 8. Verify, remove compatibility duplication and finish

- [ ] Re-run the occurrence inventory. Every remaining direct component class/anatomy has a named reason: shared API internals, supplied content slot, custom site design or playground environment. Do not count typography classes alone as component duplication.
- [ ] Remove superseded helpers/styles and unnecessary explicit asset declarations only after verifying the replacement collects the same required assets.
- [ ] Compare delivery against phase 1. Account for any asset growth by the components actually rendered; do not hide missing dependencies to achieve a smaller bundle. Preserve lazy boundaries rather than switching to global component bundles.
- [ ] Preserve bounded response caching, variant keys, ETags/304, gzip, versioned immutable URLs, decoded coverage reuse, thumbnail configuration cache and connection-aware prefetch.
- [ ] Run scoped Go tests, UI tests, component assembly checks, token-build tests and affected generated theme builds. Re-run guidance validation only when curated checks change; keep source guidance and implementation assessments separate.
- [ ] Verify real component pages and scenes at relevant widths, keyboard/touch use, reduced motion, theme/contrast changes, back navigation, throttled loading and failed optional assets. Check download correctness as well as UI behavior.
- [ ] Update the API/architecture docs to reflect final ownership and remaining intentional custom surfaces. Give a concrete completion summary with checks and limits.
- [ ] Commit coherent concern-based changes as phases complete. Leave the local review server available. Do not push or deploy without a later request.

## Completion rule

The refactor is complete when Catalog owns composition, data and environment;
Forge owns each shared component implementation; and the verified delivery paths
retain their optimizations. API imports, successful compilation and image capture
alone do not establish completion or visual fidelity.

## Apple Pencil menu investigation

- [ ] Investigate iPadOS Apple Pencil menu taps: hover/pressed styling and ripple appear and the menu closes, but selection and the primary click/HTMX action do not run. Finger and mouse taps work; Pencil works on buttons outside menus.
- [ ] Trace pen pointerdown/up/cancel, compatibility click, focusout, popover light-dismiss and menu cleanup ordering across selects, action menus and split-button menus.
- [ ] Preserve native activation and HTMX delivery; avoid synthetic double activation. Cover pen, touch scroll/cancel, mouse and keyboard paths.
- [ ] Verify with real iPadOS/Pencil when available; simulated pointer checks alone cannot establish Safari correctness.

## Approval

Approved by the user on 2026-11-01. Push and production deployment remain outside this request.

