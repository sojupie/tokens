// Integration checks for page.
package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sojupie/tokens/catalog/web/internal/playground"
	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	webrender "github.com/sojupie/tokens/catalog/web/internal/render"
	"github.com/sojupie/tokens/source/publish"
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestTokenBrowserFragmentsPreservePlaygroundQuery(t *testing.T) {
	server := scopedExportFixture()
	for _, target := range []string{"tokens-token-browser", "div#tokens-token-browser", "tokens-context-fields", "div#tokens-context-fields"} {
		r := httptest.NewRequest("GET", "/components/buttons?preview.state=hovered&q=icon", nil)
		r.Header.Set("HX-Request", "true")
		r.Header.Set("HX-Target", target)
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, r)
		body := w.Body.String()
		if w.Code != 200 || strings.Contains(body, "<!doctype") || strings.Contains(body, `id="component-preview"`) || strings.Contains(body, `id="tokens-site-header"`) {
			t.Fatalf("%s: %d %s", target, w.Code, body)
		}
		if strings.Contains(target, "browser") {
			for _, want := range []string{`id="tokens-token-browser"`, `hx-sync="this:replace"`, `hx-get="/components/buttons"`, `name="preview.state" value="hovered"`} {
				if !strings.Contains(body, want) {
					t.Fatalf("missing %s", want)
				}
			}
		} else if !strings.Contains(body, `name="context.Theme"`) {
			t.Fatal("missing context choices")
		}
	}
}

func TestNavigationUsesInertPartialToKeepDocumentCacheShared(t *testing.T) {
	server := scopedExportFixture()
	for _, partial := range []bool{false, true} {
		r := httptest.NewRequest("GET", "/components", nil)
		if partial {
			r.Header.Set("HX-Request", "true")
			r.Header.Set("HX-Target", "body")
			r.Header.Set("HX-Boosted", "true")
		}
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, r)
		body := w.Body.String()
		if w.Code != 200 {
			t.Fatal(w.Code, body)
		}
		if !strings.Contains(body, `<template hx type="partial" hx-target="#tokens-site-destinations"`) {
			t.Fatal("wrong partial", body)
		}
		for _, want := range []string{`id="tokens-site-destinations"`, `hx-head.min.js`, `hx-sync="this:replace"`, `hx-target="#tokens-site-destinations" hx-swap="innerMorph"`, `hx-target="#tokens-site-rail" hx-swap="innerMorph"`} {
			if !strings.Contains(body, want) {
				t.Fatal(want)
			}
		}
	}
}

func TestExportDialogReturnsResolvedHTMLAndPreservesMultipleSelections(t *testing.T) {
	server := scopedExportFixture()
	query := url.Values{"kind": {"component"}, "slug": {"buttons"}, "scope": {"subject"}, "content": {"tokens"}, "format": {"dtcg", "source"}, "context.Theme": {"Light", "Dark"}}
	state := server.exportDialog(context.Background(), query)
	if !state.Ready || state.Error != "" || state.Label != "Download ZIP" || !strings.Contains(state.Counts, "4 format/context combinations") {
		t.Fatalf("%+v", state)
	}
	if len(state.Query["format"]) != 2 || len(state.Query["context.Theme"]) != 2 {
		t.Fatal("multiselection lost")
	}
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/downloads/export/options?"+query.Encode(), nil))
	body := w.Body.String()
	if w.Code != 200 || !strings.Contains(body, `class="tokens-export-fields"`) || strings.Contains(body, `hx-trigger="load"`) || !strings.Contains(body, `4 format/context combinations`) || strings.Contains(body, `<dialog`) || strings.Contains(body, `<!doctype`) {
		t.Fatal(w.Code, body)
	}
	if strings.Count(body, `checked`) < 4 {
		t.Fatal("selected inputs missing", body)
	}
	options := server.exportDialogOptions(query)
	if !options.Resolved || options.Ready || options.Counts != "" {
		t.Fatalf("options must resolve before calculating exports: %+v", options)
	}
	w = httptest.NewRecorder()
	server.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/downloads/export/summary?"+query.Encode(), nil))
	body = w.Body.String()
	if !strings.Contains(body, "4 format/context combinations") || !strings.Contains(body, `hx-target="#tokens-export-confirmation"`) || strings.Contains(body, `hx-trigger="load"`) {
		t.Fatal("summary must confirm its own action without reloading", body)
	}
	query.Del("format")
	state = server.exportDialog(context.Background(), query)
	if state.Ready || !strings.Contains(state.Counts, "Select at least one") {
		t.Fatal("empty formats must stay empty", state)
	}
	query.Set("format", "dtcg")
	query.Del("context.Theme")
	state = server.exportDialog(context.Background(), query)
	if state.Ready {
		t.Fatal("empty context must stay empty")
	}
	query.Set("context.Theme", "invalid")
	state = server.exportDialog(context.Background(), query)
	if state.Ready || state.Error == "" {
		t.Fatal("unknown context accepted")
	}
}

