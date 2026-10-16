# Offline publication

For automated production updates with saved reports, approval and atomic
generation switching, see [scheduled updates](../operations/updates.md).

Token Catalog publishes a verified snapshot into a local read-only data set.
The published data is derived from the snapshot. The snapshot remains the
canonical source and is not modified.

## Create a publication

Using the newest verified snapshot:

```text
tokens publish
```

Using an explicit snapshot and output directory:

```text
tokens publish \
  --snapshot-dir snapshots/material-token-data/material-2026-09-05-1816 \
  --output published
```

The timeline includes all supported snapshots in the selected snapshot's parent
directory, up to the selected capture date. Token Catalog verifies them and compares
each consecutive pair. To use another history directory:

```text
tokens publish \
  --snapshot-dir snapshots/material-token-data/material-2026-09-05-1816 \
  --history-dir snapshots/material-token-data \
  --output published
```

The command verifies the snapshot before reading it. It builds into a staging
directory and replaces the output only after all artifacts succeed. A failed
publication leaves an existing output directory untouched.

For local contributor development, the same command can be run with:

```text
go run ./cmd/tokens publish
```

## Serve a publication

Start the read-only server against the generated directory:

```text
tokens serve --published published --listen 127.0.0.1:8080
```

The server loads the SQLite read index into ordinary Go structs at startup.
It does not crawl, publish, or write through HTTP. A new publication becomes
visible after restarting the server.

## Generated output

```text
published/
  publication.json
  tokens.sqlite
  downloads/
```

`publication.json` identifies the publication, current snapshot, snapshot
manifest hash, capture status, review status, subjects, source payloads, the
public token view, and the changes between consecutive retained snapshots.

`tokens.sqlite` is a generated read index. It stores subjects, tokens, context-
specific values, lifecycle notices, source payloads, and provenance links.
It can be deleted and rebuilt from the verified snapshot. Re-run `tokens
publish` after upgrading Token Catalog; the SQLite index is rebuilt rather than
migrated in place.

`downloads/tokens.json` contains the capture timestamp and all subjects. Each
subject also has a scoped download under
`downloads/<subject-kind>/<subject-slug>.json`. A subject download contains its
kind, slug, display name, and tokens. Each token has its name, type, lifecycle
when known, and resolved values. Context values use readable labels such as
`Theme: Dark`; source IDs and raw payload records are not included.

A component token download looks like this:

```json
{
  "capturedAt": "2026-09-05T18:16:00Z",
  "kind": "component",
  "slug": "fab-menu",
  "name": "FAB menu",
  "tokens": [
    {
      "name": "md.comp.fab-menu.primary-container.list-item.container.color",
      "type": "COLOR",
      "values": [
        {
          "context": [
            "Platform: Android",
            "Theme: Light",
            "Audience: 1P Baseline",
            "Contrast: Default"
          ],
          "default": true,
          "value": {"color": {"red": 0.827451, "green": 0.8901961, "blue": 0.99215686, "alpha": 1}}
        }
      ]
    }
  ]
}
```

The current publication does not replace or copy the immutable snapshot's raw
evidence. Raw graph data and source payloads remain available from the
snapshot directory.

## Web assets

The server embeds CSS, JavaScript, fonts and logos in the Go binary. CSS is
maintained in `catalog/web/static/app.css` and `token-table.css`. HTML templates
are written in templ; generated Go files are committed alongside them.
See [development](../development.md) for regeneration and checks.

HTMX is bundled in `catalog/web/static/htmx.min.js`, and Roboto is self-hosted.
The server does not need a CDN. Their licenses are listed in
[third-party notices](../../THIRD_PARTY_NOTICES.md).

## What is preserved

The publication keeps the distinction between:

- component, style, and foundation subjects;
- source token metadata and resolved values;
- values under different context tags;
- lifecycle notices and ordinary token state;
- source pages, source payloads, and source record IDs.

A token with several context-specific resolved values is stored as one token
with several values. It is not collapsed into whichever value happened to be
read first.

## Source tables and compound values

Each token's optional `presentation` field preserves its source token-set name,
label, display-group path and order. The browser and downloads use the same
membership. Token sets such as “Button - Color - Elevated” belong to the Buttons
component.

A shape is one compound value. Its `defaultSize` applies to unspecified corners,
and named corners override that default. A dimension with a known unit and an
omitted numeric `value` represents zero in the source format. The display keeps
explicit zero corner overrides and draws supported rounded shapes from all four
corner values.

Publication schema 4 stores this presentation metadata. Rebuild saved publications
with `tokens publish` to populate it. The server also reads schema 3 publications
using the previous table grouping.

## Change history

The global change log and each component's history show all detected changes,
newest first, grouped by observation date. Each group compares a capture with the
one before it. The first capture is a starting point, not a list of additions.

These are observation dates, not Material's exact change dates. A change can have
occurred at any point between the two captures. History is limited to snapshots
you retain. Unsupported snapshot schemas are excluded with a message; corrupt
supported snapshots stop publication instead of silently leaving a gap.

Publishing again rebuilds the timeline without duplicate change records. Keep
the older snapshots and re-run `tokens publish` after creating a new one, then
restart the server. No crawl is needed just to rebuild a publication.
