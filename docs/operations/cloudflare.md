# Cloudflare setup

Configure DNS, tunnel routing, DNSSEC, and crawler settings in Cloudflare for
the public interfaces in [API and agent access](../reference/agents.md). Make
sure `TOKENS_PUBLIC_ORIGIN` is set (see [deployment](deployment.md)); discovery
and MCP endpoints return `503` when it is unset.

Examples below use `tokencatalog.dev`; substitute your own hostname for a fork.

## 1. Tunnel routing and DNS-AID records

In **Zero Trust → Networks → Connectors → Cloudflare Tunnels**, route apex
hostname `tokencatalog.dev` (empty subdomain and path) to HTTP `tokens:8080`.
Route the whole hostname without stripping `/md3` so root `/robots.txt` and
discovery files reach the server.

Ensure `@` has a proxied CNAME to `<UUID>.cfargotunnel.com` (no apex A record).
Enable HTTP/2 on the zone, then add these DNS-only discovery records (TTL 3600;
copyable zone file at [deploy/dns-aid.zone.example](../../deploy/dns-aid.zone.example)):

| Type | Name | Priority | Target | Values / Content |
| --- | --- | --- | --- | --- |
| SVCB | `_index._agents` | 1 | `tokencatalog.dev` | `alpn="h2" port=443 mandatory=alpn,port key65400="https://tokencatalog.dev/.well-known/ai-catalog.json"` |
| SVCB | `_mcp._agents` | 1 | `tokencatalog.dev` | `alpn="mcp,h2" port=443 mandatory=alpn,port key65400="https://tokencatalog.dev/md3/mcp/server-card" key65402="mcp"` |
| TXT | `_catalog._agents` | — | — | `url=https://tokencatalog.dev/.well-known/ai-catalog.json` |

## 2. DNSSEC

Enable **DNS → Settings → DNSSEC** and add the DS record at your registrar
(automatic on Cloudflare Registrar). Confirm the status is active and a
validating resolver returns the `ad` flag.

## 3. Public agent and API traffic

- Allow unauthenticated reads (no Cloudflare Access or JS challenges) on HTML
  pages, `/robots.txt`, `/agents.txt`, `/agents.json`, `/llms.txt`,
  `/llms-full.txt`, `/openapi.json`, `/auth.md`, `/skills/`, `/.well-known/`,
  `/api/`, and `/downloads/`.
- Allow `POST` and `OPTIONS` on `/md3/mcp` (`GET` returns `405` by design).
- If **Browser Integrity Check** blocks API clients (e.g. Python `urllib` with
  HTTP 403 / error `1010`), add a Configuration Rule disabling it for
  `http.host eq "tokencatalog.dev"`.
- If Cloudflare Managed `robots.txt` or AI Crawl Control injects content
  signals, match `ai-train=no, search=yes, ai-input=yes` or disable managed
  injection so the response has one consistent policy.

## 4. Caching and Markdown negotiation

HTML routes vary on `Accept` (`text/markdown` vs `text/html`) and `HX-Request`.
Standard Cloudflare caching ignores those `Vary` headers, so do **not** apply a
blanket Cache Everything rule to HTML routes—either bypass cache on HTML/MCP or
scope any HTML Cache Rule to exclude `HX-Request: true` and
`Accept: text/markdown`. Keep versioned `/static/` assets cached and purge
discovery/robots caches after deployment.

Cloudflare's **Markdown for Agents** (Pro/Business) is optional because the
origin handles `Accept: text/markdown` directly.

## Verify after deploying

```sh
curl -i https://tokencatalog.dev/md3/ -H 'Accept: text/markdown'
curl -i https://tokencatalog.dev/md3/ -H 'Accept: text/html'
curl -i https://tokencatalog.dev/robots.txt
curl -i https://tokencatalog.dev/md3/.well-known/api-catalog
curl -i https://tokencatalog.dev/md3/.well-known/ai-catalog.json
curl -i https://tokencatalog.dev/md3/.well-known/agent-skills/index.json
curl -i https://tokencatalog.dev/md3/mcp/server-card
curl -i https://tokencatalog.dev/md3/auth.md
curl -i https://tokencatalog.dev/md3/mcp \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -H 'MCP-Protocol-Version: 2025-11-25' \
  --data '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","clientInfo":{"name":"check","version":"1"},"capabilities":{}}}'
curl -i https://tokencatalog.dev/md3/mcp \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -H 'MCP-Protocol-Version: 2025-11-25' \
  --data '{"jsonrpc":"2.0","id":2,"method":"tools/list"}'
dig @1.1.1.1 +dnssec _index._agents.tokencatalog.dev SVCB
dig @1.1.1.1 +dnssec _mcp._agents.tokencatalog.dev SVCB
dig @1.1.1.1 _catalog._agents.tokencatalog.dev TXT
```

Expected results:
- `Accept: text/markdown` returns `Content-Type: text/markdown` with no `<script>` tags.
- Discovery endpoints return `200`; `api-catalog` returns `application/linkset+json`.
- `dig +dnssec` responses include the `ad` flag.

See [URL layout](../reference/agents.md#url-layout) for collection paths and
root discovery aliases.
