package playground

import (
	"fmt"
	"net/url"

	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
)

func ConfigureBadgePreview(page *webview.ComponentPreview, query url.Values) error {
	for _, choice := range []webview.PreviewControl{
		{Name: "content", Label: "Notification", Options: []string{"99+", "1", "12", "999", "1000", "New"}},
		{Name: "placement", Label: "Placement", Options: []string{"icon", "trailing"}},
		{Name: "direction", Label: "Direction", Options: []string{"ltr", "rtl"}},
	} {
		choice.Selected = query.Get("preview." + choice.Name)
		if choice.Selected == "" {
			choice.Selected = choice.Options[0]
		}
		if !ContainsPreview(choice.Options, choice.Selected) {
			return fmt.Errorf("%w: %s %s", ErrPreviewInput, choice.Name, choice.Selected)
		}
		page.Controls = append(page.Controls, choice)
		for i := range page.Cases {
			switch choice.Name {
			case "content":
				page.Cases[i].BadgeContent = choice.Selected
			case "placement":
				page.Cases[i].BadgePlacement = choice.Selected
			case "direction":
				page.Cases[i].BadgeDirection = choice.Selected
			}
		}
	}
	return nil
}
