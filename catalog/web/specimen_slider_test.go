package web

import (
	"bytes"
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
)

func TestSliderPreviewConfigurations(t *testing.T) {
	cases := []componentCase{
		{Family: "slider", Variant: "continuous", Control: "single", Size: "medium", Shape: "label", State: "enabled"},
		{Family: "slider", Variant: "discrete", Control: "single", Size: "medium", Shape: "label", State: "enabled"},
	}
	page, err := selectComponentPreview(cases, url.Values{"preview.icon": {"volume"}, "preview.value": {"30"}, "preview.step": {"20"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, control := range page.Controls {
		if control.Name == "shape" {
			t.Fatal("token families exposed as bubble variants")
		}
	}
	for _, c := range page.Cases {
		var out bytes.Buffer
		if err := specimenSlider(c).Render(context.Background(), &out); err != nil {
			t.Fatal(err)
		}
		markup := out.String()
		if strings.Contains(markup, `md-slider-trailing`) || strings.Contains(markup, `@materialSymbol`) {
			t.Fatal("unexpected endpoint icon")
		}
		if c.Variant == "continuous" {
			if c.SliderInsetIcon() != "volume_up" || !strings.Contains(markup, `data-slider-inset`) {
				t.Fatal("missing meaningful single inset icon")
			}
		} else {
			if c.SliderInsetIcon() != "" || c.SliderValue("end") != "40" || len(c.SliderTicks()) != 4 || c.SliderInput("end")["step"] != "20" {
				t.Fatalf("invalid discrete configuration: %#v", c)
			}
		}
		if c.SliderDisplay != "always" {
			t.Fatal("preview should reveal the value label by default")
		}
	}
	_, err = selectComponentPreview(cases, url.Values{"preview.step": {"0"}})
	if !errors.Is(err, errPreviewInput) {
		t.Fatalf("invalid step accepted: %v", err)
	}
}
func TestSliderInsetIconRestrictions(t *testing.T) {
	for _, c := range []componentCase{
		{Variant: "continuous", Control: "range", Size: "medium", SliderIcon: "volume"},
		{Variant: "continuous", Control: "single", Size: "small", SliderIcon: "volume"},
		{Variant: "discrete", Control: "single", Size: "large", SliderIcon: "volume"},
	} {
		if c.SliderInsetIcon() != "" {
			t.Fatalf("unsupported inset icon: %#v", c)
		}
	}
	page := componentPreview{Cases: []componentCase{{Variant: "continuous", Control: "range", Size: "large"}}}
	if err := configureSliderPreview(&page, url.Values{"preview.lower": {"90"}, "preview.value": {"100"}}); err != nil {
		t.Fatal(err)
	}
	if page.Cases[0].SliderValue("start") != "90" || page.Cases[0].SliderValue("end") != "100" {
		t.Fatal("numeric range corrupted")
	}
	for _, c := range page.Controls {
		if c.Name == "icon" {
			t.Fatal("inset control offered for range slider")
		}
	}
}
