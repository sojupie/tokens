package web

import (
	"bytes"
	"context"
	"net/url"
	"strings"
	"testing"
)

func TestCarouselActionConfigurations(t *testing.T) {
	for _, action := range []string{"none", "show-all", "header-arrow"} {
		for _, variant := range []string{"hero", "full-screen"} {
			page := &componentPreview{Cases: []componentCase{{Family: "carousel", Variant: variant, State: "interactive"}}}
			if err := configureCarouselPreview(page, url.Values{"preview.action": {action}}); err != nil {
				t.Fatal(err)
			}
			var out bytes.Buffer
			if err := specimenCarousel(page.Cases[0]).Render(context.Background(), &out); err != nil {
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
	page := &componentPreview{Cases: []componentCase{{Family: "carousel"}}}
	if err := configureCarouselPreview(page, url.Values{"preview.action": {"invalid"}}); err == nil {
		t.Fatal("invalid action accepted")
	}
	if err := configureCarouselPreview(page, nil); err != nil || page.Cases[0].CarouselAction != "show-all" {
		t.Fatal("default action changed")
	}
}
