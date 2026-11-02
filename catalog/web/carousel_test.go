package web

import "testing"

func TestCarouselItemStates(t *testing.T) {
	if materialSymbolSVG("arrow_forward") == "" {
		t.Fatal("header action arrow must be available")
	}
	if materialSymbolSVG("chevron_left") == "" {
		t.Fatal("previous-item icon must be available")
	}
	c := componentCase{Family: "carousel", State: "disabled"}
	first, second := c.CarouselItem(0), c.CarouselItem(1)
	if first["disabled"] != true || first["data-component-state"] != "disabled" {
		t.Fatal("forced disabled state must apply to the first item")
	}
	if _, ok := second["disabled"]; ok {
		t.Fatal("neighboring items must remain enabled")
	}
	c.State = "interactive"
	if _, ok := c.CarouselItem(0)["data-component-state"]; ok {
		t.Fatal("interactive items must use native interaction states")
	}
	if _, ok := c.Attributes()["aria-disabled"]; ok {
		t.Fatal("a disabled specimen must not disable the whole carousel")
	}
}

func TestCarouselAspectRatioExamples(t *testing.T) {
	c := componentCase{Family: "carousel", Variant: "uncontained-multi-aspect-ratio"}
	seen := map[string]bool{}
	for i := range 4 {
		seen[c.CarouselSlot(i)["style"].(string)] = true
	}
	if len(seen) != 4 {
		t.Fatal("multi-aspect variant must demonstrate four distinct source ratios")
	}
	c.Variant = "uncontained"
	if _, ok := c.CarouselSlot(0)["style"]; ok {
		t.Fatal("ordinary uncontained items must keep a single width")
	}
}

func TestCarouselScrollingAndScreenOptions(t *testing.T) {
	page := &componentPreview{Cases: []componentCase{{Family: "carousel", Variant: "uncontained-multi-aspect-ratio"}, {Family: "carousel", Variant: "full-screen"}}}
	if err := configureCarouselPreview(page, map[string][]string{"preview.scrolling": {"snap"}, "preview.screen": {"portrait"}}); err != nil {
		t.Fatal(err)
	}
	if page.Cases[0].CarouselScroll != "snap" || page.Cases[1].Orientation != "portrait" {
		t.Fatal("explicit configuration lost")
	}
	if err := configureCarouselPreview(page, map[string][]string{"preview.screen": {"other"}}); err == nil {
		t.Fatal("invalid frame accepted")
	}
	if err := configureCarouselPreview(page, map[string][]string{"preview.scrolling": {"other"}}); err == nil {
		t.Fatal("invalid scrolling accepted")
	}
}

func TestCarouselControlsOnlyApplyToVisibleLayouts(t *testing.T) {
	for _, variant := range []string{"hero", "full-screen"} {
		page := &componentPreview{Cases: []componentCase{{Family: "carousel", Variant: variant}}}
		if err := configureCarouselPreview(page, nil); err != nil {
			t.Fatal(err)
		}
		for _, c := range page.Controls {
			if c.Name == "scrolling" || c.Name == "screen" {
				t.Fatalf("%s exposes irrelevant %s", variant, c.Name)
			}
		}
		if variant == "full-screen" && page.Cases[0].Orientation != "portrait" {
			t.Fatal("full screen defaults to landscape")
		}
	}
}
