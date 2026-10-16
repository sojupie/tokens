package web

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sojupie/tokens/source/publish"
)

func TestGeneratedAssetsAreBoundToPublication(t *testing.T) {
	directory := t.TempDir()
	assets := filepath.Join(directory, "assets")
	if err := os.MkdirAll(filepath.Join(assets, "dtcg", "light-default"), 0755); err != nil {
		t.Fatal(err)
	}
	css := []byte(":root { --md-sys-color-primary: red; }\n")
	database := []byte("publication database")
	digest := func(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
	components, _ := readStaticFile("static/components.css")
	buttons, _ := readStaticFile("static/buttons.css")
	manifest := map[string]string{"publicationId": "test", "snapshotId": "capture", "databaseSha256": digest(database), "cssSha256": digest(css), "buttonsSha256": digest(buttons), "componentsSha256": digest(components)}
	for file, data := range map[string][]byte{"tokens.sqlite": database, "assets/theme.css": css, "assets/dtcg/light-default/tokens.tokens.json": []byte(`{"size":{"$type":"dimension","$value":{"value":4,"unit":"px"}}}`)} {
		if err := os.WriteFile(filepath.Join(directory, file), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	data, _ := json.Marshal(manifest)
	if err := os.WriteFile(filepath.Join(assets, "manifest.json"), data, 0644); err != nil {
		t.Fatal(err)
	}
	state := publish.PublishedState{PublicationID: "test", CurrentSnapshot: publish.SnapshotView{ID: "capture"}}
	server := New(state, "")
	if err := server.UseGeneratedAssets(directory); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(server.agentGuide(), "/downloads/dtcg/light-default/tokens.tokens.json") {
		t.Fatal("generated exports missing from agent guide")
	}
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/theme.css", nil))
	if response.Code != 200 || response.Body.String() != string(css) {
		t.Fatal(response)
	}
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/downloads/dtcg/light-default/tokens.tokens.json", nil))
	if response.Code != 200 || response.Header().Get("Content-Type") != "application/design-tokens+json" || response.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal(response)
	}
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/downloads/dtcg/unknown/tokens.tokens.json", nil))
	if response.Code != 404 {
		t.Fatal(response)
	}
	if err := os.WriteFile(filepath.Join(directory, "tokens.sqlite"), []byte("new publication"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := server.UseGeneratedAssets(directory); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("accepted stale assets: %v", err)
	}
}
