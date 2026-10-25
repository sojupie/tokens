package web

import (
	"fmt"
	"strings"

	"github.com/a-h/templ"
)

func (c componentCase) Attributes() templ.Attributes {
	if c.Family == "button" {
		return c.button().Attributes()
	}
	class := map[string]string{
		"card":         "md-card md-" + c.Variant + "-card md-interactive-card",
		"icon":         "md-icon-button md-" + c.Variant + "-icon-button",
		"chip":         "md-filter-chip",
		"radio":        "md-radio",
		"field":        "md-outlined-text-field",
		"select":       "md-outlined-select",
		"menu":         "md-menu-item",
		"dialog":       "md-dialog",
		"divider":      "md-divider",
		"checkbox":     "md-checkbox",
		"switch":       "md-switch",
		"fab":          "md-button md-fab",
		"extended-fab": "md-button md-extended-fab",
		"badge":        "md-badge",
		"snackbar":     "md-snackbar",
		"tooltip":      "md-tooltip",
		"nav-bar":      "md-navigation md-navigation-bar",
		"nav-rail":     "md-navigation md-navigation-rail",
		"nav-drawer":   "md-navigation md-navigation-drawer",
		"tabs":         "md-tabs-example",
		"segmented":    "md-segmented-buttons",
		"slider":       "md-slider",
		"progress":     "md-progress",
		"loading":      "md-loading",
		"appbar":       "md-app-bar",
		"toolbar":      "md-toolbar",
		"bottom-sheet": "md-sheet md-bottom-sheet",
		"side-sheet":   "md-sheet md-side-sheet",
		"split":        "md-split-button",
		"button-group": "md-button-group",
		"fab-menu":     "md-fab-menu",
		"list":         "md-list",
		"carousel":     "md-carousel",
		"search":       "md-search-demo",
		"date":         "md-date-demo",
		"time":         "md-time-demo",
	}[c.Family]
	if c.Family == "chip" && c.Variant != "" {
		class = "md-" + c.Variant + "-chip"
	}
	if c.Family == "field" || c.Family == "select" || c.Family == "autocomplete" {
		variant := c.Variant
		if variant == "" {
			variant = "outlined"
		}
		kind := c.Family
		if kind == "field" {
			kind = "text-field"
		}
		class = "md-" + variant + "-" + kind
	}
	if c.Family == "dialog" && c.Variant == "full-screen" {
		class = "md-full-screen-dialog"
	}
	attrs := templ.Attributes{"class": class, "data-component-case": c.ID(), "data-family": c.Family, "data-component-state": c.State}
	if c.Family == "fab" || c.Family == "extended-fab" {
		attrs["data-variant"] = c.Variant
		attrs["type"] = "button"
		attrs["aria-label"] = "Create"
	}
	if c.Family == "tabs" || c.Family == "segmented" || c.Family == "slider" || c.Family == "progress" || c.Family == "loading" || c.Family == "appbar" || c.Family == "toolbar" || c.Family == "bottom-sheet" || c.Family == "side-sheet" || c.Family == "split" || c.Family == "button-group" || c.Family == "fab-menu" || c.Family == "list" || c.Family == "carousel" || c.Family == "search" || c.Family == "date" || c.Family == "time" {
		attrs["data-orientation"] = c.Orientation
		attrs["data-variant"] = c.Variant
		attrs["data-control"] = c.Control
	}
	if c.Family == "segmented" {
		attrs["data-density"] = c.SegmentDensity
		if c.SegmentDensity == "" {
			attrs["data-density"] = "0"
		}
	}
	if c.Family == "search" {
		attrs["data-disclosure"] = fmt.Sprint(c.Expanded)
		attrs["data-interactive"] = fmt.Sprint(c.State == "interactive")
	}
	if c.Family == "divider" {
		attrs["data-inset"] = c.DividerInset()
	}
	if c.Family == "list" {
		attrs["data-size"] = c.Size
		attrs["data-selection-mode"] = c.Selection
		attrs["data-media"] = c.listMedia()
		attrs["data-divider"] = c.listDivider()
		attrs["data-expand-inset"] = c.ListInset
		attrs["data-expand-surface"] = c.ListSurface
		attrs["data-expand-icon"] = c.ListExpandIcon
		attrs["aria-label"] = "Example list"
	}
	if c.Family == "button-group" {
		attrs["data-size"] = c.Size
		attrs["role"] = "group"
		attrs["aria-label"] = "Actions"
	}
	if c.Family == "split" {
		attrs["data-size"] = c.Size
		attrs["role"] = "group"
		attrs["aria-label"] = "Add to cart"
	}
	if c.Family == "appbar" && !strings.HasPrefix(c.Variant, "baseline-") && c.Variant != "bottom" {
		attrs["data-modern"] = ""
	}
	if strings.HasPrefix(c.Family, "nav-") {
		attrs["data-variant"] = c.Variant
		attrs["data-navigation-group"] = ""
		attrs["data-layout"] = c.NavLayout
		attrs["data-badge"] = c.NavBadge
		attrs["data-indicator"] = c.NavIndicator
		attrs["data-orientation"] = c.Orientation
		if c.Family == "nav-bar" {
			attrs["data-orientation"] = c.Variant
		}
		if c.Family == "nav-drawer" {
			attrs["data-orientation"] = "horizontal"
		}
	}
	if c.Family == "badge" || c.Family == "snackbar" || c.Family == "tooltip" {
		attrs["data-variant"] = c.Variant
	}
	if c.Size != "" {
		attrs["data-size"] = c.Size
		attrs["data-shape"] = c.Shape
		attrs["data-width"] = c.Width
	}
	if c.Family == "icon" {
		attrs["aria-pressed"] = fmt.Sprint(c.Selected)
		attrs["data-button-toggle"] = ""
		attrs["aria-label"] = "Mark complete"
		attrs["type"] = "button"
	}
	if c.Family == "field" || c.Family == "select" || c.Family == "autocomplete" {
		attrs["data-field-example"] = ""
		attrs["data-variant"] = c.Variant
		attrs["data-populated"] = fmt.Sprint(c.Width == "populated")
		attrs["data-shape"] = c.Shape
	}
	if c.Family == "switch" {
		attrs["data-with-icon"] = fmt.Sprint(c.Variant == "icon")
	}
	if c.Family == "chip" {
		attrs["data-elevated"] = fmt.Sprint(c.Elevated)
		attrs["data-variant"] = c.Variant
		attrs["data-shape"] = c.Shape
		attrs["data-selected"] = fmt.Sprint(c.Selected)
		if c.Variant == "assist" || c.Variant == "suggestion" {
			attrs["type"] = "button"
			attrs["disabled"] = c.State == "disabled"
			attrs["data-chip-action"] = ""
		}
	}
	if c.Family == "menu" && c.Selected {
		attrs["aria-current"] = "page"
	}

	if c.Family == "field" || c.Family == "select" || c.Family == "autocomplete" || c.Family == "checkbox" {
		attrs["aria-invalid"] = fmt.Sprint(c.Invalid)
	}
	if c.State == "disabled" {
		attrs["aria-disabled"] = "true"
		if c.Family == "icon" {
			attrs["disabled"] = true
		}
	}
	if c.Family == "toolbar" || c.Family == "side-sheet" || c.Family == "button-group" || c.Family == "list" || c.Family == "carousel" || c.Family == "search" || c.Family == "date" || c.Family == "time" {
		delete(attrs, "aria-disabled")
	}
	if c.State == "interactive" {
		delete(attrs, "data-component-state")
	}
	return attrs
}
