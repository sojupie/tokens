package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/sojupie/tokens/internal/publish"
	"github.com/sojupie/tokens/internal/usage"
)

const mcpVersion = "1.0.0"
const mcpProtocol = "2025-11-25"

var errMCPArguments = errors.New("Invalid tool parameters")

func (server *Server) mcpInfo() map[string]string {
	return map[string]string{"name": "io.github.sojupie/material-tokens", "title": "Token Catalog", "version": mcpVersion}
}

func (server *Server) mcpCard() map[string]any {
	return map[string]any{
		"$schema": "https://static.modelcontextprotocol.io/schemas/v1/server-card.schema.json",
		"name":    "io.github.sojupie/material-tokens", "title": "Token Catalog", "version": mcpVersion,
		"description": "Read Material Design 3 tokens, context-specific values, references and observed changes.",
		"websiteUrl":  server.publicOrigin,
		"remotes":     []any{map[string]any{"type": "streamable-http", "url": server.publicOrigin + "/mcp", "supportedProtocolVersions": []string{mcpProtocol, "2025-06-18", "2025-03-26"}}},
	}
}

func mcpTools() []any {
	tool := func(name, description string, properties map[string]any, required ...string) any {
		schema := map[string]any{"type": "object", "properties": properties, "additionalProperties": false}
		if len(required) > 0 {
			schema["required"] = required
		}
		return map[string]any{"name": name, "description": description, "inputSchema": schema,
			"annotations": map[string]bool{"readOnlyHint": true, "destructiveHint": false, "idempotentHint": true, "openWorldHint": false}}
	}
	return []any{
		tool("list_subjects", "Find Material Design 3 components and styles with exact token download URLs.", map[string]any{"query": map[string]string{"type": "string", "description": "Optional case-insensitive name or slug filter."}}),
		tool("get_tokens", "Read every table, token and supported context for one subject from list_subjects. Ignores browser search and filters. For filtered table downloads, use the HTTP API documented in OpenAPI.", map[string]any{
			"kind": map[string]any{"type": "string", "enum": []string{"component", "style", "foundation", "reference-token-set", "system-token-set"}},
			"slug": map[string]string{"type": "string", "description": "Exact subject slug returned by list_subjects."},
		}, "kind", "slug"),
		tool("list_changes", "Read retained observed changes; dates are detection dates, not upstream releases.", map[string]any{"subject": map[string]string{"type": "string", "description": "Optional exact subject slug."}}),
	}
}

func writeDiscoveryJSON(response http.ResponseWriter, value any) {
	response.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(response).Encode(value)
}

// Streamable HTTP in stateless JSON-response mode: no session or SSE stream.
func (server *Server) handleMCP(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Cache-Control", "no-store")
	if origin := request.Header.Get("Origin"); origin != "" && origin != server.browserOrigin {
		http.Error(response, "Origin is not allowed.", http.StatusForbidden)
		return
	}
	response.Header().Set("Access-Control-Allow-Origin", server.browserOrigin)
	response.Header().Set("Vary", "Origin")
	response.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	response.Header().Set("Access-Control-Allow-Headers", "Content-Type, MCP-Protocol-Version")
	if request.Method == http.MethodOptions {
		response.WriteHeader(http.StatusNoContent)
		return
	}
	if request.Method != http.MethodPost {
		response.Header().Set("Allow", "POST, OPTIONS")
		http.Error(response, "Only POST is supported; there is no event stream or session.", http.StatusMethodNotAllowed)
		return
	}
	if server.publicOrigin == "" {
		http.Error(response, "Configure serve --public-origin to enable MCP.", http.StatusServiceUnavailable)
		return
	}
	mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		http.Error(response, "Expected application/json.", http.StatusUnsupportedMediaType)
		return
	}
	if !strings.Contains(request.Header.Get("Accept"), "application/json") || !strings.Contains(request.Header.Get("Accept"), "text/event-stream") {
		http.Error(response, "Accept must include application/json and text/event-stream.", http.StatusNotAcceptable)
		return
	}
	version := request.Header.Get("MCP-Protocol-Version")
	if version != "" && version != mcpProtocol && version != "2025-06-18" && version != "2025-03-26" {
		http.Error(response, "Unsupported MCP protocol version.", http.StatusBadRequest)
		return
	}
	var message struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params"`
	}
	observed := usage.Event{Time: time.Now().UTC(), Event: "mcp_call", Route: "/mcp", Kind: "mcp", Method: "unknown", Status: 200}
	if server.usage != nil {
		defer func() {
			observed.DurationMS = float64(time.Since(observed.Time)) / float64(time.Millisecond)
			server.usage.Record(observed)
		}()
	}
	decoder := json.NewDecoder(http.MaxBytesReader(response, request.Body, 64<<10))
	rpcError := func(id json.RawMessage, code int, message string) {
		observed.Success = false
		if len(id) == 0 {
			id = json.RawMessage("null")
		}
		writeJSON(response, http.StatusOK, map[string]any{"jsonrpc": "2.0", "id": id, "error": map[string]any{"code": code, "message": message}})
	}
	if err := decoder.Decode(&message); err != nil {
		rpcError(nil, -32700, "Parse error")
		return
	}
	if decoder.Decode(new(any)) != io.EOF {
		rpcError(nil, -32700, "Expected one JSON-RPC message")
		return
	}
	for _, method := range []string{"initialize", "ping", "tools/list", "tools/call", "notifications/initialized"} {
		if message.Method == method {
			observed.Method = method
		}
	}
	validID := len(message.ID) == 0 || string(message.ID) == "null" || message.ID[0] == '"' || message.ID[0] == '-' || message.ID[0] >= '0' && message.ID[0] <= '9'
	if message.JSONRPC != "2.0" || message.Method == "" || !validID {
		rpcError(nil, -32600, "Invalid request")
		return
	}
	if len(message.ID) == 0 {
		observed.Success = true
		observed.Status = http.StatusAccepted
		response.WriteHeader(http.StatusAccepted)
		return
	}
	observed.Success = true
	var result any
	switch message.Method {
	case "initialize":
		var params struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		if json.Unmarshal(message.Params, &params) != nil || params.ProtocolVersion == "" {
			rpcError(message.ID, -32602, "protocolVersion is required")
			return
		}
		negotiated := mcpProtocol
		if params.ProtocolVersion == "2025-06-18" || params.ProtocolVersion == "2025-03-26" {
			negotiated = params.ProtocolVersion
		}
		result = map[string]any{"protocolVersion": negotiated, "serverInfo": server.mcpInfo(), "capabilities": map[string]any{"tools": map[string]bool{"listChanged": false}}, "instructions": "Public read-only Material Design 3 token data. No registration required. Start with list_subjects; use exact kind and slug for get_tokens. get_tokens returns the entire subject in all contexts. For filtered tables, read " + server.publicOrigin + "/openapi.json and " + server.publicOrigin + "/llms-full.txt."}
	case "ping":
		result = map[string]any{}
	case "tools/list":
		result = map[string]any{"tools": mcpTools()}
	case "tools/call":
		var params struct {
			Name      string                     `json:"name"`
			Arguments map[string]json.RawMessage `json:"arguments"`
		}
		if json.Unmarshal(message.Params, &params) != nil || params.Name == "" {
			rpcError(message.ID, -32602, "Tool name is required")
			return
		}
		observed.Tool = "unknown"
		for _, name := range []string{"list_subjects", "get_tokens", "list_changes"} {
			if params.Name == name {
				observed.Tool = name
			}
		}
		value, err := server.callMCPTool(params.Name, params.Arguments)
		if errors.Is(err, errMCPArguments) {
			rpcError(message.ID, -32602, err.Error())
			return
		}
		if err != nil {
			observed.Success = false
			result = map[string]any{"isError": true, "content": []any{map[string]string{"type": "text", "text": err.Error()}}}
		} else {
			encoded, _ := json.Marshal(value)
			result = map[string]any{"content": []any{map[string]string{"type": "text", "text": string(encoded)}}, "structuredContent": value, "isError": false}
		}
	default:
		rpcError(message.ID, -32601, "Method not found")
		return
	}
	writeJSON(response, http.StatusOK, map[string]any{"jsonrpc": "2.0", "id": message.ID, "result": result})
}

