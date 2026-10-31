package web

import "github.com/a-h/templ"
import "github.com/sojupie/tokens/forge/components/ui"

func materialSymbol(name string) templ.Component { return ui.Icon(name) }
func materialSymbolSVG(name string) string       { return ui.IconSVG(name) }
func materialSymbolPath(name string) string      { return ui.IconPath(name) }
