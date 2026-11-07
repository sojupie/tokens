// Integration checks for delivery.
package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/a-h/templ"
	webassets "github.com/sojupie/tokens/catalog/web/assets"
	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	webrender "github.com/sojupie/tokens/catalog/web/internal/render"
	"github.com/sojupie/tokens/forge/components"
	"github.com/sojupie/tokens/forge/components/ui"
	"github.com/sojupie/tokens/source/publish"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func collectedModules(t *testing.T, component templ.Component) []string {
	t.Helper()
	assets := &webview.ComponentAssets{Modules: map[string]bool{}, Server: &Server{}}
	ctx := ui.WithAssets(context.WithValue(context.Background(), webview.ComponentAssetsKey{}, assets), ui.Assets(assets.Modules))
	var output bytes.Buffer
	if err := component.Render(ctx, &output); err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, name := range webview.ComponentModuleOrder() {
		if name != "common" && assets.Modules[name] {
			names = append(names, name)
		}
	}
	return names
}

func TestRenderedComponentsCollectNestedAssets(t *testing.T) {
	for _, tc := range []struct {
		name      string
		component templ.Component
		want      []string
	}{
		{"reorder sheet", webrender.SpecimenList(webview.ComponentCase{Family: "list", Control: "reorder", Width: "icon", ListTrailing: "text"}), []string{"sheet", "list"}},
		{"switch row", webrender.SpecimenList(webview.ComponentCase{Family: "list", ListTrailing: "switch", Width: "icon"}), []string{"switch", "list"}},
		{"checkbox row", webrender.SpecimenList(webview.ComponentCase{Family: "list", ListTrailing: "checkbox", Width: "icon"}), []string{"checkbox", "list"}},
		{"radio row", webrender.SpecimenList(webview.ComponentCase{Family: "list", ListTrailing: "radio", Width: "icon"}), []string{"radio", "list"}},
		{"adaptive search", webrender.SpecimenSearch(webview.ComponentCase{Family: "search", Variant: "adaptive", Shape: "icon"}), []string{"card", "fab", "navigation", "list", "search"}},
		{"fab menu opener", webrender.SpecimenFabMenu(webview.ComponentCase{Family: "fab-menu", State: "interactive"}), []string{"fab", "fab-menu"}},
		{"rail fab", webrender.NavigationExampleSurface(webview.ComponentCase{Family: "nav-rail", NavHeader: "menu-fab"}), []string{"fab", "navigation"}},
		{"search results", webrender.SearchScene(webview.ComponentCase{Family: "search", Shape: "icon"}), []string{"list", "search"}},
		{"loading control", webrender.SpecimenLoading(webview.ComponentCase{Family: "loading", State: "interactive", Control: "determinate"}), []string{"slider", "progress", "loading"}},
		{"toolbar fab", webrender.SpecimenToolbar(webview.ComponentCase{Family: "toolbar", Variant: "floating", Size: "small"}), []string{"fab", "toolbar"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := collectedModules(t, tc.component); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestUnknownComponentModuleFailsRendering(t *testing.T) {
	assets := &webview.ComponentAssets{Modules: map[string]bool{}}
	ctx := ui.WithAssets(context.WithValue(context.Background(), webview.ComponentAssetsKey{}, assets), ui.Assets(assets.Modules))
	if err := webview.UseComponentAssets("missing").Render(ctx, &bytes.Buffer{}); err == nil {
		t.Fatal("unknown module accepted")
	}
}

func TestBehaviorDeclarationsTravelWithIndependentScopes(t *testing.T) {
	parent := &webview.ComponentAssets{Modules: map[string]bool{}, Server: &Server{}}
	ctx := ui.WithAssets(context.WithValue(context.Background(), webview.ComponentAssetsKey{}, parent), ui.Assets(parent.Modules))
	child := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if err := webview.UseCatalogBehavior("component-demos.js").Render(ctx, w); err != nil {
			return err
		}
		return webrender.SpecimenList(webview.ComponentCase{Family: "list", Control: "reorder", ListTrailing: "switch"}).Render(ctx, w)
	})
	var output bytes.Buffer
	if err := webview.ComponentAssetScope().Render(templ.WithChildren(ctx, child), &output); err != nil {
		t.Fatal(err)
	}
	body := output.String()
	for _, script := range []string{"runtime/switches.js", "sheets.js", "lists.js", "component-demos.js"} {
		if !strings.Contains(body, `"`+script+`"`) {
			t.Errorf("scope omitted behavior %s", script)
		}
	}
	if !strings.Contains(body, `type="application/json" data-component-behaviors`) || strings.Contains(body, `"carousel.js"`) {
		t.Fatal("invalid behavior declaration", body)
	}
	if len(parent.Modules) > 0 || len(parent.CatalogScripts) > 0 {
		t.Fatal("child requirements leaked into the parent scope")
	}
	if err := webview.UseCatalogBehavior("unknown.js").Render(ctx, io.Discard); err == nil {
		t.Fatal("unknown Catalog behavior accepted")
	}
}

func TestRootBehaviorDeclarationsKeepDocumentDoctypeFirst(t *testing.T) {
	server := &Server{}
	page := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if err := ui.Require("tabs").Render(ctx, w); err != nil {
			return err
		}
		_, err := io.WriteString(w, `<!doctype html><html><head><title>Test</title></head><body>Content</body></html>`)
		return err
	})
	w := httptest.NewRecorder()
	if err := server.renderHTML(w, httptest.NewRequest("GET", "/", nil), page); err != nil {
		t.Fatal(err)
	}
	body := w.Body.String()
	declaration := strings.Index(body, `data-component-behaviors`)
	if !strings.HasPrefix(body, `<!doctype html>`) || declaration < 0 || declaration > strings.Index(body, `</head>`) {
		t.Fatal("root dependencies escaped the document head", body)
	}
}

