package web

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestNavigationRailConfigurations(t *testing.T) {
	page := &componentPreview{Cases: []componentCase{{Family: "nav-rail", Variant: "expanded", Orientation: "vertical"}}}
	if err := configureNavigationPreview(page, map[string][]string{"preview.layout": {"modal"}, "preview.alignment": {"center"}, "preview.header": {"menu-fab"}, "preview.badge": {"labeled"}, "preview.indicator": {"fill"}}); err != nil {
		t.Fatal(err)
	}
	c := page.Cases[0]
	if previewDimensions(c)["orientation"] != "" {
		t.Fatal("rail label layout must follow its variant, not filter variants out")
	}
	if c.NavIndicator != "fill" || c.Orientation != "horizontal" || c.NavLayout != "modal" || c.NavAlign != "center" || !c.RailFAB() || !c.RailMenu() {
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

func TestRailMenuExistsOnlyInsideConfiguredRail(t *testing.T) {
	for _, variant := range []string{"collapsed", "expanded"} {
		for _, header := range []string{"none", "fab", "menu", "menu-fab"} {
			c := componentCase{Family: "nav-rail", Variant: variant, State: "interactive", NavHeader: header}
			var output bytes.Buffer
			if err := componentSpecimen(c).Render(context.Background(), &output); err != nil {
				t.Fatal(err)
			}
			html := output.String()
			count := strings.Count(html, "data-navigation-expand")
			if strings.Contains(html, "md-rail-external-menu") {
				t.Fatal("external menu rendered")
			}
			if c.RailMenu() {
				start, end := strings.Index(html, "<nav "), strings.Index(html, "</nav>")
				if count != 1 || start < 0 || end < start || !strings.Contains(html[start:end], "data-navigation-expand") {
					t.Fatalf("%s/%s: menu not exclusively inside rail", variant, header)
				}
			} else if count != 0 {
				t.Fatalf("%s/%s: menu rendered without configuration", variant, header)
			}
		}
	}
}

func TestRailAdditionalDestinationsAreExpandedOnly(t *testing.T) {
	c := componentCase{Family: "nav-rail"}
	for i := range 6 {
		p := navigationExampleItem(c, i, c.NavigationLabels()[i], c.NavigationIcon(i))
		if p.Secondary != (i >= 4) {
			t.Fatalf("item %d: wrong visibility policy", i)
		}
	}
}

func TestRailInteractiveStates(t *testing.T) {
	for _, state := range []string{"collapsed", "expanded", "hidden"} {
		for _, header := range []string{"none", "fab", "menu", "menu-fab"} {
			page := &componentPreview{Interactive: true, Controls: []previewControl{{Name: "variant"}}, Cases: []componentCase{{Family: "nav-rail", Variant: "collapsed", State: "interactive"}}}
			if err := configureNavigationPreview(page, map[string][]string{"preview.rail-state": {state}, "preview.header": {header}}); err != nil {
				t.Fatal(err)
			}
			c := page.Cases[0]
			for _, control := range page.Controls {
				if control.Name == "variant" || control.Name == "collapse" || control.Name == "secondary" {
					t.Fatalf("redundant control %s", control.Name)
				}
				if state == "hidden" && control.Name == "header" && (containsPreview(control.Options, "none") || containsPreview(control.Options, "fab")) {
					t.Fatal("hidden mode offered no-menu header")
				}
			}
			if state == "hidden" {
				if c.NavCollapse != "hidden" || !c.RailMenu() {
					t.Fatal("hidden rail needs an internal menu")
				}
			} else if c.Variant != state || c.NavCollapse == "hidden" {
				t.Fatalf("wrong initial state: %#v", c)
			}
			var output bytes.Buffer
			if err := componentSpecimen(c).Render(context.Background(), &output); err != nil {
				t.Fatal(err)
			}
			html := output.String()
			if strings.Contains(html, "md-rail-opener-slot") != (state == "hidden") {
				t.Fatal("immersive opener must retain a layout slot")
			}
			if strings.Contains(html, "md-rail-external-menu") != (state == "hidden") {
				t.Fatal("page opener must be exclusive to hidden mode")
			}
		}
	}
}
