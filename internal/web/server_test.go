package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/sojupie/m3-source/internal/publish"
)

func TestHandlerServesReadAPI(t *testing.T) {
	state := publish.PublishedState{
		PublicationID: "publication-1",
		CurrentSnapshot: publish.SnapshotView{
			ID:            "snapshot-1",
			CreatedAt:     "2026-09-07T00:00:00Z",
			ManifestHash:  "manifest-hash",
			CaptureStatus: "complete",
			ReviewStatus:  "provisional",
		},
		Snapshots: []publish.SnapshotView{{ID: "snapshot-1"}},
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
			{Kind: publish.SubjectStyle, Slug: "elevation"}: {
				Key:  publish.SubjectKey{Kind: publish.SubjectStyle, Slug: "elevation"},
				Name: "Elevation",
				Tokens: []publish.TokenView{{
					SourceName: "md.sys.elevation.level1",
				}},
			},
			{Kind: publish.SubjectComponent, Slug: "snackbar"}: {
				Key:  publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "snackbar"},
				Name: "Snackbar",
			},
		},
	}
	server := New(state, "")

	checks := []struct {
		path       string
		statusCode int
	}{
		{path: "/api/v1/health", statusCode: http.StatusOK},
		{path: "/api/v1/releases", statusCode: http.StatusOK},
		{path: "/api/v1/changes", statusCode: http.StatusOK},
		{path: "/api/v1/components", statusCode: http.StatusOK},
		{path: "/api/v1/components/snackbar", statusCode: http.StatusOK},
		{path: "/api/v1/components/snackbar/changes", statusCode: http.StatusOK},
		{path: "/api/v1/components/missing", statusCode: http.StatusNotFound},
	}
	for _, check := range checks {
		t.Run(check.path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, check.path, nil)
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != check.statusCode {
				t.Fatalf("status = %d, want %d", recorder.Code, check.statusCode)
			}
			if check.statusCode == http.StatusOK && recorder.Header().Get("Content-Type") != "application/json; charset=utf-8" {
				t.Fatalf("content type = %q", recorder.Header().Get("Content-Type"))
			}
		})
	}
}

func TestComponentsAreSortedAndIncludeStyles(t *testing.T) {
	state := publish.PublishedState{
		PublicationID:   "publication-1",
		CurrentSnapshot: publish.SnapshotView{ID: "snapshot-1"},
		Subjects: map[publish.SubjectKey]publish.SubjectView{
			{Kind: publish.SubjectStyle, Slug: "spacing"}: {
				Key: publish.SubjectKey{Kind: publish.SubjectStyle, Slug: "spacing"},
			},
			{Kind: publish.SubjectComponent, Slug: "button"}: {
				Key: publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "button"},
			},
		},
	}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/components", nil)
	recorder := httptest.NewRecorder()
	New(state, "").Handler().ServeHTTP(recorder, request)

	var response componentListResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if len(response.Subjects) != 2 {
		t.Fatalf("subject count = %d, want 2", len(response.Subjects))
	}
	if response.Subjects[0].Kind != publish.SubjectComponent || response.Subjects[1].Kind != publish.SubjectStyle {
		t.Fatalf("subjects = %#v", response.Subjects)
	}
}

func TestHandlerServesOnlyApprovedDownloads(t *testing.T) {
	downloadsDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(downloadsDir, "tokens.json"), []byte("published"), 0o644); err != nil {
		t.Fatal(err)
	}
	handler := New(publish.PublishedState{}, downloadsDir).Handler()

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/downloads/tokens.json", nil))
	if recorder.Code != http.StatusOK || recorder.Body.String() != "published" {
		t.Fatalf("download response = %d %q", recorder.Code, recorder.Body.String())
	}

	for _, relativePath := range []string{"../tokens.json", "component\\tokens.json", ""} {
		if safeDownloadPath(relativePath) {
			t.Fatalf("safeDownloadPath(%q) = true", relativePath)
		}
	}
}
