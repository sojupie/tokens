package web

func (c componentCase) ItemSelected(index int) bool {
	return index == 0 && c.Selected || index == 1 && !c.Selected
}
