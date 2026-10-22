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
	manifest := map[string]any{"publicationId": "test", "snapshotId": "capture", "databaseSha256": digest(database), "cssSha256": digest(css), "buttonsSha256": digest(buttons), "componentsSha256": digest(components)}
	for file, data := range map[string][]byte{"tokens.sqlite": database, "assets/theme.css": css, "assets/dtcg/light-default/tokens.tokens.json": []byte(`{"size":{"$type":"dimension","$value":{"value":4,"unit":"px"}}}`)} {
		if err := os.WriteFile(filepath.Join(directory, file), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(assets, "themes"), 0755); err != nil {
		t.Fatal(err)
	}
	themeCSS := []byte(":root{--shared:red}")
	checksums := map[string]string{}
	for _, name := range []string{"shared", "all", "navigation"} {
		if err := os.WriteFile(filepath.Join(assets, "themes", name+".css"), themeCSS, 0644); err != nil {
			t.Fatal(err)
		}
		checksums[name] = digest(themeCSS)
	}
	manifest["themeBundles"] = checksums
	data, _ := json.Marshal(manifest)
	if err := os.WriteFile(filepath.Join(assets, "manifest.json"), data, 0644); err != nil {
		t.Fatal(err)
	}
	state := publish.PublishedState{PublicationID: "test", CurrentSnapshot: publish.SnapshotView{ID: "capture"}}
	server := New(state, "")
	if err := server.UseGeneratedAssets(directory); err != nil {
		t.Fatal(err)
	}
	responseBundle := httptest.NewRecorder()
	server.Handler().ServeHTTP(responseBundle, httptest.NewRequest("GET", "/themes/navigation.css?v="+digest(themeCSS), nil))
	if responseBundle.Code != 200 || responseBundle.Body.String() != string(themeCSS) || !strings.Contains(responseBundle.Header().Get("Cache-Control"), "immutable") {
		t.Fatal("theme bundle delivery", responseBundle)
	}
	if err := os.WriteFile(filepath.Join(assets, "themes", "navigation.css"), []byte("tampered"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := server.UseGeneratedAssets(directory); err == nil {
		t.Fatal("accepted corrupted theme bundle")
	}
	if err := os.WriteFile(filepath.Join(assets, "themes", "navigation.css"), themeCSS, 0644); err != nil {
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

	external := filepath.Join(t.TempDir(), "assets")
	if err := os.Rename(assets, external); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "current-assets")
	if err := os.Symlink(external, link); err != nil {
		t.Fatal(err)
	}
	if err := server.UseGeneratedAssetsDirectory(directory, link); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/downloads/dtcg/light-default/tokens.tokens.json", nil))
	if response.Code != 200 {
		t.Fatal("external asset path was not pinned", response)
	}
	if err := server.UseGeneratedAssetsDirectory(directory, link); err == nil {
		t.Fatal("accepted missing required assets")
	}
	if err := server.UseGeneratedAssetsDirectory(directory, t.TempDir()); err == nil {
		t.Fatal("accepted directory without manifest")
	}
	if err := os.Symlink(external, assets); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "tokens.sqlite"), []byte("new publication"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := server.UseGeneratedAssets(directory); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("accepted stale assets: %v", err)
	}
}
