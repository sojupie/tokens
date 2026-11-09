// Download selection uses an immutable index to validate option changes before any expensive export work.
package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sojupie/tokens/catalog/web/internal/playground"
	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	"github.com/sojupie/tokens/source/dtcg"
	"github.com/sojupie/tokens/source/publish"
	"maps"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

type exportSummary struct {
	Selected     int               `json:"selected"`
	Exported     int               `json:"exported"`
	Dependencies int               `json:"dependencies"`
	Omitted      int               `json:"omitted"`
	Context      map[string]string `json:"context"`
}

func (server *Server) exportPage(subject *publish.SubjectView, query url.Values) webview.ExportPage {
	page := webview.ExportPage{Scope: "collection", CSS: server.cssValues != nil, Implementations: len(server.themeBundles) > 0}
	selection := dtcg.DefaultContext()
	if subject != nil {
		page.Kind, page.Slug, page.Scope = string(subject.Key.Kind), subject.Key.Slug, "subject"
		page.Set, page.Query = query.Get("set"), query.Get("q")
		if page.Set != "" || page.Query != "" {
			page.Scope = "view"
		}
		browser := buildTokenBrowser(*subject, query)
		if page.Set == "" {
			page.Set = browser.Selected.Key
		}
		for _, filter := range browser.Filters {
			selection[filter.Name] = filter.Selected
		}
		page.Component = subject.Key.Kind == publish.SubjectComponent && playground.ComponentFamily(subject.Key.Slug) != "" && server.tokenAssetsDir != ""
	}
	choices := server.exportIndex().choices
	names := make([]string, 0, len(choices))
	for name := range choices {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {

		if webview.ContextOrder(names[i]) != webview.ContextOrder(names[j]) {
			return webview.ContextOrder(names[i]) < webview.ContextOrder(names[j])
		}
		return names[i] < names[j]
	})
	for _, name := range names {
		values := make([]string, 0, len(choices[name]))
		for value := range choices[name] {
			values = append(values, value)
		}
		sort.Strings(values)
		selected := selection[name]
		if q := query.Get("context." + name); q != "" {
			selected = q
		}
		if !choices[name][selected] {
			selected = values[0]
		}
		control := webview.ExportContext{Name: name, ID: strings.ReplaceAll(strings.ToLower(name), " ", "-"), Selected: selected}
		for _, value := range values {
			control.Choices = append(control.Choices, webview.CatalogChoice{Value: value, Label: value})
		}
		page.Contexts = append(page.Contexts, control)
	}
	return page
}

func (server *Server) selectedExport(query url.Values) (map[string]bool, map[string]string, *publish.SubjectView, error) {
	context := dtcg.DefaultContext()
	for key, values := range query {
		if strings.HasPrefix(key, "context.") && len(values) > 0 {
			context[strings.TrimPrefix(key, "context.")] = values[0]
		}
	}
	var subject *publish.SubjectView
	if query.Get("slug") != "" {
		item, ok := server.state.Subjects[publish.SubjectKey{Kind: publish.SubjectKind(query.Get("kind")), Slug: query.Get("slug")}]
		if !ok {
			return nil, nil, nil, fmt.Errorf("unknown component or style")
		}
		subject = &item
	}
	names := map[string]bool{}
	add := func(tokens []publish.TokenView) {
		for _, token := range tokens {
			names[token.SourceName] = true
		}
	}
	switch query.Get("scope") {
	case "collection":
		names = maps.Clone(server.exportIndex().names)
	case "subject":
		if subject == nil {
			return nil, nil, nil, fmt.Errorf("choose a component or style")
		}
		add(subject.Tokens)
	case "view", "set":
		if subject == nil {
			return nil, nil, nil, fmt.Errorf("choose a token table")
		}
		filters := url.Values{}
		for k, v := range query {
			filters[k] = append([]string{}, v...)
		}
		if query.Get("scope") == "set" {
			filters.Del("q")
		}
		browser := buildTokenBrowser(*subject, filters)
		if len(browser.Sets) == 0 || query.Get("set") != "" && query.Get("set") != browser.Selected.Key {
			return nil, nil, nil, fmt.Errorf("unknown token table")
		}
		selected := tokenDownloadSubject(*subject, browser, query.Get("scope"))
		add(selected.Tokens)
	default:
		return nil, nil, nil, fmt.Errorf("scope must be view, set, subject or collection")
	}
	return names, context, subject, nil
}

