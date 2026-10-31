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

// Demonstrate an unread destination without showing a stale badge on the open view.
func (c componentCase) NavigationHasBadge(index int) bool {
	if c.NavBadge == "" || c.NavBadge == "none" {
		return false
	}
	if c.Family == "nav-bar" {
		unread := 1
		if !c.Selected {
			unread = 2
		}
		return index == unread
	}
	return c.Family == "nav-rail" && index == 1 && c.Selected
}
