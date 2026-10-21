# Using Token Catalog

The hosted collection is [Material Design 3](https://tokencatalog.dev/md3/).
It includes component and style tokens with their source references and
context-specific values.

## Tables and settings

Open a component or style, choose a table, and select the settings relevant to
your implementation. A single token can have different values for platform,
theme, contrast, size and other contexts. The available settings come from the
source data and vary by table.

Search narrows the rows. Context filters narrow the displayed values. References
show which other tokens a value depends on. Follow the source links when you
need the design guidance behind a token.

## Downloads

**Download tokens** opens a dialog with three formats: DTCG JSON, CSS variables,
and the conversion report. Choose Light or Dark and Default, Medium or High
contrast. Each download covers the whole collection in that context. Page,
table and search filters do not apply. The dialog disables downloads when the
publication has no generated assets.

DTCG JSON preserves supported typed values, references and source metadata.
CSS contains custom properties for browser use. The report lists the selected
context, conversion counts, omitted values and values retained only in source
extensions. Read the report when checking conversion completeness. These files
provide design data; they do not contain component implementations.
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
