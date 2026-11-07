package model

import "github.com/a-h/templ"

type FABMenuProps struct {
	ID, Label, Variant             string
	Inline                         bool
	Close                          templ.Component
	Attrs, OpenerAttrs, PanelAttrs templ.Attributes
}

func (p FABMenuProps) Attributes() templ.Attributes {
	return Attributes(templ.Attributes{"class": "md-fab-menu", "data-variant": Fallback(p.Variant, "primary")}, p.Attrs)
}
func (p FABMenuProps) OpenerResolved() templ.Attributes {
	return Attributes(templ.Attributes{"type": "button", "class": "md-button md-fab md-fab-menu-opener", "data-variant": Fallback(p.Variant, "primary") + "-container", "data-size": "default", "aria-label": p.Label, "aria-haspopup": "menu", "aria-expanded": "false", "popovertarget": p.ID + "-menu"}, p.OpenerAttrs)
}
func (p FABMenuProps) PanelResolved() templ.Attributes {
	a := templ.Attributes{"class": "md-fab-menu-panel", "id": p.ID + "-menu"}
	if !p.Inline {
		a["popover"] = "auto"
	}
	return Attributes(a, p.PanelAttrs)
}

type FABMenuItemProps struct {
	Label, Icon string
	Attrs       templ.Attributes
}
