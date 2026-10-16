package web

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sojupie/tokens/source/publish"
)

func TestComponentGalleryRendersNativeStates(t *testing.T) {
	path := os.Getenv("TOKEN_CATALOG_THEME_DB")
	if path == "" {
		t.Skip("set TOKEN_CATALOG_THEME_DB to check the captured publication")
	}
	state, err := publish.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	server := New(state, "")
	if err = server.UseGeneratedAssets(filepath.Dir(path)); err != nil {
		t.Fatal(err)
	}
	if err = server.SetPublicOrigin("https://catalog.example/md3"); err != nil {
		t.Fatal(err)
	}
	for _, family := range []string{"all", "chip", "field"} {
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/md3/lab/components?family="+family, nil))
		if response.Code != 200 {
			t.Fatal(response.Body.String())
		}
		body := response.Body.String()
		if !strings.Contains(body, `name="robots" content="noindex, follow"`) {
			t.Fatal("component lab is indexable")
		}
		if strings.Contains(body, `disabled="false"`) || strings.Contains(body, `checked="false"`) {
			t.Fatal("false boolean attributes disable or select native controls")
		}
		if family == "all" && strings.Count(body, "data-component-case=") != 1290 {
			t.Fatal("gallery omitted source configurations")
		}
	}
}
