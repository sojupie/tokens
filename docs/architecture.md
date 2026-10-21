# Architecture

Token Catalog contains Source, Catalog, and Forge in one repository.

- Source (`source/`) captures upstream data, verifies snapshots, builds
  publications, and exports selected token contexts as DTCG.
- Catalog (`catalog/`) serves publications through the website and agent
  interfaces, and retains usage totals.
- Forge (`forge/`) builds CSS and component styles from DTCG and implements
  their browser behavior.

The parts share one Go module and the `cmd/tokens` executable. The root Node
package pins Style Dictionary for Forge. They ship together; no separate package
release or service is needed between these directories.

Source owns the published data model consumed by Catalog and the DTCG exporter.
Forge's compiler and Material mappings read DTCG documents. Its current build
script also runs the CLI exporter and reads Catalog's stylesheets to assemble
the site theme. Catalog embeds the generated component assets from Forge and
loads matching publication assets when available.

```text
upstream data -> Source snapshots -> publication -> Catalog
                                        |
                                        v
                                      DTCG
                                        |
                                        v
                                Forge CSS + behavior
                                        |
                                        v
                                  Catalog website
```

Generated assets are optional for serving a publication. The existing Go theme
renderer supports publications without them. Production update jobs do not yet
run the Node build. See [DTCG exports](reference/dtcg.md) for the build procedure.

## Design decisions

- Capture records pages, payloads and failures separately, so a successful page
  request does not hide a missing token payload.
- Source payloads are stored by their SHA-256 hash, preserving the bytes behind
  an extracted value. Snapshots carry their own source evidence and checksums.
- Comparisons use token values, references and contexts. Release URLs and
  refreshed source hashes do not create token change entries.
- Publication builds into a staging directory. Activation switches the current
  generation only after the candidate passes the update checks or is approved.
- The HTTP server resolves and loads one publication at startup. Chromium runs
  during capture, independently of the serving process.

## The flow

```text
inputs/page-catalog.json
        |
        v
    tokens crawl
        |
        +--> .cache/material-token-table-responses/
        |      downloaded source-payload files
        |
        +--> crawl-results/<crawl-id>.json
               crawl receipt
        |
        v
    tokens snapshot create
        |
        v
snapshots/material-token-data/<snapshot-id>/
        |
        +--> manifest.json
        +--> page-catalog.json
        +--> source-payloads/
        +--> generated-token-graph.json.gz
        +--> source-payload-details.json
        +--> page-results.json
        +--> lifecycle-notices.json
        +--> checksums.txt
        |
        +--> tokens snapshot verify
        +--> tokens snapshot compare
```

Verified snapshots can then be published for read-only browsing:

```text
verified snapshot(s)
        |
        v
    tokens publish
        |
        +--> published/tokens.sqlite
        +--> published/publication.json
        +--> published/downloads/
        |
        v
    tokens serve
        |
        +--> in-memory published views
        +--> Templ HTML pages
        +--> HTMX fragments
        +--> JSON API
        +--> downloads
```

The generated SQLite file and downloads are derived publication data. They can
be deleted and rebuilt from verified snapshots. The web server is read-only and
does not crawl Material or modify snapshots.

To repair a partial crawl, `tokens crawl --retry-failed` reads the newest
crawl receipt, opens only its failed pages, and writes another merged crawl
receipt. Successful pages and source payloads are carried forward; retried
pages replace their earlier results.

## What is fetched?

The crawler opens each selected Material token-source page with Chrome. Normal
crawls select pages from the reviewed page URL catalog. Broad discovery can
also consider component pages ending in `/specs` and style or foundation pages
ending in `/tokens`, `/specs`, `-tokens`, or `-scale`. The crawler watches each
page's network traffic for uppercase source-payload JSON responses. It then
downloads each unique source payload and stores the compressed response in the
source-payload cache.

The page HTML is observed but is not a snapshot input. The matching
source-payload response is the fetched source data. Different Material areas
use different payload filename stems, so Token Catalog does not assume that
every source payload is named `TOKEN_TABLE`.

## What is generated?

The crawl generates a crawl receipt. It records the pages opened, source-payload
URLs observed, downloaded file hashes, retrieval metadata, and errors.

Snapshot creation reads the crawl receipt and source-payload files. It
generates:

- `generated-token-graph.json.gz`: generic Material records extracted from the
  source payloads;
- `source-payload-details.json`: source URLs, hashes, retrieval details, and
  errors;
- `page-results.json`: page and extraction results;
- `manifest.json`: snapshot identity, inputs, generator version, and file list;
- `checksums.txt`: integrity records for every snapshot file.

The original successful source-payload files are copied into the snapshot.
They are not regenerated from the graph.

## Names and states

`crawl` means opening pages and downloading source payloads.

`snapshot create` means assembling source payloads and generated token data into
one immutable bundle.

`snapshot verify` checks file hashes, checksums, JSON schemas, and bundled source
payloads. Capture completeness and publication approval are separate checks.

`snapshot compare` means comparing two verified bundles.

`captureStatus` describes whether the crawl obtained all expected data:

- `complete` — all expected pages and source payloads succeeded;
- `partial` — some expected data is missing or failed.

`reviewStatus` records the snapshot's review state:

- `provisional` — evidence only;
- `approved` — accepted after review;
- `rejected` — not accepted.

Capture status and review status are separate. A snapshot can be successfully
assembled while still being partial and provisional.

## Collection boundaries

The hosted Material Design 3 collection lives at `/md3`. Root discovery files
advertise that collection's routes. The current server loads one publication.
Source capture, conversion, Catalog presentation, and Forge component mappings
currently contain Material-specific rules.

To add another design system:

1. Add its source catalog, capture and extraction rules, completeness checks,
   and licensing notes under Source. Preserve its native token names and evidence.
2. Map its supported values and contexts into the published model and DTCG.
   Retain source extensions for information that DTCG cannot represent.
3. Add its Catalog route prefix, publication selection, presentation mappings,
   and entries in origin discovery documents. Multi-collection serving needs
   implementation before one server can host both collections.
4. Add Forge component mappings and behavior for the system's own anatomy,
   states, and variants when component output is needed. DTCG standardizes token
   values; it does not define component markup or interactions.
5. Add its interpretation rules to the agent skill's references. Keep the
   existing skill name and install command while it serves the same workflow.

Introduce shared adapter interfaces when the second implementation establishes
what they need to share. A new collection does not require a separate repository.

## Agent instructions

`skills/token-catalog/SKILL.md` is the installable user skill. Its Material
reference supplies the hosted guide through an embedded template, with the
public origin, MCP version, and available generated downloads filled at runtime.
The discovery index hashes that rendered guide. Repository development guidance
lives in the root `AGENTS.md`.

[Development](development.md) maps these stages to packages.
[Updates and recovery](operations/updates.md) describes automated capture,
review and activation. [Snapshot format](reference/snapshot-format.md) and
[publication](reference/publication.md) describe the stored artifacts.

## Derived site assets

Production keeps Forge output outside approved publication generations. The
asset-builder reads the selected SQLite database, generates six appearance
contexts, and checks the result against the server's embedded bindings. A
completed build replaces the external asset symlink before server startup.
The server pins both publication and asset directories and verifies their
manifest checksums. Source review digests remain unchanged. See
[deployment](operations/deployment.md#generated-publication-assets).
