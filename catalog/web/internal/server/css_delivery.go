package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	webassets "github.com/sojupie/tokens/catalog/web/assets"
)

// A delivery pack is scoped to the exact source bytes checked at startup.
// Download handlers continue reading original assets and token data.
type cssDeliveryFile struct {
	File, SourceSha256, Sha256 string
	content                    []byte
}
type cssDelivery struct {
	ID    string
	Files map[string]cssDeliveryFile
}

var cssDigest = regexp.MustCompile(`^[0-9a-f]{64}$`)

func (server *Server) originalStylesheet(path string) ([]byte, error) {
	switch {
	case path == "/theme.css":
		return []byte(server.themeCSS), nil
	case strings.HasPrefix(path, "/themes/"):
		data, ok := server.themeBundles[strings.TrimSuffix(strings.TrimPrefix(path, "/themes/"), ".css")]
		if !ok {
			return nil, fmt.Errorf("unknown theme %s", path)
		}
		return data, nil
	case strings.HasPrefix(path, "/static/"):
		return webassets.ReadStaticFile(strings.TrimPrefix(path, "/static/"))
	}
	return nil, fmt.Errorf("invalid delivery stylesheet %s", path)
}

func (server *Server) loadDeliveryCSS(directory string) error {
	data, err := os.ReadFile(filepath.Join(directory, "delivery-css/manifest.json"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var pack cssDelivery
	if err = json.Unmarshal(data, &pack); err != nil {
		return err
	}
	if len(pack.ID) != 16 || len(pack.Files) == 0 {
		return fmt.Errorf("invalid delivery CSS manifest")
	}
	for path, entry := range pack.Files {
		if !cssDigest.MatchString(entry.Sha256) || entry.File != entry.Sha256+".css" {
			return fmt.Errorf("invalid delivery CSS file")
		}
		original, err := server.originalStylesheet(path)
		if err != nil {
			return err
		}
		if webassets.AssetVersion(original) != entry.SourceSha256 {
			return fmt.Errorf("delivery CSS source changed: %s; rebuild release assets", path)
		}
		entry.content, err = os.ReadFile(filepath.Join(directory, "delivery-css", entry.File))
		if err != nil {
			return err
		}
		if webassets.AssetVersion(entry.content) != entry.Sha256 {
			return fmt.Errorf("delivery CSS checksum mismatch: %s", path)
		}
		pack.Files[path] = entry
	}
	server.deliveryCSS = &pack
	return nil
}

// Original URLs stay usable for exports and static consumers. The extra build
// identity prevents browser caches from mixing different property dictionaries.
func (server *Server) StylesheetURL(original string) string {
	path, _, _ := strings.Cut(original, "?")
	if server.deliveryCSS != nil {
		if entry, ok := server.deliveryCSS.Files[path]; ok {
			return path + "?v=" + entry.Sha256 + "&css=" + server.deliveryCSS.ID
		}
	}
	return original
}

func (server *Server) serveDeliveryCSS(w http.ResponseWriter, r *http.Request) bool {
	if !strings.HasSuffix(r.URL.Path, ".css") || !(r.URL.Path == "/theme.css" || strings.HasPrefix(r.URL.Path, "/static/") || strings.HasPrefix(r.URL.Path, "/themes/")) {
		return false
	}
	id := r.URL.Query().Get("css")
	if id == "" {
		return false
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		return false
	}
	if server.deliveryCSS != nil && id == server.deliveryCSS.ID {
		if entry, ok := server.deliveryCSS.Files[r.URL.Path]; ok {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
			serveCachedAsset(w, r, filepath.Base(r.URL.Path), entry.content)
			return true
		}
	}
	w.Header().Set("Cache-Control", "no-store")
	http.NotFound(w, r)
	return true
}
