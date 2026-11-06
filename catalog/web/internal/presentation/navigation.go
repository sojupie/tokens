package presentation

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/sojupie/tokens/forge/components/ui"
)

func (c ComponentCase) RailFrame() templ.Attributes {
	return templ.Attributes{"class": "md-rail-frame", "data-navigation-demo": "", "data-collapse": c.NavCollapse, "data-header": c.NavHeader, "data-alignment": c.NavAlign, "data-width": c.Width, "data-interactive": fmt.Sprint(c.State == "interactive")}
}
func (c ComponentCase) NavigationLabels() []string {
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
func (c ComponentCase) NavigationIcon(index int) string {
	if c.Family == "nav-bar" {
		return []string{"home", "folder", "star", "schedule", "settings"}[index]
	}
	return []string{"home", "folder", "star", "schedule", "archive", "delete", "settings", "help"}[index]
}
func (c ComponentCase) RailMenu() bool {
	return c.NavHeader == "" || c.NavHeader == "menu" || c.NavHeader == "menu-fab"
}
func (c ComponentCase) RailFAB() bool { return c.NavHeader == "fab" || c.NavHeader == "menu-fab" }
func (c ComponentCase) RailProps() ui.RailProps {
	p := ui.RailProps{NavigationProps: ui.NavigationProps{Label: "Navigation preview", Size: ui.Size(c.Size), Variant: c.Variant, Width: c.Width, Layout: c.NavLayout, Indicator: c.NavIndicator, Alignment: c.NavAlign, Attrs: c.Attributes()}, Menu: c.RailMenu(), PersistentMenu: c.NavCollapse == "hidden"}
	if c.RailFAB() {
		p.FAB = &ui.FABProps{Label: "Create", Icon: "edit", Attrs: templ.Attributes{"data-navigation-create": "", "aria-label": "Create file"}}
	}
	return p
}

func (c ComponentCase) NavBarPreviewStyle() templ.SafeCSS {
	width := "100%"
	if c.NavViewport == "compact" {
		width = "412px"
	}
	if c.NavViewport == "medium" {
		width = "840px"
	}
	return templ.SafeCSS("--nav-preview-width:" + width)
}

func NavBarPageItems(label string) []string {
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
