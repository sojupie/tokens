package web

import (
	"fmt"
	"github.com/a-h/templ"
	"net/url"
)

func (c componentCase) CarouselItem(index int) templ.Attributes {
	attrs := templ.Attributes{"class": "md-carousel-item", "type": "button", "data-carousel-index": fmt.Sprint(index)}
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
	hasAction := false
	for _, c := range page.Cases {
		if c.Variant != "full-screen" {
			hasAction = true
		}
	}
	if hasAction {
		page.Controls = append(page.Controls, previewControl{Name: "action", Label: "All items action", Selected: action, Options: options, Labels: map[string]string{"none": "None", "show-all": "Show all below", "header-arrow": "Header with arrow"}})
	}
	for i := range page.Cases {
		page.Cases[i].CarouselAction = action
	}
	scroll := query.Get("preview.scrolling")
	if scroll == "" {
		scroll = "free"
	}
	if !containsPreview([]string{"free", "snap"}, scroll) {
		return fmt.Errorf("%w: carousel scrolling %s", errPreviewInput, scroll)
	}
	hasUncontained := false
	for _, c := range page.Cases {
		if c.Variant == "uncontained" || c.Variant == "uncontained-multi-aspect-ratio" {
			hasUncontained = true
		}
	}
	if hasUncontained {
		page.Controls = append(page.Controls, previewControl{Name: "scrolling", Label: "Uncontained scrolling", Selected: scroll, Options: []string{"free", "snap"}, Labels: map[string]string{"free": "Free", "snap": "Snap"}})
	}
	for i := range page.Cases {
		page.Cases[i].CarouselScroll = scroll
	}
	screen := query.Get("preview.screen")
	if screen == "" {
		screen = "portrait"
	}
	if !containsPreview([]string{"landscape", "portrait"}, screen) {
		return fmt.Errorf("%w: carousel screen %s", errPreviewInput, screen)
	}
	for i := range page.Cases {
		if page.Cases[i].Variant == "full-screen" {
			page.Cases[i].Orientation = "portrait"
		}
	}

	return nil
}

// These artwork ratios are the four examples in the captured carousel specs.
func (c componentCase) CarouselSlot(index int) templ.Attributes {
	attrs := templ.Attributes{"class": "md-carousel-slot"}
	if c.Variant == "uncontained-multi-aspect-ratio" {
		ratios := []float64{16.0 / 9, 9.0 / 16, 1, 3.0 / 4}
		attrs["style"] = fmt.Sprintf("--_carousel-aspect-ratio:%g", ratios[index%len(ratios)])
	}
	return attrs
}

func (page componentPreview) HasFullScreenCarousel() bool {
	for _, c := range page.Cases {
		if c.Family == "carousel" && c.Variant == "full-screen" {
			return true
		}
	}
	return false
}
