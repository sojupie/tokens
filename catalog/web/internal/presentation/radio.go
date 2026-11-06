package presentation

import (
	"fmt"

	"github.com/a-h/templ"
)

func (c ComponentCase) RadioControl(index int) templ.Attributes {
	a := c.Attributes()
	delete(a, "aria-disabled")
	a["data-component-case"] = fmt.Sprintf("%s-option-%d", c.ID(), index+1)
	return a
}
func (c ComponentCase) RadioInput(index int) templ.Attributes {
	label := c.Label()
	if c.State == "interactive" {
		label = fmt.Sprintf("Option %d", index+1)
	}
	a := templ.Attributes{"type": "radio", "name": c.ID() + "-options", "value": fmt.Sprint(index + 1), "aria-label": label}
	if c.Selected && index == 0 {
		a["checked"] = true
	}
	if c.State == "disabled" {
		a["disabled"] = true
	}
	return a
}
