package web

import (
	"net/url"
	"strings"

	"github.com/sojupie/m3-source/internal/publish"
)

func tokenSetPath(set tokenSet) string {
	parts := []string{set.component, strings.ToLower(set.category)}
	if set.variant != "" {
		parts = append(parts, set.variant)
	}
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return strings.Join(parts, "/")
}

func tokenSetURL(subject publish.SubjectView, set tokenSet) string {
	return subjectURL(subject) + "/sets/" + tokenSetPath(set)
}

func subjectTokenDescription(subject publish.SubjectView) string {
	return "Explore Material Design tokens for " + subject.Name + ", including token references, values, contexts and change history. Download the tokens as JSON."
}
