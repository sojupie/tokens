package web

import (
	"context"
	"regexp"
	"strings"

	"github.com/jinzhu/inflection"
	"github.com/sojupie/tokens/source/publish"
)

var nameSuffix = regexp.MustCompile(`\s+\(([^()]*)\)$`)
var tableLabelSeparator = regexp.MustCompile(`\s+[-–—]\s+`)

// Singularize the final noun, not modifiers or source identifiers. Inflection
// uses established English regex rules, including irregular and uncountable nouns.
func componentName(name string) string {
	suffix := ""
	if match := nameSuffix.FindStringSubmatch(name); match != nil {
		suffix = " (" + titleTokenPhrase(inflection.Singular(strings.ToLower(match[1]))) + ")"
		name = name[:len(name)-len(match[0])]
	}
	words := strings.Fields(tokenPhrase(name))
	if len(words) == 0 {
		return strings.TrimSpace(name)
	}
	words[len(words)-1] = inflection.Singular(strings.ToLower(words[len(words)-1]))
	return titleTokenPhrase(strings.Join(words, " ")) + suffix
}

func subjectName(subject publish.SubjectView) string {
	if subject.Key.Kind == publish.SubjectComponent {
		return componentName(subject.Name)
	}
	return subject.Name
}

func subjectTokenTitle(subject publish.SubjectView) string {
	return `Material Design 3 "` + subjectName(subject) + `" tokens`
}

// Source labels remain intact in the publication. Only their presentation is
// shortened here; unfamiliar labels retain their wording rather than inferred variants.
func tokenTableQualifier(subject publish.SubjectView, set tokenSet) string {
	if set.All {
		return ""
	}
	label := strings.TrimSpace(set.Label)
	status := ""
	if strings.HasPrefix(label, "[Deprecated] ") {
		status = "Deprecated"
		label = strings.TrimPrefix(label, "[Deprecated] ")
	}
	parts := tableLabelSeparator.Split(label, -1)
	name := componentName(parts[0])
	if strings.EqualFold(name, subjectName(subject)) {
		parts = parts[1:]
	} else if strings.EqualFold(strings.TrimSuffix(name, " (Baseline)"), subjectName(subject)) && strings.HasSuffix(name, " (Baseline)") {
		parts = append([]string{"Baseline"}, parts[1:]...)
	}
	for i := range parts {
		parts[i] = titleTokenPhrase(parts[i])
	}
	if len(parts) == 2 && (parts[0] == "Color" || parts[0] == "Size") {
		parts[0], parts[1] = parts[1], parts[0]
	}
	if status != "" {
		parts = append(parts, status)
	}
	if len(parts) == 0 {
		return "Common"
	}
	return strings.Join(parts, " / ")
}

func subjectTableTitle(subject publish.SubjectView, set tokenSet) string {
	if set.All {
		return subjectTokenTitle(subject)
	}
	return subjectTokenTitle(subject) + ": " + tokenTableQualifier(subject, set)
}

func subjectTableDescription(subject publish.SubjectView, set tokenSet) string {
	text := `Material Design 3 "` + subjectName(subject) + `" tokens`
	if !set.All {
		text += " for " + tokenTableQualifier(subject, set)
	}
	return text + ". Compare values by context, follow references and download JSON."
}

func subjectChangesTitle(subject publish.SubjectView) string {
	return `Material Design 3 "` + subjectName(subject) + `" token changes`
}

func tokenTableCanonicalPath(subject publish.SubjectView, set tokenSet) string {
	if set.All || len(tokenSets(subject)) == 1 {
		return subjectURL(subject)
	}
	return tokenSetURL(subject, set)
}

func changePageHeading(ctx context.Context, subject string) string {
	if heading := metadata(ctx).Heading; heading != "" {
		return heading
	}
	return timelineTitle(subject)
}

func selectedTableLabel(subject publish.SubjectView, set tokenSet) string {
	if set.All {
		return "All tokens"
	}
	return tokenTableQualifier(subject, set)
}

func titleTokenPhrase(value string) string {
	words := strings.Fields(tokenPhrase(value))
	for i, word := range words {
		if word != "iOS" {
			words[i] = sentenceCase(word)
		}
	}
	return strings.Join(words, " ")
}
