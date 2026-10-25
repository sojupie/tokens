# Curating component guidance

Maintain one active file per component. Create the archive on the first retirement:

```text
forge/<design-system>/guidance/<component>/checks.json
forge/<design-system>/guidance/<component>/retired-checks.json  # optional
```

Commit these files. Keep captured documentation in the versioned snapshot directory,
which stays outside Git. Use Git history for earlier file versions; do not create
separate dated check files or copy the current file into reports.

The format is defined by `schemas/component-guidance.schema.json`.
The archive uses `schemas/retired-guidance.schema.json`. It contains the component
identity and complete retired checks; source receipts remain in each check’s
history. It does not duplicate the current source inventory.

## From capture to checks

1. Capture the component's documentation pages with `tokens guidance`. Include
   overview, specs, guidelines and accessibility for Material components. Record
   the receipt path, Markdown hash and extractor version for every source page.
2. Inventory the page headings in `sources[].sections`. Include the introductory
   content as `page`. Use the enclosing source region ID for a heading where
   available; do not use duplicated editor IDs such as `isPasted`.
3. Read every section, including tables, captions and linked media. Compare the
   Markdown with the rendered HTML. Inspect diagrams when they carry requirements.
   Capture linked pages when they define a requirement needed for interpretation.
4. Write independently assessable instructions in `checks`. Preserve conditions,
   exceptions, platform limits, recommendations and optional behavior. Link each
   instruction to its source page, section and captured revision.
5. Mark the section `mapped` after all its actionable instructions have checks.
   Use `no-requirements` with a reason for descriptive or irrelevant content.
   Use `unclear` with a reason when interpretation or missing media prevents
   completion. Leave unread or partly processed sections `pending`.
6. Assess implementation separately, then validate and commit the file.

A mapped section means its instructions have been curated. It does not mean they
are implemented. Partial coverage stays visible in the section inventory.

## Writing a check

Use a stable semantic ID, such as `leading-edge-placement`. Keep it when the
wording or source location changes. Do not encode ordering, dates or status in IDs.

`statement` states the expected behavior. `condition` states when it applies.
Split instructions that need different conditions or independent results. For
example, text scaling and keyboard traversal need separate checks. Do not split a
single inseparable behavior solely to inflate the number of checks.

Use `basis: explicit` for instructions stated by the source. Use `interpretation`
when deriving a check from examples or broader guidance, and explain the reasoning
in `notes`. Keep project preferences outside this source-guidance file. Upstream
platform availability describes the upstream implementation, not our coverage.

Preserve “should”, “must”, “may”, exceptions and recommendation changes in the
statement. Do not turn a suggestion into a mandatory rule or treat a superseded
configuration as the preferred implementation. Distinguish product integration
rules from behavior that an isolated component can establish.

Read tables with their headers and row relationships intact. A lost table layout
or an image description is insufficient evidence for an availability mapping or
numeric measurement. Mark the relevant section unclear until inspected.

For keyboard and accessibility-label tables, compare every row with the checks.
Retain each key/action mapping, label/role mapping and platform condition. A broad
instruction such as “support keyboard navigation” does not cover a shortcut table.
When another check already covers a row, identify that check in the section notes.

Before calling a table empty, inspect its Markdown rows and captured HTML. Missing
rows in an extraction are a capture problem, not evidence that the source has no
requirements. Mark the section unclear and recapture if the saved evidence is
insufficient. Use a specific reason for each excluded section rather than a list
of possible reasons copied across unrelated sections.

## Source history

`sourceHistory` is append-only. Its array position is the one-based source revision
used by assessments and fixes. Each event records the source URL, section,
capture receipt, Markdown hash, extractor version, date and a concrete summary.
The current statement is kept at the top of the check; Git preserves prior wording.

| Change | Use |
| --- | --- |
| `added` | First record of this instruction. This does not claim it was newly added upstream. |
| `changed` | The source instruction, condition or source location changed. Describe the difference. |
| `removed` | Confirmed removal from upstream guidance. Retain the check and its last statement. |
| `restored` | A removed or withdrawn check becomes active again. |
| `corrected` | Our curation or interpretation was wrong. Explain the correction without claiming an upstream change. |
| `withdrawn` | Our check was erroneous or duplicated another check. State why and retain its ID. |

A missing page, failed capture or missing Markdown block does not establish source
removal. Check the HTML, replacement location and upstream page before recording
`removed`. Preserve uncertainty when the available evidence is insufficient.

