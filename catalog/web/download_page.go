package web

import "net/http"

func (server *Server) handleDownloadPage(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if !query.Has("content") {
		_, _, subject, _ := server.selectedExport(query)
		query = initialExportDialog(server.exportPage(subject, query)).Query
	}
	state := server.exportDialog(r.Context(), query)
	if err := server.renderHTML(w, r, downloadPage(state)); err != nil {
		http.Error(w, "Could not render download options.", http.StatusInternalServerError)
	}
}

func nativeDownloadURL(page *exportPage) string {
	if page == nil {
		return "/download"
	}
	return "/download?" + initialExportDialog(*page).Query.Encode()
}
