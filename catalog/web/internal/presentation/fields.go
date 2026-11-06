package presentation

import (
	"fmt"

	"github.com/a-h/templ"
)

var FieldOptions = []string{"Option one", "Option two", "Option three"}

func (c ComponentCase) FieldLabel() string {
	if c.IllustrationLabels[0] != "" {
		return c.IllustrationLabels[0]
	}
	if c.Shape == "affixes" {
		return "Price per kilogram"
	}
	if c.Family == "select" {
		return "Choose a value"
	}
	if c.Family == "autocomplete" {
		return "Find a value"
	}
	return "Label"
}
func (c ComponentCase) FieldSupporting() string {
	if c.IllustrationLabels[1] != "" {
		return c.IllustrationLabels[1]
	}
	return "Supporting text"
}
func (c ComponentCase) FieldInput() templ.Attributes {
	value := ""
	if c.Width == "populated" {
		value = "Input text"
		if c.Family != "field" {
			value = FieldOptions[1]
		}
	}
	a := templ.Attributes{"id": c.ID() + "-input", "type": "text", "value": value, "placeholder": "Input text", "aria-describedby": c.ID() + "-help", "aria-invalid": fmt.Sprint(c.Invalid), "disabled": c.State == "disabled", "autocomplete": "off"}
	if c.Shape == "affixes" {
		a["aria-describedby"] = c.ID() + "-prefix " + c.ID() + "-suffix " + c.ID() + "-help"
	}
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

func (c ComponentCase) FieldHelpAttributes() templ.Attributes {
	a := templ.Attributes{}
	if c.Invalid {
		a["role"] = "alert"
	}
	return a
}
