package web

import "github.com/a-h/templ"

// Material Symbols Outlined, 24px, by Google (Apache-2.0).
// Source: https://github.com/google/material-design-icons/tree/master/symbols/web
// Paths are embedded so icons do not need a font or an external request.
func materialSymbol(name string) templ.Component {
	paths := map[string]string{
		"expand_more":   "M480-345 240-585l56-56 184 184 184-184 56 56-240 240Z",
		"chevron_right": "M504-480 320-664l56-56 240 240-240 240-56-56 184-184Z",
	}
	path, ok := paths[name]
	if !ok {
		return templ.Raw("")
	}
	return templ.Raw(`<svg class="sl-material-symbol" xmlns="http://www.w3.org/2000/svg" viewBox="0 -960 960 960" width="24" height="24" fill="currentColor" aria-hidden="true" focusable="false"><path d="` + path + `"/></svg>`)
}
