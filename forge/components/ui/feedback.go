package ui

import "github.com/a-h/templ"

type TooltipProps struct {
	ID, Title, Text, Variant string
	Inline                   bool
	Actions                  templ.Component
	Attrs                    templ.Attributes
}
type SnackbarProps struct {
	ID, Text, Variant string
	Duration          int
	Inline            bool
	Action            templ.Component
	Attrs             templ.Attributes
}

func (p SnackbarProps) Attributes() templ.Attributes {
	a := templ.Attributes{"id": p.ID, "class": "md-snackbar", "data-variant": fallback(p.Variant, "one-line"), "role": "status", "aria-live": "polite"}
	if !p.Inline {
		a["popover"] = "manual"
	}
	if p.Duration > 0 {
		a["data-snackbar-duration"] = p.Duration
	}
	return attributes(a, p.Attrs)
}
