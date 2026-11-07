// Integration checks for playground.
package server

import (
	"bytes"
	"context"
	"errors"
	"github.com/sojupie/tokens/catalog/web/internal/playground"
	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	webrender "github.com/sojupie/tokens/catalog/web/internal/render"
	"github.com/sojupie/tokens/source/publish"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestButtonGalleryUsesGeneratedConfigurations(t *testing.T) {
	path := os.Getenv("TOKEN_CATALOG_THEME_DB")
	if path == "" {
		t.Skip("set TOKEN_CATALOG_THEME_DB for generated button gallery checks")
	}
	state, err := publish.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	server := New(state, "")
	if err := server.UseGeneratedAssets(filepath.Dir(path)); err != nil {
		t.Fatal(err)
	}
	if err := server.SetPublicOrigin("https://catalog.example/md3"); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		compare string
		count   int
	}{{"states", 5}, {"sizes", 5}, {"variants", 5}, {"interactive", 1}} {
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/md3/components/buttons?preview.compare="+tc.compare, nil))
		if response.Code != 200 {
			t.Fatal(response.Body.String())
		}
		if count := strings.Count(response.Body.String(), "data-button-case="); count != tc.count {
			t.Fatalf("%s: got %d cases, want %d", tc.compare, count, tc.count)
		}
		if !strings.Contains(response.Body.String(), "Browse tokens") {
			t.Fatal("preview replaced token table")
		}
		if tc.compare == "interactive" && strings.Contains(response.Body.String(), `data-button-state=`) {
			t.Fatal("interactive preview forces an interaction state")
		}
	}
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/md3/components/buttons?preview.size=all", nil))
	if response.Code != 400 {
		t.Fatalf("unsupported size: %d", response.Code)
	}
}

func TestCarouselActionConfigurations(t *testing.T) {
	for _, action := range []string{"none", "show-all", "header-arrow"} {
		for _, variant := range []string{"hero", "full-screen"} {
			page := &webview.ComponentPreview{Cases: []webview.ComponentCase{{Family: "carousel", Variant: variant, State: "interactive"}}}
			if err := playground.ConfigureCarouselPreview(page, url.Values{"preview.action": {action}}); err != nil {
				t.Fatal(err)
			}
			var out bytes.Buffer
			if err := webrender.SpecimenCarousel(page.Cases[0]).Render(context.Background(), &out); err != nil {
				t.Fatal(err)
			}
			html := out.String()
			if strings.Contains(html, "href=") {
				t.Fatal("specimen actions must not navigate")
			}
			bottom := strings.Contains(html, `class="md-carousel-all-action"`)
			header := strings.Contains(html, `class="md-carousel-header"`)
			if bottom != (action == "show-all" && variant != "full-screen") || header != (action == "header-arrow" && variant != "full-screen") {
				t.Fatalf("%s/%s: wrong action placement", variant, action)
			}
			if (bottom || header) && !strings.Contains(html, `type="button"`) {
				t.Fatal("actions must use button primitives")
			}
		}
	}
	page := &webview.ComponentPreview{Cases: []webview.ComponentCase{{Family: "carousel"}}}
	if err := playground.ConfigureCarouselPreview(page, url.Values{"preview.action": {"invalid"}}); err == nil {
		t.Fatal("invalid action accepted")
	}
	if err := playground.ConfigureCarouselPreview(page, nil); err != nil || page.Cases[0].CarouselAction != "show-all" {
		t.Fatal("default action changed")
	}
}

