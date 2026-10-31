package ui

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

func (p TimePickerProps) hour() int   { return max(0, min(23, p.Hour)) }
func (p TimePickerProps) minute() int { return max(0, min(59, p.Minute)) }
func (p TimePickerProps) displayHour() int {
	h := p.hour()
	if p.Format24 {
		return h
	}
	h = h % 12
	if h == 0 {
		return 12
	}
	return h
}
func (p TimePickerProps) active() string {
	if p.Active == "minute" {
		return "minute"
	}
	return "hour"
}
func (p TimePickerProps) modeLabel() string {
	if p.InputMode {
		return "Toggle dial picker"
	}
	return "Toggle input picker"
}
func (p TimePickerProps) panel() templ.Attributes {
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
	a := templ.Attributes{"class": "md-time-panel", "id": p.ID + "-time", "data-time-panel": "", "data-mode": mode, "data-format": format, "data-orientation": orientation, "data-active": p.active(), "data-hour": p.hour(), "data-minute": p.minute(), "aria-label": fallback(p.Label, "Choose time")}
	if p.Inline {
		a["open"] = true
		a["data-time-inline"] = ""
	}
	return attributes(a, p.PanelAttrs)
}
func (p TimePickerProps) value(part string) templ.Attributes {
	return attributes(templ.Attributes{"class": "md-time-value", "type": "button", "data-time-value": part, "aria-label": map[string]string{"hour": "Hours", "minute": "Minutes"}[part], "aria-pressed": fmt.Sprint(p.active() == part)}, p.ValueAttrs[part])
}
func (p TimePickerProps) input(part string) templ.Attributes {
	value := p.minute()
	if part == "hour" {
		value = p.displayHour()
	}
	return attributes(templ.Attributes{"class": "md-time-field", "type": "text", "inputmode": "numeric", "maxlength": "2", "autocomplete": "off", "data-time-input": part, "value": fmt.Sprintf("%02d", value), "id": p.ID + "-" + part, "aria-label": map[string]string{"hour": "Hour", "minute": "Minute"}[part], "aria-describedby": p.ID + "-error"}, p.InputAttrs[part])
}
func (p TimePickerProps) period(period string) templ.Attributes {
	selected := (period == "pm") == (p.hour() >= 12)
	tab := -1
	if selected {
		tab = 0
	}
	return attributes(templ.Attributes{"class": "md-time-period", "type": "button", "data-time-period": period, "aria-pressed": fmt.Sprint(selected), "tabindex": tab}, p.PeriodAttrs[period])
}
func (p TimePickerProps) dial() templ.Attributes {
	label, lo, hi, value, angle := "Hours", 1, 12, p.displayHour(), p.hour()%12*30
	if p.active() == "minute" {
		label, lo, hi, value, angle = "Minutes", 0, 59, p.minute(), p.minute()*6
	} else if p.Format24 {
		lo, hi, value = 0, 23, p.hour()
	}
	ring := "outer"
	if p.active() == "hour" && p.Format24 && p.hour() >= 12 {
		ring = "inner"
	}
	return attributes(templ.Attributes{"class": "md-time-dial", "data-time-dial": "", "data-ring": ring, "role": "slider", "tabindex": "0", "aria-label": label, "aria-valuemin": lo, "aria-valuemax": hi, "aria-valuenow": value, "style": fmt.Sprintf("--_time-angle:%ddeg;", angle)}, p.DialAttrs)
}

type TimeTick struct {
	Label string
	X, Y  float64
}

func (p TimePickerProps) Ticks() []TimeTick {
	var ticks []TimeTick
	for ring := range 2 {
		if ring == 1 && (!p.Format24 || p.active() != "hour") {
			break
		}
		radius := 101.0 / 256
		if ring == 1 {
			radius = 69.0 / 256
		}
		for i := range 12 {
			label := fmt.Sprint(i)
			if p.active() == "minute" {
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
