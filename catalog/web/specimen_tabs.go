package web

import (
	"fmt"

	"github.com/a-h/templ"
)

func (c componentCase) TabAttributes(index int) templ.Attributes {
	selected := c.ItemSelected(index)
	attrs := templ.Attributes{"class": "md-tab", "type": "button", "role": "tab", "id": fmt.Sprintf("%s-tab-%d", c.ID(), index), "aria-controls": fmt.Sprintf("%s-panel-%d", c.ID(), index), "aria-selected": fmt.Sprint(selected), "tabindex": "-1"}
	if selected {
		attrs["tabindex"] = "0"
	}
	if c.State != "interactive" {
		attrs["data-component-state"] = "enabled"
		if index == 0 {
			attrs["data-component-state"] = c.State
		}
	}
	return attrs
}
