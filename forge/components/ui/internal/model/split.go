package model

import (
	"github.com/a-h/templ"
	"strings"
)

type SplitButtonProps struct {
	ID, Label, Icon, Variant, Content string
	Size                              Size
	Disabled                          bool
	Attrs, PrimaryAttrs, MenuAttrs    templ.Attributes
}

func (p SplitButtonProps) Attributes() templ.Attributes {
	return Attributes(templ.Attributes{"class": "md-split-button", "role": "group", "aria-label": p.Label, "data-variant": Fallback(p.Variant, "filled"), "data-size": Fallback(string(p.Size), "small"), "data-shape": Fallback(p.Content, "icon-label")}, p.Attrs)
}
func (p SplitButtonProps) ActionResolved(menu bool) templ.Attributes {
	a := templ.Attributes{"class": "md-button md-" + Fallback(p.Variant, "filled") + "-button md-split-leading", "type": "button", "disabled": p.Disabled, "data-split-primary": ""}
	extra := p.PrimaryAttrs
	if menu {
		a = templ.Attributes{"class": "md-button md-" + Fallback(p.Variant, "filled") + "-button md-split-trailing", "type": "button", "disabled": p.Disabled, "data-split-toggle": "", "popovertarget": p.ID + "-menu", "aria-controls": p.ID + "-menu", "aria-haspopup": "menu", "aria-expanded": "false", "aria-label": "More " + strings.ToLower(p.Label) + " options"}
		extra = p.MenuAttrs
	}
	a["data-size"] = Fallback(string(p.Size), "small")
	a["data-shape"] = "round"
	return Attributes(a, extra)
}
