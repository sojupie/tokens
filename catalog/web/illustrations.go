package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/a-h/templ"
	"github.com/sojupie/tokens/source/publish"
)

const IllustrationWidth, IllustrationHeight = 1200, 630

// Illustration files are build outputs, independent of the live playground state.
type Illustration struct {
	Path       string            `json:"path"`
	Name       string            `json:"name"`
	Alt        string            `json:"alt"`
	Files      map[string]string `json:"files,omitempty"`
	Thumbnails map[string]string `json:"thumbnails,omitempty"`
}
type IllustrationManifest struct {
	PublicationID string         `json:"publicationId"`
	Images        []Illustration `json:"images"`
}

func (server *Server) IllustrationJobs() IllustrationManifest {
	result := IllustrationManifest{PublicationID: server.state.PublicationID}
	for _, subject := range sortedSubjects(server.state) {
		result.Images = append(result.Images, Illustration{Path: subjectURL(subject), Name: subjectName(subject), Alt: illustrationAlt(subject)})
	}
	for _, path := range []string{"/", "/components", "/styles"} {
		result.Images = append(result.Images, Illustration{Path: path, Name: "Token Catalog", Alt: map[string]string{"/": "Material Design 3 cards, buttons, chips and switches in a collection interface.", "/components": "Material Design 3 action buttons, choices and a landscape card.", "/styles": "Material Design 3 typography, color roles and corner shapes."}[path]})
	}
	return result
}

func illustrationAlt(subject publish.SubjectView) string {
	descriptions := map[string]string{
		"app-bars": "An app bar above a journal of saved places.", "badges": "An inbox with unread counts on folders and a notification dot.", "bottom-sheets": "An expanded bottom sheet over a collection of saved places.", "button-groups": "A button group choosing between walking, cycling and driving routes.", "buttons": "Filled and outlined buttons for saving a coastal route and getting directions.", "cards": "Two cards with landscape paintings, place names and supporting descriptions.", "carousel": "A carousel of places to explore in a travel collection.", "checkbox": "A packing checklist with checked and unchecked items.", "chips": "Filter and assist chips above a collection of walking destinations.", "date-pickers": "A date picker over a calendar of scheduled events.", "dialogs": "A dialog over a collection of saved discoveries.", "divider": "Dividers separating personal and shared collections.", "extended-fab": "An extended floating action button for creating a collection.", "fab-menu": "A floating action button menu for creating documents in a library.", "floating-action-button": "A floating action button for adding to a collection.", "icon-buttons": "Icon buttons for managing a collection of favorite places.", "lists": "A reading list with descriptions and item counts alongside a landscape painting.", "loading-indicator": "A loading indicator while a library is being retrieved.", "menus": "An open menu beside a page of weekend notes.", "navigation-bar": "A horizontal navigation bar anchored below a recently played music library.", "navigation-drawer": "A navigation drawer beside an inbox of messages.", "navigation-rail": "A navigation rail beside a course overview and lesson list.", "progress-indicators": "A progress indicator showing a photo download.", "radio-button": "Radio buttons choosing who can view a collection.", "search": "A search interface with suggestions for saved places.", "segmented-buttons": "Segmented buttons switching between day, week and month calendar views.", "side-sheets": "A docked side sheet displaying metadata beside a landscape photo.", "sliders": "A volume slider beside album artwork and playback controls.", "snackbar": "A snackbar confirming an action in a travel journal.", "split-button": "A split button offering send options in a message composer.", "switch": "On and off switches for notification, download and playback preferences.", "tabs": "Tabs above a collection of outdoor destinations.", "text-fields": "A text field for naming a new collection.", "time-pickers": "A time picker over a calendar of scheduled events.", "toolbars": "A floating toolbar over a travel journal entry.", "tooltips": "A tooltip explaining an action in a notes workspace.", "elevation": "Overlapping filled and elevated cards with a floating action button.", "motion/overview": "A motion curve and directional control illustrating movement between states.", "shape/corner-radius": "Landscape cards and icon surfaces with contrasting corner shapes.", "spacing": "Three content cards demonstrating 8, 16 and 24dp outer spacing.", "typography/type-scale": "Display, headline, body and button typography in a travel interface.",
	}
	return descriptions[subject.Key.Slug]
}

