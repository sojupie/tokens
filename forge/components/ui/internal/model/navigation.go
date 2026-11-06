package model

import (
	"github.com/a-h/templ"
)

type NavigationProps struct {
	ID, Label, Kind, Variant, Alignment, Direction, Orientation, Layout, Indicator, Width string
	Size                                                                                  Size
	Attrs                                                                                 templ.Attributes
}

func (p NavigationProps) Attributes() templ.Attributes {
	kind := Fallback(p.Kind, "bar")
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
	return Attributes(templ.Attributes{"data-orientation": orientation, "data-layout": Fallback(p.Layout, "standard"), "data-indicator": Fallback(p.Indicator, "fill"), "data-width": p.Width, "data-size": Fallback(string(p.Size), "default"), "class": "md-navigation md-navigation-" + kind, "id": p.ID, "aria-label": p.Label, "data-variant": variant, "data-navigation-group": "", "data-alignment": Fallback(p.Alignment, "center"), "dir": Fallback(p.Direction, "ltr")}, p.Attrs)
}

type NavigationItemProps struct {
	// IconContent supplies a consumer-owned icon, such as a brand mark.
	IconContent                                            templ.Component
	Label, Icon, Href, Badge, BadgeText, BadgeLabel, Count string
	Current, Secondary                                     bool
	Attrs                                                  templ.Attributes
}

func (p NavigationItemProps) Attributes() templ.Attributes {
	a := templ.Attributes{"class": "md-nav-item", "data-navigation-item": "", "data-navigation-secondary": p.Secondary}
	if p.Current {
		a["aria-current"] = "page"
	}
	return Attributes(a, p.Attrs)
}

//forge:internal
func NavigationIcon(name string, filled bool) string {
	suffix := "_outline"
	if filled {
		suffix = "_filled"
	}
	if IconPath(name+suffix) != "" {
		return name + suffix
	}
	return name
}

// RailProps owns rail anatomy; children supply destinations and section content.
// PersistentMenu reserves the header slot when RailToggle is hosted by RailFrame.
type RailProps struct {
	NavigationProps
	Menu, PersistentMenu bool
	FAB                  *FABProps
}

func (p RailProps) ExpandedResolved() bool { return p.Variant == "expanded" }
func (p RailProps) FabResolved() FABProps {
	f := *p.FAB
	f.Extended = p.ExpandedResolved()
	f.Attrs = Attributes(templ.Attributes{"class": "md-nav-fab"}, f.Attrs)
	return f
}
