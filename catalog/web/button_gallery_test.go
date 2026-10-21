package web

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sojupie/tokens/source/publish"
)

func TestButtonGalleryUsesGeneratedConfigurations(t *testing.T) {
	path := os.Getenv("TOKEN_CATALOG_THEME_DB")
	if path == "" {
		t.Skip("set TOKEN_CATALOG_THEME_DB for generated button gallery checks")
	}
	state, err := publish.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	server := New(state, "")
	if err := server.UseGeneratedAssets(filepath.Dir(path)); err != nil {
		t.Fatal(err)
	}
	if err := server.SetPublicOrigin("https://catalog.example/md3"); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		compare string
		count   int
	}{{"states", 5}, {"sizes", 5}, {"variants", 5}, {"interactive", 1}} {
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/md3/components/buttons?preview.compare="+tc.compare, nil))
		if response.Code != 200 {
			t.Fatal(response.Body.String())
		}
		if count := strings.Count(response.Body.String(), "data-button-case="); count != tc.count {
			t.Fatalf("%s: got %d cases, want %d", tc.compare, count, tc.count)
		}
		if !strings.Contains(response.Body.String(), "Browse tokens") {
			t.Fatal("preview replaced token table")
		}
		if tc.compare == "interactive" && strings.Contains(response.Body.String(), `data-button-state=`) {
			t.Fatal("interactive preview forces an interaction state")
		}
	}
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/md3/components/buttons?preview.size=all", nil))
	if response.Code != 400 {
		t.Fatalf("unsupported size: %d", response.Code)
	}
}
