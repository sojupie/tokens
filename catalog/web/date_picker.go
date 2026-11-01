package web

import (
	"github.com/a-h/templ"
	"github.com/sojupie/tokens/forge/components/ui"
	"time"
)

var dateWeekdays = ui.DateWeekdays
var dateMonths = ui.DateMonths

type dateCell = ui.DateCell

func (c componentCase) dateBase() ui.DatePickerProps {
	return ui.DatePickerProps{ID: c.ID(), Variant: c.Variant, Range: c.DateRange, Pane: c.Control, Inline: c.State != "interactive", Value: datePickerSampleDate(2026, 10, 21), End: datePickerSampleDate(2026, 10, 26), Today: datePickerSampleDate(2026, 10, 18)}
}
func dateRows(month int) [][]dateCell {
	return (componentCase{}).dateBase().Rows(datePickerSampleDate(2026, month, 1))
}
func (c componentCase) DateEntryLabel(part string) string      { return c.dateBase().DateEntryLabel(part) }
func (c componentCase) DateEntry(part string) templ.Attributes { return c.dateBase().DateEntry(part) }
func (c componentCase) DatePanel() templ.Attributes {
	a := c.dateBase().DatePanel()
	a["data-selection"] = c.Selection
	a["data-state"] = c.State
	a["data-selection-date"] = c.dateTarget()
	return a
}
func (c componentCase) dateTarget() string {
	return map[string]string{"selected": "2026-10-21", "unselected": "2026-10-22", "today": "2026-10-18", "in-range": "2026-10-23"}[c.Selection]
}
func (c componentCase) DateDay(day dateCell) templ.Attributes {
	a := c.dateBase().DateDay(day)
	if c.Control == "day" && c.State != "interactive" && day.ISO == c.dateTarget() {
		a["data-component-state"] = c.State
	}
	return a
}
func (c componentCase) DateCell(day dateCell) templ.Attributes { return c.dateBase().DateCell(day) }
func (c componentCase) DateYear(year int) templ.Attributes {
	a := c.dateBase().DateYear(year)
	target := 2026
	if c.Selection == "unselected" {
		target++
	}
	if c.Control == "year" && c.State != "interactive" && year == target {
		a["data-component-state"] = c.State
	}
	return a
}
func (c componentCase) DateMonth(month int) templ.Attributes {
	a := c.dateBase().DateMonth(month)
	target := 10
	if c.Selection == "unselected" {
		target++
	}
	if c.Control == "month" && c.State != "interactive" && month == target {
		a["data-component-state"] = c.State
		if c.State == "disabled" {
			a["disabled"] = true
		}
	}
	return a
}
func (c componentCase) dateMenuState(a templ.Attributes) templ.Attributes {
	if c.Control == "menu-button" && c.State != "interactive" {
		a["data-component-state"] = c.State
		if c.State == "disabled" {
			a["disabled"] = true
		}
	}
	return a
}
func (c componentCase) DateMenu() templ.Attributes { return c.dateMenuState(c.dateBase().DateMenu()) }
func (c componentCase) DateYearMenu() templ.Attributes {
	return c.dateMenuState(c.dateBase().DateYearMenu())
}
func (c componentCase) DateHeadline() string { return c.dateBase().DateHeadline() }

func (c componentCase) datePickerProps() ui.DatePickerProps {
	p := ui.DatePickerProps{ID: c.ID(), Variant: c.Variant, Range: c.DateRange, Pane: c.Control, Inline: c.State != "interactive", Value: time.Date(2026, 10, 21, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC), Today: time.Date(2026, 10, 18, 0, 0, 0, 0, time.UTC), Attrs: c.Attributes(), PanelAttrs: c.DatePanel(), MenuAttrs: c.DateMenu(), YearMenuAttrs: c.DateYearMenu(), DayAttrs: map[string]templ.Attributes{}, YearAttrs: map[int]templ.Attributes{}, MonthAttrs: map[int]templ.Attributes{}}
	for _, month := range []int{10, 11} {
		for _, row := range dateRows(month) {
			for _, day := range row {
				if a := c.DateDay(day); a["data-component-state"] != nil {
					p.DayAttrs[day.ISO] = templ.Attributes{"data-component-state": a["data-component-state"]}
				}
			}
		}
	}
	for year := 1900; year <= 2100; year++ {
		p.YearAttrs[year] = c.DateYear(year)
	}
	for month := 1; month <= 12; month++ {
		p.MonthAttrs[month] = c.DateMonth(month)
	}
	return p
}

func datePickerSampleDate(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
