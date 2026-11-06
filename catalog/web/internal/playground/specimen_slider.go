package playground

import (
	"fmt"
	"net/url"
	"strconv"

	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
)

func ConfigureSliderPreview(page *webview.ComponentPreview, query url.Values) error {
	controlsWithoutShape := page.Controls[:0]
	for _, control := range page.Controls {
		if control.Name != "shape" {
			controlsWithoutShape = append(controlsWithoutShape, control)
		}
	}
	page.Controls = controlsWithoutShape
	controls := []webview.PreviewControl{
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
		controls = append(controls, webview.PreviewControl{Name: "lower", Label: "Initial lower value", Selected: "30", Options: controls[3].Options})
	} else {
		controls[3].Label = "Initial value"
	}
	if hasDiscrete {
		controls = append(controls, webview.PreviewControl{Name: "step", Label: "Step", Selected: "10", Options: []string{"5", "10", "20"}})
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
		if !ContainsPreview(control.Options, control.Selected) {
			return fmt.Errorf("%w: %s %s", ErrPreviewInput, control.Name, control.Selected)
		}
		selected[control.Name] = control.Selected
		page.Controls = append(page.Controls, control)
	}

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

		c.Shape = "label"
		c.SliderDisplay, c.SliderIcon, c.SliderStops = selected["display"], selected["icon"], selected["stops"]
		c.SliderStart, c.SliderEnd, c.SliderStep = selected["lower"], selected["value"], selected["step"]
	}
	return nil
}
