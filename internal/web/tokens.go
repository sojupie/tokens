package web

import (
	"net/url"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/sojupie/m3-source/internal/publish"
)

type tokenGroup struct {
	Name   string
	Tokens []tokenRow
}

type tokenRow struct {
	Name        string
	SourceName  string
	Category    string
	Lifecycle   string
	Value       publish.TokenValueView
	HasValue    bool
	OtherValues []publish.TokenValueView
}

func subjectTokenGroups(subject publish.SubjectView, query string) []tokenGroup {
	query = strings.ToLower(strings.TrimSpace(query))
	groups := make(map[string][]tokenRow)
	componentPrefix := ""
	if subject.Key.Kind == publish.SubjectComponent {
		componentPrefix = "md.comp." + subject.Key.Slug + "."
	}

	for _, token := range subject.Tokens {
		groupName, displayName := tokenLabels(token.SourceName, componentPrefix)
		row := tokenRow{
			Name:       displayName,
			SourceName: token.SourceName,
			Category:   displayWords(strings.ToLower(token.Category)),
			Lifecycle:  displayWords(token.Lifecycle),
		}
		if strings.EqualFold(token.Lifecycle, "active") {
			row.Lifecycle = ""
		}
		if len(token.Values) > 0 {
			primary := 0
			for index := range token.Values {
				if token.Values[index].IsDefaultContext {
					primary = index
					break
				}
			}
			row.Value = token.Values[primary]
			row.HasValue = true
			for index, value := range token.Values {
				if index != primary {
					row.OtherValues = append(row.OtherValues, value)
				}
			}
		}
		if query != "" && !tokenMatchesQuery(groupName, row, token.Values, query) {
			continue
		}
		groups[groupName] = append(groups[groupName], row)
	}

	names := make([]string, 0, len(groups))
	for name := range groups {
		names = append(names, name)
	}
	sort.Strings(names)
	result := make([]tokenGroup, 0, len(names))
	for _, name := range names {
		result = append(result, tokenGroup{Name: name, Tokens: groups[name]})
	}
	return result
}

func tokenVariantValues(subject publish.SubjectView, sourceName string) []publish.TokenValueView {
	for _, token := range subject.Tokens {
		if token.SourceName != sourceName || len(token.Values) < 2 {
			continue
		}
		primary := 0
		for index := range token.Values {
			if token.Values[index].IsDefaultContext {
				primary = index
				break
			}
		}
		values := make([]publish.TokenValueView, 0, len(token.Values)-1)
		for index, value := range token.Values {
			if index != primary {
				values = append(values, value)
			}
		}
		return values
	}
	return nil
}

func tokenVariantsURL(subject publish.SubjectView, sourceName string) string {
	query := url.Values{"variants": []string{sourceName}}
	return subjectURL(subject) + "?" + query.Encode()
}

func tokenLabels(sourceName, componentPrefix string) (group, name string) {
	if componentPrefix == "" || !strings.HasPrefix(sourceName, componentPrefix) {
		return "Tokens", sourceName
	}
	parts := strings.Split(strings.TrimPrefix(sourceName, componentPrefix), ".")
	if len(parts) == 0 || parts[0] == "" {
		return "Other", sourceName
	}
	group = displayWords(parts[0])
	nameParts := make([]string, 0, len(parts)-1)
	for _, part := range parts[1:] {
		nameParts = append(nameParts, displayWords(part))
	}
	name = strings.Join(nameParts, " · ")
	if name == "" {
		name = group
	}
	return group, name
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

func tokenMatchesQuery(groupName string, row tokenRow, values []publish.TokenValueView, query string) bool {
	if strings.Contains(strings.ToLower(groupName), query) ||
		strings.Contains(strings.ToLower(row.Name), query) ||
		strings.Contains(strings.ToLower(row.SourceName), query) ||
		strings.Contains(strings.ToLower(row.Category), query) {
		return true
	}
	for _, value := range values {
		if strings.Contains(strings.ToLower(tokenContextLabel(value)), query) ||
			strings.Contains(strings.ToLower(tokenValueText(value.Value)), query) {
			return true
		}
	}
	return false
}
