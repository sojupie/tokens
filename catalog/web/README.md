# Catalog web

`web.go` exposes the public HTTP server API. Request handling, playground state
selection, and presentation models live in separate internal packages so
templates stay decoupled from the HTTP layer.

| Directory | Responsibility |
| --- | --- |
| Root Go files | Public server API and `go:generate` entry point |
| `internal/server/` | Routing, handlers, publication loading, response cache, discovery/MCP, and integration tests |
| `internal/playground/` | Pure preview case selection and control validation |
| `internal/presentation/` | Rendering models, formatting, and asset dependency collection |
| `templates/` | Templ pages, shared layout, token tables, downloads, playgrounds, and preview helpers |
| `internal/render/` | Generated Go templates (edit `templates/` instead) |
| `assets/scripts/`, `assets/styles/` | Maintained browser TypeScript/JS and Catalog CSS |
| `assets/generated/`, `assets/vendor/` | Compiled scripts and third-party libraries |
| `assets/images/`, `assets/fonts/` | Embedded media and font licenses |
| `tests/browser/` | Browser-script unit tests |

`internal/presentation` and `internal/playground` do not import
`internal/server`; renderers resolve component links and published images
through the small `AssetDelivery` interface.

## Build and check

Run from the repository root after editing templates or browser scripts:

```sh
npm run generate
npm run check:components
npm run check:templates
go test ./catalog/web/...
npm run test:ui
```

`npm run generate` compiles Catalog templates into `internal/render/`, Forge
templates into Forge's renderer package, and TypeScript into
`assets/generated/`. Commit both sources and generated output, and rebuild the
Go binary after changing embedded assets or templates.

`assets/static.go` maps embedded folders to `/static/...` URLs and fails at
startup on duplicate names. Published token CSS and thumbnails are loaded from
the publication directory, not embedded here.
