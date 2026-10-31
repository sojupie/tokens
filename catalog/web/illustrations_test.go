package web

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIllustrationMetadataVisibleImageAndSitemapShareOneAsset(t *testing.T) {
	server := New(seoTestState(), "")
	if err := server.SetPublicOrigin("https://tokens.example/md3"); err != nil {
		t.Fatal(err)
	}
	image := Illustration{Path: "/components/snackbar", Alt: "A snackbar confirming a saved note.", Files: map[string]string{}}
	for _, variant := range []string{"light-default", "light-medium", "light-high", "dark-default", "dark-medium", "dark-high"} {
		image.Files[variant] = "snackbar-" + variant + "-0123456789abcdef.jpg"
	}
	server.illustrations = map[string]Illustration{image.Path: image}
	request := httptest.NewRequest("GET", "/md3/components/snackbar", nil)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	body := response.Body.String()
	for _, want := range []string{`property="og:image" content="https://tokens.example/md3/illustrations/images/snackbar-light-default-0123456789abcdef.jpg"`, `<img src="/md3/illustrations/images/snackbar-light-default-0123456789abcdef.jpg"`, `srcset="/md3/illustrations/images/snackbar-dark-high-0123456789abcdef.jpg"`, `width="1200" height="630"`, `name="twitter:card" content="summary_large_image"`, image.Alt} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %s", want)
		}
	}
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/md3/sitemap.xml", nil))
	if !strings.Contains(response.Body.String(), `<image:loc>https://tokens.example/md3/illustrations/images/snackbar-light-default-0123456789abcdef.jpg</image:loc>`) {
		t.Fatal(response.Body.String())
	}
}

func TestIllustrationBuildManifestAndRenderRoutes(t *testing.T) {
	server := New(seoTestState(), "")
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/illustrations/index.json", nil))
	var manifest IllustrationManifest
	if err := json.Unmarshal(response.Body.Bytes(), &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Images) != 4 {
		t.Fatal(manifest)
	}
	for _, path := range []string{"/illustrations/render/components/snackbar?theme=dark&contrast=high", "/illustrations/render/"} {
		response = httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest("GET", path, nil))
		if response.Code != 200 || response.Header().Get("X-Robots-Tag") == "" || !strings.Contains(response.Body.String(), `data-component-assets class="tokens-illustration"`) {
			t.Fatalf("%s: %d %s", path, response.Code, response.Body.String())
		}
	}
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/illustrations/render/components/unknown", nil))
	if response.Code != 404 {
		t.Fatal(response.Code)
	}
}

func TestIllustrationManifestRejectsWrongPublicationAndMissingFiles(t *testing.T) {
	server := New(seoTestState(), "")
	server.tokenAssetsDir = t.TempDir()
	directory := filepath.Join(server.tokenAssetsDir, "illustrations")
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	manifest := IllustrationManifest{PublicationID: "another-publication"}
	write := func() {
		data, _ := json.Marshal(manifest)
		if err := os.WriteFile(filepath.Join(directory, "manifest.json"), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	write()
	if err := server.UseIllustrations(); err == nil {
		t.Fatal("accepted another publication")
	}
	manifest.PublicationID = server.state.PublicationID
	manifest.Images = []Illustration{{Files: map[string]string{"light-default": "missing.jpg"}}}
	write()
	if err := server.UseIllustrations(); err == nil {
		t.Fatal("accepted missing image")
	}
}

func TestOlderContentAddressedIllustrationRemainsAvailable(t *testing.T) {
	server := New(seoTestState(), "")
	server.tokenAssetsDir = t.TempDir()
	directory := filepath.Join(server.tokenAssetsDir, "illustrations")
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	file := "components-snackbar-light-default-0123456789abcdef.jpg"
	if err := os.WriteFile(filepath.Join(directory, file), []byte{0xff, 0xd8, 0xff, 0xd9}, 0644); err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/illustrations/images/"+file, nil))
	if response.Code != 200 || !strings.Contains(response.Header().Get("Cache-Control"), "immutable") {
		t.Fatalf("%d %v", response.Code, response.Header())
	}
}

func TestCatalogThumbnailUsesItsOwnThemeAwareImage(t *testing.T) {
	server := New(seoTestState(), "")
	image := Illustration{Path: "/components/snackbar", Name: "Snackbar", Files: map[string]string{}, Thumbnails: map[string]string{}}
	for _, variant := range []string{"light-default", "light-medium", "light-high", "dark-default", "dark-medium", "dark-high"} {
		image.Files[variant] = "scene-" + variant + "-0123456789abcdef.jpg"
		image.Thumbnails[variant] = "thumbnail-" + variant + "-0123456789abcdef.jpg"
	}
	server.illustrations = map[string]Illustration{image.Path: image}
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/components", nil))
	body := response.Body.String()
	for _, want := range []string{"thumbnail-light-default-0123456789abcdef.jpg", "thumbnail-dark-high-0123456789abcdef.jpg", `width="720" height="450"`, `loading="lazy"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("thumbnail missing %s", want)
		}
	}
	if strings.Contains(body, `data-thumbnail="snackbar"`) || strings.Contains(body, "scene-light-default-0123456789abcdef.jpg") {
		t.Fatal("catalog rendered live specimen or editorial image instead of thumbnail")
	}
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/illustrations/render/components/snackbar?composition=thumbnail", nil))
	if response.Code != 200 || !strings.Contains(response.Body.String(), `data-composition="thumbnail"`) {
		t.Fatal("thumbnail source unavailable")
	}
}
