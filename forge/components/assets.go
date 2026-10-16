// Package components embeds generated component styles and shared browser behavior.
package components

import "embed"

// Files holds the component assets consumed by the catalog server.
//
//go:embed static/buttons.css static/buttons.js static/components.css
var Files embed.FS
