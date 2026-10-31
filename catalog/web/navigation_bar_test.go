package web

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestNavigationBarConfigurations(t *testing.T) {
	for _, variant := range []string{"adaptive", "vertical", "horizontal", "baseline"} {
		page := &componentPreview{Cases: []componentCase{{Family: "nav-bar", Variant: variant, State: "interactive", Selected: true}}}
		if err := configureNavBarPreview(page, map[string][]string{"preview.destinations": {"5"}, "preview.viewport": {"compact"}, "preview.badge": {"labeled"}, "preview.direction": {"rtl"}}); err != nil {
			t.Fatal(err)
		}
		c := page.Cases[0]
		if len(c.NavigationLabels()) != 5 {
			t.Fatal("missing destinations")
		}
		wantOrientation := variant
		if variant == "baseline" {
			wantOrientation = "vertical"
		}
		if c.Attributes()["data-orientation"] != wantOrientation {
			t.Errorf("%s: wrong orientation", variant)
		}
		var out bytes.Buffer
		if err := specimenNavBar(c).Render(context.Background(), &out); err != nil {
			t.Fatal(err)
		}
		html := out.String()
		if strings.Count(html, `data-navigation-item`) != 5 || strings.Count(html, `aria-current="page"`) != 1 {
			t.Fatal("bar must have five destinations and one selected destination")
		}
		for _, want := range []string{`dir="rtl"`, `data-navbar-page="Settings"`, `md-nav-badge`, `Files 3 unread files`, `class="md-nav-items"`} {
			if !strings.Contains(html, want) {
				t.Errorf("%s missing %s", variant, want)
			}
		}
		if c.NavigationItem(0)["aria-controls"] != c.ID()+"-screen" {
			t.Fatal("destination must control its view")
		}
	}
}

func TestNavigationBarRejectsUnsupportedDestinationCounts(t *testing.T) {
	for _, value := range []string{"2", "6", "0", "all"} {
		if err := configureNavBarPreview(&componentPreview{}, map[string][]string{"preview.destinations": {value}}); err == nil {
			t.Errorf("accepted count %s", value)
		}
	}
}

func TestHorizontalNavigationUsesAMediumExampleWindow(t *testing.T) {
	page := &componentPreview{Cases: []componentCase{{Family: "nav-bar", Variant: "horizontal"}}}
	if err := configureNavBarPreview(page, map[string][]string{"preview.viewport": {"compact"}}); err != nil {
		t.Fatal(err)
	}
	if page.Cases[0].NavViewport != "medium" {
		t.Fatal("horizontal items must not be demonstrated in an unsupported compact window")
	}
	for _, control := range page.Controls {
		if control.Name == "viewport" && containsPreview(control.Options, "compact") {
			t.Fatal("compact horizontal example is not valid")
		}
	}
}
