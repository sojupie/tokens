# Documentation capture

`tokens guidance` captures rendered documentation separately from token payloads.
It saves the selected page content as HTML, plain text and Markdown. Markdown
conversion preserves headings, lists, tables, links and image descriptions.
Extractor version 2 retains multiline table cells instead of flattening the table.

Component selection uses the token crawler's existing `inputs/page-catalog.json`.
For each component specs URL, it captures the fixed `overview`, `specs`,
`guidelines` and `accessibility` pages. Style and foundation token pages are
excluded. It does not perform another sitemap or link-discovery pass.

```sh
# Preview all selected URLs without opening Chromium:
go run ./cmd/tokens guidance --list

# Capture all four pages for one component:
go run ./cmd/tokens guidance --component tooltips

# Capture every component in the existing catalog:
go run ./cmd/tokens guidance
```

Repeat `--component` to select several components. Unknown component slugs fail
before capture. Use `--page-catalog PATH` to read another token crawler catalog.

`--url URL` alone captures only the supplied URLs. Combine it with `--component`
to include supplementary guidance, such as a linked transition-pattern page.
URLs are deduplicated before capture. Catalog changes are managed by the existing
token crawler; this command only reads the catalog.

Use `--chrome-path PATH` if Chromium is not on the executable search path.
`--selector` defaults to `main`. `--page-timeout` defaults to 45 seconds per page.

Each run creates a new directory under `snapshots/material-guidance/`. Existing
runs remain unchanged. The directory contains:

- `capture.json`: requested URLs, final URLs, link bases, titles, HTTP status, timestamps,
  extraction version, file paths, SHA-256 hashes and failures.
- `<url-hash>.html`: the rendered content selected for capture.
- `<url-hash>.txt`: its visible text, including page controls within that content.
- `<url-hash>.md`: structured text with navigation and Material's section-copy
  controls and inline definition popups removed.

The command scrolls through the page to request lazy content, then waits for text
to settle. A timeout, HTTP error or missing content marks the run incomplete and
returns an error. Successful pages remain available in that run. Unfinished runs
retain a `.pending-` directory.

Compare a page's `markdownSha256` across runs to detect extracted-content changes.
Its URL-derived filename stays the same, so `diff -u` works directly on the two
Markdown files. `htmlSha256` also detects markup changes. Extraction changes
increment `extractorVersion`; compare captures made with the same version when
checking upstream changes.

Only the selected component tabs and explicit URLs are captured. The receipt's `complete` field means those
pages were captured successfully, not that every page, interactive state or piece
of guidance has been discovered. The command does not open collapsed controls or
cycle through token-table configurations.

Image URLs and descriptions are retained. Image files, video frames and text
embedded in diagrams are not captured. Measurements available only in images
still need visual inspection. Captures do not classify requirements or claim
implementation coverage.

Generated captures stay outside Git. This command does not change token snapshots,
DTCG exports, publications or the website.

[Guidance curation](guidance-curation.md) defines the maintained per-component
checks file, source history, implementation evidence and validation command.
