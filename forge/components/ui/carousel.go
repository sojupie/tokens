package ui

import (
	"fmt"
	"github.com/a-h/templ"
)

type CarouselProps struct {
	Label, Variant, Orientation, Scrolling string
	Attrs                                  templ.Attributes
	Header, Actions, Controls              templ.Component
}

func (p CarouselProps) Attributes() templ.Attributes {
	return attributes(templ.Attributes{"class": "md-carousel", "aria-label": p.Label, "aria-roledescription": "carousel", "data-variant": fallback(p.Variant, "multi-browse"), "data-orientation": fallback(p.Orientation, "landscape"), "data-scrolling": fallback(p.Scrolling, "free")}, p.Attrs)
}

type CarouselItemProps struct {
	Index, Count              int
	Label, ImageSrc, ImageAlt string
	AspectRatio               float64
	SlotAttrs, Attrs          templ.Attributes
	Content                   templ.Component
}

func (p CarouselItemProps) slot() templ.Attributes {
	a := templ.Attributes{"class": "md-carousel-slot"}
	if p.AspectRatio > 0 {
		a["style"] = fmt.Sprintf("--_carousel-aspect-ratio:%g", p.AspectRatio)
	}
	return attributes(a, p.SlotAttrs)
}
func (p CarouselItemProps) Attributes() templ.Attributes {
	return attributes(templ.Attributes{"class": "md-carousel-item", "type": "button", "aria-label": fallback(p.Label, fmt.Sprintf("Item %d of %d", p.Index+1, p.Count)), "data-carousel-index": fmt.Sprint(p.Index)}, p.Attrs)
}
