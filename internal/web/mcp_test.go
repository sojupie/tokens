package web

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMCPProtocolAndTools(t *testing.T) {
	server := discoveryTestServer(t)
	rpc := func(body string) map[string]json.RawMessage {
		t.Helper()
		response := httptest.NewRecorder()
		request := httptest.NewRequest("POST", "/mcp", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Accept", "application/json, text/event-stream")
		server.Handler().ServeHTTP(response, request)
		if response.Code != 200 {
			t.Fatalf("MCP returned %d: %s", response.Code, response.Body.String())
		}
		var value map[string]json.RawMessage
		if err := json.Unmarshal(response.Body.Bytes(), &value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	init := rpc(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","clientInfo":{"name":"test","version":"1"},"capabilities":{}}}`)
	if !strings.Contains(string(init["result"]), mcpProtocol) {
		t.Fatal("no negotiated protocol")
	}
	tools := rpc(`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
	if !strings.Contains(string(tools["result"]), "get_tokens") {
		t.Fatal("no tools")
	}
	for _, call := range []struct{ params, content string }{
		{`{"name":"list_subjects","arguments":{"query":"snack"}}`, "snackbar"},
		{`{"name":"get_tokens","arguments":{"kind":"component","slug":"snackbar"}}`, "capturedAt"},
		{`{"name":"list_changes","arguments":{}}`, "changes"},
	} {
		value := rpc(`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":` + call.params + `}`)
		result := string(value["result"])
		if !strings.Contains(result, call.content) || !strings.Contains(result, `"isError":false`) || !strings.Contains(result, "structuredContent") {
			t.Fatalf("tool failed: %s", result)
		}
	}
	for _, body := range []string{`{`, `[]`, `{"jsonrpc":"1.0","method":"ping","id":1}`, `{"jsonrpc":"2.0","id":{},"method":"ping"}`, `{"jsonrpc":"2.0","id":1,"method":"missing"}`} {
		if rpc(body)["error"] == nil {
			t.Fatal("invalid RPC accepted")
		}
	}
	bad := rpc(`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_tokens","arguments":{"kind":"component","slug":"missing"}}}`)
	if !strings.Contains(string(bad["result"]), `"isError":true`) {
		t.Fatal("missing subject did not return tool error")
	}
	for _, body := range []string{`{"query":null}`, `{"query":4}`, `{"url":"https://untrusted.example"}`} {
		var arguments map[string]json.RawMessage
		_ = json.Unmarshal([]byte(body), &arguments)
		if _, err := server.callMCPTool("list_subjects", arguments); err == nil {
			t.Fatal("invalid tool argument accepted")
		}
	}
}

func TestMCPTransportChecks(t *testing.T) {
	server := discoveryTestServer(t)
	for _, check := range []struct {
		method, origin, mediaType, accept, version, body string
		status                                           int
	}{
		{"GET", "", "", "", "", "", 405},
		{"POST", "https://evil.example", "application/json", "application/json, text/event-stream", "", `{}`, 403},
		{"POST", "", "text/plain", "application/json, text/event-stream", "", `{}`, 415},
		{"POST", "", "application/json", "application/json", "", `{}`, 406},
		{"POST", "", "application/json", "application/json, text/event-stream", "invalid", `{}`, 400},
		{"POST", "https://tokens.example", "application/json", "application/json, text/event-stream", mcpProtocol, `{"jsonrpc":"2.0","method":"notifications/initialized"}`, 202},
		{"OPTIONS", "https://tokens.example", "", "", "", "", 204},
	} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(check.method, "/mcp", strings.NewReader(check.body))
		request.Header.Set("Origin", check.origin)
		request.Header.Set("Content-Type", check.mediaType)
		request.Header.Set("Accept", check.accept)
		request.Header.Set("MCP-Protocol-Version", check.version)
		server.Handler().ServeHTTP(response, request)
		if response.Code != check.status {
			t.Fatalf("%s %s: got %d want %d", check.method, check.origin, response.Code, check.status)
		}
	}
}
