package web

import (
	"context"
	"encoding/json"
	"github.com/sojupie/tokens/source/publish"
	"net/http/httptest"
	"net/url"
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
	if got := exportChoiceID("context.Language height"); got != "export-choices-context-language-height" {
		t.Fatal(got)
	}
	state := exportDialogState{Page: exportPage{Contexts: []exportContext{{Name: "Platform"}, {Name: "Theme"}, {Name: "Contrast"}, {Name: "Audience"}}}}
	attrs := state.requestAttrs("context.Theme")
	targets := attrs["hx-disable"].(string)
	if strings.Contains(targets, "context-theme") || strings.Contains(targets, "context-platform") || !strings.Contains(targets, "context-contrast") || !strings.Contains(targets, "context-audience") {
		t.Fatal(targets)
	}
	if state.requestAttrs("format")["hx-get"] != "/downloads/export/summary" {
		t.Fatal("format changes must not wait for option resolution")
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
	// This extra malformed token makes whole-collection conversion fail, but cannot
	// affect the selected subject's choices or combination count.
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
