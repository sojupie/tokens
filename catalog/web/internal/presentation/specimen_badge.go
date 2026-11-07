package presentation

func (c ComponentCase) BadgeText() string {
	if c.BadgeContent != "" {
		return c.BadgeContent
	}
	return "99+"
}

func (c ComponentCase) BadgeLabel() string {
	if c.Variant == "labeled" {
		if c.BadgeText() == "New" {
			return "New notification"
		}
		return c.BadgeText() + " unread notifications"
	}
	return "New notification"
}

func (c ComponentCase) NavigationBadgeLabel() string {
	if c.NavBadge == "labeled" {
		if c.Family == "nav-bar" && !c.Selected {
			return "3 new saved items"
		}
		return "3 unread files"
	}
	return "New notification"
}
