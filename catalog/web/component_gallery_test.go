package web

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sojupie/tokens/source/publish"
)

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
		if !strings.Contains(body, `class="tokens-component-tile"`) || !strings.Contains(body, `class="tokens-component-thumbnail-scene" inert aria-hidden="true"`) {
			t.Fatal("missing inert component illustration")
		}
		if !strings.Contains(body, `class="md-time-panel"`) {
			t.Fatal("missing time picker illustration")
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
			if response.Code != 200 || !strings.Contains(body, `class="tokens-card-grid"`) || strings.Contains(body, `data-new-component`) {
				t.Fatalf("card grid missing: %s fragment=%v status=%d", path, fragment, response.Code)
			}
			if strings.Contains(path, "q=segmented") && strings.Count(body, `class="tokens-component-tile"`) != 1 {
				t.Fatal("single search result must retain exactly one illustrated card")
			}
			if strings.Contains(path, "/styles?q=shape") && !strings.Contains(body, `data-style-thumbnail="shape/corner-radius"`) {
				t.Fatal("style search must retain its illustration")
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
