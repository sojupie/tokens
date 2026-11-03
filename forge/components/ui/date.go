package ui

import (
	"fmt"
	"github.com/a-h/templ"
	"time"
)

// DatePickerProps uses civil dates. Zero Value means no selection; Today and Month
// can be supplied for deterministic rendering. Supported years are 1900–2100.
type DatePickerProps struct {
	ID, Name, EndName, Label, Variant, Pane string
	Value, End, Month, Today                time.Time
	Range, Inline                           bool
	Attrs, PanelAttrs                       templ.Attributes
	DayAttrs                                map[string]templ.Attributes
	YearAttrs, MonthAttrs                   map[int]templ.Attributes
	MenuAttrs, YearMenuAttrs                templ.Attributes
}

var DateWeekdays = []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
var DateMonths = []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}

func (p DatePickerProps) today() time.Time {
	if !p.Today.IsZero() {
		return p.Today
	}
	return time.Now()
}
func (p DatePickerProps) month() time.Time {
	if !p.Month.IsZero() {
		return p.Month
	}
	if !p.Value.IsZero() {
		return p.Value
	}
	return p.today()
}
func dateISO(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}
func (p DatePickerProps) DateEntryLabel(part string) string {
	if part == "end" {
		return "End date"
	}
	if p.Range {
		return "Start date"
	}
	return fallback(p.Label, "Date")
}
func (p DatePickerProps) DateEntry(part string) templ.Attributes {
	value, name := p.Value, p.Name
	if part == "end" {
		value, name = p.End, p.EndName
	}
	a := templ.Attributes{"id": p.ID + "-value-" + part, "type": "text", "data-date-value": part, "value": dateISO(value), "placeholder": p.DateEntryLabel(part), "inputmode": "numeric", "maxlength": "32", "autocomplete": "off", "aria-describedby": p.ID + "-value-help-" + part}
	if name != "" {
		a["name"] = name
	}
	if p.Variant == "docked" {
		a["aria-keyshortcuts"] = "Alt+ArrowDown"
	}
	return a
}

type DateCell struct {
	Col, Row                                      int
	ISO, Label, Name                              string
	Outside, Today, Selected, InRange, Start, End bool
}

