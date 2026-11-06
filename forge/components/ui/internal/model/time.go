package model

import (
	"fmt"
	"github.com/a-h/templ"
	"math"
)

type TimePickerProps struct {
	ID, Name, Label                        string
	Hour, Minute                           int
	Format24, InputMode, Landscape, Inline bool
	Active                                 string
	Attrs, PanelAttrs                      templ.Attributes
	ValueAttrs, InputAttrs, PeriodAttrs    map[string]templ.Attributes
	DialAttrs                              templ.Attributes
}

func (p TimePickerProps) HourResolved() int   { return max(0, min(23, p.Hour)) }
func (p TimePickerProps) MinuteResolved() int { return max(0, min(59, p.Minute)) }
func (p TimePickerProps) DisplayHourResolved() int {
	h := p.HourResolved()
	if p.Format24 {
		return h
	}
	h = h % 12
	if h == 0 {
		return 12
	}
	return h
}
func (p TimePickerProps) ActiveResolved() string {
	if p.Active == "minute" {
		return "minute"
	}
	return "hour"
}
func (p TimePickerProps) ModeLabelResolved() string {
	if p.InputMode {
		return "Toggle dial picker"
	}
	return "Toggle input picker"
}
func (p TimePickerProps) PanelResolved() templ.Attributes {
	mode, format, orientation := "clock", "12h", "portrait"
	if p.InputMode {
		mode = "input"
	}
	if p.Format24 {
		format = "24h"
	}
	if p.Landscape {
		orientation = "landscape"
	}
	a := templ.Attributes{"class": "md-time-panel", "id": p.ID + "-time", "data-time-panel": "", "data-mode": mode, "data-format": format, "data-orientation": orientation, "data-active": p.ActiveResolved(), "data-hour": p.HourResolved(), "data-minute": p.MinuteResolved(), "aria-label": Fallback(p.Label, "Choose time")}
	if p.Inline {
		a["open"] = true
		a["data-time-inline"] = ""
	}
	return Attributes(a, p.PanelAttrs)
}
func (p TimePickerProps) ValueResolved(part string) templ.Attributes {
	return Attributes(templ.Attributes{"class": "md-time-value", "type": "button", "data-time-value": part, "aria-label": map[string]string{"hour": "Hours", "minute": "Minutes"}[part], "aria-pressed": fmt.Sprint(p.ActiveResolved() == part)}, p.ValueAttrs[part])
}
func (p TimePickerProps) InputResolved(part string) templ.Attributes {
	value := p.MinuteResolved()
	if part == "hour" {
		value = p.DisplayHourResolved()
	}
	return Attributes(templ.Attributes{"class": "md-time-field", "type": "text", "inputmode": "numeric", "maxlength": "2", "autocomplete": "off", "data-time-input": part, "value": fmt.Sprintf("%02d", value), "id": p.ID + "-" + part, "aria-label": map[string]string{"hour": "Hour", "minute": "Minute"}[part], "aria-describedby": p.ID + "-error"}, p.InputAttrs[part])
}
func (p TimePickerProps) PeriodResolved(period string) templ.Attributes {
	selected := (period == "pm") == (p.HourResolved() >= 12)
	tab := -1
	if selected {
		tab = 0
	}
	return Attributes(templ.Attributes{"class": "md-time-period", "type": "button", "data-time-period": period, "aria-pressed": fmt.Sprint(selected), "tabindex": tab}, p.PeriodAttrs[period])
}
func (p TimePickerProps) DialResolved() templ.Attributes {
	label, lo, hi, value, angle := "Hours", 1, 12, p.DisplayHourResolved(), p.HourResolved()%12*30
	if p.ActiveResolved() == "minute" {
		label, lo, hi, value, angle = "Minutes", 0, 59, p.MinuteResolved(), p.MinuteResolved()*6
	} else if p.Format24 {
		lo, hi, value = 0, 23, p.HourResolved()
	}
	ring := "outer"
	if p.ActiveResolved() == "hour" && p.Format24 && p.HourResolved() >= 12 {
		ring = "inner"
	}
	return Attributes(templ.Attributes{"class": "md-time-dial", "data-time-dial": "", "data-ring": ring, "role": "slider", "tabindex": "0", "aria-label": label, "aria-valuemin": lo, "aria-valuemax": hi, "aria-valuenow": value, "style": fmt.Sprintf("--_time-angle:%ddeg;", angle)}, p.DialAttrs)
}

type TimeTick struct {
	Label string
	X, Y  float64
}

func (p TimePickerProps) Ticks() []TimeTick {
	var ticks []TimeTick
	for ring := range 2 {
		if ring == 1 && (!p.Format24 || p.ActiveResolved() != "hour") {
			break
		}
		radius := 101.0 / 256
		if ring == 1 {
			radius = 69.0 / 256
		}
		for i := range 12 {
			label := fmt.Sprint(i)
			if p.ActiveResolved() == "minute" {
				label = fmt.Sprintf("%02d", i*5)
			} else if ring == 1 {
				label = fmt.Sprint(i + 12)
			} else if i == 0 {
				if p.Format24 {
					label = "00"
				} else {
					label = "12"
				}
			}
			theta := float64(i) * math.Pi / 6
			ticks = append(ticks, TimeTick{label, 50 + math.Sin(theta)*radius*100, 50 - math.Cos(theta)*radius*100})
		}
	}
	return ticks
}
func (t TimeTick) Style() string {
	return fmt.Sprintf("--_time-label-x:%g%%;--_time-label-y:%g%%;", t.X, t.Y)
}
