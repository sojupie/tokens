package model

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

type ColorFamily string

const (
	ColorPrimary   ColorFamily = "primary"
	ColorSecondary ColorFamily = "secondary"
	ColorTertiary  ColorFamily = "tertiary"
	ColorError     ColorFamily = "error"
)

const (
	Round  ButtonShape = "round"
	Square ButtonShape = "square"
)

type ButtonProps struct {
	ID, Name, Value, Type string
	Variant               ButtonVariant
	Size                  Size
	Shape                 ButtonShape
	// Color selects a system color family; empty preserves the source mapping.
	Color    ColorFamily
	Disabled bool
	Pressed  *bool
	Attrs    templ.Attributes
}

func (p ButtonProps) Attributes() templ.Attributes {
	a := templ.Attributes{"class": "md-button md-" + Fallback(string(p.Variant), "filled") + "-button", "type": Fallback(p.Type, "button"), "data-size": Fallback(string(p.Size), "small"), "data-shape": Fallback(string(p.Shape), "round"), "disabled": p.Disabled}
	if p.Color != "" {
		a["data-color"] = string(p.Color)
	}
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
	return Attributes(a, p.Attrs)
}
