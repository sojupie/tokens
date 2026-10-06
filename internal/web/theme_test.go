package web

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/sojupie/m3-source/internal/publish"
)

func TestThemeValuesAreTypedCSS(t *testing.T) {
	for raw, expected := range map[string]string{
		`{"color":{"alpha":1}}`:                     "rgba(0, 0, 0, 1)",
		`{"color":{"green":1,"alpha":1}}`:           "rgba(0, 255, 0, 1)",
		`{"color":{}}`:                              "",
		`{"color":{"alpha":null}}`:                  "",
		`{"fontWeight":null}`:                       "",
		`{"fontTracking":{"unit":"POINTS"}}`:        "normal",
		`{"length":{"value":24,"unit":"DIPS"}}`:     "24px",
		`{"fontSize":{"value":16,"unit":"POINTS"}}`: "16px",
		`{"shape":{"family":"SHAPE_FAMILY_ROUNDED_CORNERS","defaultSize":{"value":12,"unit":"DIPS"}}}`: "12px",
		`{"shape":{"family":"SHAPE_FAMILY_CIRCULAR"}}`:                                                 "calc(infinity * 1px)",
		`{"fontNames":{"values":["Roboto"]}}`:                                                          `"Roboto", sans-serif`,
		`{"opacity":0.08}`:                                                                             "0.08",
		`{"length":{"unit":"DIPS"}}`:                                                                   "",
		`{"undefined":true}`:                                                                           "",
		`{"arbitrary":"red; } body { display:none"}`:                                                   "",
	} {
		if got := themeValue(json.RawMessage(raw)); got != expected {
			t.Errorf("themeValue(%s) = %q, want %q", raw, got, expected)
		}
	}
}

func TestStylesUseMaterialRolesWithoutLegacyAliases(t *testing.T) {
	legacy := regexp.MustCompile(`--(?:sl-(?:ink|page|rule|accent|font)|table-|token-(?:ink|color))`)
	for _, file := range []string{"static/app.css", "static/token-table.css"} {
		css, err := embeddedStatic.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if match := legacy.Find(css); match != nil {
			t.Fatalf("%s still uses legacy styling alias %s; use the Material role directly", file, match)
		}
	}
}

func TestThemeReportsMissingTokensWithoutInventingFallbacks(t *testing.T) {
	server := New(publish.PublishedState{}, "")
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/theme.css", nil))
	if response.Code != 503 || !strings.Contains(response.Body.String(), "md-sys-color-primary") || response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("missing theme tokens must be reported, not replaced with a hard-coded palette")
	}
}

func TestStylesHaveNoSeparatePaletteOrTypeScale(t *testing.T) {
	for _, file := range []string{"static/app.css", "static/token-table.css"} {
		css, err := embeddedStatic.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if regexp.MustCompile(`#[0-9a-fA-F]{3,8}\b|font-size:\s*[0-9.]|border-radius:\s*[0-9.]`).Match(css) {
			t.Errorf("%s contains a hard-coded palette, text size or corner size", file)
		}
	}
}

func TestReadableRolesAndDistinctCardSurface(t *testing.T) {
	app, err := embeddedStatic.ReadFile("static/app.css")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(app), "typo-") || !strings.Contains(string(app), "background:var(--md-sys-color-surface-container-low)") || !strings.Contains(string(app), "background:var(--md-sys-color-surface);") {
		t.Fatal("use Material role names and a card surface separate from the page")
	}
	table, err := embeddedStatic.ReadFile("static/token-table.css")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(table), "font-size:var(--md-sys-typescale-body-large-size)") || !strings.Contains(string(table), "line-height:var(--md-sys-typescale-body-large-line-height)") {
		t.Fatal("primary token content must use the readable body-large role")
	}
	response := httptest.NewRecorder()
	New(publish.PublishedState{}, "").Handler().ServeHTTP(response, httptest.NewRequest("GET", "/changes", nil))
	if strings.Contains(response.Body.String(), "typo-") || !strings.Contains(response.Body.String(), `class="md-typescale-body-medium">Dates show`) {
		t.Fatal("instructions should be readable body text, not tiny metadata")
	}
}

// Opt-in check against the actual served publication, not a mock theme.
func TestThemeFromPublication(t *testing.T) {
	path := os.Getenv("SOURCE_LOOM_THEME_DB")
	if path == "" {
		t.Skip("set SOURCE_LOOM_THEME_DB to validate a real publication")
	}
	state, err := publish.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	css, err := publicationTheme(state)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(css, "--md-sys-color-primary:") || !strings.Contains(css, "--md-sys-typescale-body-large-font:") {
		t.Fatal("publication did not provide color and typography")
	}
	variants := make(map[string]bool)
	for _, theme := range []string{"light", "dark"} {
		for _, contrast := range []string{"default", "medium", "high"} {
			selector := `:root[data-theme="` + theme + `"][data-contrast="` + contrast + `"]`
			start := strings.Index(css, selector)
			if start < 0 {
				t.Fatalf("missing runtime theme %s / %s", theme, contrast)
			}
			block := strings.SplitN(css[start:], "}", 2)[0]
			if !strings.Contains(block, "color-scheme: "+theme) {
				t.Fatal("native controls do not follow the selected theme")
			}
			variants[strings.SplitN(block, "{", 2)[1]] = true
		}
	}
	if len(variants) != 6 {
		t.Fatal("the six appearance settings must use distinct published values")
	}
	response := httptest.NewRecorder()
	New(state, "").Handler().ServeHTTP(response, httptest.NewRequest("GET", "/theme.css", nil))
	if response.Code != 200 || response.Body.String() != css || response.Header().Get("Content-Type") != "text/css; charset=utf-8" {
		t.Fatal("theme endpoint did not serve the current publication's CSS")
	}
	response = httptest.NewRecorder()
	server := New(state, "")
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/changes", nil))
	if !strings.Contains(response.Body.String(), "/theme.css?v="+assetVersion([]byte(css))) || !strings.Contains(response.Body.String(), staticURL("app.css")) {
		t.Fatal("HTML must reference the current theme and stylesheet content versions")
	}
	response = httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/theme.css?v="+assetVersion([]byte(css)), nil)
	server.Handler().ServeHTTP(response, request)
	if response.Header().Get("Cache-Control") != "public, max-age=31536000, immutable" {
		t.Fatal("publication theme should be browser-cached")
	}
	etag := response.Header().Get("ETag")
	request.Header.Set("If-None-Match", etag)
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != 304 || response.Body.Len() != 0 {
		t.Fatal("unchanged theme should not transfer CSS again")
	}
	for _, file := range []string{"roboto-latin.woff2", "roboto-latin-ext.woff2"} {
		font, err := embeddedStatic.ReadFile("static/fonts/" + file)
		if err != nil || len(font) < 4 || string(font[:4]) != "wOF2" {
			t.Fatalf("missing or invalid self-hosted font: %s", file)
		}
	}
}
