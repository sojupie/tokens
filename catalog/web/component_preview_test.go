package web

import (
	"errors"
	"net/url"
	"testing"
)

func TestPreviewVariesOneDimensionAndKeepsValidCombinations(t *testing.T) {
	var cases []componentCase
	for _, variant := range []string{"tonal", "text"} {
		for _, size := range []string{"small", "large"} {
			for _, selection := range []string{"", "selected"} {
				if variant == "text" && selection != "" {
					continue
				}
				for _, state := range []string{"enabled", "hovered", "disabled"} {
					cases = append(cases, componentCase{Family: "button", Variant: variant, Size: size, Shape: "round", Selection: selection, State: state})
				}
			}
		}
	}
	for _, tc := range []struct {
		query string
		count int
	}{
		{"", 2}, {"preview.compare=sizes", 2}, {"preview.compare=variants", 2},
		{"preview.compare=variants&preview.selection=selected", 1},
		{"preview.variant=text&preview.selection=selected", 1},
		{"preview.compare=interactive&preview.state=disabled", 1},
	} {
		query, _ := url.ParseQuery(tc.query)
		page, err := selectComponentPreview(cases, query)
		if err != nil {
			t.Fatal(err)
		}
		if len(page.Cases) != tc.count {
			t.Fatalf("%s: got %d specimens, want %d", tc.query, len(page.Cases), tc.count)
		}
		for _, c := range page.Cases {
			if c.Variant == "text" && c.Selection != "" {
				t.Fatal("invented unsupported selected text button")
			}
			if page.Interactive && c.State != "interactive" {
				t.Fatal("interactive case has forced state")
			}
		}
	}
	page, err := selectComponentPreview(cases, url.Values{"q": {"shadow"}, "context.Theme": {"Dark"}, "preview.compare": {"sizes"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Hidden) != 2 {
		t.Fatalf("lost token table filters: %#v", page.Hidden)
	}
	for _, query := range []url.Values{{"preview.compare": {"all"}}, {"preview.size": {"all"}}} {
		if _, err := selectComponentPreview(cases, query); !errors.Is(err, errPreviewInput) {
			t.Fatalf("unsupported value accepted: %v", err)
		}
	}
}

func TestCheckboxPreviewSelectsIndeterminateWithoutRedundantVariant(t *testing.T) {
	cases := []componentCase{
		{Family: "checkbox", Variant: "unchecked", State: "enabled"},
		{Family: "checkbox", Variant: "checked", Selected: true, State: "enabled"},
		{Family: "checkbox", Variant: "indeterminate", Selected: true, State: "enabled"},
	}
	page, err := selectComponentPreview(cases, url.Values{"preview.selection": {"indeterminate"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Cases) != 1 || page.Cases[0].Variant != "indeterminate" {
		t.Fatalf("wrong checkbox selection: %#v", page.Cases)
	}
	for _, control := range page.Controls {
		if control.Name == "variant" {
			t.Fatal("checkbox selection duplicated as variant")
		}
	}
}

func TestSnackbarPreviewFocusesOnlyTheChosenControl(t *testing.T) {
	c := componentCase{Family: "snackbar", Control: "action", State: "focused"}
	if c.FeedbackAttributes("md-button md-feedback-action")["data-component-state"] != "focused" || c.FeedbackAttributes("md-icon-button md-snackbar-close")["data-component-state"] != "enabled" {
		t.Fatal("snackbar forced two controls into focus")
	}
	c.Control = "dismiss"
	if c.FeedbackAttributes("md-button md-feedback-action")["data-component-state"] != "enabled" || c.FeedbackAttributes("md-icon-button md-snackbar-close")["data-component-state"] != "focused" {
		t.Fatal("dismiss control does not own focus")
	}
	c.State = "interactive"
	if _, exists := c.FeedbackAttributes("md-button md-feedback-action")["data-component-state"]; exists {
		t.Fatal("interactive action has a forced state")
	}
}

func TestNavigationPreviewPreservesOrientationAndNativeSelection(t *testing.T) {
	c := componentCase{Family: "nav-rail", Variant: "expanded", Orientation: "horizontal", State: "focused", Selected: true}
	if c.NavigationItem(0)["aria-current"] != "page" || c.NavigationItem(1)["aria-current"] != nil {
		t.Fatal("navigation selection does not match the preview")
	}
	if c.NavigationItem(0)["data-component-state"] != "focused" || c.NavigationItem(1)["data-component-state"] != "enabled" {
		t.Fatal("only the target item should have a forced state")
	}
	c.State = "interactive"
	if c.NavigationItem(0)["data-component-state"] != nil {
		t.Fatal("interactive navigation retains a forced state")
	}
	if c.Attributes()["data-orientation"] != "horizontal" {
		t.Fatal("missing navigation orientation")
	}
	other := c
	other.Orientation = "vertical"
	if c.ID() == other.ID() {
		t.Fatal("navigation orientations have duplicate IDs")
	}
}

func TestTabsAndSegmentsUseNativeSelectionAndLinkedPanels(t *testing.T) {
	c := componentCase{Family: "tabs", Variant: "primary", Control: "label", State: "focused", Selected: true}
	first, second := c.TabAttributes(0), c.TabAttributes(1)
	if first["aria-selected"] != "true" || first["tabindex"] != "0" || second["aria-selected"] != "false" || second["tabindex"] != "-1" {
		t.Fatal("tabs must have one selected tab in the tab sequence")
	}
	if first["aria-controls"] == second["aria-controls"] {
		t.Fatal("tabs share a panel ID")
	}
	c.Family = "segmented"
	c.Variant = "single"
	if c.SegmentInput(0, "Home")["type"] != "radio" || c.SegmentInput(0, "Home")["checked"] != true {
		t.Fatal("single selection must use native radios")
	}
	c.Variant = "multiple"
	c.State = "disabled"
	if c.SegmentInput(0, "Home")["type"] != "checkbox" || c.SegmentInput(0, "Home")["disabled"] != true {
		t.Fatal("multiple selection must use native checkboxes and native disabled state")
	}
	c.State = "interactive"
	if c.SegmentAttributes(0)["data-component-state"] != nil || c.TabAttributes(0)["data-component-state"] != nil {
		t.Fatal("interactive controls retain a forced state")
	}
}

func TestSliderPreviewUsesNativeRangesAndIndependentHandles(t *testing.T) {
	c := componentCase{Family: "slider", Control: "range", Variant: "discrete", Orientation: "vertical", State: "pressed"}
	if len(c.SliderHandles()) != 2 || c.SliderInput("start")["step"] != "10" || c.SliderInput("start")["orient"] != "vertical" {
		t.Fatal("range preview lost its native controls")
	}
	if c.SliderHandle("start")["data-component-state"] != "pressed" || c.SliderHandle("end")["data-component-state"] != "enabled" {
		t.Fatal("preview pressed both range handles")
	}
	c.State = "disabled"
	if c.SliderInput("start")["disabled"] != true || c.SliderInput("end")["disabled"] != true {
		t.Fatal("disabled ranges must disable both inputs")
	}
	c.Control = "single"
	c.State = "interactive"
	c.Variant = "continuous"
	if len(c.SliderHandles()) != 1 || c.SliderInput("end")["step"] != "1" || c.SliderHandle("end")["data-component-state"] != nil {
		t.Fatal("interactive slider has an invented forced state")
	}
}

func TestPreviewDefaultsToVariantsAndUnknownProgress(t *testing.T) {
	for _, family := range []string{"progress", "loading"} {
		var cases []componentCase
		for _, control := range []string{"determinate", "indeterminate"} {
			for _, variant := range []string{"first", "second"} {
				cases = append(cases, componentCase{Family: family, Control: control, Variant: variant, State: "enabled"})
			}
		}
		for _, control := range []string{"", "determinate"} {
			query := url.Values{}
			if control != "" {
				query.Set("preview.control", control)
			}
			page, err := selectComponentPreview(cases, query)
			if err != nil {
				t.Fatal(err)
			}
			want := control
			if want == "" {
				want = "indeterminate"
			}
			if page.Dimension != "variant" || len(page.Cases) != 2 {
				t.Fatalf("%s: expected two variants: %#v", family, page)
			}
			for _, c := range page.Cases {
				if c.Control != want {
					t.Fatalf("%s: got %s, want %s", family, c.Control, want)
				}
			}
		}
	}
	page, err := selectComponentPreview([]componentCase{{Family: "radio", State: "enabled"}, {Family: "radio", State: "hovered"}}, url.Values{})
	if err != nil || page.Dimension != "state" || len(page.Cases) != 2 {
		t.Fatalf("single variant must fall back to states: %#v, %v", page, err)
	}
}

func TestPreviewDefaultsToMediumOrDefaultSize(t *testing.T) {
	for _, tc := range []struct {
		name   string
		family string
		sizes  []string
		query  url.Values
		want   string
	}{
		{"medium", "button", []string{"small", "medium", "large"}, nil, "medium"},
		{"default", "carousel", []string{"small", "default", "large"}, nil, "default"},
		{"slider", "slider", []string{"xsmall", "small", "medium", "large"}, nil, "medium"},
		{"explicit", "button", []string{"small", "medium", "large"}, url.Values{"preview.size": {"small"}}, "small"},
		{"fallback", "button", []string{"small", "large"}, nil, "small"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var cases []componentCase
			for _, size := range tc.sizes {
				cases = append(cases, componentCase{Family: tc.family, Size: size, State: "enabled"})
			}
			page, err := selectComponentPreview(cases, tc.query)
			if err != nil {
				t.Fatal(err)
			}
			if len(page.Cases) != 1 || page.Cases[0].Size != tc.want {
				t.Fatalf("got %#v, want size %s", page.Cases, tc.want)
			}
		})
	}
}
