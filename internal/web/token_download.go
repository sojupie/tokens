package web

import (
	"html"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/sojupie/m3-source/internal/publish"
)

func writeTokenDownloadMenu(writer io.Writer, subject publish.SubjectView, browser tokenBrowser) error {
	menu := `<details class="sl-token-download"><summary class="sl-link">Download JSON</summary><div class="sl-download-options">`
	if len(browser.Sets) > 0 {
		menu += `<a class="sl-link" href="` + html.EscapeString(tokenDownloadURL(subject, browser, "view")) + `">Current table — selected context</a><p>Matches the selected token set and search results. Includes tokens in collapsed groups.</p><a class="sl-link" href="` + html.EscapeString(tokenDownloadURL(subject, browser, "set")) + `">Selected token set — all contexts</a>`
	}
	menu += `<a class="sl-link" href="` + html.EscapeString(subjectDownloadURL(subject)) + `" download>Full ` + strings.ToLower(html.EscapeString(subjectKindLabel(subject.Key.Kind))) + ` — all token sets and contexts</a></div></details>`
	_, err := io.WriteString(writer, menu)
	return err
}

// Filtered downloads use the same token set, search, and value selection as the table.
func (server *Server) handleTokenDownload(response http.ResponseWriter, request *http.Request, relativePath string) {
	scope := request.URL.Query().Get("scope")
	if scope != "view" && scope != "set" {
		http.Error(response, "download scope must be view or set", http.StatusBadRequest)
		return
	}
	for key, subject := range server.state.Subjects {
		if relativePath != string(key.Kind)+"/"+key.Slug+".json" {
			continue
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

func tokenDownloadURL(subject publish.SubjectView, browser tokenBrowser, scope string) string {
	query := url.Values{"scope": {scope}, "set": {browser.Selected.Key}}
	if scope == "view" {
		query.Set("q", browser.Query)
		for _, filter := range browser.Filters {
			query.Set(filter.Key, filter.Selected)
		}
	}
	return subjectDownloadURL(subject) + "?" + query.Encode()
}
