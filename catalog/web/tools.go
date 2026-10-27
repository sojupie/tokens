package web

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type toolDestination struct {
	Name, Description, URL string
}

// SetToolURLs configures the separately deployed tools before serving requests.
func (server *Server) SetToolURLs(motion, lab string) error {
	for _, target := range []string{motion, lab} {
		u, err := url.Parse(target)
		if err != nil || u.User != nil || u.Fragment != "" || strings.HasPrefix(target, "//") || !(strings.HasPrefix(target, "/") && u.Host == "" && u.Scheme == "" || (u.Scheme == "https" || u.Scheme == "http") && u.Hostname() != "") {
			return fmt.Errorf("tool URL must be an absolute http(s) URL or a root-relative path")
		}
	}
	server.motionURL, server.motionLabURL = motion, lab
	return nil
}
func (server *Server) toolDestinations() []toolDestination {
	external := func(target string) string {
		if strings.HasPrefix(target, "/") {
			return server.browserOrigin + target
		}
		return target
	}
	return []toolDestination{
		{"Video motion analysis", "Track movement in a recording, compare videos, and measure motion against reference curves.", external(server.motionURL)},
		{"Carousel motion lab", "Help fix jerky carousel settling on iOS. Compare implementations and record tests on your device.", external(server.motionLabURL)},
	}
}
func (server *Server) handleTools(w http.ResponseWriter, r *http.Request) {
	if err := server.renderHTML(w, r, toolsPage()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
