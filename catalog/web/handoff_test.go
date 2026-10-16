package web

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAgentHandoffUsesConfiguredURLsAndNegotiatesMarkdown(t *testing.T) {
	server := New(seoTestState(), "")
	if err := server.SetPublicOrigin("https://catalog.test/md3"); err != nil {
		t.Fatal(err)
	}
	for _, accept := range []string{"text/html", "text/markdown"} {
		request := httptest.NewRequest("GET", "/md3/for-agents", nil)
		request.Host = "untrusted.test"
		request.Header.Set("Accept", accept)
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, request)
		body := response.Body.String()
		for _, required := range []string{"Give this to your agent", "https://catalog.test/md3/skills/material-tokens/SKILL.md", "https://catalog.test/md3/mcp", "npx skills add sojupie/tokens --skill token-catalog"} {
			if !strings.Contains(body, required) {
				t.Fatalf("%s missing %s", accept, required)
			}
		}
		if response.Code != 200 || !strings.Contains(response.Header().Get("Content-Type"), accept) || strings.Contains(body, "untrusted.test") {
			t.Fatalf("wrong response: %d %s", response.Code, body)
		}
		if accept == "text/html" {
			if !strings.Contains(body, `href="/md3/for-agents"`) || !strings.Contains(body, "agent-handoff.js?v=") {
				t.Fatal("setup page is missing navigation or copy controls")
			}
		}
	}
	response := httptest.NewRecorder()
	New(seoTestState(), "").Handler().ServeHTTP(response, httptest.NewRequest("GET", "/for-agents", nil))
	if response.Code != 404 {
		t.Fatal("private server advertised a public handoff")
	}
}
