package web

import "strings"

// One set of paths per browser also works when HTMX replaces the whole browser.
// All references are local: no icon request or JavaScript is needed to draw them.
func tokenTableSymbol(name string) string {
	return `<svg class="tokens-material-symbol" viewBox="0 0 960 960" width="24" height="24" fill="currentColor" aria-hidden="true" focusable="false"><use href="#tokens-table-icon-` + name + `"/></svg>`
}

func tokenTableSharedMarkup() string {
	var markup strings.Builder
	markup.WriteString(`<svg width="0" height="0" aria-hidden="true" focusable="false" style="position:absolute;overflow:hidden"><defs>`)
	for _, name := range []string{"search", "visibility", "visibility_off", "grid_view", "table_rows", "unfold_more", "unfold_less", "expand_more", "folder", "folder_open", "check", "info", "close", "camera", "content_copy", "arrow_downward"} {
		markup.WriteString(`<symbol id="tokens-table-icon-` + name + `" viewBox="0 -960 960 960"><path d="` + materialSymbolPaths[name] + `"/></symbol>`)
	}
	markup.WriteString(`</defs></svg><template class="tokens-reference-chrome"><header><h3 class="md-typescale-title-large"></h3><button type="button" class="md-icon-button md-standard-icon-button tokens-icon-button tokens-popover-close" popovertargetaction="hide" aria-label="Close token details">` + tokenTableSymbol("close") + `</button></header><div class="tokens-popup-value"></div><div class="tokens-reference-resolved"><span class="tokens-reference-type">Color</span></div><p class="tokens-reference-status" role="status" aria-live="polite"></p><button type="button" class="md-icon-button md-standard-icon-button tokens-icon-button" data-copy-reference>` + tokenTableSymbol("content_copy") + `</button></template>`)
	return markup.String()
}
