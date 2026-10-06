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
)

//go:embed static/app.css static/token-table.css static/token-table.js static/theme.js static/htmx.min.js static/fonts
var embeddedStatic embed.FS

var webStaticFiles = mustStaticFiles()

func mustStaticFiles() fs.FS {
	files, err := fs.Sub(embeddedStatic, "static")
	if err != nil {
		panic(err)
	}
	return files
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
	response.Header().Set("ETag", `"`+version+`"`)
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
