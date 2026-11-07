// Integration checks for tools.
package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"github.com/sojupie/tokens/catalog/web/internal/playground"
	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	webrender "github.com/sojupie/tokens/catalog/web/internal/render"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestToolsLinksLeaveTheCatalogMount(t *testing.T) {
	server := New(seoTestState(), "")
	if err := server.SetPublicOrigin("https://tokens.example/md3"); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/md3/", "/md3/tools"} {
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		body := w.Body.String()
		for _, link := range []string{`href="/md3/tools"`, `href="https://tokens.example/tools/motion/"`, `href="https://tokens.example/tools/motion-lab/"`, `hx-boost="false"`} {
			if w.Code != 200 || !strings.Contains(body, link) {
				t.Fatalf("%s: missing tools navigation %s (status %d)", path, link, w.Code)
			}
		}
		if strings.Contains(body, `href="/md3/tools/motion/"`) {
			t.Fatal("tool app incorrectly mounted inside catalog")
		}
	}
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/md3/tools/?q=test", nil))
	if w.Code != http.StatusPermanentRedirect || w.Header().Get("Location") != "/md3/tools?q=test" {
		t.Fatalf("noncanonical tools URL: %d %s", w.Code, w.Header().Get("Location"))
	}
}
func TestToolURLsAreConfigurableAndValidated(t *testing.T) {
	server := New(seoTestState(), "")
	if err := server.SetToolURLs("https://motion.example/", "/lab/"); err != nil {
		t.Fatal(err)
	}
	destinations := server.toolDestinations()
	if destinations[0].URL != "https://motion.example/" || destinations[1].URL != "/lab/" {
		t.Fatal(destinations)
	}
	for _, target := range []string{"", "javascript:alert(1)", "//other.example/", "https://user:secret@example.com/", "relative/path"} {
		if err := server.SetToolURLs(target, "/lab/"); err == nil {
			t.Fatalf("accepted invalid URL %q", target)
		}
	}
}

func TestToolsOutsideCollectionMount(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"motion", "motion-lab"} {
		if err := os.MkdirAll(filepath.Join(dir, name, "assets"), 0755); err != nil {
			t.Fatal(err)
		}
		for file, body := range map[string]string{"index.html": name, "assets/engine.wasm": "wasm", "sw.js": "worker"} {
			if err := os.WriteFile(filepath.Join(dir, name, file), []byte(body), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	var encoded bytes.Buffer
	compressor := gzip.NewWriter(&encoded)
	_, _ = compressor.Write([]byte("motion"))
	_ = compressor.Close()
	if err := os.WriteFile(filepath.Join(dir, "motion", "index.html.gz"), encoded.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	fallback := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/md3/components" {
			_, _ = io.WriteString(w, "catalog")
		} else {
			http.NotFound(w, r)
		}
	})
	handler, err := MountTools(fallback, dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		method, path string
		status       int
		body         string
	}{
		{"GET", "/tools/motion/", 200, "motion"}, {"GET", "/tools/motion-lab/", 200, "motion-lab"},
		{"GET", "/md3/components", 200, "catalog"}, {"GET", "/tools/motion/sw.js", 200, "worker"},
		{"GET", "/tools/motion/assets/engine.wasm", 200, "wasm"}, {"HEAD", "/tools/motion/", 200, ""},
		{"POST", "/tools/motion/import", 405, ""}, {"GET", "/tools/motion/assets/", 404, ""},
		{"GET", "/tools/motion/api/projects", 404, ""}, {"GET", "/tools/motion/index.html.gz", 404, ""},
	} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
		if w.Code != tc.status || (tc.body != "" && w.Body.String() != tc.body) {
			t.Fatalf("%s %s: %d %s", tc.method, tc.path, w.Code, w.Body.String())
		}
	}
	request := httptest.NewRequest("GET", "/tools/motion/", nil)
	request.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, request)
	if w.Header().Get("Content-Encoding") != "gzip" || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/html") {
		t.Fatal(w.Header())
	}
	request = httptest.NewRequest("GET", "/tools/motion/assets/engine.wasm", nil)
	request.Header.Set("Range", "bytes=0-1")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, request)
	if w.Code != 206 || w.Body.String() != "wa" || w.Header().Get("Content-Encoding") != "" {
		t.Fatal(w.Code, w.Body.String())
	}
	if _, err := MountTools(fallback, t.TempDir()); err == nil {
		t.Fatal("missing tool builds accepted")
	}
}

func TestNavigationRailConfigurations(t *testing.T) {
	page := &webview.ComponentPreview{Cases: []webview.ComponentCase{{Family: "nav-rail", Variant: "expanded", Orientation: "vertical"}}}
	if err := playground.ConfigureNavigationPreview(page, map[string][]string{"preview.layout": {"modal"}, "preview.alignment": {"center"}, "preview.header": {"menu-fab"}, "preview.badge": {"labeled"}, "preview.indicator": {"fill"}}); err != nil {
		t.Fatal(err)
	}
	c := page.Cases[0]
	if webview.PreviewDimensions(c)["orientation"] != "" {
		t.Fatal("rail label layout must follow its variant, not filter variants out")
	}
	if c.NavIndicator != "fill" || c.Orientation != "horizontal" || c.NavLayout != "modal" || c.NavAlign != "center" || !c.RailFAB() || !c.RailMenu() {
		t.Fatalf("lost configuration: %#v", c)
	}
	if c.Attributes()["data-variant"] != "expanded" || c.Attributes()["data-layout"] != "modal" {
		t.Fatal("modal must configure expanded rather than replace it")
	}
	var html bytes.Buffer
	if err := webrender.ComponentSpecimen(c).Render(context.Background(), &html); err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"md-rail-frame", "md-rail-body", "data-navigation-secondary", "data-navigation-create"} {
		if !strings.Contains(html.String(), required) {
			t.Fatalf("missing %s", required)
		}
	}
	if !strings.Contains(webview.MaterialSymbolSVG("keyboard_command_key"), "<path") {
		t.Fatal("command shortcut icon missing")
	}
}
func TestNavigationDrawerContent(t *testing.T) {
	var html bytes.Buffer
	if err := webrender.ComponentSpecimen(webview.ComponentCase{Family: "nav-drawer", Variant: "modal", State: "interactive"}).Render(context.Background(), &html); err != nil {
		t.Fatal(err)
	}
	if strings.Count(html.String(), "data-navigation-item") != 8 || !strings.Contains(html.String(), "Workspace") {
		t.Fatal("drawer sections are incomplete")
	}
	if strings.Count(html.String(), "<hr") != 1 || strings.Contains(html.String(), ">Close</button>") {
		t.Fatal("drawer must have only the section divider and no footer close button")
	}
}

