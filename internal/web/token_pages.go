package web

import (
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/sojupie/m3-source/internal/publish"
)

func tokenSetPath(set tokenSet) string {
	parts := []string{strings.Join(strings.Fields(strings.ToLower(set.component)), "-"), strings.ToLower(set.category)}
	if set.variant != "" {
		parts = append(parts, set.variant)
	}
	return strings.Join(parts, "/")
}

func tokenSetURL(subject publish.SubjectView, set tokenSet) string {
	parts := strings.Split(tokenSetPath(set), "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return subjectURL(subject) + "/sets/" + strings.Join(parts, "/")
}

func subjectTokenDescription(subject publish.SubjectView) string {
	return "Material Design tokens for " + subject.Name + ". View named references and values by context, download JSON or see what changed."
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
	label := sentenceCase(tokenPhrase(set.category))
	if set.variant == "toggle" || set.category == "SIZE" && set.variant != "" {
		label = sentenceCase(tokenPhrase(set.variant)) + " · " + label
	}
	return label
}

func tokenSetTitle(set tokenSet) string {
	subset := tokenPhrase(set.category) + " tokens"
	if set.category == "COLOR" {
		subset = "colors"
	}
	if set.variant == "toggle" || set.category == "SIZE" && set.variant != "" {
		subset = tokenPhrase(set.variant) + " " + subset
	}
	return sentenceCase(tokenSetComponent(set)) + ": " + subset + " — Material Design"
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
	text := sentenceCase(kind) + " tokens for the Material Design " + tokenSetComponent(set)
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
