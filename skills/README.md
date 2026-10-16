# Agent skills

`token-catalog/` is the installable skill for agents reading the public catalog.
Install it with the existing command:

```sh
npx skills add sojupie/tokens --skill token-catalog
```

Its Material reference also supplies the hosted API guide. The server replaces
`{{PUBLIC_ORIGIN}}`, `{{MCP_PROTOCOL}}`, and `{{DTCG_DOWNLOADS}}` with deployment
values. Keep these placeholders intact when editing the reference.

A new design system should add a reference for its token semantics and supported
routes. Add a separate skill when its workflow needs different instructions.
Repository development guidance belongs in the root `AGENTS.md`.
