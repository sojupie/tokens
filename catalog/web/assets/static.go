package assets

import (
	"crypto/sha256"
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"github.com/a-h/templ"
	"github.com/sojupie/tokens/forge/components"
	"github.com/sojupie/tokens/forge/site"
)

// Browser URLs are independent of the source/output folder layout.
//
//go:embed styles scripts/*.js vendor images fonts generated
var EmbeddedStatic embed.FS

//go:embed scripts/theme-init.js
var ThemeInitialization string

// This trusted, embedded script must run before styles to avoid a theme flash.
func InitialThemeScript() templ.Component {
	return templ.Raw("<script>" + ThemeInitialization + "</script>")
}

var WebStaticFiles = MustStaticFiles()

func MustStaticFiles() fs.FS {
	paths := map[string]string{}
	// Build the public-name index from embedded files; additions need no second manifest.
	err := fs.WalkDir(EmbeddedStatic, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		parts := strings.SplitN(path, "/", 2)
		if len(parts) != 2 {
			return nil
		}
		name := parts[1]
		if parts[0] == "fonts" {
			name = path
		}
		if previous, exists := paths[name]; exists {
			return fmt.Errorf("duplicate public asset %s: %s and %s", name, previous, path)
		}
		paths[name] = path
		return nil
	})
	if err != nil {
		panic(err)
	}
	return StaticAssets{FS: EmbeddedStatic, paths: paths}
}

// Component assets have one source in Forge and retain their public static URLs.
type StaticAssets struct {
	fs.FS
	paths map[string]string
}

func (files StaticAssets) Open(name string) (fs.File, error) {
	if strings.HasPrefix(name, "runtime/") || strings.HasPrefix(name, "bundles/") {
		return components.Files.Open("static/" + name)
	}
	switch name {
	case "site.css", "site.js", "logo.svg", "logo-250x250.webp":
		return site.Files.Open(name)
	case "motion.js":
		return components.Files.Open("static/motion.cjs")
	case "buttons.css", "buttons.js", "components.css", "components.js", "progress.js", "loading.js", "loading-shapes.json", "sheets.js", "lists.js", "carousel.js", "search.js", "date-picker.js", "time-picker.js", "fields.js", "dialogs.js", "menus.js":
		return components.Files.Open("static/" + name)
	default:
		path, ok := files.paths[name]
		if !ok {
			return nil, fs.ErrNotExist
		}
		return files.FS.Open(path)
	}
}

func ReadStaticFile(name string) ([]byte, error) {
	return fs.ReadFile(WebStaticFiles, strings.TrimPrefix(name, "static/"))
}

func AssetVersion(content []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(content))
}

func StaticURL(name string) string {
	content, err := fs.ReadFile(WebStaticFiles, name)
	if err != nil {
		panic(err)
	}
	return "/static/" + name + "?v=" + AssetVersion(content)
}
