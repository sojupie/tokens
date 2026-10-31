package ui

import (
	"fmt"
	"github.com/a-h/templ"
)

type TabsProps struct {
	Label, Variant, Content, Layout, Alignment, Direction string
	Attrs                                                 templ.Attributes
}

func (p TabsProps) Attributes() templ.Attributes {
	return attributes(templ.Attributes{"class": "md-tabs", "role": "tablist", "aria-label": p.Label, "data-variant": fallback(p.Variant, "primary"), "data-control": fallback(p.Content, "label"), "data-layout": fallback(p.Layout, "fixed"), "data-alignment": fallback(p.Alignment, "center"), "dir": fallback(p.Direction, "ltr")}, p.Attrs)
}

type TabProps struct {
	ID, PanelID, Label, Icon, Badge, BadgeText, Content, Variant string
	Selected, Disabled                                           bool
	Attrs                                                        templ.Attributes
}

func (p TabProps) Attributes() templ.Attributes {
	tab := -1
	if p.Selected {
		tab = 0
	}
	return attributes(templ.Attributes{"class": "md-tab", "type": "button", "role": "tab", "id": p.ID, "aria-controls": p.PanelID, "aria-selected": fmt.Sprint(p.Selected), "tabindex": fmt.Sprint(tab), "disabled": p.Disabled}, p.Attrs)
}
