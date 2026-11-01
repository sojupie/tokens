package web

import (
	"github.com/a-h/templ"
	"github.com/sojupie/tokens/forge/components/ui"
	"net/http"
	"strconv"
	"time"
)

type applicationView struct{ ID, Name, Components string }
type applicationState struct {
	applicationView
	Item        int
	Filter      string
	Size, Frame int
}
type applicationScene struct {
	ID, Name, Components string
	Width, Height        int
	Views                []applicationView
}

var applicationScenes = []applicationScene{
	{"mail", "Mail and calendar", "Navigation rail, badges, extended FAB, split button, snackbar, segmented buttons, date picker, time picker", 1440, 1000, []applicationView{
		{"inbox", "Inbox", "Navigation rail · Badges · Extended FAB"}, {"compose", "Compose", "Split button · Text fields"}, {"archived", "Archived conversation", "Snackbar"}, {"calendar", "Calendar", "Segmented buttons"}, {"date", "Event date", "Date picker"}, {"time", "Event time", "Time picker"}}},
	{"learning", "Learning platform", "App bar, cards, navigation rail, progress indicator, loading indicator", 1440, 1000, []applicationView{{"courses", "Courses", "App bar · Cards · Navigation rail"}, {"lesson", "Lesson", "Progress indicator · Supporting pane"}, {"loading", "Loading lesson", "Loading indicator"}}},
	{"artwork", "Artwork library and editor", "Search, chips, side sheet, sliders, toolbar, icon buttons, tooltips", 1440, 1000, []applicationView{{"library", "Library", "Search · Chips"}, {"details", "Image details", "Side sheet · Icon buttons"}, {"editor", "Editor", "Sliders · Toolbar · Tooltip"}}},
	{"workspace", "Project workspace", "Navigation drawer, checkboxes, lists, tabs, menu, dialog, switches, dividers", 1440, 1000, []applicationView{{"tasks", "Tasks", "Checkboxes · Lists · Tabs"}, {"document", "Document", "Menu · Supporting pane"}, {"delete", "Delete file", "Dialog"}, {"settings", "Preferences", "Switches · Dividers"}}},
	{"shop", "Print shop", "Buttons, radio buttons, text fields", 1440, 1000, []applicationView{{"product", "Product", "Buttons"}, {"checkout", "Checkout", "Radio buttons · Text fields"}}},
	{"museum", "Museum companion", "Carousel, bottom sheet, navigation bar, FAB, FAB menu, button group", 412, 892, []applicationView{{"discover", "Discover", "Carousel · Navigation bar"}, {"visit", "Visit details", "Bottom sheet"}, {"planner", "Visit planner", "FAB · FAB menu"}, {"audio", "Audio guide", "Button group"}}},
}

