package web

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/sojupie/tokens/source/publish"
)

type tokenBrowser struct {
	Preview  *componentPreview
	Contexts url.Values
	Sets     []tokenSet
	Selected tokenSet
	Filters  []tokenFilter
	Query    string
	Count    int
}

type tokenSet struct {
	Key          string
	Label        string
	component    string
	category     string
	variant      string
	sourceName   string
	order        int
	sourceTokens []publish.TokenView
	Groups       []tokenGroup
}

type tokenGroup struct {
	Name     string
	Count    int
	Open     bool
	Tokens   []tokenRow
	Children []tokenGroup
}

type tokenRow struct {
	Name       string
	SourceName string
	Value      string
	Color      string
	Elevation  string
	Shape      string
	Lifecycle  string
	HasValue   bool
	References []publish.TokenReference
}

type tokenFilter struct {
	Name     string
	Key      string
	Selected string
	Options  []string
}

type tokenFolder struct {
	name     string
	order    int
	tokens   []tokenRow
	children map[string]*tokenFolder
}

func tokenSets(subject publish.SubjectView) []tokenSet {
	var result []tokenSet
	sets := make(map[string]*tokenSet)
	for _, token := range subject.Tokens {
		if source := token.Presentation; source != nil {
			key := source.SetName
			set := sets[key]
			if set == nil {
				set = &tokenSet{Key: key, Label: source.SetLabel, sourceName: source.SetName, order: source.SetOrder, component: subject.Name}
				sets[key] = set
			}
			set.sourceTokens = append(set.sourceTokens, token)
			continue
		}
		isComponent := subject.Key.Kind == publish.SubjectComponent
		component, variant, _, _ := tokenLocation(token, isComponent)
		category := token.Category
		if isComponent {
			switch strings.ToUpper(category) {
			case "COLOR", "OPACITY", "ELEVATION":
				category = "COLOR"
			default:
				category = "COMMON"
				for _, size := range []string{"xsmall", "small", "medium", "large", "xlarge"} {
					if variant == size {
						category = "SIZE"
					}
				}
			}
		}
		if category == "" {
			category = "TOKEN"
		}
		key := tokenSetKey(component, category, variant)
		set := sets[key]
		if set == nil {
			label := displayWords(component)
			if label == "" {
				label = "Tokens"
			}
			if !strings.EqualFold(component, category) {
				label += " - " + displayWords(strings.ToLower(category))
			}
			if variant != "" {
				label += " - " + displayWords(variant)
			}
			set = &tokenSet{Key: key, Label: label, component: component, category: category, variant: variant}
			sets[key] = set
		}
		set.sourceTokens = append(set.sourceTokens, token)
	}
	for _, set := range sets {
		if set.sourceName != "" {
			sort.SliceStable(set.sourceTokens, func(i, j int) bool {
				a, b := set.sourceTokens[i].Presentation, set.sourceTokens[j].Presentation
				if a.Order != b.Order {
					return a.Order < b.Order
				}
				return set.sourceTokens[i].SourceName < set.sourceTokens[j].SourceName
			})
		}
		result = append(result, *set)
	}
	sort.Slice(result, func(i, j int) bool {
		left, right := result[i], result[j]
		if left.sourceName != "" && right.sourceName != "" {
			if left.order != right.order {
				return left.order < right.order
			}
			return left.sourceName < right.sourceName
		}
		if left.component != right.component {
			return left.component < right.component
		}
		if left.category != right.category {
			if categoryOrder(left.category) == categoryOrder(right.category) {
				return left.category < right.category
			}
			return categoryOrder(left.category) < categoryOrder(right.category)
		}
		if left.variant == "" || right.variant == "" {
			return left.variant == ""
		}
		return left.variant < right.variant
	})
	return result
}

