package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sojupie/m3-source/internal/publish"
)

func TestChangeTimelineShowsEarlierTransitionsGloballyAndPerSubject(t *testing.T) {
	state := publish.PublishedState{
		Snapshots: []publish.SnapshotView{{ID: "a", CreatedAt: "2026-06-01T12:00:00Z"}, {ID: "b", CreatedAt: "2026-06-02T12:00:00Z"}, {ID: "c", CreatedAt: "2026-06-03T12:00:00Z"}},
		Changes: []publish.ChangeView{
			{Subject: "snackbar", Name: "Width", Type: "changed", Before: "48", After: "56", PreviousSnapshotID: "b", SnapshotID: "c"},
			{Subject: "snackbar", Name: "Width", Type: "changed", Before: "40", After: "48", PreviousSnapshotID: "a", SnapshotID: "b"},
		},
	}
	for _, path := range []string{"/changes", "/components/snackbar/changes"} {
		recorder := httptest.NewRecorder()
		New(state, "").Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		body := recorder.Body.String()
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s: %d", path, recorder.Code)
		}
		for _, text := range []string{"Observed", "3 June 2026", "2 June 2026", "40", "56"} {
			if !strings.Contains(body, text) {
				t.Fatalf("%s missing %q", path, text)
			}
		}
		if strings.Index(body, "Observed <time datetime=\"2026-06-03") > strings.Index(body, "Observed <time datetime=\"2026-06-02") {
			t.Fatal("timeline not newest first")
		}
	}
}

func TestExistingSubjectWithoutChangesHasAnEmptyTimeline(t *testing.T) {
	key := publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "buttons"}
	state := publish.PublishedState{Subjects: map[publish.SubjectKey]publish.SubjectView{key: {Key: key}}}
	recorder := httptest.NewRecorder()
	New(state, "").Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/buttons/changes", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "At least two snapshots") {
		t.Fatalf("empty timeline = %d %s", recorder.Code, recorder.Body.String())
	}
}
