package ui

import (
	"github.com/a-h/templ"
	"strconv"
)

type ButtonVariant string

const (
	Filled   ButtonVariant = "filled"
	Tonal    ButtonVariant = "tonal"
	Outlined ButtonVariant = "outlined"
	Text     ButtonVariant = "text"
	Elevated ButtonVariant = "elevated"
)

type ButtonShape string

const (
	Round  ButtonShape = "round"
	Square ButtonShape = "square"
)

type ButtonProps struct {
	ID, Name, Value, Type string
	Variant               ButtonVariant
	Size                  Size
	Shape                 ButtonShape
	Disabled              bool
	Pressed               *bool
	Attrs                 templ.Attributes
}

func (p ButtonProps) Attributes() templ.Attributes {
	a := templ.Attributes{"class": "md-button md-" + fallback(string(p.Variant), "filled") + "-button", "type": fallback(p.Type, "button"), "data-size": fallback(string(p.Size), "small"), "data-shape": fallback(string(p.Shape), "round"), "disabled": p.Disabled}
	if p.ID != "" {
		a["id"] = p.ID
	}
	if p.Name != "" {
		a["name"] = p.Name
	}
	if p.Value != "" {
		a["value"] = p.Value
	}
	if p.Pressed != nil {
		a["aria-pressed"] = strconv.FormatBool(*p.Pressed)
		a["data-button-toggle"] = ""
	}
	return attributes(a, p.Attrs)
}
