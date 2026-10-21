package web

func (c componentCase) DividerInset() string {
	if c.Variant != "inset" {
		return "false"
	}
	if c.Shape == "both" {
		return "middle"
	}
	return "true"
}
