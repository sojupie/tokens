# Offline publication

Source Loom publishes a verified snapshot into a local read-only data set.
The published data is derived from the snapshot. The snapshot remains the
canonical source and is not modified.

## Create a publication

Using the newest verified snapshot:

```text
sourceloom publish
```

Using an explicit snapshot and output directory:

```text
sourceloom publish \
  --snapshot-dir snapshots/material-token-data/material-2026-09-05-1816 \
  --output published
```

To include change records against a previous verified snapshot:

```text
sourceloom publish \
  --snapshot-dir snapshots/material-token-data/material-2026-09-05-1816 \
  --previous-snapshot snapshots/material-token-data/material-2026-08-12-1605-fixed \
  --output published
```

The command verifies the snapshot before reading it. It builds into a staging
directory and replaces the output only after all artifacts succeed. A failed
publication leaves an existing output directory untouched.

For local contributor development, the same command can be run with:

```text
go run ./cmd/sourceloom publish
```

Normal documentation and deployment should use the compiled `sourceloom`
binary.

## Generated output

```text
published/
  publication.json
  tokens.sqlite
  downloads/
```

`publication.json` identifies the publication, current snapshot, snapshot
manifest hash, capture status, review status, subjects, source payloads, the
public token view, and any changes derived from the supplied previous snapshot.

`tokens.sqlite` is a generated read index. It stores subjects, tokens, context-
specific values, lifecycle notices, source payloads, and provenance links.
It can be deleted and rebuilt from the verified snapshot.

`downloads/tokens.json` contains all normalized subjects. Each subject also has
a scoped download under `downloads/<subject-kind>/<subject-slug>.json`. These
files carry the publication and snapshot identifiers so a downloaded token set
does not lose its source context.

The current publication does not replace or copy the immutable snapshot's raw
evidence. Raw graph data and source payloads remain available from the
snapshot directory.

## Web assets

The read-only web server embeds its assets in the `sourceloom` binary. The
stylesheet is generated from `internal/web/static/tailwind.css` with the
standalone Tailwind CLI:

```text
tailwindcss -i internal/web/static/tailwind.css -o internal/web/static/app.css --content "internal/web/**/*.templ" --minify
```

HTMX is pinned in `internal/web/static/htmx.min.js`; the server does not need a
CDN at runtime.

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

## Current limitations

- Without `--previous-snapshot`, the publication contains one snapshot and no
  change records.
- A publication compares only the explicitly supplied previous snapshot; it
  does not search the filesystem for history automatically.
- The web server exposes the read-only API, HTML subject pages, change history,
  normalized downloads, embedded styling, and HTMX subject filtering. More
  discovery and presentation features can be added later.

These are later publication steps. They do not change the immutable snapshot
contract.
