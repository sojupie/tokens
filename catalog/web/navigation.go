package web

import (
	"fmt"
	"github.com/a-h/templ"
	"net/url"
)

func configureNavigationPreview(page *componentPreview, query url.Values) error {
	for _, choice := range []previewControl{
		{Name: "layout", Label: "Expanded layout", Options: []string{"standard", "modal"}},
		{Name: "secondary", Label: "Additional destinations", Options: []string{"always", "expanded-only"}, Labels: map[string]string{"always": "Always visible", "expanded-only": "Expanded only"}},
		{Name: "alignment", Label: "Destination alignment", Options: []string{"top", "center"}},
		{Name: "header", Label: "Header", Options: []string{"menu-fab", "menu", "fab", "none"}},
		{Name: "indicator", Label: "Expanded indicator", Options: []string{"hug", "fill"}},
		{Name: "badge", Label: "Badge", Options: []string{"none", "small", "labeled"}},
	} {
		choice.Selected = query.Get("preview." + choice.Name)
		if choice.Selected == "" {
			choice.Selected = choice.Options[0]
		}
		if !containsPreview(choice.Options, choice.Selected) {
			return fmt.Errorf("%w: %s %s", errPreviewInput, choice.Name, choice.Selected)
		}
		page.Controls = append(page.Controls, choice)
		for i := range page.Cases {
			c := &page.Cases[i]
			switch choice.Name {
			case "layout":
				c.NavLayout = choice.Selected
			case "secondary":
				c.NavSecondary = choice.Selected
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
func (c componentCase) RailFrame() templ.Attributes {
	return templ.Attributes{"class": "md-rail-frame", "data-navigation-demo": "", "data-header": c.NavHeader, "data-alignment": c.NavAlign, "data-width": c.Width, "data-interactive": fmt.Sprint(c.State == "interactive")}
}
func (c componentCase) NavigationLabels() []string {
	if c.Family == "nav-drawer" {
		return []string{"Home", "Files", "Saved", "Recent", "Archive", "Trash", "Settings", "Help"}
	}
	if c.Family == "nav-rail" {
		return []string{"Home", "Files", "Saved", "Recent", "Archive", "Trash"}
	}
	labels := []string{"Home", "Files", "Saved", "Recent", "Settings"}
	count := 3
	if c.NavCount == "4" {
		count = 4
	}
	if c.NavCount == "5" {
		count = 5
	}
	return labels[:count]
}
func (c componentCase) NavigationIcon(index int) string {
	if c.Family == "nav-bar" {
		return []string{"home", "folder", "star", "schedule", "settings"}[index]
	}
	return []string{"home", "folder", "star", "schedule", "archive", "delete", "settings", "help"}[index]
}
func (c componentCase) RailMenu() bool {
	return c.NavHeader == "" || c.NavHeader == "menu" || c.NavHeader == "menu-fab"
}
func (c componentCase) RailFAB() bool { return c.NavHeader == "fab" || c.NavHeader == "menu-fab" }
func (c componentCase) NavigationFAB() templ.Attributes {
	class := "md-button md-fab md-nav-fab"
	if c.Variant == "expanded" {
		class = "md-button md-extended-fab md-nav-fab"
	}
	return templ.Attributes{"type": "button", "class": class, "data-size": "default", "data-variant": "primary-container", "data-navigation-create": "", "aria-label": "Create file"}
}

func configureNavBarPreview(page *componentPreview, query url.Values) error {
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
	for _, choice := range []previewControl{
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
		if !containsPreview(choice.Options, choice.Selected) {
			return fmt.Errorf("%w: %s %s", errPreviewInput, choice.Name, choice.Selected)
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
func (c componentCase) NavBarPreviewStyle() templ.SafeCSS {
	width := "100%"
	if c.NavViewport == "compact" {
		width = "412px"
	}
	if c.NavViewport == "medium" {
		width = "840px"
	}
	return templ.SafeCSS("--nav-preview-width:" + width)
}

func navBarPageItems(label string) []string {
	switch label {
	case "Files":
		return []string{"Installation plan.pdf · 2.4 MB", "Budget.xlsx · 128 KB", "Artwork references · 12 items", "Workshop notes.pdf · 1.1 MB"}
	case "Saved":
		return []string{"Exhibition guide", "Color reference collection", "Printing specifications", "Gallery floor plan"}
	case "Recent":
		return []string{"Budget updated · 9:24", "Installation plan opened · 8:42", "Artwork references saved · Yesterday", "Workshop notes shared · Tuesday"}
	case "Settings":
		return []string{"Notifications · On", "Offline access · 3 files", "Storage · 1.2 GB of 15 GB", "Language · English"}
	default:
		return []string{"Exhibition guide · Opened today", "Workshop notes · Opened yesterday", "Gallery floor plan · Opened Tuesday", "Print specifications · Opened Monday"}
	}
}
