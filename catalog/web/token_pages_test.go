package web

import (
	"net/url"
	"strings"
	"testing"

	"github.com/sojupie/tokens/source/publish"
)

func TestTokenSetCopySeparatesComponentAndSubset(t *testing.T) {
	set := tokenSet{component: "filled-tonal-icon-button", category: "COLOR", variant: "toggle", sourceTokens: []publish.TokenView{
		{SourceName: "md.comp.filled-tonal-icon-button.toggle.selected.container.color", Category: "COLOR"},
		{SourceName: "md.comp.filled-tonal-icon-button.toggle.unselected.container.color", Category: "COLOR"},
	}}
	if got := tokenSetComponent(set); got != "filled tonal icon button" {
		t.Fatal(got)
	}
	if got := tokenSetLabel(set); got != "Toggle · Color" {
		t.Fatal(got)
	}
	if got := tokenSetTitle(set); got != "Filled tonal icon button: toggle colors — Material Design 3" {
		t.Fatal(got)
	}
	if got := tokenSetDescription(set); got != "Color tokens for the Material Design 3 filled tonal icon button when used as a toggle, in selected and unselected states. View references and values by context, or download JSON." {
		t.Fatal(got)
	}
	set.sourceTokens = set.sourceTokens[:1]
	if strings.Contains(tokenSetDescription(set), "unselected states") {
		t.Fatal("description promises a state not present in the set")
	}
	set = tokenSet{component: "button", category: "COLOR", variant: "elevated", sourceTokens: []publish.TokenView{{Category: "COLOR"}, {Category: "OPACITY"}, {Category: "ELEVATION"}}}
	if tokenSetComponent(set) != "elevated button" || !strings.HasPrefix(tokenSetDescription(set), "Color, opacity and elevation tokens") {
		t.Fatal("variant or actual property types missing from copy")
	}
	if tokenPhrase("extended-fab") != "extended FAB" {
		t.Fatal("lost FAB acronym")
	}
}

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
	variant := tokenSet{component: "icon-button", variant: "outlined", category: "COLOR"}
	family := tokenSet{component: "outlined-icon-button", category: "COLOR"}
	groups := tokenSetMenuGroups([]tokenSet{variant, family})
	if len(groups) != 2 || groups[0].Label != "Icon button" || groups[1].Label != "Outlined icon button" || tokenSetMenuLabel(variant) != "Outlined · Color" {
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
