# Forge

Forge compiles DTCG tokens into CSS and implements browser component behavior
(keyboard handling, ripples, motion, and popovers).

| Path | Responsibility |
| --- | --- |
| `build-tokens.mjs` | Style Dictionary build, site theme assembly, and asset manifest |
| `material/` | Material Design 3 token mappings, overrides, and coverage checks |
| [`components/`](components/README.md) | Go/templ API, CSS anatomy, behaviors, and generated assets |
| [`react/`](react/README.md) | Standalone React 19 renderers sharing Forge CSS and controllers |
| [`conformance/`](conformance/README.md) | Local browser parity and regression fixtures |

## Commands

Run from the repository root:

```sh
npm ci
npm run build:components
npm run build:tokens -- --published published
npm run check:components
npm run test:tokens
npm run test:ui
```

- `npm run build:components` assembles CSS anatomy and compiles browser
  behaviors into `components/generated/` (no publication needed).
- `npm run build:tokens -- --published <dir>` exports DTCG contexts, builds the
  six site appearance themes (`<dir>/assets/themes/`) and component CSS bundles,
  and updates embedded `buttons.css` and `components.css`. Rebuild the Go server
  afterwards so embedded bindings match the asset manifest.
- In production, a separate asset-builder container writes assets outside
  approved publications; see
  [deployment](../docs/operations/deployment.md#generated-publication-assets).

See [component ownership](components/README.md),
[generated components](../docs/reference/components.md),
[buttons](../docs/reference/buttons.md), and
[DTCG exports](../docs/reference/dtcg.md).
