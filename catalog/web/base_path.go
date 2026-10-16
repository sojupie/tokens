package web

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

// mount keeps internal routing independent of the public path. Root discovery
// aliases advertise the canonical URLs under the mount rather than copying data.
func (server *Server) mount(handler http.Handler) http.Handler {
	if server.basePath == "" {
		return handler
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == server.basePath {
			target := server.basePath + "/"
			if r.URL.RawQuery != "" {
				target += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, target, http.StatusPermanentRedirect)
			return
		}
		if strings.HasPrefix(r.URL.Path, server.basePath+"/") {
			clone := r.Clone(r.Context())
			clone.URL.Path = strings.TrimPrefix(r.URL.Path, server.basePath)
			clone.URL.RawPath = ""
			handler.ServeHTTP(&baseResponseWriter{ResponseWriter: w, base: server.basePath}, clone)
			return
		}
		if agentDiscoveryPath(r.URL.Path) || wellKnownPath(r.URL.Path) || r.URL.Path == "/robots.txt" || r.URL.Path == "/sitemap.xml" || r.URL.Path == "/api/v1/health" {
			handler.ServeHTTP(&baseResponseWriter{ResponseWriter: w, base: server.basePath}, r)
			return
		}
		http.NotFound(w, r)
	})
}

type baseResponseWriter struct {
	http.ResponseWriter
	base string
}

func (w *baseResponseWriter) WriteHeader(status int) {
	for _, key := range []string{"Location", "HX-Push-Url", "HX-Replace-Url"} {
		if value := w.Header().Get(key); strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//") {
			w.Header().Set(key, w.base+value)
		}
	}
	w.ResponseWriter.WriteHeader(status)
}
func (server *Server) baseLinks(value string) string {
	return strings.ReplaceAll(value, "</", "<"+server.basePath+"/")
}

// Rewrite only URL attributes, never token text, script bodies or external URLs.
// Tokenization preserves the document and HTMX fragments without reformatting.
func prefixHTML(body []byte, base string) ([]byte, error) {
	if base == "" {
		return body, nil
	}
	var out bytes.Buffer
	tokenizer := html.NewTokenizer(bytes.NewReader(body))
	for {
		kind := tokenizer.Next()
		if kind == html.ErrorToken {
			if err := tokenizer.Err(); err != io.EOF {
				return nil, err
			}
			return out.Bytes(), nil
		}
		raw := append([]byte(nil), tokenizer.Raw()...)
		if kind == html.StartTagToken || kind == html.SelfClosingTagToken {
			token := tokenizer.Token()
			changed := false
			for i := range token.Attr {
				attr := &token.Attr[i]
				switch attr.Key {
				case "href", "src", "action", "hx-get", "hx-post", "hx-push-url", "hx-replace-url":
					if strings.HasPrefix(attr.Val, "/") && !strings.HasPrefix(attr.Val, "//") {
						attr.Val = base + attr.Val
						changed = true
					}
				}
			}
			if changed {
				out.WriteString(token.String())
				continue
			}
		}
		out.Write(raw)
	}
}
