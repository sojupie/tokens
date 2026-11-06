// Asset delivery resolves captured token CSS and component dependencies into cacheable browser resources.
package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	webassets "github.com/sojupie/tokens/catalog/web/assets"
	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	"github.com/sojupie/tokens/forge/components"
	xhtml "golang.org/x/net/html"
	"html"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"
)

// Asset contents are immutable for a running server. Cache compressed variants
// by content identity, including the publication theme, rather than recompressing
// them on each request or conditional request.
var compressedAssets sync.Map

func compressedAsset(version string, content []byte) ([]byte, error) {
	encode, _ := compressedAssets.LoadOrStore(version, sync.OnceValues(func() ([]byte, error) {
		return gzipBytes(content)
	}))
	return encode.(func() ([]byte, error))()
}

func serveCachedAsset(response http.ResponseWriter, request *http.Request, name string, content []byte) {
	version := webassets.AssetVersion(content)
	addEncodingVary(response.Header())
	representation := version
	if acceptsGzip(request) && compressibleAsset(name) && request.Header.Get("Range") == "" {
		encoded, err := compressedAsset(version, content)
		if err != nil {
			http.Error(response, "asset encoding failed", http.StatusInternalServerError)
			return
		}
		content = encoded
		representation += "-gzip"
		response.Header().Set("Content-Encoding", "gzip")
	}
	response.Header().Set("ETag", `"`+representation+`"`)
	if request.URL.Query().Get("v") == version {
		response.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {

		response.Header().Set("Cache-Control", "public, max-age=0, must-revalidate")
	}
	http.ServeContent(response, request, name, time.Time{}, bytes.NewReader(content))
}

func handleStaticAsset(response http.ResponseWriter, request *http.Request) {
	name := strings.TrimPrefix(request.URL.Path, "/static/")
	content, err := fs.ReadFile(webassets.WebStaticFiles, name)
	if err != nil {
		response.Header().Set("Cache-Control", "no-store")
		http.NotFound(response, request)
		return
	}
	serveCachedAsset(response, request, name, content)
}

// UseGeneratedAssets loads Style Dictionary output belonging to this exact
// publication. Publications created before the exporter keep their Go theme.
func (server *Server) UseGeneratedAssets(publicationDir string) error {
	return server.useGeneratedAssets(publicationDir, filepath.Join(publicationDir, "assets"), false)
}

// UseGeneratedAssetsDirectory requires external assets and pins their resolved path.
func (server *Server) UseGeneratedAssetsDirectory(publicationDir, directory string) error {
	resolved, err := filepath.EvalSymlinks(directory)
	if err != nil {
		return err
	}
	return server.useGeneratedAssets(publicationDir, resolved, true)
}

func (server *Server) useGeneratedAssets(publicationDir, directory string, required bool) error {
	data, err := os.ReadFile(filepath.Join(directory, "manifest.json"))
	if os.IsNotExist(err) && !required {
		return nil
	}
	if err != nil {
		return err
	}
	var manifest struct {
		PublicationID, SnapshotID, DatabaseSha256, CssSha256, ButtonsSha256, ComponentsSha256, CssValuesSha256 string
		ThemeBundles                                                                                           map[string]string
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return err
	}
	if manifest.PublicationID != server.state.PublicationID || manifest.SnapshotID != server.state.CurrentSnapshot.ID {
		return fmt.Errorf("generated token assets belong to another publication")
	}
	file, err := os.Open(filepath.Join(publicationDir, "tokens.sqlite"))
	if err != nil {
		return err
	}
	hash := sha256.New()
	_, copyErr := io.Copy(hash, file)
	closeErr := file.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if hex.EncodeToString(hash.Sum(nil)) != manifest.DatabaseSha256 {
		return fmt.Errorf("generated token assets are stale; run npm run build:tokens -- --published %s", publicationDir)
	}
	css, err := os.ReadFile(filepath.Join(directory, "theme.css"))
	if err != nil {
		return err
	}
	sum := sha256.Sum256(css)
	if hex.EncodeToString(sum[:]) != manifest.CssSha256 {
		return fmt.Errorf("generated theme checksum mismatch")
	}
	buttons, err := webassets.ReadStaticFile("static/buttons.css")
	if err != nil {
		return err
	}
	buttonSum := sha256.Sum256(buttons)
	if hex.EncodeToString(buttonSum[:]) != manifest.ButtonsSha256 {
		return fmt.Errorf("generated button bindings differ from the server binary; rebuild token assets and the server")
	}
	components, err := webassets.ReadStaticFile("static/components.css")
	if err != nil {
		return err
	}
	componentSum := sha256.Sum256(components)
	if hex.EncodeToString(componentSum[:]) != manifest.ComponentsSha256 {
		return fmt.Errorf("generated component bindings differ from the server binary; rebuild token assets and the server")
	}

	themes := map[string][]byte{}
	for name, checksum := range manifest.ThemeBundles {
		if name == "" || strings.Trim(name, "abcdefghijklmnopqrstuvwxyz-") != "" {
			return fmt.Errorf("invalid theme bundle name")
		}
		data, err := os.ReadFile(filepath.Join(directory, "themes", name+".css"))
		if err != nil {
			return err
		}
		if webassets.AssetVersion(data) != checksum {
			return fmt.Errorf("generated theme bundle checksum mismatch: %s", name)
		}
		themes[name] = data
	}
	if len(themes) > 0 && (themes["shared"] == nil || themes["all"] == nil) {
		return fmt.Errorf("generated themes need shared and all bundles")
	}
	server.themeBundles = themes
	server.themeCSS = string(css)
	server.themeError = nil
	if data, readErr := os.ReadFile(filepath.Join(directory, "css-values.json")); readErr == nil {
		if manifest.CssValuesSha256 != "" && webassets.AssetVersion(data) != manifest.CssValuesSha256 {
			return fmt.Errorf("CSS value index checksum mismatch")
		}
		var index struct {
			DatabaseSha256 string
			Values         map[string]string
		}
		if err := json.Unmarshal(data, &index); err != nil {
			return err
		}
		if index.DatabaseSha256 != manifest.DatabaseSha256 {
			return fmt.Errorf("CSS value index belongs to another publication")
		}
		server.cssValues = map[string]map[string]string{}
		for key, value := range index.Values {
			name, suffix, _ := strings.Cut(key, "-")
			if suffix != "" {
				suffix = "-" + suffix
			}
			if server.cssValues[name] == nil {
				server.cssValues[name] = map[string]string{}
			}
			server.cssValues[name][suffix] = value
		}
	} else if !os.IsNotExist(readErr) || manifest.CssValuesSha256 != "" {
		return readErr
	}
	server.tokenAssetsDir = directory
	return nil
}

func (server *Server) handleDTCGDownload(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Cache-Control", "public, no-cache")
	variant, file := request.PathValue("variant"), request.PathValue("file")
	valid := false
	for _, allowed := range []string{"light-default", "light-medium", "light-high", "dark-default", "dark-medium", "dark-high"} {
		if variant == allowed {
			valid = true
		}
	}
	if !valid || server.tokenAssetsDir == "" || (file != "tokens.tokens.json" && file != "tokens.css" && file != "report.json") {
		http.NotFound(response, request)
		return
	}
	if file == "tokens.tokens.json" {
		response.Header().Set("Content-Type", "application/design-tokens+json")
	} else if file == "tokens.css" {
		response.Header().Set("Content-Type", "text/css; charset=utf-8")
	} else {
		response.Header().Set("Content-Type", "application/json")
	}
	http.ServeFile(response, request, filepath.Join(server.tokenAssetsDir, "dtcg", variant, file))
}

func (server *Server) ComponentAssetLinks(assets *webview.ComponentAssets) ([]byte, error) {
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
	for order, name := range webview.ComponentModuleOrder() {
		if name == "common" || !assets.Modules[name] || assets.HeadModules[name] {
			continue
		}
		if len(server.themeBundles) > 0 {
			theme, ok := server.themeBundles[name]
			if !ok {
				return nil, fmt.Errorf("missing component theme: %s; rebuild token assets", name)
			}
			writeStylesheet(fmt.Sprintf("/themes/%s.css?v=%s", name, webassets.AssetVersion(theme)), true, !assets.DeferStyles || name == "navigation", order*2)
		}

		writeStylesheet(webassets.StaticURL("bundles/"+name+".css"), false, !assets.DeferStyles || name == "navigation", order*2+1)
	}
	var modules []string
	for _, name := range webview.ComponentModuleOrder() {
		if assets.Modules[name] {
			modules = append(modules, name)
		}
	}
	scripts, err := components.ScriptsFor(modules...)
	if err != nil {
		return nil, err
	}
	scripts = append(scripts, assets.CatalogScripts...)
	if assets.Static {

		scripts = slices.DeleteFunc(scripts, func(name string) bool { return name != "progress.js" })
	}
	if len(scripts) > 0 {
		data, err := json.Marshal(scripts)
		if err != nil {
			return nil, err
		}

		fmt.Fprintf(&links, `<script type="application/json" data-component-behaviors>%s</script>`, data)
	}
	return []byte(links.String()), nil
}

func (server *Server) pageThemes() webview.PageThemesContext {
	if data, ok := server.themeBundles["shared"]; ok {
		return webview.PageThemesContext{Shared: "/themes/shared.css?v=" + webassets.AssetVersion(data)}
	}
	return webview.PageThemesContext{}
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
