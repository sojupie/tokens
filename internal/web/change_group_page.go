package web

import (
	"net/http"
	"net/url"

	"github.com/sojupie/m3-source/internal/publish"
)

const changeGroupsPerPage = 10

// Page whole observation/component groups. Record counts cannot split a group
// or determine which other components appear beside it.
func paginateChangeGroups(changes []publish.ChangeView, page int) ([]publish.ChangeView, int, int, bool, bool) {
	type key struct{ snapshot, subject string }
	positions := make(map[key]int)
	for _, change := range changes {
		group := key{change.SnapshotID, change.Subject}
		if _, exists := positions[group]; !exists {
			positions[group] = len(positions)
		}
	}
	total := len(positions)
	pages := (total + changeGroupsPerPage - 1) / changeGroupsPerPage
	if pages == 0 {
		return nil, 1, 0, false, false
	}
	if page > pages {
		page = pages
	}
	start := (page - 1) * changeGroupsPerPage
	var selected []publish.ChangeView
	for _, change := range changes {
		position := positions[key{change.SnapshotID, change.Subject}]
		if position >= start && position < start+changeGroupsPerPage {
			selected = append(selected, change)
		}
	}
	return selected, page, total, page > 1, page < pages
}

func changeRecordsURL(base, subject, snapshot string) string {
	address, err := url.Parse(base)
	if err != nil {
		return "/changes/records"
	}
	query := address.Query()
	query.Del("page")
	query.Set("subject", subject)
	query.Set("snapshot", snapshot)
	return "/changes/records?" + query.Encode()
}

func (server *Server) handleChangeRecords(response http.ResponseWriter, request *http.Request) {
	if request.URL.Query().Get("subject") == "" {
		http.Error(response, "component or style is required", http.StatusBadRequest)
		return
	}
	changes, _ := filterChangeHistory(server.state, "", request.URL.Query())
	pageChanges, page, previous, next := paginateChanges(changes, requestedPage(request))
	groups := changeGroups(pageChanges, server.state.Subjects)
	var records []labeledChange
	for _, group := range groups {
		records = append(records, group.Changes...)
	}
	content := changeRecordsContent(records, request.URL.String(), page, len(changes), previous, next)
	response.Header().Set("Vary", "HX-Request")
	if request.Header.Get("HX-Request") != "true" {
		content = changeRecordsPage(timelineTitle(request.URL.Query().Get("subject"), server.state.Subjects), content)
	}
	if err := renderHTML(response, request, content); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
}
