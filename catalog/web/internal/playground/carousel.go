package playground

import (
	"fmt"
	"net/url"

	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
)

func ConfigureCarouselPreview(page *webview.ComponentPreview, query url.Values) error {
	action := query.Get("preview.action")
	if action == "" {
		action = "show-all"
	}
	options := []string{"none", "show-all", "header-arrow"}
	if !ContainsPreview(options, action) {
		return fmt.Errorf("%w: carousel action %s", ErrPreviewInput, action)
	}
	hasAction := false
	for _, c := range page.Cases {
		if c.Variant != "full-screen" {
			hasAction = true
		}
	}
	if hasAction {
		page.Controls = append(page.Controls, webview.PreviewControl{Name: "action", Label: "All items action", Selected: action, Options: options, Labels: map[string]string{"none": "None", "show-all": "Show all below", "header-arrow": "Header with arrow"}})
	}
	for i := range page.Cases {
		page.Cases[i].CarouselAction = action
	}
	scroll := query.Get("preview.scrolling")
	if scroll == "" {
		scroll = "free"
	}
	if !ContainsPreview([]string{"free", "snap"}, scroll) {
		return fmt.Errorf("%w: carousel scrolling %s", ErrPreviewInput, scroll)
	}
	hasUncontained := false
	for _, c := range page.Cases {
		if c.Variant == "uncontained" || c.Variant == "uncontained-multi-aspect-ratio" {
			hasUncontained = true
		}
	}
	if hasUncontained {
		page.Controls = append(page.Controls, webview.PreviewControl{Name: "scrolling", Label: "Gesture scrolling", Selected: scroll, Options: []string{"free", "snap"}, Labels: map[string]string{"free": "Free", "snap": "Snap"}})
	}
	for i := range page.Cases {
		page.Cases[i].CarouselScroll = scroll
	}
	content := query.Get("preview.content")
	if content == "" {
		content = "none"
	}
	if !ContainsPreview([]string{"none", "overlay", "above", "below"}, content) {
		return fmt.Errorf("%w: carousel content %s", ErrPreviewInput, content)
	}
	contentOptions := []string{"none", "overlay"}

	if !ContainsPreview(contentOptions, content) {
		content = "none"
	}
	page.Controls = append(page.Controls, webview.PreviewControl{Name: "content", Label: "Item text", Selected: content, Options: contentOptions, Labels: map[string]string{"none": "Image only", "overlay": "On image"}})
	viewport := query.Get("preview.viewport")
	if viewport == "" {
		viewport = "available"
	}
	widths := []string{"available", "360", "600", "840", "1200"}
	if !ContainsPreview(widths, viewport) {
		return fmt.Errorf("%w: carousel viewport %s", ErrPreviewInput, viewport)
	}
	page.Controls = append(page.Controls, webview.PreviewControl{Name: "viewport", Label: "Preview width · environment", Selected: viewport, Options: widths, Labels: map[string]string{"available": "Available width", "360": "360px", "600": "600px", "840": "840px", "1200": "1200px"}})
	for i := range page.Cases {
		page.Cases[i].CarouselContent = content
		page.Cases[i].CarouselViewport = viewport
	}
	screen := query.Get("preview.screen")
	if screen == "" {
		screen = "portrait"
	}
	if !ContainsPreview([]string{"landscape", "portrait"}, screen) {
		return fmt.Errorf("%w: carousel screen %s", ErrPreviewInput, screen)
	}
	for i := range page.Cases {
		if page.Cases[i].Variant == "full-screen" {
			page.Cases[i].Orientation = "portrait"
		}
	}

	return nil
}
