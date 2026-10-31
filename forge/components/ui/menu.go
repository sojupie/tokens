package ui

import "github.com/a-h/templ"

// MenuEntryProps describes content slots. Attrs carries command and checked state.
type MenuEntryProps struct {
	Label, Supporting, ShortcutLabel string
	IconOnly                         bool
	Leading, Trailing, Shortcut      templ.Component
	Attrs                            templ.Attributes
}
