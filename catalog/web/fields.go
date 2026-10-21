package web

import (
	"fmt"
	"github.com/a-h/templ"
)

var fieldOptions = []string{"Option one", "Option two", "Option three"}

func (c componentCase) FieldLabel() string {
	if c.Family == "select" {
		return "Choose a value"
	}
	if c.Family == "autocomplete" {
		return "Find a value"
	}
	return "Label"
}
func (c componentCase) FieldInput() templ.Attributes {
	value := ""
	if c.Width == "populated" {
		value = "Input text"
		if c.Family != "field" {
			value = fieldOptions[1]
		}
	}
	a := templ.Attributes{"id": c.ID() + "-input", "type": "text", "value": value, "placeholder": "Input text", "aria-describedby": c.ID() + "-help", "aria-invalid": fmt.Sprint(c.Invalid), "disabled": c.State == "disabled", "autocomplete": "off"}
	if c.Family != "field" {
		a["placeholder"] = " "
		a["role"] = "combobox"
		a["aria-expanded"] = "false"
		a["aria-haspopup"] = "listbox"
		a["aria-controls"] = c.ID() + "-options"
		a["data-field-combo"] = ""
		a["aria-autocomplete"] = "list"
		if c.Family == "select" {
			a["readonly"] = true
			a["aria-autocomplete"] = "none"
		}
	}
	return a
}
