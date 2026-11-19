package model

import (
	"fmt"
	"math"

	"github.com/a-h/templ"
)

type TimePickerProps struct {
	ID, Name, Label                                 string
	Hour, Minute                                    int
	Format24, InputMode, Landscape, Inline          bool
	Expressive, ScrollMode, ReturnToScroll, Preview bool
	Active                                          string
	ScrollFinite                                    bool
	ScrollMotion                                    ScrollFieldMotion
	Attrs, PanelAttrs                               templ.Attributes
	ValueAttrs, InputAttrs, PeriodAttrs             map[string]templ.Attributes
	DialAttrs                                       templ.Attributes
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
func (p TimePickerProps) IsInline() bool {
	return p.Inline && p.Expressive && (p.InputMode || p.ScrollMode)
}

func (p TimePickerProps) ModeLabelResolved() string {
	if p.InputMode {
		if p.Expressive && (p.ReturnToScroll || p.IsInline()) {
			return "Toggle scroll picker"
		}
		return "Toggle dial picker"
	}
	return "Toggle input picker"
}
func (p TimePickerProps) PanelResolved() templ.Attributes {
	mode, format, orientation := "clock", "12h", "portrait"
	if p.InputMode {
		mode = "input"
	} else if p.Expressive && p.ScrollMode {
		mode = "scroll"
	}
	if p.Format24 {
		format = "24h"
	}
	if p.Landscape {
		orientation = "landscape"
	}
	style, returnMode := "baseline", "clock"
	if p.Expressive {
		style = "expressive"
		if p.ScrollMode || p.ReturnToScroll || p.IsInline() {
			returnMode = "scroll"
		}
	}
	a := templ.Attributes{"class": "md-time-panel", "id": p.ID + "-time", "data-time-panel": "", "data-mode": mode, "data-time-style": style, "data-return-mode": returnMode, "data-format": format, "data-orientation": orientation, "data-active": p.ActiveResolved(), "data-hour": p.HourResolved(), "data-minute": p.MinuteResolved(), "aria-label": Fallback(p.Label, "Choose time")}
	if p.IsInline() {
		a["open"] = true
		a["data-time-inline"] = ""
	} else if p.Preview {
		a["open"] = true
		a["data-time-preview"] = ""
	} else {
		a["data-dialog-surface"] = ""
	}
	return Attributes(a, p.PanelAttrs)
}
func (p TimePickerProps) ValueResolved(part string) templ.Attributes {
	return Attributes(templ.Attributes{"class": "md-time-value", "type": "button", "data-time-value": part, "aria-label": map[string]string{"hour": "Hour", "minute": "Minute"}[part], "aria-pressed": fmt.Sprint(p.ActiveResolved() == part)}, p.ValueAttrs[part])
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
	return Attributes(templ.Attributes{"class": "md-time-period", "type": "button", "data-time-period": period, "role": "radio", "aria-checked": fmt.Sprint(selected), "data-selected": fmt.Sprint(selected), "tabindex": tab}, p.PeriodAttrs[period])
}
func (p TimePickerProps) DialResolved() templ.Attributes {
	label, angle := "Hours", p.HourResolved()%12*30
	if p.ActiveResolved() == "minute" {
		label, angle = "Minutes", p.MinuteResolved()*6
	}
	ring := "outer"
	if p.ActiveResolved() == "hour" && p.Format24 && p.HourResolved() >= 12 {
		ring = "inner"
	}
	return Attributes(templ.Attributes{"class": "md-time-dial", "data-time-dial": "", "data-ring": ring, "role": "group", "tabindex": "-1", "aria-label": label, "style": fmt.Sprintf("--_time-angle:%ddeg;", angle)}, p.DialAttrs)
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

func (p TimePickerProps) ScrollField(part string) ScrollFieldProps {
	first, count, index, label := 0, 60, p.MinuteResolved(), "Minute"
	if part == "hour" {
		count, index, label = 24, p.HourResolved(), "Hour"
		if !p.Format24 {
			first, count, index = 1, 12, p.DisplayHourResolved()-1
		}
	}
	options := make([]string, count)
	for i := range options {
		options[i] = fmt.Sprintf("%02d", i+first)
	}
	return ScrollFieldProps{Label: label, Options: options, Index: index, Wrap: !p.ScrollFinite, Fill: true, Motion: p.ScrollMotion,
		Attrs: templ.Attributes{"data-time-scroll": part}}
}
