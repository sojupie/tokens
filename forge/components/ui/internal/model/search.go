package model

import (
	"fmt"
	"github.com/a-h/templ"
	"strings"
)

type SearchProps struct {
	ID, Label, Value, Variant, Control, Leading, Actions, Elevation, AvatarText string
	Action, Name                                                                string
	Expanded, Disclosure                                                        bool
	Options, Page                                                               templ.Component
	Attrs, BarAttrs, PanelAttrs                                                 templ.Attributes
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
func (p SearchProps) SearchPanel() templ.Attributes {
	role := "search"
	if p.Expanded {
		role = "dialog"
	}
	return Attributes(templ.Attributes{"class": "md-search-view", "id": p.ID + "-view", "data-control": Fallback(p.Control, "contained"), "data-variant": Fallback(p.Variant, "docked"), "data-layout": Fallback(p.Variant, "docked"), "data-open": fmt.Sprint(p.Expanded), "role": role, "aria-label": p.Label, "data-elevation": p.Elevation}, p.PanelAttrs)
}
