package ui

import (
	"fmt"

	"github.com/a-h/templ"
)

type CarouselProps struct {
	Label, Variant, Orientation, Scrolling string
	// MaxItemWidth limits the large item; the available region may fit more items.
	MaxItemWidth              float64
	Attrs                     templ.Attributes
	Header, Actions, Controls templ.Component
}

func (p CarouselProps) Attributes() templ.Attributes {
	a := templ.Attributes{"class": "md-carousel", "aria-label": p.Label, "aria-roledescription": "carousel", "data-variant": fallback(p.Variant, "multi-browse"), "data-orientation": fallback(p.Orientation, "landscape"), "data-scrolling": fallback(p.Scrolling, "free")}
	if p.Variant == "full-screen" {
		a["data-orientation"] = "portrait"
	}
	if p.MaxItemWidth > 0 {
		a["style"] = fmt.Sprintf("--_carousel-maximum:%gpx", p.MaxItemWidth)
	}
	result := attributes(a, p.Attrs)
	if p.MaxItemWidth > 0 {
		if custom, ok := p.Attrs["style"]; ok {
			result["style"] = fmt.Sprintf("--_carousel-maximum:%gpx;%v", p.MaxItemWidth, custom)
		}
	}
	return result
}

type CarouselItemProps struct {
	Index, Count              int
	Label, ImageSrc, ImageAlt string
	// ImageScale zooms artwork around its focal point; zero defaults to 1.
	ImageScale float64
	// ImagePosition selects the focal point within cover-cropped artwork.
	ImagePosition    string
	AspectRatio      float64
	SlotAttrs, Attrs templ.Attributes
	Content          templ.Component
	// Before and After are unmasked caption slots for horizontal carousels.
	Before, After templ.Component
}

func (p CarouselItemProps) slot() templ.Attributes {
	a := templ.Attributes{"class": "md-carousel-slot"}
	if p.AspectRatio > 0 {
		a["style"] = fmt.Sprintf("--_carousel-aspect-ratio:%g", max(9.0/16, min(16.0/9, p.AspectRatio)))
	}
	return attributes(a, p.SlotAttrs)
}
func (p CarouselItemProps) Attributes() templ.Attributes {
	label := fmt.Sprintf("Item %d of %d", p.Index+1, p.Count)
	if p.Label != "" {
		label = p.Label + ", " + label
	}
	return attributes(templ.Attributes{"class": "md-carousel-item", "type": "button", "aria-label": label, "data-carousel-index": fmt.Sprint(p.Index)}, p.Attrs)
}

func (p CarouselItemProps) imageStyle() string {
	scale := p.ImageScale
	if scale <= 0 {
		scale = 1
	}
	position := fallback(p.ImagePosition, "50% 50%")
	return fmt.Sprintf("object-position:%s;transform:scale(%g);transform-origin:%s", position, scale, position)
}
