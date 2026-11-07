package presentation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"slices"
	"strings"
	"sync"

	"github.com/a-h/templ"
	webassets "github.com/sojupie/tokens/catalog/web/assets"
	"github.com/sojupie/tokens/forge/components"
	"github.com/sojupie/tokens/forge/components/ui"
)

// Each replaceable region collects the modules its descendants render. Links
// declare dependencies for HTMX swaps, including playground-only updates. Full
// documents hoist them to the head; the client prepares fragment styles before insertion.
type ComponentAssetsKey struct{}

// AssetDelivery exposes only the publication lookups that rendering needs.
// HTTP routing, requests and mutable server caches remain in package web.
type AssetDelivery interface {
	StylesheetURL(string) string
	ComponentAssetLinks(*ComponentAssets) ([]byte, error)
	ThumbnailIllustration(string) (Illustration, bool)
}

type ComponentAssets struct {
	Modules        map[string]bool
	HeadModules    map[string]bool
	Server         AssetDelivery
	DeferStyles    bool
	Static         bool
	CatalogScripts []string
}

var CatalogBehaviorScripts = []string{"artwork-canvas.js", "artwork.js", "component-demos.js", "dialog-demo.js"}

func UseCatalogBehavior(name string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, _ io.Writer) error {
		if !slices.Contains(CatalogBehaviorScripts, name) {
			return fmt.Errorf("unknown Catalog behavior: %s", name)
		}
		if assets, ok := ctx.Value(ComponentAssetsKey{}).(*ComponentAssets); ok && !slices.Contains(assets.CatalogScripts, name) {
			assets.CatalogScripts = append(assets.CatalogScripts, name)
		}
		return nil
	})
}

func UseComponentAssets(name string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, _ io.Writer) error {
		assets, ok := ctx.Value(ComponentAssetsKey{}).(*ComponentAssets)
		if !ok {
			return nil
		}
		if !ComponentModuleNames()[name] {
			return fmt.Errorf("unknown component module: %s", name)
		}
		assets.Modules[name] = true
		return nil
	})
}

func ComponentAssetScope(static ...bool) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) error {
		children := templ.GetChildren(ctx)
		parent, ok := ctx.Value(ComponentAssetsKey{}).(*ComponentAssets)
		if !ok {
			return children.Render(templ.ClearChildren(ctx), writer)
		}
		assets := &ComponentAssets{Modules: map[string]bool{}, HeadModules: parent.HeadModules, Server: parent.Server, Static: len(static) > 0 && static[0]}
		assets.DeferStyles = assets.Static && parent.DeferStyles
		ctx = ui.WithAssets(context.WithValue(templ.ClearChildren(ctx), ComponentAssetsKey{}, assets), ui.Assets(assets.Modules))
		var content bytes.Buffer
		if err := children.Render(ctx, &content); err != nil {
			return err
		}
		links, err := assets.Server.ComponentAssetLinks(assets)
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
func SiteNavigationStyles() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) error {
		parent, ok := ctx.Value(ComponentAssetsKey{}).(*ComponentAssets)
		if !ok {
			return nil
		}
		assets := &ComponentAssets{Modules: map[string]bool{"navigation": true}, Server: parent.Server}
		links, err := parent.Server.ComponentAssetLinks(assets)
		if err != nil {
			return err
		}
		if _, err = writer.Write(links); err != nil {
			return err
		}
		if parent.HeadModules == nil {
			parent.HeadModules = map[string]bool{}
		}
		parent.HeadModules["navigation"] = true
		return nil
	})
}

var ComponentModuleOrder = sync.OnceValue(func() []string {
	data, err := fs.ReadFile(webassets.WebStaticFiles, "bundles/modules.json")
	if err != nil {
		panic(err)
	}
	var names []string
	if err := json.Unmarshal(data, &names); err != nil {
		panic(err)
	}
	return names
})
var ComponentModuleNames = sync.OnceValue(func() map[string]bool {
	names := map[string]bool{}
	for _, name := range ComponentModuleOrder() {
		names[name] = true
	}
	return names
})

var ComponentScriptSources = sync.OnceValue(func() string {
	var sources strings.Builder
	sources.WriteByte('{')
	scripts, err := components.ScriptsFor(ComponentModuleOrder()...)
	if err != nil {
		panic(err)
	}

	scripts = append(scripts, CatalogBehaviorScripts...)
	for _, name := range scripts {
		if sources.Len() > 1 {
			sources.WriteByte(',')
		}
		key, _ := json.Marshal(name)
		value, _ := json.Marshal(strings.TrimPrefix(webassets.StaticURL(name), "/static/"))
		fmt.Fprintf(&sources, "%s:%s", key, value)
	}
	sources.WriteByte('}')
	return sources.String()
})

// Shared tokens stay on the document root; rendered modules add scoped tokens.
type PageThemeKey struct{}
type PageThemesContext struct{ Shared string }
