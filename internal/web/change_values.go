package web

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unicode"

	"github.com/sojupie/m3-source/internal/publish"
	"github.com/sojupie/m3-source/internal/tokensnapshot"
)

type changeValue struct {
	Field  string
	Before string
	After  string
}

func changedValues(change publish.ChangeView) []changeValue {
	before, beforeOK := historicalChangeValue(change.Before)
	after, afterOK := historicalChangeValue(change.After)
	if !beforeOK && !afterOK {
		return changedRawValues(change)
	}
	copy := change
	copy.Before, copy.After = "", ""
	if beforeOK {
		copy.Before = string(before.Value)
	}
	if afterOK {
		copy.After = string(after.Value)
	}
	values := changedRawValues(copy)
	if change.Type == "changed" && !reflect.DeepEqual(before.References, after.References) {
		values = append(values, changeValue{"References", changeReferences(before.References), changeReferences(after.References)})
	}
	if beforeOK && afterOK && before.Default != after.Default {
		values = append(values, changeValue{"Default context", fmt.Sprint(before.Default), fmt.Sprint(after.Default)})
	}
	return values
}

func historicalChangeValue(raw string) (tokensnapshot.ContextualValue, bool) {
	var fields map[string]json.RawMessage
	var value tokensnapshot.ContextualValue
	if json.Unmarshal([]byte(raw), &fields) != nil || fields["context"] == nil || fields["value"] == nil {
		return value, false
	}
	return value, json.Unmarshal([]byte(raw), &value) == nil
}

func changeContext(change publish.ChangeView) string {
	if change.Kind != "contextualReferenceTree" {
		return ""
	}
	value, ok := historicalChangeValue(change.After)
	if !ok {
		value, ok = historicalChangeValue(change.Before)
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

func changeReferences(refs []tokensnapshot.NamedReference) string {
	if len(refs) == 0 {
		return "—"
	}
	paths := make([]string, 0, len(refs))
	for _, ref := range refs {
		path := ref.Name
		if len(ref.References) > 0 {
			path += " → " + changeReferences(ref.References)
		}
		paths = append(paths, path)
	}
	return strings.Join(paths, " / ")
}

func changedRawValues(change publish.ChangeView) []changeValue {
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
	switch field {
	case "type", "values":
		return ""
	case "fontNames":
		return "Font family"
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
	text := displayWords(label.String())
	if first, rest, ok := strings.Cut(text, " "); ok {
		return first + " " + strings.ToLower(rest)
	}
	return text
}
