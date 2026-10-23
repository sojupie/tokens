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
	return subjectTableDescription(subject, tokenSet{All: true})
}

type tokenSetMenuGroup struct {
	Label string
	Sets  []tokenSet
}

func tokenSetMenuGroups(subject publish.SubjectView, sets []tokenSet) []tokenSetMenuGroup {
	var groups []tokenSetMenuGroup
	indexes := make(map[string]int)
	for _, set := range sets {
		label := sentenceCase(tokenPhrase(set.component))
		if subject.Key.Kind == publish.SubjectComponent {
			label = componentName(set.component)
		}
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
