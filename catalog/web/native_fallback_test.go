package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNativeDownloadPageUsesServerResolvedChoices(t *testing.T) {
	handler := New(seoTestState(), "").Handler()
	for _, path := range []string{"/download", "/download?content=tokens&scope=collection&format=source"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", path, response.Code, response.Body.String())
		}
		body := response.Body.String()
		for _, want := range []string{`action="/download"`, `formaction="/downloads/export"`, `Apply choices`, `name="format" value="source"`, `name="scope" value="collection"`} {
			if !strings.Contains(body, want) {
				t.Fatalf("%s missing %s", path, want)
			}
		}
		if strings.Contains(path, "format=source") && !strings.Contains(body, `value="source" checked`) {
			t.Fatal("server did not preserve the submitted format")
		}
	}
}
