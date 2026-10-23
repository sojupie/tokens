package web

import (
	"html"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/sojupie/tokens/source/publish"
)

func TestComponentNameEnglishInflections(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"Buttons", "Button"}, {"Icon buttons", "Icon Button"}, {"Checkboxes", "Checkbox"}, {"Checkbox", "Checkbox"},
		{"Switches", "Switch"}, {"Switch", "Switch"}, {"Bottom sheets", "Bottom Sheet"}, {"Menus", "Menu"},
		{"Badges", "Badge"}, {"Carousels", "Carousel"}, {"Progress indicators", "Progress Indicator"},
		{"Status", "Status"}, {"Statuses", "Status"}, {"Search", "Search"}, {"Tabs", "Tab"},
		{"Navigation drawers", "Navigation Drawer"}, {"Floating action buttons (FABs)", "Floating Action Button (FAB)"},
		{"Extended FAB", "Extended FAB"}, {"FAB menus", "FAB Menu"}, {"iOS switches", "iOS Switch"},
		{"Properties", "Property"}, {"People", "Person"}, {"Series", "Series"}, {"Glass", "Glass"}, {"", ""},
	} {
		t.Run(tc.input, func(t *testing.T) {
			got := componentName(tc.input)
			if got != tc.want {
				t.Fatalf("%q, want %q", got, tc.want)
			}
			if componentName(got) != got {
				t.Fatalf("name not stable: %q", got)
			}
			subject := publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectComponent}, Name: tc.input}
			if !regexp.MustCompile(`^Material Design 3 "` + regexp.QuoteMeta(tc.want) + `" tokens$`).MatchString(subjectTokenTitle(subject)) {
				t.Fatal(subjectTokenTitle(subject))
			}
		})
	}
	// Style names and source paths must not undergo component inflection.
	style := publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectStyle, Slug: "states"}, Name: "States"}
	if subjectName(style) != "States" || subjectURL(style) != "/styles/states" {
		t.Fatal("style or URL renamed")
	}
}

func TestTableNamesKeepSourceDistinctions(t *testing.T) {
	for _, tc := range []struct{ subject, label, want string }{
		{"Icon buttons", "Icon button - Color - Filled", "Filled / Color"},
		{"Icon buttons", "Icon button - Size - Medium", "Medium / Size"},
		{"Date pickers", "Date picker - Modal", "Modal"},
		{"Date pickers", "Date picker - Modal input", "Modal Input"},
		{"Date pickers", "Date picker - Docked", "Docked"},
		{"Menus", "Menus - Common", "Common"},
		{"Menus", "Menus - Color - Standard", "Standard / Color"},
		{"Menus", "Menus - Color - Vibrant", "Vibrant / Color"},
		{"Menus", "Menu (baseline)", "Baseline"},
		{"Buttons", "[Deprecated] Button – Outlined", "Outlined / Deprecated"},
		{"Toolbars", "Bottom app bar (baseline)", "Bottom App Bar (baseline)"},
		{"Buttons", "Unfamiliar source label", "Unfamiliar Source Label"},
	} {
		t.Run(tc.label, func(t *testing.T) {
			subject := publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectComponent}, Name: tc.subject}
			set := tokenSet{Label: tc.label, sourceName: "unchanged.source.id"}
			if got := tokenTableQualifier(subject, set); got != tc.want {
				t.Fatalf("%q, want %q", got, tc.want)
			}
			if set.Label != tc.label || set.sourceName != "unchanged.source.id" {
				t.Fatal("source metadata mutated")
			}
			title := subjectTableTitle(subject, set)
			if !regexp.MustCompile(`^Material Design 3 "[^"]+" tokens: ` + regexp.QuoteMeta(tc.want) + `$`).MatchString(title) {
				t.Fatal(title)
			}
			if !strings.Contains(subjectTableDescription(subject, set), tc.want) {
				t.Fatal("description lost variant")
			}
		})
	}
}