func TestExportRequestDependenciesFollowContextOrder(t *testing.T) {
	if got := webview.ExportChoiceID("context.Language height"); got != "export-choices-context-language-height" {
		t.Fatal(got)
	}
	state := webview.ExportDialogState{Page: webview.ExportPage{Contexts: []webview.ExportContext{{Name: "Platform"}, {Name: "Theme"}, {Name: "Contrast"}, {Name: "Audience"}}}}
	attrs := state.RequestAttrs("context.Theme")
	targets := attrs["hx-disable"].(string)
	if strings.Contains(targets, "context-theme") || strings.Contains(targets, "context-platform") || !strings.Contains(targets, "context-contrast") || !strings.Contains(targets, "context-audience") {
		t.Fatal(targets)
	}
	if state.RequestAttrs("format")["hx-get"] != "/downloads/export/summary" {
		t.Fatal("format changes must not wait for option resolution")
	}
	for _, name := range []string{"content", "scope", "format", "context.Theme"} {
		if state.RequestAttrs(name)["hx-indicator"] != "#tokens-export-loading-action" {
			t.Errorf("%s does not use the Download loading indicator", name)
		}
	}
}

func TestExportFooterKeepsSummaryAndLoadingActionOutsideScrollingForm(t *testing.T) {
	server := scopedExportFixture()
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/components/buttons", nil))
	body := w.Body.String()
	form := strings.Index(body, `id="tokens-export-form"`)
	if form < 0 {
		t.Fatal("missing export form")
	}
	end := form + strings.Index(body[form:], "</form>")
	summary := strings.Index(body, `id="tokens-export-summary"`)
	if summary < end {
		t.Fatal("export summary must be docked outside the scrolling form")
	}
	if !strings.Contains(body, `id="tokens-export-loading-action" hx-preserve`) || !strings.Contains(body, `class="md-button md-filled-button md-loading-button"`) {
		t.Fatal("loading primitive must remain mounted")
	}
}

// Selection feedback must remain independent of serializers and asset assembly.
func TestExportSelectionFeedbackDoesNotGenerateFiles(t *testing.T) {
	server := scopedExportFixture()
	query := url.Values{"kind": {"component"}, "slug": {"buttons"}, "scope": {"subject"}, "content": {"tokens"}, "format": {"dtcg", "source"}, "context.Theme": {"Light", "Dark"}}

	key := publish.SubjectKey{Kind: publish.SubjectStyle, Slug: "broken"}
	server.state.Subjects[key] = publish.SubjectView{Key: key, Tokens: []publish.TokenView{{SourceName: "broken", Values: []publish.TokenValueView{{Value: json.RawMessage(`{"numeric":1}`), ContextLabels: []string{"invalid label"}}}}}}
	state := server.exportDialog(context.Background(), query)
	if !state.Ready || state.Error != "" || !strings.Contains(state.Counts, "4 format/context combinations") {
		t.Fatalf("%+v", state)
	}
	if _, _, _, _, err := server.exportDocument(url.Values{"scope": {"collection"}}); err == nil {
		t.Fatal("fixture must fail conversion")
	}
	server.themeBundles = map[string][]byte{"shared": []byte(":root{}")}
	server.tokenAssetsDir = t.TempDir()
	query.Set("content", "implementation")
	query.Set("format", "zip")
	state = server.exportDialog(context.Background(), query)
	if !state.Ready || state.Error != "" {
		t.Fatalf("summary attempted asset assembly: %+v", state)
	}
}

