package presentation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unicode"

	"github.com/sojupie/tokens/source/publish"
	"github.com/sojupie/tokens/source/tokensnapshot"
)

type ChangeValue struct {
	Field  string
	Before string
	After  string
}

func ChangedValues(change publish.ChangeView) []ChangeValue {
	before, beforeOK := HistoricalChangeValue(change.Before)
	after, afterOK := HistoricalChangeValue(change.After)
	if !beforeOK && !afterOK {
		return ChangedRawValues(change)
	}
	copy := change
	copy.Before, copy.After = "", ""
	if beforeOK {
		copy.Before = string(before.Value)
	}
	if afterOK {
		copy.After = string(after.Value)
	}
	values := ChangedRawValues(copy)
	if change.Type == "changed" && !reflect.DeepEqual(before.References, after.References) {
		values = append(values, ChangeValue{"References", ChangeReferences(before.References), ChangeReferences(after.References)})
	}
	if beforeOK && afterOK && before.Default != after.Default {
		values = append(values, ChangeValue{"Default context", fmt.Sprint(before.Default), fmt.Sprint(after.Default)})
	}
	return values
}

func HistoricalChangeValue(raw string) (tokensnapshot.ContextualValue, bool) {
	var fields map[string]json.RawMessage
	var value tokensnapshot.ContextualValue
	if json.Unmarshal([]byte(raw), &fields) != nil || fields["context"] == nil || fields["value"] == nil {
		return value, false
	}
	return value, json.Unmarshal([]byte(raw), &value) == nil
}

func ChangeContext(change publish.ChangeView) string {
	if change.Kind != "contextualReferenceTree" {
		return ""
	}
	value, ok := HistoricalChangeValue(change.After)
	if !ok {
		value, ok = HistoricalChangeValue(change.Before)
	}
	if !ok {
		return "Context not recorded"
	}
	labels := append([]string(nil), value.Context...)
	if len(value.UnknownContextTags) > 0 {
		labels = append(labels, "Unlabelled context")
	}
	if len(labels) == 0 {
		return "No context tags"
	}
	return strings.Join(labels, " · ")
}

func ChangeReferences(refs []tokensnapshot.NamedReference) string {
	if len(refs) == 0 {
		return "—"
	}
	paths := make([]string, 0, len(refs))
	for _, ref := range refs {
		path := ref.Name
		if len(ref.References) > 0 {
			path += " → " + ChangeReferences(ref.References)
		}
		paths = append(paths, path)
	}
	return strings.Join(paths, " / ")
}

func ChangedRawValues(change publish.ChangeView) []ChangeValue {
	var before, after any
	if (change.Before != "" && json.Unmarshal([]byte(change.Before), &before) != nil) || (change.After != "" && json.Unmarshal([]byte(change.After), &after) != nil) {
		return []ChangeValue{{"Value", ChangeSummaryValue(change.Before), ChangeSummaryValue(change.After)}}
	}
	var values []ChangeValue
	var compare func(string, any, any)
	compare = func(field string, before, after any) {
		if reflect.DeepEqual(before, after) {
			return
		}
		left, leftMap := before.(map[string]any)
		right, rightMap := after.(map[string]any)

		if (leftMap || rightMap) && (leftMap || before == nil) && (rightMap || after == nil) && left["unit"] == nil && right["unit"] == nil && left["red"] == nil && right["red"] == nil && left["undefined"] == nil && right["undefined"] == nil {
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
				label := ChangeFieldLabel(key)
				if label == "" {
					label = field
				} else if field != "" {
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
				if fields["undefined"] == true {
					return "Not resolved"
				}
				if fields["unit"] != nil {
					if text, ok := DimensionValueText(fields); ok {
						return text
					}
					if unit, ok := fields["unit"].(string); ok {
						return "Not specified (" + DisplayUnit(unit) + ")"
					}
				}
				if fields["red"] != nil {
					if text, ok := ColorValueText(fields); ok {
						return text
					}
				}
			}
			if field == "Opacity" || strings.HasSuffix(field, " · Opacity") {
				if text, ok := FormattedNamedValue("opacity", value); ok {
					return text
				}
			}
			return FormattedTokenValue(value)
		}
		values = append(values, ChangeValue{field, text(before), text(after)})
	}
	compare("", before, after)
	return values
}

func ChangeFieldLabel(field string) string {
	switch field {
	case "type", "values", "axisValue":
		return ""
	case "fontNames":
		return "Font family"
	case "fontTracking":
		return "Letter spacing"
	case "variationAxes":
		return "Font axes"
	case "wght", "wdth", "opsz", "slnt", "GRAD", "CRSV", "FILL", "ROND", "HEXP":
		return field
	}
	var label strings.Builder
	previous := rune(0)
	for _, character := range field {
		if unicode.IsUpper(character) && unicode.IsLower(previous) {
			label.WriteByte(' ')
		}
		label.WriteRune(character)
		previous = character
	}
	text := DisplayWords(label.String())
	if first, rest, ok := strings.Cut(text, " "); ok {
		return first + " " + strings.ToLower(rest)
	}
	return text
}
