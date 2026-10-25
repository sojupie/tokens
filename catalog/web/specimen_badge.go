package web

func (c componentCase) BadgeLabel() string {
	if c.Variant == "labeled" {
		return "99+ unread notifications"
	}
	return "New notification"
}
