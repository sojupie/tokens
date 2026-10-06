package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sojupie/m3-source/internal/publish"
)

func seoTestState() publish.PublishedState {
	key := publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "snackbar"}
	return publish.PublishedState{Subjects: map[publish.SubjectKey]publish.SubjectView{
		key: {Key: key, Name: "Snackbar", Tokens: []publish.TokenView{
			{SourceName: "md.comp.snackbar.container.color", Category: "COLOR"},
			{SourceName: "md.comp.snackbar.container.width", Category: "LENGTH"},
		}},
	}}
}

func TestTokenSetPagesAreLinkedAndServerRendered(t *testing.T) {
	handler := New(seoTestState(), "").Handler()
	for _, path := range []string{"/components/snackbar", "/components/snackbar/sets/snackbar/common"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("%s: status %d", path, response.Code)
		}
		body := response.Body.String()
		for _, link := range []string{"/sets/snackbar/color", "/sets/snackbar/common"} {
			if !strings.Contains(body, `href="/components/snackbar`+link+`"`) {
				t.Fatalf("%s: missing token set link %s", path, link)
			}
		}
		if strings.HasSuffix(path, "/common") && (!strings.Contains(body, "md.comp.snackbar.container.width") || strings.Contains(body, "md.comp.snackbar.container.color")) {
			t.Fatal("token set page does not render the selected set")
		}
	}
	for _, path := range []string{"/components/snackbar/sets/", "/components/snackbar/sets/snackbar/unknown", "/components/snackbar/sets/snackbar/color/unknown"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusNotFound {
			t.Fatalf("%s: status %d, want 404", path, response.Code)
		}
	}
}
