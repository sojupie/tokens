package ui

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func TestCarouselItemContentAndLabel(t *testing.T) {
	p := CarouselItemProps{Index: 1, Count: 6, Label: "Coastal collection", AspectRatio: 4, Before: templ.Raw("<p>Above</p>"), After: templ.Raw("<button>Save</button>")}
	if p.Attributes()["aria-label"] != "Coastal collection, Item 2 of 6" {
		t.Fatal("custom label lost its position")
	}
	if p.slot()["style"] != "--_carousel-aspect-ratio:1.7777777777777777" {
		t.Fatal("aspect ratio exceeded the supported range")
	}
	var out bytes.Buffer
	if err := CarouselItem(p).Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	html := out.String()
	if strings.Index(html, "Above") > strings.Index(html, "md-carousel-item") || strings.Index(html, "Save") < strings.Index(html, "</button>") {
		t.Fatal("surrounding content is nested in the masked button")
	}
}

func TestCarouselFullScreenAndMaximum(t *testing.T) {
	p := CarouselProps{Variant: "full-screen", Orientation: "landscape", MaxItemWidth: 240}
	if p.Attributes()["data-orientation"] != "portrait" {
		t.Fatal("full-screen exposed landscape")
	}
	if p.Attributes()["style"] != "--_carousel-maximum:240px" {
		t.Fatal("maximum width missing")
	}
}

func TestCarouselMaximumWithConsumerStyle(t *testing.T) {
	p := CarouselProps{MaxItemWidth: 240, Attrs: templ.Attributes{"style": "margin:0"}}
	if p.Attributes()["style"] != "--_carousel-maximum:240px;margin:0" {
		t.Fatal("consumer style replaced maximum")
	}
}

func TestCarouselArtworkFraming(t *testing.T) {
	p := CarouselItemProps{ImageSrc: "/painting.webp", ImagePosition: "50% 100%", ImageScale: 1.2}
	var out bytes.Buffer
	if err := CarouselItem(p).Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"object-position:50% 100%", "transform:scale(1.2)", "transform-origin:50% 100%"} {
		if !strings.Contains(out.String(), value) {
			t.Fatalf("missing artwork framing: %s", value)
		}
	}
	if !strings.Contains((CarouselItemProps{}).imageStyle(), "scale(1)") {
		t.Fatal("default artwork scale changed")
	}
}
