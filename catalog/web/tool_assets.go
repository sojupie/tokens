package web

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// MountTools serves built browser tools outside the catalog's collection mount.
// It exposes files only: no uploads, project store, processing API or directory listings.
func MountTools(catalog http.Handler, directory string) (http.Handler, error) {
	mux := http.NewServeMux()
	for _, name := range []string{"motion", "motion-lab"} {
		root := filepath.Join(directory, name)
		if info, err := os.Stat(filepath.Join(root, "index.html")); err != nil || !info.Mode().IsRegular() {
			return nil, fmt.Errorf("missing built tool %s/index.html in %s", name, directory)
		}
		prefix := "/tools/" + name + "/"
		mux.Handle(prefix, http.StripPrefix(prefix, toolFiles(root)))
	}
	mux.Handle("/", catalog)
	return mux, nil
}

func toolFiles(root string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "Static application", http.StatusMethodNotAllowed)
			return
		}
		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if name == "." || name == "" {
			name = "index.html"
		}
		// Sidecars are negotiated, not separate public resources.
		if strings.HasSuffix(name, ".br") || strings.HasSuffix(name, ".gz") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Cache-Control", "public, max-age=0, must-revalidate")
		addEncodingVary(w.Header())
		selected := name
		if r.Header.Get("Range") == "" {
			for _, encoding := range []struct{ name, suffix string }{{"br", ".br"}, {"gzip", ".gz"}} {
				if !acceptsEncoding(r, encoding.name) {
					continue
				}
				if info, err := os.Stat(filepath.Join(root, name+encoding.suffix)); err == nil && info.Mode().IsRegular() {
					selected += encoding.suffix
					w.Header().Set("Content-Encoding", encoding.name)
					break
				}
			}
		}
		file, err := os.OpenInRoot(root, selected)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer file.Close()
		info, err := file.Stat()
		if err != nil || !info.Mode().IsRegular() {
			http.NotFound(w, r)
			return
		}
		if kind := mime.TypeByExtension(filepath.Ext(name)); kind != "" {
			w.Header().Set("Content-Type", kind)
		}
		http.ServeContent(w, r, name, info.ModTime(), file)
	})
}
