package web

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestNavigationRailConfigurations(t *testing.T) {
	page := &componentPreview{Cases: []componentCase{{Family: "nav-rail", Variant: "expanded", Orientation: "vertical"}}}
	if err := configureNavigationPreview(page, map[string][]string{"preview.layout": {"modal"}, "preview.collapse": {"hidden"}, "preview.alignment": {"center"}, "preview.header": {"menu-fab"}, "preview.badge": {"labeled"}, "preview.indicator": {"fill"}}); err != nil {
		t.Fatal(err)
	}
	c := page.Cases[0]
	if previewDimensions(c)["orientation"] != "" {
		t.Fatal("rail label layout must follow its variant, not filter variants out")
	}
	if c.NavIndicator != "fill" || c.Orientation != "horizontal" || c.NavLayout != "modal" || c.NavCollapse != "hidden" || c.NavAlign != "center" || !c.RailFAB() || !c.RailMenu() {
		t.Fatalf("lost configuration: %#v", c)
	}
	if c.Attributes()["data-variant"] != "expanded" || c.Attributes()["data-layout"] != "modal" {
		t.Fatal("modal must configure expanded rather than replace it")
	}
	var html bytes.Buffer
	if err := componentSpecimen(c).Render(context.Background(), &html); err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"md-rail-frame", "md-rail-body", "data-navigation-secondary", "data-navigation-create"} {
		if !strings.Contains(html.String(), required) {
			t.Fatalf("missing %s", required)
		}
	}
	if !strings.Contains(materialSymbolSVG("keyboard_command_key"), "<path") {
		t.Fatal("command shortcut icon missing")
	}
}
func TestNavigationDrawerContent(t *testing.T) {
	var html bytes.Buffer
	if err := componentSpecimen(componentCase{Family: "nav-drawer", Variant: "modal", State: "interactive"}).Render(context.Background(), &html); err != nil {
		t.Fatal(err)
	}
	if strings.Count(html.String(), "data-navigation-item") != 8 || !strings.Contains(html.String(), "Workspace") {
		t.Fatal("drawer sections are incomplete")
	}
	if strings.Count(html.String(), "<hr") != 1 || strings.Contains(html.String(), ">Close</button>") {
		t.Fatal("drawer must have only the section divider and no footer close button")
	}
}