func TestRailMenuExistsOnlyInsideConfiguredRail(t *testing.T) {
	for _, variant := range []string{"collapsed", "expanded"} {
		for _, header := range []string{"none", "fab", "menu", "menu-fab"} {
			c := webview.ComponentCase{Family: "nav-rail", Variant: variant, State: "interactive", NavHeader: header}
			var output bytes.Buffer
			if err := webrender.ComponentSpecimen(c).Render(context.Background(), &output); err != nil {
				t.Fatal(err)
			}
			html := output.String()
			count := strings.Count(html, "data-navigation-expand")
			if strings.Contains(html, "md-rail-toggle") {
				t.Fatal("external menu rendered")
			}
			if c.RailMenu() {
				start, end := strings.Index(html, "<nav "), strings.Index(html, "</nav>")
				if count != 1 || start < 0 || end < start || !strings.Contains(html[start:end], "data-navigation-expand") {
					t.Fatalf("%s/%s: menu not exclusively inside rail", variant, header)
				}
			} else if count != 0 {
				t.Fatalf("%s/%s: menu rendered without configuration", variant, header)
			}
		}
	}
}

func TestRailAdditionalDestinationsAreExpandedOnly(t *testing.T) {
	c := webview.ComponentCase{Family: "nav-rail"}
	for i := range 6 {
		p := webrender.NavigationExampleItem(c, i, c.NavigationLabels()[i], c.NavigationIcon(i))
		if p.Secondary != (i >= 4) {
			t.Fatalf("item %d: wrong visibility policy", i)
		}
	}
}

func TestRailInteractiveStates(t *testing.T) {
	for _, state := range []string{"collapsed", "expanded", "hidden"} {
		for _, header := range []string{"none", "fab", "menu", "menu-fab"} {
			page := &webview.ComponentPreview{Interactive: true, Controls: []webview.PreviewControl{{Name: "variant"}}, Cases: []webview.ComponentCase{{Family: "nav-rail", Variant: "collapsed", State: "interactive"}}}
			if err := playground.ConfigureNavigationPreview(page, map[string][]string{"preview.rail-state": {state}, "preview.header": {header}}); err != nil {
				t.Fatal(err)
			}
			c := page.Cases[0]
			for _, control := range page.Controls {
				if control.Name == "variant" || control.Name == "collapse" || control.Name == "secondary" {
					t.Fatalf("redundant control %s", control.Name)
				}
				if state == "hidden" && control.Name == "header" && (playground.ContainsPreview(control.Options, "none") || playground.ContainsPreview(control.Options, "fab")) {
					t.Fatal("hidden mode offered no-menu header")
				}
			}
			if state == "hidden" {
				if c.NavCollapse != "hidden" || !c.RailMenu() {
					t.Fatal("hidden rail needs a menu")
				}
			} else if c.Variant != state || c.NavCollapse == "hidden" {
				t.Fatalf("wrong initial state: %#v", c)
			}
			var output bytes.Buffer
			if err := webrender.ComponentSpecimen(c).Render(context.Background(), &output); err != nil {
				t.Fatal(err)
			}
			html := output.String()
			if state == "hidden" && strings.Count(html, "data-navigation-expand") != 1 {
				t.Fatal("immersive rail must have exactly one persistent toggle")
			}
			if strings.Contains(html, "md-rail-menu-slot") != (state == "hidden") {
				t.Fatal("immersive menu must retain header layout slots")
			}
			if strings.Contains(html, "md-rail-toggle") != (state == "hidden") {
				t.Fatal("persistent toggle must be exclusive to hidden mode")
			}
		}
	}
}

func TestRailVariantComparisonKeepsExpandedAtNarrowWidth(t *testing.T) {
	cases := []webview.ComponentCase{
		{Family: "nav-rail", Variant: "collapsed", Width: "default", Size: "default", State: "enabled"},
		{Family: "nav-rail", Variant: "collapsed", Width: "narrow", Size: "default", State: "enabled"},
		{Family: "nav-rail", Variant: "expanded", Width: "default", Size: "default", State: "enabled"},
	}
	for _, width := range []string{"default", "narrow"} {
		page, err := playground.SelectComponentPreview(cases, url.Values{"preview.compare": {"variants"}, "preview.width": {width}, "preview.rail-state": {"hidden"}})
		if err != nil {
			t.Fatal(err)
		}
		if len(page.Cases) != 2 || page.Cases[0].Variant != "collapsed" || page.Cases[1].Variant != "expanded" || page.Cases[0].Width != width {
			t.Fatalf("%s: variants were filtered out: %#v", width, page.Cases)
		}
		for _, control := range page.Controls {
			if control.Name == "rail-state" {
				t.Fatal("interactive initial state leaked into static comparison")
			}
		}
	}
}
