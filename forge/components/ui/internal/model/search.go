package model

import (
	"fmt"
	"github.com/a-h/templ"
	"strings"
)

type SearchProps struct {
	ID, Label, Value, Variant, Control, Leading, Actions, Elevation, AvatarText string
	Action, Name, Placeholder                                                   string
	Expanded, Disclosure                                                        bool
	Options, Page                                                               templ.Component
	Attrs, BarAttrs, PanelAttrs, InputAttrs                                     templ.Attributes
}

func (p SearchProps) SearchHasAction(action string) bool {
	for _, a := range strings.Split(p.Actions, "-") {
		if a == action {
			return true
		}
	}
	return false
}
func (p SearchProps) SearchOptionsID() string { return p.ID + "-options" }
func (p SearchProps) SearchBar() templ.Attributes {
	return Attributes(templ.Attributes{"class": "md-search-bar md-search-trigger", "data-control": Fallback(p.Control, "contained"), "data-leading": Fallback(p.Leading, "icon"), "data-actions": p.Actions}, p.BarAttrs)
}
func (p SearchProps) SearchInput(label, value string, disclosure bool) templ.Attributes {
	a := templ.Attributes{
		"type": "search", "class": "md-search-input", "aria-label": label,
		"placeholder": Fallback(p.Placeholder, label), "value": value, "autocomplete": "off",
	}
	if p.ID != "" {
		a["id"] = p.ID
	}
	if p.Name != "" {
		a["name"] = p.Name
	}
	if disclosure {
		a["aria-expanded"] = fmt.Sprint(p.Expanded)
		a["aria-controls"] = p.ID + "-results"
		a["readonly"] = !p.Expanded
	}
	return Attributes(a, p.InputAttrs)
}

func (p SearchProps) SearchPanel() templ.Attributes {
	role := "search"
	if p.Expanded {
		role = "dialog"
	}
	return Attributes(templ.Attributes{"class": "md-search-view", "id": p.ID + "-view", "data-control": Fallback(p.Control, "contained"), "data-variant": Fallback(p.Variant, "docked"), "data-layout": Fallback(p.Variant, "docked"), "data-open": fmt.Sprint(p.Expanded), "role": role, "aria-label": p.Label, "data-elevation": p.Elevation}, p.PanelAttrs)
}
