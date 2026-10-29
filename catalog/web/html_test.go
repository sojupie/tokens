package web

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sojupie/tokens/source/publish"
)

func TestHTMLPagesRenderComponentsAndStyles(t *testing.T) {
	state := publish.PublishedState{
		CurrentSnapshot: publish.SnapshotView{ID: "snapshot-1", CreatedAt: "2026-09-07T00:00:00Z"},
		Changes: []publish.ChangeView{{
			ID:                 "change-1",
			Type:               "changed",
			Subject:            "snackbar",
			Kind:               "contextualReferenceTree",
			Name:               "md.comp.snackbar.container.width",
			Before:             `{"dimension":{"value":480,"unit":"DIPS"}}`,
			After:              `{"dimension":{"value":560,"unit":"DIPS"}}`,
			PreviousSnapshotID: "snapshot-0",
			SnapshotID:         "snapshot-1",
		}},
		Subjects: map[publish.SubjectKey]publish.SubjectView{
			{Kind: publish.SubjectComponent, Slug: "snackbar"}: {
				Key:  publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "snackbar"},
				Name: "Snackbar",
				Tokens: []publish.TokenView{{
					SourceName: "md.comp.snackbar.container.width",
					Values:     []publish.TokenValueView{{Value: []byte(`{"dimension":{"value":560,"unit":"DIPS"}}`)}},
				}},
			},
			{Kind: publish.SubjectStyle, Slug: "elevation"}: {
				Key:  publish.SubjectKey{Kind: publish.SubjectStyle, Slug: "elevation"},
				Name: "Elevation",
			},
			{Kind: publish.SubjectStyle, Slug: "motion/overview"}: {
				Key:  publish.SubjectKey{Kind: publish.SubjectStyle, Slug: "motion/overview"},
				Name: "Motion",
			},
		},
	}
	handler := New(state, "").Handler()

	checks := []struct {
		path   string
		body   []string
		status int
	}{
		{path: "/", body: []string{"Material Design 3 tokens and change history", "Browse Material Design 3 components and styles", "/components/snackbar", "/styles/elevation"}, status: http.StatusOK},
		{path: "/components/snackbar", body: []string{"Snackbar", "Container", "Snackbar container width", "md.comp.snackbar.container.width", "560", "Download tokens"}, status: http.StatusOK},
		{path: "/styles/elevation", body: []string{"Elevation", "Download tokens"}, status: http.StatusOK},
		{path: "/styles/motion/overview", body: []string{"Motion", "Download tokens"}, status: http.StatusOK},
		{path: "/components?q=snack", body: []string{"snack", "hx-target=\"#subject-list\""}, status: http.StatusOK},
		{path: "/changes", body: []string{"Change log", "1 item changed", "/components/snackbar", "View changes"}, status: http.StatusOK},
		{path: "/components/snackbar/changes", body: []string{"Material Design 3 &#34;Snackbar&#34; token changes", "md.comp.snackbar.container.width", "480", "560"}, status: http.StatusOK},
	}
	for _, check := range checks {
		t.Run(check.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, check.path, nil))
			if recorder.Code != check.status {
				t.Fatalf("status = %d, want %d", recorder.Code, check.status)
			}
			if recorder.Header().Get("Content-Type") != "text/html; charset=utf-8" {
				t.Fatalf("content type = %q", recorder.Header().Get("Content-Type"))
			}
			for _, value := range check.body {
				if !strings.Contains(recorder.Body.String(), value) {
					t.Fatalf("body does not contain %q:\n%s", value, recorder.Body.String())
				}
			}
			body := recorder.Body.String()
			initial := strings.Index(body, "localStorage.getItem('tokens-theme')")
			styles := strings.Index(body, `rel="stylesheet"`)
			if initial < 0 || styles < initial || !strings.Contains(body, `<link rel="stylesheet" href="`+staticURL("site.css")+`"`) || !strings.Contains(body, `<script src="`+staticURL("theme.js")+`" defer></script>`) {
				t.Fatal("saved theme must be inline before styles; controls must be deferred")
			}
			if !strings.Contains(body, `class="tokens-brand-logo"`) || !strings.Contains(body, `fill="var(--md-sys-color-tertiary,var(--tokens-logo-fallback))"`) {
				t.Fatal("header must use the inline tertiary SVG logo")
			}
			for _, value := range []string{`<footer class="tokens-site-footer"`, `href="https://marmenlind.com"`, "Made by Olle Marmenlind", "Independent project. Not affiliated with Google."} {
				if !strings.Contains(recorder.Body.String(), value) {
					t.Fatalf("page footer does not contain %q", value)
				}
			}
		})
	}
}

func TestHTMLSubjectListSupportsHTMXFiltering(t *testing.T) {
	state := publish.PublishedState{Subjects: map[publish.SubjectKey]publish.SubjectView{
		{Kind: publish.SubjectComponent, Slug: "snackbar"}: {
			Key:  publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "snackbar"},
			Name: "Snackbar",
		},
		{Kind: publish.SubjectComponent, Slug: "button"}: {
			Key:  publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "button"},
			Name: "Button",
		},
	}}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/components?q=snack", nil)
	request.Header.Set("HX-Request", "true")
	New(state, "").Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if !strings.Contains(recorder.Body.String(), "Snackbar") {
		t.Fatalf("filtered response does not contain Snackbar: %s", recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), "Button") {
		t.Fatalf("filtered response contains Button: %s", recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), "<!doctype html>") {
		t.Fatalf("HTMX response unexpectedly rendered the full page: %s", recorder.Body.String())
	}
}

