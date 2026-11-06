package presentation

import (
	"fmt"

	"github.com/a-h/templ"
)

func (c ComponentCase) ListMediaKind() string {
	if c.ListMedia == "" {
		return "small"
	}
	return c.ListMedia
}
func (c ComponentCase) ListDividerKind() string {
	if c.ListDivider == "" {
		return "inset"
	}
	return c.ListDivider
}

// Selection lists expose one option per row. Nested check marks remain visual
// controls; the option owns focus and announces the row's selected state.
func (c ComponentCase) ListContainerAttributes() templ.Attributes {
	a := c.Attributes()
	if c.Control == "select" {
		a["role"] = "listbox"
		a["aria-label"] = "Single selection"
		a["aria-multiselectable"] = fmt.Sprint(c.Selection != "single")
		if c.Selection != "single" {
			a["aria-label"] = "Multiple selection"
		}
	}
	return a
}

func (c ComponentCase) ListNativeKind() string {
	if c.Width == "checkbox" || c.Width == "radio" {
		return c.Width
	}
	if c.ListTrailing == "checkbox" || c.ListTrailing == "radio" || c.ListTrailing == "switch" {
		return c.ListTrailing
	}
	return ""
}
func (c ComponentCase) ListSelected(index int) bool {
	return index == 0 && c.Selected || index == 1 && !c.Selected && (c.Selection == "single" || c.ListNativeKind() == "radio")
}
func (c ComponentCase) ListControl(index int) ComponentCase {
	state := "enabled"
	if c.State == "interactive" || index == 0 {
		state = c.State
	}
	return ComponentCase{Family: c.ListNativeKind(), Variant: "default", State: state, Selected: c.ListSelected(index)}
}
func (c ComponentCase) ListControlAttributes(index int) templ.Attributes {
	a := c.ListControl(index).Attributes()
	delete(a, "data-component-case")
	a["class"] = fmt.Sprint(a["class"]) + " md-list-selection-control"
	return a
}
func (c ComponentCase) ListControlInput(index int) templ.Attributes {
	kind := c.ListNativeKind()
	inputType := "checkbox"
	if kind == "radio" {
		inputType = "radio"
	}
	a := templ.Attributes{"type": inputType, "aria-label": c.ListLabel(index), "name": c.ID() + "-selection", "checked": c.ListSelected(index), "disabled": index == 0 && c.State == "disabled"}
	if kind == "switch" {
		a["role"] = "switch"
	}
	if c.Control == "select" {
		a["tabindex"] = "-1"
		a["aria-hidden"] = "true"
	}
	return a
}

func (c ComponentCase) ListLabel(index int) string {
	if index < len(c.IllustrationLabels) && c.IllustrationLabels[index] != "" {
		return c.IllustrationLabels[index]
	}
	return []string{"First item", "Second item", "Third item", "Nested item", "Another nested item"}[index]
}

type ListItemContent struct {
	Label, Supporting, Overline, Avatar, Metadata, Icon, Image, Badge string
}

func (c ComponentCase) ListItem(index int) templ.Attributes {
	return c.ListItemWithLabel(index, c.ListLabel(index))
}
func (c ComponentCase) ListItemWithLabel(index int, label string) templ.Attributes {
	attrs := templ.Attributes{"type": "button", "class": "md-list-item", "data-list-item": "", "data-list-label": label}
	if c.Control == "select" {
		attrs["role"] = "option"
		attrs["aria-label"] = label
		attrs["aria-selected"] = fmt.Sprint(c.ListSelected(index))
		attrs["tabindex"] = "-1"
		if index == 0 && c.State != "disabled" || index == 1 && c.State == "disabled" {
			attrs["tabindex"] = "0"
		}
	}
	if c.Control == "expand" && index < 3 {
		attrs["aria-expanded"] = fmt.Sprint(c.Expanded && index == 0)
		attrs["aria-controls"] = fmt.Sprintf("%s-children-%d", c.ID(), index)
	}
	if c.State != "interactive" {
		state := "enabled"
		if index == 0 {
			state = c.State
		}
		attrs["data-component-state"] = state
		attrs["data-button-state"] = state
		if state == "disabled" {
			attrs["disabled"] = true
		}
	}
	if c.ListNativeKind() != "" {
		delete(attrs, "type")
		delete(attrs, "aria-pressed")
		delete(attrs, "disabled")
		delete(attrs, "data-button-toggle")
		attrs["data-selected"] = fmt.Sprint(c.ListSelected(index))
		if index == 0 && c.State == "disabled" {
			attrs["aria-disabled"] = "true"
		}
	}
	return attrs
}

func (c ComponentCase) ListUtility(index int, action bool) templ.Attributes {
	class, label := "md-icon-button md-standard-icon-button md-list-grip", "Move "+c.ListLabel(index)
	key := "data-list-grip"
	if c.Control == "reveal" {
		class = "md-icon-button md-list-reveal-toggle"
		label = "Reveal actions for " + c.ListLabel(index)
		key = "data-list-reveal-toggle"
		if action {
			class = "md-icon-button md-list-reveal-action"
			label = "Mark " + c.ListLabel(index) + " done"
			key = "data-list-reveal-action"
		}
	}
	attrs := templ.Attributes{"type": "button", "class": class, "data-size": "small", "data-shape": "round", "data-width": "default", "aria-label": label, key: ""}
	if c.Control == "reorder" {
		attrs["aria-description"] = "Drag to reorder, or activate for Move up and Move down buttons."
		attrs["aria-expanded"] = "false"
		attrs["aria-controls"] = c.ID() + "-reorder"
		attrs["aria-haspopup"] = "dialog"
		attrs["data-sheet-open"] = ""
	}
	if c.Control == "reveal" && !action {
		attrs["aria-expanded"] = "false"
	}
	if c.State != "interactive" {
		state := "enabled"
		if index == 0 {
			state = c.State
		}
		attrs["data-component-state"] = state
		attrs["data-button-state"] = state
	}
	if c.State == "disabled" && index == 0 {
		attrs["disabled"] = true
	}
	return attrs
}

func (c ComponentCase) ListSupporting() string {
	if c.ListText == "wrapped" {
		return "Supporting text that wraps naturally as the available width changes."
	}
	if c.Size == "three-line" && c.ListText != "overline" {
		return "Supporting text\nContinued supporting text"
	}
	return "Supporting text"
}

func (c ComponentCase) ListMoveAction(index int, action string) templ.Attributes {
	attrs := templ.Attributes{"type": "button", "class": "md-list-item", "data-sheet-initial-focus": ""}
	attrs["data-list-move"] = action
	if c.State == "disabled" && index == 0 || action == "up" && index == 0 || action == "down" && index == 2 {
		attrs["disabled"] = true
	}
	return attrs
}

func (c ComponentCase) ListRowRole() string {
	if c.Control == "select" {
		return "presentation"
	}
	return "listitem"
}

func ListContentIcon(icon string) string {
	if icon == "" {
		return "folder"
	}
	return icon
}
func ListContentImage(image string, index int) string {
	if image == "" {
		return CarouselPaintings[index%len(CarouselPaintings)]
	}
	return image
}
