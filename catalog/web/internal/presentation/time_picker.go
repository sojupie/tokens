package presentation

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/sojupie/tokens/forge/components/ui"
)

type TimeTick = ui.TimeTick

func (c ComponentCase) TimeActive() string {
	if c.Shape == "minute" {
		return "minute"
	}
	return "hour"
}
func (c ComponentCase) TimePanel() templ.Attributes {
	a := templ.Attributes{"class": "md-time-panel", "id": c.ID() + "-time", "data-time-panel": "", "data-mode": c.Variant, "data-format": c.Width, "data-orientation": c.Orientation, "data-active": c.TimeActive(), "aria-label": "Choose time"}
	if c.State != "interactive" {
		a["open"] = true
		a["data-time-inline"] = ""
	}
	return a
}
func (c ComponentCase) TimeValue(part string) templ.Attributes {
	selected := part == c.TimeActive()
	a := templ.Attributes{"class": "md-time-value", "type": "button", "data-time-value": part, "aria-label": map[string]string{"hour": "Hours", "minute": "Minutes"}[part], "aria-pressed": fmt.Sprint(selected)}
	if c.Control == "value" && c.State != "interactive" && selected == (c.Selection == "selected") {
		a["data-component-state"] = c.State
	}
	return a
}
func (c ComponentCase) TimeInput(part string) templ.Attributes {
	a := templ.Attributes{"class": "md-time-field", "type": "text", "inputmode": "numeric", "maxlength": "2", "autocomplete": "off", "data-time-input": part, "value": "10", "id": c.ID() + "-" + part, "aria-label": map[string]string{"hour": "Hour", "minute": "Minute"}[part], "aria-describedby": c.ID() + "-error"}
	if c.Control == "field" && c.State != "interactive" && part == c.TimeActive() {
		a["data-component-state"] = c.State
	}
	return a
}
func (c ComponentCase) TimePeriod(period string) templ.Attributes {
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
func (c ComponentCase) TimeDial() templ.Attributes {
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
func (c ComponentCase) TimeTicks() []TimeTick { return c.TimePickerProps().Ticks() }
func (c ComponentCase) TimePickerProps() ui.TimePickerProps {
	return ui.TimePickerProps{ID: c.ID(), Hour: 10, Minute: 10, Inline: c.State != "interactive", Format24: c.Width == "24h", InputMode: c.Variant == "input", Landscape: c.Orientation == "landscape", Active: c.TimeActive(), PanelAttrs: c.TimePanel(), ValueAttrs: map[string]templ.Attributes{"hour": c.TimeValue("hour"), "minute": c.TimeValue("minute")}, InputAttrs: map[string]templ.Attributes{"hour": c.TimeInput("hour"), "minute": c.TimeInput("minute")}, PeriodAttrs: map[string]templ.Attributes{"am": c.TimePeriod("am"), "pm": c.TimePeriod("pm")}, DialAttrs: c.TimeDial()}
}

func (c ComponentCase) TimeModeLabel() string {
	if c.Variant == "input" {
		return "Toggle dial picker"
	}
	return "Toggle input picker"
}
