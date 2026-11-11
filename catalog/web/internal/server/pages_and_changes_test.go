// Integration checks for page.
package server

import (
	"fmt"
	webassets "github.com/sojupie/tokens/catalog/web/assets"
	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	"github.com/sojupie/tokens/source/publish"
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestAllTokensPreservesTablesGroupsAndDownloads(t *testing.T) {
	subject := publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "date-pickers"}, Name: "Date pickers"}
	for i, variant := range []string{"modal", "docked"} {
		subject.Tokens = append(subject.Tokens, publish.TokenView{
			SourceName: "md.comp.date-picker." + variant + ".container.color", Category: "COLOR",
			Presentation: &publish.TokenPresentation{SetName: "md.comp.date-picker." + variant, SetLabel: "Date picker - " + variant, SetOrder: i, Groups: []publish.DisplayGroup{{Name: "Enabled"}}},
		})
	}
	all := buildTokenBrowser(subject, url.Values{"set": {"all"}})
	if all.Selected.Key != "all" || all.Selected.Label != "All tokens" || all.Count != 2 || len(all.Selected.Groups) != 2 {
		t.Fatalf("wrong aggregate: %+v", all)
	}
	for i, group := range all.Selected.Groups {
		if group.Name != all.Sets[i].Label || len(group.Children) != 1 || group.Children[0].Name != "Enabled" {
			t.Fatalf("lost source hierarchy: %+v", group)
		}
	}
	if got := tokenDownloadSubject(subject, all, "set"); len(got.Tokens) != 2 {
		t.Fatal("aggregate download lost tokens")
	}
	selected := buildTokenBrowser(subject, url.Values{"set": {"md.comp.date-picker.docked"}})
	if selected.Count != 1 || selected.Selected.Groups[0].Name != "Enabled" {
		t.Fatal("individual table changed")
	}
	if got := webview.TokenSetURL(subject, all.Selected); got != "/components/date-pickers/sets/all" {
		t.Fatal(got)
	}
	if got := webview.TokenSetLinkURL(subject, all.Selected, webview.TokenBrowser{Query: "container"}); got != "/components/date-pickers/sets/all?q=container" {
		t.Fatal(got)
	}
}

func TestDefaultAndAllViewsAreServerRenderedAndAddressable(t *testing.T) {
	key := publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "date-pickers"}
	subject := publish.SubjectView{Key: key, Name: "Date pickers"}
	for i, variant := range []string{"Modal", "Docked"} {
		subject.Tokens = append(subject.Tokens, publish.TokenView{SourceName: "md.comp.date-picker." + strings.ToLower(variant) + ".container.color", Category: "COLOR", Presentation: &publish.TokenPresentation{SetName: "md.comp.date-picker." + strings.ToLower(variant), SetLabel: "Date picker - " + variant, SetOrder: i}})
	}
	server := New(publish.PublishedState{Subjects: map[publish.SubjectKey]publish.SubjectView{key: subject}}, "")
	if err := server.SetPublicOrigin("https://tokens.example/md3"); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		path, title, canonical string
		both                   bool
	}{
		{"", `Material Design 3 "Date Picker" tokens`, "", false},
		{"/sets/md.comp.date-picker.modal", `Material Design 3 "Date Picker" tokens`, "", false},
		{"/sets/all", `Material Design 3 "Date Picker" tokens: All tokens`, "/sets/all", true},
		{"?set=all", `Material Design 3 "Date Picker" tokens: All tokens`, "/sets/all", true},
	} {
		t.Run(tc.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/md3/components/date-pickers"+tc.path, nil))
			body := html.UnescapeString(response.Body.String())
			if response.Code != 200 {
				t.Fatal(response.Code)
			}
			if !strings.Contains(body, "<title>"+tc.title+" | Token Catalog</title>") {
				t.Fatal("wrong title")
			}
			if !strings.Contains(body, `rel="canonical" href="https://tokens.example/md3/components/date-pickers`+tc.canonical+`"`) {
				t.Fatal("wrong canonical")
			}
			if !strings.Contains(body, "md.comp.date-picker.modal.container.color") {
				t.Fatal("first table missing from SSR")
			}
			if strings.Contains(body, "md.comp.date-picker.docked.container.color") != tc.both {
				t.Fatal("wrong SSR table membership")
			}
			nav := regexp.MustCompile(`(?s)<nav[^>]*aria-label="Token sets"[^>]*>(.*?)</nav>`).FindStringSubmatch(body)
			if len(nav) != 2 {
				t.Fatal("missing table selector")
			}
			if strings.Contains(nav[1], "<h3") {
				t.Fatal("table selector must not include section labels")
			}
			if !strings.Contains(nav[1], `href="/md3/components/date-pickers"`) {
				t.Fatal("default option does not link to root")
			}
			if !regexp.MustCompile(`(?s)Docked.*<hr class="md-divider" data-inset="middle">\s*<section aria-label="All tokens">\s*<ul>\s*<li>\s*<a[^>]*href="/md3/components/date-pickers/sets/all"[^>]*>\s*<span class="md-menu-item-text">All tokens</span>\s*<span data-slot="trailing-icon"[^>]*>.*?</span>\s*</a>\s*</li>\s*</ul>\s*</section>\s*$`).MatchString(nav[1]) {
				t.Fatal("All tokens must be last and have its own URL")
			}
		})
	}
}

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
	if strings.Contains(body, ` selected="false"`) {
		t.Fatal("false boolean attributes still select an option in HTML")
	}
	for _, text := range []string{`value="changed" selected`, "1 item changed", "/components/snackbar/changes?", "View changes"} {
		if !strings.Contains(body, text) {
			t.Fatalf("missing %q", text)
		}
	}
	if strings.Contains(body, "font-54") || strings.Contains(body, `tokens-change-disclosure" open`) {
		t.Fatal("unwanted or expanded change records rendered")
	}
	pageURL, err := url.Parse(webview.ChangesPageURL("/changes?"+query.Encode(), 2))
	if err != nil {
		t.Fatal(err)
	}
	for key := range query {
		if pageURL.Query().Get(key) != query.Get(key) {
			t.Fatalf("pagination dropped %s", key)
		}
	}
	if pageURL.Query().Get("page") != "2" || strings.Contains(webview.ChangesPageURL(pageURL.String(), 1), "page=") {
		t.Fatal("pagination must replace or remove the page number")
	}
	changes, filters = filterChangeHistory(state, "snackbar", url.Values{"subject": {"buttons"}})
	if len(changes) != 2 || !filters.FixedSubject || filters.Subject != "snackbar" {
		t.Fatal("component timeline escaped its subject")
	}
}

