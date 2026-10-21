package web

import (
	"fmt"
	"github.com/a-h/templ"
	"html"
)

func (c componentCase) RadioControl(index int) templ.Attributes {
	a := c.Attributes()
	delete(a, "aria-disabled")
	a["data-component-case"] = fmt.Sprintf("%s-option-%d", c.ID(), index+1)
	return a
}
func (c componentCase) RadioInput(index int) templ.Attributes {
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

func downloadRadio(value string, checked bool) string {
	flag := ""
	if checked {
		flag = " checked"
	}
	return `<span class="md-radio"><input type="radio" name="scope" value="` + html.EscapeString(value) + `"` + flag + `/><span class="md-radio-icon" aria-hidden="true"></span></span>`
}
