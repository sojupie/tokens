package web

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/url"
)

const (
	apiCatalogPath = "/.well-known/api-catalog"
	aiCatalogPath  = "/.well-known/ai-catalog.json"
	skillIndexPath = "/.well-known/agent-skills/index.json"
	mcpCardPath    = "/.well-known/mcp/server-card.json"
	// Allow agent use and search, while declining model training.
	contentSignal = "ai-train=no, search=yes, ai-input=yes"
)

func wellKnownPath(path string) bool {
	switch path {
	case apiCatalogPath, aiCatalogPath, skillIndexPath, mcpCardPath, "/mcp/server-card", "/auth.md":
		return true
	}
	return false
}

func (server *Server) tokenSkill() string {
	return "---\nname: material-tokens\ndescription: Find Material Design 3 token references, values, supported contexts and observed changes using the Token Catalog public read-only API.\n---\n\n" + server.agentGuide()
}

func (server *Server) handleWellKnown(response http.ResponseWriter, request *http.Request) {
	if server.publicOrigin == "" {
		http.Error(response, "Configure serve --public-origin to enable agent discovery.", http.StatusServiceUnavailable)
		return
	}
	response.Header().Set("Cache-Control", "public, max-age=3600")
	origin := server.publicOrigin
	switch request.URL.Path {
	case apiCatalogPath:
		response.Header().Set("Content-Type", "application/linkset+json")
		writeDiscoveryJSON(response, map[string]any{"linkset": []any{map[string]any{
			"anchor":       origin + "/api/v1/",
			"service-desc": []any{map[string]string{"href": origin + "/openapi.json", "type": "application/json"}},
			"service-doc":  []any{map[string]string{"href": origin + tokenSkillPath, "type": "text/plain"}},
			"status":       []any{map[string]string{"href": origin + "/api/v1/health", "type": "application/json"}},
		}}})
	case skillIndexPath:
		writeJSON(response, http.StatusOK, map[string]any{
			"$schema": "https://schemas.agentskills.io/discovery/0.2.0/schema.json",
			"skills": []any{map[string]string{
				"name": "material-tokens", "type": "skill-md",
				"description": "Find Material Design 3 tokens, references, context-specific values and observed changes.",
				"url":         origin + tokenSkillPath,
				"digest":      fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(server.tokenSkill()))),
			}},
		})
	case aiCatalogPath:
		parsed, _ := url.Parse(origin)
		host := parsed.Hostname()
		entry := func(namespace, name, title, mediaType, path string, queries ...string) any {
			return map[string]any{"identifier": "urn:air:" + host + ":" + namespace + ":" + name,
				"displayName": title, "type": mediaType, "url": origin + path, "representativeQueries": queries}
		}
		writeJSON(response, http.StatusOK, map[string]any{
			"specVersion": "1.0", "host": map[string]string{"displayName": "Token Catalog", "identifier": origin},
			"entries": []any{
				entry("api", "material-tokens", "Token Catalog API", "application/json", "/openapi.json", "Find Material Design 3 component tokens", "Download all style tokens", "Compare observed token changes"),
				entry("skill", "material-tokens", "Read Material Design 3 tokens", "text/plain", tokenSkillPath, "How do I read token references?", "How do I choose a token context?"),
				entry("server", "material-tokens", "Token Catalog MCP server", "application/mcp-server-card+json", "/mcp/server-card", "Find tokens for a Material Design 3 component", "Read the latest token changes"),
			},
		})
	case mcpCardPath, "/mcp/server-card":
		card := server.mcpCard()
		if request.URL.Path == mcpCardPath {
			// Compatibility with the older scanner shape; the canonical card uses SEP-2127.
			card["serverInfo"] = server.mcpInfo()
			card["transport"] = map[string]string{"type": "streamable-http", "endpoint": origin + "/mcp"}
			card["capabilities"] = map[string]any{"tools": map[string]bool{"listChanged": false}}
		}
		writeJSON(response, http.StatusOK, card)
	case "/auth.md":
		response.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		fmt.Fprintf(response, "# Token Catalog auth.md\n\n## Access for agents\n\nToken Catalog is a public, read-only Material Design 3 token reference. Agents need no registration, account, API key, access token, or cookies. No credentials are issued.\n\nUse GET requests to %s/api/v1/subjects and follow each subject's jsonUrl. Read %s/openapi.json for the full API. MCP clients can connect to %s/mcp using Streamable HTTP without authentication.\n\nThere is no OAuth authorization server, protected resource, or agent registration endpoint. OAuth discovery and agent_auth metadata are therefore not published.\n", origin, origin, origin)
	}
}
