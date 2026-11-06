# Catalog web

`web.go` preserves the public server API. The implementation has three owners:
HTTP handlers load publication data, playground selection normalizes options, and
presentation models supply rendering data. Templates render that data through
Forge primitives.

| Directory | Ownership |
| --- | --- |
| Root Go files | Public server API and generation entry point |
| `internal/server/` | Routing, handlers, publication access, caches and integration tests |
| `internal/playground/` | Pure case selection and family-specific control validation |
| `internal/presentation/` | Rendering models, formatting and dependency collection |
| `templates/pages/`, `templates/shared/` | Pages, layout and shared fragments |
| `templates/tokens/`, `templates/downloads/` | Token tables, details and download forms |
| `templates/playground/`, `templates/previews/` | Controls, specimens and image markup |
| `templates/helpers/` | Render helpers that compose templates |
| `internal/render/` | Generated Go; edit source templates instead |
| `assets/scripts/`, `assets/styles/` | Maintained browser scripts and Catalog CSS |
| `assets/generated/`, `assets/vendor/` | Compiled scripts and third-party libraries |
| `assets/images/`, `assets/fonts/` | Embedded media and licenses |
| `tests/browser/` | Browser-script unit checks |

## HTTP ownership

The server keeps route registration in `server.go`. `pages.go` owns page requests;
`tokens.go` and `markdown.go` expose data representations. Download selection,
preparation and form rendering have separate files because changing options must
not prepare an export. `asset_delivery.go` resolves component dependencies;
`response_cache.go` owns encoded response reuse. `discovery.go`, `mcp.go` and
`seo.go` expose their respective discovery contracts. Tools, appearance and
illustrations retain their own owners.

Presentation and playground packages cannot import the server. Renderers use the
small `AssetDelivery` interface for component links and published-image lookup;
request handling and mutable caches stay in the HTTP package.

## Source to output

Run `npm run generate` after editing templates or browser sources. The pinned
compiler writes Catalog Go to `internal/render/` and Forge Go to its own renderer
package. TypeScript compiles to the corresponding generated asset folder.
Commit source and generated output. `npm run check:components` checks types and
output freshness; `npm run check:templates` checks both template renderers and the
Forge public facade. Use these commands rather than running templ across the repo,
which would place Go files beside their template sources.

Run `go test ./catalog/web/...` and `npm run test:ui`. Rebuild and restart after
changing embedded assets or templates.

## Public paths

`assets/static.go` derives stable public names from embedded source/output folders.
Existing `/static/...` paths remain unchanged. Duplicate asset names fail at
startup. Forge runtime and site-shell aliases remain available to Catalog and
statically built tools.

Published token CSS and rendered thumbnails remain publication outputs. Captured
source data does not belong in this folder. The embedded OpenAPI contract lives
beside its discovery handler in `internal/server/`.
