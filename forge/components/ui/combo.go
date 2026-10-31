package ui

import (
	"fmt"
	"github.com/a-h/templ"
)

// FieldProps adds affixes or a listbox to the same native text-field anatomy.
// Kind is text, select, or autocomplete. Options belong to the application.
type FieldProps struct {
	TextFieldProps
	Kind, Prefix, Suffix, PrefixDescription, SuffixDescription string
	Options                                                    []Option
	MenuIcons                                                  bool
	HelpAttrs                                                  templ.Attributes
}

func (p FieldProps) Attributes() templ.Attributes {
	v := "outlined"
	if p.Filled {
		v = "filled"
	}
	kind := fallback(p.Kind, "text-field")
	if kind == "text" {
		kind = "text-field"
	}
	return attributes(templ.Attributes{"class": "md-" + v + "-" + kind, "data-field-example": "", "data-variant": v, "data-populated": fmt.Sprint(p.Value != ""), "data-shape": map[bool]string{true: "icons", false: "label"}[p.MenuIcons]}, p.Attrs)
}
func (p FieldProps) FieldInput() templ.Attributes {
	base := p.TextFieldProps
	base.ID = p.ID + "-input"
	a := base.input()
	a["aria-describedby"] = p.ID + "-help"
	if v := p.InputAttrs["aria-describedby"]; v != nil {
		a["aria-describedby"] = v
	}
	if p.Kind == "select" || p.Kind == "autocomplete" {
		a["role"] = "combobox"
		a["aria-expanded"] = "false"
		a["aria-haspopup"] = "listbox"
		a["aria-controls"] = p.ID + "-options"
		a["data-field-combo"] = ""
		a["aria-autocomplete"] = "list"
		if p.Kind == "select" {
			a["readonly"] = true
			a["aria-autocomplete"] = "none"
		}
	}
	return a
}
