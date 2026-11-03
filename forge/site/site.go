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

// The page theme supplies the color. A noscript subtree must not contain style
// elements, because history parsers can otherwise activate them on restoration.
var fallbackLogo = func() string {
	before, rest, found := strings.Cut(brandLogo, "<style>")
	if !found {
		return brandLogo
	}
	_, after, _ := strings.Cut(rest, "</style>")
	return before + after
}()

//go:embed site-init.js
var initialLayout string

type Destination struct{ Name, URL string }

type Section struct {
	Name, Path, Icon string
	Items            []Destination
}
type Config struct {
	Catalog, Active, ActiveTool string
	Workspace                   bool
	Agents                      bool
	Tools                       []Destination
	Navigation                  []Section
}

func Default() Config {
	return Config{Catalog: "https://tokencatalog.dev/md3", Active: "/tools", Agents: true, Tools: []Destination{
		{"Material Motion Lab", "https://tokencatalog.dev/tools/motion/"},
		{"Carousel playground", "https://tokencatalog.dev/tools/motion-lab/"},
	}}
}
func (c Config) drawerBreakpoint() string {
	if c.Workspace {
		return "1440"
	}
	return "1100"
}

func (c Config) sections() []Section {
	sections := []Section{{Name: "Components", Path: "/components", Icon: "grid_view"}, {Name: "Styles", Path: "/styles", Icon: "palette"}, {Name: "Changes", Path: "/changes", Icon: "history"}}
	for i := range sections {
		for _, supplied := range c.Navigation {
			if supplied.Path == sections[i].Path {
				sections[i].Items = supplied.Items
			}
		}
	}
	if c.Agents {
		sections = append(sections, Section{Name: "For agents", Path: "/for-agents", Icon: "description"})
	}
	return append(sections, Section{Name: "Tools", Path: "/tools", Icon: "architecture", Items: c.Tools})
}
func (c Config) section() string {
	if c.ActiveTool != "" {
		return "/tools"
	}
	for _, section := range c.sections() {
		if c.Active == section.Path || strings.HasPrefix(c.Active, section.Path+"/") {
			return section.Path
		}
	}
	return ""
}
func (c Config) destinationURL(item Destination) string {
	if strings.HasPrefix(item.URL, "/") {
		return c.Catalog + item.URL
	}
	return item.URL
}
func (c Config) destinationCurrent(item Destination) string {
	if c.Active == item.URL || c.ActiveTool != "" && c.ActiveTool == item.URL {
		return "page"
	}
	return "false"
}

func (c Config) panelSection() string {
	for _, section := range c.sections() {
		if section.Path == c.section() && len(section.Items) > 0 {
			return section.Path
		}
	}
	return ""
}
