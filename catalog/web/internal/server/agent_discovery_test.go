// Integration checks for discovery.
package server

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/sojupie/tokens/source/publish"
	"golang.org/x/net/html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestHTMLAgentHintsAndJSONAlternate(t *testing.T) {
	state := seoTestState()
	key := publish.SubjectKey{Kind: publish.SubjectStyle, Slug: "motion/overview"}
	state.Subjects[key] = publish.SubjectView{Key: key, Name: "Motion"}
	componentKey := publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "snackbar"}
	subject := state.Subjects[componentKey]
	subject.Tokens[0].Values = []publish.TokenValueView{
		{ContextLabels: []string{"Theme: Light"}, Value: json.RawMessage(`{"color":{"red":1}}`)},
		{ContextLabels: []string{"Theme: Dark"}, Value: json.RawMessage(`{"color":{"red":0}}`)},
	}
	state.Subjects[componentKey] = subject
	server := New(state, "")
	if err := server.SetPublicOrigin("https://tokens.example"); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/", "/components", "/styles", "/changes", "/components/snackbar", "/styles/motion/overview", "/components/snackbar/sets/snackbar/color?q=container&context.Theme=Dark"} {
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		document, err := html.Parse(response.Body)
		if err != nil || response.Code != http.StatusOK {
			t.Fatalf("cannot read page %s", path)
		}
		links := map[string]string{}
		alternate := ""
		footerLink := false
		var walk func(*html.Node, bool)
		walk = func(node *html.Node, inFooter bool) {
			inFooter = inFooter || node.Type == html.ElementNode && node.Data == "footer"
			attrs := map[string]string{}
			for _, attr := range node.Attr {
				attrs[attr.Key] = attr.Val
			}

			if node.Type == html.ElementNode && node.Data == "link" {
				links[attrs["href"]] = attrs["rel"]
				if attrs["rel"] == "alternate" && attrs["type"] == "application/json" {
					alternate = attrs["href"]
				}
			}
			if inFooter && node.Data == "a" && attrs["href"] == "/llms-full.txt" && node.FirstChild != nil && node.FirstChild.Data == "API reference" {
				footerLink = true
			}
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				walk(child, inFooter)
			}
		}
		walk(document, false)

		if links["/agents.json"] != "describedby" || links["/llms.txt"] != "describedby" || links["/openapi.json"] != "service-desc" || links[apiCatalogPath] != "api-catalog" || links[aiCatalogPath] != "ai-catalog" || !footerLink {
			t.Fatalf("%s missing HTML hints/footer guide", path)
		}
		if !strings.Contains(response.Header().Get("Link"), `</agents.json>; rel="describedby"`) {
			t.Fatal("incorrect discovery header relation")
		}
		for _, hint := range []string{`</.well-known/api-catalog>; rel="api-catalog"`, `</.well-known/ai-catalog.json>; rel="ai-catalog"`} {
			if !strings.Contains(response.Header().Get("Link"), hint) {
				t.Fatalf("missing catalog header: %s", hint)
			}
		}
		if strings.HasPrefix(path, "/components/") || strings.HasPrefix(path, "/styles/") {
			if alternate == "" {
				t.Fatal("missing JSON alternative for " + path)
			}
			export := httptest.NewRecorder()
			server.Handler().ServeHTTP(export, httptest.NewRequest(http.MethodGet, alternate, nil))
			if export.Code != http.StatusOK || !strings.HasPrefix(export.Header().Get("Content-Type"), "application/json") {
				t.Fatal("JSON alternative is not readable")
			}
			if strings.Contains(path, "context.Theme") {
				parsed, err := url.Parse(alternate)
				if err != nil || parsed.Query().Get("context.Theme") != "Dark" || parsed.Query().Get("q") != "container" || parsed.Query().Get("scope") != "view" || parsed.Query().Get("set") != "snackbar|color|" {
					t.Fatalf("JSON alternative lost the table filters: %s", alternate)
				}
			}
		} else if alternate != "" {
			t.Fatal("index/history page advertised a token-table alternative")
		}
	}
	preview := httptest.NewRecorder()
	New(state, "").Handler().ServeHTTP(preview, httptest.NewRequest(http.MethodGet, "/", nil))
	if strings.Contains(preview.Body.String(), `href="/agents.json"`) || strings.Contains(preview.Body.String(), "API and agent guide") || preview.Header().Get("Link") != "" {
		t.Fatal("preview advertises discovery routes that require a public origin")
	}
}

