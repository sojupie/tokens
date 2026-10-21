package web

import "github.com/a-h/templ"

func (c componentCase) FABMenuOpener() templ.Attributes {
	return templ.Attributes{"type": "button", "class": "md-button md-fab md-fab-menu-opener", "data-variant": c.Variant + "-container", "data-size": "default", "aria-label": "Create", "aria-haspopup": "menu", "aria-expanded": "false", "popovertarget": c.ID() + "-menu"}
}

func (c componentCase) FABMenuPanel() templ.Attributes {
	attrs := templ.Attributes{"class": "md-fab-menu-panel", "id": c.ID() + "-menu"}
	if c.State == "interactive" {
		attrs["popover"] = "auto"
	}
	return attrs
}

func (c componentCase) FABMenuAction(index int) templ.Attributes {
	class := "md-button md-fab-menu-item"
	if index < 0 {
		class = "md-button md-fab-menu-close"
	}
	attrs := templ.Attributes{"class": class, "type": "button", "data-size": "small"}
	if index < 0 {
		attrs["aria-label"] = "Close menu"
		if c.State == "interactive" {
			attrs["popovertarget"] = c.ID() + "-menu"
			attrs["popovertargetaction"] = "hide"
		}
	} else {
		attrs["role"] = "menuitem"
		attrs["data-fab-menu-action"] = ""
		if c.State == "interactive" {
			attrs["tabindex"] = "-1"
		}
	}
	if c.State != "interactive" {
		state := "enabled"
		if c.Control == "items" && index == 0 || c.Control == "close" && index < 0 {
			state = c.State
		}
		attrs["data-component-state"] = state
		attrs["data-button-state"] = state
	}
	return attrs
}
