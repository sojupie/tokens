# Source Loom pre-compiler worklist

**Status:** Open review list
**Purpose:** Keep the remaining Source Loom contract and Token Forge policy decisions in one place.
**Compiler gate:** Do not start the Token Forge compiler until the items marked as gates are closed.

## Current position

Source Loom has:

- a Go-only crawl and snapshot pipeline;
- a checked-in page URL catalog;
- content-addressed source-payload storage;
- generic loss-preserving extraction;
- manifest, graph, source-payload details, page-results, and lifecycle-notice documents;
- standalone JSON Schemas in `schemas/`;
- deterministic snapshot creation and verification;
- semantic snapshot comparison;
- failed-page retry with merged crawl receipts;
- committed provisional snapshots;
- an accepted upstream 404 for the Snackbar token-table payload.

The latest lifecycle-aware snapshot is:

```text
m3-2026-08-11-1921-lifecycle
manifest SHA-256:
6c28a66d8082f336b038ebe64cdeed1d024fe1b59ee4a6ca4b808394cdead4f5
```

Its capture is still incomplete because the Snackbar source payload returns
404. Source must continue to report that fact. Forge may later decide whether
to accept the exception.

## 1. Source page scope

Source Loom must capture more than component specification pages. Material also
publishes token tables on foundation and style pages such as elevation, motion,
spacing, typography, shape, and color.

The generated graph is already largely page-type independent. It indexes the
records found in a source payload and does not require each record to belong to
a component. The current limitation is in page selection and catalog
validation: the crawler defaults to `/components/`, requires `/specs`, and the
page catalog stores a component-only field.

### Required before the first Token Forge-bound snapshot

- [ ] Expand the curated page catalog to include component and foundation/style
  token-source pages.
- [x] Replace component-only catalog metadata with explicit page metadata such
  as `area`, `subject`, and `pageType`.
- [x] Allow discovery and catalog validation for the supported Material style
  and foundation URL patterns.
- [x] Keep the catalog limited to reviewed pages expected to expose token-table
  payloads. Do not add every Material documentation page.
- [ ] Add representative foundation/style fixtures and confirm the generic
  extractor handles them without component-specific parser branches.
- [ ] Recrawl the expanded catalog and create a new immutable snapshot. Do not
  modify the existing component-focused snapshots.
- [ ] Review whether the expanded snapshot contains the foundation records
  Token Forge needs for `md.ref`, `md.sys`, and component references.

Static recommendation tables and explanatory motion tables remain separate page
evidence. They must not be merged into the generated token graph unless a
separate evidence contract is approved.

## 2. Snapshot contract and schemas

### Done

- [x] Versioned Go types exist for the published snapshot documents.
- [x] Standalone JSON Schemas exist for the manifest, generated token graph,
  source-payload details, page results, and lifecycle notices.
- [x] Snapshot artifacts include `lifecycle-notices.json`.
- [x] Snapshot verification checks artifact hashes and checksums.
- [x] Snapshot IDs are immutable.

### Open: Source contract gate

- [ ] Validate committed snapshot documents with a JSON Schema validator in
  tests or CI. Parsing the schema files as JSON is not enough.
- [ ] Rebuild the lifecycle snapshot twice with a fixed snapshot ID and
  creation time. Confirm byte-identical output for every artifact.
- [ ] Check that schema changes have clear version rules. Additive lifecycle
  evidence currently uses its own schema version so old snapshot IDs remain
  immutable.
- [ ] Review whether the manifest should list a role for each artifact. The
  current Go output does not have that field, although the draft documentation
  shows one.

## 3. Capture status and unavailable source data

Source must keep these facts separate:

- page was opened successfully;
- token-table URL was observed;
- source payload was fetched or reused;
- source payload was unavailable or failed;
- Material explicitly marked something as deprecated or not recommended;
- Forge chose to support or reject it.

### Decision

- [x] A 404 is a source-availability result, not a lifecycle result.
- [x] A failed current attempt must not overwrite a content-addressed,
  previously successful payload artifact.
- [x] Source capture remains strict: the Snackbar exception keeps
  `capture.complete` false.

### Open: Forge acceptance gate

- [ ] Decide whether Forge may pin an incomplete snapshot when the missing
  payload is a reviewed, known upstream exception.
- [ ] If yes, add an explicit accepted-exception entry to the Forge lock. Do
  not change Source `capture.complete` to hide the exception.
- [ ] Decide whether a later successful Snackbar payload should create a new
  snapshot and a new Forge lock, or only update the next routine snapshot.

## 4. Lifecycle notices and deprecation

`lifecycle-notices.json` currently contains explicit notices found in source
payload records. The latest snapshot contains:

- 1,667 `deprecated` notices;
- 7 `not-recommended` notices;
- token, style, and record scopes.

### Rules

- [x] Preserve the raw message.
- [x] Preserve source payload and source record identifiers.
- [x] Keep `deprecated` separate from `not-recommended`.
- [x] Do not infer `active` from the absence of a notice.
- [x] Do not infer deprecation from a 404, missing payload, or missing table.
- [ ] Review the 458 `record`-scoped notices for useful source kinds and
  possible false positives.
- [ ] Decide whether messages that name replacement tokens need a separate
  relationship record or whether the raw message is enough for Phase 1.

### Component-level notices

Rendered page notices can say that a whole component is no longer recommended.
The token payload may not contain that page message. The current model can
represent such a notice, but the crawler does not yet capture page text as a
separate evidence document.

