package presentation

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/sojupie/tokens/source/publish"
)

type TypographyPreview struct {
	Style        string
	FallbackFont bool
	Properties   []TokenProperty
}

func TokenTypographyPreview(raw json.RawMessage, name string, references []publish.TokenReference) *TypographyPreview {
	var value struct {
		Type map[string]any `json:"type"`
	}
	if json.Unmarshal(raw, &value) != nil || len(value.Type) == 0 {
		return nil
	}
	fields := value.Type
	// A component commonly points to one type-scale token. Its direct
	// dependencies are the individual typography properties.
	var parent string
	if len(references) == 1 && strings.HasPrefix(references[0].Name, "md.sys.typescale.") && len(references[0].References) > 0 {
		parent = references[0].Name
		name, references = parent, references[0].References
	}
	byProperty := make(map[string][]publish.TokenReference)
	for _, ref := range references {
		if property, ok := strings.CutPrefix(ref.Name, name+"."); ok {
			if parent != "" {
				ref = publish.TokenReference{Name: parent, References: []publish.TokenReference{ref}}
			}
			byProperty[property] = append(byProperty[property], ref)
		}
	}

	propertyReferences := func(key string) []publish.TokenReference {
		if len(references) == 1 && parent == "" && strings.HasPrefix(references[0].Name, "md.sys.typescale.") && len(references[0].References) == 0 && !strings.HasPrefix(references[0].Name, name+".") {
			return references
		}
		return byProperty[key]
	}
	preview := &TypographyPreview{}
	for _, property := range []struct{ key, label, icon, alias string }{
		{"fontNames", "Font names", "font_download", "font"},
		{"fontSize", "Font size", "format_size", "size"},
		{"fontTracking", "Font tracking", "format_letter_spacing", "tracking"},
		{"fontWeight", "Font weight", "format_bold", "weight"},
		{"lineHeight", "Line height", "format_line_spacing", "line-height"},
	} {
		if text, ok := FormattedNamedValue(property.key, fields[property.key]); ok {
			preview.Properties = append(preview.Properties, TokenProperty{
				Label: property.label, Value: text, Icon: property.icon, References: propertyReferences(property.alias),
			})
		}
	}
	if axes, ok := fields["variationAxes"].(map[string]any); ok {
		keys := make([]string, 0, len(axes))
		for key := range axes {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			axis, _ := axes[key].(map[string]any)
			if number, ok := NumberValue(axis["axisValue"]); ok {
				preview.Properties = append(preview.Properties, TokenProperty{
					Label: "Variation axis · " + key, Value: FormatNumber(number),
					Icon: "settings_outline", References: propertyReferences(strings.ToLower(key)),
				})
			}
		}
	}

	size, ok := typographyLength(fields["fontSize"])
	if !ok || size == "0px" || strings.HasPrefix(size, "-") {
		return preview
	}

	preview.Style = "--tokens-type-size:" + size
	if lineHeight, ok := typographyLength(fields["lineHeight"]); ok {
		preview.Style += ";--tokens-type-line-height:" + lineHeight
	}
	if tracking, ok := typographyLength(fields["fontTracking"]); ok {
		preview.Style += ";--tokens-type-tracking:" + tracking
	}
	if weight, ok := NumberValue(fields["fontWeight"]); ok && weight >= 1 && weight <= 1000 {
		preview.Style += ";--tokens-type-weight:" + FormatNumber(weight)
	}
	fontNames, _ := FontNamesValue(fields["fontNames"])
	font, _, _ := strings.Cut(fontNames, ", ")
	switch font {
	case "System":
		preview.Style += ";--tokens-type-family:system-ui"
	case "Google Sans", "Google Sans Flex", "Google Sans Text", "Roboto Flex":
		preview.Style += `;--tokens-type-family:"` + font + `",Roboto,system-ui`
		preview.FallbackFont = true
	case "Roboto":
		preview.Style += ";--tokens-type-family:Roboto,system-ui"
	default:
		preview.FallbackFont = font != ""
	}
	return preview
}

func typographyLength(value any) (string, bool) {
	text, ok := SourceDimensionText(value)
	if !ok {
		return "", false
	}
	// Match the Web export policy while keeping source units in the property list.
	return text[:len(text)-2] + "px", true
}
