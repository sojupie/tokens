# Token Catalog

Token Catalog collects design tokens and makes them easier to browse, compare
and download. [Material Design 3](https://tokencatalog.dev/md3/) is the first
supported collection.

The goal is fast, direct access to design data for developers and [agents](#use-with-your-agent). Tokens,
references and settings are available in readable pages and structured responses
that can be inspected, crawled and fetched with ordinary HTTP clients. Source
links connect the data to the original design guidance.

![Token Catalog showing the Material Design 3 component catalog](docs/images/catalog.png)

## Browse and download

Find tokens by component or style, inspect their references, and filter values
by platform, theme and other settings. Downloads can include the filtered table,
all values in a table, a whole component, or the collection.

For example, select a Buttons table, choose a platform and theme, then download
“This table, with your filters” to export that selection. Broader download
options include values outside the selected table and filters.

The change log compares retained captures. It shows added, removed and changed
values, including changes to individual settings. Dates record when a change
was observed. Incomplete captures remain marked as partial.

The same data is available through a public JSON API, Markdown responses and
read-only MCP tools. See [usage](docs/usage.md) and
[API and agent access](docs/reference/agents.md).

## Component playground

Component pages include interactive examples and comparisons of supported
variants, sizes and states. The examples use CSS bindings generated from the
captured tokens, with authored HTML, layout and browser behavior where tokens
do not describe implementation. Theme and contrast controls apply to both the
site and its examples.

The gallery and playground use the same Forge assets. Specimen markup,
component styling and browser behavior have separate source files; see the
[component authoring map](docs/architecture.md#component-authoring) to follow a
component from its example to its token bindings. These components are currently
part of the site; a standalone library is planned separately.

## Use with your agent

[Give this to your agent](https://tokencatalog.dev/md3/for-agents) has a copyable
prompt, skill install command, and MCP server URL.

Give your agent this instruction:

> Use Token Catalog at https://tokencatalog.dev/md3 to find the Material Design 3
> tokens for this task. Read its agent guide, fetch the relevant component or
> style, and preserve the token references and selected context.

For agents that support installable skills:

```sh
npx skills add sojupie/tokens --skill token-catalog
```

The installer lets you choose your agent. The [skill](skills/token-catalog/SKILL.md)
lives in this repo and reads the current API guide from the site. You do not need
to clone or run the catalog to use its data.

If your client supports MCP, add `https://tokencatalog.dev/md3/mcp` as a
Streamable HTTP server. Start with `list_subjects`, then call `get_tokens` with
the returned kind and slug. No account or API key is required.

Example requests:

- “Find the filled button tokens for Android in the dark theme.”
- “Compare the corner tokens for cards and dialogs.”
- “Check which motion token values changed between retained captures.”

See [API and agent access](docs/reference/agents.md) for setup and data formats.

## Usage

![Token Catalog usage over the last seven days: server requests, downloads, Markdown requests, successful MCP tool calls, and popular subjects](docs/images/usage.svg)

The graphic updates weekly from retained server request totals, with prefetches
counted separately.

## Performance and access

Pages render on the server with the token data already present in the HTML.
Reading pages and fetching JSON do not require JavaScript or an account.
Markdown responses, OpenAPI, discovery files and MCP tools make the same data
available to different clients. Small assets and deferred scripts keep the
browser's initial work limited.

The October 11, 2026 [mobile PageSpeed Insights report](https://pagespeed.web.dev/analysis/https-tokencatalog-dev/91lzhrkat6?hl=en&form_factor=mobile)
scores 100 for performance, accessibility, best practices and SEO, with 4/4
agentic browsing checks passed. The measured largest contentful paint is 1.2 s,
with 0 ms total blocking time and no layout shift.

The captured [Is It Agent Ready?](https://isitagentready.com/) check scores 80/100,
including 100 for discoverability, content and bot access control.
The remaining three checks concern OAuth/OIDC discovery, protected-resource
metadata and agent registration. Token Catalog provides public, read-only access
without accounts or protected APIs, so those authentication features are
deliberately omitted. The lower score reflects those checks being counted as
missing even though the discovery and access checks for the public service pass.

<a href="https://pagespeed.web.dev/analysis/https-tokencatalog-dev/91lzhrkat6?hl=en&amp;form_factor=mobile"><img src="docs/images/PageSpeedInsights.png" alt="October 11, 2026 mobile PageSpeed Insights report showing four scores of 100 and 4/4 agentic browsing checks" width="420" /></a>
<img src="docs/images/isitagentready.png" alt="Is It Agent Ready report for Token Catalog showing 80/100 overall and 100 for discoverability, content and bot access control" width="420" />

## Repository components

The project has three parts in one repository. They share the existing CLI,
Go module, and deployment.

| Part | What it does | Source |
| --- | --- | --- |
| Source | Captures upstream data, verifies snapshots, compares changes, builds publications, and exports DTCG | [`source/`](source/) |
| Catalog | Serves token tables, filters, downloads, history, API and agent interfaces, and usage reports | [`catalog/`](catalog/) |
| Forge | Generates CSS and component state bindings, implements browser behavior, and checks the rendered result | [`forge/`](forge/) |

[`skills/`](skills/) contains the installable agent skill and the reference used
by the hosted guide. [`cmd/tokens/`](cmd/tokens/) provides the shared CLI.
Deployment configuration remains in [`deploy/`](deploy/).

Forge generates the catalog theme and shared components, including buttons,
cards, chips, fields, menus, and dialogs. See [generated components](docs/reference/components.md)
and [buttons](docs/reference/buttons.md) for supported variants, states, and checks.

## Material Design 3 data pipeline

The capture command opens the reviewed source pages with Chromium and saves
their token payloads. Snapshot creation bundles those payloads with extracted
tokens, source URLs, coverage results and SHA-256 checksums. Verification checks
that the saved files match their checksums and schemas.

Publication compares verified snapshots and builds a SQLite read index, change
history and JSON downloads. The web server loads a publication at startup and
uses it for HTML pages, API responses and MCP tools. Its own Material component
styles also come from the published tokens.

Scheduled updates carry out the same steps, retain a report for each run and
check the candidate publication before activation. See
[architecture](docs/architecture.md) for the data flow and
[scheduled updates](docs/operations/updates.md) for review and recovery.

## Run the catalog server locally

Build the Go application with Go 1.26 or newer:

```sh
go build -o ./bin/tokens ./cmd/tokens
```

With an existing publication in `published/`, start the catalog server:

```sh
./bin/tokens serve --published published --listen 127.0.0.1:8081
```

Open [localhost:8081](http://127.0.0.1:8081) to browse the saved tokens. This starts
the website and its data endpoints. The [agent access reference](docs/reference/agents.md)
explains public-origin configuration for discovery metadata and MCP.

### Create a Material Design 3 publication

For a fresh checkout, install Chrome or Chromium and run these commands before
starting the server:

```sh
./bin/tokens crawl
./bin/tokens snapshot create
./bin/tokens snapshot verify
./bin/tokens publish
```

The crawl fetches Google's source pages and payloads over the network.
Subsequent server starts use the saved publication. On Windows, build
`bin/tokens.exe` and use that executable for the same commands.

See [development](docs/development.md) for tests and template changes, or
[deployment](docs/operations/deployment.md) for Docker, scheduled updates and
usage reports. To use the hosted data from an agent, follow
[Use with your agent](#use-with-your-agent).

## Supported design systems and source coverage

Material Design 3 is the current collection. Its capture and extraction
code reads Google's source payloads, and its website theme uses Material tokens.
Adding another design system requires a source catalog, capture and extraction
rules, publication mapping, a website theme and source attribution. The
[development guide](docs/development.md#other-design-systems) describes those
integration points.

The Material Design 3 page catalog defines capture coverage. Snapshots retain
page and payload results, and incomplete captures are marked as partial in the
change history. Coverage records accompany comparisons so source changes can
be assessed against the pages captured in each run.

## Documentation

- [Using the catalog](docs/usage.md)
- [Developing and contributing](CONTRIBUTING.md)
- [Architecture](docs/architecture.md)
- [CLI, snapshots and publication](docs/README.md#reference)
- [Deployment and recovery](docs/README.md#operations)
- [Project history](docs/history/README.md)

## License and attribution

Token Catalog engine code and original documentation are licensed under the
[Business Source License 1.1 (BSL 1.1)](LICENSE), which converts to the
[Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0) three years
after the publication date of each version or commit.

Under the Additional Use Grant, Token Catalog is free to use for all individual
developers, AI agents, non-commercial entities, and any deployment serving
publicly available, openly licensed design tokens. Production use and self-hosting
by any entity to serve, browse, or manage proprietary, internal, or non-open
corporate design systems is strictly prohibited without a separate commercial agreement.

Material Design 3 data comes from Google and retains its upstream terms.
Bundled fonts, icons and dependencies retain their own licenses.

Token Catalog is an independent project. See [third-party notices](THIRD_PARTY_NOTICES.md)
and [licensing and source attribution](docs/licensing.md) for the distinction
between the software engine and the token data it publishes.
