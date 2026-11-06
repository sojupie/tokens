package playground

import (
	"fmt"
	"net/url"

	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
)

// Query content is independent from the action buttons.
func ConfigureSearchPreview(page *webview.ComponentPreview, query url.Values) error {
	content := query.Get("preview.content")
	if content == "" {
		content = "empty"
	}
	if content != "empty" && content != "text" {
		return fmt.Errorf("%w: search content %s", ErrPreviewInput, content)
	}
	elevation := query.Get("preview.shadow")
	if elevation == "" {
		elevation = "none"
	}
	if elevation != "none" && elevation != "token" {
		return fmt.Errorf("%w: search shadow %s", ErrPreviewInput, elevation)
	}
	page.Controls = append(page.Controls, webview.PreviewControl{Name: "shadow", Label: "Shadow", Options: []string{"none", "token"}, Selected: elevation, Labels: map[string]string{"none": "None (default)", "token": "Token elevation"}})
	page.Controls = append(page.Controls, webview.PreviewControl{Name: "content", Label: "Query", Options: []string{"empty", "text"}, Selected: content, Labels: map[string]string{"text": "Sample query"}})
	for i := range page.Controls {
		c := &page.Controls[i]
		if c.Name == "disclosure" {
			c.Label = "Focus"
			c.Labels = map[string]string{"collapsed": "Unfocused", "expanded": "Focused"}
		}
		if c.Name == "variant" {
			c.Label = "Layout"
			c.Labels = map[string]string{"adaptive": "Adaptive (flexible width)", "docked": "Docked", "full-screen": "Full-screen"}
		}
		if c.Name == "control" {
			c.Labels = map[string]string{"baseline": "Divided (baseline)", "contained": "Contained"}
		}
		if c.Name == "shape" {
			c.Labels = map[string]string{"icon": "Search icon", "menu": "Navigation button", "none": "None"}
		}
		if c.Name == "width" {
			c.Labels = map[string]string{"clear": "Clear", "avatar": "Avatar", "none": "None", "settings": "Search options", "clear-settings": "Clear + options", "settings-avatar": "Options + avatar"}
		}
	}
	for i := range page.Cases {
		page.Cases[i].SearchContent = content
		page.Cases[i].SearchElevation = elevation
	}
	return nil
}
