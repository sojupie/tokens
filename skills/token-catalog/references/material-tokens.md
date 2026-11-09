# Read Material Design 3 tokens

The JSON data endpoints below use GET. MCP uses POST.
No account, API key, cookies or JavaScript are required to read the data.
The service is read-only. It provides browser tools where WebMCP is supported.
No payment or authentication protocols are required.
Read {{PUBLIC_ORIGIN}}/openapi.json for request parameters and response schemas.

## Find a component or style

Fetch {{PUBLIC_ORIGIN}}/api/v1/subjects. The subjects array contains kind, slug, name,
tokenCount, sourcePageUrls, url and jsonUrl. Follow jsonUrl rather than guessing a path;
style slugs can contain slashes. Follow url for the human-readable token table.
The older /api/v1/components endpoint returns the same index, including styles.

## Read captured JSON

- {{PUBLIC_ORIGIN}}/downloads/tokens.json?scope=components includes components only.
- {{PUBLIC_ORIGIN}}/downloads/tokens.json?scope=styles includes all non-component subjects.
- {{PUBLIC_ORIGIN}}/downloads/tokens.json?scope=all includes both.

Collection downloads return `capturedAt` and `subjects`. The subject index's
`jsonUrl` returns `capturedAt`, `kind`, `slug`, `name`, and `tokens`. Change
the scope on that URL to narrow the result.

| Subject JSON scope | Parameters that affect the result | What it includes |
| --- | --- | --- |
| `view` | `set`, `q`, `context.<dimension>` | The selected table and search results, with values for the selected context. |
| `set` | `set` | The whole table, with all captured contexts. |
| `all` | None | Every token in the subject, with all captured contexts. |

For example, [read the primary color roles in the Dark context]({{PUBLIC_ORIGIN}}/downloads/style/color.json?scope=view&set=md.sys.color&q=primary&context.Theme=Dark).
Use the exact `set` key, such as `md.sys.color` for Color Schemes or
`md.ref.palette` for Palettes. Use `q` for the table search and
`context.<dimension>` for context filters. A token can have an empty `values`
array when there is no matching value.

## Export selected tokens

Send a GET request to `/downloads/export` using one of these combinations.
See the [OpenAPI description]({{PUBLIC_ORIGIN}}/openapi.json) for the full request.

| Content | Scope | Subject and table parameters | Formats | Context choices |
| --- | --- | --- | --- | --- |
| `tokens` (default) | `view` | Require `kind` and `slug`. Use `set` to choose a table and `q` to filter it. | `dtcg`, `css`, `source`, `report` | Ask for `options=1`. |
| `tokens` | `set` | Require `kind` and `slug`. Use `set` to choose a table. | `dtcg`, `css`, `source`, `report` | Ask for `options=1`. |
| `tokens` | `subject` | Require `kind` and `slug`. | `dtcg`, `css`, `source`, `report` | Ask for `options=1`. |
| `tokens` | `collection` | Omit subject and table parameters. | `dtcg`, `css`, `source`, `report` | Ask for `options=1`. |
| `implementation` | `subject` | Require `kind=component` and the slug of a component with a playground. | `zip` | Theme and Contrast only. |
| `implementation` | `collection` | Omit subject and table parameters. | `zip` | Theme and Contrast only. |

The table scopes use the same `set` keys as captured JSON. Omit `set` to use
the default table. Use `set=all` with `scope=view` to search across all tables
in one subject. Each formatted export resolves one context per selection,
including for `scope=set`.

| Format | What the file contains |
| --- | --- |
| `dtcg` | DTCG 2025.10 tokens and required alias targets. The document's `$extensions.dev.tokencatalog.export` includes conversion and selection metadata. |
| `css` | CSS custom properties with their original names and required alias targets. Requires generated CSS assets. |
| `source` | Captured values and references for the selected context, grouped by subject. |
| `report` | Conversion and selection metadata as a separate JSON file. |

Start with [a DTCG export for Buttons]({{PUBLIC_ORIGIN}}/downloads/export?kind=component&slug=buttons&scope=subject&format=dtcg).
1. Send the token scope and any subject or table parameters with
   [`options=1`]({{PUBLIC_ORIGIN}}/downloads/export?kind=component&slug=buttons&scope=subject&options=1).
   The JSON response maps each active context dimension to its allowed values.
2. Pass a choice as `context.<dimension>`, for example `context.Platform=iOS`,
   and request `options=1` again. The response accounts for the other choices
   you supplied, so changing a platform can change the list of dimensions or values.
   Remove choices that no longer apply. Omitted choices use defaults, including
   Light, Default contrast, Web, 3P, and Static font.
