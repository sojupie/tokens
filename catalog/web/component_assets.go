package web

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
)

// Shared site controls use common.css. Optional page bundles keep their original
// rule order and live in the body so HTMX replaces them with the page contents.
func componentStylesheet(ctx context.Context) string {
	path := metadata(ctx).Path
	if path == "/" || path == "/components" {
		return "bundles/all.css"
	}
	if !strings.HasPrefix(path, "/components/") || strings.HasSuffix(path, "/changes") {
		return ""
	}
	slug := strings.Split(strings.TrimPrefix(path, "/components/"), "/")[0]
	family := componentFamily(slug)
	bundle := map[string]string{
		"nav-bar": "navigation", "nav-rail": "navigation", "nav-drawer": "navigation",
		"chip": "chips", "radio": "radio", "checkbox": "checkbox", "switch": "switch",
		"tabs": "tabs", "segmented": "segmented", "slider": "sliders", "progress": "progress",
		"loading": "loading", "appbar": "appbar", "toolbar": "toolbar",
		"bottom-sheet": "sheets", "side-sheet": "sheets", "split": "split",
		"button-group": "button-group", "fab-menu": "fab-menu", "list": "lists",
		"carousel": "carousel", "search": "search", "date": "date", "time": "time",
		"dialog": "dialog", "snackbar": "snackbar", "tooltip": "tooltip",
	}[family]
	if bundle == "" {
		return ""
	}
	return "bundles/" + bundle + ".css"
}

var componentScriptSources = sync.OnceValue(func() string {
	sources := map[string]string{}
	for _, name := range []string{
		"runtime/chips.js", "runtime/tabs.js", "runtime/button-groups.js", "runtime/navigation.js",
		"runtime/sliders.js", "runtime/toolbars.js", "runtime/app-bars.js", "runtime/snackbars.js",
		"runtime/fab-menus.js", "runtime/split-buttons.js", "runtime/tooltips.js",
		"progress.js", "sheets.js", "lists.js", "carousel.js", "search.js", "date-picker.js",
		"time-picker.js", "dialogs.js", "loading.js", "artwork.js", "component-demos.js",
	} {
		sources[name] = strings.TrimPrefix(staticURL(name), "/static/")
	}
	data, err := json.Marshal(sources)
	if err != nil {
		panic(err)
	}
	return string(data)
})