func TestGlobalHistorySummarizesWholeComponentsWithoutNestedPaging(t *testing.T) {
	state := publish.PublishedState{}
	for i := range 201 {
		state.Changes = append(state.Changes, publish.ChangeView{Subject: "component-00", SnapshotID: "new", Name: fmt.Sprintf("large-token-%03d", i)})
	}

	state.Changes = append(state.Changes, state.Changes[0])
	for i := 1; i <= 10; i++ {
		state.Changes = append(state.Changes, publish.ChangeView{Subject: fmt.Sprintf("component-%02d", i), SnapshotID: "new", Name: "small-token"})
	}
	state.Changes = append(state.Changes, publish.ChangeView{Subject: "component-00", SnapshotID: "old", Name: "earlier-token"})
	handler := New(state, "").Handler()
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/changes?page=8", nil))
	body := recorder.Body.String()
	if strings.Count(body, `class="md-card md-outlined-card tokens-change-summary"`) != 12 || !strings.Contains(body, "201 items changed") || strings.Contains(body, "large-token-") {
		t.Fatal("summary must batch by observation/component and count distinct items")
	}
	for _, unwanted := range []string{"Load changes", "tokens-pagination", "@materialSymbol", ` selected="false"`} {
		if strings.Contains(body, unwanted) {
			t.Fatalf("overview still contains %s", unwanted)
		}
	}
	if !strings.Contains(body, "/components/component-00/changes?snapshot=new") || !strings.Contains(body, "/components/component-00/changes?snapshot=old") {
		t.Fatal("summary links must select the relevant component and observation")
	}
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/components/component-00/changes?snapshot=new&page=2", nil))
	body = recorder.Body.String()
	if !strings.Contains(body, "large-token-050") || !strings.Contains(body, "large-token-099") || strings.Contains(body, "large-token-100") || strings.Contains(body, "earlier-token") {
		t.Fatal("component page must show bounded details for the chosen observation")
	}
}

