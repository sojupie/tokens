package web

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const tokenSkillPath = "/skills/material-tokens/SKILL.md"

//go:embed agent_openapi.json
var agentOpenAPI []byte

func agentDiscoveryPath(path string) bool {
	switch path {
	case "/agents.txt", "/agents.json", "/llms.txt", "/llms-full.txt", "/openapi.json", tokenSkillPath:
		return true
	}
	return false
}

func (server *Server) handleAgentDiscovery(response http.ResponseWriter, request *http.Request) {
	if server.publicOrigin == "" {
		http.Error(response, "Configure serve --public-origin to enable agent discovery.", http.StatusServiceUnavailable)
		return
	}
	response.Header().Set("Cache-Control", "public, max-age=3600")
	origin := server.publicOrigin
	if request.URL.Path == "/openapi.json" {
		var document map[string]any
		if err := json.Unmarshal(agentOpenAPI, &document); err != nil {
			http.Error(response, "Could not build API description.", http.StatusInternalServerError)
			return
		}
		document["servers"] = []map[string]string{{"url": origin}}
		writeJSON(response, http.StatusOK, document)
		return
	}
	if request.URL.Path == "/agents.json" {
		writeJSON(response, http.StatusOK, map[string]any{
			"$schema": "https://agents-txt.com/schema/agents-json/v1.0.json",
			"version": "1.0", "standard": "https://agents-txt.com",
			"site":   map[string]string{"name": "Tokens", "url": origin, "description": "Material Design token references, values, supported contexts and observed changes. An independent site, not a Google service."},
			"skills": []map[string]string{{"url": origin + tokenSkillPath, "description": "Find and read Material Design tokens and their observed change history through public JSON endpoints."}},
		})
		return
	}
	response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if request.URL.Path == "/agents.txt" {
		_, _ = fmt.Fprintf(response, "# agents.txt\n# Standard: https://agents-txt.com\n# JSON: %s/agents.json\n\nSkills: %s%s\n", origin, origin, tokenSkillPath)
		return
	}
	if request.URL.Path == tokenSkillPath {
		_, _ = fmt.Fprint(response, "---\nname: material-tokens\ndescription: Find Material Design token references, values, supported contexts and observed changes using the Tokens public read-only API.\n---\n\n", server.agentGuide())
		return
	}
	var body strings.Builder
	body.WriteString("# Tokens\n\n> An independent reference for Material Design tokens and observed changes. Not affiliated with Google.\n\n")
	fmt.Fprintf(&body, "Current snapshot: %s. Captured at: %s.\n\n", server.state.CurrentSnapshot.ID, server.state.CurrentSnapshot.CreatedAt)
	body.WriteString("## Read the data\n\n")
	for _, link := range []struct{ label, path, description string }{
		{"Usage guide", tokenSkillPath, "API endpoints, download scopes and interpretation."},
		{"API description", "/openapi.json", "OpenAPI paths, parameters and response schemas."},
		{"Subject index", "/api/v1/subjects", "Components and styles with exact page and JSON URLs."},
		{"Snapshots", "/api/v1/releases", "Capture dates and snapshot identifiers."},
		{"Change history", "/api/v1/changes", "All retained observed changes as JSON."},
		{"All tokens", "/downloads/tokens.json?scope=all", "Bulk JSON with every subject and supported context."},
		{"Page index", "/sitemap.xml", "Canonical HTML pages."},
	} {
		fmt.Fprintf(&body, "- [%s](%s%s): %s\n", link.label, origin, link.path, link.description)
	}
	if request.URL.Path == "/llms-full.txt" {
		body.WriteString("\n" + server.agentGuide() + "\n## Subject inventory\n\n")
		for _, subject := range sortedSubjects(server.state) {
			fmt.Fprintf(&body, "- [%s](%s%s) (%s): [JSON](%s%s?scope=all), %d tokens.\n", subject.Name, origin, subjectURL(subject), subject.Key.Kind, origin, subjectDownloadURL(subject), len(subject.Tokens))
		}
	}
	_, _ = fmt.Fprint(response, body.String())
}

func (server *Server) agentGuide() string {
	origin := server.publicOrigin
	return `# Read Material Design tokens

All endpoints below use GET. No account, API key, cookies or JavaScript are required.
The service is read-only. It does not provide MCP, payment or authentication protocols.
Read ` + origin + `/openapi.json for request parameters and response schemas.

## Find a component or style

Fetch ` + origin + `/api/v1/subjects. The subjects array contains kind, slug, name,
tokenCount, sourcePageUrls, url and jsonUrl. Follow jsonUrl rather than guessing a path;
style slugs can contain slashes. Follow url for the human-readable token table.
The older /api/v1/components endpoint returns the same index, including styles.

## Choose a download

- ` + origin + `/downloads/tokens.json?scope=components includes components only.
- ` + origin + `/downloads/tokens.json?scope=styles includes all non-component subjects.
- ` + origin + `/downloads/tokens.json?scope=all includes both.

Collection downloads contain capturedAt and subjects. A subject's jsonUrl returns
capturedAt, kind, slug, name and tokens. Both include all supported contexts.
For a smaller export, use the page's Download JSON form: view applies its table,
search and context filters; set includes the selected table in all contexts;
all includes every token in that subject. Copy the form's exact set and context
parameters. Do not invent contexts that the subject does not support.

## Interpret token values

Each token has name, optional displayName and type, and a values array.
Each value contains structured value data, optional context labels, default and
references. Preserve references and context labels. Do not replace references with
theme-specific hex colors or assume that every token resolves to a scalar.
An absent value is not zero. An absent context is not proof that a context exists.
Use sourcePageUrls to cite Material's source pages and capturedAt for freshness.
This site captures source data; it is not an official Google service.

## Read history

Fetch ` + origin + `/api/v1/releases for current and retained snapshots.
Fetch ` + origin + `/api/v1/changes for publicationId, snapshot and changes.
Filter changes locally by subject and snapshotId. The response is not paginated.
Each record includes detectedAt, type, name, kind, previousSnapshotId and snapshotId,
with before and after where available. Before and after are strings, which can
contain JSON. Dates show when the crawler observed a change, not its release date.
One token can change in multiple contexts; a record count is not a unique-token count.

## Read HTML instead

Use ` + origin + `/sitemap.xml for canonical pages. The site returns server-rendered
HTML. Leave HX-Request unset to receive a complete page, not an HTMX fragment.
Token rows remain in the HTML when their groups are collapsed.
Prefer one subject's JSON over repeated page scraping or a full-site download.
Reuse the subject index and discovery files rather than fetching them for every token.
`
}
