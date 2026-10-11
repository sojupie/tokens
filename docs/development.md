# Development

Install Go 1.26 or newer. Chrome or Chromium is needed for live crawling.
Node.js is needed for the browser-script checks. Serving a saved publication
and running the Go tests do not require Chromium.

## Build and capture

Run commands from the repository root:

```sh
go build -o ./bin/tokens ./cmd/tokens
./bin/tokens crawl
./bin/tokens snapshot create
./bin/tokens snapshot verify
./bin/tokens publish
./bin/tokens serve --published published --listen 127.0.0.1:8081
```

On Windows, build `bin/tokens.exe` and invoke it with PowerShell. Use
`crawl --chrome-path PATH` when Chromium cannot be found automatically.
The [CLI reference](reference/cli.md) covers explicit paths and smaller captures.

For development with existing data, run `publish` and `serve` against a verified
snapshot. Rebuilding a publication does not need another crawl. Avoid modifying
snapshot files after creation; their checksums verify the captured evidence.

## Checks

```sh
go test ./...
node --test internal/web/ui_test.cjs
```

Tests use local fixtures. Keep normal checks offline. Live captures and DNS or
Cloudflare checks are separate operational checks.

## Templates and assets

Edit `internal/web/templates.templ`, then regenerate its Go file with the pinned
templ version:

```sh
go run github.com/a-h/templ/cmd/templ@v0.3.960 generate -f internal/web/templates.templ
```

Commit both the template and generated Go file. CSS and JavaScript are maintained
in `internal/web/static/` and embedded in the binary. Rebuild and restart the
server after changing them. Styles come from the published token data through
`internal/web/theme.go`; layout rules are in the static stylesheets.

## Code layout

| Path | Responsibility |
| --- | --- |
| `cmd/tokens/` | CLI commands and server startup |
| `internal/crawl/` | Browser capture, source discovery, payload cache and receipts |
| `internal/tokensnapshot/` | Snapshot assembly, verification and semantic comparison |
| `internal/publish/` | Token projection, retained history, SQLite and JSON exports |
| `internal/update/` | Saved runs, review, activation, retry and notifications |
| `internal/web/` | HTML, filters, downloads, API and agent interfaces |
| `schemas/` | Snapshot document schemas |
| `inputs/page-catalog.json` | Reviewed Material Design 3 source pages |
| `deploy/` | Docker, systemd and deployment scripts |

## Other design systems

The current source implementation expects Material Design 3 pages and payloads.
A further collection needs its own source catalog, capture and extraction rules,
publication mapping, and licensing notes. Some names, paths, token relationships
and theme generation also assume Material Design 3. Treat those as implementation
work when adding a source; a new hostname or URL prefix alone is insufficient.

Keep source provenance and completeness checks for any new collection. Its
values should be traceable to the source independently of the rendered site.

## Hover prefetch

Internal page links prefetch after 200 ms of mouse hover or keyboard focus.
Leaving the link before the delay cancels the pending work. Downloads, API
routes, external links and HTMX controls are excluded. Each document prefetches
at most eight distinct destinations, and data-saving or 2G connections skip it.

The script uses browser speculation rules where supported, then native document
prefetch hints. Other browsers keep normal navigation. Full HTML prefetch
responses can stay in the browser's private cache for 30 seconds; ordinary
requests continue to revalidate. HTMX fragments and Markdown keep their existing
cache policy.

To check this in a browser, hover a component link briefly, then for more than
200 ms. In DevTools, check the prefetch request or the speculative loading panel.
Click the link and confirm the navigation reuses the prefetched response.
