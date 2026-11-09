package presentation

import (
	"crypto/sha256"
	"fmt"
	"html"
	"io"
	"net/url"
	"strings"

	"github.com/sojupie/tokens/source/publish"
)

func TableContextLabel(filters []TokenFilter) string {
	labels := make([]string, 0, len(filters))
	for _, name := range []string{"Contrast", "Theme", "Platform", "Audience"} {
		for _, filter := range filters {
			if filter.Name == name && filter.Selected != "" {
				labels = append(labels, filter.Selected)
			}
		}
	}
	if len(labels) == 0 {
		return "Token default"
	}
	return strings.Join(labels, ", ")
}

func TableIcon(name string) string {
	symbols := map[string]string{
		"search": "search", "eye": "visibility", "grid": "grid_view",
		"expand": "unfold_more", "dropdown": "expand_more", "folder": "folder", "folder-open": "folder_open",
	}
	return TokenTableSymbol(symbols[name])
}

func SafeSourceDocumentationURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Host != "m3.material.io" || parsed.User != nil {
		return ""
	}
	return raw
}

func WriteTokenGroups(writer io.Writer, browser TokenBrowser) error {
	if len(browser.Sets) == 0 {
		_, err := io.WriteString(writer, `<p class="tokens-empty md-typescale-body-large">No published tokens for this page.</p>`)
		return err
	}
	if browser.Count == 0 {
		_, err := io.WriteString(writer, `<p class="tokens-empty md-typescale-body-large">No tokens match these filters.</p>`)
		return err
	}
	if _, err := io.WriteString(writer, `<div class="tokens-token-groups">`); err != nil {
		return err
	}
	for _, group := range browser.Selected.Groups {
		if err := WriteTokenGroup(writer, group, true); err != nil {
			return err
		}
	}
	_, err := io.WriteString(writer, `</div>`)
	return err
}

func WriteTokenGroup(writer io.Writer, group TokenGroup, state bool) error {
	write := func(value string) error {
		_, err := io.WriteString(writer, value)
		return err
	}
	className := "tokens-token-folder"
	if !state {
		className += " tokens-token-subfolder"
	}
	open := ""
	if group.Open {
		open = ` open`
	}
	if group.Name != "" {
		if err := write(`<details class="` + className + `"` + open + `><summary>` + `<span class="tokens-folder-closed">` + TableIcon("folder") + `</span><span class="tokens-folder-open">` + TableIcon("folder-open") + `</span><span>` + html.EscapeString(group.Name) + `</span>` + TableIcon("dropdown") + `</summary>`); err != nil {
			return err
		}
	}
	if len(group.Tokens) > 0 {
		if err := write(`<div class="tokens-token-table-wrap"><table class="tokens-token-table"><caption class="tokens-visually-hidden">` + html.EscapeString(strings.TrimSpace(group.Name+" tokens")) + `</caption><colgroup><col><col class="tokens-token-value-column"></colgroup><thead class="tokens-visually-hidden"><tr><th scope="col">Token</th><th scope="col">Value</th></tr></thead><tbody>`); err != nil {
			return err
		}
		for _, token := range group.Tokens {
			if err := write(`<tr><th scope="row"><span class="tokens-token-description">` + html.EscapeString(token.Name) + `</span><code class="tokens-token-source-name">` + html.EscapeString(token.SourceName) + `</code>`); err != nil {
				return err
			}
			if token.Lifecycle != "" {
				if err := write(`<span class="tokens-token-lifecycle">` + html.EscapeString(token.Lifecycle) + `</span>`); err != nil {
					return err
				}
			}
			cell := `</th><td>`
			if !token.HasValue {
				cell = `</th><td class="tokens-token-no-value">`
			}
			if err := write(cell); err != nil {
				return err
			}
			if err := WriteTokenReferences(writer, token); err != nil {
				return err
			}
			if err := write(`</td></tr>`); err != nil {
				return err
			}
		}
		if err := write(`</tbody></table></div>`); err != nil {
			return err
		}
	}
	for _, child := range group.Children {
		if err := WriteTokenGroup(writer, child, false); err != nil {
			return err
		}
	}
	if group.Name != "" {
		return write(`</details>`)
	}
	return nil
}

