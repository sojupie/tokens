package ui

import (
	"github.com/a-h/templ"
)

type SegmentedProps struct {
	Label, Content, Density string
	Multiple                bool
	Attrs                   templ.Attributes
}

func (p SegmentedProps) Attributes() templ.Attributes {
	variant, role := "single", "radiogroup"
	if p.Multiple {
		variant, role = "multiple", "group"
	}
	return attributes(templ.Attributes{"class": "md-segmented-buttons", "role": role, "aria-label": p.Label, "data-variant": variant, "data-control": fallback(p.Content, "icon-label"), "data-density": fallback(p.Density, "0")}, p.Attrs)
}

type SegmentProps struct {
	Label, Icon, Name, Value, Content string
	Selected, Disabled, Multiple      bool
	Attrs, InputAttrs                 templ.Attributes
}

func (p SegmentProps) input() templ.Attributes {
	kind := "radio"
	if p.Multiple {
		kind = "checkbox"
	}
	return attributes(templ.Attributes{"type": kind, "name": p.Name, "value": p.Value, "checked": p.Selected, "disabled": p.Disabled, "aria-label": p.Label}, p.InputAttrs)
}
