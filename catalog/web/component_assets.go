package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sojupie/tokens/forge/components"
	"github.com/sojupie/tokens/forge/components/ui"
	xhtml "golang.org/x/net/html"
	"html"
	"io"
	"io/fs"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/a-h/templ"
)

// Each replaceable region collects the modules its descendants render. Links
// declare dependencies for HTMX swaps, including playground-only updates. Full
// documents hoist them to the head; the client prepares fragment styles before insertion.
type componentAssetsKey struct{}
type componentAssets struct {
	modules        map[string]bool
	headModules    map[string]bool
	server         *Server
	deferStyles    bool
	static         bool
	catalogScripts []string
}

var catalogBehaviorScripts = []string{"artwork-canvas.js", "artwork.js", "component-demos.js", "dialog-demo.js"}

func useCatalogBehavior(name string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, _ io.Writer) error {
		if !slices.Contains(catalogBehaviorScripts, name) {
			return fmt.Errorf("unknown Catalog behavior: %s", name)
		}
		if assets, ok := ctx.Value(componentAssetsKey{}).(*componentAssets); ok && !slices.Contains(assets.catalogScripts, name) {
			assets.catalogScripts = append(assets.catalogScripts, name)
		}
		return nil
	})
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

func componentAssetScope(static ...bool) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) error {
		children := templ.GetChildren(ctx)
		parent, ok := ctx.Value(componentAssetsKey{}).(*componentAssets)
		if !ok {
			return children.Render(templ.ClearChildren(ctx), writer)
		}
		assets := &componentAssets{modules: map[string]bool{}, headModules: parent.headModules, server: parent.server, static: len(static) > 0 && static[0]}
		assets.deferStyles = assets.static && parent.deferStyles
		ctx = ui.WithAssets(context.WithValue(templ.ClearChildren(ctx), componentAssetsKey{}, assets), ui.Assets(assets.modules))
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

// The preserved site shell needs styles whose lifetime spans body replacements.
func siteNavigationStyles() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) error {
		parent, ok := ctx.Value(componentAssetsKey{}).(*componentAssets)
		if !ok {
			return nil
		}
		assets := &componentAssets{modules: map[string]bool{"navigation": true}, server: parent.server}
		links, err := parent.server.componentAssetLinks(assets)
		if err != nil {
			return err
		}
		if _, err = writer.Write(links); err != nil {
			return err
		}
		if parent.headModules == nil {
			parent.headModules = map[string]bool{}
		}
		parent.headModules["navigation"] = true
		return nil
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
	writeStylesheet := func(url string, pageTheme, blocking bool, order int) {
		url = html.EscapeString(url)
		fmt.Fprintf(&links, `<link rel="stylesheet" href="%s" data-component-stylesheet data-component-order="%04d"`, url, order)
		if pageTheme {
			links.WriteString(` data-page-theme`)
		}
		if blocking {
			links.WriteString(`>`)
			return
		}
		fmt.Fprintf(&links, ` media="print" onload="this.media='all'" onerror="this.dataset.stylesheetError=''"><noscript><link rel="stylesheet" href="%s"></noscript>`, url)
	}
	for order, name := range componentModuleOrder() {
		if name == "common" || !assets.modules[name] || assets.headModules[name] {
			continue
		}
		if len(server.themeBundles) > 0 {
			theme, ok := server.themeBundles[name]
			if !ok {
				return nil, fmt.Errorf("missing component theme: %s; rebuild token assets", name)
			}
			writeStylesheet(fmt.Sprintf("/themes/%s.css?v=%s", name, assetVersion(theme)), true, !assets.deferStyles || name == "navigation", order*2)
		}
		// Navigation is initial shell layout; optional specimens can arrive later.
		writeStylesheet(staticURL("bundles/"+name+".css"), false, !assets.deferStyles || name == "navigation", order*2+1)
	}
	var modules []string
	for _, name := range componentModuleOrder() {
		if assets.modules[name] {
			modules = append(modules, name)
		}
	}
	scripts, err := components.ScriptsFor(modules...)
	if err != nil {
		return nil, err
	}
	scripts = append(scripts, assets.catalogScripts...)
	if assets.static {
		// Static progress examples need their SVG paths painted; other thumbnail
		// components use CSS only and must not activate interactive controllers.
		scripts = slices.DeleteFunc(scripts, func(name string) bool { return name != "progress.js" })
	}
	if len(scripts) > 0 {
		data, err := json.Marshal(scripts)
		if err != nil {
			return nil, err
		}
		// Requirements travel with each independently replaced region.
		fmt.Fprintf(&links, `<script type="application/json" data-component-behaviors>%s</script>`, data)
	}
	return []byte(links.String()), nil
}

var componentScriptSources = sync.OnceValue(func() string {
	var sources strings.Builder
	sources.WriteByte('{')
	scripts, err := components.ScriptsFor(componentModuleOrder()...)
	if err != nil {
		panic(err)
	}
	// These scripts belong to Catalog examples rather than exported Forge controls.
	scripts = append(scripts, catalogBehaviorScripts...)
	for _, name := range scripts {
		if sources.Len() > 1 {
			sources.WriteByte(',')
		}
		key, _ := json.Marshal(name)
		value, _ := json.Marshal(strings.TrimPrefix(staticURL(name), "/static/"))
		fmt.Fprintf(&sources, "%s:%s", key, value)
	}
	sources.WriteByte('}')
	return sources.String()
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

// Full documents own their component styles in the head. Fragments retain their
// dependency links for the client to prepare before a swap. Tokenizing leaves
// scripts, template markup and the rest of the rendered HTML unchanged.
func hoistComponentStyles(content []byte) ([]byte, error) {
	if !bytes.Contains(content, []byte("</head>")) {
		return content, nil
	}
	z := xhtml.NewTokenizer(bytes.NewReader(content))
	var body bytes.Buffer
	styles := map[string]string{}
	for {
		kind := z.Next()
		if kind == xhtml.ErrorToken {
			if err := z.Err(); err != io.EOF {
				return nil, err
			}
			break
		}
		start := body.Len()
		body.Write(z.Raw())
		if kind == xhtml.StartTagToken || kind == xhtml.SelfClosingTagToken {
			token := z.Token()
			if token.Data == "link" {
				component, href, order := false, "", ""
				for _, attr := range token.Attr {
					switch attr.Key {
					case "data-component-stylesheet":
						component = true
					case "href":
						href = attr.Val
					case "data-component-order":
						order = attr.Val
					}
				}
				if component {
					attrs := token.Attr[:0]
					for _, attr := range token.Attr {
						if attr.Key != "media" && attr.Key != "onload" && attr.Key != "onerror" {
							attrs = append(attrs, attr)
						}
					}
					token.Attr = append(attrs, xhtml.Attribute{Key: "onerror", Val: "this.dataset.stylesheetError=''"})
					styles[order+"\n"+href] = token.String()
					body.Truncate(start)
					continue
				}
			}
		}
	}
	keys := make([]string, 0, len(styles))
	for key := range styles {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var links strings.Builder
	for _, key := range keys {
		links.WriteString(styles[key])
	}
	return bytes.Replace(body.Bytes(), []byte("</head>"), []byte(links.String()+"</head>"), 1), nil
}
