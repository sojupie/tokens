package presentation

import "github.com/sojupie/tokens/forge/components/ui"
import "github.com/a-h/templ"

func MaterialSymbol(name string) templ.Component { return ui.Icon(name) }
func MaterialSymbolSVG(name string) string       { return ui.IconSVG(name) }
func MaterialSymbolPath(name string) string      { return ui.IconPath(name) }
