package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/sojupie/tokens/source/dtcg"
	"github.com/sojupie/tokens/source/publish"
)

type exportContext struct {
	Name, ID, Selected string
	Choices            []catalogChoice
}
type exportPage struct {
	CSS, Implementations          bool
	Kind, Slug, Set, Query, Scope string
	Contexts                      []exportContext
	Component                     bool
}
type exportSummary struct {
	Selected     int               `json:"selected"`
	Exported     int               `json:"exported"`
	Dependencies int               `json:"dependencies"`
	Omitted      int               `json:"omitted"`
	Context      map[string]string `json:"context"`
}

func (server *Server) exportPage(subject *publish.SubjectView, query url.Values) exportPage {
	page := exportPage{Scope: "collection", CSS: server.cssValues != nil, Implementations: len(server.themeBundles) > 0}
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
		page.Component = subject.Key.Kind == publish.SubjectComponent && componentFamily(subject.Key.Slug) != "" && server.tokenAssetsDir != ""
	}
	choices := server.exportIndex().choices
	names := make([]string, 0, len(choices))
	for name := range choices {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {

		if contextOrder(names[i]) != contextOrder(names[j]) {
			return contextOrder(names[i]) < contextOrder(names[j])
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
		control := exportContext{Name: name, ID: strings.ReplaceAll(strings.ToLower(name), " ", "-"), Selected: selected}
		for _, value := range values {
			control.Choices = append(control.Choices, catalogChoice{value, value})
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
	return document, report, summary, subject, nil
}

func (server *Server) handleSelectedExport(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
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
		for _, item := range sortedSubjects(server.state) {
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

func (page exportPage) Scopes() []catalogChoice {
	choices := []catalogChoice{}
	if page.Slug != "" {
		choices = append(choices, catalogChoice{"view", "Current filtered table"}, catalogChoice{"set", "Whole table"}, catalogChoice{"subject", "This component or style"})
	}
	return append(choices, catalogChoice{"collection", "Whole collection"})
}

func (page *exportPage) defaultScope() string {
	if page == nil {
		return "collection"
	}
	return page.Scope
}

type exportRenderError struct{ error }

func exportStatus(err error) int {
	var render exportRenderError
	if errors.As(err, &render) {
		return http.StatusUnprocessableEntity
	}
	return http.StatusBadRequest
}
