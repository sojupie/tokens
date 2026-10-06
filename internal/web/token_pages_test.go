package web

import (
	"strings"
	"testing"

	"github.com/sojupie/m3-source/internal/publish"
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
	if got := tokenSetTitle(set); got != "Filled tonal icon button: toggle colors — Material Design" {
		t.Fatal(got)
	}
	if got := tokenSetDescription(set); got != "Color tokens for the Material Design filled tonal icon button when used as a toggle, in selected and unselected states. View references and values by context, or download JSON." {
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
