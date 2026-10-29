package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"io/fs"
	"strings"
	"sync"

	"github.com/a-h/templ"
)

// Each replaceable region collects the modules its descendants render. Links
// travel with that region during HTMX swaps, including playground-only updates.
type componentAssetsKey struct{}
type componentAssets struct {
	modules     map[string]bool
	server      *Server
	deferStyles bool
}

func useComponentAssets(name string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, _ io.Writer) error {
		assets, ok := ctx.Value(componentAssetsKey{}).(*componentAssets)
		if !ok {
			return nil
		} // Standalone template renders do not deliver a page.
		if !componentModuleNames()[name] {
			return fmt.Errorf("unknown component module: %s", name)
		}
		assets.modules[name] = true
		return nil
	})
}

// Gallery artwork can stay blurred while its styles arrive. Detail pages retain
// blocking styles so their interactive controls never paint unfinished.
func deferThumbnailStyles() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, _ io.Writer) error {
		if assets, ok := ctx.Value(componentAssetsKey{}).(*componentAssets); ok {
			assets.deferStyles = true
		}
		return nil
	})
}

func componentAssetScope() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) error {
		children := templ.GetChildren(ctx)
		parent, ok := ctx.Value(componentAssetsKey{}).(*componentAssets)
		if !ok {
			return children.Render(templ.ClearChildren(ctx), writer)
		}
		assets := &componentAssets{modules: map[string]bool{}, server: parent.server}
		ctx = context.WithValue(templ.ClearChildren(ctx), componentAssetsKey{}, assets)
		var content bytes.Buffer
		if err := children.Render(ctx, &content); err != nil {
			return err
		}
		links, err := assets.server.componentAssetLinks(assets)
		if err != nil {
			return err
		}
		if _, err = writer.Write(links); err != nil {
			return err
		}
		_, err = content.WriteTo(writer)
		return err
	})
}

var componentModuleOrder = sync.OnceValue(func() []string {
	data, err := fs.ReadFile(webStaticFiles, "bundles/modules.json")
	if err != nil {
		panic(err)
	}
	var names []string
	if err := json.Unmarshal(data, &names); err != nil {
		panic(err)
	}
	return names
})
var componentModuleNames = sync.OnceValue(func() map[string]bool {
	names := map[string]bool{}
	for _, name := range componentModuleOrder() {
		names[name] = true
	}
	return names
})

func (server *Server) componentAssetLinks(assets *componentAssets) ([]byte, error) {
	var links strings.Builder
	writeStylesheet := func(url string, pageTheme, blocking bool) {
		url = html.EscapeString(url)
		fmt.Fprintf(&links, `<link rel="stylesheet" href="%s" data-component-stylesheet`, url)
		if pageTheme {
			links.WriteString(` data-page-theme`)
		}
		if blocking {
			links.WriteString(`>`)
			return
		}
		fmt.Fprintf(&links, ` media="print" onload="this.media='all'" onerror="this.dataset.stylesheetError=''"><noscript><link rel="stylesheet" href="%s"></noscript>`, url)
	}
	for _, name := range componentModuleOrder() {
		if name == "common" || !assets.modules[name] {
			continue
		}
		if len(server.themeBundles) > 0 {
			theme, ok := server.themeBundles[name]
			if !ok {
				return nil, fmt.Errorf("missing component theme: %s; rebuild token assets", name)
			}
			writeStylesheet(fmt.Sprintf("/themes/%s.css?v=%s", name, assetVersion(theme)), true, !assets.deferStyles || name == "navigation")
		}
		// Navigation is initial shell layout; optional specimens can arrive later.
		writeStylesheet(staticURL("bundles/"+name+".css"), false, !assets.deferStyles || name == "navigation")
	}
	return []byte(links.String()), nil
}

var componentScriptSources = sync.OnceValue(func() string {
	sources := map[string]string{}
	for _, name := range []string{
		"runtime/chips.js", "runtime/tabs.js", "runtime/button-groups.js", "runtime/segmented-buttons.js", "runtime/switches.js", "runtime/navigation.js",
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

// Shared tokens stay on the document root; rendered modules add scoped tokens.
type pageThemeKey struct{}
type pageThemesContext struct{ Shared string }

func (server *Server) pageThemes() pageThemesContext {
	if data, ok := server.themeBundles["shared"]; ok {
		return pageThemesContext{Shared: "/themes/shared.css?v=" + assetVersion(data)}
	}
	return pageThemesContext{}
}
