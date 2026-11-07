package model

import (
	"fmt"
	"github.com/a-h/templ"
)

type GroupProps struct {
	Label, Variant, Selection string
	Size                      Size
	Connected                 bool
	Attrs                     templ.Attributes
}

func (p GroupProps) Attributes() templ.Attributes {
	control := "standard"
	if p.Connected {
		control = "connected"
	}
	return Attributes(templ.Attributes{"class": "md-button-group", "role": "group", "aria-label": p.Label, "data-control": control, "data-variant": Fallback(p.Variant, "filled"), "data-size": Fallback(string(p.Size), "small"), "data-shape": Fallback(p.Selection, "single")}, p.Attrs)
}

type GroupItemProps struct {
	Label, Icon                   string
	Index                         int
	Selected, Disabled, Connected bool
	Variant                       ButtonVariant
	Size                          Size
	Attrs                         templ.Attributes
}

func (p GroupItemProps) Attributes() templ.Attributes {
	a := (ButtonProps{Variant: p.Variant, Size: p.Size, Disabled: p.Disabled}).Attributes()
	a["data-group-button"] = fmt.Sprint(p.Index)
	a["tabindex"] = "-1"
	if p.Index == 0 {
		a["tabindex"] = "0"
	}
	if p.Connected {
		a["aria-pressed"] = fmt.Sprint(p.Selected)
		a["data-button-toggle"] = ""
	}
	return Attributes(a, p.Attrs)
}
