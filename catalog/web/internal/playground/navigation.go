package playground

import (
	"fmt"
	"net/url"

	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
)

func ConfigureNavigationPreview(page *webview.ComponentPreview, query url.Values) error {
	initial := ""
	if page.Interactive {
		initial = query.Get("preview.rail-state")
		if initial == "" {
			initial = page.Cases[0].Variant
			if query.Get("preview.collapse") == "hidden" {
				initial = "hidden"
			}
		}
		if !ContainsPreview([]string{"collapsed", "expanded", "hidden"}, initial) {
			return fmt.Errorf("%w: rail state %s", ErrPreviewInput, initial)
		}
		controls := page.Controls[:0]
		for _, control := range page.Controls {
			if control.Name != "variant" {
				controls = append(controls, control)
			}
		}
		page.Controls = append(controls, webview.PreviewControl{Name: "rail-state", Label: "Initial state", Options: []string{"collapsed", "expanded", "hidden"}, Selected: initial})
		for i := range page.Cases {
			page.Cases[i].Variant = initial
			if initial == "hidden" {
				page.Cases[i].Variant = "collapsed"
				page.Cases[i].NavCollapse = "hidden"
			}
		}
	}
	for _, choice := range []webview.PreviewControl{
		{Name: "layout", Label: "Expanded layout", Options: []string{"standard", "modal"}},
		{Name: "alignment", Label: "Destination alignment", Options: []string{"top", "center"}},
		{Name: "header", Label: "Header", Options: []string{"menu-fab", "menu", "fab", "none"}},
		{Name: "indicator", Label: "Expanded indicator", Options: []string{"hug", "fill"}},
		{Name: "badge", Label: "Badge", Options: []string{"none", "small", "labeled"}},
	} {
		if choice.Name == "header" && initial == "hidden" {
			choice.Options = []string{"menu-fab", "menu"}
		}
		choice.Selected = query.Get("preview." + choice.Name)
		if choice.Selected == "" {
			choice.Selected = choice.Options[0]
		}
		if choice.Name == "header" && initial == "hidden" && (choice.Selected == "none" || choice.Selected == "fab") {
			choice.Selected = "menu-fab"
		}
		if !ContainsPreview(choice.Options, choice.Selected) {
			return fmt.Errorf("%w: %s %s", ErrPreviewInput, choice.Name, choice.Selected)
		}
		page.Controls = append(page.Controls, choice)
		for i := range page.Cases {
			c := &page.Cases[i]
			switch choice.Name {
			case "layout":
				c.NavLayout = choice.Selected
			case "alignment":
				c.NavAlign = choice.Selected
			case "header":
				c.NavHeader = choice.Selected
			case "indicator":
				c.NavIndicator = choice.Selected
			case "badge":
				c.NavBadge = choice.Selected
			}
		}
	}
	for i := range page.Cases {
		page.Cases[i].Orientation = "vertical"
		if page.Cases[i].Variant == "expanded" {
			page.Cases[i].Orientation = "horizontal"
		}
	}
	return nil
}

func ConfigureNavBarPreview(page *webview.ComponentPreview, query url.Values) error {
	allHorizontal := len(page.Cases) > 0
	hasHorizontal := false
	for _, c := range page.Cases {
		allHorizontal = allHorizontal && c.Variant == "horizontal"
		hasHorizontal = hasHorizontal || c.Variant == "horizontal" || c.Variant == "adaptive"
	}
	for i := range page.Controls {
		if page.Controls[i].Name == "variant" {
			page.Controls[i].Labels = map[string]string{"adaptive": "Flexible: automatic", "vertical": "Flexible: vertical", "horizontal": "Flexible: horizontal", "baseline": "Baseline (not recommended)"}
		}
	}
	for _, choice := range []webview.PreviewControl{
		{Name: "destinations", Label: "Destinations", Options: []string{"3", "4", "5"}},
		{Name: "viewport", Label: "Window", Options: []string{"compact", "medium", "available"}, Labels: map[string]string{"available": "Available width", "compact": "Compact (412px)", "medium": "Medium (840px)"}},
		{Name: "badge", Label: "Badge", Options: []string{"labeled", "small", "none"}},
		{Name: "alignment", Label: "Horizontal alignment", Options: []string{"center", "start", "end"}},
		{Name: "direction", Label: "Direction", Options: []string{"ltr", "rtl"}},
	} {
		if choice.Name == "alignment" && !hasHorizontal {
			continue
		}
		choice.Selected = query.Get("preview." + choice.Name)
		if choice.Selected == "" {
			choice.Selected = choice.Options[0]
		}
		if choice.Name == "viewport" && allHorizontal {
			choice.Options = []string{"available", "medium"}
			if choice.Selected == "compact" {
				choice.Selected = "medium"
			}
		}
		if !ContainsPreview(choice.Options, choice.Selected) {
			return fmt.Errorf("%w: %s %s", ErrPreviewInput, choice.Name, choice.Selected)
		}
		page.Controls = append(page.Controls, choice)
		for i := range page.Cases {
			c := &page.Cases[i]
			switch choice.Name {
			case "destinations":
				c.NavCount = choice.Selected
			case "viewport":
				c.NavViewport = choice.Selected
			case "badge":
				c.NavBadge = choice.Selected
			case "alignment":
				c.NavBarAlignment = choice.Selected
			case "direction":
				c.NavDirection = choice.Selected
			}
		}
	}
	if len(page.Cases) > 1 && page.Cases[0].NavViewport == "compact" && !allHorizontal {
		cases := page.Cases[:0]
		for _, c := range page.Cases {
			if c.Variant != "horizontal" {
				cases = append(cases, c)
			}
		}
		page.Cases = cases
	}

	return nil
}
