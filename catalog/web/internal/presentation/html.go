package presentation

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/sojupie/tokens/source/publish"
)

func SubjectListPath(styles bool) string {
	if styles {
		return "/styles"
	}
	return "/components"
}

func SubjectURL(subject publish.SubjectView) string {
	section := "styles"
	if subject.Key.Kind == publish.SubjectComponent {
		section = "components"
	}
	parts := strings.Split(subject.Key.Slug, "/")
	for index, part := range parts {
		parts[index] = url.PathEscape(part)
	}
	return "/" + section + "/" + strings.Join(parts, "/")
}

func SubjectChangesURL(subject publish.SubjectView) string {
	return SubjectURL(subject) + "/changes"
}

func SubjectKindLabel(kind publish.SubjectKind) string {
	if kind == publish.SubjectComponent {
		return "Component"
	}
	if kind == publish.SubjectStyle {
		return "Style"
	}
	return string(kind)
}

func TokenValueText(value json.RawMessage) string {
	var decoded any
	if json.Unmarshal(value, &decoded) != nil {
		return string(value)
	}
	return FormattedTokenValue(decoded)
}

func FormattedTokenValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case float64:
		return strconv.FormatFloat(typed, 'g', -1, 64)
	case bool:
		return strconv.FormatBool(typed)
	case nil:
		return "No value published"
	case []any:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			values = append(values, FormattedTokenValue(item))
		}
		return strings.Join(values, ", ")
	case map[string]any:
		if len(typed) == 1 {
			for name, value := range typed {
				if formatted, ok := FormattedNamedValue(name, value); ok {
					return formatted
				}
				return DisplayWords(name) + ": " + FormattedTokenValue(value)
			}
		}
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		fields := make([]string, 0, len(keys))
		for _, key := range keys {
			value := FormattedTokenValue(typed[key])
			if formatted, ok := FormattedNamedValue(key, typed[key]); ok {
				value = formatted
			}
			fields = append(fields, DisplayWords(key)+": "+value)
		}
		return strings.Join(fields, " · ")
	}
	return "No value published"
}

func FormattedNamedValue(name string, value any) (string, bool) {
	fields, ok := value.(map[string]any)
	switch name {
	case "color":
		if color, ok := value.(map[string]any); ok {
			return ColorValueText(color)
		}
	case "dimension", "length", "elevation", "fontSize", "fontTracking", "lineHeight", "defaultSize",
		"topLeft", "topRight", "bottomLeft", "bottomRight":
		if ok {
			if text, ok := DimensionValueText(fields); ok {
				return text, true
			}
			unit, _ := fields["unit"].(string)
			if unit != "" {
				return "No value specified (" + DisplayUnit(unit) + ")", true
			}
			return "No numeric value specified", true
		}
	case "fontNames":
		if names, ok := FontNamesValue(value); ok {
			return names, true
		}
	case "fontWeight":
		if weight, ok := NumberValue(value); ok {
			return FormatNumber(weight), true
		}
	case "opacity":
		if opacity, ok := NumberValue(value); ok {
			return FormatNumber(opacity*100) + "%", true
		}
	case "shape":
		if ok {
			return ShapeValueText(fields), true
		}
	case "type":
		if ok {
			return TypographyValueText(fields), true
		}
	}
	return "", false
}

func ColorValueText(color map[string]any) (string, bool) {
	red, green, blue, ok := ColorChannels(color)
	alpha, alphaOK := NumberValue(color["alpha"])
	if !ok {
		return "", false
	}
	if red <= 1 && green <= 1 && blue <= 1 {
		red *= 255
		green *= 255
		blue *= 255
	}
	red = TidyColorChannel(red)
	green = TidyColorChannel(green)
	blue = TidyColorChannel(blue)
	if !alphaOK {
		alpha = 1
	}
	alpha = TidyColorChannel(alpha)
	return fmt.Sprintf("rgba(%s, %s, %s, %s)",
		FormatNumber(red), FormatNumber(green), FormatNumber(blue), FormatNumber(alpha)), true
}

// Source color messages omit zero-valued channels. An empty placeholder is
// still unresolved; alpha (or any channel) must establish a concrete color.
func ColorChannels(color map[string]any) (float64, float64, float64, bool) {
	concrete := false
	for _, name := range []string{"red", "green", "blue", "alpha"} {
		if _, exists := color[name]; exists {
			concrete = true
		}
	}
	if !concrete {
		return 0, 0, 0, false
	}
	if value, exists := color["alpha"]; exists {
		alpha, ok := NumberValue(value)
		if !ok || alpha < 0 || alpha > 1 {
			return 0, 0, 0, false
		}
	}
	channels := [3]float64{}
	for index, name := range []string{"red", "green", "blue"} {
		if value, exists := color[name]; exists {
			channel, ok := NumberValue(value)
			if !ok || channel < 0 || channel > 255 {
				return 0, 0, 0, false
			}
			channels[index] = channel
		}
	}
	return channels[0], channels[1], channels[2], true
}

