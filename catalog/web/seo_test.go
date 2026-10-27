package web

import (
	"encoding/xml"
	"html"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sojupie/tokens/source/publish"
)

func seoTestState() publish.PublishedState {
	key := publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "snackbar"}
	return publish.PublishedState{Subjects: map[publish.SubjectKey]publish.SubjectView{
		key: {Key: key, Name: "Snackbar", Tokens: []publish.TokenView{
			{SourceName: "md.comp.snackbar.container.color", Category: "COLOR"},
			{SourceName: "md.comp.snackbar.container.width", Category: "LENGTH"},
		}},
	}}
}

func TestSEOMetadataAndCanonicalURLs(t *testing.T) {
	state := seoTestState()
	for i := 0; i < 51; i++ {
		state.Changes = append(state.Changes, publish.ChangeView{ID: "change", Subject: "snackbar", Type: "changed", Name: "md.comp.snackbar.container.width"})
	}
	server := New(state, "")
	if err := server.SetPublicOrigin("https://tokens.example/"); err != nil {
		t.Fatal(err)
	}
	for _, check := range []struct {
		path, title, canonical string
		noindex                bool
	}{
		{"/", "Material Design 3 Tokens and Change History", "https://tokens.example/", false},
		{"/components/snackbar", `Material Design 3 "Snackbar" tokens`, "https://tokens.example/components/snackbar", false},
		{"/components/snackbar/sets/snackbar/common", `Material Design 3 "Snackbar" tokens: Common`, "https://tokens.example/components/snackbar/sets/snackbar/common", false},
		{"/components/snackbar?set=snackbar%7Ccommon%7C", `Material Design 3 "Snackbar" tokens: Common`, "https://tokens.example/components/snackbar/sets/snackbar/common", false},
		{"/components/snackbar?utm_source=example", `Material Design 3 "Snackbar" tokens`, "https://tokens.example/components/snackbar", false},
		{"/components/snackbar/changes?page=2", `Material Design 3 "Snackbar" token changes: Page 2`, "https://tokens.example/components/snackbar/changes?page=2", false},
		{"/components?q=snack", "Material Design 3 Components and Component Tokens", "", true},
		{"/components/snackbar?context.Theme=Dark", `Material Design 3 "Snackbar" tokens`, "", true},
		{"/changes?snapshot=one", "Material Design 3 Token Change Log", "", true},
		{"/changes?type=removed", "Material Design 3 Token Change Log", "", true},
	} {
		t.Run(check.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, check.path, nil)
			request.Host = "untrusted.example"
			request.Header.Set("X-Forwarded-Host", "untrusted.example")
			server.Handler().ServeHTTP(response, request)
			body := response.Body.String()
			if response.Code != http.StatusOK || !strings.Contains(html.UnescapeString(body), "<title>"+check.title+" | Token Catalog</title>") || !strings.Contains(body, `<meta name="description" content="`) {
				t.Fatalf("%s: unexpected metadata/status %d", check.path, response.Code)
			}
			if check.canonical != "" && !strings.Contains(body, `rel="canonical" href="`+check.canonical+`"`) {
				t.Fatalf("%s: missing canonical %s", check.path, check.canonical)
			}
			if check.canonical == "" && strings.Contains(body, `rel="canonical"`) {
				t.Fatal("filtered page should not claim to duplicate an unfiltered page")
			}
			if strings.Contains(body, `content="noindex, follow"`) != check.noindex || strings.Contains(body, "untrusted.example") {
				t.Fatal("wrong indexing rule or untrusted canonical origin")
			}
			if response.Header().Get("Vary") != "HX-Request, HX-Target, Accept" {
				t.Fatal("full HTML and HTMX fragments must be separate cache variants")
			}
		})
	}
}

func TestSEOIndexPagesHaveDistinctTitlesAndDescriptions(t *testing.T) {
	server := New(seoTestState(), "")
	if err := server.SetPublicOrigin("https://tokens.example"); err != nil {
		t.Fatal(err)
	}
	seenTitles := map[string]string{}
	seenDescriptions := map[string]string{}
	for _, path := range []string{"/", "/for-agents", "/components", "/styles", "/changes"} {
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		body := response.Body.String()
		titleStart := strings.Index(body, "<title>")
		titleEnd := strings.Index(body, "</title>")
		descriptionStart := strings.Index(body, `<meta name="description" content="`)
		if response.Code != http.StatusOK || titleStart < 0 || titleEnd < titleStart || descriptionStart < 0 {
			t.Fatalf("%s: page metadata missing", path)
		}
		title := body[titleStart+len("<title>") : titleEnd]
		descriptionStart += len(`<meta name="description" content="`)
		descriptionEnd := strings.Index(body[descriptionStart:], `">`)
		if descriptionEnd < 0 {
			t.Fatalf("%s: description is not closed", path)
		}
		description := body[descriptionStart : descriptionStart+descriptionEnd]
		if previous := seenTitles[title]; previous != "" {
			t.Errorf("%s and %s share title %q", previous, path, title)
		}
		if previous := seenDescriptions[description]; previous != "" {
			t.Errorf("%s and %s share description %q", previous, path, description)
		}
		seenTitles[title] = path
		seenDescriptions[description] = path
	}
}

