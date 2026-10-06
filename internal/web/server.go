package web

import (
	"encoding/json"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sojupie/tokens/internal/publish"
)

type Server struct {
	publicOrigin string
	state        publish.PublishedState
	downloadsDir string
	themeCSS     string
	themeError   error
}

type subjectSummary struct {
	Kind                 publish.SubjectKind `json:"kind"`
	Slug                 string              `json:"slug"`
	Name                 string              `json:"name"`
	TokenCount           int                 `json:"tokenCount"`
	LifecycleNoticeCount int                 `json:"lifecycleNoticeCount"`
	SourcePageURLs       []string            `json:"sourcePageUrls"`
}

type componentListResponse struct {
	PublicationID string               `json:"publicationId"`
	Snapshot      publish.SnapshotView `json:"snapshot"`
	Subjects      []subjectSummary     `json:"subjects"`
}

type healthResponse struct {
	Status        string `json:"status"`
	PublicationID string `json:"publicationId"`
	SnapshotID    string `json:"snapshotId"`
}

type releasesResponse struct {
	PublicationID string                 `json:"publicationId"`
	Current       publish.SnapshotView   `json:"current"`
	Snapshots     []publish.SnapshotView `json:"snapshots"`
}

type changesResponse struct {
	PublicationID string               `json:"publicationId"`
	Snapshot      publish.SnapshotView `json:"snapshot"`
	Subject       string               `json:"subject,omitempty"`
	Changes       []publish.ChangeView `json:"changes"`
}

func New(state publish.PublishedState, downloadsDir string) *Server {
	css, err := publicationTheme(state)
	return &Server{state: state, downloadsDir: downloadsDir, themeCSS: css, themeError: err}
}

func (server *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /sitemap.xml", server.handleSitemap)
	mux.HandleFunc("GET /robots.txt", server.handleRobots)
	mux.HandleFunc("GET /theme.css", server.handleTheme)
	mux.HandleFunc("GET /api/v1/health", server.handleHealth)
	mux.HandleFunc("GET /api/v1/releases", server.handleReleases)
	mux.HandleFunc("GET /api/v1/changes", server.handleChanges)
	mux.HandleFunc("GET /api/v1/components", server.handleComponents)
	mux.HandleFunc("GET /api/v1/components/{slug...}", server.handleComponent)
	mux.HandleFunc("GET /static/", handleStaticAsset)
	mux.HandleFunc("GET /downloads/{path...}", server.handleDownload)
	mux.HandleFunc("GET /", server.handleOverview)
	mux.HandleFunc("GET /changes", server.handleGlobalChanges)
	mux.HandleFunc("GET /components", server.handleOverview)
	mux.HandleFunc("GET /components/{slug...}", server.handleHTMLSubject)
	mux.HandleFunc("GET /styles", server.handleOverview)
	mux.HandleFunc("GET /styles/{slug...}", server.handleHTMLSubject)
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if strings.HasPrefix(request.URL.Path, "/api/") || strings.HasPrefix(request.URL.Path, "/downloads/") {
			response.Header().Set("X-Robots-Tag", "noindex")
		}
		mux.ServeHTTP(response, request)
	})
}

func (server *Server) handleHealth(response http.ResponseWriter, _ *http.Request) {
	writeJSON(response, http.StatusOK, healthResponse{
		Status:        "ok",
		PublicationID: server.state.PublicationID,
		SnapshotID:    server.state.CurrentSnapshot.ID,
	})
}

func (server *Server) handleReleases(response http.ResponseWriter, _ *http.Request) {
	writeJSON(response, http.StatusOK, releasesResponse{
		PublicationID: server.state.PublicationID,
		Current:       server.state.CurrentSnapshot,
		Snapshots:     server.state.Snapshots,
	})
}

func (server *Server) handleComponents(response http.ResponseWriter, _ *http.Request) {
	subjects := make([]subjectSummary, 0, len(server.state.Subjects))
	for _, subject := range server.state.Subjects {
		subjects = append(subjects, subjectSummary{
			Kind:                 subject.Key.Kind,
			Slug:                 subject.Key.Slug,
			Name:                 subject.Name,
			TokenCount:           len(subject.Tokens),
			LifecycleNoticeCount: len(subject.LifecycleNotices),
			SourcePageURLs:       append([]string(nil), subject.SourcePageURLs...),
		})
	}
	sort.Slice(subjects, func(i, j int) bool {
		if subjects[i].Kind == subjects[j].Kind {
			return subjects[i].Slug < subjects[j].Slug
		}
		return subjects[i].Kind < subjects[j].Kind
	})
	writeJSON(response, http.StatusOK, componentListResponse{
		PublicationID: server.state.PublicationID,
		Snapshot:      server.state.CurrentSnapshot,
		Subjects:      subjects,
	})
}

func (server *Server) handleComponent(response http.ResponseWriter, request *http.Request) {
	slug := request.PathValue("slug")
	if strings.HasSuffix(slug, "/changes") {
		server.writeSubjectChanges(response, request, strings.TrimSuffix(slug, "/changes"))
		return
	}
	matches := make([]publish.SubjectView, 0, 1)
	for key, subject := range server.state.Subjects {
		if key.Slug == slug {
			matches = append(matches, subject)
		}
	}
	if len(matches) == 0 {
		http.NotFound(response, request)
		return
	}
	if len(matches) > 1 {
		http.Error(response, "subject slug is ambiguous", http.StatusConflict)
		return
	}
	writeJSON(response, http.StatusOK, matches[0].PublicDocument(server.state.CurrentSnapshot.CreatedAt))
}

func (server *Server) handleChanges(response http.ResponseWriter, _ *http.Request) {
	writeJSON(response, http.StatusOK, changesResponse{
		PublicationID: server.state.PublicationID,
		Snapshot:      server.state.CurrentSnapshot,
		Changes:       server.state.Changes,
	})
}

func (server *Server) writeSubjectChanges(response http.ResponseWriter, request *http.Request, subject string) {
	changes := make([]publish.ChangeView, 0)
	for _, change := range server.state.Changes {
		if change.Subject == subject {
			changes = append(changes, change)
		}
	}
	if len(changes) == 0 {
		http.NotFound(response, request)
		return
	}
	writeJSON(response, http.StatusOK, changesResponse{
		PublicationID: server.state.PublicationID,
		Snapshot:      server.state.CurrentSnapshot,
		Subject:       subject,
		Changes:       changes,
	})
}

func (server *Server) handleDownload(response http.ResponseWriter, request *http.Request) {
	relativePath := request.PathValue("path")
	if !safeDownloadPath(relativePath) {
		http.NotFound(response, request)
		return
	}
	if request.URL.Query().Has("scope") {
		server.handleTokenDownload(response, request, relativePath)
		return
	}
	if server.downloadsDir == "" {
		http.NotFound(response, request)
		return
	}
	filePath := filepath.Join(server.downloadsDir, filepath.FromSlash(relativePath))
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		http.NotFound(response, request)
		return
	}
	http.ServeFile(response, request, filePath)
}

func safeDownloadPath(relativePath string) bool {
	if relativePath == "" || strings.Contains(relativePath, `\`) || path.IsAbs(relativePath) {
		return false
	}
	cleaned := path.Clean(relativePath)
	return cleaned == relativePath && cleaned != "." && cleaned != ".." && !strings.HasPrefix(cleaned, "../")
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(value)
}
