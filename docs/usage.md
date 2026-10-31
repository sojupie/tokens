# Using Token Catalog

The hosted collection is [Material Design 3](https://tokencatalog.dev/md3/).
It includes component and style tokens with their source references and
context-specific values.

## Component playgrounds

Component pages include a playground for interacting with supported variants,
sizes and states. Use its controls to change the examples and compare them.

## Tables and settings

Open a component or style, choose a table, and select the settings relevant to
your implementation. A single token can have different values for platform,
theme, contrast, size and other contexts. The available settings come from the
source data and vary by table.

Search narrows the rows. Context filters narrow the displayed values. References
show which other tokens a value depends on. Follow the source links when you
need the design guidance behind a token.

## Downloads

**Download** opens a dialog for design tokens or component implementations.
On a component or style page, token exports default to that subject. The download
button in the token table defaults to its current filtered rows. Other scopes
include the whole table and the whole collection.

Token formats are DTCG JSON, CSS variables, source JSON, and a conversion report.
Context choices start from the page's selected values and can be changed in the
dialog. Select several formats or context values to download applicable combinations
in one ZIP, with a selection manifest. Single selections download directly.
DTCG and CSS exports include required
alias targets, even when those targets are outside the chosen table. The dialog
shows selected tokens, exported tokens, dependencies, and omissions separately.
Source JSON retains captured values and reference metadata for the selection.
Context controls use defined captured combinations, including required token
references. Settings that do not apply are omitted. Upstream labels are preserved;
for example, some `Font: iOS bold` values are also defined for Web.

Component implementations download as ZIP archives for a whole component or the
collection. Each archive contains example HTML, authored and generated CSS,
required browser scripts, nested component assets, fonts, and license notices.
Serve the extracted directory with a static HTTP server and open `index.html`.
The implementation includes the catalog's Web theme and its six appearance
combinations; it differs from a raw captured-token export. Example actions do not
implement application logic. Implementation scope and context choices reflect
these limits in the dialog.

DTCG and source JSON exports work without generated assets. Scoped CSS requires
the generated CSS value index; implementation archives require component themes.
The existing download URLs remain available.

See [DTCG exports and CSS](reference/dtcg.md) for conversion rules and CLI exports.
Source data retains its [upstream license](licensing.md).

Older custom JSON download URLs remain available for compatibility:

| Scope | Included data |
| --- | --- |
| Filtered table | The selected table's rows and values matching search and settings |
| Complete table | Every token and context in the selected table |
| Component or style | Every table and context for the subject |
| Collection | Every published component and style with all contexts |

Only the filtered-table scope applies search and context filters. These older
files use the Catalog model; their schema has not changed to DTCG.

## Comparisons

Each publication compares retained snapshots in chronological order. The first
snapshot is a baseline. Later snapshots contribute added, removed or changed
records to the change log. One token can have changes in several contexts, so
change counts are not counts of distinct token names.

Dates describe observations, not the source's exact release dates. Updated
payload URLs or hashes alone do not count as token changes. A complete capture
can contain the same values as an earlier one.

Partial captures have missing source data. Apparent additions or removals can
reflect those gaps. The change log warns when retained history includes a
partial capture.

## API and agents

[API and agent access](reference/agents.md) describes the public JSON endpoints,
OpenAPI schema, Markdown responses and MCP tools. Reads do not require an
account. The same capture dates and coverage limits apply to these interfaces.