func buildTokenBrowser(subject publish.SubjectView, query url.Values) tokenBrowser {
	browser := tokenBrowser{Query: strings.TrimSpace(query.Get("q")), Sets: tokenSets(subject)}
	browser.Contexts = make(url.Values)
	for key, values := range query {
		if strings.HasPrefix(key, "context.") {
			browser.Contexts[key] = append([]string(nil), values...)
		}
	}
	if len(browser.Sets) == 0 {
		return browser
	}

	selectedKey := query.Get("set")
	for _, set := range browser.Sets {
		if set.Key == selectedKey {
			browser.Selected = set
			break
		}
	}
	if browser.Selected.Key == "" {
		browser.Selected = browser.Sets[0]
	}

	browser.Filters = tokenFilters(browser.Selected.sourceTokens, query)
	selectedFilters := make(map[string]string, len(browser.Filters))
	for _, filter := range browser.Filters {
		if filter.Selected != "" {
			selectedFilters[filter.Name] = filter.Selected
		}
	}
	folders := make(map[string]*tokenFolder)
	for _, token := range browser.Selected.sourceTokens {
		_, _, state, path := tokenLocation(token, subject.Key.Kind == publish.SubjectComponent)
		value, hasValue := selectTokenValue(token.Values, selectedFilters)
		row := tokenRow{Name: tokenDisplayName(token), SourceName: token.SourceName, HasValue: hasValue, References: value.References}
		if hasValue {
			row.Value = tokenValueDisplay(token.Category, value.Value)
			row.Shape = shapePreviewCSS(value.Value)
			var primitive map[string]json.RawMessage
			if json.Unmarshal(value.Value, &primitive) == nil && primitive["elevation"] != nil {
				row.Elevation = themeValue(value.Value)
			}
			if strings.EqualFold(token.Category, "COLOR") {
				row.Color = tokenColor(value.Value)
				if label := tokenColorLabel(value.Value); label != "" {
					row.Value = label
				}
			}
		}
		if !strings.EqualFold(token.Lifecycle, "active") {
			row.Lifecycle = displayWords(token.Lifecycle)
		}
		if browser.Query != "" && !tokenRowMatches(browser.Query, token, row, value, hasValue) {
			continue
		}
		if source := token.Presentation; source != nil {
			group := publish.DisplayGroup{}
			if len(source.Groups) > 0 {
				group = source.Groups[0]
			}
			root := folders[group.Name]
			if root == nil {
				root = &tokenFolder{name: group.Name, order: group.Order, children: make(map[string]*tokenFolder)}
				folders[group.Name] = root
			}
			current := root
			if len(source.Groups) > 1 {
				for _, childGroup := range source.Groups[1:] {
					child := current.children[childGroup.Name]
					if child == nil {
						child = &tokenFolder{name: childGroup.Name, order: childGroup.Order, children: make(map[string]*tokenFolder)}
						current.children[childGroup.Name] = child
					}
					current = child
				}
			}
			current.tokens = append(current.tokens, row)
		} else {
			root := folders[state]
			if root == nil {
				root = &tokenFolder{name: state, order: stateOrder(state), children: make(map[string]*tokenFolder)}
				folders[state] = root
			}
			if browser.Selected.category == "COLOR" {
				root.tokens = append(root.tokens, row)
			} else {
				addTokenRow(root, path, row)
			}
		}
		browser.Count++
	}

	stateNames := make([]string, 0, len(folders))
	for name := range folders {
		stateNames = append(stateNames, name)
	}
	sort.Slice(stateNames, func(i, j int) bool {
		left, right := folders[stateNames[i]].order, folders[stateNames[j]].order
		if left == right {
			return stateNames[i] < stateNames[j]
		}
		return left < right
	})
	for _, name := range stateNames {
		group := folderView(folders[name], browser.Query != "")
		browser.Selected.Groups = append(browser.Selected.Groups, group)
	}
	return browser
}

func tokenSetKey(component, category, variant string) string {
	return strings.ToLower(strings.Join([]string{component, category, variant}, "|"))
}

func categoryOrder(category string) int {
	switch strings.ToUpper(category) {
	case "COLOR":
		return 0
	case "SIZE":
		return 1
	case "COMMON":
		return 2
	case "SHAPE":
		return 1
	case "TYPOGRAPHY":
		return 2
	case "LENGTH", "DIMENSION":
		return 3
	case "ELEVATION":
		return 4
	case "OPACITY":
		return 5
	default:
		return 6
	}
}

func displayWords(value string) string {
	words := strings.FieldsFunc(value, func(character rune) bool {
		return character == '.' || character == '-' || character == '_' || character == '·'
	})
	for index, word := range words {
		if word == "" {
			continue
		}
		first, size := utf8.DecodeRuneInString(word)
		words[index] = string(unicode.ToUpper(first)) + word[size:]
	}
	return strings.Join(words, " ")
}

