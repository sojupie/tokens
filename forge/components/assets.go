// Package components embeds generated component styles and shared browser behavior.
package components

import "embed"

// Files holds the component assets consumed by the catalog server.
//
//go:embed static/motion.cjs static/buttons.css static/buttons.js static/components.js static/components.css static/progress.js static/loading.js static/sheets.js static/lists.js static/carousel.js static/search.js static/date-picker.js static/time-picker.js static/fields.js static/dialogs.js static/menus.js static/loading-shapes.json
var Files embed.FS