func TestMaterialIconsDoNotUseHandDrawnOrCharacterFallbacks(t *testing.T) {
	for _, name := range []string{"search", "eye", "grid", "expand", "dropdown", "folder"} {
		icon := webview.TableIcon(name)
		if !strings.Contains(icon, `<use href="#tokens-table-icon-`) || !strings.Contains(icon, `class="tokens-material-symbol"`) || strings.Contains(icon, "stroke=") {
			t.Fatalf("%s is not the shared Material Symbols SVG", name)
		}
	}
}

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
		if got := webview.ChangeHierarchy(test.name); !reflect.DeepEqual(got, test.want) {
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
	groups := webview.ChangeGroups([]publish.ChangeView{{Subject: "buttons", Name: "md.comp.elevated-button.label-text.type"}, {Subject: "buttons", Name: "md.comp.outlined-button.label-text.type"}}, subjects)
	if groups[0].Changes[0].Name != "md.comp.elevated-button.label-text.type" || groups[0].Changes[1].Name != "md.comp.outlined-button.label-text.type" {
		t.Fatalf("ambiguous names: %#v", groups)
	}
}

func TestChangeGroupsCombineInterleavedComponents(t *testing.T) {
	groups := webview.ChangeGroups([]publish.ChangeView{
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
		Subjects: map[publish.SubjectKey]publish.SubjectView{key: {Key: key, Name: "App bars"}},
		Changes:  []publish.ChangeView{{Subject: "app-bars", Name: "width", Type: "added", SnapshotID: "new"}},
	}
	for _, suffix := range []string{"", "?snapshot=new", "?q=missing", "?page=2"} {
		recorder := httptest.NewRecorder()
		New(state, "").Handler().ServeHTTP(recorder, httptest.NewRequest("GET", "/components/app-bars/changes"+suffix, nil))
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `>Material Design 3 &#34;App Bar&#34; token changes</h1>`) || strings.Contains(recorder.Body.String(), "Search changes") {
			t.Fatalf("incorrect scoped heading for %s", suffix)
		}
	}
}

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
		{path: "/", body: []string{"Material Design 3 tokens, playgrounds and change history", "Browse Material Design 3 components and styles", "/components/snackbar", "/styles/elevation"}, status: http.StatusOK},
		{path: "/components/snackbar", body: []string{"Snackbar", "Container", "Snackbar container width", "md.comp.snackbar.container.width", "560", "Download"}, status: http.StatusOK},
		{path: "/styles/elevation", body: []string{"Elevation", "Download"}, status: http.StatusOK},
		{path: "/styles/motion/overview", body: []string{"Motion", "Download"}, status: http.StatusOK},
		{path: "/components?q=snack", body: []string{"snack", "hx-target=\"#subject-list\""}, status: http.StatusOK},
		{path: "/changes", body: []string{"Change log", "1 item changed", "/components/snackbar", "View changes"}, status: http.StatusOK},
		{path: "/components/snackbar/changes", body: []string{"Material Design 3 &#34;Snackbar&#34; token changes", "md.comp.snackbar.container.width", "480", "560"}, status: http.StatusOK},
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
			body := recorder.Body.String()
			initial := strings.Index(body, "localStorage.getItem('tokens-theme')")
			styles := strings.Index(body, `rel="stylesheet"`)
			if initial < 0 || styles < initial || !strings.Contains(body, `<link rel="stylesheet" href="`+webassets.StaticURL("site.css")+`"`) || !strings.Contains(body, `<script src="`+webassets.StaticURL("theme.js")+`" defer></script>`) {
				t.Fatal("saved theme must be inline before styles; controls must be deferred")
			}
			if !strings.Contains(body, `<svg xmlns="http://www.w3.org/2000/svg" class="tokens-brand-logo"`) {
				t.Fatal("header must use the inline SVG logo")
			}
			for _, value := range []string{`<footer class="tokens-site-footer"`, `href="https://marmenlind.com"`, "Made by ", "Olle Marmenlind", "Independent project. Not affiliated with Google."} {
				if !strings.Contains(recorder.Body.String(), value) {
					t.Fatalf("page footer does not contain %q", value)
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

func TestTokenValuesHaveReadableForms(t *testing.T) {
	context := publish.TokenValueView{
		ContextTags:      []string{"designSystems/abc/contextTagGroups/def/tags/ghi"},
		ContextLabels:    []string{"Theme: Dark"},
		IsDefaultContext: true,
	}
	if got := tokenContextLabel(context); got != "Default · Dark" {
		t.Fatalf("context label = %q", got)
	}
	if got := webview.TokenValueText([]byte(`{"color":{"red":1,"green":0.5,"blue":0,"alpha":1}}`)); got != "rgba(255, 127.5, 0, 1)" {
		t.Fatalf("color value = %q", got)
	}
	if got := webview.TokenValueText([]byte(`{"dimension":{"value":480,"unit":"DIPS"}}`)); got != "480dp" {
		t.Fatalf("dimension value = %q", got)
	}
}

func TestShapeTokenFormatting(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "circular full",
			input: `{"shape":{"family":"SHAPE_FAMILY_CIRCULAR"}}`,
			want:  "Full",
		},
		{
			name:  "uniform 16dp",
			input: `{"shape":{"family":"SHAPE_FAMILY_ROUNDED_CORNERS","defaultSize":{"value":16,"unit":"DIPS"}}}`,
			want:  "16dp",
		},
		{
			name:  "uniform 0dp none",
			input: `{"shape":{"family":"SHAPE_FAMILY_ROUNDED_CORNERS","defaultSize":{"unit":"DIPS"}}}`,
			want:  "0dp",
		},
		{
			name:  "large end rounding",
			input: `{"shape":{"family":"SHAPE_FAMILY_ROUNDED_CORNERS","topRight":{"value":16,"unit":"DIPS"},"bottomRight":{"value":16,"unit":"DIPS"},"defaultSize":{"unit":"DIPS"}}}`,
			want:  "Default: 0dp · Top-right: 16dp · Bottom-right: 16dp",
		},
		{
			name:  "large start rounding",
			input: `{"shape":{"family":"SHAPE_FAMILY_ROUNDED_CORNERS","topLeft":{"value":16,"unit":"DIPS"},"bottomLeft":{"value":16,"unit":"DIPS"},"defaultSize":{"unit":"DIPS"}}}`,
			want:  "Default: 0dp · Top-left: 16dp · Bottom-left: 16dp",
		},
		{
			name:  "large top rounding",
			input: `{"shape":{"family":"SHAPE_FAMILY_ROUNDED_CORNERS","topLeft":{"value":16,"unit":"DIPS"},"topRight":{"value":16,"unit":"DIPS"},"defaultSize":{"unit":"DIPS"}}}`,
			want:  "Default: 0dp · Top-left: 16dp · Top-right: 16dp",
		},
		{
			name:  "all four corners specified identically",
			input: `{"shape":{"family":"SHAPE_FAMILY_ROUNDED_CORNERS","topLeft":{"value":12,"unit":"DIPS"},"topRight":{"value":12,"unit":"DIPS"},"bottomLeft":{"value":12,"unit":"DIPS"},"bottomRight":{"value":12,"unit":"DIPS"}}}`,
			want:  "12dp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := webview.TokenValueText([]byte(tt.input))
			if got != tt.want {
				t.Fatalf("tokenValueText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHTMLChangeLogIsPaginated(t *testing.T) {
	changes := make([]publish.ChangeView, 51)
	for index := range changes {
		changes[index] = publish.ChangeView{
			ID:                 fmt.Sprintf("change-%02d", index),
			Type:               "changed",
			Subject:            "carousel",
			Name:               fmt.Sprintf("token-%02d", index),
			Before:             `{"dimension":{"value":1,"unit":"DIPS"}}`,
			After:              `{"dimension":{"value":2,"unit":"DIPS"}}`,
			PreviousSnapshotID: "snapshot-0",
		}
	}
	handler := New(publish.PublishedState{
		CurrentSnapshot: publish.SnapshotView{ID: "snapshot-1"},
		Changes:         changes,
	}, "").Handler()

	firstPage := httptest.NewRecorder()
	handler.ServeHTTP(firstPage, httptest.NewRequest(http.MethodGet, "/changes", nil))
	if !strings.Contains(firstPage.Body.String(), "51 items changed") || strings.Contains(firstPage.Body.String(), "token-49") {
		t.Fatalf("unexpected first change page:\n%s", firstPage.Body.String())
	}
	if strings.Contains(firstPage.Body.String(), `href="/changes?page=2"`) {
		t.Fatal("record count split a single component across outer pages")
	}
	firstPage = httptest.NewRecorder()
	handler.ServeHTTP(firstPage, httptest.NewRequest(http.MethodGet, "/components/carousel/changes", nil))
	if !strings.Contains(firstPage.Body.String(), "token-49") || strings.Contains(firstPage.Body.String(), "token-50") {
		t.Fatal("record page is not bounded")
	}

	secondPage := httptest.NewRecorder()
	handler.ServeHTTP(secondPage, httptest.NewRequest(http.MethodGet, "/components/carousel/changes?page=2", nil))
	if !strings.Contains(secondPage.Body.String(), "token-50") || !strings.Contains(secondPage.Body.String(), `href="/components/carousel/changes"`) {
		t.Fatalf("unexpected second change page:\n%s", secondPage.Body.String())
	}
}
