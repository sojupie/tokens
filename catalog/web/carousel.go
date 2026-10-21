package web

import (
	"fmt"
	"github.com/a-h/templ"
)

func (c componentCase) CarouselItem(index int) templ.Attributes {
	attrs := templ.Attributes{"class": "md-carousel-item", "type": "button", "aria-label": fmt.Sprintf("Item %d of 6", index+1), "data-carousel-index": fmt.Sprint(index)}
	if c.State != "interactive" && index == 0 {
		attrs["data-component-state"] = c.State
		if c.State == "disabled" {
			attrs["disabled"] = true
		}
	}
	return attrs
}