func TestAgentDiscoveryAndDataLinks(t *testing.T) {
	state := seoTestState()
	key := publish.SubjectKey{Kind: publish.SubjectStyle, Slug: "motion/overview"}
	state.Subjects[key] = publish.SubjectView{Key: key, Name: "Motion"}
	server := New(state, "")
	if err := server.SetPublicOrigin("https://tokens.example"); err != nil {
		t.Fatal(err)
	}
	handler := server.Handler()
	get := func(path string) *httptest.ResponseRecorder {
		t.Helper()
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.Host = "untrusted.example"
		request.Header.Set("X-Forwarded-Host", "untrusted.example")
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s: status %d: %s", path, response.Code, response.Body.String())
		}
		return response
	}
	for _, path := range []string{"/agents.txt", "/agents.json", "/llms.txt", "/llms-full.txt", "/openapi.json", tokenSkillPath} {
		response := get(path)
		if response.Header().Get("Access-Control-Allow-Origin") != "*" || response.Header().Get("Cache-Control") != "public, no-cache" || strings.Contains(response.Body.String(), "untrusted.example") {
			t.Fatalf("wrong discovery headers/origin for %s", path)
		}
		contentType := "text/plain; charset=utf-8"
		if strings.HasSuffix(path, ".json") {
			contentType = "application/json; charset=utf-8"
		}
		if response.Header().Get("Content-Type") != contentType {
			t.Fatal("wrong content type for " + path)
		}
	}
	for _, path := range []string{"/llms.txt", "/llms-full.txt"} {
		body := get(path).Body.String()
		for _, target := range []string{apiCatalogPath, aiCatalogPath, skillIndexPath, "/mcp/server-card", "/auth.md"} {
			if !strings.Contains(body, "(https://tokens.example"+target+")") {
				t.Fatalf("%s missing discovery link %s", path, target)
			}
			get(target)
		}
	}
	if !strings.Contains(get("/llms.txt").Body.String(), "(https://tokens.example/llms-full.txt)") {
		t.Fatal("short guide does not link to full guide")
	}
	for _, path := range []string{"/llms-full.txt", tokenSkillPath} {
		body := get(path).Body.String()
		for _, instruction := range []string{"MCP uses POST", "Accept: text/markdown", "tools/list", "get_tokens", "MCP-Protocol-Version: " + mcpProtocol} {
			if !strings.Contains(body, instruction) {
				t.Fatalf("%s missing %q", path, instruction)
			}
		}
	}
	var discovery map[string]json.RawMessage
	if err := json.Unmarshal(get("/agents.json").Body.Bytes(), &discovery); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"payments", "authorization", "mcp", "a2a", "ucp", "webmcp"} {
		if _, exists := discovery[name]; exists {
			t.Fatal("advertised unsupported capability " + name)
		}
	}
	var skills []struct{ URL string }
	if err := json.Unmarshal(discovery["skills"], &skills); err != nil || len(skills) != 1 {
		t.Fatal("missing skill declaration")
	}
	if !strings.Contains(get("/agents.txt").Body.String(), "Skills: "+skills[0].URL) || !strings.HasPrefix(get(skills[0].URL).Body.String(), "---\nname: material-tokens\n") {
		t.Fatal("text and JSON skill declarations disagree")
	}
	var index componentListResponse
	if err := json.Unmarshal(get("/api/v1/subjects").Body.Bytes(), &index); err != nil || len(index.Subjects) != 2 {
		t.Fatal("incomplete subject index")
	}
	for _, subject := range index.Subjects {
		get(subject.URL)
		response := get(subject.JSONURL)
		var document publish.PublicTokenDocument
		if err := json.Unmarshal(response.Body.Bytes(), &document); err != nil || document.Kind != subject.Kind || document.Slug != subject.Slug {
			t.Fatal("index JSON URL does not match subject")
		}
		if response.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Fatal("browser agents cannot read JSON downloads")
		}
	}
	for _, path := range []string{"/api/v1/subjects", "/downloads/tokens.json?scope=all", "/agents.json"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodOptions, path, nil))
		if response.Code != http.StatusNoContent || response.Header().Get("Access-Control-Allow-Origin") != "*" || response.Header().Get("Access-Control-Allow-Credentials") != "" {
			t.Fatal("wrong public read-only CORS policy")
		}
	}
	var spec map[string]any
	if err := json.Unmarshal(get("/openapi.json").Body.Bytes(), &spec); err != nil {
		t.Fatal(err)
	}
	compiler := jsonschema.NewCompiler()
	const location = "https://tokens.example/openapi.json"
	if err := compiler.AddResource(location, spec); err != nil {
		t.Fatal(err)
	}
	checks := map[string]string{
		"/api/v1/health":                "/api/v1/health",
		"/api/v1/subjects":              "/api/v1/subjects",
		"/api/v1/releases":              "/api/v1/releases",
		"/api/v1/changes":               "/api/v1/changes",
		"/downloads/tokens.json":        "/downloads/tokens.json?scope=all",
		"/downloads/{kind}/{slug}.json": "/downloads/style/motion/overview.json?scope=all",
	}
	for path, actual := range checks {
		pointer := strings.ReplaceAll(path, "/", "~1")
		schema, err := compiler.Compile(fmt.Sprintf("%s#/paths/%s/get/responses/200/content/application~1json/schema", location, pointer))
		if err != nil {
			t.Fatal(err)
		}
		var value any
		if err := json.Unmarshal(get(actual).Body.Bytes(), &value); err != nil {
			t.Fatal(err)
		}
		if err := schema.Validate(value); err != nil {
			t.Fatalf("API description disagrees with %s: %v", actual, err)
		}
	}
}

