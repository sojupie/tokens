# Token Catalog CLI

All commands run through `cmd/tokens` (use `go run ./cmd/tokens ...` or your
built `./bin/tokens ...`). Pass `--help` at any level for full flags.

## Commands

```text
tokens crawl                      Capture Material source pages and payloads
tokens snapshot create            Build an immutable snapshot from a crawl receipt
tokens snapshot verify            Verify snapshot hashes, schemas, and payloads
tokens snapshot compare           Semantically diff two verified snapshots
tokens publish                    Build SQLite read index, history, and downloads
tokens export                     Export a selected context as DTCG 2025.10 JSON
tokens serve                      Start the read-only catalog server
tokens update <subcommand>        Run scheduled updates (run, status, inspect, approve, retry, notify, test-notification)
tokens guidance [subcommand]      Capture docs, snapshot baselines, or validate checks.json
tokens site-shell                 Export shared header/footer HTML and assets for static tools
tokens illustrations              Render component thumbnails and write image manifests
tokens forge inspect              Scan CSS sources and write css-usage.json
```

## Capture, snapshot, and publish

```sh
tokens crawl
tokens snapshot create
tokens snapshot verify
tokens snapshot compare [--verbose] [--output review/compare.txt]
tokens publish
tokens serve --published published --listen 127.0.0.1:8081
```

- **Crawl defaults:** Page catalog `inputs/page-catalog.json`, cache
  `.cache/material-token-table-responses`, receipts `crawl-results`, 30 s page
  timeout, max 500 pages. Reuse cached payloads by default; pass `--refresh` to
  re-download, `--retry-failed` (or `--retry-failed-from RECEIPT`) to repair
  failed pages, or `--discover --scope token-source-pages --dry-run` to check
  for catalog drift. See [crawler](crawler.md).
- **Snapshots:** `snapshot create` writes to
  `snapshots/material-token-data/<snapshot-id>` (derived from UTC crawl start
  time; refuses non-empty directories) and works offline from cached payloads.
  `snapshot compare` diffs the two newest snapshots by `createdAt` semantically,
  ignoring payload hashes and release URLs. See [snapshot format](snapshot-format.md).
- **Explicit paths:** Override auto-selection when targeting specific artifacts
  (do not mix positional args and flags on `verify` or `compare`):

  ```sh
  tokens snapshot create --crawl-receipt crawl-results/<receipt>.json \
    --page-catalog inputs/page-catalog.json \
    --output snapshots/material-token-data/<id> --id <id>
  tokens snapshot verify snapshots/material-token-data/<id>
  tokens snapshot compare snapshots/material-token-data/<older> snapshots/material-token-data/<newer>
  ```

## Serving and exports

- **`tokens export`:** Exports a context to DTCG JSON (`--published DIR --output DIR`).
  Repeat `--context Name=Value` to override dimensions, or pass `--themes` for
  all six site appearances. See [DTCG exports and CSS](dtcg.md).
- **`tokens serve --assets DIR`:** Loads and verifies external generated assets
  against the publication and embedded bindings. `--motion-url` /
  `--motion-lab-url` (or `TOKENS_MOTION_URL` / `TOKENS_MOTION_LAB_URL`) set the
  tool navigation links. See [deployment](../operations/deployment.md#generated-publication-assets).
- **`tokens update`:** Manages unattended runs, review, retries, and SMTP/webhook
  alerts. See [scheduled updates](../operations/updates.md).

## Guidance, shell, and illustrations

- **`tokens guidance`:** `--list` prints Material doc URLs; `--component <slug>`
  captures pages to `snapshots/material-guidance/` (see [guidance capture](guidance-capture.md)).
  `guidance baseline` snapshots a working tree for assessments;
  `guidance validate [--against REF] [--file PATH]` validates `checks.json`
  files (see [guidance curation](guidance-curation.md)).
- **`tokens site-shell --out DIR --catalog URL [--published DIR]`:** Exports
  shared navigation/footer HTML, CSS, JS, and logo for static tools.
- **`tokens illustrations --published DIR [--assets DIR] [--chrome-path PATH]`:**
  Renders 720 × 450 catalog thumbnails across six appearances and writes
  `assets/illustrations/manifest.json`. Pass `--from-captures DIR` to encode
  existing captures without launching Chromium.
