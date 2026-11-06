// Download forms share selection rules across enhanced dialogs and full-page HTML fallbacks.
package server

import (
	"context"
	"fmt"
	"github.com/sojupie/tokens/catalog/web/internal/playground"
	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	webrender "github.com/sojupie/tokens/catalog/web/internal/render"
	"html"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

func (server *Server) exportDialogOptions(query url.Values) webview.ExportDialogState {
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
	state := webview.ExportDialogState{Query: query, Label: "Download"}
	if err != nil {
		state.Error = err.Error()
		return state
	}
	state.Page = server.exportPage(subject, query)
	if state.Implementation() {
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
	if !state.Implementation() && !state.Page.CSS {
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

func (server *Server) exportDialog(ctx context.Context, query url.Values) webview.ExportDialogState {
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
	if err := ctx.Err(); err != nil {
		return "", "Download", err
	}
	if query.Get("content") == "implementation" {
		_, _, subject, err := server.selectedExport(query)
		if err != nil {
			return "", "Download", err
		}
		if query.Get("scope") == "subject" && (subject == nil || playground.ComponentFamily(subject.Key.Slug) == "") {
			return "", "Download", fmt.Errorf("no implemented component examples in this selection")
		}
		if len(jobs) > 1 {
			return fmt.Sprintf("%d context combinations in one implementation ZIP.", len(jobs)), "Download ZIP", nil
		}
		return "Component implementation ZIP, including required assets.", "Download ZIP", nil
	}
	warning := server.exportOmissionWarning(jobs)
	if len(jobs) > 1 {
		return fmt.Sprintf("%d format/context combinations in one ZIP, including conversion reports.", len(jobs)) + warning, "Download ZIP", nil
	}
	names, _, _, err := server.selectedExport(query)
	if err != nil {
		return "", "Download", err
	}
	return fmt.Sprintf("%d selected tokens. Required references are included on export.", len(names)) + warning, "Download", nil
}

func (server *Server) handleExportDialog(w http.ResponseWriter, r *http.Request) {
	state := server.exportDialog(r.Context(), r.URL.Query())
	if err := server.renderHTML(w, r, webrender.ExportDialogOptionsResponse(state)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (server *Server) handleExportDialogSummary(w http.ResponseWriter, r *http.Request) {
	state := server.exportDialog(r.Context(), r.URL.Query())
	if err := server.renderHTML(w, r, webrender.ExportDialogSummaryResponse(state)); err != nil {
		http.Error(w, "Could not render download summary.", http.StatusInternalServerError)
	}
}

// Read precomputed value metadata; editing choices never builds an export.
func (server *Server) exportOmissionWarning(jobs []url.Values) string {
	seen := map[string]bool{}
	missing, unsupported := 0, 0
	for _, job := range jobs {
		if job.Get("format") != "css" && job.Get("format") != "dtcg" {
			continue
		}
		key := cloneExportQuery(job)
		key.Del("format")
		if seen[key.Encode()] {
			continue
		}
		seen[key.Encode()] = true
		names, selection, _, err := server.selectedExport(job)
		if err != nil {
			continue
		}
		for name := range names {
			best := -1
			convertible := false
			for _, value := range server.exportIndex().tokens[name] {
				if !value.defined {
					continue
				}
				compatible := true
				for dimension, choice := range value.labels {
					if selection[dimension] != choice {
						compatible = false
						break
					}
				}
				if compatible && len(value.labels) > best {
					best = len(value.labels)
					convertible = value.convertible
				}
			}
			if best < 0 {
				missing++
			} else if !convertible {
				unsupported++
			}
		}
	}
	var reasons []string
	if missing > 0 {
		noun, verb := "token values", "have"
		if missing == 1 {
			noun, verb = "token value", "has"
		}
		reasons = append(reasons, fmt.Sprintf("%d %s %s no defined value in the selected contexts", missing, noun, verb))
	}
	if unsupported > 0 {
		noun := "token values"
		if unsupported == 1 {
			noun = "token value"
		}
		reasons = append(reasons, fmt.Sprintf("%d %s cannot be converted to CSS or DTCG", unsupported, noun))
	}
	if len(reasons) == 0 {
		return ""
	}
	return " Omitted: " + strings.Join(reasons, "; ") + ". Choose another context or Source JSON to keep captured data."
}

func (server *Server) handleDownloadPage(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if !query.Has("content") {
		_, _, subject, _ := server.selectedExport(query)
		query = webview.InitialExportDialog(server.exportPage(subject, query)).Query
	}
	state := server.exportDialog(r.Context(), query)
	if err := server.renderHTML(w, r, webrender.DownloadPage(state)); err != nil {
		http.Error(w, "Could not render download options.", http.StatusInternalServerError)
	}
}

func downloadRadio(value string, checked bool) string {
	flag := ""
	if checked {
		flag = " checked"
	}
	return `<span class="md-radio"><input type="radio" name="scope" value="` + html.EscapeString(value) + `"` + flag + `/><span class="md-radio-icon" aria-hidden="true"></span></span>`
}
