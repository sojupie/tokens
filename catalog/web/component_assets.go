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
		"time-picker.js", "dialogs.js", "loading.js", "artwork.js", "artwork-canvas.js", "component-demos.js",
	} {
		sources[name] = strings.TrimPrefix(staticURL(name), "/static/")
	}
	data, err := json.Marshal(sources)
	if err != nil {
		panic(err)
	}
	return string(data)
})

// Page themes share the component stylesheet's lifetime. Gallery thumbnails need
// all optional tokens; playgrounds inherit only their own family's tokens.
type pageThemeKey struct{}
type pageThemesContext struct{ Shared, Page string }

func (server *Server) pageThemes(ctx context.Context) pageThemesContext {
	if len(server.themeBundles) == 0 {
		return pageThemesContext{}
	}
	result := pageThemesContext{Shared: "/themes/shared.css?v=" + assetVersion(server.themeBundles["shared"])}
	stylesheet := componentStylesheet(ctx)
	if stylesheet == "" {
		return result
	}
	name := strings.TrimSuffix(strings.TrimPrefix(stylesheet, "bundles/"), ".css")
	data, ok := server.themeBundles[name]
	if !ok {
		name = "all"
		data = server.themeBundles[name]
	}
	result.Page = "/themes/" + name + ".css?v=" + assetVersion(data)
	return result
}

func pageThemeURL(ctx context.Context) string {
	if themes, ok := ctx.Value(pageThemeKey{}).(pageThemesContext); ok {
		return themes.Page
	}
	return ""
}