func TestLandingSamplesDoNotLimitCatalogIndexes(t *testing.T) {
	state := publish.PublishedState{Subjects: map[publish.SubjectKey]publish.SubjectView{}}
	for _, kind := range []publish.SubjectKind{publish.SubjectComponent, publish.SubjectStyle} {
		for i := range 8 {
			key := publish.SubjectKey{Kind: kind, Slug: fmt.Sprintf("example-%d", i)}
			state.Subjects[key] = publish.SubjectView{Key: key, Name: key.Slug}
		}
	}
	handler := New(state, "").Handler()
	for _, check := range []struct {
		path  string
		cards int
	}{{"/", 8}, {"/components", 8}, {"/styles", 8}} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", check.path, nil))
		if response.Code != 200 {
			t.Fatalf("%s returned %d", check.path, response.Code)
		}
		if count := strings.Count(response.Body.String(), `class="tokens-component-tile"`); count != check.cards {
			t.Fatalf("%s has %d cards, want %d", check.path, count, check.cards)
		}
		if check.path == "/" && (!strings.Contains(response.Body.String(), "View all components") || !strings.Contains(response.Body.String(), "View all styles")) {
			t.Fatal("samples must link to the complete indexes")
		}
	}
}

func TestLandingCanvasIsDefaultAndKeepsHTMLFallback(t *testing.T) {
	handler := New(publish.PublishedState{}, "").Handler()
	for _, mode := range []string{"", "dom", "canvas", "unknown"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", "/?artwork="+mode, nil))
		if response.Code != 200 {
			t.Fatal(mode, response.Code)
		}
		if strings.Contains(response.Body.String(), `<canvas`) != (mode != "dom") {
			t.Fatal("wrong renderer", mode)
		}
		if strings.Count(response.Body.String(), `class="tokens-art-tile `) != 9 {
			t.Fatal("missing HTML artwork fallback", mode)
		}
	}
}

func TestComponentNameEnglishInflections(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"Buttons", "Button"}, {"Icon buttons", "Icon Button"}, {"Checkboxes", "Checkbox"}, {"Checkbox", "Checkbox"},
		{"Switches", "Switch"}, {"Switch", "Switch"}, {"Bottom sheets", "Bottom Sheet"}, {"Menus", "Menu"},
		{"Badges", "Badge"}, {"Carousels", "Carousel"}, {"Progress indicators", "Progress Indicator"},
		{"Status", "Status"}, {"Statuses", "Status"}, {"Search", "Search"}, {"Tabs", "Tab"},
		{"Navigation drawers", "Navigation Drawer"}, {"Floating action buttons (FABs)", "Floating Action Button (FAB)"},
		{"Extended FAB", "Extended FAB"}, {"FAB menus", "FAB Menu"}, {"iOS switches", "iOS Switch"},
		{"Properties", "Property"}, {"People", "Person"}, {"Series", "Series"}, {"Glass", "Glass"}, {"", ""},
	} {
		t.Run(tc.input, func(t *testing.T) {
			got := webview.ComponentName(tc.input)
			if got != tc.want {
				t.Fatalf("%q, want %q", got, tc.want)
			}
			if webview.ComponentName(got) != got {
				t.Fatalf("name not stable: %q", got)
			}
			subject := publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectComponent}, Name: tc.input}
			if !regexp.MustCompile(`^Material Design 3 "` + regexp.QuoteMeta(tc.want) + `" tokens$`).MatchString(subjectTokenTitle(subject)) {
				t.Fatal(subjectTokenTitle(subject))
			}
		})
	}

	style := publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectStyle, Slug: "states"}, Name: "States"}
	if webview.SubjectName(style) != "States" || webview.SubjectURL(style) != "/styles/states" {
		t.Fatal("style or URL renamed")
	}
}