func TestTokenValuesHaveReadableForms(t *testing.T) {
	context := publish.TokenValueView{
		ContextTags:      []string{"designSystems/abc/contextTagGroups/def/tags/ghi"},
		ContextLabels:    []string{"Theme: Dark"},
		IsDefaultContext: true,
	}
	if got := tokenContextLabel(context); got != "Default · Dark" {
		t.Fatalf("context label = %q", got)
	}
	if got := tokenValueText([]byte(`{"color":{"red":1,"green":0.5,"blue":0,"alpha":1}}`)); got != "rgba(255, 127.5, 0, 1)" {
		t.Fatalf("color value = %q", got)
	}
	if got := tokenValueText([]byte(`{"dimension":{"value":480,"unit":"DIPS"}}`)); got != "480dp" {
		t.Fatalf("dimension value = %q", got)
	}
}

func TestShapeTokenFormatting(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "circular full",
			input: `{"shape":{"family":"SHAPE_FAMILY_CIRCULAR"}}`,
			want:  "Full",
		},
		{
			name:  "uniform 16dp",
			input: `{"shape":{"family":"SHAPE_FAMILY_ROUNDED_CORNERS","defaultSize":{"value":16,"unit":"DIPS"}}}`,
			want:  "16dp",
		},
		{
			name:  "uniform 0dp none",
			input: `{"shape":{"family":"SHAPE_FAMILY_ROUNDED_CORNERS","defaultSize":{"unit":"DIPS"}}}`,
			want:  "0dp",
		},
		{
			name:  "large end rounding",
			input: `{"shape":{"family":"SHAPE_FAMILY_ROUNDED_CORNERS","topRight":{"value":16,"unit":"DIPS"},"bottomRight":{"value":16,"unit":"DIPS"},"defaultSize":{"unit":"DIPS"}}}`,
			want:  "Default: 0dp · Top-right: 16dp · Bottom-right: 16dp",
		},
		{
			name:  "large start rounding",
			input: `{"shape":{"family":"SHAPE_FAMILY_ROUNDED_CORNERS","topLeft":{"value":16,"unit":"DIPS"},"bottomLeft":{"value":16,"unit":"DIPS"},"defaultSize":{"unit":"DIPS"}}}`,
			want:  "Default: 0dp · Top-left: 16dp · Bottom-left: 16dp",
		},
		{
			name:  "large top rounding",
			input: `{"shape":{"family":"SHAPE_FAMILY_ROUNDED_CORNERS","topLeft":{"value":16,"unit":"DIPS"},"topRight":{"value":16,"unit":"DIPS"},"defaultSize":{"unit":"DIPS"}}}`,
			want:  "Default: 0dp · Top-left: 16dp · Top-right: 16dp",
		},
		{
			name:  "all four corners specified identically",
			input: `{"shape":{"family":"SHAPE_FAMILY_ROUNDED_CORNERS","topLeft":{"value":12,"unit":"DIPS"},"topRight":{"value":12,"unit":"DIPS"},"bottomLeft":{"value":12,"unit":"DIPS"},"bottomRight":{"value":12,"unit":"DIPS"}}}`,
			want:  "12dp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tokenValueText([]byte(tt.input))
			if got != tt.want {
				t.Fatalf("tokenValueText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHTMLChangeLogIsPaginated(t *testing.T) {
	changes := make([]publish.ChangeView, 51)
	for index := range changes {
		changes[index] = publish.ChangeView{
			ID:                 fmt.Sprintf("change-%02d", index),
			Type:               "changed",
			Subject:            "carousel",
			Name:               fmt.Sprintf("token-%02d", index),
			Before:             `{"dimension":{"value":1,"unit":"DIPS"}}`,
			After:              `{"dimension":{"value":2,"unit":"DIPS"}}`,
			PreviousSnapshotID: "snapshot-0",
		}
	}
	handler := New(publish.PublishedState{
		CurrentSnapshot: publish.SnapshotView{ID: "snapshot-1"},
		Changes:         changes,
	}, "").Handler()

	firstPage := httptest.NewRecorder()
	handler.ServeHTTP(firstPage, httptest.NewRequest(http.MethodGet, "/changes", nil))
	if !strings.Contains(firstPage.Body.String(), "51 items changed") || strings.Contains(firstPage.Body.String(), "token-49") {
		t.Fatalf("unexpected first change page:\n%s", firstPage.Body.String())
	}
	if strings.Contains(firstPage.Body.String(), `href="/changes?page=2"`) {
		t.Fatal("record count split a single component across outer pages")
	}
	firstPage = httptest.NewRecorder()
	handler.ServeHTTP(firstPage, httptest.NewRequest(http.MethodGet, "/components/carousel/changes", nil))
	if !strings.Contains(firstPage.Body.String(), "token-49") || strings.Contains(firstPage.Body.String(), "token-50") {
		t.Fatal("record page is not bounded")
	}

	secondPage := httptest.NewRecorder()
	handler.ServeHTTP(secondPage, httptest.NewRequest(http.MethodGet, "/components/carousel/changes?page=2", nil))
	if !strings.Contains(secondPage.Body.String(), "token-50") || !strings.Contains(secondPage.Body.String(), `href="/components/carousel/changes"`) {
		t.Fatalf("unexpected second change page:\n%s", secondPage.Body.String())
	}
}
