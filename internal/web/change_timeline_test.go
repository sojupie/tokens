package web

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/sojupie/tokens/internal/publish"
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
		if strings.Contains(body, "@materialSymbol") {
			t.Fatal("template call leaked into visible text")
		}
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

func TestChangeHierarchyPreservesHistoricalPath(t *testing.T) {
	for _, test := range []struct {
		name string
		want []string
	}{
		{"md.comp.button.outlined.hover.container.state-layer.color", []string{"Button", "Outlined", "Hovered", "Container", "State Layer", "Color"}},
		{"md.comp.elevated-button.label-text.type", []string{"Elevated Button", "Label Text", "Type"}},
		{"md.sys.color.primary", []string{"Color", "Primary"}},
		{"Width", nil},
	} {
		if got := changeHierarchy(test.name); !reflect.DeepEqual(got, test.want) {
			t.Errorf("%s: got %v, want %v", test.name, got, test.want)
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

func TestTimelineUsesExactIdentifiersAndLinkedComponentGroups(t *testing.T) {
	key := publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "buttons"}
	state := publish.PublishedState{
		Subjects: map[publish.SubjectKey]publish.SubjectView{key: {
			Key: key, Name: "Buttons", Tokens: []publish.TokenView{{SourceName: "md.comp.button.elevated.label-text.weight", DisplayName: "Button elevated label text weight"}},
		}},
		Changes: []publish.ChangeView{
			{Subject: "buttons", Name: "md.comp.button.elevated.label-text.weight", Type: "changed", Before: `{"fontWeight":500}`, After: `{"fontWeight":600}`, SnapshotID: "current"},
			{Subject: "buttons", Name: "md.comp.button.removed.height", Type: "removed", Before: "40", SnapshotID: "current"},
		},
	}
	for _, path := range []string{"/changes", "/components/buttons/changes"} {
		recorder := httptest.NewRecorder()
		New(state, "").Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		body := recorder.Body.String()
		if !strings.Contains(body, `href="/components/buttons"`) || strings.Contains(body, "@materialSymbol") {
			t.Fatal("overview must retain linked component groups without template text")
		}
		recorder = httptest.NewRecorder()
		New(state, "").Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/buttons/changes?snapshot=current", nil))
		body = recorder.Body.String()
		level := "h4"
		for _, text := range []string{"<" + level + ">md.comp.button.elevated.label-text.weight</" + level + ">", "md.comp.button.removed.height", "Font weight", "Previous value", "New value", `aria-label="Token hierarchy"`, `Elevated</li>`, `Removed</li>`, `class="tokens-material-symbol"`} {
			if !strings.Contains(body, text) {
				t.Fatalf("%s missing %q", path, text)
			}
		}
		if strings.Contains(body, "Button elevated label text weight") {
			t.Fatal("heading still uses the formatted display name")
		}
	}
}

func TestTimelineDisambiguatesRepeatedSourceDisplayNames(t *testing.T) {
	key := publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "buttons"}
	subjects := map[publish.SubjectKey]publish.SubjectView{key: {Key: key, Name: "Buttons", Tokens: []publish.TokenView{
		{SourceName: "md.comp.elevated-button.label-text.type", DisplayName: "Type style"},
		{SourceName: "md.comp.outlined-button.label-text.type", DisplayName: "Type style"},
	}}}
	groups := changeGroups([]publish.ChangeView{{Subject: "buttons", Name: "md.comp.elevated-button.label-text.type"}, {Subject: "buttons", Name: "md.comp.outlined-button.label-text.type"}}, subjects)
	if groups[0].Changes[0].Name != "md.comp.elevated-button.label-text.type" || groups[0].Changes[1].Name != "md.comp.outlined-button.label-text.type" {
		t.Fatalf("ambiguous names: %#v", groups)
	}
}

func TestChangeGroupsCombineInterleavedComponents(t *testing.T) {
	groups := changeGroups([]publish.ChangeView{
		{Subject: "buttons", Name: "first"},
		{Subject: "snackbar", Name: "second"},
		{Subject: "buttons", Name: "third"},
	}, nil)
	if len(groups) != 2 || len(groups[0].Changes) != 2 || groups[0].Changes[1].Name != "third" {
		t.Fatalf("component grouping split or lost records: %#v", groups)
	}
}

func TestScopedChangeHeadingUsesRouteAcrossFiltersAndEmptyResults(t *testing.T) {
	key := publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "app-bars"}
	state := publish.PublishedState{
		Subjects: map[publish.SubjectKey]publish.SubjectView{key: {Key: key, Name: "Search"}},
		Changes:  []publish.ChangeView{{Subject: "app-bars", Name: "width", Type: "added", SnapshotID: "new"}},
	}
	for _, suffix := range []string{"", "?snapshot=new", "?q=missing", "?page=2"} {
		recorder := httptest.NewRecorder()
		New(state, "").Handler().ServeHTTP(recorder, httptest.NewRequest("GET", "/components/app-bars/changes"+suffix, nil))
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `>Material Design 3 app bars token changes</h1>`) || strings.Contains(recorder.Body.String(), "Search changes") {
			t.Fatalf("incorrect scoped heading for %s", suffix)
		}
	}
}
