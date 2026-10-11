package web

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/sojupie/tokens/internal/usage"
)

func (server *Server) SetUsage(store *usage.Store) { server.usage = store }

type usageResponse struct {
	http.ResponseWriter
	status int
	bytes  int64
}

func (w *usageResponse) Unwrap() http.ResponseWriter { return w.ResponseWriter }
func (w *usageResponse) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
		w.ResponseWriter.WriteHeader(status)
	}
}
func (w *usageResponse) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(body)
	w.bytes += int64(n)
	return n, err
}

func (server *Server) observe(next http.Handler) http.Handler {
	if server.usage == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		response := &usageResponse{ResponseWriter: w}
		next.ServeHTTP(response, r)
		if response.status == 0 {
			response.status = http.StatusOK
		}
		p := r.URL.Path
		// Docker's health check is a local request without a Cloudflare ray.
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		ip := net.ParseIP(host)
		if p == "/api/v1/health" || p == "/" && ip != nil && ip.IsLoopback() && r.Header.Get("CF-Ray") == "" {
			return
		}
		if response.status < 400 && (strings.HasPrefix(p, "/static/") || p == "/theme.css" || r.Method == http.MethodOptions) {
			return
		}
		event := usage.Event{Time: start.UTC(), Event: "http_request", Route: "unmatched", Kind: "other", Method: "other", Status: response.status, Success: response.status < 400, Bytes: response.bytes, DurationMS: float64(time.Since(start)) / float64(time.Millisecond)}
		for _, method := range []string{"GET", "HEAD", "POST", "OPTIONS"} {
			if r.Method == method {
				event.Method = method
			}
		}
		if r.Pattern != "" {
			event.Route = r.Pattern
		}
		if p == "/mcp" {
			event.Route = "/mcp"
			event.Kind = "mcp"
		} else if strings.HasPrefix(p, "/api/") {
			event.Kind = "api"
		} else if strings.HasPrefix(p, "/downloads/") {
			event.Kind = "download"
		} else if agentDiscoveryPath(p) || wellKnownPath(p) {
			event.Kind = "discovery"
		} else if response.Header().Get("Content-Type") == "text/markdown; charset=utf-8" {
			event.Kind = "markdown"
		} else if strings.HasPrefix(response.Header().Get("Content-Type"), "text/html") {
			event.Kind = "navigation"
		}
		if r.Header.Get("HX-Request") == "true" {
			event.Kind = "htmx"
		}
		if strings.EqualFold(r.Header.Get("Purpose"), "prefetch") || strings.HasPrefix(strings.ToLower(r.Header.Get("Sec-Purpose")), "prefetch") {
			event.Kind = "prefetch"
		}
		if r.Method == http.MethodHead {
			event.Bytes = 0
		}
		for _, format := range []string{"text/html", "text/markdown", "application/json", "text/plain", "application/linkset+json"} {
			if strings.HasPrefix(response.Header().Get("Content-Type"), format) {
				event.Format = format
			}
		}
		for key, subject := range server.state.Subjects {
			if p == subjectURL(subject) || strings.HasPrefix(p, subjectURL(subject)+"/") || p == subjectDownloadURL(subject) || p == "/api/v1/components/"+key.Slug || p == "/api/v1/components/"+key.Slug+"/changes" {
				event.Subject = fmt.Sprintf("%s/%s", key.Kind, key.Slug)
				break
			}
		}
		if event.Kind == "download" {
			event.Scope = "default"
			for _, scope := range []string{"view", "set", "all", "components", "styles"} {
				if r.URL.Query().Get("scope") == scope {
					event.Scope = scope
				}
			}
		}
		server.usage.Record(event)
	})
}
