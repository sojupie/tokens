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

## Choose a download

- {{PUBLIC_ORIGIN}}/downloads/tokens.json?scope=components includes components only.
- {{PUBLIC_ORIGIN}}/downloads/tokens.json?scope=styles includes all non-component subjects.
- {{PUBLIC_ORIGIN}}/downloads/tokens.json?scope=all includes both.

Collection downloads contain capturedAt and subjects. A subject's jsonUrl returns
capturedAt, kind, slug, name and tokens. Both include all supported contexts.
A token is a named item; one token may have many values for different contexts.
For a smaller export, use the page's Download JSON form: view applies its table,
search and context filters; set includes the selected table in all contexts;
all includes every token in that subject. Copy the form's exact set and context
parameters. Do not invent contexts that the subject does not support.

For example, a filtered table can contain 7 tokens and 7 values. The same table
with scope=set can contain 7 tokens and 153 values. The whole subject with
scope=all can contain 13 tokens and 159 values across two tables. These counts
are illustrative; the HTML download previews show exact counts for the current
snapshot and filters. scope=all on /downloads/tokens.json means the entire site,
whereas scope=all on a subject download means only that subject.
MCP and WebMCP get_tokens always return the entire subject with every context;
use the filtered HTTP download endpoint when you want scope=view or scope=set.

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
Use exact kind and slug values from list_subjects when calling get_tokens.

## Read HTML or Markdown instead

Use {{PUBLIC_ORIGIN}}/sitemap.xml for canonical pages. The site returns server-rendered
HTML. Leave HX-Request unset to receive a complete page, not an HTMX fragment.
Token rows remain in the HTML when their groups are collapsed.
Send Accept: text/markdown to receive the same filtered view as Markdown with
Content-Type: text/markdown. Browsers receive HTML by default.
Markdown preserves token tables, reference chains and links to documentation,
change history and other pages. Query filters apply to both representations.
Prefer one subject's JSON over repeated page scraping or a full-site download.
Reuse the subject index and discovery files rather than fetching them for every token.
