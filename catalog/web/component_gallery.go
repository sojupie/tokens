package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/a-h/templ"
)

type componentCase struct {
	Family   string `json:"family"`
	Variant  string `json:"variant"`
	Size     string `json:"size"`
	Shape    string `json:"shape"`
	Width    string `json:"width"`
	State    string `json:"state"`
	Selected bool   `json:"selected"`
	Elevated bool   `json:"elevated"`
	Invalid  bool   `json:"invalid"`
}

func (c componentCase) ID() string {
	return fmt.Sprintf("%s-%s-%s-%s-%s-%s-%t-%t-%t", c.Family, c.Variant, c.Size, c.Shape, c.Width, c.State, c.Selected, c.Elevated, c.Invalid)
}
func (c componentCase) Label() string {
	parts := []string{c.Family}
	if c.Variant != "" {
		parts = append(parts, c.Variant)
	}
	if c.Size != "" {
		parts = append(parts, c.Size, c.Shape, c.Width+" width")
	}
	if c.Family == "chip" {
		if c.Elevated {
			parts = append(parts, "elevated")
		} else {
			parts = append(parts, "flat")
		}
	}
	if c.Family == "icon" || c.Family == "chip" || c.Family == "radio" || c.Family == "menu" {
		if c.Selected {
			parts = append(parts, "selected")
		} else {
			parts = append(parts, "unselected")
		}
	}
	if c.Invalid {
		parts = append(parts, "error")
	}
	return strings.Join(append(parts, c.State), " · ")
}

func (c componentCase) Attributes() templ.Attributes {
	class := map[string]string{"card": "md-card md-" + c.Variant + "-card md-interactive-card", "icon": "md-icon-button md-" + c.Variant + "-icon-button", "chip": "md-filter-chip", "radio": "md-radio", "field": "md-outlined-text-field", "select": "md-outlined-select", "menu": "md-menu-item", "dialog": "md-dialog", "divider": "md-divider"}[c.Family]
	attrs := templ.Attributes{"class": class, "data-component-case": c.ID(), "data-family": c.Family, "data-component-state": c.State}
	if c.Size != "" {
		attrs["data-size"] = c.Size
		attrs["data-shape"] = c.Shape
		attrs["data-width"] = c.Width
	}
	if c.Family == "icon" {
		attrs["aria-pressed"] = fmt.Sprint(c.Selected)
		attrs["data-button-toggle"] = ""
		attrs["aria-label"] = c.Label()
		attrs["type"] = "button"
	}
	if c.Family == "chip" {
		attrs["data-elevated"] = fmt.Sprint(c.Elevated)
	}
	if c.Family == "menu" && c.Selected {
		attrs["aria-current"] = "page"
	}
	if c.Family == "radio" {
		attrs["type"] = "radio"
		attrs["name"] = c.ID()
		attrs["checked"] = c.Selected
		attrs["aria-label"] = c.Label()
	}
	if c.Family == "field" || c.Family == "select" {
		attrs["aria-invalid"] = fmt.Sprint(c.Invalid)
	}
	if c.State == "disabled" {
		attrs["aria-disabled"] = "true"
		if c.Family == "icon" || c.Family == "radio" {
			attrs["disabled"] = true
		}
	}
	return attrs
}

type componentGallery struct {
	Cases            []componentCase
	Family           string
	Mapped, Excluded int
}

func (server *Server) handleComponentGallery(response http.ResponseWriter, request *http.Request) {
	if server.tokenAssetsDir == "" {
		http.Error(response, "Build generated token assets to view the component gallery.", 503)
		return
	}
	data, err := os.ReadFile(filepath.Join(server.tokenAssetsDir, "components-coverage.json"))
	if err != nil {
		http.Error(response, err.Error(), 500)
		return
	}
	var coverage struct {
		Configurations   []componentCase `json:"configurations"`
		Mapped, Excluded []json.RawMessage
	}
	if err = json.Unmarshal(data, &coverage); err != nil {
		http.Error(response, err.Error(), 500)
		return
	}
	family := request.URL.Query().Get("family")
	if family == "" {
		family = "card"
	}
	page := componentGallery{Family: family, Mapped: len(coverage.Mapped), Excluded: len(coverage.Excluded)}
	for _, c := range coverage.Configurations {
		if family == "all" || family == c.Family {
			page.Cases = append(page.Cases, c)
		}
	}
	response.Header().Set("X-Robots-Tag", "noindex")
	if err = server.renderHTML(response, request, componentGalleryPage(page)); err != nil {
		http.Error(response, err.Error(), 500)
	}
}
