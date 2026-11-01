package ui

import "github.com/a-h/templ"

type CardProps struct {
	Variant                                  string
	Headline, Supporting, ImageSrc, ImageAlt string
	Actions                                  templ.Component
	Attrs                                    templ.Attributes
}

func (p CardProps) Attributes() templ.Attributes {
	return attributes(templ.Attributes{"class": "md-card md-" + fallback(p.Variant, "filled") + "-card"}, p.Attrs)
}
