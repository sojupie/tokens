package ui

import "github.com/a-h/templ"

type TooltipProps struct {
	ID, Title, Text, Variant string
	Inline                   bool
	Actions, Body            templ.Component
	Attrs                    templ.Attributes
}

func (p TooltipProps) Attributes() templ.Attributes {
	a := templ.Attributes{"class": "md-tooltip", "data-variant": fallback(p.Variant, "plain"), "role": "tooltip"}
	if p.ID != "" {
		a["id"] = p.ID
	}
	if p.Variant == "rich" {
		a["role"] = "dialog"
		a["aria-label"] = p.Title
		if !p.Inline {
			a["popover"] = "auto"
		}
	}
	return attributes(a, p.Attrs)
}

type SnackbarProps struct {
	ID, Text, Variant   string
	Duration            int
	Inline, HideDismiss bool
	Action              templ.Component
	Attrs               templ.Attributes
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
