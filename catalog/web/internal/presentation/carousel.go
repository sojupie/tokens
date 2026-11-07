package presentation

import (
	"fmt"

	"github.com/a-h/templ"
)

func (c ComponentCase) CarouselItem(index int) templ.Attributes {
	attrs := templ.Attributes{"class": "md-carousel-item", "type": "button", "data-carousel-index": fmt.Sprint(index)}
	if c.State != "interactive" && index == 0 {
		attrs["data-component-state"] = c.State
		if c.State == "disabled" {
			attrs["disabled"] = true
		}
	}
	return attrs
}

// Specimen ratios accommodate each painting; the portrait keeps its source proportions.
func (c ComponentCase) CarouselSlot(index int) templ.Attributes {
	attrs := templ.Attributes{"class": "md-carousel-slot"}
	if c.Variant == "uncontained-multi-aspect-ratio" {
		ratios := []float64{16.0 / 9, 968.0 / 1200, 1, 3.0 / 4}
		attrs["style"] = fmt.Sprintf("--_carousel-aspect-ratio:%g", ratios[index%len(ratios)])
	}
	return attrs
}

func (page ComponentPreview) HasFullScreenCarousel() bool {
	for _, c := range page.Cases {
		if c.Family == "carousel" && c.Variant == "full-screen" {
			return true
		}
	}
	return false
}

func (page ComponentPreview) HasUncontainedCarousel() bool {
	for _, c := range page.Cases {
		if c.Family == "carousel" && (c.Variant == "uncontained" || c.Variant == "uncontained-multi-aspect-ratio") {
			return true
		}
	}
	return false
}

// Brief specimen labels; the full artwork titles remain the accessible labels.
func CarouselCaption(index int) string {
	return []string{"Trouville", "Woman with a Parasol", "Argenteuil", "Pourville", "Étretat", "The Louvre"}[index%6]
}
