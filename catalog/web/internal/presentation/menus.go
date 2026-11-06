package presentation

import (
	"fmt"

	"github.com/a-h/templ"
)

func (c ComponentCase) MenuPanel(sub bool) templ.Attributes {
	a := templ.Attributes{"class": "md-menu", "data-menu-panel": "", "data-variant": c.Variant, "data-orientation": c.Orientation, "data-content": c.Shape, "role": "menu", "aria-label": "Example actions", "data-menu-selection": c.MenuSelectionMode()}
	if sub {
		a["id"] = c.ID() + "-submenu"
		a["data-orientation"] = "vertical"
		a["aria-label"] = "More actions"
		if c.State == "interactive" {
			a["popover"] = "auto"
		} else {
			if !c.Expanded {
				a["hidden"] = true
			}
			a["data-menu-static"] = ""
		}
	} else {
		a["data-menu-main"] = ""
		if c.State == "interactive" {
			a["popover"] = "auto"
			a["id"] = c.ID() + "-menu"
		}
	}
	if c.Shape == "cascade" && c.Expanded {
		if sub {
			a["data-menu-activity"] = "active"
		} else {
			a["data-menu-activity"] = "inactive"
		}
	}
	return a
}
func (c ComponentCase) MenuItem(index int) templ.Attributes {
	a := templ.Attributes{"type": "button", "class": "md-menu-item", "role": "menuitem", "tabindex": "-1", "data-menu-action": ""}
	if index == 0 {
		a = c.Attributes()
		delete(a, "aria-current")
		delete(a, "aria-disabled")
		a["type"] = "button"
		a["data-menu-action"] = ""
		if c.State == "disabled" {
			a["aria-disabled"] = "true"
		}
		a["tabindex"] = "0"
		if c.Shape == "cascade" {
			a["role"] = "menuitem"
			delete(a, "aria-checked")
			delete(a, "data-menu-action")
			a["data-menu-submenu"] = c.ID() + "-submenu"
			a["aria-controls"] = c.ID() + "-submenu"
			a["aria-haspopup"] = "menu"
			a["aria-expanded"] = fmt.Sprint(c.Expanded && c.State != "interactive")
		}
	}
	if c.MenuSelectable(index) {
		a["role"] = "menuitemcheckbox"
		if c.MenuSelectionMode() == "single" {
			a["role"] = "menuitemradio"
		}
		a["aria-checked"] = fmt.Sprint(c.MenuSelected(index))
		a["data-selected"] = fmt.Sprint(c.MenuSelected(index))
	} else {
		delete(a, "data-selected")
	}
	if c.Shape == "icon-only" {
		a["class"] = "md-menu-item md-menu-icon-button"
		a["aria-label"] = []string{"Favorite", "Download", "Edit"}[index]
	}
	return a
}

func (c ComponentCase) MenuSelectionMode() string {
	if c.MenuMode == "" {
		return "actions"
	}
	return c.MenuMode
}
func (c ComponentCase) MenuSelectable(index int) bool {
	return c.Shape != "cascade" && (c.MenuSelectionMode() != "actions" || index == 0)
}
func (c ComponentCase) MenuSelected(index int) bool { return c.Selected && index == 0 }
