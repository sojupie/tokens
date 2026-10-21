package web

import (
	"fmt"
	"github.com/a-h/templ"
)

func (c componentCase) SheetAttributes() templ.Attributes {
	attrs := c.Attributes()
	attrs["id"] = c.ID() + "-panel"
	attrs["tabindex"] = "-1"
	return attrs
}
func (c componentCase) SheetHandle() templ.Attributes {
	minimized := c.Variant == "minimized"
	attrs := templ.Attributes{"type": "button", "class": "md-sheet-handle", "data-sheet-handle": "", "aria-controls": c.ID() + "-content", "aria-expanded": fmt.Sprint(!minimized), "aria-label": "Minimize sheet"}
	if minimized {
		attrs["aria-label"] = "Expand sheet"
	}
	if c.State == "focused" {
		attrs["data-component-state"] = "focused"
	}
	if c.Control == "modal" && c.State == "interactive" {
		attrs["autofocus"] = true
		attrs["data-sheet-initial-focus"] = ""
	}
	return attrs
}
func (c componentCase) SheetAction(index int) templ.Attributes {
	class := "md-button md-text-button"
	if c.Family == "side-sheet" {
		class = "md-button md-sheet-action"
	}
	attrs := templ.Attributes{"type": "button", "class": class, "data-size": "small", "data-shape": "round", "data-sheet-close": ""}
	if index == 1 {
		attrs["data-sheet-save"] = ""
	}
	if c.State != "interactive" {
		state := "enabled"
		if c.Family == "side-sheet" && index == 1 {
			state = c.State
		}
		attrs["data-component-state"] = state
		attrs["data-button-state"] = state
		if state == "disabled" {
			attrs["disabled"] = true
		}
	}
	return attrs
}

func (c componentCase) SheetOpener() templ.Attributes {
	attrs := templ.Attributes{"class": "md-button md-outlined-button", "type": "button", "data-sheet-open": "", "aria-controls": c.ID() + "-panel", "aria-expanded": "false"}
	if c.Control == "modal" {
		attrs["aria-haspopup"] = "dialog"
	}
	return attrs
}
