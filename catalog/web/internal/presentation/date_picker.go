package presentation

import (
	"time"

	"github.com/a-h/templ"
	"github.com/sojupie/tokens/forge/components/ui"
)

type DateCell = ui.DateCell

func (c ComponentCase) DateBase() ui.DatePickerProps {
	return ui.DatePickerProps{ID: c.ID(), Variant: c.Variant, Range: c.DateRange, Pane: c.Control, Inline: c.State != "interactive", Value: DatePickerSampleDate(2026, 10, 21), End: DatePickerSampleDate(2026, 10, 26), Today: DatePickerSampleDate(2026, 10, 18)}
}
func DateRows(month int) [][]DateCell {
	return (ComponentCase{}).DateBase().Rows(DatePickerSampleDate(2026, month, 1))
}
func (c ComponentCase) DateEntryLabel(part string) string      { return c.DateBase().DateEntryLabel(part) }
func (c ComponentCase) DateEntry(part string) templ.Attributes { return c.DateBase().DateEntry(part) }
func (c ComponentCase) DatePanel() templ.Attributes {
	a := c.DateBase().DatePanel()
	a["data-selection"] = c.Selection
	a["data-state"] = c.State
	a["data-selection-date"] = c.DateTarget()
	return a
}
func (c ComponentCase) DateTarget() string {
	return map[string]string{"selected": "2026-10-21", "unselected": "2026-10-22", "today": "2026-10-18", "in-range": "2026-10-23"}[c.Selection]
}
func (c ComponentCase) DateDay(day DateCell) templ.Attributes {
	a := c.DateBase().DateDay(day)
	if c.Control == "day" && c.State != "interactive" && day.ISO == c.DateTarget() {
		a["data-component-state"] = c.State
	}
	return a
}
func (c ComponentCase) DateCell(day DateCell) templ.Attributes { return c.DateBase().DateCell(day) }
func (c ComponentCase) DateYear(year int) templ.Attributes {
	a := c.DateBase().DateYear(year)
	target := 2026
	if c.Selection == "unselected" {
		target++
	}
	if c.Control == "year" && c.State != "interactive" && year == target {
		a["data-component-state"] = c.State
	}
	return a
}
func (c ComponentCase) DateMonth(month int) templ.Attributes {
	a := c.DateBase().DateMonth(month)
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
func (c ComponentCase) DateMenuState(a templ.Attributes) templ.Attributes {
	if c.Control == "menu-button" && c.State != "interactive" {
		a["data-component-state"] = c.State
		if c.State == "disabled" {
			a["disabled"] = true
		}
	}
	return a
}
func (c ComponentCase) DateMenu() templ.Attributes { return c.DateMenuState(c.DateBase().DateMenu()) }
func (c ComponentCase) DateYearMenu() templ.Attributes {
	return c.DateMenuState(c.DateBase().DateYearMenu())
}
func (c ComponentCase) DateHeadline() string { return c.DateBase().DateHeadline() }

func (c ComponentCase) DatePickerProps() ui.DatePickerProps {
	p := ui.DatePickerProps{ID: c.ID(), Variant: c.Variant, Range: c.DateRange, Pane: c.Control, Inline: c.State != "interactive", Value: time.Date(2026, 10, 21, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC), Today: time.Date(2026, 10, 18, 0, 0, 0, 0, time.UTC), Attrs: c.Attributes(), PanelAttrs: c.DatePanel(), MenuAttrs: c.DateMenu(), YearMenuAttrs: c.DateYearMenu(), DayAttrs: map[string]templ.Attributes{}, YearAttrs: map[int]templ.Attributes{}, MonthAttrs: map[int]templ.Attributes{}}
	for _, month := range []int{10, 11} {
		for _, row := range DateRows(month) {
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

func DatePickerSampleDate(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
