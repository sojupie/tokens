package ui

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
	return attributes(templ.Attributes{"class": "md-split-button", "role": "group", "aria-label": p.Label, "data-variant": fallback(p.Variant, "filled"), "data-size": fallback(string(p.Size), "small"), "data-shape": fallback(p.Content, "icon-label")}, p.Attrs)
}
func (p SplitButtonProps) action(menu bool) templ.Attributes {
	a := templ.Attributes{"class": "md-button md-" + fallback(p.Variant, "filled") + "-button md-split-leading", "type": "button", "disabled": p.Disabled, "data-split-primary": ""}
	extra := p.PrimaryAttrs
	if menu {
		a = templ.Attributes{"class": "md-button md-" + fallback(p.Variant, "filled") + "-button md-split-trailing", "type": "button", "disabled": p.Disabled, "data-split-toggle": "", "popovertarget": p.ID + "-menu", "aria-controls": p.ID + "-menu", "aria-haspopup": "menu", "aria-expanded": "false", "aria-label": "More " + strings.ToLower(p.Label) + " options"}
		extra = p.MenuAttrs
	}
	a["data-size"] = fallback(string(p.Size), "small")
	a["data-shape"] = "round"
	return attributes(a, extra)
}
