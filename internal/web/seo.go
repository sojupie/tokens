package web

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type pageMetadata struct {
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
		return nil
	}
	parsed, err := url.Parse(origin)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Hostname() == "" || parsed.User != nil || (parsed.Path != "" && parsed.Path != "/") || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
		return fmt.Errorf("public origin must be an absolute http(s) origin without credentials, path, query or fragment")
	}
	server.publicOrigin = strings.TrimSuffix(parsed.String(), "/")
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
		info.Title = "Material Design Tokens and Change History"
		info.Description = "Material Design tokens, named references and change history. Find a component or style and download its tokens as JSON."
	case "/components":
		info.Title = "Material Design Component Tokens"
		info.Description = "Find tokens for Material Design components, from buttons to navigation. View references and values, download JSON and see what changed."
	case "/styles":
		info.Title = "Material Design Styles and System Tokens"
		info.Description = "Material Design styles and system tokens, with named references, context-specific values, JSON downloads and change history."
	default:
		info.Title = "Material Design Token Change History"
		info.Description = "Track changes to Material Design tokens across captured snapshots, with previous values, new values and context. Dates show when changes were detected."
	}
	if !info.NoIndex {
		info.Canonical = server.publicOrigin + path
		if page, err := strconv.Atoi(query.Get("page")); err == nil && page > 1 && strings.HasSuffix(path, "/changes") && path != "/changes" {
			info.Canonical += "?page=" + strconv.Itoa(page)
		}
	}
	return info
}
