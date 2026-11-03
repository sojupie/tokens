package web

import (
	"net/http"
	"strings"
)

// A native attachment has no page-side completion event. Signal when response
// headers are written, after preparation, without buffering the download in JS.
type exportDownloadResponse struct {
	http.ResponseWriter
	token   string
	started bool
}

func (w *exportDownloadResponse) WriteHeader(status int) {
	if w.started {
		return
	}
	w.started = true
	result := "ready"
	if status >= 400 {
		result = "error"
	}
	http.SetCookie(w.ResponseWriter, &http.Cookie{Name: "tokens-download-" + w.token, Value: result, Path: "/", MaxAge: 60, SameSite: http.SameSiteStrictMode})
	w.Header().Set("Cache-Control", "no-store")
	w.ResponseWriter.WriteHeader(status)
}
func (w *exportDownloadResponse) Write(body []byte) (int, error) {
	if !w.started {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}
func downloadSignal(token string) bool {
	return token != "" && len(token) <= 64 && strings.Trim(token, "abcdefghijklmnopqrstuvwxyz0123456789-") == ""
}