- [ ] Decide whether to capture rendered page notices before the first Forge
  compiler milestone.
- [ ] If deferred, record the limitation: Source reports payload-local
  lifecycle evidence, while component-level page guidance can remain unknown.

Recommendation: defer page-text capture for the first compiler milestone. Do
not infer a component lifecycle state from token data alone.

## 5. Baseline and Expressive

### What the current snapshot shows

The captured payload for:

```text
https://m3.material.io/components/navigation-bar/specs
```

is one token-table JSON payload:

```text
TOKEN_TABLE.52865380210a9b69.json
sha256: 6f20e030df1f8a6b568204c71e7c453755e1a60800a5dc4f56af477cbfd29a0e
```

It contains four token sets, including:

- `md.comp.navigation-bar` with display name `Navigation bar (baseline)`;
- `md.comp.nav-bar` with display name `Nav bar - Common`;
- `md.comp.nav-bar.item.horizontal`;
- `md.comp.nav-bar.item.vertical`.

The payload also contains an explicit description on the baseline token set:

```text
For baseline navigation bar. For expressive, use `md.comp.nav-bar` and
corresponding `.item` token sets instead.
```

The same payload contains an upstream `motion.scheme` context group with
`standard` and `expressive` tags. The current graph preserves these objects in
the loss-preserving `raw` records, but it does not create a normalized
baseline/Expressive field or connect every token to a resolved Expressive
context.

### Current answer

Source Loom does not currently distinguish baseline and Expressive as a first-
class normalized dimension. It preserves enough raw evidence to investigate
the distinction, but Forge cannot yet query it reliably without knowing the
upstream relationships.

### Recommended model

- [ ] Preserve upstream context groups and tags as generic context data.
- [ ] Preserve token-set display names and descriptions.
- [ ] Add a normalized context relation only when the source provides a direct
  relationship.
- [ ] Do not add a simple `isExpressive` boolean. It would hide other context
  dimensions such as theme, platform, audience, contrast, and motion scheme.
- [ ] Keep `baseline` and `Expressive` as source labels or context tags, not as
  Forge support decisions.
- [ ] Add tests for the navigation-bar payload and similar pages.

The preferred shape is a generic context relation, for example:

```json
{
  "dimension": "motion.scheme",
  "tag": "expressive",
  "displayName": "Expressive",
  "sourceRecordId": "..."
}
```

The relation must include enough source identity to show why it was created.
Source Loom must not guess a relation only because a token name contains a word
such as `expressive`.

## 6. Static recommendation and availability tables

The Material pages also contain static tables that compare variants,
configurations, baseline availability, Expressive availability, and
recommendations.

These tables are not token-table payloads. They answer a different question:

> Which variant or configuration does Material recommend or expose?

They should not be merged into the generated token graph.

### Recommended boundary

- [x] Keep token values and references in the generated token graph.
- [x] Keep lifecycle evidence separate from Forge support policy.
- [ ] Decide whether Source should capture static recommendation tables as a
  separate `page-guidance.json` evidence document.
- [ ] If deferred, let Forge use manual review for the first candidate set.

Recommendation: defer static-table capture for the first compiler milestone,
but keep the decision on this list. The tables can later provide evidence for
variant availability and replacement recommendations without changing the
token graph contract.

## 7. Graph review before Forge

Review representative records from the lifecycle snapshot:

- [ ] `md.ref` values and unresolved references;
- [ ] `md.sys` roles and theme contexts;
- [ ] `md.comp` component tokens;
- [ ] baseline and Expressive navigation-bar records;
- [ ] light and dark values;
- [ ] token-set and component descriptions;
- [ ] deprecated tokens with replacement messages;
- [ ] duplicate records and repeated source values;
- [ ] records with missing or failed payloads.

The review must answer whether Forge can make decisions from the snapshot
alone. It must not require crawler internals or a live Material page.

## 8. Forge policy before the compiler

The compiler must not decide policy by accident. Define these rules first:

- [ ] Which `md.ref`, `md.sys`, and `md.comp` records can become candidates?
- [ ] How does Forge select light and dark values?
- [ ] How does Forge represent other context dimensions?
- [ ] Are deprecated source tokens retained as candidates?
- [ ] Are deprecated source tokens blocked from canonical promotion by default?
- [ ] Does `not-recommended` produce a warning, an exclusion, or a manual
  review item?
- [ ] How are replacement-token messages represented?
- [ ] What happens when a component is current but one style is deprecated?
- [ ] What happens when a component is not recommended but remains live?
- [ ] Which selected Expressive revisions are in scope?

## 9. Required commits and locks

- [ ] Review and commit the current schema and lifecycle-notice work in Source
  Loom.
- [ ] Select the provisional baseline snapshot.
- [ ] Update the Forge source lock to that snapshot ID and manifest digest.
- [ ] Update the Forge review candidate to reference the selected snapshot.
- [ ] Keep the lock status provisional until the coverage and context policy is
  approved.

## Compiler gate

Do not begin the Forge compiler until these conditions hold:

1. The published Source documents pass their standalone schemas.
2. A fixed-input snapshot rebuild is byte-identical.
3. The baseline/Expressive evidence model is either implemented or explicitly
   deferred with a documented limitation.
4. The Snackbar exception policy is recorded in the Forge lock or rejected.
5. Lifecycle notice semantics are reviewed.
6. Light/dark and other context-selection rules are approved.
7. The first candidate mapping has been reviewed by a person.
