package web

import (
	"context"
	"io/fs"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestComponentStylesFollowPageContent(t *testing.T) {
	for _, tc := range []struct{ path, want string }{
		{"/changes", ""}, {"/for-agents", ""}, {"/components/navigation-rail/changes", ""},
		{"/", "bundles/all.css"}, {"/components", "bundles/all.css"},
		{"/components/navigation-rail", "bundles/navigation.css"},
		{"/components/date-pickers", "bundles/date.css"}, {"/components/search", "bundles/search.css"},
	} {
		ctx := context.WithValue(context.Background(), pageMetadataKey{}, pageMetadata{Path: tc.path})
		if got := componentStylesheet(ctx); got != tc.want {
			t.Fatalf("%s: got %q, want %q", tc.path, got, tc.want)
		}
	}
}

func TestLazyAssetsKeepVersionedPublicDelivery(t *testing.T) {
	for _, name := range []string{"runtime/navigation.js", "bundles/common.css", "bundles/date.css", "asset-loader.js"} {
		data, err := fs.ReadFile(webStaticFiles, name)
		if err != nil {
			t.Fatal(err)
		}
		request := httptest.NewRequest("GET", staticURL(name), nil)
		request.Header.Set("Accept-Encoding", "gzip")
		response := httptest.NewRecorder()
		handleStaticAsset(response, request)
		if response.Code != 200 || response.Header().Get("Content-Encoding") != "gzip" {
			t.Fatal(name, response.Code, response.Header())
		}
		request = httptest.NewRequest("GET", staticURL(name), nil)
		request.Header.Set("Accept-Encoding", "gzip")
		request.Header.Set("If-None-Match", response.Header().Get("ETag"))
		response = httptest.NewRecorder()
		handleStaticAsset(response, request)
		if response.Code != 304 || response.Body.Len() != 0 {
			t.Fatal(name, "did not revalidate")
		}
		if !strings.Contains(staticURL(name), assetVersion(data)) {
			t.Fatal("missing content version", name)
		}
	}
}
