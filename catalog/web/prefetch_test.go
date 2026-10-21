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
		{cache: "public, no-cache"},
		{purpose: "prefetch", cache: "public, no-cache"},
		{secPurpose: "prefetch", cache: "public, no-cache"},
		{secPurpose: "prefetch;prerender", cache: "public, no-cache"},
		{purpose: "prefetch", hx: "true", cache: "public, no-cache"},
		{purpose: "prefetch", accept: "text/markdown", cache: "public, no-cache"},
	} {
		r := httptest.NewRequest("GET", "/components", nil)
		r.Header.Set("Purpose", check.purpose)
		r.Header.Set("Sec-Purpose", check.secPurpose)
		r.Header.Set("HX-Request", check.hx)
		r.Header.Set("Accept", check.accept)
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, r)
		if w.Code != 200 || w.Header().Get("Cache-Control") != check.cache || w.Header().Get("Vary") != "HX-Request, HX-Target, Accept" {
			t.Fatalf("%+v: status=%d headers=%v", check, w.Code, w.Header())
		}
		if check.hx == "" && check.accept == "" && (!strings.Contains(w.Body.String(), `href="https://github.com/sojupie/tokens"`) || !strings.Contains(w.Body.String(), "prefetch.js?v=")) {
			t.Fatal("full pages must include the GitHub link and prefetch asset")
		}
	}
}
