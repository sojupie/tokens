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

| Option | Included data | Example |
| --- | --- | --- |
| This table, with your filters | The selected table's rows and values matching your search and settings | The matching color tokens for the selected platform and theme |
| This table, with every value | Every token and setting in the selected table | All color values in that table across platforms and themes |
| All tokens for this component or style | Every table and value for the subject | All Buttons tokens, including color, typography and geometry |
| All tokens on this site | Every published component and style with all contexts | The full Material Design 3 collection |

Only the first option applies the current search and filters. The other options
expand the export to their stated scope. A subject's tables can cover different
parts, variants or states, so exporting a whole subject can include much more
than the selected table.

JSON downloads preserve token names, references and context settings. They
provide design data for your own implementation; they do not generate UI
components. Source data retains its [upstream license](licensing.md).

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
