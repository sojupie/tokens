package web

import (
	"fmt"
	"github.com/sojupie/tokens/source/publish"
	"net/http/httptest"
	"strings"
	"testing"
)

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