// UseIllustrations pins a completed image generation for the serving process.
func (server *Server) UseIllustrations() error {
	if server.tokenAssetsDir == "" {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(server.tokenAssetsDir, "illustrations", "manifest.json"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var manifest IllustrationManifest
	if err = json.Unmarshal(data, &manifest); err != nil {
		return err
	}
	if manifest.PublicationID != server.state.PublicationID {
		return fmt.Errorf("illustrations belong to another publication")
	}
	server.illustrations = map[string]Illustration{}
	for _, image := range manifest.Images {
		for _, variant := range []string{"light-default", "light-medium", "light-high", "dark-default", "dark-medium", "dark-high"} {
			if image.Files[variant] == "" {
				return fmt.Errorf("missing illustration appearance: %s %s", image.Path, variant)
			}
		}
		allFiles := make([]string, 0, len(image.Files)+len(image.Thumbnails))
		for _, file := range image.Files {
			allFiles = append(allFiles, file)
		}
		if len(image.Thumbnails) > 0 {
			for _, variant := range []string{"light-default", "light-medium", "light-high", "dark-default", "dark-medium", "dark-high"} {
				if image.Thumbnails[variant] == "" {
					return fmt.Errorf("missing thumbnail appearance: %s %s", image.Path, variant)
				}
			}
		}
		for _, file := range image.Thumbnails {
			allFiles = append(allFiles, file)
		}
		for _, file := range allFiles {
			if filepath.Base(file) != file || !illustrationFilename.MatchString(file) {
				return fmt.Errorf("invalid illustration filename")
			}
			if _, err := os.Stat(filepath.Join(server.tokenAssetsDir, "illustrations", file)); err != nil {
				return err
			}
		}
		server.illustrations[image.Path] = image
	}
	return nil
}

var illustrationFilename = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*-[0-9a-f]{16}\.jpg$`)

func (server *Server) handleIllustrationFile(w http.ResponseWriter, r *http.Request) {
	file := r.PathValue("file")
	if server.tokenAssetsDir == "" || !illustrationFilename.MatchString(file) {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeFile(w, r, filepath.Join(server.tokenAssetsDir, "illustrations", file))
}

func (server *Server) handleIllustrationRender(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Robots-Tag", "noindex")
	path := "/" + strings.Trim(r.PathValue("path"), "/")
	theme, contrast := r.URL.Query().Get("theme"), r.URL.Query().Get("contrast")
	if theme != "dark" {
		theme = "light"
	}
	if contrast != "medium" && contrast != "high" {
		contrast = "default"
	}
	var subject publish.SubjectView
	found := path == "/" || path == "/components" || path == "/styles"
	for _, item := range sortedSubjects(server.state) {
		if subjectURL(item) == path {
			subject = item
			found = true
			break
		}
	}
	if !found {
		http.NotFound(w, r)
		return
	}
	var specimen componentCase
	if subject.Key.Kind == publish.SubjectComponent {
		cases, err := server.componentThumbnails([]publish.SubjectView{subject})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		specimen = cases[subject.Key.Slug]
		specimen = scenePreset(specimen)
	}
	if r.URL.Query().Get("composition") == "thumbnail" && subject.Key.Kind == publish.SubjectComponent {
		cases, err := server.componentThumbnails([]publish.SubjectView{subject})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		specimen = cases[subject.Key.Slug]
		specimen.instanceID = "thumbnail-main"
	}
	if subject.Key.Slug == "" {
		specimen = componentCase{Family: "overview", Control: path, Variant: "filled", Size: "medium", Shape: "round", State: "enabled"}
	}
	// Use the same asset ownership and exact publication theme as normal pages.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := server.renderHTML(w, r, illustrationPage(subject, specimen, theme, contrast, r.URL.Query().Get("composition") == "thumbnail", r.URL.Query().Get("view") == "review")); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func illustrationFileURL(file string) string { return "/illustrations/images/" + url.PathEscape(file) }

func (server *Server) setIllustrationMetadata(info *pageMetadata, path string) {
	image, ok := server.illustrations[path]
	if !ok {
		return
	}
	info.Illustration = &image
	if server.publicOrigin != "" {
		info.ImagePageURL = server.publicOrigin + path
		info.ImageURL = server.publicOrigin + illustrationFileURL(image.Files["light-default"])
		info.ImageAlt = image.Alt
	}
}

func (c componentCase) sceneTitle() string {
	titles := map[string]string{"appbar": "Courses", "badge": "Messages", "bottom-sheet": "Riverside Museum", "button-group": "Now playing", "button": "Coastal study", "card": "Exhibitions", "carousel": "Discover", "checkbox": "Exhibition opening", "chip": "Collection", "date": "Edit event", "dialog": "Exhibition 2026", "divider": "Account settings", "extended-fab": "Inbox", "fab-menu": "Exhibition 2026", "fab": "Notes", "icon": "Valley study.jpg", "list": "Downloads", "loading": "Maya Chen", "menu": "Exhibition installation plan", "nav-bar": "Library", "nav-drawer": "Inbox", "nav-rail": "My courses", "progress": "Transfers", "radio": "Delivery", "search": "Search collection", "segmented": "Calendar", "side-sheet": "Photo library", "slider": "Edit photo", "snackbar": "Inbox", "split": "New message", "switch": "Notifications", "tabs": "Orders", "field": "Shipping address", "time": "Edit alarm", "toolbar": "Harbour study", "tooltip": "Exhibition budget"}
	if title := titles[c.Family]; title != "" {
		return title
	}
	return "Make room for good ideas"
}

func (c componentCase) sceneInstance(index int) componentCase {
	c.instanceID = fmt.Sprintf("illustration-choice-%d", index)
	return c
}
func (c componentCase) sceneLabel(label string) componentCase {
	c.illustrationLabels = [3]string{label}
	return c
}

// Apply the saved appearance while parsing the picture, before its first paint.
func illustrationImageTheme() templ.Component {
	return templ.Raw(`<script>(()=>{const root=document.documentElement;const variant=root.dataset.theme+'-'+root.dataset.contrast;for(const source of document.currentScript.previousElementSibling.querySelectorAll('source[data-image-appearance]'))source.media=source.dataset.imageAppearance===variant?'all':'not all';})();</script>`)
}

func (c componentCase) labelsOr(fallback [3]string) [3]string {
	for i, label := range c.illustrationLabels {
		if label != "" {
			fallback[i] = label
		}
	}
	return fallback
}

func thumbnailIllustration(ctx context.Context, path string) (Illustration, bool) {
	assets, ok := ctx.Value(componentAssetsKey{}).(*componentAssets)
	if !ok || assets.server == nil {
		return Illustration{}, false
	}
	image, ok := assets.server.illustrations[path]
	if !ok || len(image.Thumbnails) != 6 {
		return Illustration{}, false
	}
	image.Files = image.Thumbnails
	return image, true
}