3. Remove `options=1`, add the desired `format`, and send `summary=1` to see
   the counts without downloading. Some tokens can lack a value even when the
   chosen context is available. Check the omission count or conversion report.
   Remove `summary=1` to download.

A single selection returns one file. Repeat `format` or `context.<dimension>`
to request up to 64 combinations. A batch returns a ZIP with one folder per
combination, conversion reports, and `selections.txt` mapping folders to settings.
For example, [preview the counts for Buttons in two formats and two themes]({{PUBLIC_ORIGIN}}/downloads/export?kind=component&slug=buttons&scope=subject&format=dtcg&format=css&context.Theme=Light&context.Theme=Dark&summary=1).
Remove `summary=1` from that request to download the ZIP.

## Download runnable components

Request [a Buttons implementation ZIP]({{PUBLIC_ORIGIN}}/downloads/export?kind=component&slug=buttons&scope=subject&content=implementation&format=zip)
or use `scope=collection` for all eligible component examples. Subject exports
require a component with a playground. Pass its `preview.*` options to choose
the example configuration.

The ZIP includes example HTML, CSS, JavaScript, dependencies of nested components,
and assets needed by the example. Extract it and serve the directory over HTTP.
The CSS uses the catalog's Web theme and includes all six light, dark, and contrast
appearances. `context.Theme` accepts Light or Dark. `context.Contrast` accepts
Default, Medium contrast, or High contrast. These parameters set the initial
appearance. Repeat either to bundle more than one starting appearance.
Use `summary=1&format=zip` to inspect an implementation selection. The
`options=1` lookup applies to token exports.

## Interpret token values

Each token has name, optional displayName and type, and a values array.
Each value contains structured value data, optional context labels, default and
references. Preserve references and context labels. Do not replace references with
theme-specific hex colors or assume that every token resolves to a scalar.
A missing resolved value remains unavailable. In a typed source dimension with a
known unit, an omitted numeric value field represents zero. A shape is one compound
value whose named corners override defaultSize. Keep those corner overrides when
using the token. The optional presentation object preserves source token-set
membership, display groups and order. An absent context does not establish a context.
Use sourcePageUrls to cite Material's source pages and capturedAt for freshness.
This site captures source data; it is not an official Google service.

{{DTCG_DOWNLOADS}}## Read history

Fetch {{PUBLIC_ORIGIN}}/api/v1/releases for current and retained snapshots.
Fetch {{PUBLIC_ORIGIN}}/api/v1/changes for publicationId, snapshot and changes.
Filter changes locally by subject and snapshotId. The response is not paginated.
Each record includes detectedAt, type, name, kind, previousSnapshotId and snapshotId,
with before and after where available. Before and after are strings, which can
contain JSON. Dates show when the crawler observed a change, not its release date.
One token can change in multiple contexts; a record count is not a unique-token count.

## Use MCP

Read [the MCP server card]({{PUBLIC_ORIGIN}}/mcp/server-card) for the transport URL
and supported protocol versions. Send JSON-RPC requests to {{PUBLIC_ORIGIN}}/mcp
using POST with Content-Type: application/json and
Accept: application/json, text/event-stream. Initialize the connection, then use
tools/list to discover list_subjects, get_tokens and list_changes.
Use MCP-Protocol-Version: {{MCP_PROTOCOL}} for subsequent requests.
This server returns JSON responses and does not require a session ID.

- `list_subjects` accepts an optional `query` to match the name or slug. Its
  results include the page URL and `jsonUrl`.
- `get_tokens` takes the exact `kind` and `slug` and returns the entire subject
  in all captured contexts. Use the HTTP downloads for table filters, context
  selection, conversion formats, and implementation ZIPs.
- `list_changes` accepts an optional exact `subject` slug. Filter by snapshot
  locally. Results are not paginated.

WebMCP exposes the same tools in browsers that support it.

## Read HTML or Markdown instead

Use the [site sitemap]({{SITEMAP_URL}}) for canonical pages. The catalog returns
server-rendered HTML. Leave HX-Request unset to receive the complete page.
Token rows remain in the HTML when their groups are collapsed.
Send Accept: text/markdown to receive the same filtered view as Markdown with
Content-Type: text/markdown. Browsers receive HTML by default.
Markdown preserves token tables, reference chains and links to documentation,
change history and other pages. It includes every row that matches the selected
set and search. The response lists set links, their keys, and the query parameters
and context values available on that page.
Add `view=table` to a subject or set URL to get the table without the catalog
page. HTML includes the table controls. Markdown returns just the token groups
and rows. The full Markdown response links to this view with the current set,
search and context filters.
Prefer one subject's JSON over repeated page scraping or a full-site download.
Reuse the subject index and discovery files rather than fetching them for every token.