func (server *Server) callMCPTool(name string, arguments map[string]json.RawMessage) (any, error) {
	allowed := map[string][]string{"list_subjects": {"query"}, "get_tokens": {"kind", "slug"}, "list_changes": {"subject"}}
	keys, exists := allowed[name]
	if !exists {
		return nil, fmt.Errorf("%w: unknown tool %s", errMCPArguments, name)
	}
	values := map[string]string{}
	for key, raw := range arguments {
		found := false
		for _, candidate := range keys {
			if key == candidate {
				found = true
			}
		}
		if !found || json.Unmarshal(raw, new(string)) != nil || string(raw) == "null" {
			return nil, fmt.Errorf("%w: %s must be an allowed string argument", errMCPArguments, key)
		}
		var value string
		_ = json.Unmarshal(raw, &value)
		values[key] = value
	}
	switch name {
	case "list_subjects":
		subjects := make([]subjectSummary, 0)
		for _, subject := range sortedSubjects(server.state) {
			query := strings.ToLower(values["query"])
			if !strings.Contains(strings.ToLower(subject.Name+" "+subject.Key.Slug), query) {
				continue
			}
			subjects = append(subjects, subjectSummary{Kind: subject.Key.Kind, Slug: subject.Key.Slug, Name: subject.Name, TokenCount: len(subject.Tokens), LifecycleNoticeCount: len(subject.LifecycleNotices), SourcePageURLs: subject.SourcePageURLs, URL: server.publicOrigin + subjectURL(subject), JSONURL: server.publicOrigin + subjectDownloadURL(subject) + "?scope=all"})
		}
		return componentListResponse{PublicationID: server.state.PublicationID, Snapshot: server.state.CurrentSnapshot, Subjects: subjects}, nil
	case "get_tokens":
		if values["kind"] == "" || values["slug"] == "" {
			return nil, fmt.Errorf("%w: kind and slug are required", errMCPArguments)
		}
		subject, ok := server.state.Subjects[publish.SubjectKey{Kind: publish.SubjectKind(values["kind"]), Slug: values["slug"]}]
		if !ok {
			return nil, fmt.Errorf("Subject not found; use kind and slug from list_subjects")
		}
		return subject.PublicDocument(server.state.CurrentSnapshot.CreatedAt), nil
	default:
		changes := make([]publish.ChangeView, 0)
		for _, change := range server.state.Changes {
			if values["subject"] == "" || change.Subject == values["subject"] {
				changes = append(changes, change)
			}
		}
		return changesResponse{PublicationID: server.state.PublicationID, Snapshot: server.state.CurrentSnapshot, Subject: values["subject"], Changes: changes}, nil
	}
}
