package presentation

import (
	"encoding/json"
	"github.com/sojupie/tokens/source/dtcg"
	"strings"

	"github.com/sojupie/tokens/source/publish"
)

type TokenProperty struct {
	Label      string
	Value      string
	Icon       string
	References []publish.TokenReference
}

func TokenValueDetails(value map[string]json.RawMessage, references []publish.TokenReference) (kind string, properties []TokenProperty) {
	if len(value) != 1 {
		return "", nil
	}
	for key, raw := range value {
		if key == "shape" {
			properties = shapeProperties(raw, references)
		}
		return key, properties
	}
	return "", nil
}

func shapeProperties(raw json.RawMessage, references []publish.TokenReference) []TokenProperty {
	var fields map[string]any
	if json.Unmarshal(raw, &fields) != nil {
		return nil
	}
	var properties []TokenProperty
	if family, ok := fields["family"].(string); ok && family != "" {
		properties = append(properties, TokenProperty{
			Label: "Shape family", Value: DisplayWords(strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(family, "SHAPE_FAMILY_")), "_", " ")), Icon: "rounded_corner", References: references,
		})
	}
	for _, field := range []struct{ key, label string }{
		{"defaultSize", "Default size"},
		{"topLeft", "Top-left radius"},
		{"topRight", "Top-right radius"},
		{"bottomRight", "Bottom-right radius"},
		{"bottomLeft", "Bottom-left radius"},
	} {
		if text, ok := SourceDimensionText(fields[field.key]); ok {
			properties = append(properties, TokenProperty{Label: field.label, Value: text, Icon: "straighten", References: references})
		}
	}
	return properties
}

func scalarTokenProperty(token TokenRow, value string) TokenProperty {
	label, icon := "Value", "info"
	switch {
	case token.Color != "" || token.ValueKind == "color":
		label, icon = "Color", "palette"
	case token.Shape != "" || token.ValueKind == "shape":
		label, icon = "Shape", "rounded_corner"
	case token.Elevation != "" || token.ValueKind == "elevation":
		label, icon = "Elevation", "layers"
	default:
		switch token.ValueKind {
		case "length", "dimension":
			label, icon = "Measurement", "straighten"
		case "opacity":
			label, icon = "Opacity", "opacity"
		case "numeric":
			label, icon = "Number", "numbers"
		case "durationMs":
			label, icon = "Duration", "schedule"
		case "cubicBezier", "svgPath":
			label, icon = "Easing", "timeline"
		case "motionPath":
			label, icon = "Motion path", "timeline"
		case "motionReference":
			label, icon = "Motion reference", "timeline"
		case "fontNames":
			label, icon = "Font names", "font_download"
		case "fontSize":
			label, icon = "Font size", "format_size"
		case "fontTracking":
			label, icon = "Font tracking", "format_letter_spacing"
		case "fontWeight":
			label, icon = "Font weight", "format_bold"
		case "lineHeight":
			label, icon = "Line height", "format_line_spacing"
		case "axisValue":
			label, icon = "Variation axis", "settings_outline"
		case "textTransform":
			label, icon = "Text transform", "format_size"
		}
	}
	return TokenProperty{Label: label, Value: value, Icon: icon, References: token.References}
}

// InspectorValue formats primitives whose generic value text includes field names
// or omits source units. It is presentation-only.
func InspectorValue(kind string, raw json.RawMessage) string {
	var value any
	if json.Unmarshal(raw, &value) != nil {
		return ""
	}
	switch kind {
	case "durationMs", "numeric":
		if number, ok := NumberValue(value); ok {
			text := FormatNumber(number)
			if kind == "durationMs" {
				text += "ms"
			}
			return text
		}
	case "axisValue":
		if fields, ok := value.(map[string]any); ok {
			tag, _ := fields["tag"].(string)
			if number, ok := NumberValue(fields["value"]); ok {
				return strings.TrimSpace(tag + " " + FormatNumber(number))
			}
			if tag != "" {
				return tag + " 0"
			}
		}
	case "length", "dimension":
		text, _ := SourceDimensionText(value)
		return text
	case "cubicBezier":
		wrapper, err := json.Marshal(map[string]json.RawMessage{kind: raw})
		if err != nil {
			return ""
		}
		node, err := dtcg.Convert(wrapper)
		if err == nil {
			if coordinates, ok := node["$value"].([]float64); ok {
				parts := make([]string, 0, len(coordinates))
				for _, coordinate := range coordinates {
					parts = append(parts, FormatNumber(coordinate))
				}
				return "cubic-bezier(" + strings.Join(parts, ", ") + ")"
			}
		}
	case "motionPath":
		if fields, ok := value.(map[string]any); ok {
			if path, ok := fields["standardPath"].(string); ok {
				return DisplayWords(strings.ToLower(path))
			}
		}
	case "textTransform":
		if name, ok := value.(string); ok {
			return DisplayWords(strings.ToLower(strings.TrimPrefix(name, "TEXT_TRANSFORM_")))
		}
	case "svgPath":
		if path, ok := value.(string); ok {
			return path
		}
	}
	return ""
}
