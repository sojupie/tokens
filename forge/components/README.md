# Component ownership

Forge separates component data, markup, behavior and token bindings. A primitive
can render without Catalog; the host resolves its declared asset requirements.

| Directory | Edit here for |
| --- | --- |
| `ui/internal/model/` | Typed props, attribute rules, geometry and asset declarations |
| `ui/templates/` | Component markup |
| `ui/internal/render/` | Generated Go only |
| `ui/api_generated.go` | Generated public facade; edit its model or template source |
| `behaviors/`, `helpers/` | Browser controllers and their pure helpers |
| `styles/`, `foundations/` | Component anatomy and shared elevation CSS |
| `data/` | Loading-indicator geometry |
| `generated/` | Built scripts, token-bound CSS, bundles and notices |
| `modules.json` | Ownership of component styles, bindings and runtime scripts |
| `style-sources.json` | CSS assembly order; changes can affect the cascade |
| `behavior-sources.json` | Aggregate entries and dedicated output exceptions |

Run `npm run generate` for templates and browser assets. Dedicated controllers are
found automatically in `behaviors/`; aggregate membership stays explicit. Each
controller must have one owner in `modules.json`. Pure imports are bundled into
classic-script entries, so the browser does not need a new module loading model.

Run `npm run build:tokens -- --published PATH` for token bindings. This requires
publication data and also refreshes component CSS bundles. `npm run check:components`
checks types and source/output freshness; `npm run test:ui` and `npm run test:tokens`
check browser logic and token mappings.

The public Go import remains `forge/components/ui`. Public component functions and
prop aliases are generated from source declarations. A `//forge:internal` directive
marks model helpers that templates need across the package boundary but consumers
should not receive as public functions.

`assets.go` maps the existing logical `static/...` paths to `generated/` and `data/`.
Consumers and exported bundles keep their existing paths. Generated files identify
their source; change source files and regenerate rather than patching output.
