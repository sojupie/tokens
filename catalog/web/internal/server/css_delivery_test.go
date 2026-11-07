package server

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	webassets "github.com/sojupie/tokens/catalog/web/assets"
)

func TestDeliveryCSSKeepsOriginalAssetsAndDownloads(t *testing.T) {
	directory := t.TempDir()
	folder := filepath.Join(directory, "delivery-css")
	_ = os.Mkdir(folder, 0755)
	original := ":root{--long-token:red}"
	shortened := ":root{--a:red}"
	hash := webassets.AssetVersion([]byte(shortened))
	file := hash + ".css"
	pack := cssDelivery{ID: "0123456789abcdef", Files: map[string]cssDeliveryFile{"/theme.css": {File: file, SourceSha256: webassets.AssetVersion([]byte(original)), Sha256: hash}}}
	data, _ := json.Marshal(pack)
	_ = os.WriteFile(filepath.Join(folder, "manifest.json"), data, 0644)
	_ = os.WriteFile(filepath.Join(folder, file), []byte(shortened), 0644)
	server := &Server{themeCSS: original, tokenAssetsDir: directory}
	if err := server.loadDeliveryCSS(directory); err != nil {
		t.Fatal(err)
	}
	url := server.StylesheetURL("/theme.css?v=original")
	if !strings.Contains(url, "css="+pack.ID) {
		t.Fatal(url)
	}
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", url, nil)
	if !server.serveDeliveryCSS(response, request) || response.Body.String() != shortened || !strings.Contains(response.Header().Get("Cache-Control"), "immutable") {
		t.Fatal(response)
	}
	if server.serveDeliveryCSS(httptest.NewRecorder(), httptest.NewRequest("GET", "/theme.css?v=original", nil)) {
		t.Fatal("original CSS URL was intercepted")
	}
	if server.serveDeliveryCSS(httptest.NewRecorder(), httptest.NewRequest("GET", "/downloads/test?css=x", nil)) {
		t.Fatal("download URL was intercepted")
	}
	response = httptest.NewRecorder()
	server.handleTheme(response, httptest.NewRequest("GET", "/theme.css", nil))
	if response.Body.String() != original {
		t.Fatal("original theme changed")
	}
	exportDir := filepath.Join(directory, "dtcg/light-default")
	_ = os.MkdirAll(exportDir, 0755)
	_ = os.WriteFile(filepath.Join(exportDir, "tokens.css"), []byte(original), 0644)
	request = httptest.NewRequest("GET", "/downloads/dtcg/light-default/tokens.css", nil)
	request.SetPathValue("variant", "light-default")
	request.SetPathValue("file", "tokens.css")
	response = httptest.NewRecorder()
	server.handleDTCGDownload(response, request)
	if response.Body.String() != original {
		t.Fatal("download changed", response.Body.String())
	}
	server.themeCSS = original + " /* changed */"
	if err := server.loadDeliveryCSS(directory); err == nil {
		t.Fatal("stale CSS pack accepted")
	}
}
