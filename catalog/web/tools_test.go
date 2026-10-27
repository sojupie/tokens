package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestToolsLinksLeaveTheCatalogMount(t *testing.T) {
	server := New(seoTestState(), "")
	if err := server.SetPublicOrigin("https://tokens.example/md3"); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/md3/", "/md3/tools"} {
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		body := w.Body.String()
		for _, link := range []string{`href="/md3/tools"`, `href="https://tokens.example/tools/motion/"`, `href="https://tokens.example/tools/motion-lab/"`, `hx-boost="false"`} {
			if w.Code != 200 || !strings.Contains(body, link) {
				t.Fatalf("%s: missing tools navigation %s (status %d)", path, link, w.Code)
			}
		}
		if strings.Contains(body, `href="/md3/tools/motion/"`) {
			t.Fatal("tool app incorrectly mounted inside catalog")
		}
	}
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/md3/tools/?q=test", nil))
	if w.Code != http.StatusPermanentRedirect || w.Header().Get("Location") != "/md3/tools?q=test" {
		t.Fatalf("noncanonical tools URL: %d %s", w.Code, w.Header().Get("Location"))
	}
}
func TestToolURLsAreConfigurableAndValidated(t *testing.T) {
	server := New(seoTestState(), "")
	if err := server.SetToolURLs("https://motion.example/", "/lab/"); err != nil {
		t.Fatal(err)
	}
	destinations := server.toolDestinations()
	if destinations[0].URL != "https://motion.example/" || destinations[1].URL != "/lab/" {
		t.Fatal(destinations)
	}
	for _, target := range []string{"", "javascript:alert(1)", "//other.example/", "https://user:secret@example.com/", "relative/path"} {
		if err := server.SetToolURLs(target, "/lab/"); err == nil {
			t.Fatalf("accepted invalid URL %q", target)
		}
	}
}
