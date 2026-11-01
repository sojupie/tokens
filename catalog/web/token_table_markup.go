package web

import (
	"context"
	"strings"
)

// One set of paths per browser also works when HTMX replaces the whole browser.
// All references are local: no icon request or JavaScript is needed to draw them.
func tokenTableSymbol(name string) string {
	return `<svg class="tokens-material-symbol" viewBox="0 0 960 960" width="24" height="24" fill="currentColor" aria-hidden="true" focusable="false"><use href="#tokens-table-icon-` + name + `"/></svg>`
}

func tokenTableSharedMarkup() string {
	var markup strings.Builder
	markup.WriteString(`<svg width="0" height="0" aria-hidden="true" focusable="false" style="position:absolute;overflow:hidden"><defs>`)
	for _, name := range []string{"search", "visibility", "visibility_off", "grid_view", "table_rows", "unfold_more", "unfold_less", "expand_more", "folder", "folder_open", "check", "info", "close", "camera", "content_copy", "arrow_downward"} {
		markup.WriteString(`<symbol id="tokens-table-icon-` + name + `" viewBox="0 -960 960 960"><path d="` + materialSymbolPath(name) + `"/></symbol>`)
	}
	markup.WriteString(`</defs></svg>`)
	if err := tokenReferenceChrome().Render(context.Background(), &markup); err != nil {
		panic(err)
	}
	return markup.String()
}
