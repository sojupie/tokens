package web

import (
	"io/fs"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVersionedAssetsAreCachedAndRevalidated(t *testing.T) {
	for _, name := range []string{"site.css", "site.js", "app.css", "token-table.css", "theme.js", "token-table.js", "prefetch.js", "agent-handoff.js", "htmx.min.js", "logo.png", "logo-250x250.png", "logo-250x250.webp"} {
		t.Run(name, func(t *testing.T) {
			response := httptest.NewRecorder()
			handleStaticAsset(response, httptest.NewRequest("GET", staticURL(name), nil))
			if response.Code != 200 || response.Header().Get("Cache-Control") != "public, max-age=31536000, immutable" {
				t.Fatalf("versioned asset is not cacheable: %d %v", response.Code, response.Header())
			}
			etag := response.Header().Get("ETag")
			request := httptest.NewRequest("GET", "/static/"+name, nil)
			request.Header.Set("If-None-Match", "W/"+etag)
			response = httptest.NewRecorder()
			handleStaticAsset(response, request)
			if response.Code != 304 || response.Body.Len() != 0 || response.Header().Get("Cache-Control") != "public, max-age=0, must-revalidate" {
				t.Fatal("bare URL should revalidate without transferring the file again")
			}
		})
	}
}

func TestNewContentInvalidatesOldAssetVersion(t *testing.T) {
	oldVersion := assetVersion([]byte("old CSS"))
	newContent := []byte("new CSS")
	request := httptest.NewRequest("GET", "/theme.css?v="+oldVersion, nil)
	request.Header.Set("If-None-Match", `"`+oldVersion+`"`)
	response := httptest.NewRecorder()
	serveCachedAsset(response, request, "theme.css", newContent)
	if response.Code != 200 || response.Body.String() != string(newContent) || strings.Contains(response.Header().Get("Cache-Control"), "immutable") {
		t.Fatal("an old URL must not permanently cache new content")
	}
	request = httptest.NewRequest("GET", "/theme.css?v="+assetVersion(newContent), nil)
	response = httptest.NewRecorder()
	serveCachedAsset(response, request, "theme.css", newContent)
	if !strings.Contains(response.Header().Get("Cache-Control"), "immutable") {
		t.Fatal("matching theme version should be cached like a static stylesheet")
	}
}

func TestMissingAssetsAreNotCached(t *testing.T) {
	response := httptest.NewRecorder()
	handleStaticAsset(response, httptest.NewRequest("GET", "/static/missing.css?v=old", nil))
	if response.Code != 404 || response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("do not cache asset errors")
	}
}

func TestAssetVersionTracksContent(t *testing.T) {
	content, err := fs.ReadFile(webStaticFiles, "app.css")
	if err != nil {
		t.Fatal(err)
	}
	if staticURL("app.css") != "/static/app.css?v="+assetVersion(content) || assetVersion(content) == assetVersion(append(content, '\n')) {
		t.Fatal("asset URL must change automatically when its contents change")
	}
}
