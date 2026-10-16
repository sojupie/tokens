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
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/md3/lab/buttons?size=all&shape=all", nil))
	if response.Code != 200 {
		t.Fatal(response)
	}
	if count := strings.Count(response.Body.String(), "data-button-case="); count != 650 {
		t.Fatalf("rendered %d of 650 configurations", count)
	}
	if response.Header().Get("X-Robots-Tag") != "noindex" || !strings.Contains(response.Body.String(), `name="robots" content="noindex, follow"`) {
		t.Fatal("gallery is indexable")
	}
	if strings.Contains(response.Body.String(), "md-comp-filled-tonal-button") || !strings.Contains(response.Body.String(), "/md3/static/buttons.js?v=") {
		t.Fatal("gallery did not use current shared button implementation")
	}
}
