# Tokens web publication plan

**Status:** Planned
**Created:** 2026-09-07
**Purpose:** Publish verified Material token snapshots through a small self-hosted Go service.

## Decision

Tokens keeps immutable snapshot directories as its canonical source data.
The web service does not crawl Material, modify snapshots, or write its own
source of truth.

The publication flow is:

```text
verified snapshots
        |
        v
    tokens publish
        |
        +--> generated SQLite read index
        +--> generated downloads
        +--> publication metadata
        |
        v
    tokens serve
        |
        +--> in-memory current data
        +--> Templ HTML pages
        +--> HTMX interactions
        +--> JSON API
        +--> downloads
```

SQLite is a generated, disposable read model. It is not manually maintained
and it is not the canonical token store. Deleting it and rebuilding it from
verified snapshots must produce the same publication data.

The first deployment is read-only and can run locally behind Cloudflared.

## Technology choices

- Go HTTP server and standard-library routing where practical.
- Templ for server-rendered HTML.
- HTMX for search, filtering, expandable sections, and partial-page updates.
- Tailwind compiled at build time; no Tailwind runtime is required.
- `go:embed` for compiled templates and static assets in the release binary.
- SQLite as the generated durable read index.
- Go application memory for frequently used current data and component pages.
- No MongoDB, SPA, React, authentication, or admin write API in the first
  version.

The production binary should not require Node or pnpm. The build environment
may use the standalone Templ and Tailwind tools to generate Go and CSS assets.

## Publication directory

The generated publication directory should have a small, explicit shape:

```text
published/
  tokens.sqlite
  publication.json
  downloads/
```

`publication.json` identifies the published snapshots and the generation
version. The SQLite database and downloads can be deleted and regenerated.

The published data must retain links back to:

- snapshot ID;
- manifest hash;
- source page URL;
- source-payload URL;
- source record ID where available.

## `publish` command

Add:

```text
tokens publish --snapshot-dir <verified-snapshot> --output <published-dir>
```

The command must:

1. verify the input snapshot;
2. read the generated token graph and related evidence;
3. extract components, foundation/style subjects, tokens, lifecycle notices,
   source links, and semantic changes;
4. create the SQLite read index;
5. create normalized token downloads;
6. write publication metadata;
7. replace the output atomically only after generation succeeds.

Publishing a new snapshot must not mutate an older publication directory.

## SQLite read model

The first schema should remain small and query-oriented:

```text
snapshots
  id, created_at, manifest_hash, source_date

subjects
  id, slug, name, kind, source_url

tokens
  id, subject_id, name, category, value_json, lifecycle

token_changes
  id, subject_id, token_name, change_type, before_json, after_json,
  previous_snapshot_id, current_snapshot_id

snapshot_sources
  snapshot_id, source_url, payload_url
```

Use direct SQL. Do not add an ORM or a general repository abstraction for the
first version.

The database stores normalized views. Raw source payloads and generated graph
files remain in the immutable snapshot and are available as downloads.

## In-memory server state

At startup, load common read data into simple Go structs and maps:

```text
current snapshot metadata
component and style summaries
current token values
change summaries
```

Use memory for ordinary component searches and detail pages. Use SQLite for
less common history and filtering queries. Do not load all raw source payloads
into memory.

A future reload can build a replacement state and swap it in one operation.
The first version may simply restart the server after publishing.

## HTTP surface

HTML routes:

```text
GET /                         overview
GET /components               component and style list
GET /components/{slug}        subject detail
GET /components/{slug}/changes
GET /changes                  global change log
GET /downloads/...             approved downloads
```

JSON routes:

```text
GET /api/v1/health
GET /api/v1/releases
GET /api/v1/components
GET /api/v1/components/{slug}
GET /api/v1/components/{slug}/changes
```

The API and HTML pages use the same published read model. HTML fragments used
by HTMX are not a replacement for the JSON API.

## Web UI

The first UI should provide:

- component and style search;
- an individual subject page;
- grouped current token values;
- lifecycle notices with their source evidence;
- source-page and download links;
- a global change log;
- subject-specific change history.

The UI should describe dates as detection dates, for example:

```text
Detected in snapshot: September 5, 2026

md.comp.snackbar.container.width
480dp -> 560dp
```

It must not claim that the upstream change happened on that date unless
Material provides an authoritative change date.

Baseline, Expressive, theme, and other context dimensions should only be
displayed as source evidence when the graph records a direct relationship.
The publication layer must not infer them from names or page layout.

## Downloads

Start with:

- current normalized tokens for a subject;
- complete normalized token data;
- snapshot manifest;
- generated token graph;
- source-payload details;
- source payloads or an evidence archive as an advanced download.

All paths must resolve inside approved snapshot or publication directories.
The server must not expose arbitrary filesystem paths.

## Deployment

The server should bind to localhost by default:

```text
tokens serve --published ./published --listen 127.0.0.1:8080
```

Cloudflared can expose that local address. The initial service has no crawl,
publish, or write endpoints. Publishing is a local operator action followed by
a server restart.

## Implementation phases

### Phase 1: publication model

- Define publication view types for snapshots, subjects, tokens, and changes.
- Define the generated SQLite schema.
- Build a small index from one verified snapshot.
- Add tests for deterministic rebuilds.

### Phase 2: read-only API

- Add `tokens serve`.
- Add health, release, subject, token, and change endpoints.
- Add path validation and read-only tests.

### Phase 3: server-rendered UI

- Add Templ layouts and pages.
- Add component/style search.
- Add subject detail and change-history pages.
- Add source and download links.

### Phase 4: HTMX and Tailwind

- Add search and filtering interactions.
- Add expandable token groups.
- Compile and embed CSS and static assets.
- Keep the site usable without client-side application state.

### Phase 5: repeatable publication

- Add `tokens publish`.
- Generate normalized downloads and change history.
- Atomically replace the publication directory.
- Document the crawl, verify, compare, publish, and serve workflow.

### Phase 6: deployment

- Document local operation.
- Document Cloudflared exposure.
- Add a simple startup/service configuration only if manual startup becomes
  inconvenient.

## Deferred

- MongoDB or another external database;
- React islands and live component galleries;
- user accounts and authentication;
- admin editing;
- background crawling from the server;
- database writes from HTTP handlers;
- full-text search;
- database migrations beyond a schema-version check;
- automatic server reload;
- generated full snapshot archives unless individual downloads are insufficient.
