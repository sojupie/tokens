package web

import (
	"bytes"
	"container/list"
	"net/http"
	"strings"
	"sync"
)

// Publications, templates and embedded assets are fixed for a running handler.
// Rebuilding/restarting creates an empty cache, so new content needs no purge.
// The byte limit also bounds arbitrary search/filter URL combinations.
const responseCacheBytes = 32 << 20

type cachedResponse struct {
	key     string
	pattern string
	header  http.Header
	body    []byte
}
type responseCache struct {
	sync.Mutex
	entries map[string]*list.Element
	order   *list.List
	bytes   int
}
type bufferedResponse struct {
	header http.Header
	bytes.Buffer
	status int
}

func (w *bufferedResponse) Header() http.Header { return w.header }
func (w *bufferedResponse) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
}
func (w *bufferedResponse) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.Buffer.Write(body)
}

func cacheableRequest(r *http.Request) bool {
	p := r.URL.Path
	return (r.Method == "GET" || r.Method == "HEAD") && r.Header.Get("Range") == "" && p != "/mcp" && p != "/api/v1/health" && !strings.HasPrefix(p, "/usage") && !strings.HasPrefix(p, "/static/") && p != "/theme.css" && !strings.HasPrefix(p, "/themes/") && (!strings.HasPrefix(p, "/downloads/") || r.URL.Query().Has("scope"))
}
func (c *responseCache) get(key string) *cachedResponse {
	c.Lock()
	defer c.Unlock()
	if e := c.entries[key]; e != nil {
		c.order.MoveToFront(e)
		return e.Value.(*cachedResponse)
	}
	return nil
}
func (c *responseCache) put(entry *cachedResponse) {
	size := len(entry.body) + len(entry.key)
	if size > responseCacheBytes {
		return
	}
	c.Lock()
	defer c.Unlock()
	if c.entries == nil {
		c.entries = make(map[string]*list.Element)
		c.order = list.New()
	}
	if c.entries[entry.key] != nil {
		return
	}
	for c.bytes+size > responseCacheBytes || c.order.Len() >= 256 {
		e := c.order.Back()
		old := e.Value.(*cachedResponse)
		delete(c.entries, old.key)
		c.bytes -= len(old.body) + len(old.key)
		c.order.Remove(e)
	}
	c.entries[entry.key] = c.order.PushFront(entry)
	c.bytes += size
}
func etagMatches(value, etag string) bool {
	for _, candidate := range strings.Split(value, ",") {
		candidate = strings.TrimSpace(candidate)
		if candidate == "*" || strings.TrimPrefix(candidate, "W/") == strings.TrimPrefix(etag, "W/") {
			return true
		}
	}
	return false
}
func writeCachedResponse(w http.ResponseWriter, r *http.Request, entry *cachedResponse) {
	r.Pattern = entry.pattern
	for key, values := range entry.header {
		w.Header()[key] = append([]string(nil), values...)
	}
	if etagMatches(r.Header.Get("If-None-Match"), entry.header.Get("ETag")) {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	if r.Method != "HEAD" {
		_, _ = w.Write(entry.body)
	}
}
func cacheResponses(next http.Handler) http.Handler {
	cache := &responseCache{}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !cacheableRequest(r) {
			next.ServeHTTP(w, r)
			return
		}
		// Preserve all response variants; gzip and identity have different strong ETags.
		variant := r.Clone(r.Context())
		variant.Header = r.Header.Clone()
		if r.Header.Get("HX-Boosted") == "true" {
			variant.Header.Del("HX-Request")
			variant.Header.Del("HX-Target")
		}
		key := strings.Join([]string{r.URL.RequestURI(), variant.Header.Get("HX-Request"), variant.Header.Get("HX-Target"), r.Header.Get("Accept"), r.Header.Get("Accept-Encoding")}, "\n")
		if entry := cache.get(key); entry != nil {
			writeCachedResponse(w, r, entry)
			return
		}
		// Render HEAD through GET so it never seeds an empty cached representation.
		request := variant
		request.Method = "GET"
		response := &bufferedResponse{header: make(http.Header)}
		next.ServeHTTP(response, request)
		r.Pattern = request.Pattern
		if response.status != http.StatusOK || strings.Contains(response.header.Get("Cache-Control"), "no-store") || response.header.Get("Set-Cookie") != "" {
			for key, values := range response.header {
				w.Header()[key] = values
			}
			if response.status != 0 {
				w.WriteHeader(response.status)
			}
			if r.Method != "HEAD" {
				_, _ = w.Write(response.Bytes())
			}
			return
		}
		response.header.Set("Cache-Control", "public, no-cache")
		if response.header.Get("Vary") != "" {
			response.header.Add("Vary", "HX-Boosted")
		}
		response.header.Set("ETag", `"`+assetVersion(response.Bytes())+`"`)
		entry := &cachedResponse{key: key, header: response.header, body: response.Bytes(), pattern: request.Pattern}
		cache.put(entry)
		writeCachedResponse(w, r, entry)
	})
}
