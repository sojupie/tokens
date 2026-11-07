// Integration checks for playground.
package server

import (
	"bytes"
	"context"
	"github.com/sojupie/tokens/catalog/web/internal/playground"
	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	webrender "github.com/sojupie/tokens/catalog/web/internal/render"
	"github.com/sojupie/tokens/forge/components/ui"
	"github.com/sojupie/tokens/source/publish"
	"html"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDateRowsAndStates(t *testing.T) {
	rows := webview.DateRows(10)
	if len(rows) != 5 || rows[0][0].ISO != "2026-09-27" || rows[4][6].ISO != "2026-10-31" {
		t.Fatal("October calendar must contain five complete weeks")
	}
	c := webview.ComponentCase{Family: "date", Variant: "modal", DateRange: true, Control: "day", Selection: "in-range", State: "pressed"}
	for _, row := range rows {
		for _, day := range row {
			if day.ISO == "2026-10-23" {
				a := c.DateDay(day)
				if a["aria-selected"] != "true" || a["data-date-selected"] != "false" || a["data-in-range"] != "true" || a["data-component-state"] != "pressed" {
					t.Fatalf("range interior must retain its own palette: %#v", a)
				}
			}
		}
	}
	c.State = "interactive"
	if _, ok := c.DatePanel()["open"]; ok {
		t.Fatal("interactive picker must open through native dialog API")
	}
}

func TestDateEntryFields(t *testing.T) {
	c := webview.ComponentCase{Family: "date", Variant: "docked", State: "interactive"}
	a := c.DateEntry("start")
	if a["data-date-value"] != "start" || a["value"] != "2026-10-21" {
		t.Fatalf("docked entry must be an editable calendar combobox: %#v", a)
	}
	c.Variant = "modal"
	c.DateRange = true
	if c.DateEntryLabel("start") != "Start date" || c.DateEntryLabel("end") != "End date" {
		t.Fatal("range fields must have distinct labels")
	}
	if c.DateEntry("end")["id"] == c.DateEntry("start")["id"] {
		t.Fatal("range fields must have distinct IDs")
	}
	if _, ok := c.DateEntry("start")["role"]; ok {
		t.Fatal("modal entry must remain a text field")
	}
}

func TestDockedDateNavigation(t *testing.T) {
	c := webview.ComponentCase{Family: "date", Variant: "docked", Control: "month"}
	month, year := c.DateMenu(), c.DateYearMenu()
	if month["data-date-menu"] != "month" || year["data-date-menu"] != "year" || month["aria-controls"] == year["aria-controls"] {
		t.Fatal("docked month and year controls must target separate menus")
	}
	if month["aria-expanded"] != "true" || year["aria-expanded"] != "false" {
		t.Fatal("only the active menu is expanded")
	}
	if c.DateYear(2026)["class"] != "md-date-month-item" {
		t.Fatal("docked years must use captured docked menu-list styling")
	}
	c.Variant = "modal"
	if c.DateYear(2026)["class"] != "md-date-year" {
		t.Fatal("modal years must retain their grid styling")
	}
}

func TestDateRangeIsIndependentOfVariant(t *testing.T) {
	for _, variant := range []string{"docked", "modal", "input", "full-screen"} {
		c := webview.ComponentCase{Family: "date", Variant: variant, DateRange: true, Control: "day", State: "interactive"}
		if c.DatePanel()["data-range"] != "true" || webview.PreviewDimensions(c)["mode"] != "range" || c.DateEntryLabel("start") != "Start date" {
			t.Fatalf("range missing for %s", variant)
		}
		single := c
		single.DateRange = false
		if single.ID() == c.ID() {
			t.Fatal("single and range IDs collide")
		}
	}
}

func TestDatePreviewExposesEveryVariantForSingleAndRange(t *testing.T) {
	for _, rangeMode := range []bool{false, true} {
		var cases []webview.ComponentCase
		for _, variant := range []string{"docked", "modal", "input", "full-screen"} {
			for _, r := range []bool{false, true} {
				cases = append(cases, webview.ComponentCase{Family: "date", Variant: variant, DateRange: r, Control: "day", State: "enabled"})
			}
		}
		mode := "single"
		if rangeMode {
			mode = "range"
		}
		page, err := playground.SelectComponentPreview(cases, map[string][]string{"preview.compare": {"variants"}, "preview.mode": {mode}})
		if err != nil || len(page.Cases) != 4 {
			t.Fatalf("%s: %#v %v", mode, page, err)
		}
		for _, c := range page.Cases {
			if c.DateRange != rangeMode {
				t.Fatalf("selection changed for %s", c.Variant)
			}
		}
	}
}

func TestDateEntryPlaceholderMatchesPurpose(t *testing.T) {
	for _, isRange := range []bool{false, true} {
		c := webview.ComponentCase{Family: "date", Variant: "docked", DateRange: isRange}
		for _, part := range []string{"start", "end"} {
			if c.DateEntry(part)["placeholder"] != c.DateEntryLabel(part) {
				t.Fatalf("%s input placeholder differs from its label", part)
			}
			if c.DateEntry(part)["aria-describedby"] != c.ID()+"-value-help-"+part {
				t.Fatal("Date format is not associated with the field")
			}
		}
	}
}

func TestDialogProfilesAndTargets(t *testing.T) {
	c := webview.ComponentCase{Family: "dialog", Variant: "full-screen", Orientation: "bottom", Control: "close", State: "pressed", Width: "middle"}
	if c.DialogPanel()["class"] != "md-full-screen-dialog" || c.DialogPanel()["data-scroll-start"] != "middle" {
		t.Fatal("fullscreen profile missing")
	}
	if c.DialogClose()["data-component-state"] != "pressed" {
		t.Fatal("close preview state missing")
	}
	if _, ok := c.DialogAction(true)["data-component-state"]; ok {
		t.Fatal("untargeted action must retain its native state")
	}
	c.State = "interactive"
	if _, ok := c.DialogPanel()["open"]; ok {
		t.Fatal("interactive dialogs must start closed")
	}
	cases := []webview.ComponentCase{{Family: "dialog", Variant: "basic", Control: "action", Shape: "plain", State: "enabled"}, {Family: "dialog", Variant: "full-screen", Control: "action", Orientation: "header", Shape: "plain", Width: "top", State: "enabled"}}
	p, err := playground.SelectComponentPreview(cases, map[string][]string{"preview.compare": {"variants"}})
	if err != nil || len(p.Cases) != 2 {
		t.Fatalf("dialog variants must both remain in comparison: %v %v", p, err)
	}
}

func TestDialogOptionalHeadlineAndDecorations(t *testing.T) {
	for _, variant := range []string{"basic", "full-screen"} {
		c := webview.ComponentCase{Family: "dialog", Variant: variant, Shape: "icon-divided", DialogHeadline: "hidden"}
		a := c.DialogPanel()
		if a["aria-label"] != c.DialogHeading() {
			t.Fatal("headline-free dialogs need a name")
		}
		if _, ok := a["aria-labelledby"]; ok {
			t.Fatal("missing headline referenced")
		}
		if !c.DialogIcon() || !c.DialogDividers() {
			t.Fatal("icon and dividers must compose")
		}
		shown := c
		shown.DialogHeadline = "shown"
		if shown.ID() == c.ID() {
			t.Fatal("headline configurations collide")
		}
	}
}

func TestDialogVariantResetsUnsupportedControls(t *testing.T) {
	cases := []webview.ComponentCase{
		{Family: "dialog", Variant: "basic", Control: "action", Shape: "plain", State: "enabled"},
		{Family: "dialog", Variant: "full-screen", Control: "close", Shape: "plain", State: "enabled"},
	}
	p, err := playground.SelectComponentPreview(cases, map[string][]string{"preview.compare": {"states"}, "preview.variant": {"basic"}, "preview.control": {"close"}})
	if err != nil || len(p.Cases) != 1 || p.Cases[0].Variant != "basic" || p.Cases[0].Control != "action" {
		t.Fatalf("stale fullscreen control changed the selected variant: %#v %v", p, err)
	}
	for _, control := range p.Controls {
		if control.Name == "control" {
			t.Fatal("basic dialog offers an absent close control")
		}
	}
}

func TestListSemanticsFollowBehavior(t *testing.T) {
	c := webview.ComponentCase{Family: "list", Control: "select", Selection: "single", State: "disabled"}
	if c.ListItem(0)["aria-selected"] != "false" || c.ListItem(1)["aria-selected"] != "true" || c.ListItem(0)["disabled"] != true {
		t.Fatal("single selection and disabled state are not reflected in native controls")
	}
	c.Control = "expand"
	c.Expanded = true
	if c.ListItem(0)["aria-expanded"] != "true" || c.ListItem(1)["aria-expanded"] != "false" {
		t.Fatal("only the first disclosure should start expanded")
	}
	c.Control = "reveal"
	if c.ListUtility(0, true)["class"] != "md-icon-button md-list-reveal-action" {
		t.Fatal("reveal actions must use their own token profile")
	}
}

func TestListPreviewCompositions(t *testing.T) {
	c := webview.ComponentCase{Family: "list", Control: "select", Variant: "standard", Width: "video", Size: "three-line", State: "enabled"}
	page := &webview.ComponentPreview{Cases: []webview.ComponentCase{c}}
	if err := playground.ConfigureListPreview(page, map[string][]string{"preview.media": {"large"}, "preview.trailing": {"checkbox"}, "preview.divider": {"none"}, "preview.text": {"wrapped"}}); err != nil {
		t.Fatal(err)
	}
	got := page.Cases[0]
	if got.ListMediaKind() != "large" || got.ListDividerKind() != "none" || got.ListNativeKind() != "checkbox" {
		t.Fatalf("composition choices lost: %+v", got)
	}
	a := got.ListItem(0)
	if _, ok := a["type"]; ok {
		t.Fatal("native controls must not be nested inside a button")
	}
	if _, ok := a["data-button-toggle"]; ok {
		t.Fatal("native controls must own selection")
	}
	if got.ListControlInput(0)["type"] != "checkbox" {
		t.Fatal("checkbox input missing")
	}
	got.Width = "radio"
	got.ListTrailing = "text"
	if !got.ListSelected(1) || got.ListSelected(0) {
		t.Fatal("radio composition must have one selected row")
	}
	if err := playground.ConfigureListPreview(&webview.ComponentPreview{Cases: []webview.ComponentCase{c}}, map[string][]string{"preview.media": {"unknown"}}); err == nil {
		t.Fatal("invalid media size accepted")
	}
}

func TestListBehaviorCompositions(t *testing.T) {
	for _, mode := range []string{"single", "multiple"} {
		page := &webview.ComponentPreview{Cases: []webview.ComponentCase{{Family: "list", Control: "select", Selection: mode}}}
		if err := playground.ConfigureListPreview(page, nil); err != nil {
			t.Fatal(err)
		}
		for _, control := range page.Controls {
			if control.Name == "trailing" {
				for _, value := range control.Options {
					if mode == "single" && (value == "checkbox" || value == "switch") || mode == "multiple" && value == "radio" {
						t.Fatalf("invalid %s selection control: %s", mode, value)
					}
				}
			}
		}
	}
	page := &webview.ComponentPreview{Cases: []webview.ComponentCase{{Family: "list", Control: "reveal"}}}
	if err := playground.ConfigureListPreview(page, map[string][]string{"preview.trailing": {"icon"}}); err != nil {
		t.Fatal(err)
	}
	if page.Cases[0].ListTrailing == "icon" {
		t.Fatal("reveal must not duplicate its utility action with a navigation chevron")
	}
	page = &webview.ComponentPreview{Cases: []webview.ComponentCase{{Family: "list", Control: "expand"}}}
	if err := playground.ConfigureListPreview(page, map[string][]string{"preview.inset": {"inset"}, "preview.surface": {"none"}, "preview.expand-icon": {"plain"}}); err != nil {
		t.Fatal(err)
	}
	if page.Cases[0].ListInset != "inset" || page.Cases[0].ListSurface != "none" || page.Cases[0].ListExpandIcon != "plain" {
		t.Fatal("independent expansion options lost")
	}
	if _, ok := (webview.ComponentCase{Control: "reorder"}).ListItem(0)["aria-pressed"]; ok {
		t.Fatal("reordering must not toggle persistent row selection")
	}
}

func TestListSelectionOwnsFocusAndNames(t *testing.T) {
	for _, mode := range []string{"single", "multiple"} {
		c := webview.ComponentCase{Family: "list", Control: "select", Selection: mode, Width: "checkbox", State: "interactive"}
		a := c.ListContainerAttributes()
		if a["role"] != "listbox" || a["aria-multiselectable"] != "false" && mode == "single" || a["aria-multiselectable"] != "true" && mode == "multiple" {
			t.Fatalf("wrong listbox mode: %v", a)
		}
		for i := range 3 {
			item, input := c.ListItem(i), c.ListControlInput(i)
			if item["role"] != "option" || item["aria-label"] != c.ListLabel(i) || item["aria-selected"] == nil {
				t.Fatalf("option semantics missing: %v", item)
			}
			if input["aria-hidden"] != "true" || input["tabindex"] != "-1" {
				t.Fatalf("duplicate nested control: %v", input)
			}
		}
	}
	c := webview.ComponentCase{Family: "list", Control: "expand"}
	if c.ListContainerAttributes()["role"] == "listbox" || c.ListItem(0)["role"] == "option" {
		t.Fatal("disclosure rows became selection options")
	}
}

func TestMenuAnatomyAndTargets(t *testing.T) {
	for _, name := range []string{"star", "download", "edit", "check", "chevron_right"} {
		if webview.MaterialSymbolSVG(name) == "" {
			t.Fatalf("menu icon missing: %s", name)
		}
	}
	c := webview.ComponentCase{Family: "menu", Variant: "standard", Orientation: "vertical", Shape: "cascade", Expanded: true, State: "focused"}
	if c.MenuItem(0)["aria-expanded"] != "true" || c.MenuPanel(true)["data-menu-activity"] != "active" || c.MenuPanel(false)["data-menu-activity"] != "inactive" {
		t.Fatal("cascade preview must expose its active submenu")
	}
	c.State = "interactive"
	if c.MenuPanel(true)["popover"] != "auto" {
		t.Fatal("interactive submenu must use native popover")
	}
	c.Shape = "icon-only"
	c.Orientation = "horizontal"
	c.Selected = true
	c.State = "disabled"
	a := c.MenuItem(0)
	if a["aria-label"] != "Favorite" || a["aria-checked"] != "true" || a["aria-disabled"] != "true" || a["disabled"] != nil {
		t.Fatal("icon-only menu needs a label, selection and a focusable disabled state")
	}
}

func TestSubmenuControlOnlyAppearsForCascades(t *testing.T) {
	cases := []webview.ComponentCase{{Family: "menu", Variant: "standard", Orientation: "vertical", Shape: "plain", State: "enabled"}, {Family: "menu", Variant: "standard", Orientation: "vertical", Shape: "cascade", State: "enabled"}, {Family: "menu", Variant: "standard", Orientation: "vertical", Shape: "cascade", Expanded: true, State: "enabled"}}
	for _, shape := range []string{"plain", "cascade"} {
		preview, err := playground.SelectComponentPreview(cases, map[string][]string{"preview.shape": {shape}})
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, control := range preview.Controls {
			if control.Name == "disclosure" {
				found = true
			}
		}
		if found != (shape == "cascade") {
			t.Fatalf("submenu control in %s: %v", shape, found)
		}
	}
}

func TestConfiguredHorizontalMenuItemsCanBeSelected(t *testing.T) {
	for _, shape := range []string{"plain", "icon-only"} {
		c := webview.ComponentCase{Family: "menu", Variant: "standard", Orientation: "horizontal", Shape: shape, Selected: true, State: "interactive", MenuMode: "multiple"}
		for index := range 3 {
			a := c.MenuItem(index)
			expected := "false"
			if index == 0 {
				expected = "true"
			}
			if a["role"] != "menuitemcheckbox" || a["aria-checked"] != expected || a["data-selected"] != expected {
				t.Fatalf("%s item %d lacks independent selection: %v", shape, index, a)
			}
		}
	}
}

func TestMenuBehaviorConfiguration(t *testing.T) {
	for _, mode := range []string{"actions", "single", "multiple"} {
		page := &webview.ComponentPreview{Cases: []webview.ComponentCase{{Family: "menu", Orientation: "horizontal", Shape: "plain", Selected: true}}}
		if err := playground.ConfigureMenuPreview(page, map[string][]string{"preview.behavior": {mode}}); err != nil {
			t.Fatal(err)
		}
		c := page.Cases[0]
		for index := range 3 {
			a := c.MenuItem(index)
			if mode == "actions" && index > 0 {
				if a["role"] != "menuitem" {
					t.Fatal("actions must remain actions")
				}
				continue
			}
			role := "menuitemcheckbox"
			if mode == "single" {
				role = "menuitemradio"
			}
			if a["role"] != role {
				t.Fatalf("%s item %d: %v", mode, index, a)
			}
			if mode != "actions" && (a["aria-checked"] == "true") != (index == 0) {
				t.Fatal("initial selected state must use the first item")
			}
		}
	}
	page := &webview.ComponentPreview{Cases: []webview.ComponentCase{{Family: "menu"}}}
	if playground.ConfigureMenuPreview(page, map[string][]string{"preview.behavior": {"invalid"}}) == nil {
		t.Fatal("invalid behavior accepted")
	}
}

func TestNavigationBarConfigurations(t *testing.T) {
	for _, variant := range []string{"adaptive", "vertical", "horizontal", "baseline"} {
		page := &webview.ComponentPreview{Cases: []webview.ComponentCase{{Family: "nav-bar", Variant: variant, State: "interactive", Selected: true}}}
		if err := playground.ConfigureNavBarPreview(page, map[string][]string{"preview.destinations": {"5"}, "preview.viewport": {"compact"}, "preview.badge": {"labeled"}, "preview.direction": {"rtl"}}); err != nil {
			t.Fatal(err)
		}
		c := page.Cases[0]
		if len(c.NavigationLabels()) != 5 {
			t.Fatal("missing destinations")
		}
		wantOrientation := variant
		if variant == "baseline" {
			wantOrientation = "vertical"
		}
		if c.Attributes()["data-orientation"] != wantOrientation {
			t.Errorf("%s: wrong orientation", variant)
		}
		var out bytes.Buffer
		if err := webrender.SpecimenNavBar(c).Render(context.Background(), &out); err != nil {
			t.Fatal(err)
		}
		html := out.String()
		if strings.Count(html, `data-navigation-item`) != 5 || strings.Count(html, `aria-current="page"`) != 1 {
			t.Fatal("bar must have five destinations and one selected destination")
		}
		for _, want := range []string{`dir="rtl"`, `data-navbar-page="Settings"`, `md-nav-badge`, `Files 3 unread files`, `class="md-nav-items"`} {
			if !strings.Contains(html, want) {
				t.Errorf("%s missing %s", variant, want)
			}
		}
		if c.NavigationItem(0)["aria-controls"] != c.ID()+"-screen" {
			t.Fatal("destination must control its view")
		}
	}
}

func TestNavigationBarRejectsUnsupportedDestinationCounts(t *testing.T) {
	for _, value := range []string{"2", "6", "0", "all"} {
		if err := playground.ConfigureNavBarPreview(&webview.ComponentPreview{}, map[string][]string{"preview.destinations": {value}}); err == nil {
			t.Errorf("accepted count %s", value)
		}
	}
}

func TestHorizontalNavigationUsesAMediumExampleWindow(t *testing.T) {
	page := &webview.ComponentPreview{Cases: []webview.ComponentCase{{Family: "nav-bar", Variant: "horizontal"}}}
	if err := playground.ConfigureNavBarPreview(page, map[string][]string{"preview.viewport": {"compact"}}); err != nil {
		t.Fatal(err)
	}
	if page.Cases[0].NavViewport != "medium" {
		t.Fatal("horizontal items must not be demonstrated in an unsupported compact window")
	}
	for _, control := range page.Controls {
		if control.Name == "viewport" && playground.ContainsPreview(control.Options, "compact") {
			t.Fatal("compact horizontal example is not valid")
		}
	}
}

func TestPreviewFamiliesShareOneDecodedCoverageFile(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "components-coverage.json")
	if err := os.WriteFile(file, []byte(`{"configurations":[{"family":"switch","variant":"icon","state":"enabled"},{"family":"badge","variant":"labeled","state":"enabled"}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	server := New(publish.PublishedState{}, "")
	server.tokenAssetsDir = directory
	for i, slug := range []string{"switch", "badges", "switch"} {
		if i == 1 {
			if err := os.Remove(file); err != nil {
				t.Fatal(err)
			}
		}
		page, err := server.buildComponentPreview(publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectComponent, Slug: slug}}, url.Values{})
		if err != nil || page == nil || len(page.Cases) == 0 {
			t.Fatalf("%s reread coverage or lost its configurations: %v", slug, err)
		}
	}
}

func TestShowcaseUpdatesReturnOnlyCacheablePreviewHTML(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "components-coverage.json"), []byte(`{"configurations":[{"family":"switch","variant":"icon","state":"enabled"}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	key := publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "switch"}
	server := New(publish.PublishedState{Subjects: map[publish.SubjectKey]publish.SubjectView{key: {Key: key, Name: "Switch"}}}, "")
	server.tokenAssetsDir = directory
	for _, target := range []string{"", "component-preview", "section#component-preview"} {
		partial := target != ""
		request := httptest.NewRequest("GET", "/components/switch?preview.compare=interactive", nil)
		if partial {
			request.Header.Set("HX-Request", "true")
			request.Header.Set("HX-Target", target)
		}
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, request)
		if response.Code != 200 {
			t.Fatal(response.Body.String())
		}
		body := response.Body.String()
		if !partial {
			decoded := html.UnescapeString(body)
			for _, want := range []string{
				`<title>Material Design 3 "Switch" tokens and playground | Token Catalog</title>`,
				`>Material Design 3 "Switch" tokens and playground</h1>`,
				`content="Material Design 3 "Switch" tokens. Try the interactive playground, compare values by context, follow references and download CSS variables or JSON."`,
			} {
				if !strings.Contains(decoded, want) {
					t.Fatalf("missing metadata or preserved heading: %s", want)
				}
			}
		}
		if strings.Contains(body, "<!doctype html>") == partial || !strings.Contains(body, `id="component-preview"`) {
			t.Fatal("wrong document or preview representation")
		}
		if partial && (response.Header().Get("Cache-Control") != "public, no-cache" || !strings.Contains(response.Header().Get("Vary"), "HX-Target")) {
			t.Fatal("preview caching must distinguish the full document")
		}
		if !strings.Contains(body, `hx-trigger="change"`) || !strings.Contains(body, `data-preview-compare`) || !strings.Contains(body, `name="preview.compare"`) {
			t.Fatal("preview must apply connected comparison selections directly")
		}
	}
}

func TestSearchDisclosure(t *testing.T) {
	c := webview.ComponentCase{Family: "search", Variant: "docked", Control: "contained", Width: "clear", SearchContent: "text", Expanded: true, State: "enabled"}
	if c.SearchPanel()["data-open"] != "true" || c.SearchPanel()["role"] != "dialog" {
		t.Fatal("focused search must use the persistent surface")
	}
	if c.SearchQuery() != "Buttons" || c.SearchPanel()["data-open"] != "true" {
		t.Fatal("search query and disclosure must match the specimen")
	}
	c.State = "interactive"
	if _, ok := c.SearchPanel()["open"]; ok {
		t.Fatal("search must stay within the preview canvas")
	}
	c.Expanded = false
	c.Width = "none"
	c.SearchContent = "empty"
	if c.SearchQuery() != "" || c.SearchPanel()["data-open"] != "false" {
		t.Fatal("empty collapsed bar must show its placeholder")
	}
}

func TestSearchIndependentConfiguration(t *testing.T) {
	c := webview.ComponentCase{Family: "search", Variant: "adaptive", Width: "clear-settings"}
	if c.SearchQuery() != "" || !c.SearchHasAction("clear") || !c.SearchHasAction("settings") || c.SearchHasAction("avatar") {
		t.Fatal("search query and action configuration are coupled")
	}
	if c.SearchPanel()["aria-label"] != "Search tokens" || c.SearchPanel()["data-layout"] != "adaptive" {
		t.Fatal("search labeling or adaptive layout lost")
	}
	page := &webview.ComponentPreview{Cases: []webview.ComponentCase{c}}
	if err := playground.ConfigureSearchPreview(page, map[string][]string{"preview.content": {"text"}}); err != nil {
		t.Fatal(err)
	}
	if page.Cases[0].SearchQuery() != "Buttons" {
		t.Fatal("explicit sample query missing")
	}
}

func TestSearchComparisons(t *testing.T) {
	var cases []webview.ComponentCase
	for _, layout := range []string{"adaptive", "docked", "full-screen"} {
		for _, style := range []string{"contained", "baseline"} {
			for _, expanded := range []bool{false, true} {
				cases = append(cases, webview.ComponentCase{Family: "search", Variant: layout, Control: style, Expanded: expanded, State: "enabled"})
			}
		}
	}
	for _, comparison := range []string{"layouts", "styles"} {
		p, err := playground.SelectComponentPreview(cases, map[string][]string{"preview.compare": {comparison}})
		if err != nil {
			t.Fatal(err)
		}
		if len(p.Cases) != 2 {
			t.Fatalf("%s: wanted two cases, got %d", comparison, len(p.Cases))
		}
		for _, c := range p.Cases {
			if !c.Expanded {
				t.Fatal("comparisons must show focused layout/style differences")
			}
		}
	}
}

func TestSearchKeepsOneHeaderAndInput(t *testing.T) {
	for _, expanded := range []bool{false, true} {
		c := webview.ComponentCase{Family: "search", Variant: "docked", Control: "contained", Shape: "icon", Width: "clear-settings", Expanded: expanded, State: "interactive"}
		var html bytes.Buffer
		if err := webrender.ComponentSpecimen(c).Render(context.Background(), &html); err != nil {
			t.Fatal(err)
		}
		text := html.String()
		if strings.Count(text, `class="md-search-input"`) != 1 || strings.Count(text, `class="md-search-bar md-search-trigger"`) != 1 || strings.Contains(text, `<dialog class="md-search-view`) {
			t.Fatal("search must retain one header/input rather than a duplicate popup")
		}
	}
}

func TestSearchElevationIsExplicit(t *testing.T) {
	for _, style := range []string{"baseline", "contained"} {
		page := &webview.ComponentPreview{Cases: []webview.ComponentCase{{Family: "search", Control: style}}}
		if err := playground.ConfigureSearchPreview(page, map[string][]string{}); err != nil {
			t.Fatal(err)
		}
		if page.Cases[0].SearchElevation != "none" || strings.Contains(page.Cases[0].SearchPanel()["class"].(string), "md-elevation") {
			t.Fatal("search must have no shadow by default")
		}
		if err := playground.ConfigureSearchPreview(page, map[string][]string{"preview.shadow": {"token"}}); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(page.Cases[0].SearchBar()["class"].(string), "md-elevation") || !strings.Contains(page.Cases[0].SearchPanel()["class"].(string), "md-elevation") {
			t.Fatal("explicit elevation must be available to both styles")
		}
	}
}

func TestAdaptiveSearchUsesComponentPrimitives(t *testing.T) {
	c := webview.ComponentCase{Family: "search", Variant: "adaptive", Control: "contained", Shape: "icon", Width: "clear", State: "interactive"}
	var html bytes.Buffer
	assets := ui.Assets{}
	if err := webrender.ComponentSpecimen(c).Render(ui.WithAssets(context.Background(), assets), &html); err != nil {
		t.Fatal(err)
	}
	text := html.String()
	for _, module := range []string{"navigation", "list", "card", "fab", "search"} {
		if !assets[module] {
			t.Errorf("composition missing %s dependency", module)
		}
	}
	if !strings.Contains(text, "md-nav-fab") {
		t.Fatal("nested FAB must use rail anatomy")
	}
	if strings.Contains(text, `data-family="nav-drawer"`) {
		t.Fatal("adaptive composition must not include a drawer")
	}
	if strings.Count(text, `data-search-message=`) != len(webview.SearchExamples) || strings.Count(text, `class="md-list-avatar-label"`) != len(webview.SearchExamples) {
		t.Fatal("mail content must use the shared list avatar slots for every message")
	}
	if !strings.Contains(text, `data-search-reply`) || !strings.Contains(text, `data-shape="expressive"`) {
		t.Fatal("mail content must use typed action and expressive list configuration")
	}
	if strings.Count(text, `class="md-search-input"`) != 1 {
		t.Fatal("composition must keep a single search input")
	}
}

func TestSegmentedDensityConfiguration(t *testing.T) {
	for _, variant := range []string{"single", "multiple"} {
		for _, control := range []string{"label", "icon", "icon-label"} {
			for _, density := range []string{"0", "-1", "-2", "-3"} {
				original := webview.ComponentCase{Family: "segmented", Variant: variant, Control: control}
				page := &webview.ComponentPreview{Cases: []webview.ComponentCase{original}}
				if err := playground.ConfigureSegmentedPreview(page, url.Values{"preview.density": {density}}); err != nil {
					t.Fatal(err)
				}
				c := page.Cases[0]
				if c.Attributes()["data-density"] != density || c.Control != control || c.Variant != variant {
					t.Fatal("density must apply independently of variant and content")
				}
				if (c.ID() == original.ID()) != (density == "0") {
					t.Fatal("nondefault densities need distinct specimen IDs")
				}
			}
		}
	}
	if err := playground.ConfigureSegmentedPreview(&webview.ComponentPreview{}, url.Values{"preview.density": {"-4"}}); err == nil {
		t.Fatal("unsupported density must be rejected")
	}
}

func TestSegmentedSelectionRoles(t *testing.T) {
	for variant, role := range map[string]string{"single": "radiogroup", "multiple": "group"} {
		var out bytes.Buffer
		if err := webrender.SpecimenSegmented(webview.ComponentCase{Family: "segmented", Variant: variant, Control: "label", State: "interactive"}).Render(context.Background(), &out); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), `role="`+role+`"`) {
			t.Fatal(variant, "missing", role)
		}
	}
}
