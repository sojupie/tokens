package ui

import (
	"github.com/a-h/templ"
	"strings"
)

// Attributes extend a component for application hooks and playground states.
// Component options own normal semantics and geometry; callers need no CSS classes.
func attributes(base, extra templ.Attributes) templ.Attributes {
	result := make(templ.Attributes, len(base)+len(extra))
	for k, v := range base {
		result[k] = v
	}
	for k, v := range extra {
		if k == "class" {
			classes := strings.Fields(templ.Classes(base[k], v).String())
			seen := map[string]bool{}
			unique := classes[:0]
			for _, class := range classes {
				if !seen[class] {
					seen[class] = true
					unique = append(unique, class)
				}
			}
			result[k] = strings.Join(unique, " ")
		} else {
			result[k] = v
		}
	}
	return result
}
func fallback(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

type Size string

const (
	ExtraSmall Size = "xsmall"
	Small      Size = "small"
	Medium     Size = "medium"
	Large      Size = "large"
	ExtraLarge Size = "xlarge"
)

func firstAttrs(values []templ.Attributes) templ.Attributes {
	if len(values) > 0 {
		return values[0]
	}
	return nil
}
