package model

import (
	"fmt"
	"github.com/a-h/templ"
)

type ChipProps struct {
	Label, Variant, Icon, Avatar            string
	Selected, Disabled, Elevated, Removable bool
	Attrs                                   templ.Attributes
}

func (p ChipProps) Attributes() templ.Attributes {
	v := Fallback(p.Variant, "filter")
	a := templ.Attributes{"class": "md-" + v + "-chip", "data-family": "chip", "data-variant": v, "data-selected": fmt.Sprint(p.Selected), "data-elevated": fmt.Sprint(p.Elevated), "data-shape": "label"}
	if p.Icon != "" {
		a["data-shape"] = "icon"
	}
	if p.Avatar != "" {
		a["data-shape"] = "avatar"
	}
	if v == "assist" || v == "suggestion" {
		a["type"] = "button"
		a["disabled"] = p.Disabled
		a["data-chip-action"] = ""
	}
	return Attributes(a, p.Attrs)
}