func TestCarouselContentAndViewport(t *testing.T) {
	for _, variant := range []string{"multi-browse", "hero", "center-aligned-hero", "uncontained", "uncontained-multi-aspect-ratio", "full-screen"} {
		for _, content := range []string{"none", "overlay", "above", "below"} {
			page := &webview.ComponentPreview{Cases: []webview.ComponentCase{{Family: "carousel", Variant: variant, State: "interactive"}}}
			if err := playground.ConfigureCarouselPreview(page, url.Values{"preview.content": {content}, "preview.viewport": {"360"}}); err != nil {
				t.Fatal(err)
			}
			for _, control := range page.Controls {
				if control.Name == "content" && strings.Join(control.Options, ",") != "none,overlay" {
					t.Fatal("external text must remain consumer composition")
				}
			}
			var out bytes.Buffer
			if err := webrender.SpecimenCarousel(page.Cases[0]).Render(context.Background(), &out); err != nil {
				t.Fatal(err)
			}
			html := out.String()
			if !strings.Contains(html, "inline-size:min(100%,360px)") {
				t.Fatal("missing environment width")
			}
			if strings.Contains(html, `class="md-carousel-caption`) != (content == "overlay") {
				t.Fatalf("%s/%s: overlay", variant, content)
			}
			external := strings.Contains(html, `<p class="md-typescale-body-medium"`)
			if external {
				t.Fatalf("%s/%s: external content", variant, content)
			}
			if strings.Contains(html, `class="md-carousel-pagination"`) != (variant != "full-screen") || strings.Contains(html, `<output`) || strings.Contains(html, `class="md-carousel-number"`) {
				t.Fatal("must use position, not item counts")
			}
		}
	}
	for _, query := range []url.Values{{"preview.content": {"invalid"}}, {"preview.viewport": {"-1"}}} {
		if err := playground.ConfigureCarouselPreview(&webview.ComponentPreview{}, query); err == nil {
			t.Fatal("invalid option accepted")
		}
	}
}

func TestCarouselItemStates(t *testing.T) {
	if webview.MaterialSymbolSVG("arrow_forward") == "" {
		t.Fatal("header action arrow must be available")
	}
	if webview.MaterialSymbolSVG("chevron_left") == "" {
		t.Fatal("previous-item icon must be available")
	}
	c := webview.ComponentCase{Family: "carousel", State: "disabled"}
	first, second := c.CarouselItem(0), c.CarouselItem(1)
	if first["disabled"] != true || first["data-component-state"] != "disabled" {
		t.Fatal("forced disabled state must apply to the first item")
	}
	if _, ok := second["disabled"]; ok {
		t.Fatal("neighboring items must remain enabled")
	}
	c.State = "interactive"
	if _, ok := c.CarouselItem(0)["data-component-state"]; ok {
		t.Fatal("interactive items must use native interaction states")
	}
	if _, ok := c.Attributes()["aria-disabled"]; ok {
		t.Fatal("a disabled specimen must not disable the whole carousel")
	}
}

func TestCarouselAspectRatioExamples(t *testing.T) {
	c := webview.ComponentCase{Family: "carousel", Variant: "uncontained-multi-aspect-ratio"}
	seen := map[string]bool{}
	for i := range 4 {
		seen[c.CarouselSlot(i)["style"].(string)] = true
	}
	if len(seen) != 4 {
		t.Fatal("multi-aspect variant must demonstrate four distinct source ratios")
	}
	c.Variant = "uncontained"
	if _, ok := c.CarouselSlot(0)["style"]; ok {
		t.Fatal("ordinary uncontained items must keep a single width")
	}
}

func TestCarouselScrollingAndScreenOptions(t *testing.T) {
	page := &webview.ComponentPreview{Cases: []webview.ComponentCase{{Family: "carousel", Variant: "uncontained-multi-aspect-ratio"}, {Family: "carousel", Variant: "full-screen"}}}
	if err := playground.ConfigureCarouselPreview(page, map[string][]string{"preview.scrolling": {"snap"}, "preview.screen": {"portrait"}}); err != nil {
		t.Fatal(err)
	}
	if page.Cases[0].CarouselScroll != "snap" || page.Cases[1].Orientation != "portrait" {
		t.Fatal("explicit configuration lost")
	}
	if err := playground.ConfigureCarouselPreview(page, map[string][]string{"preview.screen": {"other"}}); err == nil {
		t.Fatal("invalid frame accepted")
	}
	if err := playground.ConfigureCarouselPreview(page, map[string][]string{"preview.scrolling": {"other"}}); err == nil {
		t.Fatal("invalid scrolling accepted")
	}
}

