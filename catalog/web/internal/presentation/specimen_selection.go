package presentation

func (c ComponentCase) ItemSelected(index int) bool {
	return index == 0 && c.Selected || index == 1 && !c.Selected
}