func TestSourceTableSEOMatrix(t *testing.T) {
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
		path, qualifier, canonical string
		noindex                    bool
	}{
		{"/components/date-pickers", "", "/components/date-pickers", false},
		{"/components/date-pickers/sets/md.comp.date-picker.modal", "", "/components/date-pickers", false},
		{"/components/date-pickers?set=md.comp.date-picker.docked", ": Docked", "/components/date-pickers/sets/md.comp.date-picker.docked", false},
		{"/components/date-pickers?preview.compare=interactive", "", "", true},
		{"/components/date-pickers?context.Theme=Dark", "", "", true},
		{"/components/date-pickers?q=container", "", "", true},
		{"/components/date-pickers?utm_source=example", "", "/components/date-pickers", false},
		{"/components/date-pickers/sets/md.comp.date-picker.modal?q=container", "", "", true},
		{"/components/date-pickers/changes", "changes", "/components/date-pickers/changes", false},
	} {
		t.Run(tc.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/md3"+tc.path, nil))
			body := html.UnescapeString(response.Body.String())
			if response.Code != 200 {
				t.Fatalf("status %d", response.Code)
			}
			name := `Material Design 3 "Date Picker" tokens` + tc.qualifier
			if tc.qualifier == "changes" {
				name = `Material Design 3 "Date Picker" token changes`
			}
			for _, pattern := range []string{`<title>` + regexp.QuoteMeta(name) + ` \| Token Catalog</title>`, `<h1[^>]*>` + regexp.QuoteMeta(name) + `</h1>`} {
				if !regexp.MustCompile(pattern).MatchString(body) {
					t.Fatalf("missing %s", pattern)
				}
			}
			if tc.canonical != "" && !strings.Contains(body, `rel="canonical" href="https://tokens.example/md3`+tc.canonical+`"`) {
				t.Fatal("wrong canonical")
			}
			if tc.canonical == "" && strings.Contains(body, `rel="canonical"`) {
				t.Fatal("filtered view claims canonical")
			}
			if strings.Contains(body, `content="noindex, follow"`) != tc.noindex || (response.Header().Get("X-Robots-Tag") != "") != tc.noindex {
				t.Fatal("inconsistent indexing")
			}
			if tc.qualifier == ": Modal" || tc.qualifier == ": Docked" {
				if !regexp.MustCompile(`<meta name="description" content="Material Design 3 "Date Picker" tokens for ` + tc.qualifier[2:] + `\.`).MatchString(body) {
					t.Fatal("description lost variant")
				}
			}
		})
	}
	// A single source table repeats the entire component, so prefer the main URL.
	subject.Tokens = subject.Tokens[:1]
	server.state.Subjects[key] = subject
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/md3/components/date-pickers/sets/md.comp.date-picker.modal", nil))
	if !strings.Contains(response.Body.String(), `rel="canonical" href="https://tokens.example/md3/components/date-pickers"`) {
		t.Fatal("single table not consolidated")
	}
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/md3/sitemap.xml", nil))
	if strings.Contains(response.Body.String(), "/sets/") {
		t.Fatal("noncanonical table in sitemap")
	}
}

func TestPublishedTableNamesAndAggregateCoverage(t *testing.T) {
	path := os.Getenv("TOKEN_CATALOG_THEME_DB")
	if path == "" {
		t.Skip("set TOKEN_CATALOG_THEME_DB to inspect captured table names")
	}
	state, err := publish.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	total := 0
	for _, subject := range state.Subjects {
		all := buildTokenBrowser(subject, url.Values{"set": {"all"}})
		if all.Count != len(subject.Tokens) {
			t.Fatalf("%s: aggregate has %d of %d tokens", subject.Key.Slug, all.Count, len(subject.Tokens))
		}
		names := map[string]string{}
		for _, set := range all.Sets {
			title := subjectTableTitle(subject, set)
			if other, exists := names[title]; exists {
				t.Errorf("%s: %s and %s share %q", subject.Key.Slug, other, set.Key, title)
			}
			names[title] = set.Key
			if !regexp.MustCompile(`^Material Design 3 ".+" tokens: .+\S$`).MatchString(title) {
				t.Errorf("invalid title %q", title)
			}
			if strings.Contains(title, "Expressive") && !strings.Contains(strings.ToLower(set.Label), "expressive") {
				t.Errorf("inferred Expressive status: %q", title)
			}
			total++
		}
	}
	t.Logf("checked %d subjects and %d table titles", len(state.Subjects), total)
}
