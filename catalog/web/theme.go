package web

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/sojupie/tokens/source/dtcg"
	"github.com/sojupie/tokens/source/publish"
)

var themeTokenReference = regexp.MustCompile(`var\(\s*--(md-[a-z0-9-]+)\s*[,)]`)

// The app's theme is independent of the token context a visitor is inspecting.
// Web maps Material's logical DIPS and type-scale POINTS to CSS pixels.
func publicationTheme(state publish.PublishedState) (string, error) {
	var css strings.Builder
	for _, theme := range []string{"Light", "Dark"} {
		for _, contrast := range []string{"Default", "Medium contrast", "High contrast"} {
			variant, err := publicationThemeVariant(state, theme, contrast)
			if err != nil {
				return "", fmt.Errorf("%s / %s: %w", theme, contrast, err)
			}
			css.WriteString(variant)
		}
	}
	return css.String(), nil
}

func publicationThemeVariant(state publish.PublishedState, theme, contrast string) (string, error) {
	context := map[string]string{
		"Theme": theme, "Contrast": contrast, "Platform": "Web", "Audience": "3P", "Font": "Static",
	}
	values := make(map[string]string)
	aliases := make(map[string]string)
	paletteFallback := make(map[string]string)
	fallbackContext := map[string]string{"Theme": "Light", "Contrast": "Default", "Platform": "Web", "Audience": "3P", "Font": "Static"}
	for _, subject := range sortedSubjects(state) {
		for _, token := range subject.Tokens {
			// Reference tones are fixed across appearances. Preserve an explicit
			// context value, and recover absent tones from the captured default.
			if fallback, ok := selectTokenValue(token.Values, fallbackContext); ok {
				if css := themeValue(fallback.Value); css != "" {
					if strings.HasPrefix(token.SourceName, "md.ref.palette.") {
						paletteFallback[token.SourceName] = css
					}
					for refs := fallback.References; len(refs) == 1; refs = refs[0].References {
						if name := refs[0].Name; strings.HasPrefix(name, "md.ref.palette.") {
							if prior, exists := paletteFallback[name]; exists && prior != css {
								paletteFallback[name] = ""
							} else if !exists {
								paletteFallback[name] = css
							}
						}
					}
				}
			}
			value, ok := selectTokenValue(token.Values, context)
			if !ok {
				continue
			}
			css := themeValue(value.Value)
			if css == "" {
				continue
			}
			values[token.SourceName] = css
			if node, err := dtcg.Convert(value.Value); err == nil && node["$type"] == "typography" {
				typography := node["$value"].(map[string]any)
				spacing := typography["letterSpacing"].(dtcg.Dimension)
				values[token.SourceName+".letter-spacing"] = themeNumber(spacing.Value) + spacing.Unit
			}
			// A single reference chain denotes the same primitive value. This
			// recovers system roles absent as standalone values in the publication.
			// Composite/branching references cannot be inferred this way.
			for refs := value.References; len(refs) == 1; refs = refs[0].References {
				name := refs[0].Name
				if previous, exists := aliases[name]; exists && previous != css {
					aliases[name] = "" // Conflicting evidence is not a theme value.
				} else if !exists {
					aliases[name] = css
				}
			}
		}
	}
	for name, value := range aliases {
		if values[name] == "" {
			values[name] = value
		}
	}
	for name, css := range paletteFallback {
		if _, explicit := values[name]; explicit {
			continue
		}
		if _, evidence := aliases[name]; !evidence && css != "" {
			values[name] = css
		}
	}
	cssValues := make(map[string]string, len(values))
	for name, value := range values {
		cssValues[strings.ReplaceAll(name, ".", "-")] = value
	}
	// Include preview text colour inputs alongside the tokens referenced by the styles.
	wanted := map[string]bool{"md-ref-palette-neutral0": true, "md-sys-color-on-primary": true}
	pattern := themeTokenReference
	for _, file := range []string{"static/app.css", "static/token-table.css", "static/buttons.css", "static/components.css"} {
		data, err := readStaticFile(file)
		if err != nil {
			return "", err
		}
		for _, match := range pattern.FindAllStringSubmatch(string(data), -1) {
			wanted[match[1]] = true
		}
	}
	names := make([]string, 0, len(wanted))
	var missing []string
	for name := range wanted {
		names = append(names, name)
		if cssValues[name] == "" {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	if len(missing) != 0 {
		return "", fmt.Errorf("publication is missing app theme tokens: %s", strings.Join(missing, ", "))
	}
	sort.Strings(names)
	var css strings.Builder
	contrastKey := strings.ToLower(strings.TrimSuffix(contrast, " contrast"))
	if theme == "Light" && contrast == "Default" {
		css.WriteString(":root, ")
	}
	fmt.Fprintf(&css, ":root[data-theme=%q][data-contrast=%q] {\n  color-scheme: %s;\n", strings.ToLower(theme), contrastKey, strings.ToLower(theme))
	if theme == "Light" && contrast == "Default" {
		// Pin preview text colour to the publication's black/white values, rather than
		// on-surface/on-primary roles whose luminance flips with the app theme.
		fmt.Fprintf(&css, "  --tokens-preview-on-light: %s;\n  --tokens-preview-on-dark: %s;\n", cssValues["md-ref-palette-neutral0"], cssValues["md-sys-color-on-primary"])
	}
	for _, name := range names {
		fmt.Fprintf(&css, "  --%s: %s;\n", name, cssValues[name])
	}
	css.WriteString("}\n")
	return css.String(), nil
}

// Only typed primitives become CSS. Raw source strings are never interpolated.
func themeValue(raw json.RawMessage) string {
	var value map[string]json.RawMessage
	if json.Unmarshal(raw, &value) != nil {
		return ""
	}
	if _, ok := value["color"]; ok {
		return tokenColor(raw)
	}
	var elevation struct {
		Value *float64 `json:"value"`
		Unit  string   `json:"unit"`
	}
	if json.Unmarshal(value["elevation"], &elevation) == nil && elevation.Unit == "DIPS" {
		if elevation.Value == nil {
			return "0" // Protobuf omits the zero elevation value.
		}
		// Material Web's shadow renderer takes a level, rather than a dp distance.
		for level, dp := range []float64{0, 1, 3, 6, 8, 12} {
			if *elevation.Value == dp {
				return strconv.Itoa(level)
			}
		}
		return ""
	}
	for _, key := range []string{"length", "fontSize", "lineHeight", "fontTracking"} {
		var size struct {
			Value *float64 `json:"value"`
			Unit  string   `json:"unit"`
		}
		if key == "length" && json.Unmarshal(value[key], &size) == nil && size.Value != nil && size.Unit == "PERCENT" {
			if number := themeNumber(*size.Value); number != "" {
				return number + "%"
			}
		}
		if json.Unmarshal(value[key], &size) == nil && size.Value != nil && (size.Unit == "DIPS" || size.Unit == "POINTS") {
			return themeNumber(*size.Value) + "px"
		}
	}
	if _, ok := value["shape"]; ok {
		shape := dtcg.ShapeCSS(raw)
		if shape == "50%" {
			return "calc(infinity * 1px)"
		}
		return shape
	}
	var fonts struct {
		Values []string `json:"values"`
	}
	if json.Unmarshal(value["fontNames"], &fonts) == nil && len(fonts.Values) > 0 {
		quoted := make([]string, len(fonts.Values))
		for i, name := range fonts.Values {
			// CSS quoted strings: escape slashes, quotes and control characters.
			quoted[i] = `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\a `, "\r", `\d `, "\f", `\c `).Replace(name) + `"`
		}
		return strings.Join(quoted, ", ") + ", sans-serif"
	}
	for _, key := range []string{"fontWeight", "opacity", "numeric"} {
		var number *float64
		if json.Unmarshal(value[key], &number) == nil && number != nil {
			return themeNumber(*number)
		}
	}
	if node, err := dtcg.Convert(raw); err == nil {
		switch node["$type"] {
		case "dimension":
			if _, elevation := value["elevation"]; elevation {
				return ""
			}
			dimension := node["$value"].(dtcg.Dimension)
			return themeNumber(dimension.Value) + dimension.Unit
		case "duration":
			value := node["$value"].(map[string]any)
			return themeNumber(value["value"].(float64)) + value["unit"].(string)
		case "cubicBezier":
			points := node["$value"].([]float64)
			parts := make([]string, len(points))
			for i, v := range points {
				parts[i] = themeNumber(v)
			}
			return "cubic-bezier(" + strings.Join(parts, ", ") + ")"
		case "typography":
			t := node["$value"].(map[string]any)
			size := t["fontSize"].(dtcg.Dimension)
			fonts, _ := json.Marshal(map[string]any{"fontNames": map[string]any{"values": t["fontFamily"]}})
			return themeNumber(t["fontWeight"].(float64)) + " " + themeNumber(size.Value) + size.Unit + "/" + themeNumber(t["lineHeight"].(float64)) + " " + themeValue(fonts)
		}
	}
	return ""
}

func themeNumber(value float64) string {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return ""
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func (server *Server) handleTheme(response http.ResponseWriter, request *http.Request) {
	if server.themeError != nil {
		response.Header().Set("Cache-Control", "no-store")
		http.Error(response, server.themeError.Error(), http.StatusServiceUnavailable)
		return
	}
	response.Header().Set("Content-Type", "text/css; charset=utf-8")
	serveCachedAsset(response, request, "theme.css", []byte(server.themeCSS))
}
