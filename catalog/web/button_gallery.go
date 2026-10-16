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

type buttonCase struct {
	Variant   string `json:"variant"`
	Size      string `json:"size"`
	Shape     string `json:"shape"`
	Selection string `json:"selection"`
	State     string `json:"state"`
}

func (c buttonCase) ID() string {
	return strings.Join([]string{c.Variant, c.Size, c.Shape, c.Selection, c.State}, "-")
}
func (c buttonCase) Label() string {
	selection := c.Selection
	if selection == "" {
		selection = "plain"
	}
	return fmt.Sprintf("%s · %s · %s · %s · %s", c.Variant, c.Size, c.Shape, selection, c.State)
}
func (c buttonCase) Attributes() templ.Attributes {
	attrs := templ.Attributes{"class": "md-button md-" + c.Variant + "-button", "type": "button", "data-size": c.Size, "data-shape": c.Shape, "data-button-state": c.State, "data-button-case": c.ID()}
	if c.Selection != "" {
		attrs["aria-pressed"] = fmt.Sprint(c.Selection == "selected")
		attrs["data-button-toggle"] = ""
	}
	if c.State == "disabled" {
		attrs["disabled"] = true
	}
	return attrs
}

type buttonGallery struct {
	Cases           []buttonCase
	Filters         map[string]string
	Variants, Sizes []string
	Mapped          int
}

func (server *Server) handleButtonGallery(response http.ResponseWriter, request *http.Request) {
	if server.tokenAssetsDir == "" {
		http.Error(response, "Build generated token assets to view the button gallery.", http.StatusServiceUnavailable)
		return
	}
	data, err := os.ReadFile(filepath.Join(server.tokenAssetsDir, "buttons-coverage.json"))
	if err != nil {
		http.Error(response, err.Error(), 500)
		return
	}
	var coverage struct {
		Configurations []buttonCase      `json:"configurations"`
		Variants       []string          `json:"variants"`
		Sizes          []string          `json:"sizes"`
		Mapped         []json.RawMessage `json:"mapped"`
	}
	if err := json.Unmarshal(data, &coverage); err != nil {
		http.Error(response, err.Error(), 500)
		return
	}
	page := buttonGallery{Variants: coverage.Variants, Sizes: coverage.Sizes, Mapped: len(coverage.Mapped), Filters: map[string]string{"variant": "all", "size": "small", "shape": "round", "selection": "all", "state": "all"}}
	for key := range page.Filters {
		if value := request.URL.Query().Get(key); value != "" {
			page.Filters[key] = value
		}
	}
	for _, c := range coverage.Configurations {
		values := map[string]string{"variant": c.Variant, "size": c.Size, "shape": c.Shape, "selection": c.Selection, "state": c.State}
		if values["selection"] == "" {
			values["selection"] = "plain"
		}
		matches := true
		for key, value := range values {
			if filter := page.Filters[key]; filter != "all" && filter != value {
				matches = false
				break
			}
		}
		if matches {
			page.Cases = append(page.Cases, c)
		}
	}
	response.Header().Set("X-Robots-Tag", "noindex")
	if err := server.renderHTML(response, request, buttonGalleryPage(page)); err != nil {
		http.Error(response, err.Error(), 500)
	}
}
