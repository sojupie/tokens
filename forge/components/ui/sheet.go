package ui

import "github.com/a-h/templ"

type SheetKind string

const (
	BottomSheet SheetKind = "bottom"
	SideSheet   SheetKind = "side"
)

type SheetProps struct {
	ID, Title                         string
	Kind                              SheetKind
	Modal, Detached, Start, Minimized bool
	Inline, Hidden, HideClose         bool
	HandleAttrs, CloseAttrs           templ.Attributes
	Actions, Close                    templ.Component
	Attrs                             templ.Attributes
}

func (p SheetProps) Attributes() templ.Attributes {
	kind := fallback(string(p.Kind), "bottom")
	variant, control, orientation := "expanded", "standard", "end"
	if kind == "side" {
		variant = "docked"
		if p.Detached {
			variant = "detached"
		}
	} else if p.Minimized {
		variant = "minimized"
	}
	if p.Modal {
		control = "modal"
	}
	if p.Start {
		orientation = "start"
	}
	return attributes(templ.Attributes{"id": p.ID + "-panel", "class": "md-sheet md-" + kind + "-sheet", "tabindex": "-1", "data-control": control, "data-variant": variant, "data-orientation": orientation, "aria-labelledby": p.ID + "-heading"}, p.Attrs)
}
