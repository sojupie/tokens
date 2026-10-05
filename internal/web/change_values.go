package web

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"unicode"

	"github.com/sojupie/m3-source/internal/publish"
)

type changeValue struct {
	Field  string
	Before string
	After  string
}

func changedValues(change publish.ChangeView) []changeValue {
	var before, after any
	if change.Type != "changed" || json.Unmarshal([]byte(change.Before), &before) != nil || json.Unmarshal([]byte(change.After), &after) != nil {
		return []changeValue{{"Value", changeSummaryValue(change.Before), changeSummaryValue(change.After)}}
	}
	var values []changeValue
	var compare func(string, any, any)
	compare = func(field string, before, after any) {
		if reflect.DeepEqual(before, after) {
			return
		}
		left, leftMap := before.(map[string]any)
		right, rightMap := after.(map[string]any)
		// Keep colors and measurements together so their units remain readable.
		if leftMap && rightMap && left["unit"] == nil && right["unit"] == nil && left["red"] == nil && right["red"] == nil {
			keys := make(map[string]bool)
			for key := range left {
				keys[key] = true
			}
			for key := range right {
				keys[key] = true
			}
			ordered := make([]string, 0, len(keys))
			for key := range keys {
				ordered = append(ordered, key)
			}
			sort.Strings(ordered)
			for _, key := range ordered {
				label := changeFieldLabel(key)
				if field != "" {
					label = field + " · " + label
				}
				compare(label, left[key], right[key])
			}
			return
		}
		if field == "" {
			field = "Value"
		}
		text := func(value any) string {
			if value == nil {
				return "—"
			}
			if fields, ok := value.(map[string]any); ok {
				if fields["unit"] != nil {
					if text, ok := dimensionValueText(fields); ok {
						return text
					}
				}
				if fields["red"] != nil {
					if text, ok := colorValueText(fields); ok {
						return text
					}
				}
			}
			return formattedTokenValue(value)
		}
		values = append(values, changeValue{field, text(before), text(after)})
	}
	compare("", before, after)
	return values
}

func changeFieldLabel(field string) string {
	var label strings.Builder
	previous := rune(0)
	for _, character := range field {
		if unicode.IsUpper(character) && unicode.IsLower(previous) {
			label.WriteByte(' ')
		}
		label.WriteRune(character)
		previous = character
	}
	return displayWords(label.String())
}
