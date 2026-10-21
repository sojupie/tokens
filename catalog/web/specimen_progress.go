package web

import (
	"github.com/a-h/templ"
)

func (c componentCase) ProgressAttributes() templ.Attributes {
	attrs := c.Attributes()
	attrs["role"] = "progressbar"
	attrs["aria-label"] = "Loading"
	if c.Control == "determinate" {
		attrs["aria-valuemin"] = "0"
		attrs["aria-valuemax"] = "100"
		attrs["aria-valuenow"] = "40"
	}
	return attrs
}

func (c componentCase) ProgressControl() componentCase {
	return componentCase{Family: "slider", Variant: "continuous", Control: "single", Size: "xsmall", Shape: "value-indicator", Orientation: "horizontal", State: "interactive"}
}
