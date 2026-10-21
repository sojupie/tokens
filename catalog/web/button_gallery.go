package web

import (
	"fmt"
	"strings"

	"github.com/a-h/templ"
)

type buttonCase struct {
	Variant   string `json:"variant"`
	Size      string `json:"size"`
	Shape     string `json:"shape"`
	Selection string `json:"selection"`
	State     string `json:"state"`
}

func (c buttonCase) ID() string {
	return strings.Join([]string{c.Variant, c.Size, c.Shape, c.Selection, c.State}, "-")
}
func (c buttonCase) Label() string {
	selection := c.Selection
	if selection == "" {
		selection = "plain"
	}
	return fmt.Sprintf("%s · %s · %s · %s · %s", c.Variant, c.Size, c.Shape, selection, c.State)
}
func (c buttonCase) Attributes() templ.Attributes {
	attrs := templ.Attributes{"class": "md-button md-" + c.Variant + "-button", "type": "button", "data-size": c.Size, "data-shape": c.Shape, "data-button-state": c.State, "data-button-case": c.ID()}
	if c.Selection != "" {
		attrs["aria-pressed"] = fmt.Sprint(c.Selection == "selected")
		attrs["data-button-toggle"] = ""
	}
	if c.State == "disabled" {
		attrs["disabled"] = true
	}
	if c.State == "interactive" {
		delete(attrs, "data-button-state")
	}
	return attrs
}
