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
- source-payload cache: `.cache/material-token-table-responses`;
- crawl receipts: `crawl-results`;
- page timeout: 10 seconds;
- maximum pages: 500.

`--max-pages` limits how many failed pages a repair crawl retries. Unretried
failures remain in the merged crawl receipt and keep its capture incomplete.

Use `--refresh` to re-download source payloads whose token-table URL is already
in the cache. Without it, the crawl observes the page but reuses the cached
source payload when its artifact is present.
