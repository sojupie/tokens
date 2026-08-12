# M3 Source CLI

The command is organized around the work being done:

```text
m3source crawl
m3source snapshot create
m3source snapshot verify
m3source snapshot compare
```

Show help at any level:

```text
go run ./cmd/m3source --help
go run ./cmd/m3source crawl --help
go run ./cmd/m3source snapshot --help
go run ./cmd/m3source snapshot create --help
```

## Normal workflow

Run a crawl using the checked-in page URL catalog:

```text
go run ./cmd/m3source crawl
```

Create a snapshot from the newest crawl receipt:

```text
go run ./cmd/m3source snapshot create
```

The command derives the snapshot ID from the crawl start time in UTC and uses
`snapshots/material-token-data/<snapshot-id>` as the output directory. It
refuses to use an existing non-empty output directory.

Verify the newest snapshot:

```text
go run ./cmd/m3source snapshot verify
```

Compare the two latest snapshot runs:

```text
go run ./cmd/m3source snapshot compare
```

Retry only the failed pages from the newest crawl receipt:

```text
go run ./cmd/m3source crawl --retry-failed
```

Select a specific receipt when needed:

```text
go run ./cmd/m3source crawl \
  --retry-failed-from crawl-results/20260811T143000.000000000Z.json
```

The repair crawl opens only pages that previously had a page error, missing
token-table observation, or failed/missing source payload. Its new crawl
receipt retains the previous successful pages and replaces only the retried
pages and their source payload records. In that receipt, `pagesAttempted` is
the number opened by the repair, while `pagesCarriedForward` is the number
retained from the earlier crawl. `complete` describes the merged receipt. If
the retry cannot capture any page successfully, it exits without writing a
new receipt.

## Broad token-source inventory

Use the broader discovery scope when reviewing foundation and style pages in
addition to component specification pages:

```text
go run ./cmd/m3source crawl \
  --discover \
  --scope token-source-pages \
  --dry-run
```

The scope considers component pages ending in `/specs` and style token pages
ending in `/tokens`, `/specs`, `-tokens`, or `-scale`. It produces candidate page results and drift. It
does not automatically make unreviewed pages part of the normal catalog.
After reviewing the additions, `--discover` and `--update-page-catalog` may be
combined for one live discovery and capture. The update writes the reviewed
candidate catalog before the crawl continues.

## Explicit paths

Every automatic choice can be overridden when working with a specific receipt
or snapshot:

```text
go run ./cmd/m3source snapshot create \
  --crawl-receipt crawl-results/20260811T143000.000000000Z.json \
  --page-catalog inputs/page-catalog.json \
  --output snapshots/material-token-data/m3-2026-08-11-1430-example \
  --id m3-2026-08-11-1430-example

go run ./cmd/m3source snapshot verify \
  snapshots/material-token-data/m3-2026-08-11-1430-example

go run ./cmd/m3source snapshot compare \
  snapshots/material-token-data/m3-2026-08-11-1430-older \
  snapshots/material-token-data/m3-2026-08-11-1500-newer
```

With no paths, `snapshot compare` selects the two latest snapshot runs by the
manifest `createdAt` timestamp. The positional forms and flag forms cannot be mixed for `verify` or `compare`.
Unknown flags, unexpected positional arguments, missing paths, unsupported
schema versions, incomplete captures, and non-empty snapshot output
directories fail before snapshot assembly proceeds.

## Crawl controls

The crawl command uses these defaults:

- page URL catalog: `inputs/page-catalog.json`;
- live discovery scope: `component-specs`;
- source-payload cache: `.cache/material-token-table-responses`;
- crawl receipts: `crawl-results`;
- page timeout: 10 seconds;
- maximum pages: 500.

`--max-pages` limits how many failed pages a repair crawl retries. Unretried
failures remain in the merged crawl receipt and keep its capture incomplete.

Use `--scope token-source-pages` with `--discover` for a deliberate inventory
of component, foundation, and style token-source pages. The scope does not
mean every page on the Material site.

During a crawl, the command reports page selection, Chromium startup, page
progress, source-payload observations, and source-payload reuse, storage, or
failure. A dry run reports observed payload URLs but explicitly skips payload
persistence, so its receipt cannot create a snapshot.

By default, the crawl observes each page but reuses a valid cached source
payload without an HTTP download. Use `--refresh` to re-download source
payloads whose URL is already in the cache. Cached payloads still appear in the
receipt because a complete crawl needs them for offline snapshot assembly.

If only the generated snapshot needs to be rerun, use `snapshot create` again
with the existing crawl receipt. That path reads cached source-payload artifacts
and does not open Chrome or crawl pages.
