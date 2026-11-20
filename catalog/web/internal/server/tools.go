// Tools share Catalog navigation and may be served from an independent static build; mounting must preserve their own routes.
package server

import (
	"fmt"
	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	webrender "github.com/sojupie/tokens/catalog/web/internal/render"
	"github.com/sojupie/tokens/forge/site"
	"github.com/sojupie/tokens/source/publish"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

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
func (server *Server) toolDestinations() []webview.ToolDestination {
	external := func(target string) string {
		if strings.HasPrefix(target, "/") {
			return server.browserOrigin + target
		}
		return target
	}
	return []webview.ToolDestination{
		{Name: "Token Foundry", Description: "Open local DTCG token files or folders, edit tokens and variants in a table, and review the file diff before saving.", URL: site.FoundryURL, Image: "tool-foundry.svg", ImageAlt: "Token Foundry workspace showing DTCG token groups, resolved alias values, and a pending diff."},
		{Name: "Material Motion Lab", Description: "Measure how an interface moves in a recording. Compare its timing and motion with another video or Material Design 3 reference curves.", URL: external(server.motionURL), Image: "tool-motion.svg", ImageAlt: "A tracked carousel card in a video, with its measured width change compared against MD3 standard easing."},
		{Name: "Carousel playground", Description: "Help investigate uneven carousel motion after dragging on iOS. Try different implementations on your device and record how they settle.", URL: external(server.motionLabURL), Image: "tool-carousel-lab.svg", ImageAlt: "Full recorded returns to rest for the same carousel baseline: about 304 ms on PC and 739 ms on iOS."},
	}
}
func (server *Server) handleTools(w http.ResponseWriter, r *http.Request) {
	if err := server.renderHTML(w, r, webrender.ToolsPage()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// MountTools serves built browser tools outside the catalog's collection mount.
// It exposes files only: no uploads, project store, processing API or directory listings.
func MountTools(catalog http.Handler, directory string) (http.Handler, error) {
	mux := http.NewServeMux()
	for _, prefix := range []string{site.MotionLabPath, site.CarouselPlaygroundPath} {
		name := path.Base(prefix)
		root := filepath.Join(directory, name)
		if info, err := os.Stat(filepath.Join(root, "index.html")); err != nil || !info.Mode().IsRegular() {
			return nil, fmt.Errorf("missing built tool %s/index.html in %s", name, directory)
		}
		mux.Handle(prefix, http.StripPrefix(prefix, toolFiles(root)))
	}
	const oldMotionPath = "/tools/motion"
	redirectMotion := func(w http.ResponseWriter, r *http.Request) {
		suffix := strings.TrimPrefix(strings.TrimPrefix(r.URL.EscapedPath(), oldMotionPath), "/")
		target := site.MotionLabPath + suffix
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusPermanentRedirect)
	}
	mux.HandleFunc("GET "+oldMotionPath, redirectMotion)
	mux.HandleFunc("GET "+oldMotionPath+"/", redirectMotion)
	mux.Handle("/", catalog)
	return mux, nil
}

func toolFiles(root string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "Static application", http.StatusMethodNotAllowed)
			return
		}
		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if name == "." || name == "" {
			name = "index.html"
		}

		if strings.HasSuffix(name, ".br") || strings.HasSuffix(name, ".gz") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Cache-Control", "public, max-age=0, must-revalidate")
		addEncodingVary(w.Header())
		selected := name
		if r.Header.Get("Range") == "" {
			for _, encoding := range []struct{ name, suffix string }{{"br", ".br"}, {"gzip", ".gz"}} {
				if !acceptsEncoding(r, encoding.name) {
					continue
				}
				if info, err := os.Stat(filepath.Join(root, name+encoding.suffix)); err == nil && info.Mode().IsRegular() {
					selected += encoding.suffix
					w.Header().Set("Content-Encoding", encoding.name)
					break
				}
			}
		}
		file, err := os.OpenInRoot(root, selected)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer file.Close()
		info, err := file.Stat()
		if err != nil || !info.Mode().IsRegular() {
			http.NotFound(w, r)
			return
		}
		if kind := mime.TypeByExtension(filepath.Ext(name)); kind != "" {
			w.Header().Set("Content-Type", kind)
		}
		http.ServeContent(w, r, name, info.ModTime(), file)
	})
}

// SiteNavigation is also used when exporting the shell for static tools.
func SiteNavigation(state publish.PublishedState) []site.Section {
	sections := []site.Section{{Name: "Components", Path: "/components"}, {Name: "Styles", Path: "/styles"}}
	groups := map[string][]site.Destination{}
	for _, subject := range state.Subjects {
		index := 0
		if subject.Key.Kind == publish.SubjectStyle {
			index = 1
		} else if subject.Key.Kind != publish.SubjectComponent {
			continue
		}
		item := site.Destination{Name: webview.SubjectName(subject), URL: webview.SubjectURL(subject)}
		if index == 0 {
			if group := componentNavigationGroup(subject.Key.Slug); group != "" {
				groups[group] = append(groups[group], item)
				continue
			}
		}
		sections[index].Items = append(sections[index].Items, item)
	}
	sortItems := func(items []site.Destination) {
		sort.Slice(items, func(a, b int) bool {
			return strings.ToLower(items[a].Name) < strings.ToLower(items[b].Name)
		})
	}
	for group, items := range groups {
		sortItems(items)
		sections[0].Items = append(sections[0].Items, site.Destination{Name: group, Items: items})
	}
	for i := range sections {
		sortItems(sections[i].Items)
	}
	return sections
}

func componentNavigationGroup(slug string) string {
	switch slug {
	case "buttons", "button-groups", "icon-buttons", "split-button", "segmented-buttons", "floating-action-button", "extended-fab", "fab-menu":
		return "Buttons"
	case "date-pickers", "time-pickers":
		return "Date & time pickers"
	case "loading-indicator", "progress-indicators":
		return "Loading & progress"
	case "navigation-bar", "navigation-drawer", "navigation-rail":
		return "Navigation"
	case "bottom-sheets", "side-sheets":
		return "Sheets"
	}
	return ""
}
