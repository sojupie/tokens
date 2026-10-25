package web

import (
	"fmt"
	"github.com/a-h/templ"
	"time"
)

var dateWeekdays = []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
var dateMonths = []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}

func (c componentCase) DateEntryLabel(part string) string {
	if part == "end" {
		return "End date"
	}
	if c.DateRange {
		return "Start date"
	}
	return "Date"
}

func (c componentCase) DateEntry(part string) templ.Attributes {
	value := "2026-10-21"
	if part == "end" {
		value = "2026-10-26"
	}
	a := templ.Attributes{"id": c.ID() + "-value-" + part, "type": "text", "data-date-value": part, "value": value, "placeholder": c.DateEntryLabel(part), "inputmode": "numeric", "maxlength": "32", "autocomplete": "off", "aria-describedby": c.ID() + "-value-help-" + part}
	if c.Variant == "docked" {
		a["aria-keyshortcuts"] = "Alt+ArrowDown"
	}
	return a
}

type dateCell struct {
	Col, Row                                      int
	ISO, Label, Name                              string
	Outside, Today, Selected, InRange, Start, End bool
}

func dateRows(month int) [][]dateCell {
	first := time.Date(2026, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	start := first.AddDate(0, 0, -int(first.Weekday()))
	end := first.AddDate(0, 1, -1)
	n := (int(first.Weekday()) + end.Day() + 6) / 7
	rows := make([][]dateCell, n)
	for row := range n {
		for col := range 7 {
			d := start.AddDate(0, 0, row*7+col)
			iso := d.Format("2006-01-02")
			rows[row] = append(rows[row], dateCell{Col: col + 1, Row: row + 2, ISO: iso, Label: fmt.Sprint(d.Day()), Name: d.Format("Monday, January 2, 2006"), Outside: d.Month() != first.Month(), Today: iso == "2026-10-18", Selected: iso == "2026-10-21", Start: iso == "2026-10-21", End: iso == "2026-10-26", InRange: iso >= "2026-10-21" && iso <= "2026-10-26"})
		}
	}
	return rows
}
func (c componentCase) DatePanel() templ.Attributes {
	calendar := "modal"
	if c.Variant == "docked" {
		calendar = "docked"
	}
	mode := "calendar"
	if c.Variant == "input" {
		mode = "input"
	}
	a := templ.Attributes{"class": "md-date-panel", "id": c.ID() + "-date", "aria-label": "Choose date", "data-calendar": calendar, "data-mode": mode, "data-variant": c.Variant, "data-range": fmt.Sprint(c.DateRange), "data-pane": c.Control, "data-month": "2026-10", "data-selection": c.Selection, "data-state": c.State}
	if c.State != "interactive" {
		a["open"] = true
		a["data-date-inline"] = ""
	}
	return a
}
func (c componentCase) DateDay(d dateCell) templ.Attributes {
	selected := d.Selected
	inRange := false
	if c.DateRange {
		selected = d.Start || d.End
		inRange = d.InRange
	}
	a := templ.Attributes{"class": "md-date-day", "type": "button", "role": "gridcell", "aria-label": d.Name, "aria-colindex": fmt.Sprint(d.Col), "aria-rowindex": fmt.Sprint(d.Row), "aria-selected": fmt.Sprint(selected || inRange), "data-date": d.ISO, "data-date-selected": fmt.Sprint(selected), "data-today": fmt.Sprint(d.Today), "data-outside-month": fmt.Sprint(d.Outside), "tabindex": "-1"}
	if c.DateRange {
		a["data-in-range"] = fmt.Sprint(inRange)
	}
	if selected && (!c.DateRange || d.Start) {
		a["tabindex"] = "0"
	}
	if d.Today {
		a["aria-current"] = "date"
	}
	target := map[string]string{"selected": "2026-10-21", "unselected": "2026-10-22", "today": "2026-10-18", "in-range": "2026-10-23"}[c.Selection]
	if c.Control == "day" && c.State != "interactive" && d.ISO == target {
		a["data-component-state"] = c.State
	}
	return a
}
func (c componentCase) DateCell(d dateCell) templ.Attributes {
	a := templ.Attributes{"class": "md-date-cell", "role": "presentation"}
	if c.DateRange {
		a["data-in-range"] = fmt.Sprint(d.InRange)
		a["data-range-start"] = fmt.Sprint(d.Start)
		a["data-range-end"] = fmt.Sprint(d.End)
	}
	return a
}
func (c componentCase) DateYear(year int) templ.Attributes {
	selected := year == 2026
	a := templ.Attributes{"class": "md-date-year", "type": "button", "data-date-year": fmt.Sprint(year), "tabindex": "-1", "aria-pressed": fmt.Sprint(selected)}
	if c.Variant == "docked" {
		a["class"] = "md-date-month-item"
	}
	if selected {
		a["tabindex"] = "0"
	}
	target := 2026
	if c.Selection == "unselected" {
		target = 2027
	}
	if c.Control == "year" && c.State != "interactive" && year == target {
		a["data-component-state"] = c.State
	}
	return a
}
func (c componentCase) DateMonth(month int) templ.Attributes {
	selected := month == 10
	a := templ.Attributes{"class": "md-date-month-item", "type": "button", "data-date-month": fmt.Sprint(month), "tabindex": "-1", "aria-pressed": fmt.Sprint(selected)}
	if selected {
		a["tabindex"] = "0"
	}
	target := 10
	if c.Selection == "unselected" {
		target = 11
	}
	if c.Control == "month" && c.State != "interactive" && month == target {
		a["data-component-state"] = c.State
		if c.State == "disabled" {
			a["disabled"] = true
		}
	}
	return a
}

func (c componentCase) DateMenu() templ.Attributes {
	a := templ.Attributes{"type": "button", "class": "md-date-menu-button", "data-date-menu": "year", "aria-expanded": fmt.Sprint(c.Control == "year")}
	a["aria-controls"] = c.ID() + "-years"
	a["aria-keyshortcuts"] = "Shift+Y"
	a["title"] = "Choose year (Shift+Y)"
	if c.Variant == "docked" {
		a["aria-controls"] = c.ID() + "-months"
		a["data-date-menu"] = "month"
		a["aria-label"] = "Choose month"
		a["aria-keyshortcuts"] = "Shift+M"
		a["title"] = "Choose month (Shift+M)"
		a["aria-expanded"] = fmt.Sprint(c.Control == "month")
	}
	if c.Control == "menu-button" && c.State != "interactive" {
		a["data-component-state"] = c.State
		if c.State == "disabled" {
			a["disabled"] = true
		}
	}
	return a
}

func (c componentCase) DateYearMenu() templ.Attributes {
	a := c.DateMenu()
	a["data-date-menu"] = "year"
	a["aria-label"] = "Choose year"
	a["aria-keyshortcuts"] = "Shift+Y"
	a["title"] = "Choose year (Shift+Y)"
	a["aria-controls"] = c.ID() + "-years"
	a["aria-keyshortcuts"] = "Shift+Y"
	a["title"] = "Choose year (Shift+Y)"
	a["aria-expanded"] = fmt.Sprint(c.Control == "year")
	return a
}

func (c componentCase) DateHeadline() string {
	if c.Variant == "input" {
		return "Enter date"
	}
	if c.DateRange {
		return "Oct 21 – Oct 26"
	}
	return "Wed, Oct 21"
}
