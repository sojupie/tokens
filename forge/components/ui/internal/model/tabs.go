package model

import (
	"fmt"
	"github.com/a-h/templ"
)

type TabsProps struct {
	Label, Variant, Content, Layout, Alignment, Direction string
	Attrs                                                 templ.Attributes
}

func (p TabsProps) Attributes() templ.Attributes {
	return Attributes(templ.Attributes{"class": "md-tabs", "role": "tablist", "aria-label": p.Label, "data-variant": Fallback(p.Variant, "primary"), "data-control": Fallback(p.Content, "label"), "data-layout": Fallback(p.Layout, "fixed"), "data-alignment": Fallback(p.Alignment, "center"), "dir": Fallback(p.Direction, "ltr")}, p.Attrs)
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
	return Attributes(templ.Attributes{"class": "md-tab", "type": "button", "role": "tab", "id": p.ID, "aria-controls": p.PanelID, "aria-selected": fmt.Sprint(p.Selected), "tabindex": fmt.Sprint(tab), "disabled": p.Disabled}, p.Attrs)
}

func (p TabProps) ContentResolved() string { return Fallback(p.Content, "label") }
