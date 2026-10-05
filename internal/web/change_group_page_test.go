package web

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sojupie/m3-source/internal/publish"
)

func TestOuterPaginationKeepsWholeComponentsAndLoadsRecordsSeparately(t *testing.T) {
	state := publish.PublishedState{}
	for i := range 201 {
		state.Changes = append(state.Changes, publish.ChangeView{Subject: "component-00", SnapshotID: "new", Name: fmt.Sprintf("large-token-%03d", i)})
	}
	for i := 1; i <= 10; i++ {
		state.Changes = append(state.Changes, publish.ChangeView{Subject: fmt.Sprintf("component-%02d", i), SnapshotID: "new", Name: "small-token"})
	}
	selected, page, groups, previous, next := paginateChangeGroups(state.Changes, 1)
	if len(selected) != 210 || page != 1 || groups != 11 || previous || !next {
		t.Fatalf("component pagination split or lost records: %d, %d, %d, %v, %v", len(selected), page, groups, previous, next)
	}
	selected, _, _, _, _ = paginateChangeGroups(state.Changes, 2)
	if len(selected) != 1 || selected[0].Subject != "component-10" {
		t.Fatal("record count changed outer component pagination")
	}
	handler := New(state, "").Handler()
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/changes", nil))
	body := recorder.Body.String()
	if strings.Count(body, `<details class="sl-change-group"`) != 10 || strings.Contains(body, "large-token-") || len(body) > 30000 {
		t.Fatal("overview must render ten compact groups, not their inner records")
	}
	if strings.Contains(body, "@materialSymbol") || strings.Contains(body, `selected="false"`) {
		t.Fatal("template text or invalid boolean attributes rendered")
	}
	recorder = httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/changes/records?subject=component-00&snapshot=new&page=2", nil)
	request.Header.Set("HX-Request", "true")
	handler.ServeHTTP(recorder, request)
	body = recorder.Body.String()
	if !strings.Contains(body, "large-token-050") || !strings.Contains(body, "large-token-099") || strings.Contains(body, "large-token-100") || strings.Contains(body, "<!doctype html>") {
		t.Fatal("expanded group did not load its own bounded record page")
	}
	if !strings.Contains(body, `hx-target="closest .sl-change-records"`) || !strings.Contains(body, "page=3") {
		t.Fatal("record navigation must remain inside its component group")
	}
}

func TestMaterialIconsDoNotUseHandDrawnOrCharacterFallbacks(t *testing.T) {
	for _, name := range []string{"search", "eye", "grid", "expand", "dropdown", "folder"} {
		icon := tableIcon(name)
		if !strings.Contains(icon, `viewBox="0 -960 960 960"`) || !strings.Contains(icon, `class="sl-material-symbol"`) || strings.Contains(icon, "stroke=") {
			t.Fatalf("%s is not the shared Material Symbols SVG", name)
		}
	}
}