func tokenLocation(token publish.TokenView, isComponent bool) (component, variant, state string, path []string) {
	parts := strings.Split(token.SourceName, ".")
	if len(parts) >= 4 && parts[0] == "md" && parts[1] == "comp" {
		component, path = parts[2], append([]string(nil), parts[3:]...)
	} else if len(parts) >= 4 && parts[0] == "md" && (parts[1] == "sys" || parts[1] == "ref") {
		concept := parts[2]
		switch concept {
		case "typescale":
			concept = "Type scale"
		default:
			concept = displayWords(concept)
		}
		component, path = concept, append([]string(nil), parts[3:]...)
	} else if len(parts) >= 3 && parts[0] == "md" {
		component, path = displayWords(parts[1]), append([]string(nil), parts[2:]...)
	} else {
		path = append([]string(nil), parts...)
	}
	if isComponent && len(path) > 0 && !tokenStructuralWord(path[0]) {
		variant = path[0]
		path = path[1:]
	}
	state = "General"
	if isComponent {
		state = "Enabled"
		for index, part := range path {
			if isTokenState(part) && part != "selected" && part != "unselected" {
				state = tokenStateLabel(part)
				path = append(path[:index], path[index+1:]...)
				break
			}
		}
	}
	category := strings.ToLower(token.Category)
	if len(path) > 0 && strings.EqualFold(path[len(path)-1], category) {
		path = path[:len(path)-1]
	}
	if len(path) > 1 && strings.EqualFold(path[0], category) {
		path = path[1:]
	}
	return component, variant, state, path
}

func tokenStructuralWord(word string) bool {
	if isTokenState(word) {
		return true
	}
	switch strings.ToLower(word) {
	case "leading-space", "trailing-space", "icon-label-space", "with-icon":
		return true
	case "container", "icon", "label", "label-text", "state-layer", "leading-icon", "trailing-icon", "outline", "indicator", "active-indicator", "thumb", "track", "handle", "headline", "subtitle", "supporting-text", "title", "avatar", "badge", "close-button", "divider", "scrim", "overlay", "surface", "content", "segment", "touch-target", "shape", "shadow", "font", "inset", "spacing", "margin", "padding", "size", "height", "width", "color", "elevation", "opacity", "corner", "round", "square", "focus-indicator", "menu", "selected-icon", "unselected-icon":
		return true
	default:
		return false
	}
}

func isTokenState(word string) bool {
	switch strings.ToLower(word) {
	case "enabled", "disabled", "hover", "hovered", "focus", "focused", "press", "pressed", "selected", "unselected", "dragged", "expanded", "collapsed", "checked", "unchecked":
		return true
	default:
		return false
	}
}

func tokenStateLabel(state string) string {
	switch strings.ToLower(state) {
	case "hover":
		return "Hovered"
	case "focus":
		return "Focused"
	case "press":
		return "Pressed"
	default:
		return displayWords(state)
	}
}

func tokenFilters(tokens []publish.TokenView, query url.Values) []tokenFilter {
	values := make(map[string]map[string]bool)
	var contexts []map[string]string
	for _, token := range tokens {
		for _, value := range token.Values {
			context := make(map[string]string)
			for _, label := range value.ContextLabels {
				name, option, ok := strings.Cut(label, ":")
				if !ok {
					continue
				}
				name, option = strings.TrimSpace(name), strings.TrimSpace(option)
				if name == "" || option == "" {
					continue
				}
				if values[name] == nil {
					values[name] = make(map[string]bool)
				}
				values[name][option] = true
				context[name] = option
			}
			if len(context) > 0 {
				contexts = append(contexts, context)
			}
		}
	}
	names := make([]string, 0, len(values))
	for name, options := range values {
		if len(options) > 0 {
			names = append(names, name)
		}
	}
	sort.Slice(names, func(i, j int) bool {
		order := func(name string) int {
			switch name {
			case "Theme":
				return 0
			case "Contrast":
				return 1
			case "Platform":
				return 2
			case "Audience":
				return 3
			default:
				return 4
			}
		}
		if order(names[i]) != order(names[j]) {
			return order(names[i]) < order(names[j])
		}
		return names[i] < names[j]
	})
	filters := make([]tokenFilter, 0, len(names))
	for _, name := range names {
		available := make(map[string]bool)
		for _, context := range contexts {
			if option := context[name]; option != "" {
				available[option] = true
			}
		}
		options := make([]string, 0, len(values[name]))
		for option := range available {
			options = append(options, option)
		}
		if len(options) == 0 {
			continue
		}
		sort.Slice(options, func(i, j int) bool {
			left, right := contextOptionOrder(name, options[i]), contextOptionOrder(name, options[j])
			if left == right {
				return options[i] < options[j]
			}
			return left < right
		})
		selected := query.Get("context." + name)
		if !containsString(options, selected) {
			selected = options[0]
			if name == "Theme" && containsString(options, "Light") {
				selected = "Light"
			}
			if name == "Contrast" && containsString(options, "Default") {
				selected = "Default"
			}
			if name == "Platform" && containsString(options, "Web") {
				selected = "Web"
			}
			if name == "Audience" && containsString(options, "3P") {
				selected = "3P"
			}
		}
		filters = append(filters, tokenFilter{Name: name, Key: "context." + name, Selected: selected, Options: options})
		// A missing dimension is unspecified, not evidence for additional options.
		// Retain it for shared values, but derive further options only from explicit tags.
		remaining := contexts[:0]
		for _, context := range contexts {
			if option := context[name]; option == "" || option == selected {
				remaining = append(remaining, context)
			}
		}
		contexts = remaining
	}
	return filters
}

