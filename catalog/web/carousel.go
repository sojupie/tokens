package web

import (
	"fmt"
	"github.com/a-h/templ"
	"net/url"
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

func configureCarouselPreview(page *componentPreview, query url.Values) error {
	action := query.Get("preview.action")
	if action == "" {
		action = "show-all"
	}
	options := []string{"none", "show-all", "header-arrow"}
	if !containsPreview(options, action) {
		return fmt.Errorf("%w: carousel action %s", errPreviewInput, action)
	}
	page.Controls = append(page.Controls, previewControl{Name: "action", Label: "All items action", Selected: action, Options: options, Labels: map[string]string{"none": "None", "show-all": "Show all below", "header-arrow": "Header with arrow"}})
	for i := range page.Cases {
		page.Cases[i].CarouselAction = action
	}
	return nil
}