func DimensionValueText(dimension map[string]any) (string, bool) {
	return SourceDimensionText(dimension)
}

func TypographyValueText(fields map[string]any) string {
	parts := make([]string, 0, 6)
	if names, ok := FontNamesValue(fields["fontNames"]); ok {
		parts = append(parts, names)
	}
	if size, ok := FormattedNamedValue("fontSize", fields["fontSize"]); ok {
		parts = append(parts, size)
	}
	if weight, ok := NumberValue(fields["fontWeight"]); ok {
		parts = append(parts, "weight "+FormatNumber(weight))
	}
	if lineHeight, ok := FormattedNamedValue("lineHeight", fields["lineHeight"]); ok {
		parts = append(parts, "line height "+lineHeight)
	}
	if tracking, ok := FormattedNamedValue("fontTracking", fields["fontTracking"]); ok {
		parts = append(parts, "tracking "+tracking)
	}
	known := map[string]bool{"fontNames": true, "fontSize": true, "fontWeight": true, "lineHeight": true, "fontTracking": true}
	keys := make([]string, 0, len(fields))
	for key := range fields {
		if known[key] {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		parts = append(parts, DisplayWords(key)+": "+FormattedTokenValue(fields[key]))
	}
	return strings.Join(parts, " · ")
}

func FontNamesValue(value any) (string, bool) {
	fields, ok := value.(map[string]any)
	if !ok {
		return "", false
	}
	values, ok := fields["values"].([]any)
	if !ok || len(values) == 0 {
		return "", false
	}
	names := make([]string, 0, len(values))
	for _, item := range values {
		if name, ok := item.(string); ok && name != "" {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return "", false
	}
	return strings.Join(names, ", "), true
}

func SourceDimensionText(val any) (string, bool) {
	m, ok := val.(map[string]any)
	if !ok {
		return "", false
	}
	unit, ok := m["unit"].(string)
	if !ok || (unit != "DIPS" && unit != "POINTS" && unit != "PIXELS") {
		return "", false
	}
	value := 0.0
	if raw, exists := m["value"]; exists {
		var valid bool
		value, valid = NumberValue(raw)
		if !valid {
			return "", false
		}
	}
	return FormatNumber(value) + DisplayUnit(unit), true
}

func ShapeValueText(fields map[string]any) string {
	family, _ := fields["family"].(string)
	if family == "SHAPE_FAMILY_CIRCULAR" || strings.EqualFold(family, "circular") {
		return "Full"
	}
	base, hasDefault := SourceDimensionText(fields["defaultSize"])
	keys := []string{"topLeft", "topRight", "bottomRight", "bottomLeft"}
	labels := []string{"Top-left", "Top-right", "Bottom-right", "Bottom-left"}
	var effective [4]string
	var specified [4]bool
	for i, key := range keys {
		effective[i], specified[i] = SourceDimensionText(fields[key])
		if !specified[i] && hasDefault {
			effective[i] = base
		}
	}
	uniform := effective[0] != ""
	for _, value := range effective[1:] {
		uniform = uniform && value == effective[0]
	}
	var parts []string
	if uniform {
		parts = append(parts, effective[0])
	} else {
		if hasDefault {
			parts = append(parts, "Default: "+base)
		}
		for i, value := range effective {
			if specified[i] && (!hasDefault || value != base) {
				parts = append(parts, labels[i]+": "+value)
			}
		}
	}
	if family != "" && family != "SHAPE_FAMILY_ROUNDED_CORNERS" {
		parts = append([]string{DisplayWords(strings.ToLower(strings.TrimPrefix(family, "SHAPE_FAMILY_")))}, parts...)
	}
	if len(parts) == 0 {
		return "Shape not specified"
	}
	return strings.Join(parts, " · ")
}

func DisplayUnit(unit string) string {
	switch strings.TrimSpace(unit) {
	case "DIPS":
		return "dp"
	case "POINTS":
		return "pt"
	case "PIXELS":
		return "px"
	default:
		return strings.TrimSpace(unit)
	}
}

func TidyColorChannel(value float64) float64 {
	nearest := math.Round(value)
	if math.Abs(value-nearest) < 0.001 {
		return nearest
	}
	return value
}

func NumberValue(value any) (float64, bool) {
	number, ok := value.(float64)
	return number, ok
}

func FormatNumber(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func ChangeTypeLabel(changeType string) string {
	switch changeType {
	case "added":
		return "Added"
	case "removed":
		return "Removed"
	case "changed":
		return "Changed"
	default:
		return changeType
	}
}

func ChangeSummaryValue(value string) string {
	if value == "" {
		return "—"
	}
	return TokenValueText(json.RawMessage(value))
}

func ChangesPageURL(base string, page int) string {
	address, err := url.Parse(base)
	if err != nil {
		return base
	}
	query := address.Query()
	query.Del("page")
	if page > 1 {
		query.Set("page", strconv.Itoa(page))
	}
	address.RawQuery = query.Encode()
	return address.String()
}
