# Token Catalog

Token Catalog contains three parts in one repository:

- `source/` captures upstream data, verifies snapshots, builds publications, and exports DTCG.
- `catalog/` serves the website, API, MCP, discovery documents, and usage reports.
- `forge/` generates CSS and implements component behavior from selected tokens.

The Go packages share one module and the `cmd/tokens` executable. Run commands
from the repository root. Node dependencies and scripts use the root package.json.
See `docs/development.md` for commands and `docs/architecture.md` for data flow.

Preserve source values, references, contexts, provenance, and capture completeness.
Material-specific mappings must remain explicit when adding another design system.
Use the current captured token data and source guidance for component states.
Check computed styles and behavior on the real pages as well as the component lab.

Keep token overrides in `forge/material/overrides.json`. Generate CSS usage reports
from the source; do not add approval workflows or manually maintained review records.
Use Go for CLI commands, validation and reports. Node handles Style Dictionary
and browser-side JavaScript.

`skills/token-catalog/` contains instructions for agents using the public service.
Its `references/material-tokens.md` is also embedded in the hosted API guide.
Keep public URLs and skill installation commands stable when moving code.

Keep copy plain and connected. Use established domain terms and omit filler,
contrast slogans, invented component names, and tool branding.

Keep generated captures and publications out of commits. Preserve unrelated
working changes. Leave visual changes available on a local server for review
before pushing.