func TestSitemapAndRobots(t *testing.T) {
	server := New(seoTestState(), "")
	if err := server.SetPublicOrigin("https://tokens.example"); err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil))
	var document sitemapDocument
	if err := xml.Unmarshal(response.Body.Bytes(), &document); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusOK || document.XMLNS != "http://www.sitemaps.org/schemas/sitemap/0.9" || len(document.URLs) != 10 {
		t.Fatalf("unexpected sitemap: %s", response.Body.String())
	}
	seen := map[string]bool{}
	for _, entry := range document.URLs {
		if seen[entry.Location] || !strings.HasPrefix(entry.Location, "https://tokens.example/") || strings.Contains(entry.Location, "?") {
			t.Fatalf("unexpected sitemap URL %s", entry.Location)
		}
		seen[entry.Location] = true
		page := httptest.NewRecorder()
		server.Handler().ServeHTTP(page, httptest.NewRequest(http.MethodGet, entry.Location, nil))
		if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), `rel="canonical" href="`+entry.Location+`"`) {
			t.Fatalf("sitemap URL is not a self-canonical HTML page: %s", entry.Location)
		}
	}
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))
	if !strings.Contains(response.Body.String(), "Sitemap: https://tokens.example/sitemap.xml") || strings.Contains(response.Body.String(), "Disallow: /") {
		t.Fatalf("unexpected robots.txt: %s", response.Body.String())
	}
	for _, path := range []string{"/api/v1/health", "/downloads/component/snackbar.json"} {
		response = httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Header().Get("X-Robots-Tag") != "noindex" {
			t.Fatalf("%s: missing noindex header", path)
		}
	}
}

func TestPublicOriginValidationAndLocalIndexing(t *testing.T) {
	server := New(seoTestState(), "")
	for _, origin := range []string{"example.com", "ftp://example.com", "https://user:pass@example.com", "https://example.com/a/../b", "https://example.com//md3", "https://example.com?query=one", "https://example.com#fragment", "https://example.com?"} {
		if err := server.SetPublicOrigin(origin); err == nil {
			t.Errorf("accepted invalid origin %q", origin)
		}
	}
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if !strings.Contains(response.Body.String(), `content="noindex, follow"`) || strings.Contains(response.Body.String(), `rel="canonical"`) {
		t.Fatal("unconfigured/local site should not be indexed or invent an origin")
	}
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatal("sitemap without configured origin should not be published")
	}
}

func TestTokenSetPagesAreLinkedAndServerRendered(t *testing.T) {
	handler := New(seoTestState(), "").Handler()
	for _, path := range []string{"/components/snackbar", "/components/snackbar/sets/snackbar/common"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("%s: status %d", path, response.Code)
		}
		body := response.Body.String()
		if strings.Contains(body, "tokens-token-set-links") || strings.Contains(body, `<select name="set">`) || strings.Count(body, `aria-label="Token sets"`) != 1 {
			t.Fatal("page has redundant token-set navigation")
		}
		if !strings.Contains(body, `popover="auto"`) || !strings.Contains(body, `aria-current="page"`) || !strings.Contains(html.UnescapeString(body), `Material Design 3 "Snackbar" tokens`) {
			t.Fatal("missing native popover, current selection or clear main heading")
		}
		for _, link := range []string{"", "/sets/snackbar/common", "/sets/all"} {
			if !strings.Contains(body, `href="/components/snackbar`+link+`"`) {
				t.Fatalf("%s: missing token set link %s", path, link)
			}
		}
		if strings.HasSuffix(path, "/common") && (!strings.Contains(body, "md.comp.snackbar.container.width") || strings.Contains(body, "md.comp.snackbar.container.color")) {
			t.Fatal("token set page does not render the selected set")
		}
	}
	for _, path := range []string{"/components/snackbar/sets/", "/components/snackbar/sets/snackbar/unknown", "/components/snackbar/sets/snackbar/color/unknown"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusNotFound {
			t.Fatalf("%s: status %d, want 404", path, response.Code)
		}
	}
}

func TestTokenSetRouteResolvesSourceNamesWithSpacesAndCapitals(t *testing.T) {
	state := seoTestState()
	key := publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "snackbar"}
	subject := state.Subjects[key]
	subject.Tokens = []publish.TokenView{{SourceName: "md.comp.Outlined Icon Button.container.color", Category: "COLOR"}}
	state.Subjects[key] = subject
	response := httptest.NewRecorder()
	New(state, "").Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/components/snackbar/sets/outlined-icon-button/color", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "md.comp.Outlined Icon Button.container.color") {
		t.Fatalf("source label did not round-trip through its route: status %d", response.Code)
	}
}
