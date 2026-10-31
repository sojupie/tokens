package web

import (
	"fmt"
	"net/url"
)

func configureBadgePreview(page *componentPreview, query url.Values) error {
	for _, choice := range []previewControl{
		{Name: "content", Label: "Notification", Options: []string{"99+", "1", "12", "999", "1000", "New"}},
		{Name: "placement", Label: "Placement", Options: []string{"icon", "trailing"}},
		{Name: "direction", Label: "Direction", Options: []string{"ltr", "rtl"}},
	} {
		choice.Selected = query.Get("preview." + choice.Name)
		if choice.Selected == "" {
			choice.Selected = choice.Options[0]
		}
		if !containsPreview(choice.Options, choice.Selected) {
			return fmt.Errorf("%w: %s %s", errPreviewInput, choice.Name, choice.Selected)
		}
		page.Controls = append(page.Controls, choice)
		for i := range page.Cases {
			switch choice.Name {
			case "content":
				page.Cases[i].BadgeContent = choice.Selected
			case "placement":
				page.Cases[i].BadgePlacement = choice.Selected
			case "direction":
				page.Cases[i].BadgeDirection = choice.Selected
			}
		}
	}
	return nil
}
func (c componentCase) BadgeText() string {
	if c.BadgeContent != "" {
		return c.BadgeContent
	}
	return "99+"
}

func (c componentCase) BadgeLabel() string {
	if c.Variant == "labeled" {
		if c.BadgeText() == "New" {
			return "New notification"
		}
		return c.BadgeText() + " unread notifications"
	}
	return "New notification"
}

func (c componentCase) NavigationBadgeLabel() string {
	if c.NavBadge == "labeled" {
		if c.Family == "nav-bar" && !c.Selected {
			return "3 new saved items"
		}
		return "3 unread files"
	}
	return "New notification"
}
