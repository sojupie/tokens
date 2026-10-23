package web

import (
	"github.com/sojupie/tokens/source/publish"
	"net/url"
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
	all := buildTokenBrowser(subject, url.Values{})
	if all.Selected.Key != "" || all.Selected.Label != "All tokens" || all.Count != 2 || len(all.Selected.Groups) != 2 {
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
	if got := tokenSetURL(subject, all.Selected); got != "/components/date-pickers" {
		t.Fatal(got)
	}
	if got := tokenSetLinkURL(subject, all.Selected, tokenBrowser{Query: "container"}); got != "/components/date-pickers?q=container" {
		t.Fatal(got)
	}
}
