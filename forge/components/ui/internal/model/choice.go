package model

import (
	"github.com/a-h/templ"
	"strconv"
)

type ChoiceProps struct {
	ID, Name, Value, Label, LabelledBy                  string
	Checked, Disabled, Required, Invalid, Indeterminate bool
	Attrs, InputAttrs                                   templ.Attributes
	VisibleLabel                                        bool
}

func (p ChoiceProps) InputResolved(kind string) templ.Attributes {
	a := templ.Attributes{"type": kind, "value": Fallback(p.Value, "on"), "checked": p.Checked, "disabled": p.Disabled, "required": p.Required, "aria-invalid": strconv.FormatBool(p.Invalid)}
	if p.ID != "" {
		a["id"] = p.ID
	}
	if p.Name != "" {
		a["name"] = p.Name
	}
	if p.LabelledBy != "" {
		a["aria-labelledby"] = p.LabelledBy
	} else if !p.VisibleLabel && p.Label != "" {
		a["aria-label"] = p.Label
	}
	if p.Indeterminate {
		a["data-indeterminate"] = "true"
	}
	a = Attributes(a, p.InputAttrs)
	if p.VisibleLabel {
		delete(a, "aria-label")
	}
	return a
}