func (server *Server) exportDocument(query url.Values) (dtcg.Node, dtcg.Report, exportSummary, *publish.SubjectView, error) {
	names, context, subject, err := server.selectedExport(query)
	if err != nil {
		return nil, dtcg.Report{}, exportSummary{}, nil, err
	}
	document, report, err := dtcg.Export(server.state, context)
	if err != nil {
		return nil, report, exportSummary{}, nil, err
	}
	document, roots, dependencies := dtcg.SelectNames(document, names)
	omitted := make([]dtcg.Omission, 0)
	for _, item := range report.Omitted {
		if names[item.Name] {
			omitted = append(omitted, item)
		}
	}
	report.Omitted = omitted
	report.Exported = dtcg.TokenCount(document)
	report.MappedNames = roots + dependencies
	report.Aliases = 0
	report.ExtensionOnly = nil
	for name, node := range dtcg.NamedNodes(document) {
		if alias, ok := node["$value"].(string); ok && strings.HasPrefix(alias, "{") {
			report.Aliases++
		}
		if _, ok := node["$value"]; !ok {
			report.ExtensionOnly = append(report.ExtensionOnly, name)
		}
	}
	sort.Strings(report.ExtensionOnly)
	summary := exportSummary{Selected: len(names), Exported: roots, Dependencies: dependencies, Omitted: len(omitted), Context: context}

	metadata := document["$extensions"].(map[string]any)["dev.tokencatalog.export"].(map[string]any)
	metadata["conversionReport"] = report
	metadata["selection"] = summary
	return document, report, summary, subject, nil
}

