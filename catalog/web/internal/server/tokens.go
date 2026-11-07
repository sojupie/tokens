package server

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strings"

	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	"github.com/sojupie/tokens/source/publish"
)

type tokenFolder struct {
	name     string
	order    int
	tokens   []webview.TokenRow
	children map[string]*tokenFolder
}

func buildTokenBrowser(subject publish.SubjectView, query url.Values) webview.TokenBrowser {
	browser := webview.TokenBrowser{Query: strings.TrimSpace(query.Get("q")), Sets: webview.TokenSets(subject)}
	browser.PreviewQuery = make(url.Values)
	for key, values := range query {
		if strings.HasPrefix(key, "preview.") {
			browser.PreviewQuery[key] = append([]string(nil), values...)
		}
	}
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
		browser.Selected = webview.DefaultTokenSet(subject, browser.Sets)
	}
	if selectedKey == "all" {
		browser.Selected = webview.AllTokenSet(subject)
	}

	browser.Filters = webview.TokenFilters(browser.Selected.SourceTokens, query)
	selectedFilters := make(map[string]string, len(browser.Filters))
	for _, filter := range browser.Filters {
		if filter.Selected != "" {
			selectedFilters[filter.Name] = filter.Selected
		}
	}
	folders := make(map[string]*tokenFolder)
	for _, token := range browser.Selected.SourceTokens {
		_, _, state, path := webview.TokenLocation(token, subject.Key.Kind == publish.SubjectComponent)
		value, hasValue := selectTokenValue(token.Values, selectedFilters)
		row := webview.TokenRow{Name: tokenDisplayName(token), SourceName: token.SourceName, HasValue: hasValue, References: value.References}
		if !hasValue {
			row.Value = "No value for selected context"
			if len(token.Values) == 0 {
				row.Value = "No captured value"
			}
		}
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
			row.Lifecycle = webview.DisplayWords(token.Lifecycle)
		}
		if browser.Query != "" && !tokenRowMatches(browser.Query, token, row, value, hasValue) {
			continue
		}
		rowFolders := folders
		if browser.Selected.All {

			for index, set := range browser.Sets {
				belongs := token.Presentation != nil && set.Key == token.Presentation.SetName
				if token.Presentation == nil {
					for _, member := range set.SourceTokens {
						if member.SourceName == token.SourceName {
							belongs = true
							break
						}
					}
				}
				if !belongs {
					continue
				}
				root := folders[set.Key]
				if root == nil {
					root = &tokenFolder{name: set.Label, order: index, children: make(map[string]*tokenFolder)}
					folders[set.Key] = root
				}
				rowFolders = root.children
				break
			}
		}
		if source := token.Presentation; source != nil {
			group := publish.DisplayGroup{}
			if len(source.Groups) > 0 {
				group = source.Groups[0]
			}
			root := rowFolders[group.Name]
			if root == nil {
				root = &tokenFolder{name: group.Name, order: group.Order, children: make(map[string]*tokenFolder)}
				rowFolders[group.Name] = root
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
			root := rowFolders[state]
			if root == nil {
				root = &tokenFolder{name: state, order: stateOrder(state), children: make(map[string]*tokenFolder)}
				rowFolders[state] = root
			}
			if browser.Selected.Category == "COLOR" {
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
	formatted, ok := webview.ColorValueText(color)
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
	red, green, blue, resolved := webview.ColorChannels(color)
	if !resolved {
		return ""
	}
	if red <= 1 && green <= 1 && blue <= 1 {
		red *= 255
		green *= 255
		blue *= 255
	}
	alpha, alphaOK := webview.NumberValue(color["alpha"])
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
	return webview.TokenValueText(raw)
}

func addTokenRow(root *tokenFolder, path []string, row webview.TokenRow) {
	if len(path) > 0 {
		path = path[:len(path)-1]
	}
	current := root
	for _, part := range path {
		name := webview.DisplayWords(part)
		child := current.children[name]
		if child == nil {
			child = &tokenFolder{name: name, children: make(map[string]*tokenFolder)}
			current.children[name] = child
		}
		current = child
	}
	current.tokens = append(current.tokens, row)
}

func folderView(folder *tokenFolder, open bool) webview.TokenGroup {
	view := webview.TokenGroup{Name: folder.name, Tokens: folder.tokens, Open: open}
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

func tokenRowMatches(query string, token publish.TokenView, row webview.TokenRow, value publish.TokenValueView, hasValue bool) bool {
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
	name := webview.DisplayWords(strings.Join(parts, "."))
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
