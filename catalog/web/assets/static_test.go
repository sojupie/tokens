package assets

import (
	"io/fs"
	"testing"
)

func TestPublicNamesSurviveSourceAndOutputSeparation(t *testing.T) {
	for _, name := range []string{"app.css", "asset-loader.js", "thumbnail-images.js", "theme.js", "htmx.min.js", "logo.svg", "fonts/roboto-latin.woff2"} {
		data, err := fs.ReadFile(WebStaticFiles, name)
		if err != nil || len(data) == 0 {
			t.Errorf("public asset %s: %v", name, err)
		}
	}
	for _, name := range []string{"scripts/asset-loader.ts", "generated/asset-loader.js", "styles/app.css", "../go.mod"} {
		if _, err := fs.ReadFile(WebStaticFiles, name); err == nil {
			t.Errorf("source path exposed as public asset: %s", name)
		}
	}
}