func TestCarouselControlsOnlyApplyToVisibleLayouts(t *testing.T) {
	for _, variant := range []string{"hero", "full-screen"} {
		page := &webview.ComponentPreview{Cases: []webview.ComponentCase{{Family: "carousel", Variant: variant}}}
		if err := playground.ConfigureCarouselPreview(page, nil); err != nil {
			t.Fatal(err)
		}
		for _, c := range page.Controls {
			if c.Name == "scrolling" || c.Name == "screen" {
				t.Fatalf("%s exposes irrelevant %s", variant, c.Name)
			}
		}
		if variant == "full-screen" && page.Cases[0].Orientation != "portrait" {
			t.Fatal("full screen defaults to landscape")
		}
	}
}

func TestComponentGalleryRendersNativeStates(t *testing.T) {
	path := os.Getenv("TOKEN_CATALOG_THEME_DB")
	if path == "" {
		t.Skip("set TOKEN_CATALOG_THEME_DB to check the captured publication")
	}
	state, err := publish.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	server := New(state, "")
	if err = server.UseGeneratedAssets(filepath.Dir(path)); err != nil {
		t.Fatal(err)
	}
	if err = server.SetPublicOrigin("https://catalog.example/md3"); err != nil {
		t.Fatal(err)
	}

	for _, slug := range []string{"cards", "chips", "text-fields", "menus", "icon-buttons", "radio-button", "dialogs", "divider", "checkbox", "switch", "floating-action-button", "extended-fab", "badges", "snackbar", "tooltips", "navigation-bar", "navigation-rail", "navigation-drawer", "tabs", "segmented-buttons", "sliders", "progress-indicators", "loading-indicator", "app-bars", "toolbars", "bottom-sheets", "side-sheets", "split-button", "button-groups", "fab-menu", "lists", "carousel", "search", "date-pickers", "time-pickers"} {
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/md3/components/"+slug+"?preview.compare=states", nil))
		if response.Code != 200 {
			t.Fatalf("%s: %s", slug, response.Body.String())
		}
		body := response.Body.String()
		if !strings.Contains(body, `class="tokens-preview-grid"`) || !strings.Contains(body, `name="preview.compare"`) {
			t.Fatalf("missing preview: %s", slug)
		}
		if count := strings.Count(body, `class="tokens-component-example"`); count < 1 || count > 6 {
			t.Fatalf("%s rendered %d configurations", slug, count)
		}
		if strings.Contains(body, "data-new-component") {
			t.Fatalf("obsolete implementation label: %s", slug)
		}
		if response.Header().Get("X-Robots-Tag") != "noindex, follow" || !strings.Contains(body, `content="noindex, follow"`) {
			t.Fatalf("preview query must not be indexed: %s", slug)
		}
		if strings.Contains(body, `href="#card-`) || strings.Contains(body, `href="#menu-`) || strings.Contains(body, `action="#component-preview"`) {
			t.Fatalf("preview navigates to an example anchor: %s", slug)
		}
		if !strings.Contains(body, `hx-get="/md3/components/`+slug+`"`) || !strings.Contains(body, `hx-target="#component-preview"`) {
			t.Fatal("preview form must update its own section")
		}
		if strings.Contains(body, `<input class="md-radio"`) {
			t.Fatalf("radio input lacks its feedback host: %s", slug)
		}
		if strings.Contains(body, ` disabled="false"`) || strings.Contains(body, ` checked="false"`) || strings.Contains(body, ` hidden="false"`) {
			t.Fatalf("false native boolean attributes: %s", slug)
		}
	}

	for _, slug := range []string{"progress-indicators", "loading-indicator"} {
		for _, control := range []string{"determinate", "indeterminate"} {
			response := httptest.NewRecorder()
			server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/md3/components/"+slug+"?preview.compare=interactive&preview.control="+control, nil))
			body := response.Body.String()
			if response.Code != 200 || !strings.Contains(body, `role="progressbar"`) {
				t.Fatalf("progress %s: %d", control, response.Code)
			}
			if control == "determinate" {
				if !strings.Contains(body, `aria-valuenow="40"`) || !strings.Contains(body, `class="md-progress-control"`) || !strings.Contains(body, `class="md-slider-input"`) {
					t.Fatal("determinate progress must expose its value and use the shared slider")
				}
			} else if strings.Contains(body, `aria-valuenow=`) || strings.Contains(body, `class="md-progress-control"`) {
				t.Fatal("indeterminate progress must omit a numeric value and slider")
			}
		}

	}
	for _, path := range []string{
		"/md3/components/app-bars?preview.compare=interactive&preview.control=subtitle&preview.variant=large-flexible",
		"/md3/components/app-bars?preview.compare=states&preview.control=fab&preview.variant=bottom",
		"/md3/components/toolbars?preview.compare=interactive&preview.control=vibrant&preview.variant=floating-fab&preview.orientation=vertical&preview.size=medium&preview.shape=icon",
	} {
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest("GET", path, nil))
		if response.Code != 200 {
			t.Fatalf("bar configuration %s: %d %s", path, response.Code, response.Body.String())
		}
		if strings.Contains(path, "toolbars") && (!strings.Contains(response.Body.String(), `role="toolbar"`) || !strings.Contains(response.Body.String(), `aria-orientation="vertical"`) || !strings.Contains(response.Body.String(), `md-toolbar-fab`)) {
			t.Fatal("toolbar must expose its orientation and separate FAB")
		}
	}

	for _, path := range []string{"/md3/components"} {
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest("GET", path, nil))
		if response.Code != 200 {
			t.Fatalf("%s: %d", path, response.Code)
		}
		body := response.Body.String()
		if !strings.Contains(body, `class="tokens-component-tile"`) || !strings.Contains(body, `data-thumbnail-missing`) {
			t.Fatal("missing image placeholder")
		}
		if strings.Contains(body, `class="md-time-panel"`) || strings.Contains(body, `data-thumbnail=`) {
			t.Fatal("missing image must not fall back to live components")
		}
		if !strings.Contains(body, `class="tokens-component-tile-link"`) {
			t.Fatal("missing separate card link")
		}
	}
	for _, path := range []string{"/md3/components", "/md3/styles", "/md3/components?q=button", "/md3/styles?q=shape", "/md3/components?q=segmented", "/md3/styles?q=nonexistent"} {
		for _, fragment := range []bool{false, true} {
			request := httptest.NewRequest("GET", path, nil)
			if fragment {
				request.Header.Set("HX-Request", "true")
			}
			response := httptest.NewRecorder()
			server.Handler().ServeHTTP(response, request)
			body := response.Body.String()
			if response.Code != 200 || !strings.Contains(body, `class="tokens-card-grid tokens-layout-grid"`) || strings.Contains(body, `data-new-component`) {
				t.Fatalf("card grid missing: %s fragment=%v status=%d", path, fragment, response.Code)
			}
			if strings.Contains(path, "q=segmented") && strings.Count(body, `class="tokens-component-tile"`) != 1 {
				t.Fatal("single search result must retain exactly one illustrated card")
			}
			if strings.Contains(path, "/styles?q=shape") && (!strings.Contains(body, `data-thumbnail-missing`) || strings.Contains(body, `data-style-thumbnail=`)) {
				t.Fatal("missing style image must not fall back to live components")
			}
			if strings.Contains(path, "q=nonexistent") && !strings.Contains(body, "No matching subjects.") {
				t.Fatal("empty search needs a result message")
			}
		}
	}
	thumbnails, err := server.componentThumbnails(sortedSubjects(state))
	if err != nil {
		t.Fatal(err)
	}
	if c := thumbnails["segmented-buttons"]; c.Control != "label" || c.Variant != "single" || !c.Selected {
		t.Fatalf("segmented illustration must match its label-only anatomy: %+v", c)
	}
	if c := thumbnails["carousel"]; c.CarouselContent != "overlay" {
		t.Fatalf("carousel capture must show text on the image: %+v", c)
	}
	again, err := server.componentThumbnails(sortedSubjects(state))
	if err != nil {
		t.Fatal(err)
	}
	for _, slug := range []string{"menus", "dialogs"} {
		if c, ok := thumbnails[slug]; ok && c.State != "enabled" {
			t.Fatalf("%s thumbnail must show its component, not its opener", slug)
		}
	}
	for slug, c := range thumbnails {
		if c != again[slug] || (c.State != "enabled") {
			t.Fatalf("unstable thumbnail for %s", slug)
		}
	}
	for _, path := range []string{"/md3/lab/buttons", "/md3/lab/components"} {
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest("GET", path, nil))
		if response.Code != 404 {
			t.Fatalf("removed route %s returned %d", path, response.Code)
		}
	}
}

