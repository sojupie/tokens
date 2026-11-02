package web

import (
	"github.com/a-h/templ"
)

func (c componentCase) NavigationItem(index int) templ.Attributes {
	attrs := templ.Attributes{"class": "md-nav-item", "type": "button", "data-navigation-item": "", "aria-controls": c.ID() + "-selection"}
	if c.Family == "nav-rail" || (c.Family == "nav-bar" && c.State == "interactive") {
		attrs["aria-controls"] = c.ID() + "-screen"
	}
	if (index == 0 && c.Selected) || (index == 1 && !c.Selected) {
		attrs["aria-current"] = "page"
	}
	if c.State != "interactive" {
		attrs["data-component-state"] = "enabled"
		if index == 0 {
			attrs["data-component-state"] = c.State
		}
	}
	if c.NavigationHasBadge(index) {
		attrs["aria-label"] = c.NavigationLabels()[index] + " " + c.NavigationBadgeLabel()
	}
	return attrs
}

func (c componentCase) NavigationToggleLabel() string {
	if c.Variant == "expanded" {
		return "Collapse navigation"
	}
	return "Expand navigation"
}

// Badge presence is independent of which destination is selected.
func (c componentCase) NavigationHasBadge(index int) bool {
	return (c.Family == "nav-bar" || c.Family == "nav-rail") && c.NavBadge != "" && c.NavBadge != "none" && index == 1
}
