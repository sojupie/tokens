package web

import (
	"mime"
	"net/http"
	"strings"

	"github.com/sojupie/tokens/source/publish"
)

// Filtered downloads use the same token set, search, and value selection as the table.
func (server *Server) handleTokenDownload(response http.ResponseWriter, request *http.Request, relativePath string) {
	scope := request.URL.Query().Get("scope")
	if scope == "everything" {
		http.Redirect(response, request, "/downloads/tokens.json?scope=all", http.StatusSeeOther)
		return
	}
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

func (server *Server) handleCollectionDownload(response http.ResponseWriter, request *http.Request) {
	scope := request.URL.Query().Get("scope")
	if scope == "" {
		scope = "all"
	}
	if scope != "all" && scope != "components" && scope != "styles" {
		http.Error(response, "download scope must be components, styles or all", http.StatusBadRequest)
		return
	}
	subjects := []publish.PublicSubject{}
	for _, subject := range sortedSubjects(server.state) {
		component := subject.Key.Kind == publish.SubjectComponent
		if (scope == "components" && !component) || (scope == "styles" && component) {
			continue
		}
		subjects = append(subjects, subject.Public())
	}
	response.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": "tokens-" + scope + ".json"}))
	writeJSON(response, http.StatusOK, struct {
		CapturedAt string                  `json:"capturedAt"`
		Subjects   []publish.PublicSubject `json:"subjects"`
	}{server.state.CurrentSnapshot.CreatedAt, subjects})
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
