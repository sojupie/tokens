package web

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponseCacheValidationAndInvalidation(t *testing.T) {
	version, calls := "first", 0
	source := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path == "/missing" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintf(w, "%s:%s:%s", version, r.URL.RawQuery, r.Header.Get("HX-Target"))
	})
	handler := cacheResponses(source)
	fetch := func(method, path, etag, target string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, nil)
		r.Header.Set("If-None-Match", etag)
		r.Header.Set("HX-Target", target)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	first := fetch("HEAD", "/components", "", "")
	if first.Code != 200 || first.Body.Len() != 0 || calls != 1 {
		t.Fatal("HEAD must seed the GET representation")
	}
	next := fetch("GET", "/components", "", "")
	if next.Body.String() != "first::" || calls != 1 {
		t.Fatal("cache did not reuse rendered response")
	}
	unchanged := fetch("GET", "/components", "W/"+next.Header().Get("ETag"), "")
	if unchanged.Code != 304 || unchanged.Body.Len() != 0 || calls != 1 {
		t.Fatal("unchanged response must validate without rendering")
	}
	if fetch("GET", "/components?q=other", next.Header().Get("ETag"), "").Code != 200 || fetch("GET", "/components", next.Header().Get("ETag"), "component-preview").Code != 200 {
		t.Fatal("query and fragment variants must remain separate")
	}
	for range 2 {
		fetch("GET", "/missing", "", "")
	}
	if calls != 5 {
		t.Fatalf("errors were cached: calls=%d", calls)
	}
	version = "second"
	handler = cacheResponses(source)
	changed := fetch("GET", "/components", next.Header().Get("ETag"), "")
	if changed.Code != 200 || changed.Header().Get("ETag") == next.Header().Get("ETag") {
		t.Fatal("new publication/handler did not invalidate response")
	}
}

func TestResponseCacheIsBounded(t *testing.T) {
	cache := &responseCache{}
	for i := range 300 {
		cache.put(&cachedResponse{key: fmt.Sprint(i), header: make(http.Header), body: []byte("value")})
	}
	if len(cache.entries) != 256 || cache.get("0") != nil || cache.get("299") == nil {
		t.Fatal("cache must evict old filter combinations")
	}
	cache.put(&cachedResponse{key: "oversized", body: make([]byte, responseCacheBytes+1)})
	if cache.get("oversized") != nil {
		t.Fatal("oversized response exceeded cache budget")
	}
}

func TestBoostedNavigationSharesDocumentCache(t *testing.T) {
	handler := New(seoTestState(), "").Handler()
	get := func(boosted, partial bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest("GET", "/components", nil)
		if boosted || partial {
			r.Header.Set("HX-Request", "true")
		}
		if boosted {
			r.Header.Set("HX-Boosted", "true")
			r.Header.Set("HX-Target", "body")
		}
		if partial {
			r.Header.Set("HX-Target", "subject-list")
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	full, boosted, fragment := get(false, false), get(true, false), get(false, true)
	if full.Header().Get("ETag") != boosted.Header().Get("ETag") || full.Body.String() != boosted.Body.String() {
		t.Fatal("boosted navigation must reuse the complete document")
	}
	if full.Header().Get("ETag") == fragment.Header().Get("ETag") {
		t.Fatal("index filtering must retain its separate fragment")
	}
	if full.Header().Get("Cache-Control") != "public, no-cache" {
		t.Fatal("documents must validate before reuse")
	}
}