func TestTableNamesKeepSourceDistinctions(t *testing.T) {
	for _, tc := range []struct{ subject, label, want string }{
		{"Icon buttons", "Icon button - Color - Filled", "Filled / Color"},
		{"Icon buttons", "Icon button - Size - Medium", "Medium / Size"},
		{"Date pickers", "Date picker - Modal", "Modal"},
		{"Date pickers", "Date picker - Modal input", "Modal Input"},
		{"Date pickers", "Date picker - Docked", "Docked"},
		{"Menus", "Menus - Common", "Common"},
		{"Menus", "Menus - Color - Standard", "Standard / Color"},
		{"Menus", "Menus - Color - Vibrant", "Vibrant / Color"},
		{"Menus", "Menu (baseline)", "Baseline"},
		{"Buttons", "[Deprecated] Button – Outlined", "Outlined / Deprecated"},
		{"Toolbars", "Bottom app bar (baseline)", "Bottom App Bar (baseline)"},
		{"Buttons", "Unfamiliar source label", "Unfamiliar Source Label"},
	} {
		t.Run(tc.label, func(t *testing.T) {
			subject := publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectComponent}, Name: tc.subject}
			set := webview.TokenSet{Label: tc.label, SourceName: "unchanged.source.id"}
			if got := webview.TokenTableQualifier(subject, set); got != tc.want {
				t.Fatalf("%q, want %q", got, tc.want)
			}
			if set.Label != tc.label || set.SourceName != "unchanged.source.id" {
				t.Fatal("source metadata mutated")
			}
			title := subjectTableTitle(subject, set)
			if !regexp.MustCompile(`^Material Design 3 "[^"]+" tokens: ` + regexp.QuoteMeta(tc.want) + `$`).MatchString(title) {
				t.Fatal(title)
			}
			if !strings.Contains(subjectTableDescription(subject, set), tc.want) {
				t.Fatal("description lost variant")
			}
		})
	}
}

func TestSourceTableSEOMatrix(t *testing.T) {
	key := publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "date-pickers"}
	subject := publish.SubjectView{Key: key, Name: "Date pickers"}
	for i, variant := range []string{"Modal", "Docked"} {
		subject.Tokens = append(subject.Tokens, publish.TokenView{SourceName: "md.comp.date-picker." + strings.ToLower(variant) + ".container.color", Category: "COLOR", Presentation: &publish.TokenPresentation{SetName: "md.comp.date-picker." + strings.ToLower(variant), SetLabel: "Date picker - " + variant, SetOrder: i}})
	}
	server := New(publish.PublishedState{Subjects: map[publish.SubjectKey]publish.SubjectView{key: subject}}, "")
	if err := server.SetPublicOrigin("https://tokens.example/md3"); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		path, qualifier, canonical string
		noindex                    bool
	}{
		{"/components/date-pickers", "", "/components/date-pickers", false},
		{"/components/date-pickers/sets/md.comp.date-picker.modal", "", "/components/date-pickers", false},
		{"/components/date-pickers?set=md.comp.date-picker.docked", ": Docked", "/components/date-pickers/sets/md.comp.date-picker.docked", false},
		{"/components/date-pickers?preview.compare=interactive", "", "", true},
		{"/components/date-pickers?context.Theme=Dark", "", "", true},
		{"/components/date-pickers?q=container", "", "", true},
		{"/components/date-pickers?utm_source=example", "", "/components/date-pickers", false},
		{"/components/date-pickers/sets/md.comp.date-picker.modal?q=container", "", "", true},
		{"/components/date-pickers/changes", "changes", "/components/date-pickers/changes", false},
	} {
		t.Run(tc.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/md3"+tc.path, nil))
			body := html.UnescapeString(response.Body.String())
			if response.Code != 200 {
				t.Fatalf("status %d", response.Code)
			}
			name := `Material Design 3 "Date Picker" tokens` + tc.qualifier
			if tc.qualifier == "changes" {
				name = `Material Design 3 "Date Picker" token changes`
			}
			for _, pattern := range []string{`<title>` + regexp.QuoteMeta(name) + ` \| Token Catalog</title>`, `<h1[^>]*>` + regexp.QuoteMeta(name) + `</h1>`} {
				if !regexp.MustCompile(pattern).MatchString(body) {
					t.Fatalf("missing %s", pattern)
				}
			}
			if tc.canonical != "" && !strings.Contains(body, `rel="canonical" href="https://tokens.example/md3`+tc.canonical+`"`) {
				t.Fatal("wrong canonical")
			}
			if tc.canonical == "" && strings.Contains(body, `rel="canonical"`) {
				t.Fatal("filtered view claims canonical")
			}
			if strings.Contains(body, `content="noindex, follow"`) != tc.noindex || (response.Header().Get("X-Robots-Tag") != "") != tc.noindex {
				t.Fatal("inconsistent indexing")
			}
			if tc.qualifier == ": Modal" || tc.qualifier == ": Docked" {
				if !regexp.MustCompile(`<meta name="description" content="Material Design 3 "Date Picker" tokens for ` + tc.qualifier[2:] + `\.`).MatchString(body) {
					t.Fatal("description lost variant")
				}
			}
		})
	}

	subject.Tokens = subject.Tokens[:1]
	server.state.Subjects[key] = subject
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/md3/components/date-pickers/sets/md.comp.date-picker.modal", nil))
	if !strings.Contains(response.Body.String(), `rel="canonical" href="https://tokens.example/md3/components/date-pickers"`) {
		t.Fatal("single table not consolidated")
	}
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/md3/sitemap.xml", nil))
	if strings.Contains(response.Body.String(), "/sets/") {
		t.Fatal("noncanonical table in sitemap")
	}
}