func (server *Server) handleSelectedExport(w http.ResponseWriter, r *http.Request) {
	if token := r.URL.Query().Get("_download"); downloadSignal(token) {
		w = &exportDownloadResponse{ResponseWriter: w, token: token}
	}
	query := r.URL.Query()
	query.Del("_download")
	if query.Get("content") != "implementation" {
		if query.Get("options") == "1" {
			choices, err := server.exportContextChoices(query)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, choices)
			return
		}
		active, err := server.activeExportQuery(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		query = active
		clone := r.Clone(r.Context())
		u := *r.URL
		u.RawQuery = query.Encode()
		clone.URL = &u
		r = clone
	}
	if server.handleExportBatch(w, r, query) {
		return
	}
	if content := query.Get("content"); content != "" && content != "tokens" && content != "implementation" {
		http.Error(w, "content must be tokens or implementation", http.StatusBadRequest)
		return
	}
	if query.Get("content") == "implementation" {
		if query.Get("format") != "zip" && query.Get("format") != "summary" {
			http.Error(w, "component implementations use ZIP format", http.StatusBadRequest)
			return
		}
		if query.Get("summary") == "1" {
			query.Set("format", "summary")
			clone := r.Clone(r.Context())
			u := *r.URL
			u.RawQuery = query.Encode()
			clone.URL = &u
			r = clone
		}
		server.handleImplementationExport(w, r)
		return
	}
	if query.Get("summary") == "1" || query.Get("format") == "summary" {
		_, _, summary, _, err := server.exportDocument(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, summary)
		return
	}
	body, contentType, filename, err := server.exportTokenFile(query)
	if err != nil {
		http.Error(w, err.Error(), exportStatus(err))
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.Header().Set("Cache-Control", "public, no-cache")
	_, _ = w.Write(body)
}

func (server *Server) exportTokenFile(query url.Values) ([]byte, string, string, error) {
	document, report, summary, subject, err := server.exportDocument(query)
	if err != nil {
		return nil, "", "", err
	}
	return server.renderTokenFile(query, document, report, summary, subject)
}

func (server *Server) renderTokenFile(query url.Values, document dtcg.Node, report dtcg.Report, summary exportSummary, subject *publish.SubjectView) ([]byte, string, string, error) {
	var err error
	format := query.Get("format")
	filename := "tokens"
	if subject != nil && query.Get("scope") != "collection" {
		filename = strings.ReplaceAll(subject.Key.Slug, "/", "-")
	}
	var body []byte
	contentType := "application/json"
	switch format {
	case "dtcg":
		body, err = json.MarshalIndent(document, "", "  ")
		filename += ".tokens.json"
	case "report":
		body, err = json.MarshalIndent(struct {
			dtcg.Report
			Selection exportSummary `json:"selection"`
		}{report, summary}, "", "  ")
		filename += "-report.json"
	case "css":
		body, err = server.exportCSS(document)
		filename += ".css"
		contentType = "text/css; charset=utf-8"
	case "source":
		names, context, _, _ := server.selectedExport(query)
		subjects := []publish.PublicSubject{}
		var candidates []publish.SubjectView
		if query.Get("scope") == "collection" {
			candidates = sortedSubjects(server.state)
		} else {
			candidates = []publish.SubjectView{*subject}
		}
		for _, item := range candidates {
			selected := publish.SubjectView{Key: item.Key, Name: item.Name, Tokens: []publish.TokenView{}}
			for _, token := range item.Tokens {
				if !names[token.SourceName] {
					continue
				}
				value, ok, e := dtcg.Select(token.Values, context)
				if e != nil {
					err = e
					break
				}
				token.Values = nil
				if ok {
					token.Values = append(token.Values, value)
				}
				selected.Tokens = append(selected.Tokens, token)
			}
			if len(selected.Tokens) > 0 {
				subjects = append(subjects, selected.Public())
			}
		}
		if err == nil {
			body, err = json.MarshalIndent(struct {
				Context  map[string]string       `json:"context"`
				Subjects []publish.PublicSubject `json:"subjects"`
			}{context, subjects}, "", "  ")
		}
		filename += "-source.json"
	default:
		return nil, "", "", fmt.Errorf("format must be css, dtcg, source or report")
	}
	if err != nil {
		return nil, "", "", exportRenderError{err}
	}
	return body, contentType, filename, nil
}

func (server *Server) exportCSS(document dtcg.Node) ([]byte, error) {
	if server.cssValues == nil {
		return nil, fmt.Errorf("CSS value index unavailable; rebuild token assets")
	}
	values := map[string]string{}
	for name, node := range dtcg.NamedNodes(document) {
		extensions, _ := node["$extensions"].(map[string]any)
		source, _ := extensions[dtcg.Extension].(map[string]any)
		raw, err := json.Marshal(source["sourceValue"])
		if err != nil {
			return nil, err
		}
		key := dtcg.ValueKey(raw)
		prefix := strings.ReplaceAll(name, ".", "-")
		found := false
		for suffix, value := range server.cssValues[key] {
			values[prefix+suffix] = value
			found = true
		}
		if !found {
			return nil, fmt.Errorf("missing compiled CSS value for %s; rebuild token assets", name)
		}
		if alias, ok := node["$value"].(string); ok && strings.HasPrefix(alias, "{") {
			values[prefix] = "var(--" + strings.ReplaceAll(strings.TrimSuffix(strings.TrimPrefix(alias, "{"), "}"), ".", "-") + ")"
		}
	}
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	var css strings.Builder
	css.WriteString(":root {\n")
	for _, name := range names {
		fmt.Fprintf(&css, "  --%s: %s;\n", name, values[name])
	}
	css.WriteString("}\n")
	return []byte(css.String()), nil
}

type exportRenderError struct{ error }

func exportStatus(err error) int {
	var render exportRenderError
	if errors.As(err, &render) {
		return http.StatusUnprocessableEntity
	}
	return http.StatusBadRequest
}

// Built once from the immutable publication. Request handlers never parse captured
// values or rediscover the collection's context vocabulary.
type exportIndex struct {
	tokens  map[string][]indexedExportValue
	choices map[string]map[string]bool
	names   map[string]bool
}
type indexedExportValue struct {
	labels      map[string]string
	references  []publish.TokenReference
	defined     bool
	convertible bool
}

func (server *Server) exportIndex() *exportIndex {
	server.exportIndexOnce.Do(func() {
		index := &exportIndex{tokens: map[string][]indexedExportValue{}, choices: map[string]map[string]bool{}, names: map[string]bool{}}
		converted := map[string]bool{}
		for _, subject := range server.state.Subjects {
			for _, token := range subject.Tokens {
				index.names[token.SourceName] = true
				for _, value := range token.Values {
					var object map[string]json.RawMessage
					entry := indexedExportValue{labels: map[string]string{}, references: value.References}
					entry.defined = json.Unmarshal(value.Value, &object) == nil && object["undefined"] == nil
					if entry.defined {
						key := string(value.Value)
						valid, known := converted[key]
						if !known {
							_, err := dtcg.Convert(value.Value)
							valid = err == nil
							converted[key] = valid
						}
						entry.convertible = valid
					}
					for _, label := range value.ContextLabels {
						name, choice, ok := strings.Cut(label, ":")
						if !ok {
							continue
						}
						name, choice = strings.TrimSpace(name), strings.TrimSpace(choice)
						entry.labels[name] = choice
						if index.choices[name] == nil {
							index.choices[name] = map[string]bool{}
						}
						index.choices[name][choice] = true
					}
					index.tokens[token.SourceName] = append(index.tokens[token.SourceName], entry)
				}
			}
		}
		server.exportSelectionIndex = index
	})
	return server.exportSelectionIndex
}

// Context labels describe individual captured values, not a global matrix.
// Include referenced tokens because their settings can affect the exported CSS.
func (server *Server) exportContextChoices(query url.Values) (map[string][]string, error) {
	names, defaults, _, err := server.selectedExport(query)
	if err != nil {
		return nil, err
	}
	index := server.exportIndex()
	known := index.choices
	for key, values := range query {
		if !strings.HasPrefix(key, "context.") {
			continue
		}
		name := strings.TrimPrefix(key, "context.")
		for _, value := range values {
			if options, ok := known[name]; ok {
				if !options[value] {
					return nil, fmt.Errorf("unknown %s context %q", name, value)
				}
			} else if expected, ok := dtcg.DefaultContext()[name]; !ok || expected != value {
				return nil, fmt.Errorf("unknown context %q", name)
			}
		}
	}
	queue := []string{}
	for name := range names {
		queue = append(queue, name)
	}
	for i := 0; i < len(queue); i++ {
		for _, value := range index.tokens[queue[i]] {

			compatible := true
			for key, required := range value.labels {
				selected := query["context."+key]
				if len(selected) == 0 {
					selected = []string{defaults[key]}
				}
				match := false
				for _, candidate := range selected {
					if candidate == required {
						match = true
						break
					}
				}
				if !match {
					compatible = false
					break
				}
			}
			if !compatible {
				continue
			}
			for _, reference := range value.references {
				if !names[reference.Name] {
					names[reference.Name] = true
					queue = append(queue, reference.Name)
				}
			}
		}
	}
	choices := map[string]map[string]bool{}
	for name := range names {
		for _, value := range index.tokens[name] {
			if !value.defined {
				continue
			}
			labels := value.labels
			for dimension, candidate := range labels {
				compatible := true
				for other, required := range labels {
					if other == dimension {
						continue
					}
					selected := query["context."+other]
					if len(selected) == 0 {
						selected = []string{defaults[other]}
					}
					match := false
					for _, selectedValue := range selected {
						if selectedValue == required {
							match = true
							break
						}
					}
					if !match {
						compatible = false
						break
					}
				}
				if compatible {
					if choices[dimension] == nil {
						choices[dimension] = map[string]bool{}
					}
					choices[dimension][candidate] = true
				}
			}
		}
	}
	result := map[string][]string{}
	for name, values := range choices {
		for value := range values {
			result[name] = append(result[name], value)
		}
		sort.Strings(result[name])
	}
	return result, nil
}

func (server *Server) activeExportQuery(query url.Values) (url.Values, error) {
	choices, err := server.exportContextChoices(query)
	if err != nil {
		return nil, err
	}
	query = cloneExportQuery(query)
	for key := range query {
		if strings.HasPrefix(key, "context.") && len(choices[strings.TrimPrefix(key, "context.")]) == 0 {
			query.Del(key)
		}
	}
	return query, nil
}
