# Tokens roadmap

**Status:** Draft for execution  
**Created:** 2026-08-11  
**Scope:** Material source discovery, capture, extraction, diffing, and snapshots

## 1. Mission

Tokens reports what Material published. It does not implement components or
decide what Token Forge supports.

The project should be useful beyond Token Forge. Its stable product is an immutable,
documented source snapshot that any tool can download and verify. A website,
API, package, or database can make that data easier to explore, but none is the
canonical store.

## 2. Boundary

Tokens owns:

- curated and live discovery of component, foundation, and style token-source
  pages;
- browser-assisted token-table observation;
- content-addressed payload capture;
- a loss-preserving Material-specific graph;
- semantic diffs between snapshots;
- source-payload details, page results, and immutable snapshot publication.

Tokens does not own:

- canonical DTCG design decisions;
- CSS, components, framework adapters, or application code;
- automatic promotion into Token Forge;
- a runtime service required by library consumers.

## 3. Technical direction

Use Go for the full ingestion pipeline. chromedp is the canonical browser path.
The retired Playwright prototype and Node package layer are not part of the
project.

Keep commands conceptually separate even if they initially share one binary:

```text
tokens crawl --discover --dry-run  # audit component-spec page drift
tokens crawl --discover --scope token-source-pages --dry-run  # inventory broader token-source pages
tokens crawl
tokens snapshot create
tokens snapshot verify
tokens snapshot compare
```

Keep implementation packages internal until an external Go API has a concrete
consumer.

## 4. Storage

Use files and content-addressed artifacts:

- Git: code, schemas, curated page URL catalog, small fixtures, and documentation.
- Local cache: downloaded source-payload files and crawl receipts.
- Snapshot bundle: source payloads, generated token graph, source-payload details,
  page results, and checksums.
- Object storage or GitHub Releases: immutable public snapshots and optional raw
  evidence bundles.

Do not add an operational database. Batch files fit immutable versioning,
hashing, offline use, and mirroring. A generated SQLite file can later serve
local queries without becoming a source of truth.

### Deferred: rendered measurements and component preview images

Do not make measurement-diagram extraction part of the first token snapshot or
the first Token Forge compiler milestone. The diagrams are difficult to interpret
reliably, and their labels may describe geometry, relationships, or guidance
rather than tokens. Revisit this only when a concrete component or conformance
requirement needs the data and the source representation can be verified.

Token-table component images are different. They are useful representative
previews, but they do not show every variant and are not measurement evidence.
For now, preserve their upstream URL in the loss-preserving source payload.
If offline image access becomes useful later, add an optional, separately
provenanced component-preview artifact. Do not use preview images as compiler
input or as evidence of variant coverage.

## 5. Versioning

Track these independently:

1. Tool semantic version.
2. Page URL catalog schema version.
3. Generated token graph schema version.
4. Snapshot manifest schema version.
5. Snapshot ID and manifest SHA-256.
6. Upstream page, payload URL, revision, and payload SHA-256.

An upstream update creates evidence and a semantic diff. It does not mutate an
existing snapshot or force a Token Forge release.

## 6. Phases

### Phase 0: repository cleanup

- Rename project documentation and the public command name. **Done:** the
  product is Tokens and the executable is `tokens`.
- Remove the Node and Playwright prototype. **Done.**
- Keep the current Go crawler behavior and tests passing. **Done for the
  current implementation.**
- Correct the repository remote. **Done:** `origin` points to
  `https://github.com/sojupie/tokens.git`.

**Exit:** the repository has one language and one ingestion path.

### Phase 1: deterministic capture

- Remove button-specific inspection from capture. **Done.**
- Validate response status, media type, JSON shape, and hash. **Partially
  done; source-payload variants are classified and uncompressed hashes are
  checked, but full upstream semantic validation remains open.**
- Separate volatile run reports from deterministic capture records. **Done.**
- Add cache verification and portable cache import/export. **Cache integrity
  verification done; portable cache import/export remains open.**
- Test page URL catalog drift, duplicate source payloads, corrupt cache entries, and stable
  manifests with recorded fixtures.

**Exit:** a page URL catalog capture can be reproduced and verified without recrawling.

### Phase 2: generic token extraction

- Define versioned generated-token-graph and source-payload-detail schemas.
  **Done for the current version-3 documents and standalone schema files.**
- Preserve names, IDs, states, contexts, reference trees, types, units, and
  unknown source fields.
- Extract representative foundations and component families without
  component-specific parser branches.
- Add human-readable inspection commands.

**Exit:** color, typography, shape, motion, buttons, groups, and icon buttons
extract through one model.

### Phase 3: semantic diff

- Compare token additions, removals, literals, references, contexts, activation,
  deprecation, and source-schema changes.
- Distinguish source metadata churn from design-value changes.
- Emit machine-readable and review-friendly reports.

**Exit:** a reviewer can understand why two snapshots differ without reading
raw payloads.

### Phase 4: immutable snapshots

- Implement deterministic bundle assembly and verification. **Done locally;
  source-payload bytes, gzip contents, safe artifact paths, catalog data, and
  current schemas are verified.**
- Generate the manifest, page URL catalog, source payloads, generated token graph,
  source-payload details, page results, and checksums.
- Publish to static storage and mirror to a durable release channel.
- Document retention and raw-evidence redistribution policy.

**Exit:** a consumer can pin, download, and verify a snapshot offline.

### Phase 5: public access

- Add a static documentation browser for snapshots and diffs.
- Add a downloadable SQLite index only if query use justifies it.
- Add an API or package only when a real consumer needs one.

**Exit:** people can explore the data without changing the canonical storage
model.

## 7. Tests and operations

Ordinary tests are offline and use small committed fixtures. Test discovery and
URL filtering, cache integrity, schema drift, extraction fidelity, semantic
diffs, deterministic serialization, bundle verification, and corrupt inputs.

Run live crawling manually or on a scheduled audit. A scheduled job can report
page URL catalog or token drift and publish an evidence candidate. It must not rewrite a
page URL catalog, publish over an existing snapshot, or promote Token Forge tokens without
review.

## 8. Immediate work

1. Rebuild the current provisional snapshot from fixed inputs and confirm
   byte-identical artifacts for every version-3 document.
2. Review lifecycle notices for duplicates and false positives.
3. Decide the explicit policy for a known unavailable upstream payload such as
   Snackbar before creating a Token Forge source lock.
4. Review baseline, Expressive, theme, and other context evidence without
   collapsing it into a guessed boolean.
5. Review representative `md.ref`, `md.sys`, `md.comp`, foundation, and
   component records before defining Forge candidates.
6. Decide public object storage, mirroring, and retention policy.
7. Make CLI documentation and distribution binary-first: use `tokens` in
   normal examples, keep `go run ./cmd/tokens` in contributor-only
   instructions, document local binary builds, and define versioned release
   artifacts for supported platforms.
