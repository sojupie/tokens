package model

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

func (p SliderProps) LimitsResolved() (float64, float64, float64) {
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
	attrs := Attributes(templ.Attributes{"class": "md-slider", "role": "group", "aria-label": Fallback(p.Label, "Value"), "data-size": Fallback(string(p.Size), "small"), "data-variant": variant, "data-control": control, "data-orientation": orientation, "data-value-display": Fallback(p.Display, "interaction"), "data-end-stops": Fallback(p.EndStops, "shown")}, p.Attrs)
	lo, hi, _ := p.LimitsResolved()
	start, end := 0.0, (p.ValueResolved("end")-lo)/(hi-lo)
	if p.Start != nil {
		start = (p.ValueResolved("start") - lo) / (hi - lo)
	}
	// Paint the native inputs' initial positions before the runtime is ready.
	style := fmt.Sprintf("--_slider-start:%g;--_slider-end:%g;--_slider-mid:%g;", start, end, (start+end)/2)
	if custom, ok := p.Attrs["style"]; ok {
		style += fmt.Sprint(custom)
	}
	attrs["style"] = style
	return attrs
}
func (p SliderProps) HandlesResolved() []string {
	if p.Start != nil {
		return []string{"start", "end"}
	}
	return []string{"end"}
}
func (p SliderProps) ValueResolved(handle string) float64 {
	lo, hi, _ := p.LimitsResolved()
	end := max(lo, min(hi, p.Value))
	if handle == "start" && p.Start != nil {
		return max(lo, min(end, *p.Start))
	}
	return end
}
func (p SliderProps) InputResolved(handle string) templ.Attributes {
	lo, hi, step := p.LimitsResolved()
	label := Fallback(p.Label, "Value")
	name := Fallback(p.Name, p.ID)
	if p.Start != nil {
		if handle == "start" {
			label += " — lower"
		} else {
			label += " — upper"
		}
		name += "-" + handle
	}
	a := templ.Attributes{"id": p.ID + "-" + handle, "type": "range", "name": name, "min": lo, "max": hi, "step": step, "value": p.ValueResolved(handle), "aria-label": label, "disabled": p.Disabled}
	if p.Vertical {
		a["orient"] = "vertical"
	}
	return a
}
func (p SliderProps) TicksResolved() []float64 {
	if !p.Discrete {
		return nil
	}
	lo, hi, step := p.LimitsResolved()
	var ticks []float64

	if (hi-lo)/step > 1000 {
		return nil
	}
	for value := lo + step; value < hi; value += step {
		ticks = append(ticks, (value-lo)/(hi-lo)*100)
	}
	return ticks
}
func (p SliderProps) HandleResolved(handle string) templ.Attributes {
	extra := p.EndAttrs
	if handle == "start" {
		extra = p.StartAttrs
	}
	return Attributes(templ.Attributes{"class": "md-slider-input", "data-handle": handle}, extra)
}

//forge:internal
func SliderNumber(value float64) string { return fmt.Sprint(value) }
