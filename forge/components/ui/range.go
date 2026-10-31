package ui

import (
	"fmt"
	"github.com/a-h/templ"
)

type SliderProps struct {
	ID, Name, Label       string
	Min, Max, Value, Step float64
	// Start enables a two-handle range; Value supplies its upper bound.
	Start                        *float64
	Size                         Size
	Vertical, Discrete, Disabled bool
	Display, EndStops            string
	Icon                         templ.Component
	Attrs, StartAttrs, EndAttrs  templ.Attributes
}

func (p SliderProps) limits() (float64, float64, float64) {
	lo, hi := p.Min, p.Max
	if hi <= lo {
		hi = lo + 100
	}
	step := p.Step
	if step <= 0 {
		step = 1
	}
	return lo, hi, step
}
func (p SliderProps) Attributes() templ.Attributes {
	orientation, variant, control := "horizontal", "continuous", "single"
	if p.Vertical {
		orientation = "vertical"
	}
	if p.Discrete {
		variant = "discrete"
	}
	if p.Start != nil {
		control = "range"
	}
	return attributes(templ.Attributes{"class": "md-slider", "role": "group", "aria-label": fallback(p.Label, "Value"), "data-size": fallback(string(p.Size), "small"), "data-variant": variant, "data-control": control, "data-orientation": orientation, "data-value-display": fallback(p.Display, "interaction"), "data-end-stops": fallback(p.EndStops, "shown")}, p.Attrs)
}
func (p SliderProps) handles() []string {
	if p.Start != nil {
		return []string{"start", "end"}
	}
	return []string{"end"}
}
func (p SliderProps) value(handle string) float64 {
	lo, hi, _ := p.limits()
	end := max(lo, min(hi, p.Value))
	if handle == "start" && p.Start != nil {
		return max(lo, min(end, *p.Start))
	}
	return end
}
func (p SliderProps) input(handle string) templ.Attributes {
	lo, hi, step := p.limits()
	label := fallback(p.Label, "Value")
	name := fallback(p.Name, p.ID)
	if p.Start != nil {
		if handle == "start" {
			label += " — lower"
		} else {
			label += " — upper"
		}
		name += "-" + handle
	}
	a := templ.Attributes{"id": p.ID + "-" + handle, "type": "range", "name": name, "min": lo, "max": hi, "step": step, "value": p.value(handle), "aria-label": label, "disabled": p.Disabled}
	if p.Vertical {
		a["orient"] = "vertical"
	}
	return a
}
func (p SliderProps) ticks() []float64 {
	if !p.Discrete {
		return nil
	}
	lo, hi, step := p.limits()
	var ticks []float64
	// Native ranges can have arbitrarily fine steps; cap decorative ticks.
	if (hi-lo)/step > 1000 {
		return nil
	}
	for value := lo + step; value < hi; value += step {
		ticks = append(ticks, (value-lo)/(hi-lo)*100)
	}
	return ticks
}
func (p SliderProps) handle(handle string) templ.Attributes {
	extra := p.EndAttrs
	if handle == "start" {
		extra = p.StartAttrs
	}
	return attributes(templ.Attributes{"class": "md-slider-input", "data-handle": handle}, extra)
}
func sliderNumber(value float64) string { return fmt.Sprint(value) }
