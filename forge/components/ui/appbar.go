package ui

import "github.com/a-h/templ"

type AppBarProps struct {
	Title, Subtitle, Variant, Orientation, Size string
	Expanded                                    bool
	Leading, Actions, Search                    templ.Component
	Attrs                                       templ.Attributes
}

func (p AppBarProps) Attributes() templ.Attributes {
	variant := fallback(p.Variant, "small")
	return attributes(templ.Attributes{"class": "md-app-bar", "aria-label": p.Title, "data-variant": variant, "data-orientation": fallback(p.Orientation, "horizontal"), "data-size": fallback(p.Size, "small"), "data-modern": "", "data-control": map[bool]string{true: "subtitle", false: "title"}[p.Subtitle != ""]}, p.Attrs)
}
