package model

import "github.com/a-h/templ"

type Option struct {
	Value, Label            string
	Disabled, DividerBefore bool
	Leading                 templ.Component
	Attrs                   templ.Attributes
}
