---
name: token-catalog
description: Look up Material Design 3 design tokens, context-specific values, references, and observed changes from Token Catalog. Use when implementing or reviewing Material Design UI and exact token data is needed.
---

# Use Token Catalog

Token Catalog provides public, read-only token data at
https://tokencatalog.dev/md3. No account or API key is required.

1. Read https://tokencatalog.dev/md3/skills/material-tokens/SKILL.md for the
   current data format, download scopes, and interpretation rules. The
   [bundled reference](references/material-tokens.md) contains the same guide
   as a server template; use the hosted version for deployment URLs, protocol
   versions, and available generated downloads.
2. Fetch https://tokencatalog.dev/md3/api/v1/subjects and find the component or
   style needed for the task. Use its exact `kind` and `slug`. Follow `jsonUrl`
   to download its tokens; style slugs can contain slashes.
3. Inspect the supported contexts before selecting a value. Preserve token
   names, references, and context labels. Report the chosen context and cite
   the source pages from `sourcePageUrls` when explaining a value.

Fetch only the relevant subjects. A subject download includes every context.
Use the filtered export described in the live guide when a single table or
context is sufficient. Avoid downloading the full collection for one component.

For MCP-enabled clients, connect to https://tokencatalog.dev/md3/mcp using
Streamable HTTP. Use `list_subjects`, then `get_tokens` with the returned kind
and slug. `list_changes` reads observed changes. Plain HTTP works without MCP.
Read https://tokencatalog.dev/md3/openapi.json for API parameters and schemas.
HTML pages also support `Accept: text/markdown`.

Do not invent missing values or contexts. An absent value is not zero. Preserve
references rather than replacing them with one theme's resolved color. Captures
can be partial, and change dates are observation dates rather than upstream
release dates. Use `capturedAt` and the releases endpoint to check freshness.

Apply the data to the user's existing framework and project conventions.
Token Catalog is an independent reference, not an official Google service.
