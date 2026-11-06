package presentation

import (
	"fmt"

	"github.com/a-h/templ"
)

func (c ComponentCase) GroupLabel(index int) string {
	if index < len(c.IllustrationLabels) && c.IllustrationLabels[index] != "" {
		return c.IllustrationLabels[index]
	}
	if c.Control == "connected" {
		return []string{"First", "Second", "Third"}[index]
	}
	return []string{"Back", "Edit", "Next"}[index]
}

func (c ComponentCase) GroupButton(index int) templ.Attributes {
	state := "enabled"
	if c.State == "interactive" || index == 1 {
		state = c.State
	}
	selection := ""
	if c.Control == "connected" {
		selection = "unselected"
		if index == 1 && c.Selected || index == 0 && !c.Selected && c.Shape == "single" {
			selection = "selected"
		}
	}
	attrs := (ButtonCase{Variant: c.Variant, Size: c.Size, Shape: "round", State: state, Selection: selection}).Attributes()
	attrs["data-group-button"] = fmt.Sprint(index)
	attrs["tabindex"] = "-1"
	if index == 0 {
		attrs["tabindex"] = "0"
	}
	return attrs
}
