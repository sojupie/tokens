package web

import "testing"

func TestMenuAnatomyAndTargets(t *testing.T) {
	for _, name := range []string{"star", "download", "edit", "check", "chevron_right"} {
		if materialSymbolSVG(name) == "" {
			t.Fatalf("menu icon missing: %s", name)
		}
	}
	c := componentCase{Family: "menu", Variant: "standard", Orientation: "vertical", Shape: "cascade", Expanded: true, State: "focused"}
	if c.MenuItem(0)["aria-expanded"] != "true" || c.MenuPanel(true)["data-menu-activity"] != "active" || c.MenuPanel(false)["data-menu-activity"] != "inactive" {
		t.Fatal("cascade preview must expose its active submenu")
	}
	c.State = "interactive"
	if c.MenuPanel(true)["popover"] != "auto" {
		t.Fatal("interactive submenu must use native popover")
	}
	c.Shape = "icon-only"
	c.Orientation = "horizontal"
	c.Selected = true
	c.State = "disabled"
	a := c.MenuItem(0)
	if a["aria-label"] != "Favorite" || a["aria-checked"] != "true" || a["disabled"] != true {
		t.Fatal("icon-only menu needs a label, selection and native disabled state")
	}
}

func TestSubmenuControlOnlyAppearsForCascades(t *testing.T) {
	cases := []componentCase{{Family: "menu", Variant: "standard", Orientation: "vertical", Shape: "plain", State: "enabled"}, {Family: "menu", Variant: "standard", Orientation: "vertical", Shape: "cascade", State: "enabled"}, {Family: "menu", Variant: "standard", Orientation: "vertical", Shape: "cascade", Expanded: true, State: "enabled"}}
	for _, shape := range []string{"plain", "cascade"} {
		preview, err := selectComponentPreview(cases, map[string][]string{"preview.shape": {shape}})
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, control := range preview.Controls {
			if control.Name == "disclosure" {
				found = true
			}
		}
		if found != (shape == "cascade") {
			t.Fatalf("submenu control in %s: %v", shape, found)
		}
	}
}

func TestConfiguredHorizontalMenuItemsCanBeSelected(t *testing.T) {
	for _, shape := range []string{"plain", "icon-only"} {
		c := componentCase{Family: "menu", Variant: "standard", Orientation: "horizontal", Shape: shape, Selected: true, State: "interactive", MenuMode: "multiple"}
		for index := range 3 {
			a := c.MenuItem(index)
			expected := "false"
			if index == 0 {
				expected = "true"
			}
			if a["role"] != "menuitemcheckbox" || a["aria-checked"] != expected || a["data-selected"] != expected {
				t.Fatalf("%s item %d lacks independent selection: %v", shape, index, a)
			}
		}
	}
}

func TestMenuBehaviorConfiguration(t *testing.T) {
	for _, mode := range []string{"actions", "single", "multiple"} {
		page := &componentPreview{Cases: []componentCase{{Family: "menu", Orientation: "horizontal", Shape: "plain", Selected: true}}}
		if err := configureMenuPreview(page, map[string][]string{"preview.behavior": {mode}}); err != nil {
			t.Fatal(err)
		}
		c := page.Cases[0]
		for index := range 3 {
			a := c.MenuItem(index)
			if mode == "actions" && index > 0 {
				if a["role"] != "menuitem" {
					t.Fatal("actions must remain actions")
				}
				continue
			}
			role := "menuitemcheckbox"
			if mode == "single" {
				role = "menuitemradio"
			}
			if a["role"] != role {
				t.Fatalf("%s item %d: %v", mode, index, a)
			}
			if mode != "actions" && (a["aria-checked"] == "true") != (index == 0) {
				t.Fatal("initial selected state must use the first item")
			}
		}
	}
	page := &componentPreview{Cases: []componentCase{{Family: "menu"}}}
	if configureMenuPreview(page, map[string][]string{"preview.behavior": {"invalid"}}) == nil {
		t.Fatal("invalid behavior accepted")
	}
}
