package web

import (
	"github.com/sojupie/tokens/source/publish"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPreviewFamiliesShareOneDecodedCoverageFile(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "components-coverage.json")
	if err := os.WriteFile(file, []byte(`{"configurations":[{"family":"switch","variant":"icon","state":"enabled"},{"family":"badge","variant":"labeled","state":"enabled"}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	server := New(publish.PublishedState{}, "")
	server.tokenAssetsDir = directory
	for i, slug := range []string{"switch", "badges", "switch"} {
		if i == 1 {
			if err := os.Remove(file); err != nil {
				t.Fatal(err)
			}
		}
		page, err := server.buildComponentPreview(publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectComponent, Slug: slug}}, url.Values{})
		if err != nil || page == nil || len(page.Cases) == 0 {
			t.Fatalf("%s reread coverage or lost its configurations: %v", slug, err)
		}
	}
}

func TestShowcaseUpdatesReturnOnlyCacheablePreviewHTML(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "components-coverage.json"), []byte(`{"configurations":[{"family":"switch","variant":"icon","state":"enabled"}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	key := publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "switch"}
	server := New(publish.PublishedState{Subjects: map[publish.SubjectKey]publish.SubjectView{key: {Key: key, Name: "Switch"}}}, "")
	server.tokenAssetsDir = directory
	for _, partial := range []bool{false, true} {
		request := httptest.NewRequest("GET", "/components/switch?preview.compare=interactive", nil)
		if partial {
			request.Header.Set("HX-Request", "true")
			request.Header.Set("HX-Target", "component-preview")
		}
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, request)
		if response.Code != 200 {
			t.Fatal(response.Body.String())
		}
		body := response.Body.String()
		if strings.Contains(body, "<!doctype html>") == partial || !strings.Contains(body, `id="component-preview"`) {
			t.Fatal("wrong document or preview representation")
		}
		if partial && (response.Header().Get("Cache-Control") != "public, no-cache" || !strings.Contains(response.Header().Get("Vary"), "HX-Target")) {
			t.Fatal("preview caching must distinguish the full document")
		}
		if !strings.Contains(body, `hx-trigger="change"`) || !strings.Contains(body, `data-field-select`) {
			t.Fatal("preview must apply pill selections directly")
		}
	}
}
