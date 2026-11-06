package model

import (
	"github.com/a-h/templ"
	"strings"
)

type AppBarProps struct {
	Title, Subtitle, Variant, Orientation, Size, ScrollTarget string
	Expanded                                                  bool
	Leading, Actions, Search                                  templ.Component
	Attrs                                                     templ.Attributes
}

func (p AppBarProps) Attributes() templ.Attributes {
	variant := Fallback(p.Variant, "small")
	a := templ.Attributes{"class": "md-app-bar", "aria-label": p.Title, "data-variant": variant, "data-orientation": Fallback(p.Orientation, "horizontal"), "data-size": Fallback(p.Size, "small"), "data-modern": "", "data-control": map[bool]string{true: "subtitle", false: "title"}[p.Subtitle != ""]}
	if strings.HasPrefix(variant, "baseline-") || variant == "bottom" {
		delete(a, "data-modern")
	}
	if p.ScrollTarget != "" {
		a["data-appbar-scroll-target"] = p.ScrollTarget
	}
	return Attributes(a, p.Attrs)
}

func (p AppBarProps) ExpandedResolved() bool {
	v := Fallback(p.Variant, "small")
	return p.Expanded || v != "bottom" && v != "small" && v != "centered" && v != "baseline-small" && v != "baseline-centered"
}
