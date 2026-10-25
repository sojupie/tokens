package web

import "testing"

func TestDateRowsAndStates(t *testing.T) {
	rows := dateRows(10)
	if len(rows) != 5 || rows[0][0].ISO != "2026-09-27" || rows[4][6].ISO != "2026-10-31" {
		t.Fatal("October calendar must contain five complete weeks")
	}
	c := componentCase{Family: "date", Variant: "modal", DateRange: true, Control: "day", Selection: "in-range", State: "pressed"}
	for _, row := range rows {
		for _, day := range row {
			if day.ISO == "2026-10-23" {
				a := c.DateDay(day)
				if a["aria-selected"] != "true" || a["data-date-selected"] != "false" || a["data-in-range"] != "true" || a["data-component-state"] != "pressed" {
					t.Fatalf("range interior must retain its own palette: %#v", a)
				}
			}
		}
	}
	c.State = "interactive"
	if _, ok := c.DatePanel()["open"]; ok {
		t.Fatal("interactive picker must open through native dialog API")
	}
}

func TestDateEntryFields(t *testing.T) {
	c := componentCase{Family: "date", Variant: "docked", State: "interactive"}
	a := c.DateEntry("start")
	if a["data-date-value"] != "start" || a["value"] != "2026-10-21" {
		t.Fatalf("docked entry must be an editable calendar combobox: %#v", a)
	}
	c.Variant = "modal"
	c.DateRange = true
	if c.DateEntryLabel("start") != "Start date" || c.DateEntryLabel("end") != "End date" {
		t.Fatal("range fields must have distinct labels")
	}
	if c.DateEntry("end")["id"] == c.DateEntry("start")["id"] {
		t.Fatal("range fields must have distinct IDs")
	}
	if _, ok := c.DateEntry("start")["role"]; ok {
		t.Fatal("modal entry must remain a text field")
	}
}

func TestDockedDateNavigation(t *testing.T) {
	c := componentCase{Family: "date", Variant: "docked", Control: "month"}
	month, year := c.DateMenu(), c.DateYearMenu()
	if month["data-date-menu"] != "month" || year["data-date-menu"] != "year" || month["aria-controls"] == year["aria-controls"] {
		t.Fatal("docked month and year controls must target separate menus")
	}
	if month["aria-expanded"] != "true" || year["aria-expanded"] != "false" {
		t.Fatal("only the active menu is expanded")
	}
	if c.DateYear(2026)["class"] != "md-date-month-item" {
		t.Fatal("docked years must use captured docked menu-list styling")
	}
	c.Variant = "modal"
	if c.DateYear(2026)["class"] != "md-date-year" {
		t.Fatal("modal years must retain their grid styling")
	}
}

func TestDateRangeIsIndependentOfVariant(t *testing.T) {
	for _, variant := range []string{"docked", "modal", "input", "full-screen"} {
		c := componentCase{Family: "date", Variant: variant, DateRange: true, Control: "day", State: "interactive"}
		if c.DatePanel()["data-range"] != "true" || previewDimensions(c)["mode"] != "range" || c.DateEntryLabel("start") != "Start date" {
			t.Fatalf("range missing for %s", variant)
		}
		single := c
		single.DateRange = false
		if single.ID() == c.ID() {
			t.Fatal("single and range IDs collide")
		}
	}
}

func TestDatePreviewExposesEveryVariantForSingleAndRange(t *testing.T) {
	for _, rangeMode := range []bool{false, true} {
		var cases []componentCase
		for _, variant := range []string{"docked", "modal", "input", "full-screen"} {
			for _, r := range []bool{false, true} {
				cases = append(cases, componentCase{Family: "date", Variant: variant, DateRange: r, Control: "day", State: "enabled"})
			}
		}
		mode := "single"
		if rangeMode {
			mode = "range"
		}
		page, err := selectComponentPreview(cases, map[string][]string{"preview.compare": {"variants"}, "preview.mode": {mode}})
		if err != nil || len(page.Cases) != 4 {
			t.Fatalf("%s: %#v %v", mode, page, err)
		}
		for _, c := range page.Cases {
			if c.DateRange != rangeMode {
				t.Fatalf("selection changed for %s", c.Variant)
			}
		}
	}
}

func TestDateEntryPlaceholderMatchesPurpose(t *testing.T) {
	for _, isRange := range []bool{false, true} {
		c := componentCase{Family: "date", Variant: "docked", DateRange: isRange}
		for _, part := range []string{"start", "end"} {
			if c.DateEntry(part)["placeholder"] != c.DateEntryLabel(part) {
				t.Fatalf("%s input placeholder differs from its label", part)
			}
			if c.DateEntry(part)["aria-describedby"] != c.ID()+"-value-help-"+part {
				t.Fatal("Date format is not associated with the field")
			}
		}
	}
}
