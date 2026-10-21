package web

import (
	"fmt"
	"github.com/a-h/templ"
	"net/url"
)

func (c componentCase) MenuPanel(sub bool) templ.Attributes {
	a := templ.Attributes{"class": "md-menu", "data-menu-panel": "", "data-variant": c.Variant, "data-orientation": c.Orientation, "data-content": c.Shape, "role": "menu", "aria-label": "Example actions", "data-menu-selection": c.menuMode()}
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
func (c componentCase) MenuItem(index int) templ.Attributes {
	a := templ.Attributes{"type": "button", "class": "md-menu-item", "role": "menuitem", "tabindex": "-1", "data-menu-action": ""}
	if index == 0 {
		a = c.Attributes()
		delete(a, "aria-current")
		delete(a, "aria-disabled")
		a["type"] = "button"
		a["data-menu-action"] = ""
		if c.State == "disabled" {
			a["disabled"] = true
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
		if c.menuMode() == "single" {
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

func (c componentCase) menuMode() string {
	if c.MenuMode == "" {
		return "actions"
	}
	return c.MenuMode
}
func (c componentCase) MenuSelectable(index int) bool {
	return c.Shape != "cascade" && (c.menuMode() != "actions" || index == 0)
}
func (c componentCase) MenuSelected(index int) bool { return c.Selected && index == 0 }
func configureMenuPreview(page *componentPreview, query url.Values) error {
	if page.Cases[0].Shape == "cascade" {
		return nil
	}
	mode := query.Get("preview.behavior")
	if mode == "" {
		mode = "actions"
	}
	if !containsPreview([]string{"actions", "single", "multiple"}, mode) {
		return fmt.Errorf("%w: menu behavior %s", errPreviewInput, mode)
	}
	page.Controls = append(page.Controls, previewControl{Name: "behavior", Label: "Behavior", Selected: mode, Options: []string{"actions", "single", "multiple"}, Labels: map[string]string{"actions": "Actions", "single": "Single select", "multiple": "Multi select"}})
	for i := range page.Cases {
		page.Cases[i].MenuMode = mode
	}
	return nil
}
