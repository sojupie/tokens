package web

func (c componentCase) BadgeLabel() string {
	if c.Variant == "labeled" {
		return "99+ unread notifications"
	}
	return "New notification"
}

func (c componentCase) NavigationBadgeLabel() string {
	if c.NavBadge == "labeled" {
		return "3 unread files"
	}
	return "New notification"
}
