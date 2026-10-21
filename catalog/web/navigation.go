package web

import (
	"fmt"
	"github.com/a-h/templ"
	"net/url"
)

func configureNavigationPreview(page *componentPreview, query url.Values) error {
	for _, choice := range []previewControl{
		{Name: "layout", Label: "Expanded layout", Options: []string{"standard", "modal"}},
		{Name: "collapse", Label: "When collapsed", Options: []string{"collapsed-rail", "hidden"}},
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
			case "collapse":
				c.NavCollapse = choice.Selected
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
	return templ.Attributes{"class": "md-rail-frame", "data-navigation-demo": "", "data-collapse": c.NavCollapse, "data-header": c.NavHeader, "data-alignment": c.NavAlign, "data-width": c.Width, "data-interactive": fmt.Sprint(c.State == "interactive")}
}
func (c componentCase) NavigationLabels() []string {
	if c.Family == "nav-drawer" {
		return []string{"Home", "Files", "Saved", "Recent", "Archive", "Trash", "Settings", "Help"}
	}
	if c.Family == "nav-rail" {
		return []string{"Home", "Files", "Saved", "Recent", "Archive", "Trash"}
	}
	return []string{"Home", "Files", "Saved"}
}
func (c componentCase) NavigationIcon(index int) string {
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
