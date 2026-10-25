package web

import (
	"context"
	"fmt"
	"github.com/a-h/templ"
	"net/http"
)

func (c componentCase) CarouselItem(index int) templ.Attributes {
	attrs := templ.Attributes{"class": "md-carousel-item", "type": "button", "aria-label": fmt.Sprintf("Item %d of 6", index+1), "data-carousel-index": fmt.Sprint(index)}
	if c.State != "interactive" && index == 0 {
		attrs["data-component-state"] = c.State
		if c.State == "disabled" {
			attrs["disabled"] = true
		}
	}
	return attrs
}

func (server *Server) handleCarouselItems(response http.ResponseWriter, request *http.Request) {
	info := pageMetadata{Title: "Carousel items", NoIndex: true}
	request = request.WithContext(context.WithValue(request.Context(), pageMetadataKey{}, info))
	if err := server.renderHTML(response, request, carouselItemsPage()); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
}
