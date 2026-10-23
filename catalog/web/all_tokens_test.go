package web

import (
	"github.com/sojupie/tokens/source/publish"
	"html"
	"net/http/httptest"
	"net/url"
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
	if got := tokenSetURL(subject, all.Selected); got != "/components/date-pickers/sets/all" {
		t.Fatal(got)
	}
	if got := tokenSetLinkURL(subject, all.Selected, tokenBrowser{Query: "container"}); got != "/components/date-pickers/sets/all?q=container" {
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
			if !strings.Contains(nav[1], `href="/md3/components/date-pickers"`) {
				t.Fatal("default option does not link to root")
			}
			if !regexp.MustCompile(`(?s)Docked.*href="/md3/components/date-pickers/sets/all"[^>]*>All tokens</a>\s*$`).MatchString(nav[1]) {
				t.Fatal("All tokens must be last and have its own URL")
			}
		})
	}
}
