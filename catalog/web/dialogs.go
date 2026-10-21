package web

import "github.com/a-h/templ"

func (c componentCase) DialogPanel() templ.Attributes {
	a := c.Attributes()
	a["id"] = c.ID() + "-panel"
	a["data-dialog-preview"] = ""
	a["data-actions"] = c.Orientation
	a["data-scroll-start"] = c.Width
	a["data-variant"] = c.Variant
	a["data-shape"] = c.Shape
	a["role"] = "dialog"
	if c.Variant != "full-screen" {
		a["role"] = "alertdialog"
	}
	a["data-headline"] = c.DialogHeadline
	if c.DialogHeadline == "hidden" {
		a["aria-label"] = c.DialogHeading()
	} else {
		a["aria-labelledby"] = c.ID() + "-heading"
	}
	a["aria-describedby"] = c.ID() + "-support"
	if c.State != "interactive" {
		a["open"] = true
		a["data-dialog-inline"] = ""
	}
	return a
}
func (c componentCase) DialogHeading() string {
	if c.Variant == "full-screen" {
		return "New item"
	}
	return "Edit item"
}
func (c componentCase) DialogAction(save bool) templ.Attributes {
	a := templ.Attributes{"type": "button", "class": "md-button md-text-button"}
	if save {
		a["data-dialog-save"] = ""
		if c.State != "interactive" && c.Control == "action" {
			a["data-component-state"] = c.State
		}
	} else {
		a["data-dialog-cancel"] = ""
	}
	return a
}
func (c componentCase) DialogClose() templ.Attributes {
	a := templ.Attributes{"type": "button", "class": "md-icon-button md-standard-icon-button md-dialog-close", "data-size": "small", "data-shape": "round", "data-width": "default", "data-dialog-cancel": "", "aria-label": "Close"}
	if c.State != "interactive" && c.Control == "close" {
		a["data-component-state"] = c.State
	}
	return a
}

func (c componentCase) DialogIcon() bool { return c.Shape == "icon" || c.Shape == "icon-divided" }
func (c componentCase) DialogDividers() bool {
	return c.Shape == "divided" || c.Shape == "icon-divided"
}
