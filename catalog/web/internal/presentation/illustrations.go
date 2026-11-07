package presentation

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/a-h/templ"
)

// Illustration files are build outputs, independent of the live playground state.
type Illustration struct {
	Path              string            `json:"path"`
	Name              string            `json:"name"`
	Alt               string            `json:"alt"`
	Files             map[string]string `json:"files,omitempty"`
	Thumbnails        map[string]string `json:"thumbnails,omitempty"`
	SmallThumbnails   map[string]string `json:"smallThumbnails,omitempty"`
	Placeholders      map[string]string `json:"placeholders,omitempty"`
	PlaceholderFormat string            `json:"placeholderFormat,omitempty"`
}

func IllustrationFileURL(file string) string { return "/illustrations/images/" + url.PathEscape(file) }

func ThumbnailSrcset(image Illustration, variant string) string {
	full := IllustrationFileURL(image.Files[variant])
	if small := image.SmallThumbnails[variant]; small != "" {
		return IllustrationFileURL(small) + " 360w, " + full + " 720w"
	}
	return full
}

func ThumbnailPlaceholderStyle(image Illustration) string {
	var css strings.Builder
	format := "jpeg"
	if image.PlaceholderFormat == "webp" {
		format = "webp"
		css.WriteString("--_thumbnail-blur:6px;")
	}
	for _, variant := range []string{"light-default", "light-medium", "light-high", "dark-default", "dark-medium", "dark-high"} {
		if data := image.Placeholders[variant]; data != "" {
			fmt.Fprintf(&css, "--_thumbnail-%s:url(data:image/%s;base64,%s);", variant, format, data)
		}
	}
	return css.String()
}

// Apply the saved appearance while parsing the picture, before its first paint.
func IllustrationImageTheme() templ.Component {
	return templ.Raw(`<script>(()=>{const root=document.documentElement;const variant=root.dataset.theme+'-'+root.dataset.contrast;for(const source of document.currentScript.previousElementSibling.querySelectorAll('source[data-image-appearance]'))source.media=source.dataset.imageAppearance===variant?'all':'not all';})();</script>`)
}

func (c ComponentCase) LabelsOr(fallback [3]string) [3]string {
	for i, label := range c.IllustrationLabels {
		if label != "" {
			fallback[i] = label
		}
	}
	return fallback
}

func ThumbnailIllustration(ctx context.Context, path string) (Illustration, bool) {
	assets, ok := ctx.Value(ComponentAssetsKey{}).(*ComponentAssets)
	if !ok || assets.Server == nil {
		return Illustration{}, false
	}
	return assets.Server.ThumbnailIllustration(path)
}
