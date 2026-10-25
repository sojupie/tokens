package web

import (
	"fmt"

	"github.com/a-h/templ"
)

func (c componentCase) BarAction(index int) templ.Attributes {
	attrs := templ.Attributes{"type": "button", "class": "md-icon-button md-standard-icon-button", "data-size": "small", "data-shape": "round", "data-width": "default"}
	labels := []string{"Back", "Favorite", "More options"}
	if c.Family == "toolbar" {
		labels = []string{"Edit", "Save", "More"}
		attrs["class"] = "md-icon-button md-toolbar-item"
		if c.Shape != "icon" {
			attrs["class"] = "md-button md-toolbar-item"
		}
		attrs["data-toolbar-item"] = fmt.Sprint(index)
		if index == 2 {
			attrs["data-menu-open"] = c.ID() + "-toolbar-menu"
			attrs["aria-haspopup"] = "menu"
			attrs["aria-expanded"] = "false"
		}
		attrs["tabindex"] = "-1"
		if index == 0 {
			attrs["aria-pressed"] = fmt.Sprint(c.Selected)
			attrs["data-button-toggle"] = ""
			attrs["tabindex"] = "0"
		}
		if c.State != "interactive" {
			state := "enabled"
			if index == 0 {
				state = c.State
			}
			attrs["data-component-state"] = state
			attrs["data-button-state"] = state
			if state == "disabled" {
				attrs["disabled"] = true
			}
		}
	}
	if c.Family == "toolbar" && c.State == "disabled" {
		if index == 0 {
			attrs["tabindex"] = "-1"
		}
		if index == 1 {
			attrs["tabindex"] = "0"
		}
	}
	attrs["aria-label"] = labels[index]
	return attrs
}

func (c componentCase) BarFAB() templ.Attributes {
	return templ.Attributes{"class": "md-button md-fab md-toolbar-fab", "type": "button", "aria-label": "Create", "data-size": c.Size, "data-color": c.Control}
}

func (c componentCase) BarExpanded() bool {
	return c.Variant != "bottom" && c.Variant != "small" && c.Variant != "centered" && c.Variant != "baseline-small" && c.Variant != "baseline-centered"
}

func (c componentCase) AppbarSearch() componentCase {
	return componentCase{instanceID: c.ID() + "-search", Family: "search", Variant: "adaptive", Control: "contained", Shape: "icon", Width: "clear", State: "interactive"}
}
