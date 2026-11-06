package presentation

import (
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/sojupie/tokens/source/publish"
)

func TokenSetPath(set TokenSet) string {
	if set.All {
		return "all"
	}
	if set.SourceName != "" {
		return set.SourceName
	}
	parts := []string{strings.Join(strings.Fields(strings.ToLower(set.Component)), "-"), strings.ToLower(set.Category)}
	if set.Variant != "" {
		parts = append(parts, set.Variant)
	}
	return strings.Join(parts, "/")
}

func TokenSetURL(subject publish.SubjectView, set TokenSet) string {
	if set.All {
		return SubjectURL(subject) + "/sets/all"
	}
	parts := strings.Split(TokenSetPath(set), "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return SubjectURL(subject) + "/sets/" + strings.Join(parts, "/")
}

type TokenSetMenuGroup struct {
	Label string
	Sets  []TokenSet
}

func TokenSetMenuGroups(subject publish.SubjectView, sets []TokenSet) []TokenSetMenuGroup {
	var groups []TokenSetMenuGroup
	indexes := make(map[string]int)
	for _, set := range sets {
		label := SentenceCase(TokenPhrase(set.Component))
		if subject.Key.Kind == publish.SubjectComponent {
			label = ComponentName(set.Component)
		}
		if label == "" {
			label = "System"
		}
		index, found := indexes[label]
		if !found {
			index = len(groups)
			indexes[label] = index
			groups = append(groups, TokenSetMenuGroup{Label: label})
		}
		groups[index].Sets = append(groups[index].Sets, set)
	}
	return groups
}

func TokenSetLinkURL(subject publish.SubjectView, set TokenSet, browser TokenBrowser) string {
	query := make(url.Values)
	if browser.Query != "" {
		query.Set("q", browser.Query)
	}
	if len(browser.Contexts) > 0 {
		for _, filter := range TokenFilters(set.SourceTokens, browser.Contexts) {
			if browser.Contexts.Get(filter.Key) == filter.Selected {
				query.Set(filter.Key, filter.Selected)
			}
		}
	}
	path := TokenSetURL(subject, set)
	if IsDefaultTokenSet(subject, set) {
		path = SubjectURL(subject)
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return path
}

func CurrentTokenSet(selected bool) string {
	if selected {
		return "page"
	}
	return "false"
}

func TokenPhrase(value string) string {
	words := strings.Fields(strings.ToLower(DisplayWords(value)))
	for i, word := range words {
		if word == "fab" {
			words[i] = "FAB"
		} else if word == "ios" {
			words[i] = "iOS"
		}
	}
	return strings.Join(words, " ")
}

func SentenceCase(value string) string {
	first, size := utf8.DecodeRuneInString(value)
	if size == 0 {
		return ""
	}
	return string(unicode.ToUpper(first)) + value[size:]
}
