package presentation

import (
	"fmt"
	"strings"

	"github.com/a-h/templ"
	"github.com/sojupie/tokens/forge/components/ui"
)

var SearchExamples = []string{"Buttons", "Cards", "Checkboxes", "Menus", "Search", "Switches"}

func (c ComponentCase) SearchQuery() string {
	if c.SearchContent == "text" {
		return "Buttons"
	}
	return ""
}
func (c ComponentCase) SearchBar() templ.Attributes {
	a := templ.Attributes{"class": "md-search-bar md-search-trigger", "data-control": c.Control, "data-leading": c.Shape, "data-actions": c.Width}
	if c.SearchElevation == "token" {
		a["class"] = "md-search-bar md-search-trigger md-elevation"
	}
	if c.State != "interactive" && !c.Expanded {
		a["data-component-state"] = c.State
	}
	return a
}
func (c ComponentCase) SearchPanel() templ.Attributes {
	a := templ.Attributes{"class": "md-search-view", "id": c.ID() + "-view", "data-control": c.Control, "data-variant": c.SearchLayout(), "data-layout": c.Variant, "data-open": fmt.Sprint(c.Expanded), "role": "search", "aria-label": "Search tokens"}
	a["data-elevation"] = c.SearchElevation
	if c.SearchElevation == "token" {
		a["class"] = "md-search-view md-elevation"
	}
	if c.Expanded {
		a["role"] = "dialog"
	}
	return a
}

func (c ComponentCase) SearchLayout() string {
	if c.Variant == "adaptive" {
		return "docked"
	}
	return c.Variant
}
func (c ComponentCase) SearchHasAction(action string) bool {
	for _, a := range strings.Split(c.Width, "-") {
		if a == action {
			return true
		}
	}
	return false
}
func (c ComponentCase) SearchOptionsID() string {
	return c.ID() + "-options"
}
func (c ComponentCase) SearchNavigation() ComponentCase {
	return ComponentCase{Family: "nav-drawer", Variant: "modal", State: "interactive", Control: c.ID()}
}
func (c ComponentCase) SearchNavigationPanel() templ.Attributes {
	a := c.SearchNavigation().Attributes()
	a["id"] = c.ID() + "-navigation"
	a["aria-label"] = "Example navigation"
	a["data-navigation-modal"] = ""
	a["data-search-navigation-panel"] = ""
	return a
}

func (c ComponentCase) SearchMailRail() ui.RailProps {
	return ui.RailProps{NavigationProps: ui.NavigationProps{ID: c.ID() + "-mail-nav", Label: "Mail navigation", Variant: "collapsed", Width: "default", Alignment: "top", Attrs: templ.Attributes{"class": "md-search-mail-nav"}}, FAB: &ui.FABProps{Label: "Compose", Icon: "edit", Variant: "tertiary-container", Attrs: templ.Attributes{"data-search-compose": ""}}}
}
func (c ComponentCase) SearchMailDestination(index int) templ.Attributes {
	return templ.Attributes{"aria-controls": c.ID() + "-mail-screen", "data-search-mail-destination": fmt.Sprint(index)}
}
func (c ComponentCase) SearchMessageList() ui.ListProps {
	return ui.ListProps{ID: c.ID() + "-messages", Label: "Inbox messages", Appearance: ui.Expressive, Lines: ui.ThreeLines, Leading: ui.LeadingAvatar}
}
func (c ComponentCase) SearchMessage(index int, label string) ui.ListItemProps {
	selected := index == 0
	return ui.ListItemProps{Label: label, Supporting: "Review the latest component updates and examples.", Overline: "Design team", Avatar: "D", Metadata: "09:30", LeadingKind: ui.LeadingAvatar, Selected: &selected, Attrs: templ.Attributes{"data-search-message": label}}
}
func (c ComponentCase) SearchMessageSurface() templ.Attributes {
	return templ.Attributes{"class": "md-search-mail-detail", "id": c.ID() + "-mail-screen", "data-search-background": ""}
}
func (c ComponentCase) SearchReply() ui.ButtonProps {
	return ui.ButtonProps{Variant: ui.Filled, Attrs: templ.Attributes{"data-search-reply": ""}}
}
