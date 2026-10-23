package web

import (
	"net/url"
	"strings"
	"testing"

	"github.com/sojupie/tokens/source/publish"
)

func TestTokenSetLinksKeepOnlySupportedContexts(t *testing.T) {
	subject := publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "buttons"}}
	set := tokenSet{component: "button", category: "COLOR", variant: "elevated", sourceTokens: []publish.TokenView{
		{Values: []publish.TokenValueView{{ContextLabels: []string{"Theme: Dark", "Platform: Web"}}, {ContextLabels: []string{"Theme: Light", "Platform: Android"}}}},
	}}
	browser := tokenBrowser{Contexts: url.Values{"context.Theme": {"Dark"}, "context.Platform": {"Android"}, "context.Unknown": {"unsupported"}}}
	link, err := url.Parse(tokenSetLinkURL(subject, set, browser))
	if err != nil {
		t.Fatal(err)
	}
	if link.Query().Get("context.Theme") != "Dark" || link.Query().Has("context.Platform") || link.Query().Has("context.Unknown") {
		t.Fatalf("incompatible context carried into link: %s", link)
	}
	if link.Path != "/components/buttons/sets/button/color/elevated" {
		t.Fatal(link)
	}
	browser.Contexts = nil
	if strings.Contains(tokenSetLinkURL(subject, set, browser), "?") {
		t.Fatal("default selection should link to the clean indexable URL")
	}
}

func TestSelectorDistinguishesSourceFamiliesAndVariants(t *testing.T) {
	variant := tokenSet{Label: "Icon button - Color - Outlined", component: "icon-button", variant: "outlined", category: "COLOR"}
	family := tokenSet{component: "outlined-icon-button", category: "COLOR"}
	groups := tokenSetMenuGroups(publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectComponent}}, []tokenSet{variant, family})
	if len(groups) != 2 || groups[0].Label != "Icon Button" || groups[1].Label != "Outlined Icon Button" || tokenTableQualifier(publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectComponent}, Name: "Icon buttons"}, variant) != "Outlined / Color" {
		t.Fatalf("selector merged different source families: %+v", groups)
	}
}

func TestStyleTokenSetsHaveProperLabels(t *testing.T) {
	subject := publish.SubjectView{
		Key:  publish.SubjectKey{Kind: publish.SubjectStyle, Slug: "shape/corner-radius"},
		Name: "Shape",
		Tokens: []publish.TokenView{
			{SourceName: "md.sys.shape.corner.large", Category: "SHAPE"},
		},
	}
	sets := tokenSets(subject)
	if len(sets) != 1 {
		t.Fatalf("len(sets) = %d, want 1", len(sets))
	}
	if sets[0].Label != "Shape" {
		t.Fatalf("sets[0].Label = %q, want %q", sets[0].Label, "Shape")
	}
}
