package web

import (
	"bytes"
	"context"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCarouselAllItemsUsesSeparatePage(t *testing.T) {
	var out bytes.Buffer
	if err := specimenCarousel(componentCase{Family: "carousel", Variant: "hero", State: "interactive"}).Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `href="/components/carousel/items"`) || strings.Contains(out.String(), `<details`) || strings.Contains(out.String(), `md-carousel-all-items`) {
		t.Fatal("carousel must link to the alternative without rendering it inline")
	}
	server := &Server{}
	response := httptest.NewRecorder()
	server.handleCarouselItems(response, httptest.NewRequest("GET", "/components/carousel/items", nil))
	if response.Code != 200 {
		t.Fatalf("items page: %d %s", response.Code, response.Body.String())
	}
	if strings.Count(response.Body.String(), `role="img"`) != 6 {
		t.Fatal("items page must render every illustration")
	}
}
