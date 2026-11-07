// Package web exposes the Catalog HTTP server and shared site-shell configuration.
// The implementation lives in internal/server so consumers do not depend on
// handler internals. Public URLs are defined by that server's route registration.
package web

import (
	"github.com/sojupie/tokens/catalog/web/internal/server"
	"github.com/sojupie/tokens/forge/site"
	"github.com/sojupie/tokens/source/publish"
	"net/http"
)

type Server = server.Server
type Illustration = server.Illustration
type IllustrationManifest = server.IllustrationManifest

const IllustrationWidth = server.IllustrationWidth
const IllustrationHeight = server.IllustrationHeight

func New(state publish.PublishedState, downloadsDir string) *Server {
	return server.New(state, downloadsDir)
}
func SiteNavigation(state publish.PublishedState) []site.Section { return server.SiteNavigation(state) }
func MountTools(catalog http.Handler, directory string) (http.Handler, error) {
	return server.MountTools(catalog, directory)
}