func TestSelectionThumbnailsHaveOneLabelPerOption(t *testing.T) {
	for _, tc := range []struct {
		family, defaultLabel string
		labels               []string
	}{
		{"checkbox", "Receive updates", []string{"Design tokens", "Components", "All collections"}},
		{"switch", "Allow notifications", []string{"Notifications", "Offline access"}},
	} {
		var out bytes.Buffer
		if err := webrender.ComponentThumbnail(webview.ComponentCase{Family: tc.family, State: "enabled"}).Render(context.Background(), &out); err != nil {
			t.Fatal(err)
		}
		markup := out.String()
		if strings.Contains(markup, tc.defaultLabel) {
			t.Fatal("thumbnail must not include the standalone specimen's default label")
		}
		for _, label := range tc.labels {
			if strings.Count(markup, label) != 1 {
				t.Fatalf("expected one visible label for %s", label)
			}
		}
	}
}

func TestCheckboxThumbnailUsesSelectedTokensForMixedState(t *testing.T) {
	var output bytes.Buffer
	if err := webrender.ComponentThumbnail(webview.ComponentCase{Family: "checkbox", State: "enabled"}).Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	mixed := 0
	for _, input := range strings.Split(output.String(), "<input")[1:] {
		tag := strings.SplitN(input, ">", 2)[0]
		if strings.Contains(tag, `data-indeterminate="true"`) {
			mixed++
			if !strings.Contains(tag, " checked") {
				t.Fatal("mixed thumbnail uses unselected checkbox tokens")
			}
		}
	}
	if mixed != 1 {
		t.Fatalf("got %d mixed checkboxes, want 1", mixed)
	}
}

