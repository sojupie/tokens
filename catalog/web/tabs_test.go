package web

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestTabsLayoutsAndBadgeOwnership(t *testing.T) {
	for _, variant := range []string{"primary", "secondary"} {
		for _, control := range []string{"label", "icon", "icon-label"} {
			page := &componentPreview{Cases: []componentCase{{Family: "tabs", Variant: variant, Control: control, State: "interactive", Selected: true}}}
			err := configureTabsPreview(page, map[string][]string{"preview.layout": {"scrollable"}, "preview.badge": {"labeled"}, "preview.direction": {"rtl"}})
			if err != nil {
				t.Fatal(err)
			}
			c := page.Cases[0]
			var out bytes.Buffer
			if err := specimenTabs(c).Render(context.Background(), &out); err != nil {
				t.Fatal(err)
			}
			html := out.String()
			if strings.Count(html, `role="tab"`) != 6 || strings.Count(html, `role="tabpanel"`) != 6 || strings.Count(html, `aria-selected="true"`) != 1 {
				t.Fatal("tab/panel count or selection mismatch")
			}
			if strings.Count(html, `class="md-badge"`) != 1 || !strings.Contains(html, `dir="rtl"`) {
				t.Fatal("missing badge or RTL")
			}
			if c.TabHasBadge(0) || !c.TabHasBadge(1) {
				t.Fatal("badge must identify unread inactive tab")
			}
		}
	}
}
func TestTabsRejectCrampedFixedConfiguration(t *testing.T) {
	for _, count := range []string{"0", "1", "7"} {
		if configureTabsPreview(&componentPreview{}, map[string][]string{"preview.count": {count}}) == nil {
			t.Fatal("invalid fixed count", count)
		}
	}
}
func TestNavigationBadgeSurvivesEitherInitialSelection(t *testing.T) {
	for _, selected := range []bool{true, false} {
		page := &componentPreview{Cases: []componentCase{{Family: "nav-bar", Variant: "vertical", Selected: selected}}}
		if err := configureNavBarPreview(page, nil); err != nil {
			t.Fatal(err)
		}
		c := page.Cases[0]
		count := 0
		for i := range c.NavigationLabels() {
			if c.NavigationHasBadge(i) {
				count++
				if c.NavigationItem(i)["aria-current"] == "page" {
					t.Fatal("badge on viewed content")
				}
			}
		}
		if count != 1 {
			t.Fatal("missing default badge")
		}
		for _, control := range page.Controls {
			if control.Name == "text-size" {
				t.Fatal("text-size control should be removed")
			}
		}
	}
}

func TestTabsLayoutSwitchNormalizesCount(t *testing.T) {
	for _, layout := range []string{"fixed", "scrollable"} {
		count := "6"
		if layout == "scrollable" {
			count = "2"
		}
		p := &componentPreview{Cases: []componentCase{{Family: "tabs"}}}
		if err := configureTabsPreview(p, map[string][]string{"preview.layout": {layout}, "preview.count": {count}}); err != nil {
			t.Fatal(err)
		}
		if p.Cases[0].TabCount == count {
			t.Fatal("unsupported count retained")
		}
	}
}

func TestTabsExposeSelectionOnlyForStateComparison(t *testing.T) {
	var cases []componentCase
	for _, selected := range []bool{true, false} {
		for _, variant := range []string{"primary", "secondary"} {
			for _, state := range []string{"enabled", "hovered", "focused", "pressed"} {
				cases = append(cases, componentCase{Family: "tabs", Variant: variant, Control: "label", Selected: selected, State: state})
			}
		}
	}
	for _, comparison := range []string{"interactive", "variants", "states"} {
		page, err := selectComponentPreview(cases, map[string][]string{"preview.compare": {comparison}, "preview.selection": {"unselected"}})
		if err != nil {
			t.Fatal(err)
		}
		selection := false
		for _, control := range page.Controls {
			if control.Name == "viewport" {
				t.Fatal("tabs should not expose an artificial window selector")
			}
			if control.Name == "selection" {
				selection = true
				if control.Label != "First tab" || control.Labels["unselected"] != "Inactive" {
					t.Fatal("ambiguous selection control")
				}
			}
		}
		if selection != (comparison == "states") {
			t.Fatal("irrelevant selection control", comparison)
		}
		if comparison != "states" {
			for _, c := range page.Cases {
				if !c.Selected {
					t.Fatal("normal examples should start with first tab active")
				}
			}
		}
	}
}