func TestLazyAssetsKeepVersionedPublicDelivery(t *testing.T) {
	for _, name := range []string{"runtime/navigation.js", "bundles/common.css", "bundles/date.css", "asset-loader.js"} {
		data, err := fs.ReadFile(webassets.WebStaticFiles, name)
		if err != nil {
			t.Fatal(err)
		}
		request := httptest.NewRequest("GET", webassets.StaticURL(name), nil)
		request.Header.Set("Accept-Encoding", "gzip")
		response := httptest.NewRecorder()
		handleStaticAsset(response, request)
		if response.Code != 200 || response.Header().Get("Content-Encoding") != "gzip" {
			t.Fatal(name, response.Code, response.Header())
		}
		request = httptest.NewRequest("GET", webassets.StaticURL(name), nil)
		request.Header.Set("Accept-Encoding", "gzip")
		request.Header.Set("If-None-Match", response.Header().Get("ETag"))
		response = httptest.NewRecorder()
		handleStaticAsset(response, request)
		if response.Code != 304 || response.Body.Len() != 0 {
			t.Fatal(name, "did not revalidate")
		}
		if !strings.Contains(webassets.StaticURL(name), webassets.AssetVersion(data)) {
			t.Fatal("missing content version", name)
		}
	}
}

func TestEveryComponentAssetIsServed(t *testing.T) {
	var scripts map[string]string
	if err := json.Unmarshal([]byte(webview.ComponentScriptSources()), &scripts); err != nil {
		t.Fatal(err)
	}
	assets := make(map[string]string)
	for _, name := range webview.ComponentModuleOrder() {
		assets["bundles/"+name+".css"] = "text/css"
	}
	for name := range scripts {
		assets[name] = "text/javascript"
	}
	for name, contentType := range assets {
		t.Run(name, func(t *testing.T) {
			response := httptest.NewRecorder()
			handleStaticAsset(response, httptest.NewRequest("GET", webassets.StaticURL(name), nil))
			if response.Code != 200 || response.Body.Len() == 0 || !strings.HasPrefix(response.Header().Get("Content-Type"), contentType) {
				t.Fatalf("component asset unavailable: %d %v", response.Code, response.Header())
			}
		})
	}
}

func TestBehaviorSourceURLsCoverTheComponentManifest(t *testing.T) {
	data, err := components.Files.ReadFile("modules.json")
	if err != nil {
		t.Fatal(err)
	}
	var modules map[string]struct{ Scripts []string }
	if err := json.Unmarshal(data, &modules); err != nil {
		t.Fatal(err)
	}
	var sources map[string]string
	if err := json.Unmarshal([]byte(webview.ComponentScriptSources()), &sources); err != nil {
		t.Fatal(err)
	}
	for name, module := range modules {
		for _, script := range module.Scripts {
			if sources[script] != strings.TrimPrefix(webassets.StaticURL(script), "/static/") {
				t.Errorf("module %s script %s missing its versioned delivery URL; rebuild module assets", name, script)
			}
		}
	}
}

