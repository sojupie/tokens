# Curating component guidance

Maintain one active file per component (plus an archive created on the first
retirement):

```text
forge/<design-system>/guidance/<component>/checks.json
forge/<design-system>/guidance/<component>/retired-checks.json  # optional
```

Commit these files; keep captured documentation snapshots in
`snapshots/material-guidance/` outside Git. Schemas:
`schemas/component-guidance.schema.json` and
`schemas/retired-guidance.schema.json`.

## Workflow

1. Capture the component's documentation pages with `tokens guidance` (overview,
   specs, guidelines, and accessibility for Material) and record each page's
   receipt path, Markdown hash, and extractor version.
2. Inventory page headings in `sources[].sections` (use `page` for intro
   content; prefer enclosing region IDs over duplicate editor IDs like
   `isPasted`).
3. Read every section—including tables, captions, and diagrams—comparing
   Markdown with the captured HTML. Capture linked pages when they define
   requirements.
4. Write independently testable entries in `checks`, linked to their source
   page, section, and revision.
5. Set each section status:
   - `mapped` — all actionable instructions have checks (curated, not
     necessarily implemented).
   - `no-requirements` — descriptive or irrelevant content, with a specific
     reason.
   - `unclear` — missing media, broken extraction, or ambiguous spec, with a
     reason.
   - `pending` — unread or unfinished.
6. Assess implementation separately, validate, and commit.

## Writing checks

- **IDs:** Stable semantic slugs (e.g. `leading-edge-placement`) with no
  ordering, dates, or status. Keep the ID if wording or location changes.
- **Scope:** Put the expected behavior in `statement` and when it applies in
  `condition`. Preserve modal verbs (`must`, `should`, `may`), exceptions, and
  platform scope. Split rules that have different conditions or pass/fail
  independently; keep a single inseparable behavior together.
- **Basis:** Use `explicit` for directly stated rules and `interpretation`
  (with reasoning in `notes`) for rules derived from examples or broader
  guidance. Keep project preferences out of source-guidance files.
- **Tables and diagrams:** Inspect captured HTML before treating a table as
  empty—missing rows may be an extraction bug. Preserve every row of keyboard
  and accessibility-label tables (reference an existing check in section notes
  if already covered).
- **Embedded components:** When a component embeds another primitive (e.g. a
  button), compare any apparent spacing or sizing difference against the
  primitive's own spec before adding a local override.

## Source history

`sourceHistory` is append-only; its 1-based index is the `sourceRevision` used
by assessments and fixes. Each entry records the URL, section, receipt,
Markdown hash, extractor version, date, and summary.

| Change | When to use |
| --- | --- |
| `added` | First record of this check (whether new upstream or newly curated). |
| `changed` | Upstream wording, condition, or section moved. |
| `removed` | Confirmed removed upstream (verify HTML and live page first). |
| `restored` | Previously removed or withdrawn check becomes active again. |
| `corrected` | Fix to our own curation or interpretation. |
| `withdrawn` | Erroneous or duplicate check. |

When a new capture changes a page hash, update source metadata and mark
affected sections `pending`. Ignore pure editorial edits; for real rule
changes, append a history event, update `statement`/`condition`, and reassess.

To retire a check (`removed` or `withdrawn`), append the event and move the
complete record (preserving assessment and fixes) to `retired-checks.json` in
the same commit. Each ID lives in exactly one file. To restore it, move it back,
append `restored`, and reassess.

## Implementation assessment

| Status | Required evidence |
| --- | --- |
| `unassessed` | None (new check or stale after a source revision). |
| `implemented` | Commit/baseline + code location showing the rule is handled. |
| `missing` | Commit/baseline + inspected code scope showing the absent behavior. |
| `verified` | Commit/baseline + test or browser run exercising the condition. |
| `failing` | Commit/baseline + test or browser run showing the failure. |
| `not-applicable` | Reason explaining which condition or platform does not apply. |
| `unclear` | Reason explaining the unresolved spec or evidence gap. |

- **Evidence:** Include `kind` (`code`, `test`, or `browser`), repo-relative
  `location`, reproduction command/steps, and observed `result` (plus browser,
  viewport, and configuration for browser checks). Code inspection supports
  `implemented`, not `verified`.
- **Revisions:** Use `assessedCommit` for committed code, or `assessedBaseline`
  (created via `tokens guidance baseline`, stored in
  `forge/<design-system>/guidance/baselines.json` with a local archive) for an
  initial pass on a dirty tree. Active checks must match `implementation.sourceRevision`
  to the latest `sourceHistory` length.
- **Fixes:** Append real implementation fixes to `fixes` with their full commit
  SHA, addressed `sourceRevision`, and summary. Commit the code first, then
  commit the updated `checks.json` separately (never invent commits for
  pre-existing code). Re-verify after recording a fix.

## Validation

```sh
go run ./cmd/tokens guidance validate
go run ./cmd/tokens guidance validate --against origin/main
go run ./cmd/tokens guidance validate --file forge/material/guidance/navigation-rail/checks.json
```

`validate` checks schemas, canonical paths, unique IDs, source/section links,
append-only history, assessment revisions, and local commit existence (and
against a base ref, rejects deleted files, missing IDs, rewritten history, or
unlogged statement changes). `go test ./...` runs the schema and internal
reference checks automatically.
