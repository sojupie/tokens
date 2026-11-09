package model

import "github.com/a-h/templ"

type ProgressProps struct {
	ID, Label, Variant, Shape string
	Size                      Size
	// A nil Value is indeterminate. Values are percentages in [0,100].
	Value     *float64
	Contained bool
	// CSSOnly uses the common stylesheet for thin, flat, indeterminate linear progress.
	CSSOnly bool
	Attrs   templ.Attributes
}

func (p ProgressProps) Attributes(loading bool) templ.Attributes {
	class, variant := "md-progress", Fallback(p.Variant, "linear")
	if loading {
		class = "md-loading"
		variant = "uncontained"
		if p.Contained {
			variant = "contained"
		}
	}
	control := "indeterminate"
	if p.Value != nil {
		control = "determinate"
	}
	a := templ.Attributes{"class": class, "role": "progressbar", "aria-label": Fallback(p.Label, "Loading"), "data-variant": variant, "data-control": control, "data-shape": Fallback(p.Shape, "flat"), "data-size": Fallback(string(p.Size), "thin")}
	if p.ID != "" {
		a["id"] = p.ID
	}
	if p.Value != nil {
		value := max(0, min(100, *p.Value))
		a["aria-valuemin"] = 0
		a["aria-valuemax"] = 100
		a["aria-valuenow"] = value
	}
	return Attributes(a, p.Attrs)
}
