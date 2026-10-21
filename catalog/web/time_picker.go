package web

import (
	"fmt"
	"github.com/a-h/templ"
	"math"
)

type timeTick struct {
	Label string
	X, Y  float64
}

func (c componentCase) TimeActive() string {
	if c.Shape == "minute" {
		return "minute"
	}
	return "hour"
}
func (c componentCase) TimePanel() templ.Attributes {
	a := templ.Attributes{"class": "md-time-panel", "id": c.ID() + "-time", "data-time-panel": "", "data-mode": c.Variant, "data-format": c.Width, "data-orientation": c.Orientation, "data-active": c.TimeActive(), "aria-label": "Choose time"}
	if c.State != "interactive" {
		a["open"] = true
		a["data-time-inline"] = ""
	}
	return a
}
func (c componentCase) TimeValue(part string) templ.Attributes {
	selected := part == c.TimeActive()
	a := templ.Attributes{"class": "md-time-value", "type": "button", "data-time-value": part, "aria-label": map[string]string{"hour": "Hours", "minute": "Minutes"}[part], "aria-pressed": fmt.Sprint(selected)}
	if c.Control == "value" && c.State != "interactive" && selected == (c.Selection == "selected") {
		a["data-component-state"] = c.State
	}
	return a
}
func (c componentCase) TimeInput(part string) templ.Attributes {
	a := templ.Attributes{"class": "md-time-field", "type": "text", "inputmode": "numeric", "maxlength": "2", "autocomplete": "off", "data-time-input": part, "value": "10", "id": c.ID() + "-" + part, "aria-label": map[string]string{"hour": "Hours", "minute": "Minutes"}[part], "aria-describedby": c.ID() + "-error"}
	if c.Control == "field" && c.State != "interactive" && part == c.TimeActive() {
		a["data-component-state"] = c.State
	}
	return a
}
func (c componentCase) TimePeriod(period string) templ.Attributes {
	selected := period == "am"
	tab := "-1"
	if selected {
		tab = "0"
	}
	a := templ.Attributes{"class": "md-time-period", "type": "button", "data-time-period": period, "aria-pressed": fmt.Sprint(selected), "tabindex": tab}
	if c.Control == "period" && c.State != "interactive" && selected == (c.Selection == "selected") {
		a["data-component-state"] = c.State
	}
	return a
}
func (c componentCase) TimeDial() templ.Attributes {
	label, max, min := "Hours", 12, 1
	angle := 300
	if c.TimeActive() == "minute" {
		label = "Minutes"
		max = 59
		min = 0
		angle = 60
	} else if c.Width == "24h" {
		max = 23
		min = 0
	}
	return templ.Attributes{"class": "md-time-dial", "data-time-dial": "", "data-ring": "outer", "role": "slider", "tabindex": "0", "aria-label": label, "aria-valuemin": fmt.Sprint(min), "aria-valuemax": fmt.Sprint(max), "aria-valuenow": "10", "style": fmt.Sprintf("--_time-angle:%ddeg;", angle)}
}
func (c componentCase) TimeTicks() []timeTick {
	result := []timeTick{}
	for ring := range 2 {
		if ring == 1 && (c.Width != "24h" || c.TimeActive() != "hour") {
			break
		}
		radius := 101.0 / 256
		if ring == 1 {
			radius = 69.0 / 256
		}
		for i := range 12 {
			label := fmt.Sprint(i)
			if c.TimeActive() == "minute" {
				label = fmt.Sprintf("%02d", i*5)
			} else if ring == 1 {
				label = fmt.Sprint(i + 12)
			} else if i == 0 {
				if c.Width == "24h" {
					label = "00"
				} else {
					label = "12"
				}
			}
			theta := float64(i) * math.Pi / 6
			result = append(result, timeTick{Label: label, X: 50 + math.Sin(theta)*radius*100, Y: 50 - math.Cos(theta)*radius*100})
		}
	}
	return result
}
func (t timeTick) Style() string {
	return fmt.Sprintf("--_time-label-x:%g%%;--_time-label-y:%g%%;", t.X, t.Y)
}