func TestButtonPreviewColorFamily(t *testing.T) {
	cases := []webview.ComponentCase{{Family: "button", Variant: "filled", Size: "small", Shape: "round", State: "enabled"}}
	for _, color := range []string{"default", "primary", "secondary", "tertiary", "error"} {
		page, err := playground.SelectComponentPreview(cases, url.Values{"preview.color": {color}})
		if err != nil {
			t.Fatal(err)
		}
		want := color
		if want == "default" {
			want = ""
		}
		if len(page.Cases) != 1 || page.Cases[0].Color != want {
			t.Fatalf("color %s: unexpected cases %#v", color, page.Cases)
		}
	}
	if _, err := playground.SelectComponentPreview(cases, url.Values{"preview.color": {"unknown"}}); !errors.Is(err, playground.ErrPreviewInput) {
		t.Fatalf("unexpected invalid color result: %v", err)
	}
}

func TestPreviewVariesOneDimensionAndKeepsValidCombinations(t *testing.T) {
	var cases []webview.ComponentCase
	for _, variant := range []string{"tonal", "text"} {
		for _, size := range []string{"small", "large"} {
			for _, selection := range []string{"", "selected"} {
				if variant == "text" && selection != "" {
					continue
				}
				for _, state := range []string{"enabled", "hovered", "disabled"} {
					cases = append(cases, webview.ComponentCase{Family: "button", Variant: variant, Size: size, Shape: "round", Selection: selection, State: state})
				}
			}
		}
	}
	for _, tc := range []struct {
		query string
		count int
	}{
		{"", 2}, {"preview.compare=sizes", 2}, {"preview.compare=variants", 2},
		{"preview.compare=variants&preview.selection=selected", 1},
		{"preview.variant=text&preview.selection=selected", 1},
		{"preview.compare=interactive&preview.state=disabled", 1},
	} {
		query, _ := url.ParseQuery(tc.query)
		page, err := playground.SelectComponentPreview(cases, query)
		if err != nil {
			t.Fatal(err)
		}
		if len(page.Cases) != tc.count {
			t.Fatalf("%s: got %d specimens, want %d", tc.query, len(page.Cases), tc.count)
		}
		for _, c := range page.Cases {
			if c.Variant == "text" && c.Selection != "" {
				t.Fatal("invented unsupported selected text button")
			}
			if page.Interactive && c.State != "interactive" {
				t.Fatal("interactive case has forced state")
			}
		}
	}
	page, err := playground.SelectComponentPreview(cases, url.Values{"q": {"shadow"}, "context.Theme": {"Dark"}, "preview.compare": {"sizes"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Hidden) != 2 {
		t.Fatalf("lost token table filters: %#v", page.Hidden)
	}
	for _, query := range []url.Values{{"preview.compare": {"all"}}, {"preview.size": {"all"}}} {
		if _, err := playground.SelectComponentPreview(cases, query); !errors.Is(err, playground.ErrPreviewInput) {
			t.Fatalf("unsupported value accepted: %v", err)
		}
	}
}

func TestCheckboxPreviewSelectsIndeterminateWithoutRedundantVariant(t *testing.T) {
	cases := []webview.ComponentCase{
		{Family: "checkbox", Variant: "unchecked", State: "enabled"},
		{Family: "checkbox", Variant: "checked", Selected: true, State: "enabled"},
		{Family: "checkbox", Variant: "indeterminate", Selected: true, State: "enabled"},
	}
	page, err := playground.SelectComponentPreview(cases, url.Values{"preview.selection": {"indeterminate"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Cases) != 1 || page.Cases[0].Variant != "indeterminate" {
		t.Fatalf("wrong checkbox selection: %#v", page.Cases)
	}
	for _, control := range page.Controls {
		if control.Name == "variant" {
			t.Fatal("checkbox selection duplicated as variant")
		}
	}
}

func TestSnackbarPreviewFocusesOnlyTheChosenControl(t *testing.T) {
	c := webview.ComponentCase{Family: "snackbar", Control: "action", State: "focused"}
	if c.FeedbackAttributes("md-button md-feedback-action")["data-component-state"] != "focused" || c.FeedbackAttributes("md-icon-button md-snackbar-close")["data-component-state"] != "enabled" {
		t.Fatal("snackbar forced two controls into focus")
	}
	c.Control = "dismiss"
	if c.FeedbackAttributes("md-button md-feedback-action")["data-component-state"] != "enabled" || c.FeedbackAttributes("md-icon-button md-snackbar-close")["data-component-state"] != "focused" {
		t.Fatal("dismiss control does not own focus")
	}
	c.State = "interactive"
	if _, exists := c.FeedbackAttributes("md-button md-feedback-action")["data-component-state"]; exists {
		t.Fatal("interactive action has a forced state")
	}
}

func TestNavigationPreviewPreservesOrientationAndNativeSelection(t *testing.T) {
	c := webview.ComponentCase{Family: "nav-rail", Variant: "expanded", Orientation: "horizontal", State: "focused", Selected: true}
	if c.NavigationItem(0)["aria-current"] != "page" || c.NavigationItem(1)["aria-current"] != nil {
		t.Fatal("navigation selection does not match the preview")
	}
	if c.NavigationItem(0)["data-component-state"] != "focused" || c.NavigationItem(1)["data-component-state"] != "enabled" {
		t.Fatal("only the target item should have a forced state")
	}
	c.State = "interactive"
	if c.NavigationItem(0)["data-component-state"] != nil {
		t.Fatal("interactive navigation retains a forced state")
	}
	if c.Attributes()["data-orientation"] != "horizontal" {
		t.Fatal("missing navigation orientation")
	}
	other := c
	other.Orientation = "vertical"
	if c.ID() == other.ID() {
		t.Fatal("navigation orientations have duplicate IDs")
	}
}

func TestTabsAndSegmentsUseNativeSelectionAndLinkedPanels(t *testing.T) {
	c := webview.ComponentCase{Family: "tabs", Variant: "primary", Control: "label", State: "focused", Selected: true}
	first, second := c.TabAttributes(0), c.TabAttributes(1)
	if first["aria-selected"] != "true" || first["tabindex"] != "0" || second["aria-selected"] != "false" || second["tabindex"] != "-1" {
		t.Fatal("tabs must have one selected tab in the tab sequence")
	}
	if first["aria-controls"] == second["aria-controls"] {
		t.Fatal("tabs share a panel ID")
	}
	c.Family = "segmented"
	c.Variant = "single"
	if c.SegmentInput(0, "Home")["type"] != "radio" || c.SegmentInput(0, "Home")["checked"] != true {
		t.Fatal("single selection must use native radios")
	}
	c.Variant = "multiple"
	c.State = "disabled"
	if c.SegmentInput(0, "Home")["type"] != "checkbox" || c.SegmentInput(0, "Home")["disabled"] != true {
		t.Fatal("multiple selection must use native checkboxes and native disabled state")
	}
	c.State = "interactive"
	if c.SegmentAttributes(0)["data-component-state"] != nil || c.TabAttributes(0)["data-component-state"] != nil {
		t.Fatal("interactive controls retain a forced state")
	}
}

func TestSliderPreviewUsesNativeRangesAndIndependentHandles(t *testing.T) {
	c := webview.ComponentCase{Family: "slider", Control: "range", Variant: "discrete", Orientation: "vertical", State: "pressed"}
	if len(c.SliderHandles()) != 2 || c.SliderInput("start")["step"] != "10" || c.SliderInput("start")["orient"] != "vertical" {
		t.Fatal("range preview lost its native controls")
	}
	if c.SliderHandle("start")["data-component-state"] != "pressed" || c.SliderHandle("end")["data-component-state"] != "enabled" {
		t.Fatal("preview pressed both range handles")
	}
	c.State = "disabled"
	if c.SliderInput("start")["disabled"] != true || c.SliderInput("end")["disabled"] != true {
		t.Fatal("disabled ranges must disable both inputs")
	}
	c.Control = "single"
	c.State = "interactive"
	c.Variant = "continuous"
	if len(c.SliderHandles()) != 1 || c.SliderInput("end")["step"] != "1" || c.SliderHandle("end")["data-component-state"] != nil {
		t.Fatal("interactive slider has an invented forced state")
	}
}

func TestPreviewDefaultsToVariantsAndUnknownProgress(t *testing.T) {
	for _, family := range []string{"progress", "loading"} {
		var cases []webview.ComponentCase
		for _, control := range []string{"determinate", "indeterminate"} {
			for _, variant := range []string{"first", "second"} {
				cases = append(cases, webview.ComponentCase{Family: family, Control: control, Variant: variant, State: "enabled"})
			}
		}
		for _, control := range []string{"", "determinate"} {
			query := url.Values{}
			if control != "" {
				query.Set("preview.control", control)
			}
			page, err := playground.SelectComponentPreview(cases, query)
			if err != nil {
				t.Fatal(err)
			}
			want := control
			if want == "" {
				want = "indeterminate"
			}
			if page.Dimension != "variant" || len(page.Cases) != 2 {
				t.Fatalf("%s: expected two variants: %#v", family, page)
			}
			for _, c := range page.Cases {
				if c.Control != want {
					t.Fatalf("%s: got %s, want %s", family, c.Control, want)
				}
			}
		}
	}
	page, err := playground.SelectComponentPreview([]webview.ComponentCase{{Family: "radio", State: "enabled"}, {Family: "radio", State: "hovered"}}, url.Values{})
	if err != nil || page.Dimension != "state" || len(page.Cases) != 2 {
		t.Fatalf("single variant must fall back to states: %#v, %v", page, err)
	}
}

func TestPreviewDefaultsToMediumOrDefaultSize(t *testing.T) {
	for _, tc := range []struct {
		name   string
		family string
		sizes  []string
		query  url.Values
		want   string
	}{
		{"medium", "button", []string{"small", "medium", "large"}, nil, "medium"},
		{"default", "carousel", []string{"small", "default", "large"}, nil, "default"},
		{"slider", "slider", []string{"xsmall", "small", "medium", "large"}, nil, "medium"},
		{"explicit", "button", []string{"small", "medium", "large"}, url.Values{"preview.size": {"small"}}, "small"},
		{"fallback", "button", []string{"small", "large"}, nil, "small"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var cases []webview.ComponentCase
			for _, size := range tc.sizes {
				cases = append(cases, webview.ComponentCase{Family: tc.family, Size: size, State: "enabled"})
			}
			page, err := playground.SelectComponentPreview(cases, tc.query)
			if err != nil {
				t.Fatal(err)
			}
			if len(page.Cases) != 1 || page.Cases[0].Size != tc.want {
				t.Fatalf("got %#v, want size %s", page.Cases, tc.want)
			}
		})
	}
}
