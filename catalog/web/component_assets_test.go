package web

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"

	"github.com/a-h/templ"
	"github.com/sojupie/tokens/forge/components/ui"
	"github.com/sojupie/tokens/source/publish"
	"io"
	"io/fs"
	"net/http/httptest"
	"strings"
	"testing"
)

func collectedModules(t *testing.T, component templ.Component) []string {
	t.Helper()
	assets := &componentAssets{modules: map[string]bool{}, server: &Server{}}
	ctx := ui.WithAssets(context.WithValue(context.Background(), componentAssetsKey{}, assets), ui.Assets(assets.modules))
	var output bytes.Buffer
	if err := component.Render(ctx, &output); err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, name := range componentModuleOrder() {
		if name != "common" && assets.modules[name] {
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
		{"reorder sheet", specimenList(componentCase{Family: "list", Control: "reorder", Width: "icon", ListTrailing: "text"}), []string{"sheet", "list"}},
		{"switch row", specimenList(componentCase{Family: "list", ListTrailing: "switch", Width: "icon"}), []string{"switch", "list"}},
		{"checkbox row", specimenList(componentCase{Family: "list", ListTrailing: "checkbox", Width: "icon"}), []string{"checkbox", "list"}},
		{"radio row", specimenList(componentCase{Family: "list", ListTrailing: "radio", Width: "icon"}), []string{"radio", "list"}},
		{"adaptive search", specimenSearch(componentCase{Family: "search", Variant: "adaptive", Shape: "icon"}), []string{"card", "fab", "navigation", "list", "search"}},
		{"fab menu opener", specimenFabMenu(componentCase{Family: "fab-menu", State: "interactive"}), []string{"fab", "fab-menu"}},
		{"rail fab", navigationExampleSurface(componentCase{Family: "nav-rail", NavHeader: "menu-fab"}), []string{"fab", "navigation"}},
		{"search results", searchScene(componentCase{Family: "search", Shape: "icon"}), []string{"list", "search"}},
		{"loading control", specimenLoading(componentCase{Family: "loading", State: "interactive", Control: "determinate"}), []string{"slider", "progress", "loading"}},
		{"toolbar fab", specimenToolbar(componentCase{Family: "toolbar", Variant: "floating", Size: "small"}), []string{"fab", "toolbar"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := collectedModules(t, tc.component); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestUnknownComponentModuleFailsRendering(t *testing.T) {
	assets := &componentAssets{modules: map[string]bool{}}
	ctx := ui.WithAssets(context.WithValue(context.Background(), componentAssetsKey{}, assets), ui.Assets(assets.modules))
	if err := useComponentAssets("missing").Render(ctx, &bytes.Buffer{}); err == nil {
		t.Fatal("unknown module accepted")
	}
}

func TestLazyAssetsKeepVersionedPublicDelivery(t *testing.T) {
	for _, name := range []string{"runtime/navigation.js", "bundles/common.css", "bundles/date.css", "asset-loader.js"} {
		data, err := fs.ReadFile(webStaticFiles, name)
		if err != nil {
			t.Fatal(err)
		}
		request := httptest.NewRequest("GET", staticURL(name), nil)
		request.Header.Set("Accept-Encoding", "gzip")
		response := httptest.NewRecorder()
		handleStaticAsset(response, request)
		if response.Code != 200 || response.Header().Get("Content-Encoding") != "gzip" {
			t.Fatal(name, response.Code, response.Header())
		}
		request = httptest.NewRequest("GET", staticURL(name), nil)
		request.Header.Set("Accept-Encoding", "gzip")
		request.Header.Set("If-None-Match", response.Header().Get("ETag"))
		response = httptest.NewRecorder()
		handleStaticAsset(response, request)
		if response.Code != 304 || response.Body.Len() != 0 {
			t.Fatal(name, "did not revalidate")
		}
		if !strings.Contains(staticURL(name), assetVersion(data)) {
			t.Fatal("missing content version", name)
		}
	}
}

func TestEveryComponentAssetIsServed(t *testing.T) {
	var scripts map[string]string
	if err := json.Unmarshal([]byte(componentScriptSources()), &scripts); err != nil {
		t.Fatal(err)
	}
	assets := make(map[string]string)
	for _, name := range componentModuleOrder() {
		assets["bundles/"+name+".css"] = "text/css"
	}
	for name := range scripts {
		assets[name] = "text/javascript"
	}
	for name, contentType := range assets {
		t.Run(name, func(t *testing.T) {
			response := httptest.NewRecorder()
			handleStaticAsset(response, httptest.NewRequest("GET", staticURL(name), nil))
			if response.Code != 200 || response.Body.Len() == 0 || !strings.HasPrefix(response.Header().Get("Content-Type"), contentType) {
				t.Fatalf("component asset unavailable: %d %v", response.Code, response.Header())
			}
		})
	}
}

func TestRenderedAssetLinksKeepSharedThemeAndRejectMissingModules(t *testing.T) {
	server := &Server{themeBundles: map[string][]byte{"shared": []byte("shared"), "list": []byte("list"), "sheet": []byte("sheet")}}
	ctx := context.WithValue(context.Background(), pageThemeKey{}, server.pageThemes())
	if !strings.HasPrefix(themeURL(ctx), "/themes/shared.css?v=") {
		t.Fatal("missing shared head")
	}
	links, err := server.componentAssetLinks(&componentAssets{modules: map[string]bool{"sheet": true, "list": true}, deferStyles: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"sheet", "list"} {
		if !strings.Contains(string(links), staticURL("bundles/"+name+".css")) || !strings.Contains(string(links), "/themes/"+name+".css?v=") {
			t.Fatal("missing assets", name, string(links))
		}
	}
	if strings.Count(string(links), `media="print" onload="this.media='all'"`) != 4 || strings.Count(string(links), "<noscript>") != 4 {
		t.Fatal("component styles must load without blocking and retain a no-script fallback", string(links))
	}
	if strings.Contains(string(links), "all.css") {
		t.Fatal("loaded unrelated modules")
	}
	detailLinks, err := server.componentAssetLinks(&componentAssets{modules: map[string]bool{"list": true}})
	if err != nil || strings.Contains(string(detailLinks), `media="print"`) {
		t.Fatal("interactive detail styles must retain blocking delivery", err)
	}
	if _, err := server.componentAssetLinks(&componentAssets{modules: map[string]bool{"switch": true}}); err == nil {
		t.Fatal("missing theme accepted")
	}
}

func TestRenderHTMLCollectsAssetsBeforeWritingContent(t *testing.T) {
	server := &Server{}
	component := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if _, err := io.WriteString(w, "<main>"); err != nil {
			return err
		}
		if err := specimenList(componentCase{Family: "list", Control: "reorder", Width: "icon", ListTrailing: "switch"}).Render(ctx, w); err != nil {
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
		link := strings.Index(body, staticURL("bundles/"+name+".css"))
		if link < 0 || link > strings.Index(body, `data-list-host`) {
			t.Fatalf("module %s did not precede content", name)
		}
	}

}

func TestPlaygroundAssetsTravelInsideReplacementRegion(t *testing.T) {
	server := &Server{}
	page := componentPreview{Cases: []componentCase{{Family: "list", Control: "reorder", ListTrailing: "switch", Width: "icon"}}}
	response := httptest.NewRecorder()
	if err := server.renderHTML(response, httptest.NewRequest("GET", "/components/lists", nil), componentPreviewView(page)); err != nil {
		t.Fatal(err)
	}
	body := response.Body.String()
	start := strings.Index(body, `id="component-preview"`)
	for _, name := range []string{"switch", "sheet", "list"} {
		if at := strings.Index(body, staticURL("bundles/"+name+".css")); at < start {
			t.Fatalf("%s link outside replacement region", name)
		}
	}
}

func TestFooterRegistersItsDownloadDialog(t *testing.T) {
	if got := collectedModules(t, siteDownloadMenu()); !reflect.DeepEqual(got, []string{"radio", "checkbox", "messages", "loading", "dialog"}) {
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
		{"agents", forAgentsPage("prompt", "/mcp")},
		{"overview", overviewPage(publish.SnapshotView{}, nil, nil, "canvas")},
		{"changes", changesPage(state, "", "/changes", changeFilters{}, "", changes, 1, 1, false, false)},
		{"component changes", changesPage(state, "buttons", "/components/buttons/changes", changeFilters{}, "", changes, 1, 1, false, false)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := &Server{themeBundles: map[string][]byte{"shared": []byte("shared")}}
			for _, name := range componentModuleOrder() {
				server.themeBundles[name] = []byte(name)
			}
			response := httptest.NewRecorder()
			if err := server.renderHTML(response, httptest.NewRequest("GET", "/", nil), tc.page); err != nil {
				t.Fatal(err)
			}
			body := response.Body.String()
			if tc.name == "agents" && strings.Contains(body, staticURL("bundles/split.css")) {
				t.Fatal("loaded split-button styles on a page without split buttons")
			}
			headEnd := strings.Index(body, "</head>")
			card := strings.Index(body, `class="md-card`)
			for _, asset := range []string{staticURL("bundles/card.css"), "/themes/card.css?v="} {
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
	for _, name := range componentModuleOrder() {
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
		for _, asset := range []string{"/themes/navigation.css?v=", staticURL("bundles/navigation.css")} {
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
	// Independently rendered regions must still declare their own dependencies.
	links, err := server.componentAssetLinks(&componentAssets{modules: map[string]bool{"navigation": true}})
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
