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

## Next stages

- [ ] Audit script execution prerequisites, CSS measurements, initialization and
  cleanup entry points. Distinguish fetching from evaluation and initialization.
- [ ] Choose an enforceable prerequisite contract. Prefer dependencies expressed
  in source imports where practical; avoid maintaining a parallel hand-written
  graph of dependencies already represented in code.
- [ ] Carry behavior requirements in each rendered asset scope, including nested
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
