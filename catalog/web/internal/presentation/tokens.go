package presentation

import (
	"fmt"
	"math"
	"net/url"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/sojupie/tokens/source/publish"
)

type TokenBrowser struct {
	Preview      *ComponentPreview
	Contexts     url.Values
	PreviewQuery url.Values
	Sets         []TokenSet
	Selected     TokenSet
	Filters      []TokenFilter
	Query        string
	Count        int
}

type TokenSet struct {
	All          bool
	Key          string
	Label        string
	Component    string
	Category     string
	Variant      string
	SourceName   string
	Order        int
	SourceTokens []publish.TokenView
	Groups       []TokenGroup
}

type TokenGroup struct {
	Name     string
	Count    int
	Open     bool
	Tokens   []TokenRow
	Children []TokenGroup
}

type TokenRow struct {
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

type TokenFilter struct {
	Name     string
	Key      string
	Selected string
	Options  []string
}

// Elevation's payload includes these supporting definitions outside its table.
// Keep them in the publication and subject exports, but omit them from browsing.
func TokenTableTokens(subject publish.SubjectView) []publish.TokenView {
	if subject.Key.Kind != publish.SubjectStyle || subject.Key.Slug != "elevation" {
		return subject.Tokens
	}
	var tokens []publish.TokenView
	for _, token := range subject.Tokens {
		switch token.SourceName {
		case "md.sys.color.primary", "md.ref.palette.primary20", "md.ref.palette.primary30",
			"md.ref.palette.primary40", "md.ref.palette.primary80", "md.ref.palette.primary90", "md.ref.palette.primary95":
			continue
		}
		tokens = append(tokens, token)
	}
	return tokens
}

func TokenSets(subject publish.SubjectView) []TokenSet {
	var result []TokenSet
	sets := make(map[string]*TokenSet)
	for _, token := range TokenTableTokens(subject) {
		if source := token.Presentation; source != nil {
			key := source.SetName
			set := sets[key]
			if set == nil {
				set = &TokenSet{Key: key, Label: source.SetLabel, SourceName: source.SetName, Order: source.SetOrder, Component: subject.Name}
				sets[key] = set
			}
			set.SourceTokens = append(set.SourceTokens, token)
			continue
		}
		isComponent := subject.Key.Kind == publish.SubjectComponent
		component, variant, _, _ := TokenLocation(token, isComponent)
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
		key := TokenSetKey(component, category, variant)
		set := sets[key]
		if set == nil {
			label := DisplayWords(component)
			if label == "" {
				label = "Tokens"
			}
			if !strings.EqualFold(component, category) {
				label += " - " + DisplayWords(strings.ToLower(category))
			}
			if variant != "" {
				label += " - " + DisplayWords(variant)
			}
			set = &TokenSet{Key: key, Label: label, Component: component, Category: category, Variant: variant}
			sets[key] = set
		}
		set.SourceTokens = append(set.SourceTokens, token)
	}
	for _, set := range sets {
		if set.SourceName != "" {
			sort.SliceStable(set.SourceTokens, func(i, j int) bool {
				a, b := set.SourceTokens[i].Presentation, set.SourceTokens[j].Presentation
				if a.Order != b.Order {
					return a.Order < b.Order
				}
				return set.SourceTokens[i].SourceName < set.SourceTokens[j].SourceName
			})
		}
		result = append(result, *set)
	}
	sort.Slice(result, func(i, j int) bool {
		left, right := result[i], result[j]
		if left.SourceName != "" && right.SourceName != "" {
			if left.Order != right.Order {
				return left.Order < right.Order
			}
			return left.SourceName < right.SourceName
		}
		if left.Component != right.Component {
			return left.Component < right.Component
		}
		if left.Category != right.Category {
			if CategoryOrder(left.Category) == CategoryOrder(right.Category) {
				return left.Category < right.Category
			}
			return CategoryOrder(left.Category) < CategoryOrder(right.Category)
		}
		if left.Variant == "" || right.Variant == "" {
			return left.Variant == ""
		}
		return left.Variant < right.Variant
	})
	return result
}

func TokenSetKey(component, category, variant string) string {
	return strings.ToLower(strings.Join([]string{component, category, variant}, "|"))
}

func CategoryOrder(category string) int {
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

func DisplayWords(value string) string {
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

func TokenLocation(token publish.TokenView, isComponent bool) (component, variant, state string, path []string) {
	parts := strings.Split(token.SourceName, ".")
	if len(parts) >= 4 && parts[0] == "md" && parts[1] == "comp" {
		component, path = parts[2], append([]string(nil), parts[3:]...)
	} else if len(parts) >= 4 && parts[0] == "md" && (parts[1] == "sys" || parts[1] == "ref") {
		concept := parts[2]
		switch concept {
		case "typescale":
			concept = "Type scale"
		default:
			concept = DisplayWords(concept)
		}
		component, path = concept, append([]string(nil), parts[3:]...)
	} else if len(parts) >= 3 && parts[0] == "md" {
		component, path = DisplayWords(parts[1]), append([]string(nil), parts[2:]...)
	} else {
		path = append([]string(nil), parts...)
	}
	if isComponent && len(path) > 0 && !TokenStructuralWord(path[0]) {
		variant = path[0]
		path = path[1:]
	}
	state = "General"
	if isComponent {
		state = "Enabled"
		for index, part := range path {
			if IsTokenState(part) && part != "selected" && part != "unselected" {
				state = TokenStateLabel(part)
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

func TokenStructuralWord(word string) bool {
	if IsTokenState(word) {
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

func IsTokenState(word string) bool {
	switch strings.ToLower(word) {
	case "enabled", "disabled", "hover", "hovered", "focus", "focused", "press", "pressed", "selected", "unselected", "dragged", "expanded", "collapsed", "checked", "unchecked":
		return true
	default:
		return false
	}
}

func TokenStateLabel(state string) string {
	switch strings.ToLower(state) {
	case "hover":
		return "Hovered"
	case "focus":
		return "Focused"
	case "press":
		return "Pressed"
	default:
		return DisplayWords(state)
	}
}

func TokenFilters(tokens []publish.TokenView, query url.Values) []TokenFilter {
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
	filters := make([]TokenFilter, 0, len(names))
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
			left, right := ContextOptionOrder(name, options[i]), ContextOptionOrder(name, options[j])
			if left == right {
				return options[i] < options[j]
			}
			return left < right
		})
		selected := query.Get("context." + name)
		if !ContainsString(options, selected) {
			selected = options[0]
			if name == "Theme" && ContainsString(options, "Light") {
				selected = "Light"
			}
			if name == "Contrast" && ContainsString(options, "Default") {
				selected = "Default"
			}
			if name == "Platform" && ContainsString(options, "Web") {
				selected = "Web"
			}
			if name == "Audience" && ContainsString(options, "3P") {
				selected = "3P"
			}
		}
		filters = append(filters, TokenFilter{Name: name, Key: "context." + name, Selected: selected, Options: options})

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

func ContextOptionOrder(name, value string) int {
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

func ContainsString(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func ColorCellClass(color string) string {
	if color != "" {
		return " tokens-token-color"
	}
	return ""
}

func TokenColorTextColor(color string) string {
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
