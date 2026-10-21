package web

import (
	"github.com/a-h/templ"
	"github.com/sojupie/tokens/source/publish"
	"net/url"
)

// Fixed specimen choices keep catalog cards stable when upstream adds variants.
func (server *Server) componentThumbnails(subjects []publish.SubjectView) (map[string]componentCase, error) {
	server.thumbnailMu.Lock()
	defer server.thumbnailMu.Unlock()
	if server.thumbnails == nil {
		server.thumbnails = make(map[string]componentCase)
	}
	result := map[string]componentCase{}
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

// Thumbnail composition chooses variants, never component geometry or token values.
func (c componentCase) thumbnailVariant(variant string) componentCase {
	c.Variant = variant
	return c
}
func (c componentCase) thumbnailSelected(selected bool) componentCase {
	c.Selected = selected
	return c
}

func (c componentCase) thumbnailToolbarAction(index int) templ.Attributes {
	attrs := c.BarAction(min(index, 2))
	attrs["aria-label"] = []string{"Download", "Copy", "More actions"}[index]
	delete(attrs, "aria-pressed")
	delete(attrs, "data-button-toggle")
	delete(attrs, "data-menu-open")
	delete(attrs, "aria-haspopup")
	delete(attrs, "aria-expanded")
	return attrs
}
