package ui

import (
	"github.com/a-h/templ"
)

type NavigationProps struct {
	ID, Label, Kind, Variant, Alignment, Direction, Orientation, Layout, Indicator, Width string
	Attrs                                                                                 templ.Attributes
}

func (p NavigationProps) Attributes() templ.Attributes {
	kind := fallback(p.Kind, "bar")
	variant := p.Variant
	if variant == "" {
		variant = map[string]string{"bar": "vertical", "rail": "collapsed", "drawer": "standard"}[kind]
	}
	orientation := p.Orientation
	if orientation == "" {
		orientation = "vertical"
		if kind == "drawer" || kind == "rail" && variant == "expanded" {
			orientation = "horizontal"
		}
		if kind == "bar" {
			orientation = variant
		}
	}
	return attributes(templ.Attributes{"data-orientation": orientation, "data-layout": fallback(p.Layout, "standard"), "data-indicator": fallback(p.Indicator, "fill"), "data-width": p.Width, "class": "md-navigation md-navigation-" + kind, "id": p.ID, "aria-label": p.Label, "data-variant": variant, "data-navigation-group": "", "data-alignment": fallback(p.Alignment, "center"), "dir": fallback(p.Direction, "ltr")}, p.Attrs)
}

type NavigationItemProps struct {
	Label, Icon, Href, Badge, BadgeText, BadgeLabel, Count string
	Current, Secondary                                     bool
	Attrs                                                  templ.Attributes
}

func (p NavigationItemProps) Attributes() templ.Attributes {
	a := templ.Attributes{"class": "md-nav-item", "data-navigation-item": "", "data-navigation-secondary": p.Secondary}
	if p.Current {
		a["aria-current"] = "page"
	}
	return attributes(a, p.Attrs)
}

func navigationIcon(name string, filled bool) string {
	suffix := "_outline"
	if filled {
		suffix = "_filled"
	}
	if IconPath(name+suffix) != "" {
		return name + suffix
	}
	return name
}
