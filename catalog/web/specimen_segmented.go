package web

import (
	"github.com/a-h/templ"
)

func (c componentCase) SegmentAttributes(index int) templ.Attributes {
	attrs := templ.Attributes{"class": "md-segmented-button"}
	if c.State != "interactive" {
		attrs["data-component-state"] = "enabled"
		if index == 0 {
			attrs["data-component-state"] = c.State
		}
	}
	return attrs
}

func (c componentCase) SegmentInput(index int, label string) templ.Attributes {
	kind := "radio"
	if c.Variant == "multiple" {
		kind = "checkbox"
	}
	return templ.Attributes{"type": kind, "name": c.ID(), "checked": c.ItemSelected(index), "disabled": c.State == "disabled", "aria-label": label}
}

func (c componentCase) SegmentRole() string {
	if c.Variant == "multiple" {
		return "group"
	}
	return "radiogroup"
}
