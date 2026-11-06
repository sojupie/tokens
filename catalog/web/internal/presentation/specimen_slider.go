package presentation

import (
	"strconv"

	"github.com/a-h/templ"
	"github.com/sojupie/tokens/forge/components/ui"
)

func (c ComponentCase) SliderHandles() []string {
	if c.Control == "range" {
		return []string{"start", "end"}
	}
	return []string{"end"}
}

func (c ComponentCase) SliderValue(handle string) string {
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

func (c ComponentCase) SliderInput(handle string) templ.Attributes {
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

func (c ComponentCase) SliderHandle(handle string) templ.Attributes {
	attrs := templ.Attributes{"class": "md-slider-input", "data-handle": handle}
	if c.State != "interactive" {
		attrs["data-component-state"] = "enabled"
		if handle == c.SliderHandles()[0] {
			attrs["data-component-state"] = c.State
		}
	}
	return attrs
}

func (c ComponentCase) SliderAttributes() templ.Attributes {
	attrs := c.Attributes()
	attrs["data-value-display"] = c.SliderDisplay
	attrs["data-end-stops"] = c.SliderStops
	return attrs
}
func (c ComponentCase) SliderInsetIcon() string {
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
func (c ComponentCase) SliderTicks() []int {
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

func SliderSurfaceLabel(labels []string) string {
	if len(labels) > 0 {
		return labels[0]
	}
	return "Value"
}
func (c ComponentCase) SliderInputNamed(handle, label string) templ.Attributes {
	a := c.SliderInput(handle)
	a["aria-label"] = label
	return a
}

func (c ComponentCase) SliderProps(label string) ui.SliderProps {
	end, _ := strconv.ParseFloat(c.SliderValue("end"), 64)
	step, _ := strconv.ParseFloat(c.SliderStep, 64)
	if c.Variant == "discrete" && step == 0 {
		step = 10
	}
	p := ui.SliderProps{ID: c.ID(), Name: c.ID() + "-end", Label: label, Value: end, Max: 100, Step: step, Size: ui.Size(c.Size), Vertical: c.Orientation == "vertical", Discrete: c.Variant == "discrete", Disabled: c.State == "disabled", Display: c.SliderDisplay, EndStops: c.SliderStops, Attrs: c.SliderAttributes(), StartAttrs: c.SliderHandle("start"), EndAttrs: c.SliderHandle("end")}
	if c.Control == "range" {
		start, _ := strconv.ParseFloat(c.SliderValue("start"), 64)
		p.Start = &start
		p.Name = c.ID()
	}
	if icon := c.SliderInsetIcon(); icon != "" {
		p.Icon = ui.Icon(icon)
	}
	return p
}
