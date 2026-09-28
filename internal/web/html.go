package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/a-h/templ"
	"github.com/sojupie/m3-source/internal/publish"
)

func (server *Server) handleOverview(response http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" && request.URL.Path != "/components" && request.URL.Path != "/styles" {
		http.NotFound(response, request)
		return
	}
	if request.URL.Path == "/" {
		if err := renderHTML(response, request, overviewPage(server.state.CurrentSnapshot, sortedSubjects(server.state))); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	styles := request.URL.Path == "/styles"
	query := request.URL.Query().Get("q")
	subjects := filterSubjects(sortedSubjects(server.state), query, styles)
	if request.Header.Get("HX-Request") == "true" {
		response.Header().Set("Vary", "HX-Request")
		if err := renderHTML(response, request, subjectList(subjects, styles)); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if err := renderHTML(response, request, subjectListPage(server.state.CurrentSnapshot, subjects, styles, query)); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
}

func (server *Server) handleHTMLSubject(response http.ResponseWriter, request *http.Request) {
	slug := request.PathValue("slug")
	if strings.HasSuffix(slug, "/changes") {
		server.handleHTMLChanges(response, request, strings.TrimSuffix(slug, "/changes"))
		return
	}
	subjects := make([]publish.SubjectView, 0, 1)
	for _, subject := range server.state.Subjects {
		if subject.Key.Slug != slug {
			continue
		}
		if strings.HasPrefix(request.URL.Path, "/components/") && subject.Key.Kind != publish.SubjectComponent {
			continue
		}
		if strings.HasPrefix(request.URL.Path, "/styles/") && subject.Key.Kind == publish.SubjectComponent {
			continue
		}
		subjects = append(subjects, subject)
	}
	if len(subjects) == 0 {
		http.NotFound(response, request)
		return
	}
	if len(subjects) > 1 {
		http.Error(response, "subject slug is ambiguous", http.StatusConflict)
		return
	}
	if err := renderHTML(response, request, subjectPage(server.state.CurrentSnapshot, subjects[0], subjectChanges(server.state, subjects[0].Key.Slug))); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
}

func (server *Server) handleHTMLChanges(response http.ResponseWriter, request *http.Request, subject string) {
	changes := subjectChanges(server.state, subject)
	if len(changes) == 0 {
		http.NotFound(response, request)
		return
	}
	if err := renderHTML(response, request, changesPage(server.state.CurrentSnapshot, subject, changes)); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
}

func (server *Server) handleGlobalChanges(response http.ResponseWriter, request *http.Request) {
	if err := renderHTML(response, request, changesPage(server.state.CurrentSnapshot, "", server.state.Changes)); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
}

func renderHTML(response http.ResponseWriter, request *http.Request, component templ.Component) error {
	response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return component.Render(request.Context(), response)
}

func sortedSubjects(state publish.PublishedState) []publish.SubjectView {
	subjects := make([]publish.SubjectView, 0, len(state.Subjects))
	for _, subject := range state.Subjects {
		subjects = append(subjects, subject)
	}
	sort.Slice(subjects, func(i, j int) bool {
		if subjects[i].Key.Kind == subjects[j].Key.Kind {
			return subjects[i].Key.Slug < subjects[j].Key.Slug
		}
		return subjects[i].Key.Kind < subjects[j].Key.Kind
	})
	return subjects
}

func filterSubjects(subjects []publish.SubjectView, query string, styles bool) []publish.SubjectView {
	query = strings.ToLower(strings.TrimSpace(query))
	filtered := make([]publish.SubjectView, 0, len(subjects))
	for _, subject := range subjects {
		if styles != (subject.Key.Kind != publish.SubjectComponent) {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(subject.Key.Slug), query) && !strings.Contains(strings.ToLower(subject.Name), query) {
			continue
		}
		filtered = append(filtered, subject)
	}
	return filtered
}

func subjectListPath(styles bool) string {
	if styles {
		return "/styles"
	}
	return "/components"
}

func subjectURL(subject publish.SubjectView) string {
	section := "styles"
	if subject.Key.Kind == publish.SubjectComponent {
		section = "components"
	}
	parts := strings.Split(subject.Key.Slug, "/")
	for index, part := range parts {
		parts[index] = url.PathEscape(part)
	}
	return "/" + section + "/" + strings.Join(parts, "/")
}

func subjectChangesURL(subject publish.SubjectView) string {
	return subjectURL(subject) + "/changes"
}

func subjectKindLabel(kind publish.SubjectKind) string {
	if kind == publish.SubjectComponent {
		return "Component"
	}
	if kind == publish.SubjectStyle {
		return "Style"
	}
	return string(kind)
}

func tokenContextLabel(tags []string) string {
	if len(tags) == 0 {
		return "Default context"
	}
	return strings.Join(tags, " / ")
}

func subjectChanges(state publish.PublishedState, subject string) []publish.ChangeView {
	changes := make([]publish.ChangeView, 0)
	for _, change := range state.Changes {
		if change.Subject == subject {
			changes = append(changes, change)
		}
	}
	return changes
}

func changePageTitle(subject string) string {
	if subject == "" {
		return "Change log"
	}
	return subject + " changes"
}

func changeTypeLabel(changeType string) string {
	switch changeType {
	case "added":
		return "Added"
	case "removed":
		return "Removed"
	case "changed":
		return "Changed"
	default:
		return changeType
	}
}

func changeValue(value string) string {
	if value == "" {
		return "—"
	}
	var formatted bytes.Buffer
	if json.Indent(&formatted, []byte(value), "", "  ") == nil {
		return formatted.String()
	}
	return value
}
