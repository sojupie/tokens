// Package site provides the shared Token Catalog navigation and footer.
package site

import (
	"embed"
	"strings"
)

//go:embed site.css site.js logo.svg logo-250x250.webp logo-250x250.png
var Files embed.FS

//go:embed logo.svg
var brandLogo string

type Destination struct{ Name, URL string }
type Config struct {
	Catalog, Active, ActiveTool string
	Workspace                   bool
	Agents                      bool
	Tools                       []Destination
}

func Default() Config {
	return Config{Catalog: "https://tokencatalog.dev/md3", Active: "/tools", Agents: true, Tools: []Destination{
		{"Material Motion Lab", "https://tokencatalog.dev/tools/motion/"},
		{"Carousel playground", "https://tokencatalog.dev/tools/motion-lab/"},
	}}
}
func (c Config) current(path string) string {
	if path == "/tools" && c.ActiveTool != "" {
		return "false"
	}
	if c.Active == path || strings.HasPrefix(c.Active, path+"/") {
		return "page"
	}
	return "false"
}

func (c Config) drawerBreakpoint() string {
	if c.Workspace {
		return "1440"
	}
	return "1100"
}

func toolCurrent(c Config, tool Destination) string {
	if c.ActiveTool == tool.URL {
		return "page"
	}
	return "false"
}
