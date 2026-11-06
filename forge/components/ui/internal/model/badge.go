package model

import (
	"errors"
	"github.com/a-h/templ"
	"strconv"
	"strings"
)

// BadgeContent bounds the visual label. The owning destination supplies the
// full notification count in its accessible name when it exceeds this label.
func BadgeContent(content string) string {
	content = strings.TrimSpace(content)
	if count, err := strconv.ParseUint(content, 10, 64); (err == nil && count > 999) || errors.Is(err, strconv.ErrRange) {
		return "999+"
	}
	runes := []rune(content)
	if len(runes) > 4 {
		return string(runes[:4])
	}
	return content
}

//forge:internal
func BadgeAttributes(variant string, content string, attrs templ.Attributes) templ.Attributes {
	result := templ.Attributes{"class": "md-badge", "data-variant": variant, "aria-hidden": "true", "data-characters": strconv.Itoa(len([]rune(BadgeContent(content))))}
	return Attributes(result, attrs)
}
