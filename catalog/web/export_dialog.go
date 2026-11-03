package web

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/a-h/templ"
	"github.com/sojupie/tokens/forge/components/ui"
)

// The dialog is an HTML client of the same selection/export rules as the public API.
// Its form still submits a native download; only editing choices uses htmx.
type exportDialogState struct {
	Page                 exportPage
	Query                url.Values
	Allowed              map[string][]string
	Counts, Error, Label string
	Ready, Resolved      bool
}

func initialExportDialog(page exportPage) exportDialogState {
	query := url.Values{"kind": {page.Kind}, "slug": {page.Slug}, "set": {page.Set}, "q": {page.Query}, "scope": {page.Scope}, "content": {"tokens"}, "format": {"dtcg"}}
	for _, control := range page.Contexts {
		query.Set("context."+control.Name, control.Selected)
	}
	return exportDialogState{Page: page, Query: query, Label: "Download"}
}

func (state exportDialogState) implementation() bool {
	return state.Query.Get("content") == "implementation"
}
func (state exportDialogState) choices(name string) []string {
	if state.Allowed == nil {
		return nil
	}
	return state.Allowed[name]
}
func (state exportDialogState) choice(name string, choice catalogChoice) ui.ChoiceProps {
	disabled := false
	switch name {
	case "content":
		disabled = choice.Value == "implementation" && !state.Page.Implementations
	case "scope":
		disabled = state.implementation() && (choice.Value == "view" || choice.Value == "set" || choice.Value == "subject" && !state.Page.Component)
	case "format":
		disabled = state.implementation() || choice.Value == "css" && !state.Page.CSS
	default:
		if state.Allowed != nil {
			disabled = !slices.Contains(state.Allowed[name], choice.Value)
		}
	}
	return ui.ChoiceProps{ID: exportChoiceID(name) + "-" + url.QueryEscape(choice.Value), Name: name, Value: choice.Value, Label: choice.Label, Checked: slices.Contains(state.Query[name], choice.Value), Disabled: disabled}
}
func (state exportDialogState) contextSummary() string {
	var labels []string
	for _, control := range state.Page.Contexts {
		if state.Allowed != nil && len(state.choices("context."+control.Name)) == 0 {
			continue
		}
		if selected := state.Query["context."+control.Name]; len(selected) > 0 {
			labels = append(labels, control.Name+": "+strings.Join(selected, ", "))
		}
	}
	return strings.Join(labels, " · ")
}

func (state exportDialogState) reportURL() string {
	query := cloneExportQuery(state.Query)
	query.Set("format", "report")
	return "/downloads/export?" + query.Encode()
}

func (server *Server) exportDialogOptions(query url.Values) exportDialogState {
	query = cloneExportQuery(query)
	for key, values := range query {
		if strings.HasPrefix(key, "_context.") {
			name := strings.TrimPrefix(key, "_")
			if len(query[name]) == 0 {
				query[name] = append([]string(nil), values...)
			}
			query.Del(key)
		}
	}
	if query.Get("content") == "implementation" {
		if query.Get("format") != "zip" {
			query["_token-format"] = append([]string(nil), query["format"]...)
		}
	} else if query.Get("format") == "zip" {
		query["format"] = append([]string(nil), query["_token-format"]...)
		query.Del("_token-format")
	}
	_, _, subject, err := server.selectedExport(query)
	state := exportDialogState{Query: query, Label: "Download"}
	if err != nil {
		state.Error = err.Error()
		return state
	}
	state.Page = server.exportPage(subject, query)
	if state.implementation() {
		if !state.Page.Implementations {
			state.Error = "Component implementations are unavailable."
			return state
		}
		if !state.Page.Component {
			query.Set("scope", "collection")
		} else if query.Get("scope") != "subject" && query.Get("scope") != "collection" {
			query.Set("scope", "subject")
		}
		query.Set("format", "zip")
		state.Allowed = map[string][]string{}
		for _, control := range state.Page.Contexts {
			if control.Name == "Theme" || control.Name == "Contrast" {
				for _, choice := range control.Choices {
					state.Allowed["context."+control.Name] = append(state.Allowed["context."+control.Name], choice.Value)
				}
			}
		}
	} else {
		if content := query.Get("content"); content != "tokens" && content != "" {
			state.Error = "Choose design tokens or component implementation."
			return state
		}
		choices, e := server.exportContextChoices(query)
		if e != nil {
			state.Error = e.Error()
			return state
		}
		state.Allowed = map[string][]string{}
		for name, values := range choices {
			state.Allowed["context."+name] = values
		}
	}
	if !state.implementation() && !state.Page.CSS {
		query["format"] = slices.DeleteFunc(query["format"], func(format string) bool { return format == "css" })
	}
	missing := len(query["format"]) == 0
	for _, control := range state.Page.Contexts {
		key := "context." + control.Name
		allowed := state.Allowed[key]
		selected := query[key]
		if len(allowed) == 0 {
			continue
		}
		valid := []string{}
		for _, value := range selected {
			if slices.Contains(allowed, value) {
				valid = append(valid, value)
			}
		}
		// Preserve an intentional empty multiselection. Fall back only when a previous
		// choice ceased to apply (for example after changing platform).
		if len(selected) > 0 && len(valid) == 0 {
			valid = []string{allowed[0]}
		}
		query[key] = valid
		if len(valid) == 0 {
			missing = true
		}
	}
	if missing {
		state.Counts = "Select at least one file format and one value for each context."
		return state
	}
	state.Resolved = true
	return state
}

