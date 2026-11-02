package web

import (
	"fmt"
	"github.com/a-h/templ"
	"github.com/sojupie/tokens/forge/components/ui"
	"net/url"
	"strings"
)

var searchExamples = []string{"Buttons", "Cards", "Checkboxes", "Menus", "Search", "Switches"}

func (c componentCase) SearchQuery() string {
	if c.SearchContent == "text" {
		return "Buttons"
	}
	return ""
}
func (c componentCase) SearchBar() templ.Attributes {
	a := templ.Attributes{"class": "md-search-bar md-search-trigger", "data-control": c.Control, "data-leading": c.Shape, "data-actions": c.Width}
	if c.SearchElevation == "token" {
		a["class"] = "md-search-bar md-search-trigger md-elevation"
	}
	if c.State != "interactive" && !c.Expanded {
		a["data-component-state"] = c.State
	}
	return a
}
func (c componentCase) SearchPanel() templ.Attributes {
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

// Query content is independent from the action buttons.
func configureSearchPreview(page *componentPreview, query url.Values) error {
	content := query.Get("preview.content")
	if content == "" {
		content = "empty"
	}
	if content != "empty" && content != "text" {
		return fmt.Errorf("%w: search content %s", errPreviewInput, content)
	}
	elevation := query.Get("preview.shadow")
	if elevation == "" {
		elevation = "none"
	}
	if elevation != "none" && elevation != "token" {
		return fmt.Errorf("%w: search shadow %s", errPreviewInput, elevation)
	}
	page.Controls = append(page.Controls, previewControl{Name: "shadow", Label: "Shadow", Options: []string{"none", "token"}, Selected: elevation, Labels: map[string]string{"none": "None (default)", "token": "Token elevation"}})
	page.Controls = append(page.Controls, previewControl{Name: "content", Label: "Query", Options: []string{"empty", "text"}, Selected: content, Labels: map[string]string{"text": "Sample query"}})
	for i := range page.Controls {
		c := &page.Controls[i]
		if c.Name == "disclosure" {
			c.Label = "Focus"
			c.Labels = map[string]string{"collapsed": "Unfocused", "expanded": "Focused"}
		}
		if c.Name == "variant" {
			c.Label = "Layout"
			c.Labels = map[string]string{"adaptive": "Adaptive (flexible width)", "docked": "Docked", "full-screen": "Full-screen"}
		}
		if c.Name == "control" {
			c.Labels = map[string]string{"baseline": "Divided (baseline)", "contained": "Contained"}
		}
		if c.Name == "shape" {
			c.Labels = map[string]string{"icon": "Search icon", "menu": "Navigation button", "none": "None"}
		}
		if c.Name == "width" {
			c.Labels = map[string]string{"clear": "Clear", "avatar": "Avatar", "none": "None", "settings": "Search options", "clear-settings": "Clear + options", "settings-avatar": "Options + avatar"}
		}
	}
	for i := range page.Cases {
		page.Cases[i].SearchContent = content
		page.Cases[i].SearchElevation = elevation
	}
	return nil
}
func (c componentCase) SearchLayout() string {
	if c.Variant == "adaptive" {
		return "docked"
	}
	return c.Variant
}
func (c componentCase) SearchHasAction(action string) bool {
	for _, a := range strings.Split(c.Width, "-") {
		if a == action {
			return true
		}
	}
	return false
}
func (c componentCase) SearchOptionsID() string {
	return c.ID() + "-options"
}
func (c componentCase) SearchNavigation() componentCase {
	return componentCase{Family: "nav-drawer", Variant: "modal", State: "interactive", Control: c.ID()}
}
func (c componentCase) SearchNavigationPanel() templ.Attributes {
	a := c.SearchNavigation().Attributes()
	a["id"] = c.ID() + "-navigation"
	a["aria-label"] = "Example navigation"
	a["data-navigation-modal"] = ""
	a["data-search-navigation-panel"] = ""
	return a
}

func (c componentCase) SearchMailRail() ui.RailProps {
	return ui.RailProps{NavigationProps: ui.NavigationProps{ID: c.ID() + "-mail-nav", Label: "Mail navigation", Variant: "collapsed", Width: "default", Alignment: "top", Attrs: templ.Attributes{"class": "md-search-mail-nav"}}, FAB: &ui.FABProps{Label: "Compose", Icon: "edit", Variant: "tertiary-container", Attrs: templ.Attributes{"data-search-compose": ""}}}
}
func (c componentCase) SearchMailDestination(index int) templ.Attributes {
	return templ.Attributes{"aria-controls": c.ID() + "-mail-screen", "data-search-mail-destination": fmt.Sprint(index)}
}
func (c componentCase) SearchMessageList() ui.ListProps {
	return ui.ListProps{ID: c.ID() + "-messages", Label: "Inbox messages", Appearance: ui.Expressive, Lines: ui.ThreeLines, Leading: ui.LeadingAvatar}
}
func (c componentCase) SearchMessage(index int, label string) ui.ListItemProps {
	selected := index == 0
	return ui.ListItemProps{Label: label, Supporting: "Review the latest component updates and examples.", Overline: "Design team", Avatar: "D", Metadata: "09:30", LeadingKind: ui.LeadingAvatar, Selected: &selected, Attrs: templ.Attributes{"data-search-message": label}}
}
func (c componentCase) SearchMessageSurface() templ.Attributes {
	return templ.Attributes{"class": "md-search-mail-detail", "id": c.ID() + "-mail-screen", "data-search-background": ""}
}
func (c componentCase) SearchReply() ui.ButtonProps {
	return ui.ButtonProps{Variant: ui.Filled, Attrs: templ.Attributes{"data-search-reply": ""}}
}
func (c componentCase) searchProps(label, value string, disclosure bool) ui.SearchProps {
	return ui.SearchProps{ID: c.ID(), AvatarText: "A", Options: searchExampleOptions(c.ID()), Label: label, Value: value, Variant: c.Variant, Control: c.Control, Leading: c.Shape, Actions: c.Width, Elevation: c.SearchElevation, Expanded: c.Expanded, Disclosure: disclosure, BarAttrs: c.SearchBar(), PanelAttrs: c.SearchPanel()}
}
