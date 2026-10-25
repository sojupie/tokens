package web

import (
	"fmt"
	"github.com/a-h/templ"
	"net/url"
)

func configureListPreview(page *componentPreview, query url.Values) error {
	c := page.Cases[0]
	choices := []previewControl{{Name: "divider", Label: "Dividers", Options: []string{"inset", "full-width", "none"}}, {Name: "trailing", Label: "Trailing content", Options: []string{"text", "icon", "none"}}}
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
		choices = append(choices, previewControl{Name: "inset", Label: "Child inset", Options: []string{"none", "inset"}}, previewControl{Name: "surface", Label: "Expanded surface", Options: []string{"tonal", "none"}}, previewControl{Name: "expand-icon", Label: "Disclosure icon", Options: []string{"contained", "plain"}})
	}
	if c.Width == "video" {
		choices = append(choices, previewControl{Name: "media", Label: "Video size", Options: []string{"small", "large"}})
	}
	if c.Size != "one-line" {
		choices = append(choices, previewControl{Name: "text", Label: "Text layout", Options: []string{"supporting", "wrapped", "overline"}})
	}
	for _, choice := range choices {
		choice.Selected = query.Get("preview." + choice.Name)
		if choice.Selected == "" {
			choice.Selected = choice.Options[0]
		}
		if !containsPreview(choice.Options, choice.Selected) {
			// Changing the leading control can remove trailing-control choices.
			if choice.Name == "trailing" && containsPreview([]string{"icon", "checkbox", "radio", "switch"}, choice.Selected) {
				choice.Selected = choice.Options[0]
			} else {
				return fmt.Errorf("%w: %s %s", errPreviewInput, choice.Name, choice.Selected)
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
func (c componentCase) listMedia() string {
	if c.ListMedia == "" {
		return "small"
	}
	return c.ListMedia
}
func (c componentCase) listDivider() string {
	if c.ListDivider == "" {
		return "inset"
	}
	return c.ListDivider
}

// Selection lists expose one option per row. Nested check marks remain visual
// controls; the option owns focus and announces the row's selected state.
func (c componentCase) ListContainerAttributes() templ.Attributes {
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

func (c componentCase) ListNativeKind() string {
	if c.Width == "checkbox" || c.Width == "radio" {
		return c.Width
	}
	if c.ListTrailing == "checkbox" || c.ListTrailing == "radio" || c.ListTrailing == "switch" {
		return c.ListTrailing
	}
	return ""
}
func (c componentCase) ListSelected(index int) bool {
	return index == 0 && c.Selected || index == 1 && !c.Selected && (c.Selection == "single" || c.ListNativeKind() == "radio")
}
func (c componentCase) ListControl(index int) componentCase {
	state := "enabled"
	if c.State == "interactive" || index == 0 {
		state = c.State
	}
	return componentCase{Family: c.ListNativeKind(), Variant: "default", State: state, Selected: c.ListSelected(index)}
}
func (c componentCase) ListControlAttributes(index int) templ.Attributes {
	a := c.ListControl(index).Attributes()
	delete(a, "data-component-case")
	a["class"] = fmt.Sprint(a["class"]) + " md-list-selection-control"
	return a
}
func (c componentCase) ListControlInput(index int) templ.Attributes {
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

func (c componentCase) ListLabel(index int) string {
	return []string{"First item", "Second item", "Third item", "Nested item", "Another nested item"}[index]
}

type listItemContent struct {
	Label, Supporting, Overline, Avatar, Metadata string
}

func (c componentCase) ListItem(index int) templ.Attributes {
	return c.ListItemWithLabel(index, c.ListLabel(index))
}
func (c componentCase) ListItemWithLabel(index int, label string) templ.Attributes {
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

func (c componentCase) ListUtility(index int, action bool) templ.Attributes {
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

func (c componentCase) ListSupporting() string {
	if c.ListText == "wrapped" {
		return "Supporting text that wraps naturally as the available width changes."
	}
	if c.Size == "three-line" && c.ListText != "overline" {
		return "Supporting text\nContinued supporting text"
	}
	return "Supporting text"
}

func (c componentCase) ListMoveAction(index int, action string) templ.Attributes {
	attrs := templ.Attributes{"type": "button", "class": "md-list-item", "data-sheet-initial-focus": ""}
	attrs["data-list-move"] = action
	if c.State == "disabled" && index == 0 || action == "up" && index == 0 || action == "down" && index == 2 {
		attrs["disabled"] = true
	}
	return attrs
}

func (c componentCase) ListRowRole() string {
	if c.Control == "select" {
		return "presentation"
	}
	return "listitem"
}
