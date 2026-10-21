// Package cssusage reports CSS references without interpreting the browser cascade.
package cssusage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type ValueSource struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Selector string `json:"selector"`
	Property string `json:"property"`
	Value    string `json:"value"`
	Reason   string `json:"reason,omitempty"`
}

type Declaration struct {
	File             string        `json:"file"`
	Line             int           `json:"line"`
	Selector         string        `json:"selector"`
	Property         string        `json:"property"`
	Value            string        `json:"value"`
	Kind             string        `json:"kind"`
	Tokens           []string      `json:"tokens,omitempty"`
	Families         []string      `json:"families"`
	ManualValues     []ValueSource `json:"manualValues,omitempty"`
	TokenValues      []ValueSource `json:"tokenValues,omitempty"`
	UnknownVariables []string      `json:"unknownVariables,omitempty"`
}
type Override struct {
	Token     string            `json:"token"`
	Value     json.RawMessage   `json:"value"`
	Upstream  map[string]string `json:"upstream"`
	Effective map[string]string `json:"effective"`
}
type Report struct {
	Declarations []Declaration `json:"declarations"`
	Overrides    []Override    `json:"overrides"`
}

var comments = regexp.MustCompile(`(?s)/\*.*?\*/`)
var blocks = regexp.MustCompile(`(?s)([^{}]+)\{([^{}]*)\}`)
var variables = regexp.MustCompile(`var\(\s*(--[\w-]+)`)
var fallback = regexp.MustCompile(`var\(\s*--[\w-]+\s*,\s*([^()]*)\)`)

func families(selector string) []string {
	var result []string
	for _, item := range []struct{ class, family string }{{".md-button", "button"}, {".md-card", "card"}, {".md-interactive-card", "card"}, {".md-icon-button", "icon"}, {".md-filter-chip", "chip"}, {".md-radio", "radio"}, {".md-outlined-text-field", "field"}, {".md-outlined-select", "select"}, {".md-menu", "menu"}, {".md-dialog", "dialog"}, {".md-divider", "divider"}, {".md-carousel", "carousel"}, {".md-search", "search"}, {".md-date", "date"}, {".md-time", "time"}, {".md-list", "list"}} {
		if strings.Contains(selector, item.class) && !contains(result, item.family) {
			result = append(result, item.family)
		}
	}
	if len(result) == 0 {
		result = []string{"site"}
	}
	return result
}
func contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

func declarations(file, css string) []Declaration {
	css = comments.ReplaceAllStringFunc(css, func(s string) string {
		return strings.Map(func(r rune) rune {
			if r == '\n' {
				return r
			}
			return ' '
		}, s)
	})
	var rows []Declaration
	for _, block := range blocks.FindAllStringSubmatchIndex(css, -1) {
		selector := strings.TrimSpace(css[block[2]:block[3]])
		body := css[block[4]:block[5]]
		start, depth := 0, 0
		var quote rune
		escaped := false
		// Semicolons in strings or functions (including data URLs) are not declaration separators.
		for i, r := range body + ";" {
			if escaped {
				escaped = false
				continue
			}
			if r == '\\' {
				escaped = true
				continue
			}
			if quote != 0 {
				if r == quote {
					quote = 0
				}
				continue
			}
			if r == '\'' || r == '"' {
				quote = r
				continue
			}
			if r == '(' {
				depth++
			}
			if r == ')' {
				depth--
			}
			if r != ';' || depth != 0 {
				continue
			}
			text := body[start:min(i, len(body))]
			if property, value, ok := strings.Cut(strings.TrimSpace(text), ":"); ok {
				offset := block[4] + start + len(text) - len(strings.TrimLeft(text, " \n\t\r"))
				rows = append(rows, Declaration{File: file, Line: strings.Count(css[:offset], "\n") + 1, Selector: selector, Property: strings.TrimSpace(property), Value: strings.TrimSpace(value), Families: families(selector)})
			}
			start = i + 1
		}
	}
	return rows
}