func TestPublishedTableNamesAndAggregateCoverage(t *testing.T) {
	path := os.Getenv("TOKEN_CATALOG_THEME_DB")
	if path == "" {
		t.Skip("set TOKEN_CATALOG_THEME_DB to inspect captured table names")
	}
	state, err := publish.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	total := 0
	for _, subject := range state.Subjects {
		all := buildTokenBrowser(subject, url.Values{"set": {"all"}})
		if all.Count != len(subject.Tokens) {
			t.Fatalf("%s: aggregate has %d of %d tokens", subject.Key.Slug, all.Count, len(subject.Tokens))
		}
		names := map[string]string{}
		for _, set := range all.Sets {
			title := subjectTableTitle(subject, set)
			if other, exists := names[title]; exists {
				t.Errorf("%s: %s and %s share %q", subject.Key.Slug, other, set.Key, title)
			}
			names[title] = set.Key
			if !regexp.MustCompile(`^Material Design 3 ".+" tokens: .+\S$`).MatchString(title) {
				t.Errorf("invalid title %q", title)
			}
			if strings.Contains(title, "Expressive") && !strings.Contains(strings.ToLower(set.Label), "expressive") {
				t.Errorf("inferred Expressive status: %q", title)
			}
			total++
		}
	}
	t.Logf("checked %d subjects and %d table titles", len(state.Subjects), total)
}

func TestDocumentPrefetchCacheAndVariants(t *testing.T) {
	server := New(seoTestState(), "")
	for _, check := range []struct {
		purpose, secPurpose, hx, accept, cache string
	}{
		{cache: "public, no-cache"},
		{purpose: "prefetch", cache: "public, no-cache"},
		{secPurpose: "prefetch", cache: "public, no-cache"},
		{secPurpose: "prefetch;prerender", cache: "public, no-cache"},
		{purpose: "prefetch", hx: "true", cache: "public, no-cache"},
		{purpose: "prefetch", accept: "text/markdown", cache: "public, no-cache"},
	} {
		r := httptest.NewRequest("GET", "/components", nil)
		r.Header.Set("Purpose", check.purpose)
		r.Header.Set("Sec-Purpose", check.secPurpose)
		r.Header.Set("HX-Request", check.hx)
		r.Header.Set("Accept", check.accept)
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, r)
		if w.Code != 200 || w.Header().Get("Cache-Control") != check.cache || w.Header().Get("Vary") != "HX-Request, HX-Target, Accept" {
			t.Fatalf("%+v: status=%d headers=%v", check, w.Code, w.Header())
		}
		if check.hx == "" && check.accept == "" && (!strings.Contains(w.Body.String(), `href="https://github.com/sojupie/tokens"`) || !strings.Contains(w.Body.String(), "prefetch.js?v=")) {
			t.Fatal("full pages must include the GitHub link and prefetch asset")
		}
	}
}

