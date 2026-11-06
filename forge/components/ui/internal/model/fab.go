package model

import (
	"github.com/a-h/templ"
)

type FABProps struct {
	Label, Icon, Variant, Width string
	Size                        Size
	Extended, Disabled          bool
	Attrs                       templ.Attributes
}

func (p FABProps) Attributes() templ.Attributes {
	class := "md-button md-fab"
	if p.Extended {
		class = "md-button md-extended-fab"
	}
	return Attributes(templ.Attributes{"class": class, "type": "button", "aria-label": p.Label, "data-variant": Fallback(p.Variant, "primary-container"), "data-size": Fallback(string(p.Size), "default"), "data-width": Fallback(p.Width, "default"), "disabled": p.Disabled}, p.Attrs)
}
