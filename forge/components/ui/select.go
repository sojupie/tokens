package ui

import "github.com/a-h/templ"

type SelectProps struct {
	ID, Name, Label, Value        string
	Outlined, HideLabel, Disabled bool
	Options                       []Option
	Attrs, InputAttrs, MenuAttrs  templ.Attributes
}

func (p SelectProps) displayLabel() string {
	for _, option := range p.Options {
		if option.Value == p.Value {
			return option.Label
		}
	}
	return p.Value
}
func (p SelectProps) Attributes() templ.Attributes {
	class := "md-chip-select"
	if p.Outlined {
		class = "md-outlined-select"
	}
	return attributes(templ.Attributes{"class": class, "data-field-example": "", "data-populated": "true"}, p.Attrs)
}
func (p SelectProps) InputAttributes() templ.Attributes {
	a := templ.Attributes{"id": p.ID, "role": "combobox", "aria-label": p.Label, "aria-expanded": "false", "aria-haspopup": "listbox", "aria-controls": p.ID + "-options", "aria-autocomplete": "none", "data-field-combo": "", "value": p.displayLabel(), "disabled": p.Disabled}
	if p.Outlined {
		a["type"] = "text"
		a["readonly"] = true
		a["placeholder"] = " "
	} else {
		a["type"] = "button"
		a["class"] = "md-filter-chip"
		a["data-field-select"] = ""
	}
	return attributes(a, p.InputAttrs)
}