func TestSourceTablesRetainMembershipGroupOrderAndTokenOrder(t *testing.T) {
	token := func(name, group string, groupOrder, order int) publish.TokenView {
		return publish.TokenView{SourceName: name, Category: "ELEVATION", Lifecycle: "active", Presentation: &publish.TokenPresentation{SetName: "md.comp.button.elevated", SetLabel: "Button - Color - Elevated", SetOrder: 2, Groups: []publish.DisplayGroup{{Name: group, Order: groupOrder}}, Order: order}}
	}
	subject := publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "buttons"}, Name: "Buttons", Tokens: []publish.TokenView{
		token("md.comp.button.elevated.z", "Enabled", 1, 1),
		token("md.comp.button.elevated.a", "Enabled", 1, 2),
		token("md.comp.button.elevated.disabled.elevation", "Disabled", 2, 0),
	}}
	browser := buildTokenBrowser(subject, url.Values{"set": {"md.comp.button.elevated"}})
	if len(browser.Sets) != 1 || browser.Selected.Label != "Button - Color - Elevated" {
		t.Fatalf("sets=%+v", browser.Sets)
	}
	if len(browser.Selected.Groups) != 2 || browser.Selected.Groups[0].Name != "Enabled" || browser.Selected.Groups[1].Name != "Disabled" {
		t.Fatalf("groups=%+v", browser.Selected.Groups)
	}
	rows := browser.Selected.Groups[0].Tokens
	if rows[0].SourceName != "md.comp.button.elevated.z" {
		t.Fatalf("source row order lost: %+v", rows)
	}
	if got := webview.TokenSetURL(subject, browser.Selected); got != "/components/buttons/sets/md.comp.button.elevated" {
		t.Fatal(got)
	}
	if browser.Selected.Label != "Button - Color - Elevated" {
		t.Fatal("menu invented a label")
	}
	canonical, ok := legacyTokenSet(subject, "button/color/elevated")
	if !ok || canonical.SourceName != "md.comp.button.elevated" {
		t.Fatal("legacy table URL lost its source-table replacement")
	}
	subject.Tokens = subject.Tokens[:1]
	subject.Tokens[0].Presentation.Groups = nil
	browser = buildTokenBrowser(subject, url.Values{"set": {"md.comp.button.elevated"}})
	var flat strings.Builder
	if err := webview.WriteTokenGroup(&flat, browser.Selected.Groups[0], true); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(flat.String(), "<details") {
		t.Fatal("a source table without display groups must remain flat")
	}
}

func TestShapeDescriptionAndPreviewPreserveCornerOverrides(t *testing.T) {
	for _, test := range []struct{ raw, text, radius string }{
		{`{"shape":{"family":"SHAPE_FAMILY_ROUNDED_CORNERS","defaultSize":{"value":16,"unit":"DIPS"},"topLeft":{"unit":"DIPS"}}}`, "Default: 16dp · Top-left: 0dp", "0px 16px 16px 16px"},
		{`{"shape":{"family":"SHAPE_FAMILY_ROUNDED_CORNERS","defaultSize":{"unit":"DIPS"},"topRight":{"value":16,"unit":"DIPS"},"bottomRight":{"value":16,"unit":"DIPS"}}}`, "Default: 0dp · Top-right: 16dp · Bottom-right: 16dp", "0px 16px 16px 0px"},
		{`{"shape":{"family":"SHAPE_FAMILY_CIRCULAR"}}`, "Full", "50%"},
		{`{"shape":{"family":"SHAPE_FAMILY_CUT_CORNERS","defaultSize":{"value":16,"unit":"DIPS"}}}`, "Cut Corners · 16dp", ""},
	} {
		if got := webview.TokenValueText(json.RawMessage(test.raw)); got != test.text {
			t.Errorf("text=%q want %q", got, test.text)
		}
		if got := shapePreviewCSS(json.RawMessage(test.raw)); got != test.radius {
			t.Errorf("radius=%q want %q", got, test.radius)
		}
	}
	var output strings.Builder
	if err := webview.WriteTokenReferences(&output, webview.TokenRow{Name: "Large end", Value: "Default: 0dp · Top-right: 16dp · Bottom-right: 16dp", Shape: "0px 16px 16px 0px", HasValue: true}); err != nil {
		t.Fatal(err)
	}
	if strings.Count(output.String(), `border-radius:0px 16px 16px 0px`) != 1 {
		t.Fatal("the table must provide one directional shape for reuse in the popup")
	}
	markdown, err := htmlMarkdown([]byte(`<main><h2>Shape</h2><table><tr><th>Large end</th><td>`+output.String()+`</td></tr></table></main>`), "https://example.test/styles/shape", "")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(markdown, "Default: 0dp") != 1 || !strings.Contains(markdown, "## Shape") {
		t.Fatalf("duplicated popup or missing table label: %s", markdown)
	}
}

