package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestCollectionBasePath(t *testing.T) {
	server := New(seoTestState(), "")
	if err := server.SetPublicOrigin("https://tokencatalog.dev/md3/"); err != nil {
		t.Fatal(err)
	}
	handler := server.Handler()
	get := func(path, accept string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest("GET", path, nil)
		r.Header.Set("Accept", accept)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	for _, path := range []string{"/", "/md3"} {
		w := get(path+"?q=hello", "")
		if w.Code != 308 || w.Header().Get("Location") != "/md3/?q=hello" {
			t.Fatalf("root redirect: %d %s", w.Code, w.Header())
		}
	}
	for _, path := range []string{"/md3/", "/md3/components/snackbar", "/md3/components/snackbar/sets/snackbar/common", "/md3/styles", "/md3/changes"} {
		w := get(path, "")
		if w.Code != 200 {
			t.Fatalf("%s: %d", path, w.Code)
		}
		if !strings.Contains(w.Header().Get("Link"), "</md3/openapi.json>") {
			t.Fatal("unmounted discovery Link")
		}
		z := html.NewTokenizer(strings.NewReader(w.Body.String()))
		for z.Next() != html.ErrorToken {
			token := z.Token()
			for _, attr := range token.Attr {
				switch attr.Key {
				case "href", "src", "action", "hx-get":
					if strings.HasPrefix(attr.Val, "/") && !strings.HasPrefix(attr.Val, "/md3/") {
						t.Fatalf("%s: unmounted %s=%s", path, attr.Key, attr.Val)
					}
				}
			}
		}
	}
	w := get("/md3/components/snackbar", "text/markdown")
	if w.Code != 200 || !strings.Contains(w.Header().Get("Content-Type"), "text/markdown") || strings.Contains(w.Body.String(), "https://tokencatalog.dev/components/") {
		t.Fatal("incorrect Markdown base")
	}
	for _, path := range []string{"/robots.txt", "/.well-known/api-catalog", "/.well-known/ai-catalog.json", "/.well-known/agent-skills/index.json", "/llms-full.txt", "/auth.md", "/md3/mcp/server-card", "/md3/api/v1/subjects", "/md3/sitemap.xml", "/md3/static/webmcp.js"} {
		w := get(path, "")
		if w.Code != 200 {
			t.Fatalf("%s: %d", path, w.Code)
		}
		if path == "/md3/static/webmcp.js" {
			continue
		}
		if !strings.Contains(w.Body.String(), "https://tokencatalog.dev/md3") {
			t.Fatalf("%s lacks canonical base", path)
		}
	}
	w = get("/md3/downloads/component/snackbar.json?scope=everything", "")
	if w.Code != 303 || w.Header().Get("Location") != "/md3/downloads/tokens.json?scope=all" {
		t.Fatalf("download redirect: %d %s", w.Code, w.Header())
	}
	for _, path := range []string{"/components/snackbar", "/md30/components/snackbar", "/downloads/tokens.json?scope=all"} {
		if w := get(path, ""); w.Code != 404 {
			t.Fatalf("%s: %d", path, w.Code)
		}
	}
	for _, origin := range []string{"https://tokencatalog.dev", "https://tokencatalog.dev/md3", "https://evil.example"} {
		r := httptest.NewRequest("POST", "/md3/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
		r.Header.Set("Origin", origin)
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Accept", "application/json, text/event-stream")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		want := http.StatusForbidden
		if origin == "https://tokencatalog.dev" {
			want = http.StatusOK
		}
		if w.Code != want {
			t.Fatalf("MCP Origin %s: %d want %d", origin, w.Code, want)
		}
	}
	w = get("/md3/openapi.json", "")
	var doc struct {
		Servers []struct{ URL string }
		Paths   map[string]json.RawMessage
	}
	if err := json.Unmarshal(w.Body.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Servers[0].URL != "https://tokencatalog.dev/md3" {
		t.Fatal("wrong OpenAPI server")
	}
	if !strings.Contains(string(doc.Paths["/downloads/{kind}/{slug}.json"]), `"view"`) {
		t.Fatal("filtered scope missing from OpenAPI")
	}
}
