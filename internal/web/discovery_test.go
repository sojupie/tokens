package web

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func discoveryTestServer(t *testing.T) *Server {
	t.Helper()
	server := New(seoTestState(), "")
	if err := server.SetPublicOrigin("https://tokens.example"); err != nil {
		t.Fatal(err)
	}
	return server
}

func TestWellKnownCatalogs(t *testing.T) {
	server := discoveryTestServer(t)
	for _, path := range []string{apiCatalogPath, aiCatalogPath, skillIndexPath, mcpCardPath, "/mcp/server-card", "/auth.md"} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest("GET", path, nil)
			request.Host = "untrusted.example"
			server.Handler().ServeHTTP(response, request)
			if response.Code != 200 || response.Header().Get("Access-Control-Allow-Origin") != "*" || strings.Contains(response.Body.String(), "untrusted.example") {
				t.Fatalf("bad discovery response: %v %s", response.Code, response.Body.String())
			}
			if path != "/auth.md" && !json.Valid(response.Body.Bytes()) {
				t.Fatal("invalid JSON")
			}
			preview := httptest.NewRecorder()
			New(seoTestState(), "").Handler().ServeHTTP(preview, request)
			if preview.Code != 503 {
				t.Fatal("preview must not invent discovery URLs")
			}
			preflight := httptest.NewRecorder()
			server.Handler().ServeHTTP(preflight, httptest.NewRequest("OPTIONS", path, nil))
			if preflight.Code != 204 {
				t.Fatal("missing CORS preflight")
			}
		})
	}
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", apiCatalogPath, nil))
	if response.Header().Get("Content-Type") != "application/linkset+json" {
		t.Fatal("wrong catalog media type")
	}
	var catalog struct{ Linkset []map[string]json.RawMessage }
	_ = json.Unmarshal(response.Body.Bytes(), &catalog)
	if len(catalog.Linkset) != 1 {
		t.Fatal("missing API")
	}
	for _, rel := range []string{"anchor", "service-desc", "service-doc", "status"} {
		if catalog.Linkset[0][rel] == nil {
			t.Fatal("missing relation " + rel)
		}
	}
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", skillIndexPath, nil))
	var index struct {
		Schema string `json:"$schema"`
		Skills []struct{ URL, Digest string }
	}
	_ = json.Unmarshal(response.Body.Bytes(), &index)
	if index.Schema == "" || len(index.Skills) != 1 {
		t.Fatal("missing skill index")
	}
	skill := httptest.NewRecorder()
	server.Handler().ServeHTTP(skill, httptest.NewRequest("GET", index.Skills[0].URL, nil))
	if index.Skills[0].Digest != fmt.Sprintf("sha256:%x", sha256.Sum256(skill.Body.Bytes())) {
		t.Fatal("digest differs from served artifact")
	}
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", aiCatalogPath, nil))
	var ai struct{ Entries []map[string]json.RawMessage }
	_ = json.Unmarshal(response.Body.Bytes(), &ai)
	if len(ai.Entries) != 3 {
		t.Fatal("missing capability")
	}
	for _, entry := range ai.Entries {
		if (entry["url"] == nil) == (entry["data"] == nil) {
			t.Fatal("exactly one of url/data required")
		}
		var queries []string
		_ = json.Unmarshal(entry["representativeQueries"], &queries)
		if len(queries) < 2 || len(queries) > 5 {
			t.Fatal("missing representative queries")
		}
	}
}

func TestMarkdownNegotiation(t *testing.T) {
	server := discoveryTestServer(t)
	for _, path := range []string{"/", "/components", "/styles", "/changes", "/components/snackbar", "/components/snackbar?q=container"} {
		for _, accept := range []string{"text/markdown", "text/html;q=0.5, text/markdown;q=0.9", "text/markdown;q=0", "text/markdown;q=0.2, text/html", ""} {
			response := httptest.NewRecorder()
			request := httptest.NewRequest("GET", path, nil)
			request.Header.Set("Accept", accept)
			server.Handler().ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("%s returned %d", path, response.Code)
			}
			markdown := wantsMarkdown(request)
			if strings.HasPrefix(response.Header().Get("Content-Type"), "text/markdown") != markdown {
				t.Fatal("incorrect negotiation")
			}
			if !strings.Contains(response.Header().Get("Vary"), "Accept") || !strings.Contains(response.Header().Get("Vary"), "HX-Request") {
				t.Fatal("missing cache variants")
			}
			if markdown && (strings.Contains(response.Body.String(), "<script") || !strings.Contains(response.Body.String(), "# ")) {
				t.Fatalf("invalid Markdown for %s: %s", path, response.Body.String())
			}
		}
	}
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/api/v1/subjects", nil)
	request.Header.Set("Accept", "text/markdown")
	server.Handler().ServeHTTP(response, request)
	if !json.Valid(response.Body.Bytes()) {
		t.Fatal("JSON API was converted")
	}
}

func TestMarkdownPreservesViewContent(t *testing.T) {
	markdown, err := htmlMarkdown([]byte(`<html><head><script>secret</script></head><body><nav>Menu</nav><main><h1>Tokens</h1><p>Read <a href="/components">components</a> here.</p><table><thead><tr><th>Name</th><th>Value</th></tr></thead><tbody><tr><td>md.test</td><td>4 dp</td></tr></tbody></table></main><footer>Footer</footer></body></html>`))
	if err != nil {
		t.Fatal(err)
	}
	for _, content := range []string{"# Tokens", "Read [components](/components) here.", "| Name | Value |", "| md.test | 4 dp |"} {
		if !strings.Contains(markdown, content) {
			t.Fatalf("lost %q: %s", content, markdown)
		}
	}
	if strings.Contains(markdown, "secret") || strings.Contains(markdown, "Menu") || strings.Contains(markdown, "Footer") {
		t.Fatal("included chrome or scripts")
	}
}

func TestContentSignals(t *testing.T) {
	response := httptest.NewRecorder()
	discoveryTestServer(t).Handler().ServeHTTP(response, httptest.NewRequest("GET", "/robots.txt", nil))
	if !strings.Contains(response.Body.String(), "Content-Signal: "+contentSignal) || !strings.Contains(response.Body.String(), "Agentmap: https://tokens.example"+aiCatalogPath) || response.Header().Get("Content-Signal") != contentSignal {
		t.Fatal("content preferences are missing or inconsistent")
	}
}
