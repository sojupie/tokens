// Search discovery and page metadata describe the same published subjects as the visible catalog.
package server

import (
	"encoding/xml"
	"fmt"
	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

// SetPublicOrigin configures absolute SEO URLs before the server starts.
// Never derive these from untrusted Host or forwarded headers.
func (server *Server) SetPublicOrigin(origin string) error {
	if origin == "" {
		server.publicOrigin = ""
		server.basePath = ""
		server.browserOrigin = ""
		return nil
	}
	parsed, err := url.Parse(origin)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Hostname() == "" || parsed.User != nil || (strings.Contains(parsed.Path, "//") || path.Clean(parsed.Path) != strings.TrimSuffix(parsed.Path, "/") && parsed.Path != "" && parsed.Path != "/" || parsed.RawPath != "") || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
		return fmt.Errorf("public origin must be an absolute http(s) origin without credentials, query or fragment, with an optional clean base path")
	}
	server.publicOrigin = strings.TrimSuffix(parsed.String(), "/")
	server.basePath = strings.TrimSuffix(parsed.Path, "/")
	server.browserOrigin = parsed.Scheme + "://" + parsed.Host
	return nil
}

func hasContentFilters(query url.Values) bool {
	if query.Get("view") == "table" {
		return true
	}
	for key, values := range query {
		if strings.HasPrefix(key, "preview.") || key == "q" || strings.HasPrefix(key, "context.") || key == "snapshot" || key == "type" || key == "subject" {
			for _, value := range values {
				if value != "" {
					return true
				}
			}
		}
	}
	return false
}

func (server *Server) basicMetadata(path string, query url.Values) webview.PageMetadata {
	info := webview.PageMetadata{NoIndex: server.publicOrigin == "" || hasContentFilters(query), AgentDiscovery: server.publicOrigin != ""}
	switch path {
	case "/":
		info.Title = "Material Design 3 Tokens, Playgrounds and Change History"
		info.Description = "Browse Material Design 3 components and styles. Try interactive playgrounds, compare token values by context, follow named references, download CSS variables or JSON and review observed changes."
	case "/for-agents":
		info.Title = "Use Material Design 3 Tokens with a Coding Agent"
		info.Description = "Connect a coding agent to Token Catalog. Copy setup instructions, install the skill or use MCP to find Material Design 3 tokens, contexts and observed changes."
	case "/tools":
		info.Title = "Tools for Motion Analysis and Interaction Testing"
		info.Description = "Analyze motion in video recordings, compare reference curves, and inspect carousel interaction behavior with Token Catalog tools."
	case "/components":
		info.Title = "Material Design 3 Components, Playgrounds and Component Tokens"
		info.Description = "Explore tokens and interactive playgrounds for Material Design 3 components, including buttons, navigation and input. Compare context-specific values, named references and observed changes. Download tokens or component CSS."
	case "/styles":
		info.Title = "Material Design 3 Styles, Foundations and Token Sources"
		info.Description = "Explore Material Design 3 styles and token sources, including color, typography and shape. Compare context-specific values, named references and observed changes. Download tokens as CSS variables or JSON."
	case "/changes":
		info.Title = "Material Design 3 Token Change Log"
		info.Description = "Review observed changes across Material Design 3 token snapshots. Compare previous and current values by token, subject and observation date."
	default:
		info.Title = "Material Design 3 Token Change History"
		info.Description = "Track changes to Material Design 3 tokens across captured snapshots, with previous values, new values and context. Dates show when changes were detected."
	}
	if !info.NoIndex {
		info.Canonical = server.publicOrigin + path
		if page, err := strconv.Atoi(query.Get("page")); err == nil && page > 1 && strings.HasSuffix(path, "/changes") && path != "/changes" {
			info.Canonical += "?page=" + strconv.Itoa(page)
		}
	}
	return info
}

type sitemapEntry struct {
	Location string         `xml:"loc"`
	Images   []sitemapImage `xml:"image:image,omitempty"`
}

type sitemapImage struct {
	Location string `xml:"image:loc"`
}

type sitemapDocument struct {
	ImageNS string         `xml:"xmlns:image,attr,omitempty"`
	XMLName xml.Name       `xml:"urlset"`
	XMLNS   string         `xml:"xmlns,attr"`
	URLs    []sitemapEntry `xml:"url"`
}

func (server *Server) handleSitemap(response http.ResponseWriter, request *http.Request) {
	if server.publicOrigin == "" {
		http.Error(response, "Configure serve --public-origin to enable the sitemap.", http.StatusServiceUnavailable)
		return
	}
	document := sitemapDocument{ImageNS: "http://www.google.com/schemas/sitemap-image/1.1", XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	add := func(path string) {
		entry := sitemapEntry{Location: server.publicOrigin + path}
		if image, ok := server.illustrations[path]; ok && image.Files["light-default"] != "" {
			entry.Images = []sitemapImage{{Location: server.publicOrigin + webview.IllustrationFileURL(image.Files["light-default"])}}
		}
		document.URLs = append(document.URLs, entry)
	}
	for _, path := range []string{"/", "/components", "/styles", "/changes", "/for-agents", "/tools"} {
		add(path)
	}
	for _, tool := range server.toolDestinations() {
		if strings.HasPrefix(tool.URL, server.browserOrigin+"/") && !strings.Contains(tool.URL, "?") {
			document.URLs = append(document.URLs, sitemapEntry{Location: tool.URL})
		}
	}
	for _, subject := range sortedSubjects(server.state) {
		add(webview.SubjectURL(subject))
		add(webview.SubjectChangesURL(subject))
		sets := webview.TokenSets(subject)
		if len(sets) <= 1 {
			continue
		}
		defaultSet := webview.DefaultTokenSet(subject, sets)
		for _, set := range sets {
			if set.Key != defaultSet.Key {
				add(webview.TokenSetURL(subject, set))
			}
		}
		add(webview.TokenSetURL(subject, webview.AllTokenSet(subject)))
	}
	data, err := xml.MarshalIndent(document, "", "  ")
	if err != nil {
		http.Error(response, "Could not build sitemap.", http.StatusInternalServerError)
		return
	}
	response.Header().Set("Content-Type", "application/xml; charset=utf-8")
	response.Header().Set("Cache-Control", "no-cache")
	_, _ = fmt.Fprint(response, xml.Header, "<?xml-stylesheet type=\"text/css\" href=\""+server.basePath+"/sitemap.css\"?>\n", string(data), "\n")
}

func handleSitemapCSS(response http.ResponseWriter, _ *http.Request) {
	response.Header().Set("Content-Type", "text/css; charset=utf-8")
	response.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = fmt.Fprint(response, "urlset { display: block; padding: 1em; } url { display: block; margin-bottom: .5em; overflow-wrap: anywhere; } loc { display: block; }")
}

func (server *Server) handleRobots(response http.ResponseWriter, _ *http.Request) {
	response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	response.Header().Set("Cache-Control", "no-cache")
	if server.publicOrigin == "" {
		_, _ = fmt.Fprint(response, "User-agent: *\nDisallow: /\n")
		return
	}
	// The root sitemap can cover both the collection and its sibling tool apps.
	_, _ = fmt.Fprintf(response, "User-agent: *\nContent-Signal: "+contentSignal+"\nAllow: /\nSitemap: %s/sitemap.xml\n", server.browserOrigin)
}
