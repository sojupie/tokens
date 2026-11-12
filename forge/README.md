# Forge

Forge builds CSS from DTCG tokens and implements the browser behavior that tokens
cannot describe, including keyboard handling, ripples, and shape animation.

| Path | Responsibility |
| --- | --- |
| `build-tokens.mjs` | Style Dictionary build and catalog theme assembly |
| `material/` | Material Design 3 component token mappings and coverage checks |
| `components/` | Shared spring sampler, browser behavior and embedded generated styles |
| [`react/`](react/README.md) | React renderers that use the same Forge styles and browser controllers |
| `conformance/` | Browser checks against token-derived properties |

Run these commands from the repository root:

```sh
npm ci
npm run build:components
npm run build:tokens -- --published published
npm run test:tokens
npm run test:ui
```

Component source ownership:

| Path | Edit here for |
| --- | --- |
| `components/styles/` | CSS anatomy, layout, and interaction styling, separated by concern |
| `components/behaviors/` | Chips, tabs, navigation, sliders, toolbars, snackbars, FAB menus, split buttons, and tooltips |
| `components/style-sources.json`, `behavior-sources.json` | Assembly order; CSS order preserves the existing cascade |
| `components/generated/` | Generated served CSS, browser scripts, bundles and notices; edit their sources instead |
| `components/ui/internal/model/`, `components/ui/templates/` | Typed prop rules and component markup |
| `components/ui/internal/render/` | Generated component rendering Go |
| `components/foundations/`, `components/data/` | Shared elevation CSS and loading geometry |
| `theme-subsets.ts` | Theme dependencies, shared inputs and preview scopes |
| `component-bundles.ts`, `css-output.ts` | CSS partitioning, declaration grouping, theme deltas, and minification |
| `material/` | Explicit token bindings, source coverage, and value overrides |

`components/generated/components-base.css` and `components.js` are assembled files.
The [component ownership guide](components/README.md) maps every source and output.
Run `npm run build:components` after editing their sources, or
`npm run check:components` to check freshness. The token build also assembles them. It generates the individual `static/runtime/`
scripts from the same behavior sources.
Dedicated controllers, including `menus.js`, are authored in `behaviors/` and compiled to `generated/`.
Catalog's specimen markup lives in `catalog/web/templates/playground/specimen_<family>.templ`;
its example messages and workspace actions live in `catalog/web/assets/scripts/component-demos.js`.
Those examples consume behavior events rather than putting their content in Forge.

The build exports selected contexts from a saved publication, generates CSS, and
writes publication assets. It also updates `forge/components/generated/buttons.css`,
and `components.css`, plus the CSS bundles in `static/bundles/`, which the Go
server embeds. Rebuild the server after regenerating these files.
The build assembles Catalog's theme from its static stylesheet references and
configured color/motion scheme. Expressive site springs are selected separately
from the Standard-motion public downloads.
The theme defines shared values once and emits differences for the six appearance
contexts. `theme-subsets.ts` derives a shared theme and family themes from the
component CSS bundles. It follows CSS aliases across appearances and includes
literal JavaScript reads. Dynamic spring inputs are explicit. The build writes
these files under `<publication>/assets/themes/` and records their checksums in
the publication asset manifest. Public token downloads retain their complete selected context. Component
CSS groups compatible bindings and removes whitespace before partitioning shared
controls and optional families. Catalog selects the page stylesheet and loads
optional behavior through `catalog/web/assets/generated/asset-loader.js`. The complete
`components.css` and `components.js` assets remain available.

Production builds run in a separate asset-builder image before server startup
and after accepted updates. Assets stay outside approved generations; see
[deployment](../docs/operations/deployment.md#generated-publication-assets).

Forge covers action controls, chips and selection controls, cards and lists,
fields and pickers, menus, navigation, search, sheets, dialogs, toolbars,
loading and progress indicators, and supporting surfaces. Coverage varies by
family; the contracts describe implemented states and remaining limits. See [generated components](../docs/reference/components.md),
[buttons](../docs/reference/buttons.md), and [DTCG exports](../docs/reference/dtcg.md)
for supported variants, checks, and exclusions.

Another design system can use the DTCG-to-CSS build. Its components need explicit
mappings for anatomy, states, and behavior. The Material generator must reject
unknown token properties rather than silently choosing a fallback.
