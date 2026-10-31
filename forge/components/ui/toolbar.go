package ui

import "github.com/a-h/templ"

type ToolbarProps struct {
	Label, Variant, Orientation, Size, Content string
	Attrs                                      templ.Attributes
}

func (p ToolbarProps) Attributes() templ.Attributes {
	return attributes(templ.Attributes{"class": "md-toolbar", "role": "toolbar", "aria-label": p.Label, "aria-orientation": fallback(p.Orientation, "horizontal"), "data-orientation": fallback(p.Orientation, "horizontal"), "data-variant": fallback(p.Variant, "docked"), "data-size": fallback(p.Size, "small"), "data-shape": fallback(p.Content, "icon")}, p.Attrs)
}
