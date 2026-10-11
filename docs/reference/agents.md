# APIs and agent access

The Material Design 3 collection is served at `https://tokencatalog.dev/md3`.
Routes below are relative to `/md3`, unless stated otherwise. The public API
and MCP tools provide read-only access without accounts or credentials.

Start with [/md3/agents.json](https://tokencatalog.dev/md3/agents.json),
[/md3/llms.txt](https://tokencatalog.dev/md3/llms.txt), or the
[OpenAPI schema](https://tokencatalog.dev/md3/openapi.json).

```sh
curl https://tokencatalog.dev/md3/api/v1/subjects
curl https://tokencatalog.dev/md3/api/v1/releases
curl https://tokencatalog.dev/md3/api/v1/changes
curl https://tokencatalog.dev/md3/ -H 'Accept: text/markdown'
```

Subject responses include a `jsonUrl` for the token download. The usage skill
explains references, contexts and capture dates.

## Interfaces

| Feature | Route or behavior |
| --- | --- |
| Markdown negotiation | HTML pages return `text/markdown; charset=utf-8` when explicitly preferred through `Accept`. HTML remains the default. Both variants include `Vary: HX-Request, Accept`. |
| Content Signals | `robots.txt` and the `Content-Signal` response header declare `ai-train=no, search=yes, ai-input=yes`. |
| API catalog | `/.well-known/api-catalog`, `application/linkset+json`, links to OpenAPI, the usage skill and health. |
| Agent skill index | `/.well-known/agent-skills/index.json`, v0.2.0 schema and SHA-256 of the exact served skill bytes. |
| ARD / AI catalog | `/.well-known/ai-catalog.json`, public CORS, API, skill and MCP entries with representative queries. |
| MCP | `/mcp`, stateless Streamable HTTP, JSON responses, `initialize`, `ping`, `tools/list`, `tools/call`, and notification acceptance. |
| MCP server card | `/mcp/server-card`, current SEP-2127 shape; `/.well-known/mcp/server-card.json` also includes legacy scanner fields. |
| WebMCP | Feature-detected browser registration on page load: `list_subjects`, `get_tokens`, `list_changes`. Unsupported browsers keep the existing UI. |
| Access instructions | `/auth.md` describes public, credential-free access. |

MCP exposes the same three read-only tools as WebMCP. `get_tokens` requires a
subject's exact `kind` and `slug` from `list_subjects`. All contexts are returned.
The server advertises only tools, with no resources, prompts, server events or
sessions. MCP clients must send `Accept: application/json, text/event-stream`
and `Content-Type: application/json`. Supported protocol versions are
2025-11-25, 2025-06-18 and 2025-03-26. Browser MCP requests must have the configured
public origin; native clients may omit Origin. The API remains public with `*`
CORS, while MCP allows same-origin browser use.

Markdown comes from the rendered page, preserving table selection, search,
contexts and pagination. It omits navigation, scripts and interactive controls.
No `x-markdown-tokens` count is claimed because no tokenizer is configured.

OAuth/OIDC discovery, OAuth Protected Resource Metadata and registration
metadata are intentionally absent: there is no issuer, protected API or account
provisioning flow. `/auth.md` documents this but does not claim Auth.md agent
registration support. Scanners may continue reporting these checks as absent.
Adding discovery metadata alone would advertise authentication that does not exist.
No A2A agent is implemented or advertised.

The content policy permits search and agent input, and declines model training.
Deployment configuration is covered in [Cloudflare setup](../operations/cloudflare.md).


## URL layout

`https://tokencatalog.dev/md3/` is the canonical Material Design 3 collection.
All pages, static assets, API requests, downloads, OpenAPI and MCP use `/md3/`.
The root `/` redirects to `/md3/`. Unprefixed component and download routes are
not served. `/robots.txt`, `/sitemap.xml`, `/agents.txt`, `/agents.json`,
`/llms.txt`, `/llms-full.txt`, `/openapi.json`, `/auth.md`, `/skills/` (the published
skill file only), and the supported `/.well-known/` files also work at the
root so discovery clients can find the collection. Their contents advertise
canonical `/md3/` URLs. `/api/v1/health` remains a root alias for internal checks.
The same discovery files are available under `/md3/` too.

