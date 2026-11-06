package presentation

import (
	"fmt"
	"strconv"

	"github.com/a-h/templ"
)

func (c ComponentCase) TabAttributes(index int) templ.Attributes {
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

func (c ComponentCase) TabLabels() []string {
	if c.IllustrationLabels[0] != "" || c.TabCount == "" {
		labels := c.LabelsOr([3]string{"Home", "Files", "Saved"})
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
func (c ComponentCase) TabHasBadge(index int) bool {
	target := 0
	if c.ItemSelected(0) {
		target = 1
	}
	return index == target && c.TabBadge != "" && c.TabBadge != "none"
}
