package web

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sojupie/tokens/source/publish"
)

func TestCollectionDownloadsAndSharedControls(t *testing.T) {
	state := publish.PublishedState{CurrentSnapshot: publish.SnapshotView{ID: "fixture", CreatedAt: "2026-10-07T00:00:00Z"}, Subjects: map[publish.SubjectKey]publish.SubjectView{}}
	for _, key := range []publish.SubjectKey{{Kind: publish.SubjectComponent, Slug: "buttons"}, {Kind: publish.SubjectStyle, Slug: "color"}} {
		state.Subjects[key] = publish.SubjectView{Key: key, Name: key.Slug, Tokens: []publish.TokenView{{SourceName: "md.sys.color.primary"}}}
	}
	handler := New(state, "").Handler()
	for _, scope := range []string{"all", "components", "styles"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", "/downloads/tokens.json?scope="+scope+"&q=no-match", nil))
		if response.Code != 200 || !strings.Contains(response.Header().Get("Content-Disposition"), "attachment") {
			t.Fatalf("download %s: %d %v", scope, response.Code, response.Header())
		}
		var document struct {
			CapturedAt string                  `json:"capturedAt"`
			Subjects   []publish.PublicSubject `json:"subjects"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &document); err != nil {
			t.Fatal(err)
		}
		want := 1
		if scope == "all" {
			want = 2
		}
		if len(document.Subjects) != want || document.CapturedAt != state.CurrentSnapshot.CreatedAt {
			t.Fatalf("wrong export %s: %#v", scope, document)
		}
		for _, subject := range document.Subjects {
			if scope == "components" && subject.Kind != publish.SubjectComponent || scope == "styles" && subject.Kind == publish.SubjectComponent {
				t.Fatalf("wrong subject in %s", scope)
			}
			if len(subject.Tokens) != 1 {
				t.Fatal("missing tokens")
			}
		}
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "/downloads/tokens.json?scope=invalid", nil))
	if response.Code != 400 {
		t.Fatal("accepted invalid collection scope")
	}
	for _, path := range []string{"/", "/components", "/styles", "/components/buttons", "/styles/color", "/changes"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", path, nil))
		body := response.Body.String()
		if response.Code != 200 {
			t.Fatalf("page %s: %d", path, response.Code)
		}
		for _, text := range []string{`class="tokens-brand-logo"`, `rel="icon"`, `id="tokens-site-download"`, `value="all"`, `value="components"`, `value="styles"`} {
			if !strings.Contains(body, text) {
				t.Fatalf("%s missing %s", path, text)
			}
		}
		if strings.Count(body, `id="tokens-site-download"`) != 1 || strings.Count(body, `src="/static/token-table.js`) != 1 {
			t.Fatalf("duplicate dialog or script on %s", path)
		}
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "/downloads/component/buttons.json?scope=everything", nil))
	if response.Code != 303 || response.Header().Get("Location") != "/downloads/tokens.json?scope=all" {
		t.Fatal("subject-wide export does not lead to all tokens")
	}
}
