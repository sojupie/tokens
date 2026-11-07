package playground

import (
	"errors"
	"fmt"
	"net/url"

	"sort"
	"strings"

	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
)

var ErrPreviewInput = errors.New("invalid playground selection")

// Only one dimension varies in a comparison. The source configurations define
// valid combinations; the page never constructs a Cartesian product.
func SelectComponentPreview(cases []webview.ComponentCase, query url.Values) (*webview.ComponentPreview, error) {

	if len(cases) > 0 && cases[0].Family == "list" {
		expanded := append([]webview.ComponentCase(nil), cases...)
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
		expanded := make([]webview.ComponentCase, 0, len(cases)*3)
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
		for key, value := range webview.PreviewDimensions(c) {
			if value != "" && !ContainsPreview(options[key], value) {
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
	if !ContainsPreview(comparisons, comparison) {
		return nil, fmt.Errorf("%w: comparison %s", ErrPreviewInput, comparison)
	}
	varying := map[string]string{"states": "state", "sizes": "size", "variants": "variant", "layouts": "variant", "styles": "control"}[comparison]
	result := &webview.ComponentPreview{Interactive: comparison == "interactive", Dimension: varying}
	result.Controls = append(result.Controls, webview.PreviewControl{Name: "compare", Label: "Compare", Selected: comparison, Options: comparisons})
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
			dims := webview.PreviewDimensions(candidate)
			matches := true
			for prior, value := range selected {
				if prior != varying && dims[prior] != "" && dims[prior] != value {
					matches = false
					break
				}
			}
			if value := dims[key]; matches && value != "" && !ContainsPreview(values, value) {
				values = append(values, value)
			}
		}
		if len(values) == 0 {
			continue
		}
		value := query.Get("preview." + key)
		if value == "" {
			value = defaults[key]
			if key == "size" {
				for _, preferred := range []string{"medium", "default"} {
					if ContainsPreview(values, preferred) {
						value = preferred
						break
					}
				}
			}
			if !ContainsPreview(values, value) {
				value = values[0]
			}
		}
		if !ContainsPreview(options[key], value) {
			return nil, fmt.Errorf("%w: %s %s", ErrPreviewInput, key, value)
		}
		if !ContainsPreview(values, value) {
			value = values[0]
		}
		if cases[0].Family == "tabs" && key == "selection" && comparison != "states" {
			value = "selected"
		}
		if search && comparison == "states" && key == "disclosure" {
			value = "collapsed"
		}
		if comparison == "interactive" && key == "state" {
			value = "enabled"
		}
		selected[key] = value
		if key != varying && len(values) > 1 && !(search && comparison == "states" && key == "disclosure") && !(comparison == "interactive" && key == "state") && !(cases[0].Family == "tabs" && key == "selection" && comparison != "states") {
			result.Controls = append(result.Controls, webview.PreviewControl{Name: key, Label: PreviewControlLabel(cases[0].Family, key), Selected: value, Options: values})
		}
	}
	for _, c := range cases {
		if c.Family == "search" && comparison == "layouts" && c.Variant == "adaptive" {
			continue
		}
		matches := true
		for key, value := range webview.PreviewDimensions(c) {
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
	if len(result.Cases) > 0 && result.Cases[0].Family == "badge" {
		if err := ConfigureBadgePreview(result, query); err != nil {
			return nil, err
		}
	}
	if len(result.Cases) > 0 && result.Cases[0].Family == "carousel" {
		if err := ConfigureCarouselPreview(result, query); err != nil {
			return nil, err
		}
	}
	if len(result.Cases) > 0 && result.Cases[0].Family == "menu" {
		if err := ConfigureMenuPreview(result, query); err != nil {
			return nil, err
		}
	}
	if len(result.Cases) > 0 && result.Cases[0].Family == "slider" {
		if err := ConfigureSliderPreview(result, query); err != nil {
			return nil, err
		}
	}
	if len(result.Cases) > 0 && result.Cases[0].Family == "toolbar" {
		if err := ConfigureToolbarPreview(result, query); err != nil {
			return nil, err
		}
	}
	if len(result.Cases) > 0 && result.Cases[0].Family == "segmented" {
		if err := ConfigureSegmentedPreview(result, query); err != nil {
			return nil, err
		}
	}
	if len(result.Cases) > 0 && result.Cases[0].Family == "search" {
		if err := ConfigureSearchPreview(result, query); err != nil {
			return nil, err
		}
	}
	if len(result.Cases) > 0 && result.Cases[0].Family == "tabs" {
		if err := ConfigureTabsPreview(result, query); err != nil {
			return nil, err
		}
	}
	if len(result.Cases) > 0 && result.Cases[0].Family == "nav-bar" {
		if err := ConfigureNavBarPreview(result, query); err != nil {
			return nil, err
		}
	}
	if len(result.Cases) > 0 && result.Cases[0].Family == "nav-rail" {
		if err := ConfigureNavigationPreview(result, query); err != nil {
			return nil, err
		}
	}
	if len(result.Cases) > 0 && result.Cases[0].Family == "list" {
		if err := ConfigureListPreview(result, query); err != nil {
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
	if len(result.Cases) > 0 && result.Cases[0].Family == "button" {
		colors := []string{"default", "primary", "secondary", "tertiary", "error"}
		color := query.Get("preview.color")
		if color == "" {
			color = "default"
		}
		if !ContainsPreview(colors, color) {
			return nil, fmt.Errorf("%w: color %s", ErrPreviewInput, color)
		}
		result.Controls = append(result.Controls, webview.PreviewControl{Name: "color", Label: "Color", Selected: color, Options: colors})
		if color != "default" {
			for i := range result.Cases {
				result.Cases[i].Color = color
			}
		}
	}
	for key, values := range query {
		if strings.HasPrefix(key, "preview.") {
			continue
		}
		for _, value := range values {
			result.Hidden = append(result.Hidden, webview.PreviewHidden{Name: key, Value: value})
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

func ContainsPreview(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

func PreviewControlLabel(family, key string) string {
	if family == "nav-rail" && key == "width" {
		return "Collapsed width"
	}
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
	return webview.PreviewLabel(key)
}
