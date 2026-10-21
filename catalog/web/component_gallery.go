package web

import (
	"fmt"
	"strings"

	"github.com/a-h/templ"
)

type componentCase struct {
	MenuMode        string `json:"menuMode,omitempty"`
	SegmentDensity  string `json:"segmentDensity,omitempty"`
	SearchContent   string `json:"searchContent,omitempty"`
	SearchElevation string `json:"searchElevation,omitempty"`
	DialogHeadline  string `json:"dialogHeadline,omitempty"`
	DateRange       bool   `json:"dateRange,omitempty"`
	ListMedia       string `json:"listMedia,omitempty"`
	ListDivider     string `json:"listDivider,omitempty"`
	ListTrailing    string `json:"listTrailing,omitempty"`
	ListText        string `json:"listText,omitempty"`
	ListInset       string `json:"listInset,omitempty"`
	ListSurface     string `json:"listSurface,omitempty"`
	ListExpandIcon  string `json:"listExpandIcon,omitempty"`
	NavLayout       string `json:"navLayout,omitempty"`
	NavCollapse     string `json:"navCollapse,omitempty"`
	NavAlign        string `json:"navAlign,omitempty"`
	NavHeader       string `json:"navHeader,omitempty"`
	NavIndicator    string `json:"navIndicator,omitempty"`
	NavBadge        string `json:"navBadge,omitempty"`
	Expanded        bool   `json:"expanded,omitempty"`
	Orientation     string `json:"orientation,omitempty"`
	Control         string `json:"control,omitempty"`
	Selection       string `json:"selection,omitempty"`
	Family          string `json:"family"`
	Variant         string `json:"variant"`
	Size            string `json:"size"`
	Shape           string `json:"shape"`
	Width           string `json:"width"`
	State           string `json:"state"`
	Selected        bool   `json:"selected"`
	Elevated        bool   `json:"elevated"`
	Invalid         bool   `json:"invalid"`
}

func (c componentCase) ID() string {
	id := fmt.Sprintf("%s-%s-%s-%s-%s-%s-%t-%t-%t", c.Family, c.Variant, c.Size, c.Shape, c.Width, c.State, c.Selected, c.Elevated, c.Invalid)
	if c.DialogHeadline == "hidden" {
		id += "-no-headline"
	}
	if c.DateRange {
		id += "-range"
	}
	if c.Expanded {
		id += "-expanded"
	}
	if (c.Family == "list" || c.Family == "date" || c.Family == "time") && c.Selection != "" {
		id += "-" + c.Selection
	}
	if c.Control != "" {
		id += "-" + c.Control
	}
	if c.Orientation != "" {
		id += "-" + c.Orientation
	}
	if c.SegmentDensity != "" && c.SegmentDensity != "0" {
		id += "-density" + c.SegmentDensity
	}
	return id
}
func (c componentCase) Label() string {
	if c.Family == "button" {
		return c.button().Label()
	}
	parts := []string{c.Family}
	if c.Variant != "" {
		parts = append(parts, c.Variant)
	}
	if c.Size != "" {
		parts = append(parts, c.Size, c.Shape, c.Width+" width")
	}
	if c.Family == "chip" {
		if c.Elevated {
			parts = append(parts, "elevated")
		} else {
			parts = append(parts, "flat")
		}
	}
	if c.Family == "icon" || c.Family == "chip" || c.Family == "radio" || c.Family == "menu" {
		if c.Selected {
			parts = append(parts, "selected")
		} else {
			parts = append(parts, "unselected")
		}
	}
	if c.Invalid {
		parts = append(parts, "error")
	}
	return strings.Join(append(parts, c.State), " · ")
}

func (c componentCase) button() buttonCase {
	return buttonCase{Variant: c.Variant, Size: c.Size, Shape: c.Shape, Selection: c.Selection, State: c.State}
}