func TestRenderedAssetLinksKeepSharedThemeAndRejectMissingModules(t *testing.T) {
	server := &Server{themeBundles: map[string][]byte{"shared": []byte("shared"), "list": []byte("list"), "sheet": []byte("sheet")}}
	ctx := context.WithValue(context.Background(), webview.PageThemeKey{}, server.pageThemes())
	if !strings.HasPrefix(webview.ThemeURL(ctx), "/themes/shared.css?v=") {
		t.Fatal("missing shared head")
	}
	links, err := server.ComponentAssetLinks(&webview.ComponentAssets{Modules: map[string]bool{"sheet": true, "list": true}, DeferStyles: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"sheet", "list"} {
		if !strings.Contains(string(links), webassets.StaticURL("bundles/"+name+".css")) || !strings.Contains(string(links), "/themes/"+name+".css?v=") {
			t.Fatal("missing assets", name, string(links))
		}
	}
	if strings.Count(string(links), `media="print" onload="this.media='all'"`) != 4 || strings.Count(string(links), "<noscript>") != 4 {
		t.Fatal("component styles must load without blocking and retain a no-script fallback", string(links))
	}
	if strings.Contains(string(links), "all.css") {
		t.Fatal("loaded unrelated modules")
	}
	detailLinks, err := server.ComponentAssetLinks(&webview.ComponentAssets{Modules: map[string]bool{"list": true}})
	if err != nil || strings.Contains(string(detailLinks), `media="print"`) {
		t.Fatal("interactive detail styles must retain blocking delivery", err)
	}
	if _, err := server.ComponentAssetLinks(&webview.ComponentAssets{Modules: map[string]bool{"switch": true}}); err == nil {
		t.Fatal("missing theme accepted")
	}
}

func TestRenderHTMLCollectsAssetsBeforeWritingContent(t *testing.T) {
	server := &Server{}
	component := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if _, err := io.WriteString(w, "<main>"); err != nil {
			return err
		}
		if err := webrender.SpecimenList(webview.ComponentCase{Family: "list", Control: "reorder", Width: "icon", ListTrailing: "switch"}).Render(ctx, w); err != nil {
			return err
		}
		_, err := io.WriteString(w, "</main>")
		return err
	})
	response := httptest.NewRecorder()
	if err := server.renderHTML(response, httptest.NewRequest("GET", "/unrelated-route", nil), component); err != nil {
		t.Fatal(err)
	}
	body := response.Body.String()
	for _, name := range []string{"list", "sheet", "switch"} {
		link := strings.Index(body, webassets.StaticURL("bundles/"+name+".css"))
		if link < 0 || link > strings.Index(body, `data-list-host`) {
			t.Fatalf("module %s did not precede content", name)
		}
	}

}

func TestPlaygroundAssetsTravelInsideReplacementRegion(t *testing.T) {
	server := &Server{}
	page := webview.ComponentPreview{Cases: []webview.ComponentCase{{Family: "list", Control: "reorder", ListTrailing: "switch", Width: "icon"}}}
	response := httptest.NewRecorder()
	if err := server.renderHTML(response, httptest.NewRequest("GET", "/components/lists", nil), webrender.ComponentPreviewView(page)); err != nil {
		t.Fatal(err)
	}
	body := response.Body.String()
	start := strings.Index(body, `id="component-preview"`)
	for _, name := range []string{"switch", "sheet", "list"} {
		if at := strings.Index(body, webassets.StaticURL("bundles/"+name+".css")); at < start {
			t.Fatalf("%s link outside replacement region", name)
		}
	}
}

func TestFooterRegistersItsDownloadDialog(t *testing.T) {
	if got := collectedModules(t, webrender.SiteDownloadMenu()); !reflect.DeepEqual(got, []string{"radio", "checkbox", "messages", "loading", "dialog"}) {
		t.Fatalf("footer dependencies: %v", got)
	}
}

