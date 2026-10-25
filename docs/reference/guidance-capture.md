# Documentation capture

`tokens guidance` captures rendered documentation separately from token payloads.
It saves the selected page content as HTML, plain text and Markdown. Markdown
conversion preserves headings, lists, tables, links and image descriptions.

Capture all four Tooltip documentation pages:

```sh
go run ./cmd/tokens guidance \
  --url https://m3.material.io/components/tooltips/overview \
  --url https://m3.material.io/components/tooltips/specs \
  --url https://m3.material.io/components/tooltips/guidelines \
  --url https://m3.material.io/components/tooltips/accessibility
```

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

Only the requested URLs are captured. The receipt's `complete` field means those
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
