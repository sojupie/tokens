// Playground handlers load immutable coverage once, then delegate pure option selection to the playground package.
package server

import (
	"encoding/json"
	"github.com/sojupie/tokens/catalog/web/internal/playground"
	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	"github.com/sojupie/tokens/source/publish"
	"net/url"
	"os"
	"path/filepath"
)

func (server *Server) buildComponentPreview(subject publish.SubjectView, query url.Values) (*webview.ComponentPreview, error) {
	if server.tokenAssetsDir == "" || subject.Key.Kind != publish.SubjectComponent {
		return nil, nil
	}
	family := playground.ComponentFamily(subject.Key.Slug)
	if family == "" {
		return nil, nil
	}
	file := "components-coverage.json"
	if family == "button" {
		file = "buttons-coverage.json"
	}
	configurations, err := server.previewConfigurations(file)
	if err != nil {
		return nil, err
	}
	var cases []webview.ComponentCase
	for _, c := range configurations {
		if family == "button" {
			c.Family = "button"
		}
		if c.Family == family || family == "field" && (c.Family == "select" || c.Family == "autocomplete") {
			cases = append(cases, c)
		}
	}
	if len(cases) == 0 {
		return nil, nil
	}
	return playground.SelectComponentPreview(cases, query)
}

// Coverage is immutable for this server's loaded publication. Decode each file
// once, rather than reparsing megabytes for every card and every keystroke.
func (server *Server) previewConfigurations(file string) ([]webview.ComponentCase, error) {
	server.previewMu.Lock()
	defer server.previewMu.Unlock()
	if cases, ok := server.previewFiles[file]; ok {
		return cases, nil
	}
	data, err := os.ReadFile(filepath.Join(server.tokenAssetsDir, file))
	if err != nil {
		return nil, err
	}
	var coverage struct {
		Configurations []webview.ComponentCase `json:"configurations"`
	}
	if err := json.Unmarshal(data, &coverage); err != nil {
		return nil, err
	}
	if server.previewFiles == nil {
		server.previewFiles = make(map[string][]webview.ComponentCase)
	}
	server.previewFiles[file] = coverage.Configurations
	return coverage.Configurations, nil
}

// Fixed specimen choices keep catalog cards stable when upstream adds variants.
func (server *Server) componentThumbnails(subjects []publish.SubjectView) (map[string]webview.ComponentCase, error) {
	server.thumbnailMu.Lock()
	defer server.thumbnailMu.Unlock()
	if server.thumbnails == nil {
		server.thumbnails = make(map[string]webview.ComponentCase)
	}
	result := map[string]webview.ComponentCase{}
	for _, subject := range subjects {
		if subject.Key.Kind != publish.SubjectComponent {
			continue
		}
		if specimen, ok := server.thumbnails[subject.Key.Slug]; ok {
			result[subject.Key.Slug] = specimen
			continue
		}
		query := url.Values{"preview.compare": {"interactive"}, "preview.size": {"medium"}}
		switch subject.Key.Slug {
		case "button-groups":
			query.Set("preview.control", "connected")
			query.Set("preview.variant", "filled")
			query.Set("preview.size", "small")
			query.Set("preview.shape", "single")
			query.Set("preview.selection", "selected")
		case "segmented-buttons":
			query.Set("preview.control", "label")
			query.Set("preview.variant", "single")
			query.Set("preview.selection", "selected")
		case "sliders":
			query.Set("preview.size", "small")
		case "buttons":
			query.Set("preview.variant", "filled")
		case "split-button":
			query.Set("preview.variant", "filled")
			query.Set("preview.size", "small")
		case "carousel":
			query.Set("preview.content", "overlay")
		case "cards":
			query.Set("preview.variant", "filled")
		case "icon-buttons":
			query.Set("preview.variant", "tonal")
			query.Set("preview.selection", "selected")
		case "dialogs":
			query.Set("preview.compare", "states")
			query.Set("preview.variant", "basic")
		case "menus":
			query.Set("preview.compare", "states")
			query.Set("preview.variant", "standard")
			query.Set("preview.selection", "selected")
		case "floating-action-button", "extended-fab":
			query.Set("preview.variant", "primary")
			query.Set("preview.size", "default")
		case "navigation-bar", "navigation-rail", "navigation-drawer":
			query.Del("preview.size")
			query.Set("preview.selection", "selected")
		case "snackbar":
			query.Set("preview.compare", "states")
		case "tooltips":
			query.Set("preview.compare", "states")
			query.Set("preview.variant", "rich")
		case "time-pickers":
			query.Set("preview.compare", "states")
			query.Set("preview.variant", "clock")
			query.Set("preview.orientation", "landscape")
			query.Set("preview.control", "value")
		case "date-pickers":
			query.Set("preview.compare", "states")
			query.Set("preview.variant", "docked")
			query.Set("preview.control", "day")
			query.Set("preview.selection", "selected")
		case "lists":
			query.Set("preview.size", "two-line")
			query.Set("preview.shape", "expressive")
			query.Set("preview.variant", "segmented")
		case "fab-menu":
			query.Set("preview.compare", "states")
			query.Set("preview.variant", "primary")
		case "bottom-sheets", "side-sheets":
			query.Set("preview.compare", "states")
			query.Set("preview.control", "standard")
		case "app-bars":
			query.Set("preview.compare", "states")
			query.Set("preview.variant", "small")
		case "toolbars":
			query.Set("preview.variant", "floating")
			query.Set("preview.control", "vibrant")
			query.Set("preview.size", "small")
			query.Set("preview.shape", "icon")
			query.Set("preview.selection", "unselected")
		case "loading-indicator":
			query.Del("preview.size")
			query.Set("preview.variant", "contained")
			query.Set("preview.compare", "states")
		case "progress-indicators":
			query.Set("preview.size", "normal")
			query.Set("preview.variant", "circular")
			query.Set("preview.shape", "wave")
			query.Set("preview.compare", "states")
		case "badges":
			query.Set("preview.variant", "labeled")
		case "checkbox":
			query.Set("preview.selection", "selected")
		case "switch":
			query.Set("preview.variant", "icon")
			query.Set("preview.selection", "selected")
		case "chips":
			query.Set("preview.variant", "filter")
			query.Set("preview.selection", "selected")
		case "radio-button":
			query.Set("preview.compare", "states")
			query.Set("preview.selection", "selected")
		}
		preview, err := server.buildComponentPreview(subject, query)
		if err != nil {
			return nil, err
		}
		if preview != nil && len(preview.Cases) > 0 {
			specimen := preview.Cases[0]
			specimen.State = "enabled"
			server.thumbnails[subject.Key.Slug] = specimen
			result[subject.Key.Slug] = specimen
		}
	}
	return result, nil
}
