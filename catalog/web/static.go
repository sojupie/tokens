package web

import (
	"bytes"
	"context"
	"crypto/sha256"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/sojupie/tokens/forge/components"
)

//go:embed static/app.css static/token-table.css static/token-table.js static/theme.js static/site-ui.js static/artwork.js static/webmcp.js static/prefetch.js static/agent-handoff.js static/htmx.min.js static/fonts static/logo.png static/logo-250x250.png static/logo-250x250.webp
var embeddedStatic embed.FS

//go:embed static/theme-init.js
var themeInitialization string

// This trusted, embedded script must run before styles to avoid a theme flash.
func initialThemeScript() templ.Component {
	return templ.Raw("<script>" + themeInitialization + "</script>")
}

var webStaticFiles = mustStaticFiles()

func mustStaticFiles() fs.FS {
	files, err := fs.Sub(embeddedStatic, "static")
	if err != nil {
		panic(err)
	}
	return staticAssets{FS: files}
}

// Component assets have one source in Forge and retain their public static URLs.
type staticAssets struct{ fs.FS }

func (files staticAssets) Open(name string) (fs.File, error) {
	switch name {
	case "buttons.css", "buttons.js", "components.css", "components.js", "progress.js", "loading.js", "loading-shapes.json", "sheets.js", "lists.js", "carousel.js", "search.js", "date-picker.js", "time-picker.js", "fields.js", "dialogs.js", "menus.js":
		return components.Files.Open("static/" + name)
	default:
		return files.FS.Open(name)
	}
}

func readStaticFile(name string) ([]byte, error) {
	return fs.ReadFile(webStaticFiles, strings.TrimPrefix(name, "static/"))
}

func assetVersion(content []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(content))
}

func staticURL(name string) string {
	content, err := fs.ReadFile(webStaticFiles, name)
	if err != nil {
		panic(err) // A referenced embedded asset missing from the build is a bug.
	}
	return "/static/" + name + "?v=" + assetVersion(content)
}

type themeVersionKey struct{}

func themeURL(ctx context.Context) string {
	if version, ok := ctx.Value(themeVersionKey{}).(string); ok {
		return "/theme.css?v=" + version
	}
	return "/theme.css"
}

func serveCachedAsset(response http.ResponseWriter, request *http.Request, name string, content []byte) {
	version := assetVersion(content)
	addEncodingVary(response.Header())
	representation := version
	if acceptsGzip(request) && compressibleAsset(name) && request.Header.Get("Range") == "" {
		encoded, err := gzipBytes(content)
		if err != nil {
			http.Error(response, "asset encoding failed", http.StatusInternalServerError)
			return
		}
		content = encoded
		representation += "-gzip"
		response.Header().Set("Content-Encoding", "gzip")
	}
	response.Header().Set("ETag", `"`+representation+`"`)
	if request.URL.Query().Get("v") == version {
		response.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		// Bare URLs and old version parameters must not cache new content forever.
		response.Header().Set("Cache-Control", "public, max-age=0, must-revalidate")
	}
	http.ServeContent(response, request, name, time.Time{}, bytes.NewReader(content))
}

func handleStaticAsset(response http.ResponseWriter, request *http.Request) {
	name := strings.TrimPrefix(request.URL.Path, "/static/")
	content, err := fs.ReadFile(webStaticFiles, name)
	if err != nil {
		response.Header().Set("Cache-Control", "no-store")
		http.NotFound(response, request)
		return
	}
	serveCachedAsset(response, request, name, content)
}