func TestContentPagesIncludeCardAssets(t *testing.T) {
	change := publish.ChangeView{Subject: "buttons", Name: "md.comp.button.color", Type: "changed", DetectedAt: "2026-10-23", SnapshotID: "current"}
	changes := []publish.ChangeView{change}
	state := publish.PublishedState{}
	for _, tc := range []struct {
		name string
		page templ.Component
	}{
		{"agents", webrender.ForAgentsPage("prompt", "/mcp")},
		{"overview", webrender.OverviewPage(publish.SnapshotView{}, nil, nil, "canvas")},
		{"changes", webrender.ChangesPage(state, "", "/changes", webview.ChangeFilters{}, "", changes, 1, 1, false, false)},
		{"component changes", webrender.ChangesPage(state, "buttons", "/components/buttons/changes", webview.ChangeFilters{}, "", changes, 1, 1, false, false)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := &Server{themeBundles: map[string][]byte{"shared": []byte("shared")}}
			for _, name := range webview.ComponentModuleOrder() {
				server.themeBundles[name] = []byte(name)
			}
			response := httptest.NewRecorder()
			if err := server.renderHTML(response, httptest.NewRequest("GET", "/", nil), tc.page); err != nil {
				t.Fatal(err)
			}
			body := response.Body.String()
			if tc.name == "agents" && strings.Contains(body, webassets.StaticURL("bundles/split.css")) {
				t.Fatal("loaded split-button styles on a page without split buttons")
			}
			headEnd := strings.Index(body, "</head>")
			card := strings.Index(body, `class="md-card`)
			for _, asset := range []string{webassets.StaticURL("bundles/card.css"), "/themes/card.css?v="} {
				at := strings.Index(body, asset)
				if at < 0 || at > headEnd || card < 0 {
					t.Fatalf("card asset %s missing from document head", asset)
				}
			}
		})
	}
}

func TestPreservedSiteNavigationStylesLiveInDocumentHead(t *testing.T) {
	server := New(publish.PublishedState{Subjects: map[publish.SubjectKey]publish.SubjectView{}}, "")
	server.themeBundles = map[string][]byte{"shared": []byte(":root{}")}
	for _, name := range webview.ComponentModuleOrder() {
		server.themeBundles[name] = []byte(":root{}")
	}
	for _, path := range []string{"/components", "/styles"} {
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest("GET", path, nil))
		if response.Code != 200 {
			t.Fatal(path, response.Code, response.Body.String())
		}
		body := response.Body.String()
		headEnd := strings.Index(body, "</head>")
		for _, asset := range []string{"/themes/navigation.css?v=", webassets.StaticURL("bundles/navigation.css")} {
			at := strings.Index(body, asset)
			if at < 0 || at > headEnd || strings.Count(body, asset) != 1 {
				t.Fatalf("navigation asset must occur once in persistent head: %s", asset)
			}
			end := strings.Index(body[at:], ">")
			if strings.Contains(body[at:at+end], `media="print"`) {
				t.Fatal("shell styles must block first paint")
			}
		}
	}

	links, err := server.ComponentAssetLinks(&webview.ComponentAssets{Modules: map[string]bool{"navigation": true}})
	if err != nil || !strings.Contains(string(links), "/themes/navigation.css") {
		t.Fatal("partial navigation lost its styles", err, string(links))
	}
}

func TestComponentStylesAreHoistedDeduplicatedAndOrdered(t *testing.T) {
	input := []byte(`<!doctype html><html><head><title>Test</title></head><body><script>const html="<head>";</script><main data-component-assets><link rel="stylesheet" href="/b.css" data-component-stylesheet data-component-order="0003" media="print" onload="this.media='all'"><link rel="stylesheet" href="/a.css" data-component-stylesheet data-component-order="0001"><link rel="stylesheet" href="/b.css" data-component-stylesheet data-component-order="0003"><p title="A &amp; B">Text</p></main></body></html>`)
	body, err := hoistComponentStyles(input)
	if err != nil {
		t.Fatal(err)
	}
	html := string(body)
	headEnd := strings.Index(html, "</head>")
	if strings.Count(html, `href="/b.css"`) != 1 || strings.Index(html, `href="/a.css"`) > strings.Index(html, `href="/b.css"`) || strings.Index(html, `href="/b.css"`) > headEnd {
		t.Fatal("styles not ordered once in head", html)
	}
	if strings.Contains(html, `media="print"`) || !strings.Contains(html, `<p title="A &amp; B">Text</p>`) || !strings.Contains(html, `<script>const html="<head>";</script>`) {
		t.Fatal("altered content or kept deferred stylesheet", html)
	}
	fragment := []byte(`<section><link rel="stylesheet" href="/a.css" data-component-stylesheet></section>`)
	got, err := hoistComponentStyles(fragment)
	if err != nil || !bytes.Equal(got, fragment) {
		t.Fatal("fragment lost dependency declaration")
	}
}