func contextOptionOrder(name, value string) int {
	preferred := map[string][]string{
		"Theme":    {"Dark", "Light"},
		"Contrast": {"Medium contrast", "Default", "High contrast"},
		"Platform": {"Android", "Compose", "iOS", "Web"},
		"Audience": {"1P Baseline", "3P"},
	}
	for index, option := range preferred[name] {
		if option == value {
			return index
		}
	}
	return len(preferred[name])
}

func containsString(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func selectTokenValue(values []publish.TokenValueView, filters map[string]string) (publish.TokenValueView, bool) {
	selected, found := publish.TokenValueView{}, false
	bestMatches, bestDefault := -1, false
	for _, value := range values {
		if tokenValueIsUndefined(value.Value) && len(value.References) == 0 {
			continue
		}
		labels := make(map[string]string, len(value.ContextLabels))
		for _, label := range value.ContextLabels {
			name, option, ok := strings.Cut(label, ":")
			if ok {
				labels[strings.TrimSpace(name)] = strings.TrimSpace(option)
			}
		}
		matches, compatible := 0, true
		for name, option := range filters {
			if actual, exists := labels[name]; exists {
				if actual != option {
					compatible = false
					break
				}
				matches++
			}
		}
		if !compatible {
			continue
		}
		if !found || matches > bestMatches || (matches == bestMatches && value.IsDefaultContext && !bestDefault) {
			selected, found = value, true
			bestMatches, bestDefault = matches, value.IsDefaultContext
		}
	}
	return selected, found
}

func tokenValueIsUndefined(raw json.RawMessage) bool {
	var value any
	if json.Unmarshal(raw, &value) != nil {
		return false
	}
	return containsUndefined(value)
}

func containsUndefined(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		if undefined, ok := typed["undefined"].(bool); ok && undefined {
			return true
		}
		for _, child := range typed {
			if containsUndefined(child) {
				return true
			}
		}
	case []any:
		for _, child := range typed {
			if containsUndefined(child) {
				return true
			}
		}
	}
	return false
}

func tokenColor(raw json.RawMessage) string {
	var value map[string]any
	if json.Unmarshal(raw, &value) != nil {
		return ""
	}
	color, ok := value["color"].(map[string]any)
	if !ok {
		color = value
	}
	formatted, ok := colorValueText(color)
	if !ok {
		return ""
	}
	return formatted
}

func tokenColorLabel(raw json.RawMessage) string {
	var value map[string]any
	if json.Unmarshal(raw, &value) != nil {
		return ""
	}
	color, ok := value["color"].(map[string]any)
	if !ok {
		color = value
	}
	red, green, blue, resolved := colorChannels(color)
	if !resolved {
		return ""
	}
	if red <= 1 && green <= 1 && blue <= 1 {
		red *= 255
		green *= 255
		blue *= 255
	}
	alpha, alphaOK := numberValue(color["alpha"])
	if !alphaOK {
		alpha = 1
	}
	channels := []float64{red, green, blue, alpha * 255}
	for _, channel := range channels {
		if math.Abs(channel-math.Round(channel)) >= 0.001 {
			return ""
		}
	}
	if alpha < 1 {
		return fmt.Sprintf("#%02X%02X%02X%02X", uint8(math.Round(red)), uint8(math.Round(green)), uint8(math.Round(blue)), uint8(math.Round(alpha*255)))
	}
	return fmt.Sprintf("#%02X%02X%02X", uint8(math.Round(red)), uint8(math.Round(green)), uint8(math.Round(blue)))
}

