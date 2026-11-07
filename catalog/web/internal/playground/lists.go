package playground

import (
	"fmt"
	"net/url"

	"github.com/a-h/templ"
	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
)

func ConfigureListPreview(page *webview.ComponentPreview, query url.Values) error {
	c := page.Cases[0]
	choices := []webview.PreviewControl{{Name: "divider", Label: "Dividers", Options: []string{"inset", "full-width", "none"}}, {Name: "trailing", Label: "Trailing content", Options: []string{"text", "icon", "none"}}}
	if c.Control == "select" && c.Width != "checkbox" && c.Width != "radio" {
		if c.Selection == "single" {
			choices[1].Options = append(choices[1].Options, "radio")
		} else {
			choices[1].Options = append(choices[1].Options, "checkbox", "switch")
		}
	}
	if c.Control == "reveal" {
		choices[1].Options = []string{"text", "none"}
	}
	if c.Control == "expand" {
		choices = append(choices, webview.PreviewControl{Name: "inset", Label: "Child inset", Options: []string{"none", "inset"}}, webview.PreviewControl{Name: "surface", Label: "Expanded surface", Options: []string{"tonal", "none"}}, webview.PreviewControl{Name: "expand-icon", Label: "Disclosure icon", Options: []string{"contained", "plain"}})
	}
	if c.Width == "video" {
		choices = append(choices, webview.PreviewControl{Name: "media", Label: "Video size", Options: []string{"small", "large"}})
	}
	if c.Size != "one-line" {
		choices = append(choices, webview.PreviewControl{Name: "text", Label: "Text layout", Options: []string{"supporting", "wrapped", "overline"}})
	}
	for _, choice := range choices {
		choice.Selected = query.Get("preview." + choice.Name)
		if choice.Selected == "" {
			choice.Selected = choice.Options[0]
		}
		if !ContainsPreview(choice.Options, choice.Selected) {

			if choice.Name == "trailing" && ContainsPreview([]string{"icon", "checkbox", "radio", "switch"}, choice.Selected) {
				choice.Selected = choice.Options[0]
			} else {
				return fmt.Errorf("%w: %s %s", ErrPreviewInput, choice.Name, choice.Selected)
			}
		}
		page.Controls = append(page.Controls, choice)
		for i := range page.Cases {
			switch choice.Name {
			case "divider":
				page.Cases[i].ListDivider = choice.Selected
			case "trailing":
				page.Cases[i].ListTrailing = choice.Selected
			case "media":
				page.Cases[i].ListMedia = choice.Selected
			case "text":
				page.Cases[i].ListText = choice.Selected
			case "inset":
				page.Cases[i].ListInset = choice.Selected
			case "surface":
				page.Cases[i].ListSurface = choice.Selected
			case "expand-icon":
				page.Cases[i].ListExpandIcon = choice.Selected
			}
		}
	}
	if page.Cases[0].ListNativeKind() == "radio" {
		for i := range page.Cases {
			page.Cases[i].Selection = "single"
		}
		for i := range page.Controls {
			if page.Controls[i].Name == "mode" {
				page.Controls[i].Selected = "single"
				page.Controls[i].Options = []string{"single"}
			}
		}
	}
	return nil
}

func ContentListCase(id string) webview.ComponentCase {
	return webview.ComponentCase{Family: "list", Variant: "standard", Shape: "expressive", Size: "two-line", Control: "none", ListDivider: "none", State: "interactive", InstanceID: id}
}
func ContentListItem(c webview.ComponentCase, index int, label string) templ.Attributes {
	a := c.ListItemWithLabel(index, label)
	delete(a, "type")
	return a
}

func ContentListAttributes(c webview.ComponentCase, label string) templ.Attributes {
	a := c.ListContainerAttributes()
	a["aria-label"] = label
	return a
}
