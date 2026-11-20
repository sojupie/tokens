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
	return c.TimePickerProps().PanelResolved()
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
	a := templ.Attributes{"class": "md-time-period", "type": "button", "data-time-period": period, "data-selected": fmt.Sprint(selected), "tabindex": tab}
	if c.Control == "period" && c.State != "interactive" && selected == (c.Selection == "selected") {
		a["data-component-state"] = c.State
	}
	return a
}
func (c ComponentCase) TimeTicks() []TimeTick { return c.TimePickerProps().Ticks() }
func (c ComponentCase) TimePickerProps() ui.TimePickerProps {
	return ui.TimePickerProps{
		ID: c.ID(), Hour: 10, Minute: 10, Inline: c.TimePlacement == "inline",
		InlineContainer: c.TimeContainer, Width: ui.TimePickerWidth(c.TimeWidth),
		Preview: c.State != "interactive" && c.TimePlacement != "inline", Format24: c.Width == "24h",
		Expressive: c.TimeStyle == "expressive", InputMode: c.Variant == "input",
		ScrollMode: c.Variant == "scroll", Landscape: c.Orientation == "landscape", Active: c.TimeActive(),
		ScrollFinite: c.TimeScrollFinite,
		ScrollMotion: ui.ScrollFieldMotion(c.ScrollMotion),
		ValueAttrs:   map[string]templ.Attributes{"hour": c.TimeValue("hour"), "minute": c.TimeValue("minute")},
		InputAttrs:   map[string]templ.Attributes{"hour": c.TimeInput("hour"), "minute": c.TimeInput("minute")},
		PeriodAttrs:  map[string]templ.Attributes{"am": c.TimePeriod("am"), "pm": c.TimePeriod("pm")},
	}
}

func (c ComponentCase) ScrollFieldProps() ui.ScrollFieldProps {
	values := []string{"16", "17", "18", "19", "20", "21", "22", "23", "24"}
	label := "Temperature"
	switch c.ScrollContent {
	case "three-digits":
		values = []string{"100", "200", "400", "800", "1600"}
		label = "ISO"
	case "four-digits":
		values = []string{"1000", "1050", "1100", "1150", "1200"}
		label = "Light level"
	case "text":
		values = []string{"Off", "Eco", "Fan", "Cool", "Heat"}
		label = "Climate mode"
	}
	attrs := templ.Attributes{"data-component-state": c.State}
	return ui.ScrollFieldProps{ID: c.ID(), Name: "reading", Label: label, Size: c.Size,
		Motion:  ui.ScrollFieldMotion(c.ScrollMotion),
		Options: values, Index: len(values) / 2, Wrap: c.ScrollWrap, Fill: c.Width == "fill",
		Disabled: c.State == "disabled", DirectEntry: c.ScrollDirectEntry, StartInTextEntry: c.ScrollTextEntry, Attrs: attrs}
}

func (c ComponentCase) ScrollFieldSecondary() ui.ScrollFieldProps {
	p := c.ScrollFieldProps()
	p.ID += "-secondary"
	p.Name = "humidity"
	p.Label = "Humidity"
	p.Options = []string{"25", "30", "35", "40", "45", "50", "55", "60", "65"}
	p.Index = len(p.Options) / 2
	return p
}