func TestSitemapReadableAndPreviewDiscovery(t *testing.T) {
	server := New(seoTestState(), "")
	for _, path := range []string{"/agents.txt", "/agents.json", "/llms.txt", "/openapi.json"} {
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusServiceUnavailable {
			t.Fatal("preview must not advertise an invented public origin")
		}
	}
	if err := server.SetPublicOrigin("https://tokens.example"); err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil))
	if !strings.Contains(response.Body.String(), "\n  <url>\n    <loc>") || !strings.Contains(response.Body.String(), `href="/sitemap.css"`) {
		t.Fatal("sitemap should separate URLs in source and browser display")
	}
}

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
	markdown, err := htmlMarkdown([]byte(`<html><head><script>secret</script></head><body><nav>Menu</nav><main><h1>Tokens</h1><section id="component-preview"><h2>Preview</h2><p>Demo email text</p><ul><li><button>Only action</button></li></ul></section><div aria-hidden="true">Decorative content</div><p>Read <a href="/components">components</a> here.</p><table><thead><tr><th>Name</th><th>Value</th></tr></thead><tbody><tr><td>md.test</td><td>4 dp</td></tr></tbody></table><nav><a href="?page=2">Next</a><a href="https://example.org/docs">Documentation</a></nav></main><footer>Footer</footer></body></html>`), "https://tokens.example/changes")
	if err != nil {
		t.Fatal(err)
	}
	for _, content := range []string{"# Tokens", "Read [components](https://tokens.example/components) here.", "| Name | Value |", "| md.test | 4 dp |", "[Next](https://tokens.example/changes?page=2)\n[Documentation](https://example.org/docs)"} {
		if !strings.Contains(markdown, content) {
			t.Fatalf("lost %q: %s", content, markdown)
		}
	}
	if strings.Contains(markdown, "secret") || strings.Contains(markdown, "Menu") || strings.Contains(markdown, "Footer") || strings.Contains(markdown, "Demo email") || strings.Contains(markdown, "Decorative content") || strings.Contains(markdown, "## Preview") {
		t.Fatal("included chrome or scripts")
	}
}

func TestMarkdownSubjectLinks(t *testing.T) {
	server := discoveryTestServer(t)
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/components/snackbar", nil)
	request.Host = "untrusted.example"
	request.Header.Set("Accept", "text/markdown")
	server.Handler().ServeHTTP(response, request)
	body := response.Body.String()
	for _, link := range []string{"[Change history (0)](https://tokens.example/components/snackbar/changes)"} {
		if !strings.Contains(body, link) {
			t.Fatalf("missing %q: %s", link, body)
		}
	}
	if strings.Contains(body, "untrusted.example") {
		t.Fatal("links used the request host instead of the configured origin")
	}
}