When a new capture changes a page, replace its source metadata and mark affected
sections pending for review. Compare against the previous capture. A hash change
identifies content that needs review; it does not identify a changed guideline.
Editorial changes need no check revision. A changed instruction needs a history
event, updated statement/condition and an assessment against the new revision.
A relocated instruction keeps its check ID. Refresh section references and record
the provenance change explicitly.

After confirming a removal, append `removed` and move the complete check to
`retired-checks.json`. Use `withdrawn` for an erroneous or duplicate check, with
an explanation. Preserve its statement, assessment, fix commits and earlier
source events. Commit both file changes together. An ID occurs in exactly one
file. Do not create an empty archive before it is needed.

To restore a check, move it back to `checks.json`, append `restored`, and reassess
or reset the implementation at that revision. Keep its original ID. Retired
records remain unchanged until restored. Removal is not an implementation fix.

## Implementation assessment

| Status | Required evidence |
| --- | --- |
| `unassessed` | No assessment commit or evidence. Use for new or stale assessments. |
| `implemented` | Exact assessed commit or baseline and code location explaining how the condition is handled. |
| `missing` | Exact assessed commit or baseline and code evidence identifying the absent required behavior and inspected scope. |
| `verified` | Exact assessed commit or baseline and a test or browser result that exercises the condition. |
| `failing` | Exact assessed commit or baseline and a test or browser result showing the unmet expectation. |
| `not-applicable` | Reason explaining which condition or platform does not apply. |
| `unclear` | Reason explaining the unresolved source interpretation or implementation evidence. |

Evidence records have `kind` (`code`, `test` or `browser`), `location` and `result`.
Record the command or reproduction steps and the observed outcome. Browser
results must include the relevant browser, configuration and environment. Use
repository-relative code/test locations and stable URLs for shared artifacts.
Include reproduction steps so another person can repeat the check. Record
screenshots or saved captures where useful. A code reference alone does not
establish runtime correctness. Passing unrelated tests does not verify a check.

Use `assessedCommit` for an implementation inspected at a committed revision.
For the initial assessment of a dirty working tree, use `assessedBaseline` instead.
Create one shared baseline with `tokens guidance baseline`. It snapshots tracked
and nonignored new files through a private Git index without changing the real
index, working files or branch. The generated registry is
`forge/<design-system>/guidance/baselines.json`; each ID is the captured Git tree.
The local archive and its SHA-256 preserve the inspected source outside Git.
Keep the registry with the assessments and retain the archive for reproduction.

Every assessment in an initial pass should reference the same baseline. Existing
code is baseline evidence, not a new fix. Leave `fixes` empty unless an actual
implementation change is made later. Do not reconstruct historical fix commits.

`implemented` means code supports the rule; it does not establish rendered or
interactive correctness. `missing` identifies a required behavior absent from the
inspected code. Use `verified` and `failing` only for exercised tests or browser
observations, with exact reproduction details. Preserve uncertainty when either
the source instruction or implementation evidence is insufficient.

An active check's `implementation.sourceRevision` must match its current history
length. Following a source change, reassess it or reset it to `unassessed` at the
new revision. Historical assessments remain available through Git.

Append each implementation fix to `fixes` with its full commit ID, the source
revision it addresses and a concrete summary. A fix commit may address several
checks. Keep multiple attempts if a later commit corrects an incomplete fix.
Commit the implementation first, then record its full commit ID and commit the
updated checks file. The metadata commit is separate because a commit cannot
contain its own final hash.

Recording a fix does not promote the assessment to `verified`; verify the actual
implementation commit separately. Do not invent commits for pre-existing behavior.

## Validation

Run from the repository root:

```sh
go run ./cmd/tokens guidance validate
# Before merging a branch, compare with its base as well:
go run ./cmd/tokens guidance validate --against origin/main
```

Use `--file forge/material/guidance/navigation-rail/checks.json` for one component.
The command checks the schema, canonical path, unique IDs, source/section links,
lifecycle order, assessment revisions, required evidence and locally available
commit IDs. Against the selected Git revision, it rejects deleted component files,
checks missing from both files, rewritten source/fix history and changed
statements without a history event.

The normal Go test suite also validates the schema and internal references of
every canonical component file. Git-history and commit-existence checks run through
the CLI.

Validation is offline. It does not execute evidence, establish the accuracy of
curation, inspect captured files or prove that a component conforms. Review source
and implementation evidence as part of the same change. There is no approval gate
or separate per-run compliance report.
