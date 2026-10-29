package web

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestToolsOutsideCollectionMount(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"motion", "motion-lab"} {
		if err := os.MkdirAll(filepath.Join(dir, name, "assets"), 0755); err != nil {
			t.Fatal(err)
		}
		for file, body := range map[string]string{"index.html": name, "assets/engine.wasm": "wasm", "sw.js": "worker"} {
			if err := os.WriteFile(filepath.Join(dir, name, file), []byte(body), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	var encoded bytes.Buffer
	compressor := gzip.NewWriter(&encoded)
	_, _ = compressor.Write([]byte("motion"))
	_ = compressor.Close()
	if err := os.WriteFile(filepath.Join(dir, "motion", "index.html.gz"), encoded.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	fallback := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/md3/components" {
			_, _ = io.WriteString(w, "catalog")
		} else {
			http.NotFound(w, r)
		}
	})
	handler, err := MountTools(fallback, dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		method, path string
		status       int
		body         string
	}{
		{"GET", "/tools/motion/", 200, "motion"}, {"GET", "/tools/motion-lab/", 200, "motion-lab"},
		{"GET", "/md3/components", 200, "catalog"}, {"GET", "/tools/motion/sw.js", 200, "worker"},
		{"GET", "/tools/motion/assets/engine.wasm", 200, "wasm"}, {"HEAD", "/tools/motion/", 200, ""},
		{"POST", "/tools/motion/import", 405, ""}, {"GET", "/tools/motion/assets/", 404, ""},
		{"GET", "/tools/motion/api/projects", 404, ""}, {"GET", "/tools/motion/index.html.gz", 404, ""},
	} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
		if w.Code != tc.status || (tc.body != "" && w.Body.String() != tc.body) {
			t.Fatalf("%s %s: %d %s", tc.method, tc.path, w.Code, w.Body.String())
		}
	}
	request := httptest.NewRequest("GET", "/tools/motion/", nil)
	request.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, request)
	if w.Header().Get("Content-Encoding") != "gzip" || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/html") {
		t.Fatal(w.Header())
	}
	request = httptest.NewRequest("GET", "/tools/motion/assets/engine.wasm", nil)
	request.Header.Set("Range", "bytes=0-1")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, request)
	if w.Code != 206 || w.Body.String() != "wa" || w.Header().Get("Content-Encoding") != "" {
		t.Fatal(w.Code, w.Body.String())
	}
	if _, err := MountTools(fallback, t.TempDir()); err == nil {
		t.Fatal("missing tool builds accepted")
	}
}
