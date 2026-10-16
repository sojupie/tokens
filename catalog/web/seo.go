package web

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
)

type pageMetadata struct {
	SiteDownload   downloadPreview
	Title          string
	Description    string
	Canonical      string
	NoIndex        bool
	Heading        string
	AgentDiscovery bool
	JSONAlternate  string
}

type pageMetadataKey struct{}

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

func metadata(ctx context.Context) pageMetadata {
	value, _ := ctx.Value(pageMetadataKey{}).(pageMetadata)
	return value
}

func pageTitle(ctx context.Context, fallback string) string {
	if title := metadata(ctx).Title; title != "" {
		return title
	}
	return fallback
}

func hasContentFilters(query url.Values) bool {
	for key, values := range query {
		if key == "q" || strings.HasPrefix(key, "context.") || key == "snapshot" || key == "type" || key == "subject" {
			for _, value := range values {
				if value != "" {
					return true
				}
			}
		}
	}
	return false
}

func (server *Server) basicMetadata(path string, query url.Values) pageMetadata {
	info := pageMetadata{NoIndex: server.publicOrigin == "" || hasContentFilters(query), AgentDiscovery: server.publicOrigin != ""}
	switch path {
	case "/":
		info.Title = "Material Design 3 Tokens and Change History"
		info.Description = "Browse Material Design 3 components and styles. Compare token values by context, follow named references, download JSON and review observed changes."
	case "/for-agents":
		info.Title = "Use Material Design 3 Tokens with a Coding Agent"
		info.Description = "Connect a coding agent to Token Catalog. Copy setup instructions, install the skill or use MCP to find Material Design 3 tokens, contexts and observed changes."
	case "/components":
		info.Title = "Material Design 3 Components and Component Tokens"
		info.Description = "Explore tokens for Material Design 3 components, including buttons, navigation and input. Compare context-specific values, named references and observed changes."
	case "/styles":
		info.Title = "Material Design 3 Styles, Foundations and Token Sources"
		info.Description = "Explore Material Design 3 styles and token sources, including color, typography and shape. Compare context-specific values, named references and observed changes."
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
