package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sojupie/tokens/source/publish"
)

var errPreviewInput = errors.New("invalid preview selection")

type previewControl struct {
	Name, Label, Selected string
	Options               []string
	Labels                map[string]string
}
type previewHidden struct{ Name, Value string }
type componentPreview struct {
	Action      string
	Dimension   string
	Cases       []componentCase
	Controls    []previewControl
	Hidden      []previewHidden
	Interactive bool
}

func (page componentPreview) Caption(c componentCase) string {
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
		return dateVariantLabel(c.Variant)
	}
	return previewLabel(previewDimensions(c)[page.Dimension])
}

func previewLabel(value string) string {
	switch value {
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

func previewDimensions(c componentCase) map[string]string {
	d := map[string]string{"variant": c.Variant, "size": c.Size, "shape": c.Shape, "width": c.Width, "state": c.State, "control": c.Family, "orientation": c.Orientation}
	if c.Family == "nav-rail" {
		d["orientation"] = ""
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

func (server *Server) buildComponentPreview(subject publish.SubjectView, query url.Values) (*componentPreview, error) {
	if server.tokenAssetsDir == "" || subject.Key.Kind != publish.SubjectComponent {
		return nil, nil
	}
	family := map[string]string{"buttons": "button", "cards": "card", "icon-buttons": "icon", "chips": "chip", "radio-button": "radio", "text-fields": "field", "menus": "menu", "dialogs": "dialog", "divider": "divider", "checkbox": "checkbox", "switch": "switch", "floating-action-button": "fab", "extended-fab": "extended-fab", "badges": "badge", "snackbar": "snackbar", "tooltips": "tooltip", "navigation-bar": "nav-bar", "navigation-rail": "nav-rail", "navigation-drawer": "nav-drawer", "tabs": "tabs", "segmented-buttons": "segmented", "sliders": "slider", "progress-indicators": "progress", "loading-indicator": "loading", "app-bars": "appbar", "toolbars": "toolbar", "bottom-sheets": "bottom-sheet", "side-sheets": "side-sheet", "split-button": "split", "button-groups": "button-group", "fab-menu": "fab-menu", "lists": "list", "carousel": "carousel", "search": "search", "date-pickers": "date", "time-pickers": "time"}[subject.Key.Slug]
	if family == "" {
		return nil, nil
	}
	file := "components-coverage.json"
	if family == "button" {
		file = "buttons-coverage.json"
	}
	configurations, err := server.previewConfigurations(file)
	if err != nil {
		return nil, err
	}
	var cases []componentCase
	for _, c := range configurations {
		if family == "button" {
			c.Family = "button"
		}
		if c.Family == family || family == "field" && (c.Family == "select" || c.Family == "autocomplete") {
			cases = append(cases, c)
		}
	}
	if len(cases) == 0 {
		return nil, nil
	}
	return selectComponentPreview(cases, query)
}

// Coverage is immutable for this server's loaded publication. Decode each file
// once, rather than reparsing megabytes for every card and every keystroke.
func (server *Server) previewConfigurations(file string) ([]componentCase, error) {
	server.previewMu.Lock()
	defer server.previewMu.Unlock()
	if cases, ok := server.previewFiles[file]; ok {
		return cases, nil
	}
	data, err := os.ReadFile(filepath.Join(server.tokenAssetsDir, file))
	if err != nil {
		return nil, err
	}
	var coverage struct {
		Configurations []componentCase `json:"configurations"`
	}
	if err := json.Unmarshal(data, &coverage); err != nil {
		return nil, err
	}
	if server.previewFiles == nil {
		server.previewFiles = make(map[string][]componentCase)
	}
	server.previewFiles[file] = coverage.Configurations
	return coverage.Configurations, nil
}

// Only one dimension varies in a comparison. The source configurations define
// valid combinations; the page never constructs a Cartesian product.
func selectComponentPreview(cases []componentCase, query url.Values) (*componentPreview, error) {
	// Native selection controls are compositions of existing list and control
	// renderers; they do not add a token configuration product.
	if len(cases) > 0 && cases[0].Family == "list" {
		expanded := append([]componentCase(nil), cases...)
		for _, c := range cases {
			if c.Control == "select" && c.Width == "none" {
				for _, width := range []string{"checkbox", "radio"} {
					copy := c
					copy.Width = width
					expanded = append(expanded, copy)
				}
			}
		}
		cases = expanded
	}
	if len(cases) > 0 && cases[0].Family == "split" {
		expanded := make([]componentCase, 0, len(cases)*3)
		for _, c := range cases {
			for _, content := range []string{"icon-label", "label", "icon"} {
				c.Shape = content
				expanded = append(expanded, c)
			}
		}
		cases = expanded
	}
	if len(cases) > 0 && cases[0].Family == "toolbar" {
		queryCopy := make(url.Values, len(query))
		for key, values := range query {
			queryCopy[key] = append([]string(nil), values...)
		}
		query = queryCopy
		if query.Get("preview.variant") == "floating-fab" {
			query.Set("preview.variant", "floating")
			if query.Get("preview.size") == "" {
				query.Set("preview.size", "small")
			}
		}
		if query.Get("preview.compare") == "variants" {
			query.Set("preview.orientation", "horizontal")
			query.Set("preview.size", "none")
		}
		if query.Get("preview.compare") == "sizes" {
			query.Set("preview.variant", "floating")
		}
	}
	options := map[string][]string{}
	for _, c := range cases {
		for key, value := range previewDimensions(c) {
			if value != "" && !containsPreview(options[key], value) {
				options[key] = append(options[key], value)
			}
		}
	}
	comparisons := []string{"states"}
	if len(options["size"]) > 1 {
		comparisons = append(comparisons, "sizes")
	}
	if len(options["variant"]) > 1 {
		comparisons = append(comparisons, "variants")
	}
	comparisons = append(comparisons, "interactive")
	search := len(cases) > 0 && cases[0].Family == "search"
	if search {
		comparisons = []string{"layouts", "styles", "states", "interactive"}
	}
	comparison := query.Get("preview.compare")
	if search && comparison == "variants" {
		comparison = "layouts"
	}
	if comparison == "" {
		comparison = "states"
		if len(options["variant"]) > 1 {
			comparison = "variants"
		}
	}
	if search && comparison == "variants" {
		comparison = "layouts"
	}
	if !containsPreview(comparisons, comparison) {
		return nil, fmt.Errorf("%w: comparison %s", errPreviewInput, comparison)
	}
	varying := map[string]string{"states": "state", "sizes": "size", "variants": "variant", "layouts": "variant", "styles": "control"}[comparison]
	result := &componentPreview{Interactive: comparison == "interactive", Dimension: varying}
	result.Controls = append(result.Controls, previewControl{Name: "compare", Label: "Compare", Selected: comparison, Options: comparisons})
	defaults := map[string]string{"variant": "tonal", "size": "small", "shape": "round", "state": "enabled", "selection": "plain", "control": "indeterminate"}
	if len(cases) > 0 && cases[0].Family == "menu" {
		defaults["variant"] = "standard"
	}
	if len(cases) > 0 && cases[0].Family == "slider" {
		defaults["size"] = "xsmall"
		defaults["shape"] = "value-indicator"
	}
	if len(cases) > 0 && (cases[0].Family == "fab" || cases[0].Family == "extended-fab") {
		defaults["size"] = "default"
		defaults["variant"] = "primary-container"
	}
	if len(cases) > 0 && cases[0].Family == "search" {
		defaults["variant"] = "adaptive"
		defaults["control"] = "contained"
		if comparison == "layouts" || comparison == "styles" {
			defaults["disclosure"] = "expanded"
		}
	}
	if len(cases) > 0 && cases[0].Family == "toolbar" {
		defaults["size"] = "none"
	}
	selected := map[string]string{}
	keys := []string{"control", "mode", "disclosure", "variant", "headline", "orientation", "size", "shape", "width", "selection", "elevation", "error", "state"}
	if len(cases) > 0 && cases[0].Family == "date" {
		keys = []string{"mode", "variant", "control", "selection", "state"}
	}
	if len(cases) > 0 && cases[0].Family == "menu" {
		keys = []string{"control", "mode", "variant", "orientation", "size", "shape", "disclosure", "width", "selection", "elevation", "error", "state"}
	}
	if len(cases) > 0 && cases[0].Family == "dialog" {
		keys = []string{"variant", "headline", "control", "mode", "orientation", "shape", "width", "state"}
	}
	for _, key := range keys {
		var values []string
		for _, candidate := range cases {
			dims := previewDimensions(candidate)
			matches := true
			for prior, value := range selected {
				if prior != varying && dims[prior] != "" && dims[prior] != value {
					matches = false
					break
				}
			}
			if value := dims[key]; matches && value != "" && !containsPreview(values, value) {
				values = append(values, value)
			}
		}
		if len(values) == 0 {
			continue
		}
		value := query.Get("preview." + key)
		if value == "" {
			value = defaults[key]
			if !containsPreview(values, value) {
				value = values[0]
			}
		}
		if !containsPreview(options[key], value) {
			return nil, fmt.Errorf("%w: %s %s", errPreviewInput, key, value)
		}
		if !containsPreview(values, value) {
			value = values[0]
		}
		if search && comparison == "states" && key == "disclosure" {
			value = "collapsed"
		}
		if comparison == "interactive" && key == "state" {
			value = "enabled"
		}
		selected[key] = value
		if key != varying && len(values) > 1 && !(search && comparison == "states" && key == "disclosure") && !(comparison == "interactive" && key == "state") {
			result.Controls = append(result.Controls, previewControl{Name: key, Label: previewControlLabel(cases[0].Family, key), Selected: value, Options: values})
		}
	}
	for _, c := range cases {
		if c.Family == "search" && comparison == "layouts" && c.Variant == "adaptive" {
			continue
		}
		matches := true
		for key, value := range previewDimensions(c) {
			if key != varying && value != "" && value != selected[key] {
				matches = false
				break
			}
		}
		if matches {
			if result.Interactive {
				c.State = "interactive"
			}
			result.Cases = append(result.Cases, c)
		}
	}
	if len(result.Cases) > 0 && result.Cases[0].Family == "menu" {
		if err := configureMenuPreview(result, query); err != nil {
			return nil, err
		}
	}
	if len(result.Cases) > 0 && result.Cases[0].Family == "toolbar" {
		if err := configureToolbarPreview(result, query); err != nil {
			return nil, err
		}
	}
	if len(result.Cases) > 0 && result.Cases[0].Family == "segmented" {
		if err := configureSegmentedPreview(result, query); err != nil {
			return nil, err
		}
	}
	if len(result.Cases) > 0 && result.Cases[0].Family == "search" {
		if err := configureSearchPreview(result, query); err != nil {
			return nil, err
		}
	}
	if len(result.Cases) > 0 && result.Cases[0].Family == "nav-rail" {
		if err := configureNavigationPreview(result, query); err != nil {
			return nil, err
		}
	}
	if len(result.Cases) > 0 && result.Cases[0].Family == "list" {
		if err := configureListPreview(result, query); err != nil {
			return nil, err
		}
	}
	if len(cases) > 0 && cases[0].Family == "date" {
		for i := range result.Controls {
			if result.Controls[i].Name == "variant" {
				result.Controls[i].Labels = map[string]string{"docked": "Docked date picker", "modal": "Modal date picker", "input": "Modal date input", "full-screen": "Full-screen date picker"}
			}
		}
	}
	if len(cases) > 0 && cases[0].Family == "dialog" {
		for i := range result.Controls {
			if result.Controls[i].Name == "shape" {
				result.Controls[i].Labels = map[string]string{"plain": "None", "icon": "Icon", "divided": "Dividers", "icon-divided": "Icon and dividers"}
			}
		}
	}
	for key, values := range query {
		if strings.HasPrefix(key, "preview.") {
			continue
		}
		for _, value := range values {
			result.Hidden = append(result.Hidden, previewHidden{Name: key, Value: value})
		}
	}
	sort.Slice(result.Hidden, func(i, j int) bool {
		if result.Hidden[i].Name != result.Hidden[j].Name {
			return result.Hidden[i].Name < result.Hidden[j].Name
		}
		return result.Hidden[i].Value < result.Hidden[j].Value
	})
	return result, nil
}
func dateVariantLabel(value string) string {
	return map[string]string{"docked": "Docked date picker", "modal": "Modal date picker", "input": "Modal date input", "full-screen": "Full-screen date picker"}[value]
}

func (c previewControl) OptionLabel(value string) string {
	if label := c.Labels[value]; label != "" {
		return label
	}
	return previewLabel(value)
}

func containsPreview(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

func previewControlLabel(family, key string) string {
	if family == "nav-rail" && key == "orientation" {
		return "Label layout"
	}
	if family == "nav-rail" && key == "size" {
		return "Item height"
	}
	if family == "divider" {
		if key == "control" {
			return "Content"
		}
		if key == "width" {
			return "Outer spacing"
		}
		if key == "shape" {
			return "Inset sides"
		}
	}
	if family == "menu" {
		if key == "shape" {
			return "Content"
		}
		if key == "disclosure" {
			return "Submenu"
		}
	}
	if family == "dialog" {
		switch key {
		case "control":
			return "State target"
		case "orientation":
			return "Actions"
		case "width":
			return "Scroll position"
		case "shape":
			return "Decoration"
		case "mode":
			return "Content length"
		}
	}

	if family == "field" || family == "select" || family == "autocomplete" {
		if key == "shape" {
			return "Content"
		}
		if key == "width" {
			return "Value"
		}
	}

	if family == "chip" {
		if key == "shape" {
			return "Leading content"
		}
		if key == "width" {
			return "Removal"
		}
	}
	if family == "time" {
		switch key {
		case "width":
			return "Time format"
		case "control":
			return "State target"
		case "shape":
			return "Active field"
		case "orientation":
			return "Layout"
		}
	}
	if family == "date" {
		if key == "mode" {
			return "Selection"
		}
		if key == "control" {
			return "Calendar view"
		}
		if key == "selection" {
			return "Date appearance"
		}
	}
	if family == "search" {
		switch key {
		case "control":
			return "Style"
		case "variant":
			return "Layout"
		case "shape":
			return "Leading content"
		case "width":
			return "Trailing action"
		}
	}
	if family == "carousel" {
		if key == "control" {
			return "Item outline"
		}
		if key == "size" {
			return "Preview size"
		}
	}
	if family == "list" {
		switch key {
		case "control":
			return "Behavior"
		case "variant":
			return "Layout"
		case "width":
			return "Leading content"
		case "shape":
			return "Shapes"
		case "size":
			return "Text lines"
		case "mode":
			return "Selection mode"
		}
	}
	if family == "fab-menu" && key == "control" {
		return "State target"
	}
	if family == "button-group" && key == "control" {
		return "Layout"
	}
	if family == "button-group" && key == "shape" {
		return "Selection mode"
	}
	if strings.HasSuffix(family, "-sheet") && key == "control" {
		return "Behavior"
	}
	if strings.HasSuffix(family, "-sheet") && key == "orientation" {
		return "Edge"
	}
	if family == "toolbar" && key == "control" {
		return "Color"
	}
	if (family == "toolbar" || family == "split") && key == "shape" {
		return "Content"
	}
	if family == "slider" && key == "shape" {
		return "Value label"
	}
	return previewLabel(key)
}
