package web

import (
	"github.com/a-h/templ"
	"github.com/sojupie/tokens/forge/components/ui"
)

func (c componentCase) ProgressAttributes() templ.Attributes {
	attrs := c.Attributes()
	attrs["role"] = "progressbar"
	attrs["aria-label"] = "Loading preview content"
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

func (c componentCase) progressProps(label string) ui.ProgressProps {
	p := ui.ProgressProps{Label: label, Variant: c.Variant, Shape: c.Shape, Size: ui.Size(c.Size), Contained: c.Variant == "contained", Attrs: c.ProgressAttributes()}
	p.Attrs["aria-label"] = label
	if c.Control == "determinate" {
		value := 40.0
		p.Value = &value
	}
	return p
}
