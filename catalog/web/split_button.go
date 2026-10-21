package web

import "github.com/a-h/templ"

func (c componentCase) SplitAction(trailing bool) templ.Attributes {
	attrs := (buttonCase{Variant: c.Variant, Size: c.Size, Shape: "round", State: c.State}).Attributes()
	if c.State != "interactive" {
		attrs["data-component-state"] = c.State
	}
	class := "md-split-leading"
	if trailing {
		class = "md-split-trailing"
		attrs["aria-label"] = "More cart options"
		attrs["aria-haspopup"] = "menu"
		attrs["aria-expanded"] = "false"
		attrs["popovertarget"] = c.ID() + "-menu"
	} else {
		attrs["data-split-primary"] = ""
		if c.Shape == "icon" {
			attrs["aria-label"] = "Add to cart"
		}
	}
	attrs["class"] = attrs["class"].(string) + " " + class
	return attrs
}
