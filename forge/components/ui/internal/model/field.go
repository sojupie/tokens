package model

import (
	"github.com/a-h/templ"
	"strconv"
)

type TextFieldProps struct {
	ID, Name, Label, Value, Type, Supporting, Placeholder string
	Filled, Disabled, Required, ReadOnly, Invalid         bool
	Leading, Trailing                                     templ.Component
	Attrs, InputAttrs                                     templ.Attributes
}

func (p TextFieldProps) InputResolved() templ.Attributes {
	a := templ.Attributes{"value": p.Value, "type": Fallback(p.Type, "text"), "placeholder": Fallback(p.Placeholder, " "), "disabled": p.Disabled, "required": p.Required, "readonly": p.ReadOnly, "aria-invalid": strconv.FormatBool(p.Invalid)}
	if p.ID != "" {
		a["id"] = p.ID
	}
	if p.Name != "" {
		a["name"] = p.Name
	}
	if p.Supporting != "" {
		a["aria-describedby"] = p.ID + "-help"
	}
	return Attributes(a, p.InputAttrs)
}