func TestStaticScopeKeepsStylesWithoutInteractiveScripts(t *testing.T) {
	parent := &webview.ComponentAssets{Modules: map[string]bool{}, Server: &Server{}}
	ctx := ui.WithAssets(context.WithValue(context.Background(), webview.ComponentAssetsKey{}, parent), ui.Assets(parent.Modules))
	child := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return ui.Require("carousel", "progress", "loading").Render(ctx, w)
	})
	var output bytes.Buffer
	if err := webview.ComponentAssetScope(true).Render(templ.WithChildren(ctx, child), &output); err != nil {
		t.Fatal(err)
	}
	body := output.String()
	for _, name := range []string{"carousel", "progress", "loading"} {
		if !strings.Contains(body, "bundles/"+name+".css") {
			t.Errorf("static scope lost %s styling", name)
		}
	}
	if !strings.Contains(body, `["progress.js"]`) || strings.Contains(body, `"carousel.js"`) || strings.Contains(body, `"loading.js"`) {
		t.Fatal("static scope activated an interactive controller", body)
	}
}

func TestBehaviorSourceOrderPreservesProviders(t *testing.T) {

	decoder := json.NewDecoder(strings.NewReader(webview.ComponentScriptSources()))
	if _, err := decoder.Token(); err != nil {
		t.Fatal(err)
	}
	var names []string
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			t.Fatal(err)
		}
		names = append(names, key.(string))
		var value string
		if err := decoder.Decode(&value); err != nil {
			t.Fatal(err)
		}
	}
	for _, pair := range [][2]string{{"artwork-canvas.js", "artwork.js"}, {"dialogs.js", "dialog-demo.js"}, {"progress.js", "loading.js"}} {
		if slices.Index(names, pair[0]) < 0 || slices.Index(names, pair[0]) >= slices.Index(names, pair[1]) {
			t.Fatalf("provider order %v: %v", pair, names)
		}
	}
}

