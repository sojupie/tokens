package playground

import (
	"fmt"
	"net/url"

	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
)

func ConfigureMenuPreview(page *webview.ComponentPreview, query url.Values) error {
	if page.Cases[0].Shape == "cascade" {
		return nil
	}
	mode := query.Get("preview.behavior")
	if mode == "" {
		mode = "actions"
	}
	if !ContainsPreview([]string{"actions", "single", "multiple"}, mode) {
		return fmt.Errorf("%w: menu behavior %s", ErrPreviewInput, mode)
	}
	page.Controls = append(page.Controls, webview.PreviewControl{Name: "behavior", Label: "Behavior", Selected: mode, Options: []string{"actions", "single", "multiple"}, Labels: map[string]string{"actions": "Actions", "single": "Single select", "multiple": "Multi select"}})
	for i := range page.Cases {
		page.Cases[i].MenuMode = mode
	}
	return nil
}
