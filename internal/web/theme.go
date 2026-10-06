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

	"github.com/sojupie/m3-source/internal/publish"
)

// The app's theme is independent of the token context a visitor is inspecting.
// Web maps Material's logical DIPS and type-scale POINTS to CSS pixels.
var appThemeContext = map[string]string{
	"Theme": "Light", "Contrast": "Default", "Platform": "Web", "Audience": "3P", "Font": "Static",
}

func publicationTheme(state publish.PublishedState) (string, error) {
	values := make(map[string]string)
	aliases := make(map[string]string)
	for _, subject := range sortedSubjects(state) {
		for _, token := range subject.Tokens {
			value, ok := selectTokenValue(token.Values, appThemeContext)
			if !ok {
				continue
			}
			css := themeValue(value.Value)
			if css == "" {
				continue
			}
			values[token.SourceName] = css
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
	cssValues := make(map[string]string, len(values))
	for name, value := range values {
		cssValues[strings.ReplaceAll(name, ".", "-")] = value
	}
	// Export only the tokens the styles actually use, not the full dataset.
	wanted := make(map[string]bool)
	pattern := regexp.MustCompile(`var\(--(md-[a-z0-9-]+)\)`)
	for _, file := range []string{"static/app.css", "static/token-table.css"} {
		data, err := embeddedStatic.ReadFile(file)
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
	css.WriteString("/* Served publication; Light / Default contrast / Web / 3P / Static font. */\n:root {\n")
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
	for _, key := range []string{"length", "fontSize", "lineHeight", "fontTracking"} {
		var size struct {
			Value *float64 `json:"value"`
			Unit  string   `json:"unit"`
		}
		if json.Unmarshal(value[key], &size) == nil && size.Value != nil && (size.Unit == "DIPS" || size.Unit == "POINTS") {
			return themeNumber(*size.Value) + "px"
		}
		if key == "fontTracking" && size.Value == nil && size.Unit == "POINTS" {
			// No numeric tracking was captured for some headline roles.
			// Keep native CSS tracking rather than inventing a token value.
			return "normal"
		}
	}
	var shape struct {
		Family      string `json:"family"`
		DefaultSize struct {
			Value *float64 `json:"value"`
			Unit  string   `json:"unit"`
		} `json:"defaultSize"`
	}
	if json.Unmarshal(value["shape"], &shape) == nil {
		if shape.Family == "SHAPE_FAMILY_CIRCULAR" {
			return "calc(infinity * 1px)"
		}
		if shape.Family == "SHAPE_FAMILY_ROUNDED_CORNERS" && shape.DefaultSize.Value != nil && shape.DefaultSize.Unit == "DIPS" {
			return themeNumber(*shape.DefaultSize.Value) + "px"
		}
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
	for _, key := range []string{"fontWeight", "opacity"} {
		var number *float64
		if json.Unmarshal(value[key], &number) == nil && number != nil {
			return themeNumber(*number)
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

func (server *Server) handleTheme(response http.ResponseWriter, _ *http.Request) {
	response.Header().Set("Cache-Control", "no-cache")
	if server.themeError != nil {
		http.Error(response, server.themeError.Error(), http.StatusServiceUnavailable)
		return
	}
	response.Header().Set("Content-Type", "text/css; charset=utf-8")
	fmt.Fprint(response, server.themeCSS)
}
