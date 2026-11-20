# Component ownership

Forge separates component props, markup, behavior, and token bindings so
primitives render independently of Catalog; hosts resolve each component's
declared asset requirements.

| Path | Edit here for |
| --- | --- |
| `ui/internal/model/` | Typed props, attribute rules, geometry, and asset declarations |
| `ui/templates/` | Component templ markup |
| `ui/internal/render/` | Generated Go renderers (do not edit directly) |
| `ui/api_generated.go` | Generated public facade for `forge/components/ui` |
| `behaviors/`, `helpers/` | Browser controllers and stateless shared helpers |
| `styles/`, `foundations/` | Component CSS anatomy and shared elevation rules |
| `data/` | Loading-indicator geometry |
| `generated/` | Built scripts, token-bound CSS, bundles, and notices |
| `modules.json` | Module ownership of styles, bindings, and runtime scripts |
| `style-sources.json` | Explicit CSS assembly order |
| `behavior-sources.json` | Aggregate script entries and dedicated output exceptions |

## Build

- Run `npm run generate` to compile templates and browser assets. Each
  controller in `behaviors/` must have one owner in `modules.json`; stateless
  helpers are bundled into classic-script entries by esbuild.
- Run `npm run build:tokens -- --published PATH` to regenerate token bindings
  and CSS bundles from a publication.
- Run `npm run check:components`, `npm run test:ui`, and `npm run test:tokens`
  to verify types, freshness, behaviors, and mappings.

Public Go consumers import `forge/components/ui`. Mark internal model helpers
needed by templates with `//forge:internal` so they are excluded from the
generated public facade. `assets.go` maps logical `static/...` paths to
`generated/` and `data/`.
