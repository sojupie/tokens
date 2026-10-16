package web

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/sojupie/tokens/skills"
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
			"site":   map[string]string{"name": "Token Catalog", "url": origin, "description": "Material Design 3 token references, values, supported contexts and observed changes. An independent site, not a Google service."},
			"skills": []map[string]string{{"url": origin + tokenSkillPath, "description": "Find and read Material Design 3 tokens and their observed change history through public JSON endpoints."}},
		})
		return
	}
	response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if request.URL.Path == "/agents.txt" {
		_, _ = fmt.Fprintf(response, "# agents.txt\n# Standard: https://agents-txt.com\n# JSON: %s/agents.json\n\nSkills: %s%s\n", origin, origin, tokenSkillPath)
		return
	}
	if request.URL.Path == tokenSkillPath {
		_, _ = fmt.Fprint(response, server.tokenSkill())
		return
	}
	var body strings.Builder
	body.WriteString("# Token Catalog\n\n> An independent reference for Material Design 3 tokens and observed changes. Not affiliated with Google.\n\n")
	fmt.Fprintf(&body, "Current snapshot: %s. Captured at: %s.\n\n", server.state.CurrentSnapshot.ID, server.state.CurrentSnapshot.CreatedAt)
	body.WriteString("## Read the data\n\n")
	for _, link := range []struct{ label, path, description string }{
		{"Usage guide", tokenSkillPath, "API endpoints, download scopes and interpretation."},
		{"API description", "/openapi.json", "OpenAPI paths, parameters and response schemas."},
		{"API catalog", apiCatalogPath, "API description, usage documentation and health endpoint."},
		{"AI catalog", aiCatalogPath, "API, agent skill and MCP discovery."},
		{"Skills index", skillIndexPath, "Available agent skills and their content digests."},
		{"MCP server card", "/mcp/server-card", "Public MCP transport and supported protocol versions."},
		{"Authentication guide", "/auth.md", "Public access; no credentials or registration needed."},
		{"Subject index", "/api/v1/subjects", "Components and styles with exact page and JSON URLs."},
		{"Snapshots", "/api/v1/releases", "Capture dates and snapshot identifiers."},
		{"Change history", "/api/v1/changes", "All retained observed changes as JSON."},
		{"All tokens", "/downloads/tokens.json?scope=all", "Bulk JSON with every subject and supported context."},
		{"Page index", "/sitemap.xml", "Canonical HTML pages."},
	} {
		fmt.Fprintf(&body, "- [%s](%s%s): %s\n", link.label, origin, link.path, link.description)
	}
	if request.URL.Path == "/llms.txt" {
		fmt.Fprintf(&body, "- [Full API and agent guide](%s/llms-full.txt): Usage instructions, MCP, Markdown and the subject inventory.\n", origin)
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
	dtcgGuide := ""
	if server.tokenAssetsDir != "" {
		dtcgGuide = `## DTCG and CSS downloads

Fetch ` + origin + `/downloads/dtcg/light-default/tokens.tokens.json for DTCG 2025.10
values and aliases in the Light, Default contrast, Web, 3P, Static font context.
Fetch tokens.css in the same directory for generated CSS variables and report.json
for context details, omitted values and shapes requiring source extensions.
Other appearance contexts are light-medium, light-high, dark-default, dark-medium
and dark-high. Each file contains one selected context and its dependencies.

`
	}

	return strings.NewReplacer(
		"{{PUBLIC_ORIGIN}}", origin,
		"{{DTCG_DOWNLOADS}}", dtcgGuide,
		"{{MCP_PROTOCOL}}", mcpProtocol,
	).Replace(skills.MaterialGuide)
}
