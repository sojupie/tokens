# Contributing

[Development](docs/development.md) covers setup, checks and the code layout.
The [documentation index](docs/README.md) links to CLI and deployment guides.

Keep changes focused and describe the behavior they change. Include relevant
checks and any effect on saved snapshots or published downloads.

## Source data

Review source-page additions and removals before changing
`inputs/page-catalog.json`. A failed request does not establish that an upstream
page or token was removed. Preserve source URLs, capture status and checksums.

Keep incomplete captures marked as partial. Use small fixtures for tests and
keep generated captures, credentials and publications out of routine commits.
The snapshot format documents which files belong in an immutable bundle.

## Interface changes

Preserve token contexts and references in filters and downloads. Check HTML,
JSON and agent interfaces when changing their shared data. Regenerate templ
output after editing templates and include it in the same change.

Write documentation in plain language. Use established domain terms, explain
concrete behavior, and keep commands in the guide that owns them. Historical
plans belong under `docs/history/` with their date and status.

## Licenses

Original contributions are covered by the repository's MIT license. Preserve
upstream copyright and license notices when adding third-party code or data.
Record the source and license in [third-party notices](THIRD_PARTY_NOTICES.md).
Do not apply the project's MIT license to someone else's design-system data.
