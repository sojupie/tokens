package model

import "github.com/a-h/templ"

type CardProps struct {
	Variant                                  string
	Headline, Supporting, ImageSrc, ImageAlt string
	Actions                                  templ.Component
	Attrs                                    templ.Attributes
}

func (p CardProps) Attributes() templ.Attributes {
	return Attributes(templ.Attributes{"class": "md-card md-" + Fallback(p.Variant, "filled") + "-card"}, p.Attrs)
}
