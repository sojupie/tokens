package ui

import "github.com/a-h/templ"

type CardProps struct {
	Variant                                  string
	Headline, Supporting, ImageSrc, ImageAlt string
	Actions                                  templ.Component
	Attrs                                    templ.Attributes
}
