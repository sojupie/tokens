package web

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNativeDownloadSignalsPreparationResult(t *testing.T) {
	server := scopedExportFixture()
	for _, test := range []struct{ query, result string }{
		{"scope=subject&kind=component&slug=buttons&context.Theme=Light&format=dtcg", "ready"},
		{"scope=subject&kind=component&slug=missing&format=dtcg", "error"},
	} {
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/downloads/export?"+test.query+"&_download=test-123", nil))
		cookies := w.Result().Cookies()
		if len(cookies) != 1 || cookies[0].Name != "tokens-download-test-123" || cookies[0].Value != test.result {
			t.Fatalf("%d: %v", w.Code, cookies)
		}
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("download acknowledgement must not be cached")
		}
		if test.result == "ready" && !strings.HasPrefix(w.Header().Get("Content-Disposition"), "attachment;") {
			t.Fatal("lost native download")
		}
	}
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/downloads/export?scope=subject&kind=component&slug=buttons&context.Theme=Light&format=dtcg", nil))
	if len(w.Result().Cookies()) != 0 {
		t.Fatal("ordinary API downloads must not set acknowledgement cookies")
	}
}
