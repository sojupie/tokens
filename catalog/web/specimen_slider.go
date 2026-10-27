package web

import (
	"fmt"
	"github.com/a-h/templ"
	"net/url"
	"strconv"
)

func (c componentCase) SliderHandles() []string {
	if c.Control == "range" {
		return []string{"start", "end"}
	}
	return []string{"end"}
}

func (c componentCase) SliderValue(handle string) string {
	if c.Control == "range" {
		if handle == "start" {
			if c.SliderStart != "" {
				return c.SliderStart
			}
			return "30"
		}
		if c.SliderEnd != "" {
			return c.SliderEnd
		}
		return "70"
	}
	if c.SliderEnd != "" {
		return c.SliderEnd
	}
	return "40"
}

func (c componentCase) SliderInput(handle string) templ.Attributes {
	label := "Value"
	if c.Control == "range" {
		label = "Upper value"
		if handle == "start" {
			label = "Lower value"
		}
	}
	step := "1"
	if c.Variant == "discrete" {
		step = "10"
		if c.SliderStep != "" {
			step = c.SliderStep
		}
	}
	return templ.Attributes{"type": "range", "min": "0", "max": "100", "step": step, "value": c.SliderValue(handle), "name": c.ID() + "-" + handle, "aria-label": label, "disabled": c.State == "disabled", "orient": c.Orientation}
}

func (c componentCase) SliderHandle(handle string) templ.Attributes {
	attrs := templ.Attributes{"class": "md-slider-input", "data-handle": handle}
	if c.State != "interactive" {
		attrs["data-component-state"] = "enabled"
		if handle == c.SliderHandles()[0] {
			attrs["data-component-state"] = c.State
		}
	}
	return attrs
}

func (c componentCase) SliderAttributes() templ.Attributes {
	attrs := c.Attributes()
	attrs["data-value-display"] = c.SliderDisplay
	attrs["data-end-stops"] = c.SliderStops
	return attrs
}
func (c componentCase) SliderInsetIcon() string {
	if c.Variant != "continuous" || c.Control == "range" || (c.Size != "medium" && c.Size != "large" && c.Size != "xlarge") {
		return ""
	}
	switch c.SliderIcon {
	case "volume":
		return "volume_up"
	case "brightness":
		return "brightness_high"
	}
	return ""
}
func (c componentCase) SliderTicks() []int {
	step := 10
	if c.SliderStep != "" {
		step, _ = strconv.Atoi(c.SliderStep)
	}
	var ticks []int
	if step <= 0 {
		return ticks
	}
	for value := step; value < 100; value += step {
		ticks = append(ticks, value)
	}
	return ticks
}
func configureSliderPreview(page *componentPreview, query url.Values) error {
	controlsWithoutShape := page.Controls[:0]
	for _, control := range page.Controls {
		if control.Name != "shape" {
			controlsWithoutShape = append(controlsWithoutShape, control)
		}
	}
	page.Controls = controlsWithoutShape
	controls := []previewControl{
		{Name: "display", Label: "Show value", Selected: "always", Options: []string{"interaction", "always", "hidden"}, Labels: map[string]string{"interaction": "On interaction", "always": "Always", "hidden": "Hidden"}},
		{Name: "icon", Label: "Inset icon", Selected: "none", Options: []string{"none", "volume", "brightness"}, Labels: map[string]string{"none": "None", "volume": "Volume", "brightness": "Brightness"}},
		{Name: "stops", Label: "End stops", Selected: "shown", Options: []string{"shown", "hidden"}},
		{Name: "value", Label: "Initial value / upper", Selected: "40", Options: []string{"0", "10", "20", "30", "40", "50", "60", "70", "80", "90", "100"}},
	}
	hasRange, hasDiscrete := false, false
	for _, c := range page.Cases {
		hasRange = hasRange || c.Control == "range"
		hasDiscrete = hasDiscrete || c.Variant == "discrete"
	}
	if hasRange {
		controls[3].Selected = "70"
		controls = append(controls, previewControl{Name: "lower", Label: "Initial lower value", Selected: "30", Options: controls[3].Options})
	} else {
		controls[3].Label = "Initial value"
	}
	if hasDiscrete {
		controls = append(controls, previewControl{Name: "step", Label: "Step", Selected: "10", Options: []string{"5", "10", "20"}})
	}
	selected := map[string]string{}
	for _, control := range controls {
		if control.Name == "icon" {
			available := false
			for _, c := range page.Cases {
				c.SliderIcon = "volume"
				available = available || c.SliderInsetIcon() != ""
			}
			if !available {
				selected["icon"] = "none"
				continue
			}
		}
		if value := query.Get("preview." + control.Name); value != "" {
			control.Selected = value
		}
		if !containsPreview(control.Options, control.Selected) {
			return fmt.Errorf("%w: %s %s", errPreviewInput, control.Name, control.Selected)
		}
		selected[control.Name] = control.Selected
		page.Controls = append(page.Controls, control)
	}
	// Keep the controls and rendered values in agreement after changing step or range.
	if hasDiscrete {
		step, _ := strconv.Atoi(selected["step"])
		for _, key := range []string{"lower", "value"} {
			if selected[key] != "" {
				n, _ := strconv.Atoi(selected[key])
				selected[key] = strconv.Itoa((n + step/2) / step * step)
			}
		}
	}
	if hasRange {
		lo, _ := strconv.Atoi(selected["lower"])
		hi, _ := strconv.Atoi(selected["value"])
		if lo > hi {
			selected["lower"] = selected["value"]
		}
	}
	for i := range page.Controls {
		name := page.Controls[i].Name
		if name == "lower" || name == "value" {
			page.Controls[i].Selected = selected[name]
		}
	}
	for i := range page.Cases {
		c := &page.Cases[i]
		// Label and value-indicator name token groups for the same anatomy, not slider variants.
		c.Shape = "label"
		c.SliderDisplay, c.SliderIcon, c.SliderStops = selected["display"], selected["icon"], selected["stops"]
		c.SliderStart, c.SliderEnd, c.SliderStep = selected["lower"], selected["value"], selected["step"]
	}
	return nil
}
