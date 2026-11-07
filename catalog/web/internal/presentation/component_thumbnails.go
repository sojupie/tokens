package presentation

import (
	"github.com/a-h/templ"
)

// Thumbnail composition chooses variants, never component geometry or token values.
func (c ComponentCase) ThumbnailVariant(variant string) ComponentCase {
	c.Variant = variant
	return c
}
func (c ComponentCase) ThumbnailSelected(selected bool) ComponentCase {
	c.Selected = selected
	return c
}

func (c ComponentCase) ThumbnailToolbarAction(index int) templ.Attributes {
	attrs := c.BarAction(min(index, 2))
	attrs["aria-label"] = []string{"Download", "Copy", "More actions"}[index]
	delete(attrs, "aria-pressed")
	delete(attrs, "data-button-toggle")
	delete(attrs, "data-menu-open")
	delete(attrs, "aria-haspopup")
	delete(attrs, "aria-expanded")
	return attrs
}
