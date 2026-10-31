package ui

import (
	"github.com/a-h/templ"
	"strconv"
)

type ChoiceProps struct {
	ID, Name, Value, Label, LabelledBy                  string
	Checked, Disabled, Required, Invalid, Indeterminate bool
	Attrs, InputAttrs                                   templ.Attributes
	visibleLabel                                        bool
}

func (p ChoiceProps) input(kind string) templ.Attributes {
	a := templ.Attributes{"type": kind, "value": fallback(p.Value, "on"), "checked": p.Checked, "disabled": p.Disabled, "required": p.Required, "aria-invalid": strconv.FormatBool(p.Invalid)}
	if p.ID != "" {
		a["id"] = p.ID
	}
	if p.Name != "" {
		a["name"] = p.Name
	}
	if p.LabelledBy != "" {
		a["aria-labelledby"] = p.LabelledBy
	} else if !p.visibleLabel && p.Label != "" {
		a["aria-label"] = p.Label
	}
	if p.Indeterminate {
		a["data-indeterminate"] = "true"
	}
	a = attributes(a, p.InputAttrs)
	if p.visibleLabel {
		delete(a, "aria-label")
	}
	return a
}
