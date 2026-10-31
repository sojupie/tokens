package ui

import (
	"github.com/a-h/templ"
	"strings"
)

type AppBarProps struct {
	Title, Subtitle, Variant, Orientation, Size string
	Expanded                                    bool
	Leading, Actions, Search                    templ.Component
	Attrs                                       templ.Attributes
}

func (p AppBarProps) Attributes() templ.Attributes {
	variant := fallback(p.Variant, "small")
	a := templ.Attributes{"class": "md-app-bar", "aria-label": p.Title, "data-variant": variant, "data-orientation": fallback(p.Orientation, "horizontal"), "data-size": fallback(p.Size, "small"), "data-modern": "", "data-control": map[bool]string{true: "subtitle", false: "title"}[p.Subtitle != ""]}
	if strings.HasPrefix(variant, "baseline-") || variant == "bottom" {
		delete(a, "data-modern")
	}
	return attributes(a, p.Attrs)
}

func (p AppBarProps) expanded() bool {
	v := fallback(p.Variant, "small")
	return p.Expanded || v != "bottom" && v != "small" && v != "centered" && v != "baseline-small" && v != "baseline-centered"
}
