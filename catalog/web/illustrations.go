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

	return result
}

func illustrationAlt(subject publish.SubjectView) string {
	return subjectName(subject) + " preview."
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
			if len(image.Files) > 0 && image.Files[variant] == "" {
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
	found := false
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
	}
	if r.URL.Query().Get("composition") != "thumbnail" || subject.Key.Slug == "" {
		http.NotFound(w, r)
		return
	}
	specimen.instanceID = "thumbnail-main"
	// Use the same asset ownership and exact publication theme as normal pages.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := server.renderHTML(w, r, illustrationPage(subject, specimen, theme, contrast)); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func illustrationFileURL(file string) string { return "/illustrations/images/" + url.PathEscape(file) }

func (server *Server) setIllustrationMetadata(info *pageMetadata, path string) {
	image, ok := server.illustrations[path]
	if !ok || len(image.Files) == 0 {
		return
	}
	info.Illustration = &image
	if server.publicOrigin != "" {
		info.ImagePageURL = server.publicOrigin + path
		info.ImageURL = server.publicOrigin + illustrationFileURL(image.Files["light-default"])
		info.ImageAlt = image.Alt
	}
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