func TestSourceDimensionZeroRequiresAConcreteDimension(t *testing.T) {
	for _, test := range []struct {
		raw, want string
		valid     bool
	}{
		{`{"unit":"DIPS"}`, "0dp", true},
		{`{"unit":"POINTS"}`, "0pt", true},
		{`{}`, "", false},
		{`{"unit":"DIPS","value":null}`, "", false},
		{`{"unit":"DIPS","value":"unknown"}`, "", false},
	} {
		var value map[string]any
		if err := json.Unmarshal([]byte(test.raw), &value); err != nil {
			t.Fatal(err)
		}
		got, valid := webview.DimensionValueText(value)
		if valid != test.valid || got != test.want {
			t.Errorf("%s: got %q/%v", test.raw, got, valid)
		}
	}
}

func TestStatusExamplesDescribeContent(t *testing.T) {
	for variant, want := range map[string]string{"labeled": "99+ unread notifications", "dot": "New notification"} {
		var out bytes.Buffer
		if err := webrender.SpecimenBadge(webview.ComponentCase{Family: "badge", Variant: variant}).Render(context.Background(), &out); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), `aria-label="`+want+`"`) {
			t.Errorf("%s missing %s", variant, want)
		}
	}
	c := webview.ComponentCase{Family: "progress", Control: "determinate"}
	if c.ProgressAttributes()["aria-label"] != "Loading preview content" {
		t.Error("Progress name must identify the affected content")
	}
}

func TestSnackbarDocumentsItsFocusShortcut(t *testing.T) {
	c := webview.ComponentCase{Family: "snackbar", State: "interactive"}
	var out bytes.Buffer
	if err := webrender.SpecimenSnackbar(c).Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`aria-describedby="` + c.ID() + `-shortcut"`, `id="` + c.ID() + `-shortcut"`, `Press F6 to focus the snackbar action.`} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %s", want)
		}
	}
}

func TestBadgePreviewExercisesAnchorsAndOverflowCounts(t *testing.T) {
	page := &webview.ComponentPreview{Cases: []webview.ComponentCase{{Family: "badge", Variant: "labeled"}}}
	if err := playground.ConfigureBadgePreview(page, map[string][]string{"preview.content": {"1000"}, "preview.placement": {"trailing"}, "preview.direction": {"rtl"}}); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := webrender.SpecimenBadge(page.Cases[0]).Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`dir="rtl"`, `data-badge-placement="trailing"`, `1000 unread notifications`, `999+`, `aria-hidden="true"`} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %s: %s", want, out.String())
		}
	}
	if err := playground.ConfigureBadgePreview(page, map[string][]string{"preview.content": {"arbitrary"}}); err == nil {
		t.Fatal("invalid preview content accepted")
	}
}

func TestTokenSetLinksKeepOnlySupportedContexts(t *testing.T) {
	subject := publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "buttons"}}
	set := webview.TokenSet{Component: "button", Category: "COLOR", Variant: "elevated", SourceTokens: []publish.TokenView{
		{Values: []publish.TokenValueView{{ContextLabels: []string{"Theme: Dark", "Platform: Web"}}, {ContextLabels: []string{"Theme: Light", "Platform: Android"}}}},
	}}
	browser := webview.TokenBrowser{Contexts: url.Values{"context.Theme": {"Dark"}, "context.Platform": {"Android"}, "context.Unknown": {"unsupported"}}}
	link, err := url.Parse(webview.TokenSetLinkURL(subject, set, browser))
	if err != nil {
		t.Fatal(err)
	}
	if link.Query().Get("context.Theme") != "Dark" || link.Query().Has("context.Platform") || link.Query().Has("context.Unknown") {
		t.Fatalf("incompatible context carried into link: %s", link)
	}
	if link.Path != "/components/buttons/sets/button/color/elevated" {
		t.Fatal(link)
	}
	browser.Contexts = nil
	if strings.Contains(webview.TokenSetLinkURL(subject, set, browser), "?") {
		t.Fatal("default selection should link to the clean indexable URL")
	}
}

