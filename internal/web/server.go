package web

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/sojupie/m3-source/internal/publish"
)

type Server struct {
	state publish.PublishedState
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

func New(state publish.PublishedState) *Server {
	return &Server{state: state}
}

func (server *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", server.handleHealth)
	mux.HandleFunc("GET /api/v1/releases", server.handleReleases)
	mux.HandleFunc("GET /api/v1/components", server.handleComponents)
	mux.HandleFunc("GET /api/v1/components/{slug...}", server.handleComponent)
	mux.HandleFunc("GET /", server.handleOverview)
	mux.HandleFunc("GET /components", server.handleOverview)
	mux.HandleFunc("GET /components/{slug...}", server.handleHTMLSubject)
	mux.HandleFunc("GET /styles", server.handleOverview)
	mux.HandleFunc("GET /styles/{slug...}", server.handleHTMLSubject)
	return mux
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
	writeJSON(response, http.StatusOK, matches[0])
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(value)
}
