package presentation

import (
	"strings"

	"github.com/a-h/templ"
)

func (c ComponentCase) FeedbackAttributes(class string) templ.Attributes {
	if c.Family == "chip" && c.Variant != "" {
		class = "md-" + c.Variant + "-chip"
	}
	if c.Family == "dialog" && c.Variant == "full-screen" {
		class = "md-full-screen-dialog"
	}
	attrs := templ.Attributes{"class": class, "type": "button"}
	if c.State != "interactive" {
		state := c.State
		if c.Family == "snackbar" && ((c.Control == "dismiss") != strings.Contains(class, "md-snackbar-close")) {
			state = "enabled"
		}
		attrs["data-component-state"] = state
	}
	return attrs
}