func TestSelectorDistinguishesSourceFamiliesAndVariants(t *testing.T) {
	variant := webview.TokenSet{Label: "Icon button - Color - Outlined", Component: "icon-button", Variant: "outlined", Category: "COLOR"}
	family := webview.TokenSet{Component: "outlined-icon-button", Category: "COLOR"}
	groups := webview.TokenSetMenuGroups(publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectComponent}}, []webview.TokenSet{variant, family})
	if len(groups) != 2 || groups[0].Label != "Icon Button" || groups[1].Label != "Outlined Icon Button" || webview.TokenTableQualifier(publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectComponent}, Name: "Icon buttons"}, variant) != "Outlined / Color" {
		t.Fatalf("selector merged different source families: %+v", groups)
	}
}

func TestStyleTokenSetsHaveProperLabels(t *testing.T) {
	subject := publish.SubjectView{
		Key:  publish.SubjectKey{Kind: publish.SubjectStyle, Slug: "shape/corner-radius"},
		Name: "Shape",
		Tokens: []publish.TokenView{
			{SourceName: "md.sys.shape.corner.large", Category: "SHAPE"},
		},
	}
	sets := webview.TokenSets(subject)
	if len(sets) != 1 {
		t.Fatalf("len(sets) = %d, want 1", len(sets))
	}
	if sets[0].Label != "Shape" {
		t.Fatalf("sets[0].Label = %q, want %q", sets[0].Label, "Shape")
	}
}

func TestColorTableDefaultsToSchemeAndPreservesExplicitSelection(t *testing.T) {
	subject := publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectStyle, Slug: "color"}, Name: "Color"}
	for i, set := range []struct{ name, label string }{{"md.ref.palette", "Palettes"}, {"md.sys.color", "Color Schemes"}} {
		subject.Tokens = append(subject.Tokens, publish.TokenView{SourceName: set.name + ".primary", Presentation: &publish.TokenPresentation{SetName: set.name, SetLabel: set.label, SetOrder: i}})
	}
	for _, test := range []struct{ selection, want string }{{"", "md.sys.color"}, {"md.ref.palette", "md.ref.palette"}, {"all", "all"}} {
		browser := buildTokenBrowser(subject, url.Values{"set": {test.selection}})
		if browser.Selected.Key != test.want {
			t.Fatalf("selection %q: got %q, want %q", test.selection, browser.Selected.Key, test.want)
		}
	}

	server := New(publish.PublishedState{Subjects: map[publish.SubjectKey]publish.SubjectView{subject.Key: subject}}, "")
	handler := server.Handler()
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/styles/color", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `href="/styles/color/sets/md.ref.palette"`) {
		t.Fatal("Color Schemes must link to the explicit Palettes page")
	}

	request := httptest.NewRequest(http.MethodGet, "/styles/color/sets/md.ref.palette", nil)
	request.Header.Set("HX-Request", "true")
	request.Header.Set("HX-Target", "tokens-token-browser")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	body := response.Body.String()
	if response.Code != http.StatusOK || !strings.Contains(body, "md.ref.palette.primary") || strings.Contains(body, "md.sys.color.primary") || !strings.Contains(body, `href="/styles/color" aria-current="false"`) {
		t.Fatal("Palettes must render its own tokens and link back to Color Schemes")
	}

	request = httptest.NewRequest(http.MethodGet, "/styles/color", nil)
	request.Header.Set("Accept", "text/markdown")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	body = response.Body.String()
	if response.Code != http.StatusOK || !strings.Contains(body, "**Showing Color Schemes**") || !strings.Contains(body, "[Palettes](/styles/color/sets/md.ref.palette)") {
		t.Fatalf("Markdown lost the selected token set or palette link: %s", body)
	}

	if err := server.SetPublicOrigin("https://tokens.example/md3"); err != nil {
		t.Fatal(err)
	}
	handler = server.Handler()
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil))
	body = response.Body.String()
	if response.Code != http.StatusOK || !strings.Contains(body, "<loc>https://tokens.example/md3/styles/color/sets/md.ref.palette</loc>") || strings.Contains(body, "/sets/md.sys.color") {
		t.Fatalf("sitemap lost the Palettes page or duplicated Color Schemes: %s", body)
	}
}
