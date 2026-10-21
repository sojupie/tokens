# Forge

Forge builds CSS from DTCG tokens and implements the browser behavior that tokens
cannot describe, including keyboard handling, ripples, and shape animation.

| Path | Responsibility |
| --- | --- |
| `build-tokens.mjs` | Style Dictionary build and catalog theme assembly |
| `material/` | Material Design 3 component token mappings and coverage checks |
| `components/` | Shared browser behavior and embedded generated styles |
| `conformance/` | Browser checks against token-derived properties |

Run these commands from the repository root:

```sh
npm ci
npm run build:tokens -- --published published
npm run test:tokens
npm run test:ui
```

The build exports selected contexts from a saved publication, generates CSS, and
writes publication assets. It also updates `forge/components/static/buttons.css`,
and `components.css`, which the Go server embeds. Rebuild the server after regenerating that file.
The build currently assembles Catalog's theme by reading its static stylesheets.
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

Earlier Token Forge plans are preserved in
[the historical notes](../docs/history/token-forge/README.md).