func TestGeneratedAssetsAreBoundToPublication(t *testing.T) {
	directory := t.TempDir()
	assets := filepath.Join(directory, "assets")
	if err := os.MkdirAll(filepath.Join(assets, "dtcg", "light-default"), 0755); err != nil {
		t.Fatal(err)
	}
	css := []byte(":root { --md-sys-color-primary: red; }\n")
	database := []byte("publication database")
	digest := func(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
	components, _ := webassets.ReadStaticFile("static/components.css")
	buttons, _ := webassets.ReadStaticFile("static/buttons.css")
	manifest := map[string]any{"publicationId": "test", "snapshotId": "capture", "databaseSha256": digest(database), "cssSha256": digest(css), "buttonsSha256": digest(buttons), "componentsSha256": digest(components)}
	for file, data := range map[string][]byte{"tokens.sqlite": database, "assets/theme.css": css, "assets/dtcg/light-default/tokens.tokens.json": []byte(`{"size":{"$type":"dimension","$value":{"value":4,"unit":"px"}}}`)} {
		if err := os.WriteFile(filepath.Join(directory, file), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(assets, "themes"), 0755); err != nil {
		t.Fatal(err)
	}
	themeCSS := []byte(":root{--shared:red}")
	checksums := map[string]string{}
	for _, name := range []string{"shared", "all", "navigation"} {
		if err := os.WriteFile(filepath.Join(assets, "themes", name+".css"), themeCSS, 0644); err != nil {
			t.Fatal(err)
		}
		checksums[name] = digest(themeCSS)
	}
	manifest["themeBundles"] = checksums
	data, _ := json.Marshal(manifest)
	if err := os.WriteFile(filepath.Join(assets, "manifest.json"), data, 0644); err != nil {
		t.Fatal(err)
	}
	state := publish.PublishedState{PublicationID: "test", CurrentSnapshot: publish.SnapshotView{ID: "capture"}}
	server := New(state, "")
	if err := server.UseGeneratedAssets(directory); err != nil {
		t.Fatal(err)
	}
	responseBundle := httptest.NewRecorder()
	server.Handler().ServeHTTP(responseBundle, httptest.NewRequest("GET", "/themes/navigation.css?v="+digest(themeCSS), nil))
	if responseBundle.Code != 200 || responseBundle.Body.String() != string(themeCSS) || !strings.Contains(responseBundle.Header().Get("Cache-Control"), "immutable") {
		t.Fatal("theme bundle delivery", responseBundle)
	}
	if err := os.WriteFile(filepath.Join(assets, "themes", "navigation.css"), []byte("tampered"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := server.UseGeneratedAssets(directory); err == nil {
		t.Fatal("accepted corrupted theme bundle")
	}
	if err := os.WriteFile(filepath.Join(assets, "themes", "navigation.css"), themeCSS, 0644); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(server.agentGuide(), "/downloads/dtcg/light-default/tokens.tokens.json") {
		t.Fatal("generated exports missing from agent guide")
	}
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/theme.css", nil))
	if response.Code != 200 || response.Body.String() != string(css) {
		t.Fatal(response)
	}
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/downloads/dtcg/light-default/tokens.tokens.json", nil))
	if response.Code != 200 || response.Header().Get("Content-Type") != "application/design-tokens+json" || response.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal(response)
	}
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/downloads/dtcg/unknown/tokens.tokens.json", nil))
	if response.Code != 404 {
		t.Fatal(response)
	}

	external := filepath.Join(t.TempDir(), "assets")
	if err := os.Rename(assets, external); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "current-assets")
	if err := os.Symlink(external, link); err != nil {
		t.Fatal(err)
	}
	if err := server.UseGeneratedAssetsDirectory(directory, link); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/downloads/dtcg/light-default/tokens.tokens.json", nil))
	if response.Code != 200 {
		t.Fatal("external asset path was not pinned", response)
	}
	if err := server.UseGeneratedAssetsDirectory(directory, link); err == nil {
		t.Fatal("accepted missing required assets")
	}
	if err := server.UseGeneratedAssetsDirectory(directory, t.TempDir()); err == nil {
		t.Fatal("accepted directory without manifest")
	}
	if err := os.Symlink(external, assets); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "tokens.sqlite"), []byte("new publication"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := server.UseGeneratedAssets(directory); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("accepted stale assets: %v", err)
	}
}

func TestPageEncodingNegotiation(t *testing.T) {
	body := strings.Repeat("Material component preview and token values. ", 100)
	for _, test := range []struct {
		accept     string
		compressed bool
	}{{"", false}, {"gzip, deflate, br", true}, {"gzip;q=0", false}, {"*;q=1, gzip;q=0", false}, {"gzip;q=0.5", true}} {
		r := httptest.NewRequest("GET", "/components", nil)
		r.Header.Set("Accept-Encoding", test.accept)
		w := httptest.NewRecorder()
		w.Header().Set("Vary", "HX-Request, Accept")
		if _, err := writePageBody(w, r, []byte(body)); err != nil {
			t.Fatal(err)
		}
		result := w.Body.String()
		if test.compressed {
			reader, err := gzip.NewReader(w.Body)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := io.ReadAll(reader)
			if err != nil {
				t.Fatal(err)
			}
			reader.Close()
			result = string(decoded)
			if w.Header().Get("Content-Encoding") != "gzip" {
				t.Fatal("missing encoding")
			}
		} else if w.Header().Get("Content-Encoding") != "" {
			t.Fatal("compressed an excluded representation")
		}
		if result != body {
			t.Fatal("encoding changed content")
		}
		if strings.Join(w.Header().Values("Vary"), ",") != "HX-Request, Accept,Accept-Encoding" {
			t.Fatal("lost representation negotiation")
		}
	}
}
func TestEncodedAssetsRetainConditionalAndRangeRequests(t *testing.T) {
	body := []byte(strings.Repeat(".md-button { color: purple; }", 100))
	get := func(method, etag, byteRange string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "/static/test.css", nil)
		r.Header.Set("Accept-Encoding", "gzip")
		if etag != "" {
			r.Header.Set("If-None-Match", etag)
		}
		if byteRange != "" {
			r.Header.Set("Range", byteRange)
		}
		w := httptest.NewRecorder()
		serveCachedAsset(w, r, "test.css", body)
		return w
	}
	full := get("GET", "", "")
	if full.Code != 200 || full.Header().Get("Content-Encoding") != "gzip" || full.Body.Len() >= len(body) {
		t.Fatal("asset was not compressed")
	}
	cached := get("GET", full.Header().Get("ETag"), "")
	if cached.Code != 304 || cached.Body.Len() != 0 {
		t.Fatal("conditional request changed")
	}
	head := get("HEAD", "", "")
	if head.Code != 200 || head.Body.Len() != 0 || head.Header().Get("Content-Length") != full.Header().Get("Content-Length") {
		t.Fatal("HEAD differs from GET representation")
	}
	partial := get("GET", "", "bytes=0-3")
	if partial.Code != 206 || partial.Header().Get("Content-Encoding") != "" || partial.Body.String() != string(body[:4]) {
		t.Fatal("range request changed")
	}
}

