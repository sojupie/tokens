# Catalog

Catalog serves published token data through HTML, Markdown, JSON, MCP, and
agent discovery documents. It also records request totals and builds usage reports.

`web/` owns the website, templates, filters, downloads, and public interfaces.
`usage/` owns usage retention and reporting. The server reads a saved publication
and embeds Forge's component assets in the same Go binary.

Run the server through `cmd/tokens` from the repository root. See
[development](../docs/development.md) for commands and
[API and agent access](../docs/reference/agents.md) for public routes.

Material Design 3 is currently the only hosted collection. Another collection
will need its own route prefix, publication, presentation mappings, and discovery
entries. The current server loads one publication; multi-collection serving still
requires implementation.
