package web

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/a-h/templ"
)

func (c componentCase) TabAttributes(index int) templ.Attributes {
	selected := c.ItemSelected(index)
	attrs := templ.Attributes{"class": "md-tab", "type": "button", "role": "tab", "id": fmt.Sprintf("%s-tab-%d", c.ID(), index), "aria-controls": fmt.Sprintf("%s-panel-%d", c.ID(), index), "aria-selected": fmt.Sprint(selected), "tabindex": "-1"}
	if c.TabHasBadge(index) {
		attrs["aria-label"] = c.TabLabels()[index] + " - 3 new updates"
		if c.TabBadge == "small" {
			attrs["aria-label"] = c.TabLabels()[index] + " - New updates"
		}
		attrs["data-tab-label"] = c.TabLabels()[index]
	}
	if selected {
		attrs["tabindex"] = "0"
	}
	if c.State != "interactive" {
		attrs["data-component-state"] = "enabled"
		if index == 0 {
			attrs["data-component-state"] = c.State
		}
	}
	return attrs
}

func configureTabsPreview(page *componentPreview, query url.Values) error {
	layout := query.Get("preview.layout")
	if layout == "" {
		layout = "fixed"
	}
	counts := []string{"3", "2", "4"}
	if layout == "scrollable" {
		counts = []string{"6", "3", "4", "5", "8"}
	}
	for _, choice := range []previewControl{
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
		if choice.Name == "count" && !containsPreview(choice.Options, choice.Selected) && containsPreview([]string{"2", "3", "4", "5", "6", "8"}, choice.Selected) {
			choice.Selected = choice.Options[0]
		}
		if !containsPreview(choice.Options, choice.Selected) {
			return fmt.Errorf("%w: %s %s", errPreviewInput, choice.Name, choice.Selected)
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
func (c componentCase) TabLabels() []string {
	if c.illustrationLabels[0] != "" || c.TabCount == "" {
		labels := c.labelsOr([3]string{"Home", "Files", "Saved"})
		return labels[:]
	}
	count, _ := strconv.Atoi(c.TabCount)
	labels := []string{"Overview", "Activity", "Files", "Shared", "Recent", "Saved"}
	if c.TabLayout == "scrollable" {
		labels = []string{"Recently added", "Shared with me", "My collections", "Downloads", "Saved items", "Offline files", "Archived projects", "Team resources"}
	}
	if count < 2 || count > len(labels) {
		count = 3
	}
	return labels[:count]
}
func (c componentCase) TabHasBadge(index int) bool {
	target := 0
	if c.ItemSelected(0) {
		target = 1
	}
	return index == target && c.TabBadge != "" && c.TabBadge != "none"
}
