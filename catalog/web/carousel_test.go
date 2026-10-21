package web

import "testing"

func TestCarouselItemStates(t *testing.T) {
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
