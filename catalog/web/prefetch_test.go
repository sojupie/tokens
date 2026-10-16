package web

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDocumentPrefetchCacheAndVariants(t *testing.T) {
	server := New(seoTestState(), "")
	for _, check := range []struct {
		purpose, secPurpose, hx, accept, cache string
	}{
		{cache: "no-cache"},
		{purpose: "prefetch", cache: "private, max-age=30"},
		{secPurpose: "prefetch", cache: "private, max-age=30"},
		{secPurpose: "prefetch;prerender", cache: "private, max-age=30"},
		{purpose: "prefetch", hx: "true", cache: "no-cache"},
		{purpose: "prefetch", accept: "text/markdown", cache: "no-cache"},
	} {
		r := httptest.NewRequest("GET", "/components", nil)
		r.Header.Set("Purpose", check.purpose)
		r.Header.Set("Sec-Purpose", check.secPurpose)
		r.Header.Set("HX-Request", check.hx)
		r.Header.Set("Accept", check.accept)
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, r)
		if w.Code != 200 || w.Header().Get("Cache-Control") != check.cache || w.Header().Get("Vary") != "HX-Request, Accept" {
			t.Fatalf("%+v: status=%d headers=%v", check, w.Code, w.Header())
		}
		if check.hx == "" && check.accept == "" && (!strings.Contains(w.Body.String(), `href="https://github.com/sojupie/tokens"`) || !strings.Contains(w.Body.String(), "prefetch.js?v=")) {
			t.Fatal("full pages must include the GitHub link and prefetch asset")
		}
	}
}