func WriteTokenReferences(writer io.Writer, token TokenRow) error {
	write := func(value string) error { _, err := io.WriteString(writer, value); return err }
	id := fmt.Sprintf("token-info-%x", sha256.Sum256([]byte(token.SourceName+"\x00"+token.Name+"\x00"+token.Value)))
	label := token.Value
	if label == "" || label == "Color not resolved" || label == "Undefined: true" {
		if len(token.References) > 0 {
			label = token.References[0].Name
		} else if label != "Color not resolved" {
			label = "No captured value"
		}
	}
	style := ""
	if token.Color != "" {
		style = ` style="--tokens-preview-color:` + html.EscapeString(token.Color) + `;--tokens-preview-on-color:` + TokenColorTextColor(token.Color) + `"`
	}
	preview := `<span class="tokens-preview-value">` + html.EscapeString(label) + `</span>`
	if token.Shape != "" {
		preview = `<span class="tokens-shape-swatch" style="border-radius:` + html.EscapeString(token.Shape) + `" aria-hidden="true"></span><span class="tokens-preview-value">` + html.EscapeString(label) + `</span>`
	}
	if token.Elevation != "" {
		preview = `<span class="md-elevation tokens-elevation-swatch" style="--_elevation-level:` + html.EscapeString(token.Elevation) + `" aria-hidden="true"></span><span class="tokens-visually-hidden">` + html.EscapeString(label) + `</span>`
	}
	compact := `<span class="tokens-compact-value"` + style + `>`
	if token.Color != "" {
		compact += `<span class="tokens-color-dot" aria-hidden="true"></span>`
	}
	compact += html.EscapeString(label) + `</span>`
	if err := write(`<div class="tokens-value-cell"><button type="button" class="tokens-large-preview` + ColorCellClass(token.Color) + `"` + style + ` popovertarget="` + id + `" aria-label="View ` + html.EscapeString(token.Name) + ` details">` + preview + `<span class="tokens-preview-name" aria-hidden="true">` + html.EscapeString(token.Name) + `</span></button>` + compact + `<button type="button" class="md-icon-button md-standard-icon-button tokens-icon-button tokens-info-button" popovertarget="` + id + `" aria-label="References for ` + html.EscapeString(token.Name) + `">` + TokenTableSymbol("info") + `</button></div>`); err != nil {
		return err
	}

	if err := write(`<div id="` + id + `" class="md-elevation tokens-reference-popover" popover role="dialog" aria-label="` + html.EscapeString(token.Name) + `"><div class="tokens-reference-chain"><code class="tokens-popup-source">` + html.EscapeString(token.SourceName) + `</code>`); err != nil {
		return err
	}
	if err := WriteReferenceBranches(writer, token.References); err != nil {
		return err
	}
	return write(`</div>`)
}

func WriteReferenceBranches(writer io.Writer, references []publish.TokenReference) error {
	if _, err := io.WriteString(writer, `<ul>`); err != nil {
		return err
	}
	for _, reference := range references {
		name := html.EscapeString(reference.Name)
		if _, err := io.WriteString(writer, `<li><div class="tokens-reference-node">`+TokenTableSymbol("camera")+`<code title="`+name+`">`+name+`</code></div>`+TokenTableSymbol("arrow_downward")); err != nil {
			return err
		}
		if len(reference.References) > 0 {
			if err := WriteReferenceBranches(writer, reference.References); err != nil {
				return err
			}
		}
		if _, err := io.WriteString(writer, `</li>`); err != nil {
			return err
		}
	}
	_, err := io.WriteString(writer, `</ul>`)
	return err
}
