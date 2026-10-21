package web

import (
	"github.com/a-h/templ"
)

func (c componentCase) SliderHandles() []string {
	if c.Control == "range" {
		return []string{"start", "end"}
	}
	return []string{"end"}
}

func (c componentCase) SliderValue(handle string) string {
	if c.Control == "range" {
		if handle == "start" {
			return "30"
		}
		return "70"
	}
	return "40"
}

func (c componentCase) SliderInput(handle string) templ.Attributes {
	label := "Value"
	if c.Control == "range" {
		label = "Upper value"
		if handle == "start" {
			label = "Lower value"
		}
	}
	step := "1"
	if c.Variant == "discrete" {
		step = "10"
	}
	return templ.Attributes{"type": "range", "min": "0", "max": "100", "step": step, "value": c.SliderValue(handle), "name": c.ID() + "-" + handle, "aria-label": label, "disabled": c.State == "disabled", "orient": c.Orientation}
}

func (c componentCase) SliderHandle(handle string) templ.Attributes {
	attrs := templ.Attributes{"class": "md-slider-input", "data-handle": handle}
	if c.State != "interactive" {
		attrs["data-component-state"] = "enabled"
		if handle == c.SliderHandles()[0] {
			attrs["data-component-state"] = c.State
		}
	}
	return attrs
}