func (a applicationScene) URL(view string) string { return "/illustrations/apps/" + a.ID + "/" + view }
func findApplication(id string) (applicationScene, bool) {
	for _, a := range applicationScenes {
		if a.ID == id {
			return a, true
		}
	}
	return applicationScene{}, false
}
func (a applicationScene) View(id string) (applicationView, bool) {
	for _, v := range a.Views {
		if v.ID == id {
			return v, true
		}
	}
	return applicationView{}, false
}
func (server *Server) handleApplicationSceneReview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Robots-Tag", "noindex")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := server.renderHTML(w, r, applicationReviewPage(applicationScenes)); err != nil {
		http.Error(w, err.Error(), 500)
	}
}
func (server *Server) handleApplicationScene(w http.ResponseWriter, r *http.Request) {
	a, ok := findApplication(r.PathValue("app"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	v, ok := a.View(r.PathValue("view"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	selection := applicationState{applicationView: v}
	if a.ID == "mail" {
		selection.Item = 1
	}
	if raw := r.URL.Query().Get("item"); raw != "" {
		n, err := strconv.Atoi(raw)
		limit := 8
		if a.ID == "artwork" {
			limit = 9
		}
		if err != nil || n < 0 || n >= limit {
			http.Error(w, "Unknown scene item", http.StatusBadRequest)
			return
		}
		selection.Item = n
	}
	selection.Size = 1
	selection.Frame = 1
	for _, choice := range []struct {
		name   string
		target *int
		max    int
	}{{"size", &selection.Size, 2}, {"frame", &selection.Frame, 3}} {
		if raw := r.URL.Query().Get(choice.name); raw != "" {
			n, err := strconv.Atoi(raw)
			if err != nil || n < 1 || n > choice.max {
				http.Error(w, "Unknown product option", 400)
				return
			}
			*choice.target = n
		}
	}
	selection.Filter = r.URL.Query().Get("filter")
	if selection.Filter != "" && selection.Filter != "saved" && selection.Filter != "favorites" {
		http.Error(w, "Unknown scene filter", http.StatusBadRequest)
		return
	}
	if raw := r.URL.Query().Get("width"); raw != "" {
		width, err := strconv.Atoi(raw)
		if err != nil || width < 320 || width > 1920 {
			http.Error(w, "Scene width must be between 320 and 1920", http.StatusBadRequest)
			return
		}
		a.Width = width
	}
	w.Header().Set("X-Robots-Tag", "noindex")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := server.renderHTML(w, r, applicationScenePage(a, selection)); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func applicationPainting(app string, index int) string {
	paintings := []string{"argenteuil-1970.17.42.webp", "autumn---on-the-hudson-river-1963.9.1.webp", "beach-at-trouville-1970.17.12.webp", "blue-morning-1963.10.82.webp", "cliffs-at-pourville-1985.64.27.webp", "entrance-to-the-grand-canal-from-the-molo-venice-1945.15.4.webp", "fanciful-landscape-1963.9.2.webp", "giant-magnolias-on-a-blue-velvet-cloth-1996.14.1.webp", "green-river-cliffs-wyoming-2011.2.1.webp", "italian-coast-scene-with-ruined-tower-1993.55.1.webp", "lake-lucerne-1990.50.1.webp", "landscape-2014.136.109.webp", "landscape-distant-village-2015.19.10.webp", "le-havre-2015.19.58.webp", "low-tide-at-scheveningen-2025.11.7.webp", "nymphenburg-palace-munich-1961.9.63.webp", "palmer-river-2021.11.1.webp", "pastoral-landscape-1991.96.1.webp", "peaceful-valley-1965.10.1.webp", "place-du-carrousel-paris-1970.17.55.webp", "sunset-at-scheveningen-a-fleet-of-fishing-vessels-at-anchor-2012.48.1.webp", "swiss-landscape-2003.41.1.webp", "the-fortress-of-konigstein-1993.8.1.webp", "the-grand-canal-2001.87.1.webp", "the-island-and-bridge-of-san-bartolomeo-rome-2001.23.1.webp", "the-juniata-evening-2010.107.1.webp", "the-louvre-afternoon-rainy-weather-2014.136.59.webp", "the-maas-at-dordrecht-1940.2.1.webp", "the-roman-theater-at-taormina-2004.166.33.webp", "the-voyage-of-life-youth-1971.16.2.webp", "washerwomen-on-the-beach-of-etretat-1970.17.17.webp", "yarmouth-jetty-2016.115.1.webp"}
	offset := map[string]int{"learning": 0, "artwork": 8, "shop": 18, "museum": 20}[app]
	return paintings[(offset+index)%len(paintings)]
}

func applicationInboxRows(archived bool, selected int) [][4]string {
	rows := sceneInboxRows(false)
	if !archived {
		return rows
	}
	return append(rows[:selected:selected], rows[selected+1:]...)
}
func applicationUnread(name string) string {
	return map[string]string{"Maya Chen": "2", "Alex Morgan": "5", "Jamie Park": "1"}[name]
}

func applicationNavigationCase(app string, rail bool) ui.NavigationProps {
	p := ui.NavigationProps{ID: app + "-navigation", Label: "Application destinations", Kind: "drawer", Variant: "standard", Layout: "standard"}
	if rail {
		p.Kind = "rail"
		p.Variant = "collapsed"
		p.Orientation = "vertical"
	}
	if app == "mail" {
		p.Variant = "expanded"
		p.Orientation = "horizontal"
		p.Indicator = "fill"
		p.Width = "narrow"
	}
	return p
}
func applicationNavigationItem(c ui.NavigationProps, index, selected int) templ.Attributes {
	return (ui.NavigationItemProps{Current: index == selected}).Attributes()
}
func applicationComposeAttributes(c ui.NavigationProps) templ.Attributes {
	return templ.Attributes{"class": "md-button md-extended-fab md-nav-fab", "data-variant": "primary-container", "data-size": "default", "aria-label": "Compose"}
}
func applicationMailLabel(row [4]string) string {
	label := row[1] + ". " + row[2]
	if count := applicationUnread(row[1]); count != "" {
		label += ". " + count + " unread message"
		if count != "1" {
			label += "s"
		}
	}
	return label
}

var applicationArtworkTitles = []string{"Green River Cliffs, Wyoming", "Italian Coast Scene with Ruined Tower", "Lake Lucerne", "Landscape", "Landscape, Distant Village", "Le Havre", "Low Tide at Scheveningen", "Nymphenburg Palace, Munich", "Palmer River"}
var applicationCourseTitles = []string{"Landscape painting: light and atmosphere", "Color foundations", "Composition in practice", "Drawing from observation", "Water and reflections", "Architectural studies", "Understanding exposure", "A sketchbook practice"}

func applicationLearningDestination(view, filter string) int {
	if filter == "saved" {
		return 1
	}
	if view != "courses" {
		return 2
	}
	return 0
}
func applicationArtworkDestination(view, filter string) int {
	if filter == "favorites" {
		return 1
	}
	if view == "editor" {
		return 2
	}
	return 0
}
func applicationMessageSubject(index int) string {
	return []string{"Workshop schedule", "Print proofs", "Design review", "Your booking is confirmed", "Workshop notes", "Your reservation", "October invoice", "This week"}[index]
}

func applicationPrice(state applicationState) int {
	price := 48
	if state.Size == 2 {
		price += 18
	}
	if state.Frame > 1 {
		price += 20
	}
	return price
}

func applicationMessageIndex(name string) int {
	for index, row := range sceneInboxRows(false) {
		if row[1] == name {
			return index
		}
	}
	return 0
}

func applicationUnreadTotal(state applicationState) string {
	total := 8
	if state.ID == "archived" {
		n, _ := strconv.Atoi(applicationUnread(sceneInboxRows(false)[state.Item][1]))
		total -= n
	}
	return strconv.Itoa(total)
}

func applicationPercent(value float64) *float64 { return &value }
func applicationDate(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