func TestResponseCacheValidationAndInvalidation(t *testing.T) {
	version, calls := "first", 0
	source := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path == "/missing" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintf(w, "%s:%s:%s", version, r.URL.RawQuery, r.Header.Get("HX-Target"))
	})
	handler := cacheResponses(source)
	fetch := func(method, path, etag, target string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, nil)
		r.Header.Set("If-None-Match", etag)
		r.Header.Set("HX-Target", target)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	first := fetch("HEAD", "/components", "", "")
	if first.Code != 200 || first.Body.Len() != 0 || calls != 1 {
		t.Fatal("HEAD must seed the GET representation")
	}
	if !strings.Contains(first.Header().Get("Server-Timing"), `desc="miss"`) || !strings.Contains(first.Header().Get("Server-Timing"), "render;dur=") {
		t.Fatal("missing render timing")
	}
	next := fetch("GET", "/components", "", "")
	if !strings.Contains(next.Header().Get("Server-Timing"), `desc="hit"`) || strings.Contains(next.Header().Get("Server-Timing"), "render;") {
		t.Fatal("cache hit reused stale render timing")
	}
	if next.Body.String() != "first::" || calls != 1 {
		t.Fatal("cache did not reuse rendered response")
	}
	unchanged := fetch("GET", "/components", "W/"+next.Header().Get("ETag"), "")
	if unchanged.Code != 304 || unchanged.Body.Len() != 0 || calls != 1 {
		t.Fatal("unchanged response must validate without rendering")
	}
	if fetch("GET", "/components?q=other", next.Header().Get("ETag"), "").Code != 200 || fetch("GET", "/components", next.Header().Get("ETag"), "component-preview").Code != 200 {
		t.Fatal("query and fragment variants must remain separate")
	}
	for range 2 {
		fetch("GET", "/missing", "", "")
	}
	if calls != 5 {
		t.Fatalf("errors were cached: calls=%d", calls)
	}
	version = "second"
	handler = cacheResponses(source)
	changed := fetch("GET", "/components", next.Header().Get("ETag"), "")
	if changed.Code != 200 || changed.Header().Get("ETag") == next.Header().Get("ETag") {
		t.Fatal("new publication/handler did not invalidate response")
	}
}

func TestResponseCacheIsBounded(t *testing.T) {
	cache := &responseCache{}
	for i := range 300 {
		cache.put(&cachedResponse{key: fmt.Sprint(i), header: make(http.Header), body: []byte("value")})
	}
	if len(cache.entries) != 256 || cache.get("0") != nil || cache.get("299") == nil {
		t.Fatal("cache must evict old filter combinations")
	}
	cache.put(&cachedResponse{key: "oversized", body: make([]byte, responseCacheBytes+1)})
	if cache.get("oversized") != nil {
		t.Fatal("oversized response exceeded cache budget")
	}
}

func TestBoostedNavigationSharesDocumentCache(t *testing.T) {
	handler := New(seoTestState(), "").Handler()
	get := func(boosted, partial bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest("GET", "/components", nil)
		if boosted || partial {
			r.Header.Set("HX-Request", "true")
		}
		if boosted {
			r.Header.Set("HX-Boosted", "true")
			r.Header.Set("HX-Target", "body")
		}
		if partial {
			r.Header.Set("HX-Target", "subject-list")
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	full, boosted, fragment := get(false, false), get(true, false), get(false, true)
	if full.Header().Get("ETag") != boosted.Header().Get("ETag") || full.Body.String() != boosted.Body.String() {
		t.Fatal("boosted navigation must reuse the complete document")
	}
	if full.Header().Get("ETag") == fragment.Header().Get("ETag") {
		t.Fatal("index filtering must retain its separate fragment")
	}
	if full.Header().Get("Cache-Control") != "public, no-cache" {
		t.Fatal("documents must validate before reuse")
	}
}
