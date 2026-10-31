package web

import (
	"fmt"
	"github.com/a-h/templ"
	"github.com/sojupie/tokens/source/publish"
	"net/http"
)

type sceneReviewItem struct {
	Name, Path, Device string
	Width, Height      int
}

func sceneDevice(family string) string {
	switch family {
	case "badge", "bottom-sheet", "button-group", "carousel", "checkbox", "date", "divider", "fab", "list", "loading", "nav-bar", "radio", "snackbar", "switch", "time":
		return "phone"
	case "appbar", "card", "chip", "extended-fab", "fab-menu", "nav-drawer", "search", "segmented", "slider", "tabs", "toolbar":
		return "tablet"
	default:
		return "desktop"
	}
}
func sceneSize(family string) (int, int) {
	switch sceneDevice(family) {
	case "phone":
		return 412, 892
	case "tablet":
		return 840, 1120
	default:
		return 1440, 1000
	}
}
func sceneFrameStyle(c componentCase) templ.SafeCSS {
	w, h := sceneSize(c.Family)
	return templ.SafeCSS(fmt.Sprintf("--scene-width:%dpx;--scene-height:%dpx;--scene-scale:%g", w, h, min(1000/float64(w), 560/float64(h))))
}
func sceneBarCase(family, title string) componentCase {
	variant := "small"
	if family == "appbar" {
		variant = "medium-flexible"
	}
	return componentCase{Family: "appbar", Variant: variant, State: "enabled", Control: "title", instanceID: "scene-bar-" + title}
}
func scenePreset(c componentCase) componentCase {
	c.State = "enabled"
	c.instanceID = "scene-" + c.Family
	switch c.Family {
	case "appbar":
		c.Variant = "medium-flexible"
		c.Control = "title"
	case "time":
		c.Orientation = "portrait"
		c.Width = "12h"
	case "date":
		c.Variant = "modal"
	case "search":
		c.Variant = "docked"
		c.Expanded = false
	case "side-sheet":
		c.Control = "standard"
	case "list":
		c.Size = "two-line"
		c.Control = "none"
	case "button-group":
		c.Control = "standard"
		c.Variant = "tonal"
		c.Shape = "single"
		c.Size = "small"
	case "nav-drawer":
		c.Variant = "standard"
		c.Selected = true
	case "nav-rail":
		c.Variant = "collapsed"
		c.Selected = true
	case "nav-bar":
		c.Variant = "vertical"
		c.Selected = true
	case "progress":
		c.Variant = "linear"
		c.Shape = "flat"
		c.Control = "determinate"
	case "radio":
		c.Selected = true
	case "toolbar":
		c.Variant = "floating"
		c.Orientation = "horizontal"
		c.Shape = "icon"
	case "slider":
		c.SliderDisplay = "interaction"
		c.Variant = "continuous"
		c.Control = "single"
		c.Orientation = "horizontal"
		c.Size = "small"
	case "segmented":
		c.Control = "label"
		c.illustrationLabels = [3]string{"Day", "Week", "Month"}
		c.Selected = false
	case "loading":
		c.Control = "indeterminate"
	case "chip":
		c.Variant = "filter"
		c.Selected = true
		c.illustrationLabels = [3]string{"Landscape"}
	case "tabs":
		c.Selected = true
		c.Control = "label"
		c.illustrationLabels = [3]string{"Active", "Completed", "Cancelled"}
	case "fab-menu":
		c.illustrationLabels = [3]string{"Upload file", "New folder", "New document"}
	case "split":
		c.Shape = "icon-label"
	}
	return c
}
func (server *Server) handleSceneReview(w http.ResponseWriter, r *http.Request) {
	subjects := []publish.SubjectView{}
	for _, subject := range sortedSubjects(server.state) {
		if subject.Key.Kind == publish.SubjectComponent {
			subjects = append(subjects, subject)
		}
	}
	cases, err := server.componentThumbnails(subjects)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	items := []sceneReviewItem{}
	for _, subject := range subjects {
		c := cases[subject.Key.Slug]
		width, height := sceneSize(c.Family)
		items = append(items, sceneReviewItem{Name: subjectName(subject), Path: subjectURL(subject), Device: sceneDevice(c.Family), Width: width, Height: height})
	}
	w.Header().Set("X-Robots-Tag", "noindex")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := server.renderHTML(w, r, sceneReviewPage(items)); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func sceneNamedAttributes(c componentCase, label string) templ.Attributes {
	attrs := c.Attributes()
	attrs["aria-label"] = label
	return attrs
}

func sceneCollectionPainting(family string, index int) string {
	if index < 2 {
		return scenePainting(family, []string{"coast", "hills"}[index])
	}
	if family == "chip" {
		return []string{"landscape-distant-village-2015.19.10.webp", "le-havre-2015.19.58.webp"}[(index-2)%2]
	}
	return []string{"the-island-and-bridge-of-san-bartolomeo-rome-2001.23.1.webp", "the-fortress-of-konigstein-1993.8.1.webp"}[(index-2)%2]
}

func sceneInboxRows(archived bool) [][4]string {
	rows := [][4]string{
		{"M", "Maya Chen", "Workshop schedule — The updated timings are attached.", "9:24"},
		{"A", "Alex Morgan", "Re: Print proofs — The second version looks good.", "9:10"},
		{"J", "Jamie Park", "Design review — Thursday at 9, meeting room 2.", "8:42"},
		{"S", "Studio North", "Your booking is confirmed — 4 November, 10:00.", "Yesterday"},
		{"R", "Robin Ellis", "Workshop notes — Thanks for joining yesterday.", "Yesterday"},
		{"L", "Library services", "Your reservation is ready to collect.", "Tue"},
		{"F", "Finance", "October invoice — Payment received, thank you.", "Tue"},
		{"T", "Team updates", "This week — Three things to know before Monday.", "Mon"},
	}
	if archived {
		return rows[1:]
	}
	return rows
}

func sceneFieldHelp(id, help string) templ.Attributes {
	if help == "" {
		return nil
	}
	return templ.Attributes{"aria-describedby": id + "-help"}
}

// Headers follow the task: app bars belong to app navigation, not every canvas.
func sceneUsesAppBar(family string) bool {
	switch family {
	case "appbar", "bottom-sheet", "date", "divider", "list", "loading", "radio", "switch", "time", "snackbar":
		return true
	}
	return false
}
func sceneHasOwnHeading(family string) bool {
	switch family {
	case "icon", "menu", "tooltip", "toolbar", "button":
		return true
	}
	return sceneUsesAppBar(family)
}
