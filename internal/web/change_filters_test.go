package web

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/sojupie/m3-source/internal/publish"
)

func TestChangeFiltersRunBeforePaginationAndPreserveSelection(t *testing.T) {
	state := publish.PublishedState{Snapshots: []publish.SnapshotView{{ID: "old", CreatedAt: "2026-06-01T12:00:00Z"}, {ID: "new", CreatedAt: "2026-06-02T12:00:00Z"}}}
	for i := range 55 {
		state.Changes = append(state.Changes, publish.ChangeView{Subject: "buttons", Name: fmt.Sprintf("font-%d", i), Type: "added", SnapshotID: "new"})
	}
	state.Changes = append(state.Changes,
		publish.ChangeView{Subject: "snackbar", Name: "md.comp.snackbar.container.width", Type: "changed", SnapshotID: "new", Before: "40", After: "48"},
		publish.ChangeView{Subject: "snackbar", Name: "old.width", Type: "removed", SnapshotID: "old"},
	)
	query := url.Values{"subject": {"snackbar"}, "type": {"changed"}, "snapshot": {"new"}, "q": {"width"}}
	changes, filters := filterChangeHistory(state, "", query)
	if len(changes) != 1 || len(filters.Subjects) != 2 || len(filters.Snapshots) != 2 {
		t.Fatalf("filtered result = %#v, choices = %#v", changes, filters)
	}
	recorder := httptest.NewRecorder()
	New(state, "").Handler().ServeHTTP(recorder, httptest.NewRequest("GET", "/changes?"+query.Encode(), nil))
	body := recorder.Body.String()
	if strings.Contains(body, `selected="false"`) {
		t.Fatal("false boolean attributes still select an option in HTML")
	}
	for _, text := range []string{`value="changed" selected`, "1 total changes", "0 added · 1 updated · 0 removed", `hx-trigger="toggle[event.target.open] once"`, "Load changes"} {
		if !strings.Contains(body, text) {
			t.Fatalf("missing %q", text)
		}
	}
	if strings.Contains(body, "font-54") || strings.Contains(body, `sl-change-disclosure" open`) {
		t.Fatal("unwanted or expanded change records rendered")
	}
	pageURL, err := url.Parse(changesPageURL("/changes?"+query.Encode(), 2))
	if err != nil {
		t.Fatal(err)
	}
	for key := range query {
		if pageURL.Query().Get(key) != query.Get(key) {
			t.Fatalf("pagination dropped %s", key)
		}
	}
	if pageURL.Query().Get("page") != "2" || strings.Contains(changesPageURL(pageURL.String(), 1), "page=") {
		t.Fatal("pagination must replace or remove the page number")
	}
	changes, filters = filterChangeHistory(state, "snackbar", url.Values{"subject": {"buttons"}})
	if len(changes) != 2 || !filters.FixedSubject || filters.Subject != "snackbar" {
		t.Fatal("component timeline escaped its subject")
	}
}
