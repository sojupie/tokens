package presentation

// One set of paths per browser also works when HTMX replaces the whole browser.
// All references are local: no icon request or JavaScript is needed to draw them.
func TokenTableSymbol(name string) string {
	return `<svg class="tokens-material-symbol" viewBox="0 0 960 960" width="24" height="24" fill="currentColor" aria-hidden="true" focusable="false"><use href="#tokens-table-icon-` + name + `"/></svg>`
}
