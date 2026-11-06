// Integration checks for playground.
package server

import (
	"bytes"
	"context"
	"errors"
	"github.com/sojupie/tokens/catalog/web/internal/playground"
	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	webrender "github.com/sojupie/tokens/catalog/web/internal/render"
	"net/url"
	"strings"
	"testing"
)

func TestSliderPreviewConfigurations(t *testing.T) {
	cases := []webview.ComponentCase{
		{Family: "slider", Variant: "continuous", Control: "single", Size: "medium", Shape: "label", State: "enabled"},
		{Family: "slider", Variant: "discrete", Control: "single", Size: "medium", Shape: "label", State: "enabled"},
	}
	page, err := playground.SelectComponentPreview(cases, url.Values{"preview.icon": {"volume"}, "preview.value": {"30"}, "preview.step": {"20"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, control := range page.Controls {
		if control.Name == "shape" {
			t.Fatal("token families exposed as bubble variants")
		}
	}
	for _, c := range page.Cases {
		var out bytes.Buffer
		if err := webrender.SpecimenSlider(c).Render(context.Background(), &out); err != nil {
			t.Fatal(err)
		}
		markup := out.String()
		if strings.Contains(markup, `md-slider-trailing`) || strings.Contains(markup, `@materialSymbol`) {
			t.Fatal("unexpected endpoint icon")
		}
		if c.Variant == "continuous" {
			if c.SliderInsetIcon() != "volume_up" || !strings.Contains(markup, `data-slider-inset`) {
				t.Fatal("missing meaningful single inset icon")
			}
		} else {
			if c.SliderInsetIcon() != "" || c.SliderValue("end") != "40" || len(c.SliderTicks()) != 4 || c.SliderInput("end")["step"] != "20" {
				t.Fatalf("invalid discrete configuration: %#v", c)
			}
		}
		if c.SliderDisplay != "always" {
			t.Fatal("preview should reveal the value label by default")
		}
	}
	_, err = playground.SelectComponentPreview(cases, url.Values{"preview.step": {"0"}})
	if !errors.Is(err, playground.ErrPreviewInput) {
		t.Fatalf("invalid step accepted: %v", err)
	}
}
func TestSliderInsetIconRestrictions(t *testing.T) {
	for _, c := range []webview.ComponentCase{
		{Variant: "continuous", Control: "range", Size: "medium", SliderIcon: "volume"},
		{Variant: "continuous", Control: "single", Size: "small", SliderIcon: "volume"},
		{Variant: "discrete", Control: "single", Size: "large", SliderIcon: "volume"},
	} {
		if c.SliderInsetIcon() != "" {
			t.Fatalf("unsupported inset icon: %#v", c)
		}
	}
	page := webview.ComponentPreview{Cases: []webview.ComponentCase{{Variant: "continuous", Control: "range", Size: "large"}}}
	if err := playground.ConfigureSliderPreview(&page, url.Values{"preview.lower": {"90"}, "preview.value": {"100"}}); err != nil {
		t.Fatal(err)
	}
	if page.Cases[0].SliderValue("start") != "90" || page.Cases[0].SliderValue("end") != "100" {
		t.Fatal("numeric range corrupted")
	}
	for _, c := range page.Controls {
		if c.Name == "icon" {
			t.Fatal("inset control offered for range slider")
		}
	}
}

func TestTabsLayoutsAndBadgeOwnership(t *testing.T) {
	for _, variant := range []string{"primary", "secondary"} {
		for _, control := range []string{"label", "icon", "icon-label"} {
			page := &webview.ComponentPreview{Cases: []webview.ComponentCase{{Family: "tabs", Variant: variant, Control: control, State: "interactive", Selected: true}}}
			err := playground.ConfigureTabsPreview(page, map[string][]string{"preview.layout": {"scrollable"}, "preview.badge": {"labeled"}, "preview.direction": {"rtl"}})
			if err != nil {
				t.Fatal(err)
			}
			c := page.Cases[0]
			var out bytes.Buffer
			if err := webrender.SpecimenTabs(c).Render(context.Background(), &out); err != nil {
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
		if playground.ConfigureTabsPreview(&webview.ComponentPreview{}, map[string][]string{"preview.count": {count}}) == nil {
			t.Fatal("invalid fixed count", count)
		}
	}
}
func TestNavigationBadgeSurvivesEitherInitialSelection(t *testing.T) {
	for _, selected := range []bool{true, false} {
		page := &webview.ComponentPreview{Cases: []webview.ComponentCase{{Family: "nav-bar", Variant: "vertical", Selected: selected}}}
		if err := playground.ConfigureNavBarPreview(page, nil); err != nil {
			t.Fatal(err)
		}
		c := page.Cases[0]
		count := 0
		for i := range c.NavigationLabels() {
			if c.NavigationHasBadge(i) {
				count++
				if i != 1 {
					t.Fatal("selection changed the badge destination")
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
		p := &webview.ComponentPreview{Cases: []webview.ComponentCase{{Family: "tabs"}}}
		if err := playground.ConfigureTabsPreview(p, map[string][]string{"preview.layout": {layout}, "preview.count": {count}}); err != nil {
			t.Fatal(err)
		}
		if p.Cases[0].TabCount == count {
			t.Fatal("unsupported count retained")
		}
	}
}

func TestTabsExposeSelectionOnlyForStateComparison(t *testing.T) {
	var cases []webview.ComponentCase
	for _, selected := range []bool{true, false} {
		for _, variant := range []string{"primary", "secondary"} {
			for _, state := range []string{"enabled", "hovered", "focused", "pressed"} {
				cases = append(cases, webview.ComponentCase{Family: "tabs", Variant: variant, Control: "label", Selected: selected, State: state})
			}
		}
	}
	for _, comparison := range []string{"interactive", "variants", "states"} {
		page, err := playground.SelectComponentPreview(cases, map[string][]string{"preview.compare": {comparison}, "preview.selection": {"unselected"}})
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

func TestToolbarPreviewTreatsFABAsConfigurationAndPreservesOldLinks(t *testing.T) {
	cases := []webview.ComponentCase{
		{Family: "toolbar", Variant: "docked", Control: "standard", Orientation: "horizontal", Size: "none", Shape: "icon", State: "enabled"},
		{Family: "toolbar", Variant: "floating", Control: "vibrant", Orientation: "vertical", Size: "medium", Shape: "icon", State: "enabled"},
	}
	query := url.Values{"preview.compare": {"interactive"}, "preview.variant": {"floating-fab"}, "preview.control": {"vibrant"}, "preview.orientation": {"vertical"}, "preview.size": {"medium"}}
	page, err := playground.SelectComponentPreview(cases, query)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Cases) != 1 || page.Cases[0].Variant != "floating" || page.Cases[0].Size != "medium" {
		t.Fatalf("wrong FAB composition: %#v", page)
	}
	if query.Get("preview.variant") != "floating-fab" {
		t.Fatal("modified caller query")
	}
}

func TestToolbarComparisonsUseCompatibleContexts(t *testing.T) {
	cases := []webview.ComponentCase{
		{Family: "toolbar", Variant: "docked", Control: "standard", Orientation: "horizontal", Size: "none", Shape: "icon", State: "enabled"},
		{Family: "toolbar", Variant: "floating", Control: "standard", Orientation: "horizontal", Size: "none", Shape: "icon", State: "enabled"},
		{Family: "toolbar", Variant: "floating", Control: "standard", Orientation: "vertical", Size: "medium", Shape: "icon", State: "enabled"},
	}
	page, err := playground.SelectComponentPreview(cases, url.Values{"preview.compare": {"variants"}, "preview.orientation": {"vertical"}, "preview.size": {"medium"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Cases) != 2 {
		t.Fatalf("variant comparison lost a variant: %#v", page.Cases)
	}
	for _, control := range page.Controls {
		if control.Name == "orientation" || control.Name == "size" {
			t.Fatal("incompatible comparison filter exposed")
		}
	}
}

func TestChipActionsAndVariants(t *testing.T) {
	for _, variant := range []string{"filter", "assist", "suggestion", "input"} {
		c := webview.ComponentCase{Family: "chip", Variant: variant, State: "disabled", Shape: "icon", Width: "removable", Selected: true}
		a := c.Attributes()
		if a["class"] != "md-"+variant+"-chip" || a["aria-disabled"] != "true" {
			t.Fatalf("missing native chip class or disabled state: %v", a)
		}
		d := webview.PreviewDimensions(c)
		if variant == "assist" || variant == "suggestion" {
			if _, ok := d["selection"]; ok {
				t.Fatal("action chips must not expose selection")
			}
			if a["disabled"] != true {
				t.Fatal("action chips need native disabled buttons")
			}
		}
		if variant == "input" {
			if _, ok := d["elevation"]; ok {
				t.Fatal("input chips have no elevated profile")
			}
		}
	}
	cases := []webview.ComponentCase{{Family: "chip", Variant: "filter", Shape: "icon", Width: "plain", State: "enabled"}, {Family: "chip", Variant: "assist", Shape: "icon", Width: "plain", State: "enabled"}, {Family: "chip", Variant: "suggestion", Shape: "icon", Width: "plain", State: "enabled"}, {Family: "chip", Variant: "input", Shape: "icon", Width: "removable", State: "enabled"}}
	preview, err := playground.SelectComponentPreview(cases, map[string][]string{"preview.compare": {"variants"}})
	if err != nil || len(preview.Cases) != 4 {
		t.Fatalf("variant comparison omitted a chip type: %v %v", preview, err)
	}
}
