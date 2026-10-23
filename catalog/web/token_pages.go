package web

import (
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/sojupie/tokens/source/publish"
)

func tokenSetPath(set tokenSet) string {
	if set.sourceName != "" {
		return set.sourceName
	}
	parts := []string{strings.Join(strings.Fields(strings.ToLower(set.component)), "-"), strings.ToLower(set.category)}
	if set.variant != "" {
		parts = append(parts, set.variant)
	}
	return strings.Join(parts, "/")
}

func tokenSetURL(subject publish.SubjectView, set tokenSet) string {
	if set.All {
		return subjectURL(subject)
	}
	parts := strings.Split(tokenSetPath(set), "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return subjectURL(subject) + "/sets/" + strings.Join(parts, "/")
}

// Old name-derived table URLs can redirect only when every token belongs to
// one source table. A merged legacy table has no single canonical replacement.
func legacyTokenSet(subject publish.SubjectView, path string) (tokenSet, bool) {
	legacy := subject
	legacy.Tokens = append([]publish.TokenView(nil), subject.Tokens...)
	sourceSets := make(map[string]string)
	for i, token := range legacy.Tokens {
		if token.Presentation != nil {
			sourceSets[token.SourceName] = token.Presentation.SetName
		}
		legacy.Tokens[i].Presentation = nil
	}
	for _, set := range tokenSets(legacy) {
		if tokenSetPath(set) != path {
			continue
		}
		name := ""
		for _, token := range set.sourceTokens {
			source := sourceSets[token.SourceName]
			if source == "" || name != "" && name != source {
				return tokenSet{}, false
			}
			name = source
		}
		for _, source := range tokenSets(subject) {
			if source.sourceName == name && name != "" {
				return source, true
			}
		}
	}
	return tokenSet{}, false
}

func subjectTokenDescription(subject publish.SubjectView) string {
	kind := strings.ToLower(subjectKindLabel(subject.Key.Kind))
	return "Explore Material Design 3 " + kind + " tokens for " + subject.Name + ". Compare values by context, follow named references, download JSON and review observed changes."
}

type tokenSetMenuGroup struct {
	Label string
	Sets  []tokenSet
}

func tokenSetMenuGroups(sets []tokenSet) []tokenSetMenuGroup {
	var groups []tokenSetMenuGroup
	indexes := make(map[string]int)
	for _, set := range sets {
		label := sentenceCase(tokenPhrase(set.component))
		if label == "" {
			label = "System"
		}
		index, found := indexes[label]
		if !found {
			index = len(groups)
			indexes[label] = index
			groups = append(groups, tokenSetMenuGroup{Label: label})
		}
		groups[index].Sets = append(groups[index].Sets, set)
	}
	return groups
}

func tokenSetMenuLabel(set tokenSet) string {
	if set.sourceName != "" {
		return set.Label
	}
	if set.variant != "" && set.variant != "toggle" && set.category != "SIZE" {
		return sentenceCase(tokenPhrase(set.variant)) + " · " + sentenceCase(tokenPhrase(set.category))
	}
	return tokenSetLabel(set)
}

func tokenSetLinkURL(subject publish.SubjectView, set tokenSet, browser tokenBrowser) string {
	query := make(url.Values)
	if browser.Query != "" {
		query.Set("q", browser.Query)
	}
	if len(browser.Contexts) > 0 {
		for _, filter := range tokenFilters(set.sourceTokens, browser.Contexts) {
			if browser.Contexts.Get(filter.Key) == filter.Selected {
				query.Set(filter.Key, filter.Selected)
			}
		}
	}
	path := tokenSetURL(subject, set)
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return path
}

func currentTokenSet(selected bool) string {
	if selected {
		return "page"
	}
	return "false"
}

func tokenPhrase(value string) string {
	words := strings.Fields(strings.ToLower(displayWords(value)))
	for i, word := range words {
		if word == "fab" {
			words[i] = "FAB"
		} else if word == "ios" {
			words[i] = "iOS"
		}
	}
	return strings.Join(words, " ")
}

func sentenceCase(value string) string {
	first, size := utf8.DecodeRuneInString(value)
	if size == 0 {
		return ""
	}
	return string(unicode.ToUpper(first)) + value[size:]
}

func tokenSetComponent(set tokenSet) string {
	component := tokenPhrase(set.component)
	if component == "" {
		component = "system"
	}
	if set.variant != "" && set.variant != "toggle" && set.category != "SIZE" {
		component = tokenPhrase(set.variant) + " " + component
	}
	return component
}

func tokenSetLabel(set tokenSet) string {
	if set.sourceName != "" {
		return set.Label
	}
	label := sentenceCase(tokenPhrase(set.category))
	if set.variant == "toggle" || set.category == "SIZE" && set.variant != "" {
		label = sentenceCase(tokenPhrase(set.variant)) + " · " + label
	}
	return label
}

func tokenSetTitle(set tokenSet) string {
	if set.sourceName != "" {
		return set.Label + " tokens — Material Design 3"
	}
	subset := tokenPhrase(set.category) + " tokens"
	if set.category == "COLOR" {
		subset = "colors"
	}
	if set.variant == "toggle" || set.category == "SIZE" && set.variant != "" {
		subset = tokenPhrase(set.variant) + " " + subset
	}
	return sentenceCase(tokenSetComponent(set)) + ": " + subset + " — Material Design 3"
}

func tokenSetDescription(set tokenSet) string {
	properties := make(map[string]bool)
	for _, token := range set.sourceTokens {
		properties[strings.ToUpper(token.Category)] = true
	}
	var labels []string
	for _, property := range []struct{ key, label string }{{"COLOR", "color"}, {"OPACITY", "opacity"}, {"ELEVATION", "elevation"}, {"SHAPE", "shape"}, {"TYPOGRAPHY", "typography"}} {
		if properties[property.key] {
			labels = append(labels, property.label)
		}
	}
	if properties["LENGTH"] || properties["DIMENSION"] {
		labels = append(labels, "dimension")
	}
	kind := tokenPhrase(set.category)
	if len(labels) > 0 {
		kind = strings.Join(labels, ", ")
		if len(labels) > 1 {
			kind = strings.Join(labels[:len(labels)-1], ", ") + " and " + labels[len(labels)-1]
		}
	}
	text := sentenceCase(kind) + " tokens for the Material Design 3 " + tokenSetComponent(set)
	if set.category == "SIZE" && set.variant != "" {
		text += " in " + tokenPhrase(set.variant) + " size"
	}
	if set.variant == "toggle" {
		text += " when used as a toggle"
		selected, unselected := false, false
		for _, token := range set.sourceTokens {
			for _, part := range strings.Split(token.SourceName, ".") {
				selected = selected || part == "selected"
				unselected = unselected || part == "unselected"
			}
		}
		if selected && unselected {
			text += ", in selected and unselected states"
		}
	}
	return text + ". View references and values by context, or download JSON."
}
