package web

import (
	"html"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/sojupie/m3-source/internal/publish"
)

func writeTokenDownloadMenu(writer io.Writer, subject publish.SubjectView, browser tokenBrowser) error {
	menu := `<div class="sl-token-download"><button type="button" data-download-open aria-haspopup="dialog">Download JSON</button><dialog aria-labelledby="sl-download-title"><form method="get" action="` + html.EscapeString(subjectDownloadURL(subject)) + `"><h2 id="sl-download-title">Download tokens</h2>`
	if browser.Selected.Label != "" {
		menu += `<p class="sl-download-table">` + html.EscapeString(browser.Selected.Label) + `</p>`
	}
	menu += `<fieldset><legend>What should the file include?</legend>`
	option := func(value, label, description string, checked bool) {
		menu += `<label class="sl-download-option"><input type="radio" name="scope" value="` + value + `"`
		if checked {
			menu += ` checked`
		}
		menu += `><span><strong>` + html.EscapeString(label) + `</strong><small>` + html.EscapeString(description) + `</small></span></label>`
	}
	if len(browser.Sets) > 0 {
		option("view", "This table, with your filters", "Matches your search and selected values.", true)
		option("set", "This table, with every value", "Includes all available settings. Ignores search and filters.", false)
	}
	allLabel := "All tokens on this page"
	if subject.Name != "" {
		allLabel = "All " + subject.Name + " tokens"
	}
	option("all", allLabel, "Includes every table and every value. Ignores search and filters.", len(browser.Sets) == 0)
	menu += `</fieldset>`
	input := func(name, value string) {
		menu += `<input type="hidden" name="` + html.EscapeString(name) + `" value="` + html.EscapeString(value) + `">`
	}
	input("set", browser.Selected.Key)
	input("q", browser.Query)
	for _, filter := range browser.Filters {
		input(filter.Key, filter.Selected)
	}
	menu += `<footer><button type="button" data-download-cancel>Cancel</button><button type="submit">Download JSON</button></footer></form></dialog><noscript><style>.sl-token-download>button{display:none}.sl-token-download dialog:not([open]){display:block;position:static}</style></noscript></div>`
	_, err := io.WriteString(writer, menu)
	return err
}

// Filtered downloads use the same token set, search, and value selection as the table.
func (server *Server) handleTokenDownload(response http.ResponseWriter, request *http.Request, relativePath string) {
	scope := request.URL.Query().Get("scope")
	if scope != "view" && scope != "set" && scope != "all" {
		http.Error(response, "download scope must be view, set or all", http.StatusBadRequest)
		return
	}
	for key, subject := range server.state.Subjects {
		if relativePath != string(key.Kind)+"/"+key.Slug+".json" {
			continue
		}
		if scope == "all" {
			response.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": strings.ReplaceAll(key.Slug, "/", "-") + ".json"}))
			writeJSON(response, http.StatusOK, subject.PublicDocument(server.state.CurrentSnapshot.CreatedAt))
			return
		}
		query := request.URL.Query()
		if scope == "set" {
			query.Del("q")
		}
		browser := buildTokenBrowser(subject, query)
		if len(browser.Sets) == 0 || (query.Get("set") != "" && query.Get("set") != browser.Selected.Key) {
			http.Error(response, "unknown token set", http.StatusBadRequest)
			return
		}
		export := tokenDownloadSubject(subject, browser, scope)
		filename := strings.ReplaceAll(subject.Key.Slug, "/", "-") + "-" + strings.ReplaceAll(browser.Selected.Key, "|", "-")
		if scope == "set" {
			filename += "-all-contexts"
		} else {
			for _, filter := range browser.Filters {
				filename += "-" + strings.ReplaceAll(strings.ToLower(filter.Selected), " ", "-")
			}
			if browser.Query != "" {
				filename += "-filtered"
			}
		}
		response.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename + ".json"}))
		writeJSON(response, http.StatusOK, export.PublicDocument(server.state.CurrentSnapshot.CreatedAt))
		return
	}
	http.NotFound(response, request)
}

func tokenDownloadSubject(subject publish.SubjectView, browser tokenBrowser, scope string) publish.SubjectView {
	visible := make(map[string]bool)
	var collect func([]tokenGroup)
	collect = func(groups []tokenGroup) {
		for _, group := range groups {
			for _, row := range group.Tokens {
				visible[row.SourceName] = true
			}
			collect(group.Children)
		}
	}
	collect(browser.Selected.Groups)
	filters := make(map[string]string)
	for _, filter := range browser.Filters {
		filters[filter.Name] = filter.Selected
	}
	export := publish.SubjectView{Key: subject.Key, Name: subject.Name, Tokens: []publish.TokenView{}}
	for _, token := range browser.Selected.sourceTokens {
		if scope == "view" {
			if !visible[token.SourceName] {
				continue
			}
			value, ok := selectTokenValue(token.Values, filters)
			token.Values = []publish.TokenValueView{}
			if ok {
				token.Values = append(token.Values, value)
			}
		}
		export.Tokens = append(export.Tokens, token)
	}
	return export
}
