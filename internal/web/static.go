package web

import (
	"embed"
	"io/fs"
)

//go:embed static/app.css static/token-table.css static/token-table.js static/htmx.min.js
var embeddedStatic embed.FS

var webStaticFiles = mustStaticFiles()

func mustStaticFiles() fs.FS {
	files, err := fs.Sub(embeddedStatic, "static")
	if err != nil {
		panic(err)
	}
	return files
}