func (c componentCase) Attributes() templ.Attributes {
	if c.Family == "button" {
		return c.button().Attributes()
	}
	class := map[string]string{"card": "md-card md-" + c.Variant + "-card md-interactive-card", "icon": "md-icon-button md-" + c.Variant + "-icon-button", "chip": "md-filter-chip", "radio": "md-radio", "field": "md-outlined-text-field", "select": "md-outlined-select", "menu": "md-menu-item", "dialog": "md-dialog", "divider": "md-divider", "checkbox": "md-checkbox", "switch": "md-switch", "fab": "md-button md-fab", "extended-fab": "md-button md-extended-fab", "badge": "md-badge", "snackbar": "md-snackbar", "tooltip": "md-tooltip", "nav-bar": "md-navigation md-navigation-bar", "nav-rail": "md-navigation md-navigation-rail", "nav-drawer": "md-navigation md-navigation-drawer", "tabs": "md-tabs-example", "segmented": "md-segmented-buttons", "slider": "md-slider", "progress": "md-progress", "loading": "md-loading", "appbar": "md-app-bar", "toolbar": "md-toolbar", "bottom-sheet": "md-sheet md-bottom-sheet", "side-sheet": "md-sheet md-side-sheet", "split": "md-split-button", "button-group": "md-button-group", "fab-menu": "md-fab-menu", "list": "md-list", "carousel": "md-carousel", "search": "md-search-demo", "date": "md-date-demo", "time": "md-time-demo"}[c.Family]
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
		attrs["aria-label"] = c.Label()
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

func (c componentCase) FeedbackAttributes(class string) templ.Attributes {
	if c.Family == "chip" && c.Variant != "" {
		class = "md-" + c.Variant + "-chip"
	}
	if c.Family == "dialog" && c.Variant == "full-screen" {
		class = "md-full-screen-dialog"
	}
	attrs := templ.Attributes{"class": class, "type": "button"}
	if c.State != "interactive" {
		state := c.State
		if c.Family == "snackbar" && ((c.Control == "dismiss") != strings.Contains(class, "md-snackbar-close")) {
			state = "enabled"
		}
		attrs["data-component-state"] = state
	}
	return attrs
}

func (c componentCase) NavigationItem(index int) templ.Attributes {
	attrs := templ.Attributes{"class": "md-nav-item", "type": "button", "data-navigation-item": "", "aria-controls": c.ID() + "-selection"}
	if c.Family == "nav-rail" {
		attrs["aria-controls"] = c.ID() + "-screen"
	}
	if (index == 0 && c.Selected) || (index == 1 && !c.Selected) {
		attrs["aria-current"] = "page"
	}
	if c.State != "interactive" {
		attrs["data-component-state"] = "enabled"
		if index == 0 {
			attrs["data-component-state"] = c.State
		}
	}
	return attrs
}

func (c componentCase) NavigationToggleLabel() string {
	if c.Variant == "expanded" {
		return "Collapse navigation"
	}
	return "Expand navigation"
}

func (c componentCase) ItemSelected(index int) bool {
	return index == 0 && c.Selected || index == 1 && !c.Selected
}
func (c componentCase) TabAttributes(index int) templ.Attributes {
	selected := c.ItemSelected(index)
	attrs := templ.Attributes{"class": "md-tab", "type": "button", "role": "tab", "id": fmt.Sprintf("%s-tab-%d", c.ID(), index), "aria-controls": fmt.Sprintf("%s-panel-%d", c.ID(), index), "aria-selected": fmt.Sprint(selected), "tabindex": "-1"}
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
func (c componentCase) SegmentAttributes(index int) templ.Attributes {
	attrs := templ.Attributes{"class": "md-segmented-button"}
	if c.State != "interactive" {
		attrs["data-component-state"] = "enabled"
		if index == 0 {
			attrs["data-component-state"] = c.State
		}
	}
	return attrs
}
func (c componentCase) SegmentInput(index int, label string) templ.Attributes {
	kind := "radio"
	if c.Variant == "multiple" {
		kind = "checkbox"
	}
	return templ.Attributes{"type": kind, "name": c.ID(), "checked": c.ItemSelected(index), "disabled": c.State == "disabled", "aria-label": label}
}

func (c componentCase) SliderHandles() []string {
	if c.Control == "range" {
		return []string{"start", "end"}
	}
	return []string{"end"}
}
func (c componentCase) SliderValue(handle string) string {
	if c.Control == "range" {
		if handle == "start" {
			return "30"
		}
		return "70"
	}
	return "40"
}
func (c componentCase) SliderInput(handle string) templ.Attributes {
	label := "Value"
	if c.Control == "range" {
		label = "Upper value"
		if handle == "start" {
			label = "Lower value"
		}
	}
	step := "1"
	if c.Variant == "discrete" {
		step = "10"
	}
	return templ.Attributes{"type": "range", "min": "0", "max": "100", "step": step, "value": c.SliderValue(handle), "name": c.ID() + "-" + handle, "aria-label": label, "disabled": c.State == "disabled", "orient": c.Orientation}
}
func (c componentCase) SliderHandle(handle string) templ.Attributes {
	attrs := templ.Attributes{"class": "md-slider-input", "data-handle": handle}
	if c.State != "interactive" {
		attrs["data-component-state"] = "enabled"
		if handle == c.SliderHandles()[0] {
			attrs["data-component-state"] = c.State
		}
	}
	return attrs
}

func (c componentCase) ProgressAttributes() templ.Attributes {
	attrs := c.Attributes()
	attrs["role"] = "progressbar"
	attrs["aria-label"] = "Loading"
	if c.Control == "determinate" {
		attrs["aria-valuemin"] = "0"
		attrs["aria-valuemax"] = "100"
		attrs["aria-valuenow"] = "40"
	}
	return attrs
}

func (c componentCase) ProgressControl() componentCase {
	return componentCase{Family: "slider", Variant: "continuous", Control: "single", Size: "xsmall", Shape: "value-indicator", Orientation: "horizontal", State: "interactive"}
}

func (c componentCase) BarAction(index int) templ.Attributes {
	attrs := templ.Attributes{"type": "button", "class": "md-icon-button md-standard-icon-button", "data-size": "small", "data-shape": "round", "data-width": "default"}
	labels := []string{"Back", "Favorite", "More options"}
	if c.Family == "toolbar" {
		labels = []string{"Edit", "Save", "More"}
		attrs["class"] = "md-icon-button md-toolbar-item"
		if c.Shape != "icon" {
			attrs["class"] = "md-button md-toolbar-item"
		}
		attrs["data-toolbar-item"] = fmt.Sprint(index)
		if index == 2 {
			attrs["data-menu-open"] = c.ID() + "-toolbar-menu"
			attrs["aria-haspopup"] = "menu"
			attrs["aria-expanded"] = "false"
		}
		attrs["tabindex"] = "-1"
		if index == 0 {
			attrs["aria-pressed"] = fmt.Sprint(c.Selected)
			attrs["data-button-toggle"] = ""
			attrs["tabindex"] = "0"
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
	}
	if c.Family == "toolbar" && c.State == "disabled" {
		if index == 0 {
			attrs["tabindex"] = "-1"
		}
		if index == 1 {
			attrs["tabindex"] = "0"
		}
	}
	attrs["aria-label"] = labels[index]
	return attrs
}
func (c componentCase) BarFAB() templ.Attributes {
	return templ.Attributes{"class": "md-button md-fab md-toolbar-fab", "type": "button", "aria-label": "Create", "data-size": c.Size, "data-color": c.Control}
}

func (c componentCase) BarExpanded() bool {
	return c.Variant != "bottom" && c.Variant != "small" && c.Variant != "centered" && c.Variant != "baseline-small" && c.Variant != "baseline-centered"
}

func (c componentCase) DividerInset() string {
	if c.Variant != "inset" {
		return "false"
	}
	if c.Shape == "both" {
		return "middle"
	}
	return "true"
}
