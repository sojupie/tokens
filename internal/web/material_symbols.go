package web

import "github.com/a-h/templ"

// Material Symbols Outlined, 24px, by Google (Apache-2.0).
// Source: https://github.com/google/material-design-icons/tree/master/symbols/web
// Paths are embedded so icons do not need a font or an external request.
func materialSymbol(name string) templ.Component {
	return templ.Raw(materialSymbolSVG(name))
}

func materialSymbolSVG(name string) string {
	paths := map[string]string{
		"search":        "M784-120 532-372q-30 24-69 38t-83 14q-109 0-184.5-75.5T120-580q0-109 75.5-184.5T380-840q109 0 184.5 75.5T640-580q0 44-14 83t-38 69l252 252-56 56ZM380-400q75 0 127.5-52.5T560-580q0-75-52.5-127.5T380-760q-75 0-127.5 52.5T200-580q0 75 52.5 127.5T380-400Z",
		"visibility":    "M480-320q75 0 127.5-52.5T660-500q0-75-52.5-127.5T480-680q-75 0-127.5 52.5T300-500q0 75 52.5 127.5T480-320Zm0-72q-45 0-76.5-31.5T372-500q0-45 31.5-76.5T480-608q45 0 76.5 31.5T588-500q0 45-31.5 76.5T480-392Zm0 192q-146 0-266-81.5T40-500q54-137 174-218.5T480-800q146 0 266 81.5T920-500q-54 137-174 218.5T480-200Zm0-300Zm0 220q113 0 207.5-59.5T832-500q-50-101-144.5-160.5T480-720q-113 0-207.5 59.5T128-500q50 101 144.5 160.5T480-280Z",
		"grid_view":     "M120-520v-320h320v320H120Zm0 400v-320h320v320H120Zm400-400v-320h320v320H520Zm0 400v-320h320v320H520ZM200-600h160v-160H200v160Zm400 0h160v-160H600v160Zm0 400h160v-160H600v160Zm-400 0h160v-160H200v160Zm400-400Zm0 240Zm-240 0Zm0-240Z",
		"unfold_more":   "M480-120 300-300l58-58 122 122 122-122 58 58-180 180ZM358-598l-58-58 180-180 180 180-58 58-122-122-122 122Z",
		"folder":        "M160-160q-33 0-56.5-23.5T80-240v-480q0-33 23.5-56.5T160-800h240l80 80h320q33 0 56.5 23.5T880-640v400q0 33-23.5 56.5T800-160H160Zm0-80h640v-400H447l-80-80H160v480Zm0 0v-480 480Z",
		"arrow_back":    "m313-440 224 224-57 56-320-320 320-320 57 56-224 224h487v80H313Z",
		"open_in_new":   "M200-120q-33 0-56.5-23.5T120-200v-560q0-33 23.5-56.5T200-840h280v80H200v560h560v-280h80v280q0 33-23.5 56.5T760-120H200Zm188-212-56-56 372-372H560v-80h280v280h-80v-144L388-332Z",
		"check":         "M382-240 154-468l57-57 171 171 367-367 57 57-424 424Z",
		"expand_more":   "M480-345 240-585l56-56 184 184 184-184 56 56-240 240Z",
		"chevron_right": "M504-480 320-664l56-56 240 240-240 240-56-56 184-184Z",
	}
	path, ok := paths[name]
	if !ok {
		return ""
	}
	return `<svg class="sl-material-symbol" xmlns="http://www.w3.org/2000/svg" viewBox="0 -960 960 960" width="24" height="24" fill="currentColor" aria-hidden="true" focusable="false"><path d="` + path + `"/></svg>`
}