func (p DatePickerProps) Rows(month time.Time) [][]DateCell {
	first := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	start := first.AddDate(0, 0, -int(first.Weekday()))
	end := first.AddDate(0, 1, -1)
	n := (int(first.Weekday()) + end.Day() + 6) / 7
	rows := make([][]DateCell, n)
	for row := range n {
		for col := range 7 {
			d := start.AddDate(0, 0, row*7+col)
			iso := dateISO(d)
			rows[row] = append(rows[row], DateCell{Col: col + 1, Row: row + 2, ISO: iso, Label: fmt.Sprint(d.Day()), Name: d.Format("Monday, January 2, 2006"), Outside: d.Month() != first.Month(), Today: iso == dateISO(p.today()), Selected: iso == dateISO(p.Value), Start: iso == dateISO(p.Value), End: !p.End.IsZero() && iso == dateISO(p.End), InRange: !p.Value.IsZero() && !p.End.IsZero() && iso >= dateISO(p.Value) && iso <= dateISO(p.End)})
		}
	}
	return rows
}
func (p DatePickerProps) DatePanel() templ.Attributes {
	calendar, mode := "modal", "calendar"
	if p.Variant == "docked" {
		calendar = "docked"
	}
	if p.Variant == "input" {
		mode = "input"
	}
	a := templ.Attributes{"class": "md-date-panel", "id": p.ID + "-date", "aria-label": fallback(p.Label, "Choose date"), "data-calendar": calendar, "data-mode": mode, "data-variant": fallback(p.Variant, "modal"), "data-range": fmt.Sprint(p.Range), "data-pane": fallback(p.Pane, "day"), "data-month": p.month().Format("2006-01"), "data-today": dateISO(p.today())}
	if p.Inline {
		a["open"] = true
		a["data-date-inline"] = ""
	}
	a = attributes(a, p.PanelAttrs)
	// Match the initial calendar height before the client initializes it.
	style, _ := a["style"].(string)
	a["style"] = style + fmt.Sprintf(";--_date-rows:%d", len(p.Rows(p.month())))
	return a
}
func (p DatePickerProps) DateDay(d DateCell) templ.Attributes {
	selected, inRange := d.Selected, false
	if p.Range {
		selected = d.Start || d.End
		inRange = d.InRange
	}
	tab := -1
	if selected && (!p.Range || d.Start) {
		tab = 0
	}
	a := templ.Attributes{"class": "md-date-day", "type": "button", "role": "gridcell", "aria-label": d.Name, "aria-colindex": fmt.Sprint(d.Col), "aria-rowindex": fmt.Sprint(d.Row), "aria-selected": fmt.Sprint(selected || inRange), "data-date": d.ISO, "data-date-selected": fmt.Sprint(selected), "data-today": fmt.Sprint(d.Today), "data-outside-month": fmt.Sprint(d.Outside), "tabindex": fmt.Sprint(tab)}
	if p.Range {
		a["data-in-range"] = fmt.Sprint(inRange)
	}
	if d.Today {
		a["aria-current"] = "date"
	}
	return attributes(a, p.DayAttrs[d.ISO])
}
func (p DatePickerProps) DateCell(d DateCell) templ.Attributes {
	a := templ.Attributes{"class": "md-date-cell", "role": "presentation"}
	if p.Range {
		a["data-in-range"] = fmt.Sprint(d.InRange)
		a["data-range-start"] = fmt.Sprint(d.Start)
		a["data-range-end"] = fmt.Sprint(d.End)
	}
	return a
}
func (p DatePickerProps) DateYear(year int) templ.Attributes {
	class := "md-date-year"
	if p.Variant == "docked" {
		class = "md-date-month-item"
	}
	tab := -1
	if year == p.month().Year() {
		tab = 0
	}
	return attributes(templ.Attributes{"class": class, "type": "button", "data-date-year": fmt.Sprint(year), "tabindex": fmt.Sprint(tab), "aria-pressed": fmt.Sprint(year == p.month().Year())}, p.YearAttrs[year])
}
func (p DatePickerProps) DateMonth(month int) templ.Attributes {
	tab := -1
	if month == int(p.month().Month()) {
		tab = 0
	}
	return attributes(templ.Attributes{"class": "md-date-month-item", "type": "button", "data-date-month": fmt.Sprint(month), "tabindex": fmt.Sprint(tab), "aria-pressed": fmt.Sprint(month == int(p.month().Month()))}, p.MonthAttrs[month])
}
func (p DatePickerProps) DateMenu() templ.Attributes {
	part, shortcut := "year", "Shift+Y"
	if p.Variant == "docked" {
		part, shortcut = "month", "Shift+M"
	}
	return attributes(templ.Attributes{"type": "button", "class": "md-date-menu-button", "data-date-menu": part, "aria-label": "Choose " + part, "aria-controls": p.ID + "-" + part + "s", "aria-expanded": fmt.Sprint(p.Pane == part), "aria-keyshortcuts": shortcut, "title": "Choose " + part + " (" + shortcut + ")"}, p.MenuAttrs)
}
func (p DatePickerProps) DateYearMenu() templ.Attributes {
	return attributes(templ.Attributes{"type": "button", "class": "md-date-menu-button", "data-date-menu": "year", "aria-label": "Choose year", "aria-controls": p.ID + "-years", "aria-expanded": fmt.Sprint(p.Pane == "year"), "aria-keyshortcuts": "Shift+Y", "title": "Choose year (Shift+Y)"}, p.YearMenuAttrs)
}
func (p DatePickerProps) DateHeadline() string {
	if p.Variant == "input" {
		return "Enter date"
	}
	if p.Value.IsZero() {
		return "Select date"
	}
	if p.Range {
		return p.Value.Format("Jan 2") + " – " + p.End.Format("Jan 2")
	}
	return p.Value.Format("Mon, Jan 2")
}
