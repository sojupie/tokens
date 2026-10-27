// Package site provides the shared Token Catalog navigation and footer.
package site

import (
	"embed"
	"strings"
)

//go:embed site.css site.js logo-250x250.webp
var Files embed.FS

type Destination struct{ Name, URL string }
type Config struct {
	Catalog, Logo, Active string
	Agents                bool
	Tools                 []Destination
}

func Default() Config {
	return Config{Catalog: "https://tokencatalog.dev/md3", Logo: "https://tokencatalog.dev/md3/static/logo-250x250.webp", Active: "/tools", Agents: true, Tools: []Destination{
		{"Video motion analysis", "https://tokencatalog.dev/tools/motion/"},
		{"Carousel motion lab", "https://tokencatalog.dev/tools/motion-lab/"},
	}}
}
func (c Config) current(path string) string {
	if c.Active == path || strings.HasPrefix(c.Active, path+"/") {
		return "page"
	}
	return "false"
}
