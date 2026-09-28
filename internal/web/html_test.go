package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sojupie/m3-source/internal/publish"
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
		{path: "/", body: []string{"Published Material tokens", "/components/snackbar", "/styles/elevation"}, status: http.StatusOK},
		{path: "/components/snackbar", body: []string{"Snackbar", "md.comp.snackbar.container.width", "560", "/downloads/component/snackbar.json"}, status: http.StatusOK},
		{path: "/styles/elevation", body: []string{"Elevation", "Subject kind: <code class=\"sl-code\">style"}, status: http.StatusOK},
		{path: "/styles/motion/overview", body: []string{"Motion", "Subject kind: <code class=\"sl-code\">style"}, status: http.StatusOK},
		{path: "/components?q=snack", body: []string{"snack", "hx-target=\"#subject-list\""}, status: http.StatusOK},
		{path: "/changes", body: []string{"Change log", "md.comp.snackbar.container.width", "480", "560"}, status: http.StatusOK},
		{path: "/components/snackbar/changes", body: []string{"snackbar changes", "Compared with snapshot: <code>snapshot-0"}, status: http.StatusOK},
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