func TestMarkdownTokenTableFormatting(t *testing.T) {
	source := []byte(`<main><noscript><style>secret CSS</style></noscript><details><summary>Enabled</summary><table><caption>Enabled tokens</caption><thead><tr><th scope="col">Token</th><th scope="col">Value</th></tr></thead><tbody><tr><th scope="row"><span>Container color</span><code>md.comp.container_color</code></th><td><details class="tokens-token-reference"><summary><code>md.sys.color</code><span class="tokens-reference-toggle">Show reference chainHide reference chain</span></summary><div><code>md.comp.container_color</code><ul><li><code>md.sys.color</code></li><li><code>md.ref.color</code></li></ul></div></details><span>#322F35</span></td></tr><tr><th scope="row">Width</th><td>4 dp</td></tr></tbody></table></details></main>`)
	result, err := htmlMarkdown(source, "https://tokens.example/components/test")
	if err != nil {
		t.Fatal(err)
	}
	for _, content := range []string{"### Enabled\n\n| Token | Value |", "Container color `md.comp.container_color`", "reference chain: `md.comp.container_color` → `md.sys.color` → `md.ref.color`", "#322F35", "| Width | 4 dp |"} {
		if !strings.Contains(result, content) {
			t.Fatalf("lost %q: %s", content, result)
		}
	}
	if strings.Count(result, "| --- | --- |") != 1 || strings.Contains(result, "secret CSS") || strings.Contains(result, "Show reference") || strings.Contains(result, "Hide reference") {
		t.Fatalf("included UI text or repeated table headers: %s", result)
	}
}

func TestMarkdownSubpageViews(t *testing.T) {
	state := seoTestState()
	style := state.Subjects[publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "snackbar"}]
	style.Key = publish.SubjectKey{Kind: publish.SubjectStyle, Slug: "elevation"}
	style.Name = "Elevation"
	state.Subjects[style.Key] = style
	for i := 0; i < 51; i++ {
		state.Changes = append(state.Changes, publish.ChangeView{ID: fmt.Sprint(i), Subject: "snackbar", Type: "changed", Name: "md.comp.snackbar.container.width"})
	}
	server := New(state, "")
	if err := server.SetPublicOrigin("https://tokens.example"); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/styles/elevation", "/components/snackbar?q=container", "/components/snackbar/sets/snackbar/common", "/components/snackbar?context.Theme=Dark", "/changes", "/components/snackbar/changes", "/components/snackbar/changes?page=2"} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest("GET", path, nil)
			request.Header.Set("Accept", "text/markdown")
			server.Handler().ServeHTTP(response, request)
			body := response.Body.String()
			if response.Code != 200 || !strings.Contains(body, "# ") || strings.Contains(body, "&lt;style") || strings.Contains(body, "\n \n") {
				t.Fatalf("invalid subpage Markdown: %d %s", response.Code, body)
			}
			if path == "/components/snackbar/changes" && !strings.Contains(body, "[Next](https://tokens.example/components/snackbar/changes?page=2)") {
				t.Fatalf("pagination link missing: %s", body)
			}
		})
	}
}

func TestContentSignals(t *testing.T) {
	response := httptest.NewRecorder()
	discoveryTestServer(t).Handler().ServeHTTP(response, httptest.NewRequest("GET", "/robots.txt", nil))
	if !strings.Contains(response.Body.String(), "Content-Signal: "+contentSignal) || strings.Contains(response.Body.String(), "Agentmap:") || response.Header().Get("Content-Signal") != contentSignal {
		t.Fatal("content preferences are missing or inconsistent")
	}
}

func TestEmptyChangesReturnsArray(t *testing.T) {
	response := httptest.NewRecorder()
	discoveryTestServer(t).Handler().ServeHTTP(response, httptest.NewRequest("GET", "/api/v1/changes", nil))
	var body struct {
		Changes json.RawMessage `json:"changes"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if response.Code != 200 || string(body.Changes) != "[]" {
		t.Fatalf("empty changes must be an array: %d %s", response.Code, response.Body.String())
	}
}

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
