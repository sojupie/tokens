package playground

import (
	"fmt"
	"net/url"

	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
)

func ConfigureTabsPreview(page *webview.ComponentPreview, query url.Values) error {
	layout := query.Get("preview.layout")
	if layout == "" {
		layout = "fixed"
	}
	counts := []string{"3", "2", "4"}
	if layout == "scrollable" {
		counts = []string{"6", "3", "4", "5", "8"}
	}
	for _, choice := range []webview.PreviewControl{
		{Name: "layout", Label: "Layout", Options: []string{"fixed", "scrollable"}},
		{Name: "count", Label: "Tabs", Options: counts},
		{Name: "badge", Label: "Badge", Options: []string{"none", "small", "labeled"}},
		{Name: "alignment", Label: "Alignment", Options: []string{"center", "start"}},
		{Name: "direction", Label: "Direction", Options: []string{"ltr", "rtl"}},
	} {
		if choice.Name == "alignment" && layout == "scrollable" {
			continue
		}
		choice.Selected = query.Get("preview." + choice.Name)
		if choice.Selected == "" {
			choice.Selected = choice.Options[0]
		}
		if choice.Name == "count" && !ContainsPreview(choice.Options, choice.Selected) && ContainsPreview([]string{"2", "3", "4", "5", "6", "8"}, choice.Selected) {
			choice.Selected = choice.Options[0]
		}
		if !ContainsPreview(choice.Options, choice.Selected) {
			return fmt.Errorf("%w: %s %s", ErrPreviewInput, choice.Name, choice.Selected)
		}
		page.Controls = append(page.Controls, choice)
		for i := range page.Cases {
			c := &page.Cases[i]
			switch choice.Name {
			case "layout":
				c.TabLayout = choice.Selected
			case "count":
				c.TabCount = choice.Selected
			case "badge":
				c.TabBadge = choice.Selected
			case "alignment":
				c.TabAlignment = choice.Selected
			case "direction":
				c.TabDirection = choice.Selected
			}
		}
	}
	for i := range page.Controls {
		if page.Controls[i].Name == "selection" {
			page.Controls[i].Label = "First tab"
			page.Controls[i].Labels = map[string]string{"selected": "Active", "unselected": "Inactive"}
		}
	}
	return nil
}
