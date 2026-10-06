package web

import (
	"net/url"
	"strings"

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
	return "Explore Material Design tokens for " + subject.Name + ", including token references, values, contexts and change history. Download the tokens as JSON."
}
