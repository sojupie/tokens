package ui

import "github.com/a-h/templ"

type FABMenuProps struct {
	ID, Label, Variant             string
	Inline                         bool
	Close                          templ.Component
	Attrs, OpenerAttrs, PanelAttrs templ.Attributes
}

func (p FABMenuProps) Attributes() templ.Attributes {
	return attributes(templ.Attributes{"class": "md-fab-menu", "data-variant": fallback(p.Variant, "primary")}, p.Attrs)
}
func (p FABMenuProps) opener() templ.Attributes {
	return attributes(templ.Attributes{"type": "button", "class": "md-button md-fab md-fab-menu-opener", "data-variant": fallback(p.Variant, "primary") + "-container", "data-size": "default", "aria-label": p.Label, "aria-haspopup": "menu", "aria-expanded": "false", "popovertarget": p.ID + "-menu"}, p.OpenerAttrs)
}
func (p FABMenuProps) panel() templ.Attributes {
	a := templ.Attributes{"class": "md-fab-menu-panel", "id": p.ID + "-menu"}
	if !p.Inline {
		a["popover"] = "auto"
	}
	return attributes(a, p.PanelAttrs)
}

type FABMenuItemProps struct {
	Label, Icon string
	Attrs       templ.Attributes
}
