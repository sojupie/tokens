# Domain migration

October 2026: the hosted collection moved from `tokens.marmenlind.com` to
`tokencatalog.dev/md3`. These notes record the old-host redirect configuration.
Current setup is in [Cloudflare setup](../operations/cloudflare.md).

For migration, in the old domain's **Rules → Redirect Rules**, match
`http.host eq "tokens.marmenlind.com"`. Use a dynamic destination
`concat("https://tokencatalog.dev/md3", http.request.uri.path)`, preserve the
query string, and use status **308** to preserve POST methods. Keep the old
proxied DNS record while this redirect is needed. Update any clients that use
MCP directly to the new endpoint; some clients do not follow redirects.

