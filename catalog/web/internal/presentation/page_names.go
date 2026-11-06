package presentation

import (
	"context"
	"regexp"
	"strings"

	"github.com/jinzhu/inflection"
	"github.com/sojupie/tokens/source/publish"
)

var NameSuffix = regexp.MustCompile(`\s+\(([^()]*)\)$`)
var TableLabelSeparator = regexp.MustCompile(`\s+[-–—]\s+`)

// Singularize the final noun, not modifiers or source identifiers. Inflection
// uses established English regex rules, including irregular and uncountable nouns.
func ComponentName(name string) string {
	suffix := ""
	if match := NameSuffix.FindStringSubmatch(name); match != nil {
		suffix = " (" + TitleTokenPhrase(inflection.Singular(strings.ToLower(match[1]))) + ")"
		name = name[:len(name)-len(match[0])]
	}
	words := strings.Fields(TokenPhrase(name))
	if len(words) == 0 {
		return strings.TrimSpace(name)
	}
	words[len(words)-1] = inflection.Singular(strings.ToLower(words[len(words)-1]))
	return TitleTokenPhrase(strings.Join(words, " ")) + suffix
}

func SubjectName(subject publish.SubjectView) string {
	if subject.Key.Kind == publish.SubjectComponent {
		return ComponentName(subject.Name)
	}
	return subject.Name
}

// Source labels remain intact in the publication. Only their presentation is
// shortened here; unfamiliar labels retain their wording rather than inferred variants.
func TokenTableQualifier(subject publish.SubjectView, set TokenSet) string {
	if set.All {
		return ""
	}
	label := strings.TrimSpace(set.Label)
	status := ""
	if strings.HasPrefix(label, "[Deprecated] ") {
		status = "Deprecated"
		label = strings.TrimPrefix(label, "[Deprecated] ")
	}
	parts := TableLabelSeparator.Split(label, -1)
	name := ComponentName(parts[0])
	if strings.EqualFold(name, SubjectName(subject)) {
		parts = parts[1:]
	} else if strings.EqualFold(strings.TrimSuffix(name, " (Baseline)"), SubjectName(subject)) && strings.HasSuffix(name, " (Baseline)") {
		parts = append([]string{"Baseline"}, parts[1:]...)
	}
	for i := range parts {
		parts[i] = TitleTokenPhrase(parts[i])
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

func ChangePageHeading(ctx context.Context, subject string) string {
	if heading := Metadata(ctx).Heading; heading != "" {
		return heading
	}
	return TimelineTitle(subject)
}

func SelectedTableLabel(subject publish.SubjectView, set TokenSet) string {
	if set.All {
		return "All tokens"
	}
	return TokenTableQualifier(subject, set)
}

func TitleTokenPhrase(value string) string {
	words := strings.Fields(TokenPhrase(value))
	for i, word := range words {
		if word != "iOS" {
			words[i] = SentenceCase(word)
		}
	}
	return strings.Join(words, " ")
}

func AllTokenSet(subject publish.SubjectView) TokenSet {
	return TokenSet{Key: "all", All: true, Label: "All tokens", SourceTokens: TokenTableTokens(subject)}
}

func IsDefaultTokenSet(subject publish.SubjectView, set TokenSet) bool {
	sets := TokenSets(subject)
	return !set.All && len(sets) > 0 && set.Key == sets[0].Key
}
