# Agent discovery and Cloudflare setup

Deploy the application changes using the normal [deployment workflow](../deploy/README.md).
For GitHub deployments, set `TOKENS_PUBLIC_ORIGIN` to `https://tokencatalog.dev/md3`
in the repository’s **Settings → Environments → production → Environment variables**.
The deploy action saves it in `/etc/tokens/server.env` for Docker and systemd.
For manual deployments, set it directly in that file. Discovery files and MCP
return 503 until this is configured. DNS and Cloudflare account settings are
separate from the application deployment and have not been changed by this patch.

## Implemented

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
The supplied `ai-input=no` example would conflict with this site's agent API use.
Change `contentSignal` in `internal/web/discovery.go` if a different policy is wanted.

## Cloudflare changes

### 1. Connect the new domain and publish discovery DNS

In the **tokencatalog.dev** zone, confirm the domain uses the Cloudflare
nameservers shown in Overview. In **Zero Trust → Networks → Connectors →
Cloudflare Tunnels**, open the existing Tokens tunnel and add a published
application route: hostname `tokencatalog.dev` (leave subdomain and path empty),
service `HTTP`, URL `tokens:8080`. Route the whole hostname because robots.txt
and discovery files also live at the root. Do not strip `/md3` or put it in the
service URL. The application handles that path itself.

Cloudflare should create a proxied CNAME for `@` pointing to the tunnel's
`<UUID>.cfargotunnel.com` hostname. If it does not, add it in **DNS → Records**.
Do not add an A record with the server's IP. Resolve any existing conflicting
apex web record before creating the Tunnel record; keep unrelated mail records.

Then add these DNS-only discovery records, TTL 3600:

| Type | Name | Priority | Target | Values / Content |
| --- | --- | --- | --- | --- |
| SVCB | `_index._agents` | 1 | `tokencatalog.dev` | `alpn="h2" port=443 mandatory=alpn,port key65400="https://tokencatalog.dev/.well-known/ai-catalog.json"` |
| SVCB | `_mcp._agents` | 1 | `tokencatalog.dev` | `alpn="mcp,h2" port=443 mandatory=alpn,port key65400="https://tokencatalog.dev/md3/mcp/server-card" key65402="mcp"` |
| TXT | `_catalog._agents` | — | — | `url=https://tokencatalog.dev/.well-known/ai-catalog.json` |

[Copyable zone records](../deploy/dns-aid.zone.example). DNS names contain only
the hostname; `/md3` belongs in URLs. The numbered SVCB keys and `mcp` ALPN
label are experimental. The server card advertises the actual MCP endpoint.
Enable HTTP/2 in the zone. No A2A records are needed.

Sources: [Tunnel routing](https://developers.cloudflare.com/tunnel/concepts/routing/),
[DNS-AID draft](https://datatracker.ietf.org/doc/draft-mozleywilliams-dnsop-dnsaid/),
[Cloudflare SVCB records](https://developers.cloudflare.com/dns/manage-dns-records/reference/dns-record-types/#svcb-and-https).

### 2. Enable DNSSEC and complete registrar setup

Open **DNS → Settings → DNSSEC** and enable it. Follow Cloudflare's displayed
DS-record instructions at your domain registrar. If Cloudflare Registrar manages
the domain, Cloudflare handles the DS record. Confirm the DNSSEC status is active
and a validating resolver returns the AD flag for the discovery records.
Enabling zone signing without the parent DS chain does not authenticate discovery.
See [Cloudflare DNSSEC setup](https://developers.cloudflare.com/dns/dnssec/).

### 3. Allow public agent traffic

Review Cloudflare Access, WAF custom rules, Bot Fight Mode / Super Bot Fight Mode,
AI Crawl Control and rate-limit rules for this hostname. Ensure legitimate
reads of the HTML pages, `/robots.txt`, `/agents.txt`, `/agents.json`, `/llms.txt`,
`/llms-full.txt`, `/openapi.json`, `/auth.md`, `/skills/`, `/.well-known/`, `/api/`
and `/downloads/` do not require a login or JavaScript challenge. Allow POST and
OPTIONS at `/mcp`; GET at that endpoint returns 405 by design. Keep ordinary
rate limiting and DDoS protection. Some bot settings cannot be skipped through
custom WAF rules; adjust the relevant product setting if it still blocks requests.

Browser Integrity Check can block ordinary API clients, including Python's
default user agent, with HTTP 403 and `error code: 1010`. If that happens, create
a Configuration Rule matching `http.host eq "tokencatalog.dev"` and set
Browser Integrity Check to Off for this hostname. This allows clients to use
their normal user agent. See [Browser Integrity Check](https://developers.cloudflare.com/waf/tools/browser-integrity-check/).

### 4. Keep content signals consistent

If Cloudflare Managed robots.txt or AI Crawl Control inserts content signals,
use the same `ai-train=no, search=yes, ai-input=yes` policy or disable the managed
robots.txt injection for this site. Confirm the live robots response contains
one consistent policy. The origin sends its own `Content-Signal` header, which
Cloudflare Markdown for Agents preserves.

### 5. Review caching and optional Markdown conversion

Apply the following checks to `tokencatalog.dev`. Paths listed here are relative
to `/md3` unless described as root discovery files.

Remove any Cache Everything / custom cache rule that caches HTML without
separating `Accept` and `HX-Request` variants. Bypass cache for `/`, `/components`,
`/components/*`, `/styles`, `/styles/*`, `/changes` and `/mcp` if such a rule exists.
Leave versioned `/static/` assets cached. Purge any older cached robots or
discovery responses after deployment.

**Markdown for Agents is optional** because the origin now handles negotiation.
If you want Cloudflare's converter (including its token-count header when
available), select the zone → **AI Crawl Control → Markdown for Agents**, or
create a Configuration Rule for this hostname. Cloudflare currently documents
Pro/Business availability. Test both representations after enabling it.
See [Cloudflare Markdown for Agents](https://developers.cloudflare.com/fundamentals/reference/markdown-for-agents/).

## Verify after deploying

Replace the example origin in these commands:

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

The Markdown response should contain the page heading and data, with no HTML
scripts, and `Content-Type: text/markdown`. Discovery should return 200; the
API catalog must use `application/linkset+json`. DNSSEC responses should include
`ad` in the flags. Test WebMCP in a supporting browser; the automated JS tests
exercise registration, API reads and cleanup, but cannot provide browser support.
Rerun the readiness scan after DNS propagation and deployment; authentication
and registration checks can remain absent for this public service.

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

For migration, in the old domain's **Rules → Redirect Rules**, match
`http.host eq "tokens.marmenlind.com"`. Use a dynamic destination
`concat("https://tokencatalog.dev/md3", http.request.uri.path)`, preserve the
query string, and use status **308** to preserve POST methods. Keep the old
proxied DNS record while this redirect is needed. Update any clients that use
MCP directly to the new endpoint; some clients do not follow redirects.

In the new zone's **Rules → Configuration Rules**, disable Browser Integrity
Check for `http.host eq "tokencatalog.dev"` if ordinary API clients are blocked.
In **AI Crawl Control**, leave crawler blocking disabled, and turn off managed
robots.txt if it overwrites the site's policy. Markdown conversion is optional;
the application already supports it. If custom caching rules apply, bypass
`/md3/` except versioned `/md3/static/` files, and bypass root discovery files.
Purge old cached discovery responses after deploying.

Set the production GitHub Actions variable (not a secret) to
`TOKENS_PUBLIC_ORIGIN=https://tokencatalog.dev/md3`, then run Deploy. There is
no separate base-path variable. The tunnel token can remain the same when you
reuse the existing tunnel.