func colorCellClass(color string) string {
	if color != "" {
		return " tokens-token-color"
	}
	return ""
}

func tokenColorTextColor(color string) string {
	var red, green, blue, alpha float64
	if _, err := fmt.Sscanf(color, "rgba(%f, %f, %f, %f)", &red, &green, &blue, &alpha); err != nil {
		return "var(--tokens-preview-on-light)"
	}
	red, green, blue = red*alpha+255*(1-alpha), green*alpha+255*(1-alpha), blue*alpha+255*(1-alpha)
	linear := func(channel float64) float64 {
		channel /= 255
		if channel <= 0.04045 {
			return channel / 12.92
		}
		return math.Pow((channel+0.055)/1.055, 2.4)
	}
	if 0.2126*linear(red)+0.7152*linear(green)+0.0722*linear(blue) < 0.179 {
		return "var(--tokens-preview-on-dark)"
	}
	return "var(--tokens-preview-on-light)"
}

func tokenValueDisplay(category string, raw json.RawMessage) string {
	if strings.EqualFold(category, "COLOR") {
		if color := tokenColor(raw); color != "" {
			return color
		}
		var value map[string]json.RawMessage
		if json.Unmarshal(raw, &value) == nil {
			if _, hasColor := value["color"]; hasColor {
				return "Color not resolved"
			}
		}
	}
	return tokenValueText(raw)
}

func addTokenRow(root *tokenFolder, path []string, row tokenRow) {
	if len(path) > 0 {
		path = path[:len(path)-1]
	}
	current := root
	for _, part := range path {
		name := displayWords(part)
		child := current.children[name]
		if child == nil {
			child = &tokenFolder{name: name, children: make(map[string]*tokenFolder)}
			current.children[name] = child
		}
		current = child
	}
	current.tokens = append(current.tokens, row)
}

func folderView(folder *tokenFolder, open bool) tokenGroup {
	view := tokenGroup{Name: folder.name, Tokens: folder.tokens, Open: open}
	for _, child := range folder.children {
		view.Children = append(view.Children, folderView(child, open))
	}
	sort.Slice(view.Children, func(i, j int) bool {
		a, b := folder.children[view.Children[i].Name], folder.children[view.Children[j].Name]
		if a.order != b.order {
			return a.order < b.order
		}
		return a.name < b.name
	})
	view.Count = len(view.Tokens)
	for _, child := range view.Children {
		view.Count += child.Count
	}
	return view
}

func tokenRowMatches(query string, token publish.TokenView, row tokenRow, value publish.TokenValueView, hasValue bool) bool {
	query = strings.ToLower(query)
	if strings.Contains(strings.ToLower(token.SourceName), query) || strings.Contains(strings.ToLower(row.Name), query) ||
		strings.Contains(strings.ToLower(token.Category), query) || strings.Contains(strings.ToLower(row.Lifecycle), query) {
		return true
	}
	if referencesMatch(query, row.References) {
		return true
	}
	return hasValue && (strings.Contains(strings.ToLower(row.Value), query) || strings.Contains(strings.ToLower(tokenContextLabel(value)), query))
}

func referencesMatch(query string, references []publish.TokenReference) bool {
	for _, reference := range references {
		if strings.Contains(strings.ToLower(reference.Name), query) || referencesMatch(query, reference.References) {
			return true
		}
	}
	return false
}

func tokenDisplayName(token publish.TokenView) string {
	if token.DisplayName != "" {
		return token.DisplayName
	}
	parts := strings.Split(strings.TrimPrefix(token.SourceName, "md.comp."), ".")
	selection := ""
	for index := len(parts) - 1; index >= 0; index-- {
		if parts[index] == "selected" || parts[index] == "unselected" {
			selection = parts[index]
			parts = append(parts[:index], parts[index+1:]...)
		}
	}
	name := displayWords(strings.Join(parts, "."))
	if first, rest, ok := strings.Cut(name, " "); ok {
		name = first + " " + strings.ToLower(rest)
	}
	if selection != "" {
		name += " - toggle (" + selection + ")"
	}
	return name
}

func stateOrder(state string) int {
	order := map[string]int{"Enabled": 0, "Disabled": 1, "Hovered": 2, "Focused": 3, "Pressed": 4, "Selected": 5, "Unselected": 6}
	if value, ok := order[state]; ok {
		return value
	}
	return len(order)
}
