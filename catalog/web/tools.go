package web

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type toolDestination struct {
	Name, Description, URL, Image, ImageAlt string
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
		{"Material Motion Lab", "Measure how an interface moves in a recording. Compare its timing and motion with another video or Material Design 3 reference curves.", external(server.motionURL), "tool-motion.svg", "A tracked carousel card in a video, with its measured width change compared against MD3 standard easing."},
		{"Carousel playground", "Help investigate uneven carousel motion after dragging on iOS. Try different implementations on your device and record how they settle.", external(server.motionLabURL), "tool-carousel-lab.svg", "Full recorded returns to rest for the same carousel baseline: about 304 ms on PC and 739 ms on iOS."},
	}
}
func (server *Server) handleTools(w http.ResponseWriter, r *http.Request) {
	if err := server.renderHTML(w, r, toolsPage()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
