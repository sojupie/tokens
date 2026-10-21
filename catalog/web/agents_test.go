package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/sojupie/tokens/source/publish"
	"golang.org/x/net/html"
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
		footerColumns := []string{}
		var walk func(*html.Node, bool)
		walk = func(node *html.Node, inFooter bool) {
			inFooter = inFooter || node.Type == html.ElementNode && node.Data == "footer"
			attrs := map[string]string{}
			for _, attr := range node.Attr {
				attrs[attr.Key] = attr.Val
			}
			if attrs["class"] == "tokens-footer-content" {
				for child := node.FirstChild; child != nil; child = child.NextSibling {
					if child.Type != html.ElementNode || child.Data != "div" {
						continue
					}
					var text strings.Builder
					var collectText func(*html.Node)
					collectText = func(item *html.Node) {
						if item.Type == html.TextNode {
							text.WriteString(item.Data)
						}
						for nested := item.FirstChild; nested != nil; nested = nested.NextSibling {
							collectText(nested)
						}
					}
					collectText(child)
					footerColumns = append(footerColumns, text.String())
				}
			}
			if node.Type == html.ElementNode && node.Data == "link" {
				links[attrs["href"]] = attrs["rel"]
				if attrs["rel"] == "alternate" && attrs["type"] == "application/json" {
					alternate = attrs["href"]
				}
			}
			if inFooter && node.Data == "a" && attrs["href"] == "/llms-full.txt" && node.FirstChild != nil && node.FirstChild.Data == "API and data reference" {
				footerLink = true
			}
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				walk(child, inFooter)
			}
		}
		walk(document, false)
		if len(footerColumns) != 4 || !strings.Contains(footerColumns[0], "Download tokens") || !strings.HasPrefix(footerColumns[1], "For humans") || !strings.Contains(footerColumns[2], "API and data reference") || !strings.Contains(footerColumns[1], "Give this to your agent") || !strings.HasPrefix(footerColumns[3], "About") || !strings.Contains(footerColumns[3], "Made by Olle Marmenlind") {
			t.Fatalf("wrong footer columns: %v", footerColumns)
		}
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
