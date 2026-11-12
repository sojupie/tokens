package presentation

import (
	"strings"
)

type PreviewControl struct {
	Name, Label, Selected string
	Options               []string
	Labels                map[string]string
}
type PreviewHidden struct{ Name, Value string }
type ComponentPreview struct {
	Action      string
	Dimension   string
	Cases       []ComponentCase
	Controls    []PreviewControl
	Hidden      []PreviewHidden
	Interactive bool
}

func NavBarVariantLabels() map[string]string {
	return map[string]string{
		"adaptive":   "Flexible: automatic",
		"vertical":   "Icon above label",
		"horizontal": "Icon beside label",
		"baseline":   "Baseline (not recommended)",
	}
}

func (page ComponentPreview) Caption(c ComponentCase) string {
	if c.Family == "nav-bar" && page.Dimension == "variant" && !page.Interactive {
		return NavBarVariantLabels()[c.Variant]
	}
	if page.Interactive {
		return "Interactive"
	}
	if c.Family == "search" && page.Dimension == "control" {
		if c.Control == "baseline" {
			return "Divided (baseline)"
		}
		return "Contained"
	}
	if c.Family == "date" && page.Dimension == "variant" {
		return DateVariantLabel(c.Variant)
	}
	return PreviewLabel(PreviewDimensions(c)[page.Dimension])
}

func PreviewLabel(value string) string {
	switch value {
	case "uncontained-multi-aspect-ratio":
		return "Uncontained multi-aspect ratio"
	case "center-aligned-hero":
		return "Center-aligned hero"
	case "right-bottom":
		return "Right and bottom"
	case "scrolled":
		return "On scroll"
	case "floating-fab":
		return "Floating with FAB"
	case "medium-flexible":
		return "Medium flexible"
	case "large-flexible":
		return "Large flexible"
	case "baseline-small":
		return "Baseline small"
	case "baseline-centered":
		return "Baseline centered"
	case "baseline-medium":
		return "Baseline medium"
	case "baseline-large":
		return "Baseline large"
	case "legacy-linear":
		return "Baseline linear"
	case "legacy-circular":
		return "Baseline circular"
	case "four-color":
		return "Four colors"
	case "icon-label":
		return "Icon and label"
	case "value-indicator":
		return "Value indicator"
	case "single":
		return "Single"
	case "multiple":
		return "Multiple"
	case "xsmall":
		return "Extra small"
	case "xlarge":
		return "Extra large"
	case "field":
		return "Text field"
	case "select":
		return "Select"
	}
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

func PreviewDimensions(c ComponentCase) map[string]string {
	d := map[string]string{"variant": c.Variant, "size": c.Size, "shape": c.Shape, "width": c.Width, "state": c.State, "control": c.Family, "orientation": c.Orientation}
	if c.Family == "nav-rail" {
		d["orientation"] = ""

		if c.Variant == "expanded" {
			d["width"] = ""
		}
	}
	if c.Control != "" {
		d["control"] = c.Control
	}
	switch c.Family {
	case "button":
		d["selection"] = c.Selection
		if c.Selection == "" {
			d["selection"] = "plain"
		}
	case "icon", "radio", "menu", "checkbox", "switch", "nav-bar", "nav-rail", "nav-drawer", "tabs", "segmented", "toolbar":
		d["selection"] = "unselected"
		if c.Selected {
			d["selection"] = "selected"
		}
	}
	if c.Family == "menu" && c.Shape == "cascade" {
		d["disclosure"] = "collapsed"
		if c.Expanded {
			d["disclosure"] = "expanded"
		}
	}
	if c.Family == "dialog" {
		d["headline"] = "shown"
		if c.DialogHeadline == "hidden" {
			d["headline"] = "hidden"
		}
	}
	if c.Family == "dialog" && c.Variant != "full-screen" {
		d["mode"] = "short"
		if c.Expanded {
			d["mode"] = "long"
		}
	}
	if c.Family == "time" {
		d["selection"] = c.Selection
	}
	if c.Family == "date" {
		d["selection"] = c.Selection
		d["mode"] = "single"
		if c.DateRange {
			d["mode"] = "range"
		}
	}
	if c.Family == "search" {
		d["disclosure"] = "collapsed"
		if c.Expanded {
			d["disclosure"] = "expanded"
		}
	}
	if c.Family == "list" {
		if c.Control == "expand" {
			d["disclosure"] = "collapsed"
			if c.Expanded {
				d["disclosure"] = "expanded"
			}
		}
		if c.Selection != "" {
			d["mode"] = c.Selection
		}
		if c.Control == "select" || c.Control == "reorder" {
			d["selection"] = "unselected"
			if c.Selected {
				d["selection"] = "selected"
			}
		}
	}
	if c.Family == "button-group" && c.Control == "connected" {
		d["selection"] = "unselected"
		if c.Selected {
			d["selection"] = "selected"
		}
	}
	if c.Family == "checkbox" {
		d["variant"] = ""
		if c.Variant == "indeterminate" {
			d["selection"] = "indeterminate"
		}
	}
	if c.Family == "chip" {
		if c.Variant != "filter" {
			delete(d, "width")
		}
		if c.Variant == "filter" || c.Variant == "input" {
			d["selection"] = "unselected"
			if c.Selected {
				d["selection"] = "selected"
			}
		}
		if c.Variant == "input" {
			delete(d, "elevation")
		}
		if c.Variant != "input" {
			d["elevation"] = "flat"
			if c.Elevated {
				d["elevation"] = "elevated"
			}
		}
	}
	if c.Family == "field" || c.Family == "select" || c.Family == "autocomplete" || c.Family == "checkbox" {
		d["error"] = "normal"
		if c.Invalid {
			d["error"] = "error"
		}
	}
	return d
}

func DateVariantLabel(value string) string {
	return map[string]string{"docked": "Docked date picker", "modal": "Modal date picker", "input": "Modal date input", "full-screen": "Full-screen date picker"}[value]
}

func (c PreviewControl) OptionLabel(value string) string {
	if label := c.Labels[value]; label != "" {
		return label
	}
	return PreviewLabel(value)
}
