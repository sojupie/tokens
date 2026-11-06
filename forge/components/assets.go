// Package components exposes Forge's built assets through stable logical paths.
// Sources live in behaviors/, styles/ and data/; generated/ is rebuildable output.
package components

import (
	"embed"
	"io"
	"io/fs"
	"strings"
)

//go:embed generated data modules.json
var embedded embed.FS

// Files preserves the static/... contract used by Catalog and exported bundles.
// Disk ownership can change without forcing consumers to change their URLs.
var Files assetFiles

type assetFiles struct{}

func (assetFiles) Open(name string) (fs.File, error) {
	if name == "static/loading-shapes.json" {
		return embedded.Open("data/loading-shapes.json")
	}
	if strings.HasPrefix(name, "static/") {
		name = "generated/" + strings.TrimPrefix(name, "static/")
	}
	return embedded.Open(name)
}
func (files assetFiles) ReadFile(name string) ([]byte, error) {
	file, err := files.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}
