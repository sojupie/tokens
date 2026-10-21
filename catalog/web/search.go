package web

import (
	"fmt"
	"github.com/a-h/templ"
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

func searchMailNavigation() componentCase {
	return componentCase{Family: "nav-rail", Variant: "collapsed", State: "interactive", Size: "default", Width: "default", Orientation: "vertical", Selected: true, NavHeader: "fab"}
}
func (c componentCase) SearchMailNavigation() templ.Attributes {
	child := searchMailNavigation()
	a := child.Attributes()
	a["data-component-case"] = c.ID() + "-mail-" + child.Family
	a["class"] = a["class"].(string) + " md-search-mail-nav"
	return a
}
func (c componentCase) SearchMailDestination(index int) templ.Attributes {
	a := searchMailNavigation().NavigationItem(index)
	a["aria-controls"] = c.ID() + "-mail-screen"
	a["data-search-mail-destination"] = fmt.Sprint(index)
	return a
}
func (c componentCase) SearchCompose() templ.Attributes {
	child := componentCase{Family: "fab", Variant: "tertiary-container", Size: "default", State: "interactive"}
	a := child.Attributes()
	a["data-component-case"] = c.ID() + "-compose-" + child.Family
	a["data-search-compose"] = ""
	a["aria-label"] = "Compose"
	return a
}
func searchMessageList() componentCase {
	return componentCase{Family: "list", Variant: "standard", Size: "three-line", Shape: "baseline", Width: "avatar", State: "interactive", Control: "select", Selection: "single", Selected: true, ListDivider: "none", ListTrailing: "text", ListText: "overline"}
}
func (c componentCase) SearchMessageList() templ.Attributes {
	a := searchMessageList().Attributes()
	a["data-component-case"] = c.ID() + "-messages"
	a["aria-label"] = "Inbox messages"
	return a
}
func (c componentCase) SearchMessageItem(index int, label string) templ.Attributes {
	a := searchMessageList().ListItemWithLabel(index, label)
	a["data-search-message"] = label
	return a
}
func (c componentCase) SearchMessageSurface() templ.Attributes {
	a := (componentCase{Family: "card", Variant: "filled", State: "enabled"}).Attributes()
	a["class"] = "md-card md-filled-card md-search-mail-detail"
	a["id"] = c.ID() + "-mail-screen"
	a["data-component-case"] = c.ID() + "-message-card"
	return a
}

func (c componentCase) SearchReply() templ.Attributes {
	a := (componentCase{Family: "button", Variant: "filled", Size: "small", Shape: "round", State: "interactive"}).Attributes()
	a["data-button-case"] = c.ID() + "-reply"
	a["data-search-reply"] = ""
	return a
}
