package model

import (
	"fmt"
	"github.com/a-h/templ"
)

type IconButtonProps struct {
	Label, Icon, Variant, Width string
	Size                        Size
	Shape                       ButtonShape
	Pressed                     *bool
	Disabled                    bool
	Attrs                       templ.Attributes
}

func (p IconButtonProps) Attributes() templ.Attributes {
	a := templ.Attributes{"class": "md-icon-button md-" + Fallback(p.Variant, "standard") + "-icon-button", "type": "button", "aria-label": p.Label, "data-size": Fallback(string(p.Size), "small"), "data-shape": Fallback(string(p.Shape), "round"), "data-width": Fallback(p.Width, "default"), "disabled": p.Disabled}
	if p.Pressed != nil {
		a["aria-pressed"] = fmt.Sprint(*p.Pressed)
		a["data-button-toggle"] = ""
	}
	return Attributes(a, p.Attrs)
}
