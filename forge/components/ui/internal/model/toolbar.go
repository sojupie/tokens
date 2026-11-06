package model

import "github.com/a-h/templ"

type ToolbarProps struct {
	Label, Variant, Orientation, Size, Content string
	Attrs                                      templ.Attributes
}

func (p ToolbarProps) Attributes() templ.Attributes {
	return Attributes(templ.Attributes{"class": "md-toolbar", "role": "toolbar", "aria-label": p.Label, "aria-orientation": Fallback(p.Orientation, "horizontal"), "data-orientation": Fallback(p.Orientation, "horizontal"), "data-variant": Fallback(p.Variant, "docked"), "data-size": Fallback(p.Size, "small"), "data-shape": Fallback(p.Content, "icon")}, p.Attrs)
}
