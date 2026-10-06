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

func TestThemeReportsMissingTokensWithoutInventingFallbacks(t *testing.T) {
	server := New(publish.PublishedState{}, "")
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/theme.css", nil))
	if response.Code != 503 || !strings.Contains(response.Body.String(), "md-sys-color-primary") || response.Header().Get("Cache-Control") != "no-cache" {
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
	response := httptest.NewRecorder()
	New(state, "").Handler().ServeHTTP(response, httptest.NewRequest("GET", "/theme.css", nil))
	if response.Code != 200 || response.Body.String() != css || response.Header().Get("Content-Type") != "text/css; charset=utf-8" {
		t.Fatal("theme endpoint did not serve the current publication's CSS")
	}
	for _, file := range []string{"roboto-latin.woff2", "roboto-latin-ext.woff2"} {
		font, err := embeddedStatic.ReadFile("static/fonts/" + file)
		if err != nil || len(font) < 4 || string(font[:4]) != "wOF2" {
			t.Fatalf("missing or invalid self-hosted font: %s", file)
		}
	}
}