func (server *Server) exportDialog(ctx context.Context, query url.Values) exportDialogState {
	state := server.exportDialogOptions(query)
	if !state.Resolved {
		return state
	}
	var err error
	state.Counts, state.Label, err = server.exportDialogSummary(ctx, state.Query)
	if err != nil {
		state.Error = err.Error()
		return state
	}
	state.Ready = true
	return state
}

func (server *Server) exportDialogSummary(ctx context.Context, query url.Values) (string, string, error) {
	query = cloneExportQuery(query)
	for key := range query {
		if strings.HasPrefix(key, "_") || query.Get("content") == "implementation" && strings.HasPrefix(key, "context.") && key != "context.Theme" && key != "context.Contrast" {
			query.Del(key)
		}
	}
	if query.Get("content") != "implementation" {
		var err error
		query, err = server.activeExportQuery(query)
		if err != nil {
			return "", "Download", err
		}
	}
	jobs, err := exportSelections(query)
	if err != nil {
		return "", "Download", err
	}
	files := 0
	var summary exportSummary
	for _, job := range jobs {
		if err := ctx.Err(); err != nil {
			return "", "Download", err
		}
		if query.Get("content") == "implementation" {
			entries, _, e := server.implementationFiles(ctx, job)
			err = e
			files += len(entries)
		} else {
			_, _, summary, _, err = server.exportDocument(job)
			files++
		}
		if err != nil {
			return "", "Download", err
		}
	}
	if len(jobs) > 1 {
		return fmt.Sprintf("%d format/context combinations · %d files in one ZIP.", len(jobs), files), "Download ZIP", nil
	}
	if query.Get("content") == "implementation" {
		return fmt.Sprintf("%d files in the implementation archive.", files), "Download ZIP", nil
	}
	return fmt.Sprintf("%d selected tokens · %d exported · %d required references · %d omitted", summary.Selected, summary.Exported, summary.Dependencies, summary.Omitted), "Download", nil
}

func (server *Server) handleExportDialog(w http.ResponseWriter, r *http.Request) {
	state := server.exportDialogOptions(r.URL.Query())
	if err := server.renderHTML(w, r, exportDialogOptionsResponse(state)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (server *Server) handleExportDialogSummary(w http.ResponseWriter, r *http.Request) {
	state := server.exportDialog(r.Context(), r.URL.Query())
	if err := server.renderHTML(w, r, exportDialogSummaryResponse(state)); err != nil {
		http.Error(w, "Could not render download summary.", http.StatusInternalServerError)
	}
}

// Request targets follow the same context order used to resolve allowed choices.
// Format changes affect only the summary; structural choices refresh dependent options.
func (state exportDialogState) requestAttrs(name string) templ.Attributes {
	endpoint := "/downloads/export/options"
	targets := []string{"#tokens-export-submit-pending"}
	dependent := name == "content" || name == "scope"
	if name == "content" {
		targets = append(targets, "#export-choices-scope")
	}
	if dependent {
		targets = append(targets, "#export-choices-format")
	}
	for _, control := range state.Page.Contexts {
		if dependent {
			targets = append(targets, "#"+exportChoiceID("context."+control.Name))
		}
		if name == "context."+control.Name {
			dependent = true
		}
	}
	if name == "format" {
		endpoint = "/downloads/export/summary"
	}
	attrs := templ.Attributes{"hx-get": endpoint, "hx-trigger": "change", "hx-include": "closest form", "hx-sync": "closest form:replace", "hx-disable": strings.Join(targets, ","), "hx-target": "#tokens-export-fields", "hx-swap": "outerMorph", "hx-indicator": "#tokens-export-submit-pending"}
	if name == "format" {
		attrs["hx-swap"] = "none"
	}
	return attrs
}

func exportChoiceID(name string) string {
	return "export-choices-" + strings.ToLower(strings.NewReplacer(".", "-", " ", "-").Replace(name))
}
