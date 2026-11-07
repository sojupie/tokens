package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	webrender "github.com/sojupie/tokens/catalog/web/internal/render"
	"github.com/sojupie/tokens/source/publish"
)

const IllustrationWidth, IllustrationHeight = 1200, 630

type IllustrationManifest struct {
	PublicationID string                 `json:"publicationId"`
	Images        []webview.Illustration `json:"images"`
}

func (server *Server) IllustrationJobs() IllustrationManifest {
	result := IllustrationManifest{PublicationID: server.state.PublicationID}
	for _, subject := range sortedSubjects(server.state) {
		result.Images = append(result.Images, webview.Illustration{Path: webview.SubjectURL(subject), Name: webview.SubjectName(subject), Alt: illustrationAlt(subject)})
	}

	return result
}

func illustrationAlt(subject publish.SubjectView) string {
	return webview.SubjectName(subject) + " preview."
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
	server.illustrations = map[string]webview.Illustration{}
	for _, image := range manifest.Images {
		for _, data := range image.Placeholders {
			if _, err := base64.StdEncoding.DecodeString(data); err != nil {
				return fmt.Errorf("invalid thumbnail placeholder: %w", err)
			}
		}
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
		for _, file := range image.SmallThumbnails {
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

// Match the capture filename so retained assets from older releases also opt out.
var smallThumbnailFilename = regexp.MustCompile(`-thumbnail-360-(light|dark)-(default|medium|high)-[0-9a-f]{16}\.jpg$`)

func (server *Server) handleIllustrationFile(w http.ResponseWriter, r *http.Request) {
	file := r.PathValue("file")
	if server.tokenAssetsDir == "" || !illustrationFilename.MatchString(file) {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	if smallThumbnailFilename.MatchString(file) {
		w.Header().Set("X-Robots-Tag", "noindex")
	}
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
		if webview.SubjectURL(item) == path {
			subject = item
			found = true
			break
		}
	}
	if !found {
		http.NotFound(w, r)
		return
	}
	var specimen webview.ComponentCase
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
	specimen.InstanceID = "thumbnail-main"

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := server.renderHTML(w, r, webrender.IllustrationPage(subject, specimen, theme, contrast)); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (server *Server) setIllustrationMetadata(info *webview.PageMetadata, path string) {
	image, ok := server.illustrations[path]
	if !ok {
		return
	}
	file := image.Files["light-default"]
	width, height := IllustrationWidth, IllustrationHeight
	if file != "" {
		info.Illustration = &image
	} else {
		file = image.Thumbnails["light-default"]
		width, height = 720, 450
	}
	if server.publicOrigin == "" || file == "" {
		return
	}
	info.ImagePageURL = server.publicOrigin + path
	info.ImageURL = server.publicOrigin + webview.IllustrationFileURL(file)
	info.ImageAlt = image.Alt
	info.ImageWidth, info.ImageHeight = width, height
}

// ThumbnailIllustration resolves a published image for the rendering boundary.
func (server *Server) ThumbnailIllustration(path string) (webview.Illustration, bool) {
	image, ok := server.illustrations[path]
	if !ok || len(image.Thumbnails) != 6 {
		return webview.Illustration{}, false
	}
	image.Files = image.Thumbnails
	return image, true
}