// Inspect follows every possible custom-property definition. Scope-dependent or
// literal defaults are reported as mixed, not treated as a proven runtime value.
func Inspect(files map[string]string, bindingsCSS, themeCSS string) Report {
	var rows []Declaration
	var names []string
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		rows = append(rows, declarations(name, files[name])...)
	}
	definitions := map[string][]ValueSource{}
	definitionSeen := map[string]bool{}
	for _, row := range append(append([]Declaration{}, rows...), declarations("forge/components/static/components.css", bindingsCSS)...) {
		if strings.HasPrefix(row.Property, "--") {
			key := row.Selector + "\x00" + row.Property + "\x00" + row.Value
			if definitionSeen[key] {
				continue
			}
			definitionSeen[key] = true
			definitions[row.Property] = append(definitions[row.Property], ValueSource{File: row.File, Line: row.Line, Selector: row.Selector, Property: row.Property, Value: row.Value})
		}
	}
	source := map[string][]ValueSource{}
	for _, row := range declarations("theme.css", themeCSS) {
		if strings.HasPrefix(row.Property, "--md-") {
			values := source[row.Property]
			merged := false
			for j := range values {
				if values[j].Value == row.Value {
					values[j].Selector += ", " + row.Selector
					merged = true
					break
				}
			}
			if !merged {
				values = append(values, ValueSource{File: row.File, Line: row.Line, Selector: row.Selector, Property: row.Property, Value: row.Value})
			}
			source[row.Property] = values
		}
	}
	for i := range rows {
		roots := map[string]bool{}
		manual, unknown := false, false
		manualSources := map[ValueSource]bool{}
		var trace func(ValueSource, map[string]bool)
		trace = func(origin ValueSource, seen map[string]bool) {
			value := origin.Value
			refs := variables.FindAllStringSubmatch(value, -1)
			if len(refs) == 0 {
				manual = true
				manualSources[origin] = true
				return
			}
			for _, match := range fallback.FindAllStringSubmatch(value, -1) {
				if strings.TrimSpace(match[1]) != "" {
					manual = true
					leaf := origin
					leaf.Value = strings.TrimSpace(match[1])
					leaf.Reason = "fallback"
					manualSources[leaf] = true
				}
			}
			for _, ref := range refs {
				name := ref[1]
				if len(source[name]) > 0 {
					roots[name] = true
					continue
				}
				if seen[name] || len(definitions[name]) == 0 {
					unknown = true
					if !contains(rows[i].UnknownVariables, name) {
						rows[i].UnknownVariables = append(rows[i].UnknownVariables, name)
					}
					continue
				}
				seen[name] = true
				for _, expression := range definitions[name] {
					trace(expression, seen)
				}
				delete(seen, name)
			}
		}
		trace(ValueSource{File: rows[i].File, Line: rows[i].Line, Selector: rows[i].Selector, Property: rows[i].Property, Value: rows[i].Value}, map[string]bool{})
		for origin := range manualSources {
			rows[i].ManualValues = append(rows[i].ManualValues, origin)
		}
		sort.Slice(rows[i].ManualValues, func(a, b int) bool {
			x, y := rows[i].ManualValues[a], rows[i].ManualValues[b]
			if x.File != y.File {
				return x.File < y.File
			}
			if x.Line != y.Line {
				return x.Line < y.Line
			}
			return x.Value < y.Value
		})
		switch {
		case unknown:
			rows[i].Kind = "Unresolved"
		case len(roots) > 0 && manual:
			rows[i].Kind = "Mixed"
		case len(roots) > 0:
			rows[i].Kind = "Token"
		default:
			rows[i].Kind = "Manual"
		}
		for name := range roots {
			rows[i].Tokens = append(rows[i].Tokens, name)
		}
		sort.Strings(rows[i].Tokens)
		for _, name := range rows[i].Tokens {
			rows[i].TokenValues = append(rows[i].TokenValues, source[name]...)
		}
	}
	return Report{Declarations: rows, Overrides: []Override{}}
}

func (report Report) ForFamily(family string) Report {
	filtered := Report{Overrides: report.Overrides}
	for _, row := range report.Declarations {
		if family == "all" || contains(row.Families, family) {
			filtered.Declarations = append(filtered.Declarations, row)
		}
	}
	return filtered
}
func (report Report) Count(kind string) int {
	n := 0
	for _, row := range report.Declarations {
		if row.Kind == kind {
			n++
		}
	}
	return n
}

func Run(assets string) (Report, error) {
	files := map[string]string{}
	manifest, err := os.ReadFile("forge/components/style-sources.json")
	if err != nil {
		return Report{}, err
	}
	var styles []string
	if err := json.Unmarshal(manifest, &styles); err != nil {
		return Report{}, err
	}
	sources := []string{"forge/components/static/elevation.css", "catalog/web/static/app.css", "catalog/web/static/token-table.css"}
	for _, style := range styles {
		sources = append(sources, filepath.Join("forge/components", style))
	}
	for _, file := range sources {
		data, err := os.ReadFile(file)
		if err != nil {
			return Report{}, err
		}
		files[file] = string(data)
	}
	readAsset := func(name string) (string, error) {
		file := filepath.Join(assets, name)
		if assets == "" {
			file = filepath.Join("forge/components/static", name)
		}
		data, err := os.ReadFile(file)
		return string(data), err
	}
	buttons, err := readAsset("buttons.css")
	if err != nil {
		return Report{}, err
	}
	files["forge/components/static/buttons.css"] = buttons
	bindings, err := readAsset("components.css")
	if err != nil {
		return Report{}, err
	}
	theme := ""
	if assets != "" {
		theme, err = readAsset("theme.css")
		if err != nil {
			return Report{}, err
		}
	} else {
		// Root variable identities are recoverable from generated bindings without a publication.
		for _, ref := range variables.FindAllStringSubmatch(buttons+bindings, -1) {
			if strings.HasPrefix(ref[1], "--md-") {
				theme += fmt.Sprintf(":root{%s:initial;}\n", ref[1])
			}
		}
	}
	report := Inspect(files, bindings, theme)
	if assets != "" {
		data, err := os.ReadFile(filepath.Join(assets, "overrides.json"))
		if err != nil {
			return report, err
		}
		if err = json.Unmarshal(data, &report.Overrides); err != nil {
			return report, err
		}
	}
	return report, nil
}
