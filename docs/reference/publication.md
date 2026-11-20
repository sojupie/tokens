# Offline publication

`tokens publish` builds a read-only publication from a verified snapshot
without modifying the snapshot. For automated production updates, review, and
atomic switching, see [scheduled updates](../operations/updates.md).

## Create and serve

```sh
tokens publish
tokens serve --published published --listen 127.0.0.1:8080
```

Override the snapshot, history, or output paths when needed:

```sh
tokens publish \
  --snapshot-dir snapshots/material-token-data/material-2026-09-05-1816 \
  --history-dir snapshots/material-token-data \
  --output published
```

`publish` verifies the snapshot (plus all supported snapshots in `--history-dir`
up to its capture date), builds into a staging directory, and atomically swaps
the output directory on success. `serve` loads `tokens.sqlite` into memory at
startup and serves read-only.

## Generated output

```text
published/
  publication.json
  tokens.sqlite
  downloads/
    tokens.json
    <subject-kind>/<subject-slug>.json
```

- **`publication.json`:** Publication metadata, manifest hash, capture/review
  status, subjects, payloads, and consecutive snapshot diffs.
- **`tokens.sqlite`:** Derived read index storing subjects, tokens,
  context-specific values, lifecycle notices, and provenance. Rebuilt from
  verified snapshots rather than migrated in place.
- **`downloads/`:** Whole-collection (`tokens.json`) and per-subject JSON
  exports using readable context labels (e.g. `Theme: Dark`), without raw
  payload records:

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
          "context": ["Platform: Android", "Theme: Light", "Audience: 1P Baseline", "Contrast: Default"],
          "default": true,
          "value": {"color": {"red": 0.827451, "green": 0.8901961, "blue": 0.99215686, "alpha": 1}}
        }
      ]
    }
  ]
}
```

## Embedded web assets

The Go binary embeds site CSS, JavaScript, self-hosted Roboto fonts, and
vendored HTMX (`catalog/web/assets/vendor/htmx.min.js`) so the server needs no
CDN. See [development](../development.md) for template/asset builds and
[third-party notices](../../THIRD_PARTY_NOTICES.md) for licenses.

## Data model and presentation

- Preserves component/style/foundation boundaries, all context-specific values
  per token, lifecycle notices, and source provenance.
- Publication schema 4 stores each token's `presentation` metadata (source
  token-set name, label, display-group path, and order); schema 3 publications
  remain readable.
- Compound rounded shapes apply `defaultSize` to unspecified corners while
  preserving explicit zero overrides (see [DTCG conversion](dtcg.md#material-conversion)).

## Change history

Change logs group detected differences by observation date (newest first),
comparing each retained snapshot against the previous one (the earliest snapshot
is the baseline).

- `Added` / `Removed` require a matching extracted payload (by filename, ignoring
  release directories) in both snapshots.
- A payload present on only one side records `Coverage added` / `Coverage lost`.
- A new or missing context on an existing token records `Context added` /
  `Context removed`.
- Unsupported historical schemas are skipped with a warning; corrupt supported
  snapshots fail the build. Re-running `tokens publish` rebuilds the timeline
  cleanly without another crawl.
