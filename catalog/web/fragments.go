package web

import (
	"context"
	"crypto/sha256"
	"fmt"
	"html"
	"io"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/sojupie/tokens/source/publish"
)

func tokenBrowserView(subject publish.SubjectView, browser tokenBrowser) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) error {
		if _, err := io.WriteString(writer, `<div id="tokens-token-browser" class="tokens-token-browser" data-component-assets>`); err != nil {
			return err
		}
		if err := componentAssetScope().Render(templ.WithChildren(ctx, tokenBrowserContents(subject, browser)), writer); err != nil {
			return err
		}
		_, err := io.WriteString(writer, `</div>`)
		return err
	})
}

func tokenBrowserContents(subject publish.SubjectView, browser tokenBrowser) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) error {
		write := func(value string) error {
			_, err := io.WriteString(writer, value)
			return err
		}
		if len(browser.Sets) == 0 {
			return write(`<p class="tokens-empty md-typescale-body-large">No published tokens for this page.</p>`)
		}
		if err := write(tokenTableSharedMarkup() + `<form class="tokens-token-filters" hx-boost="false" hx-target="#tokens-token-browser" hx-swap="outerHTML" hx-sync="this:replace" hx-push-url="true" hx-get="` + html.EscapeString(subjectURL(subject)) + `" method="get" action="` + html.EscapeString(subjectURL(subject)) + `"><input type="hidden" name="set" value="` + html.EscapeString(browser.Selected.Key) + `"><div class="tokens-token-toolbar">`); err != nil {
			return err
		}
		keys := make([]string, 0, len(browser.PreviewQuery))
		for key := range browser.PreviewQuery {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			values := browser.PreviewQuery[key]
			for _, value := range values {
				if err := write(`<input type="hidden" name="` + html.EscapeString(key) + `" value="` + html.EscapeString(value) + `">`); err != nil {
					return err
				}
			}
		}
		if err := tokenSetPicker(subject, browser).Render(ctx, writer); err != nil {
			return err
		}
		if err := write(`<div class="tokens-token-actions"><label class="tokens-token-search">` + tableIcon("search") + `<span class="tokens-visually-hidden">Search tokens</span><input id="tokens-token-search" type="search" name="q" value="` + html.EscapeString(browser.Query) + `" placeholder="Search"></label><div class="tokens-token-tools" role="group" aria-label="Table controls">`); err != nil {
			return err
		}
		if err := tableTools().Render(ctx, writer); err != nil {
			return err
		}
		if err := write(`</div></div></div><div class="tokens-token-columns"><span>Token</span>`); err != nil {
			return err
		}
		if len(browser.Filters) > 0 {
			if err := useComponentAssets("dialog").Render(ctx, writer); err != nil {
				return err
			}
			if err := write(`<button class="md-filter-chip tokens-context-button" type="button" data-action="context" aria-haspopup="dialog" aria-expanded="false" aria-controls="tokens-context-picker">` + `<span class="tokens-context-label" title="` + html.EscapeString(tableContextLabel(browser.Filters)) + `">` + html.EscapeString(tableContextLabel(browser.Filters)) + `</span>` + `<span data-slot="trailing-icon" aria-hidden="true">` + tableIcon("dropdown") + `</span></button></div>`); err != nil {
				return err
			}
			if err := tableContextDialog(browser.Filters, subjectURL(subject)).Render(ctx, writer); err != nil {
				return err
			}

		} else if err := write(`<span>Value</span></div>`); err != nil {
			return err
		}
		if err := write(`<noscript><button class="md-button md-filled-button" type="submit">Apply filters</button></noscript></form>`); err != nil {
			return err
		}
		if browser.Count == 0 {
			return write(`<p class="tokens-empty md-typescale-body-large">No tokens match these filters.</p>`)
		}
		if err := write(`<div class="tokens-token-groups">`); err != nil {
			return err
		}
		for _, group := range browser.Selected.Groups {
			if err := writeTokenGroup(writer, group, true); err != nil {
				return err
			}
		}
		return write(`</div>`)
	})
}

func tableContextLabel(filters []tokenFilter) string {
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

func tableIcon(name string) string {
	symbols := map[string]string{
		"search": "search", "eye": "visibility", "grid": "grid_view",
		"expand": "unfold_more", "dropdown": "expand_more", "folder": "folder", "folder-open": "folder_open",
	}
	return tokenTableSymbol(symbols[name])
}

func subjectPageComponent(subject publish.SubjectView, changes []publish.ChangeView, browser tokenBrowser) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) error {
		content := templ.ComponentFunc(func(ctx context.Context, writer io.Writer) error {
			write := func(value string) error {
				_, err := io.WriteString(writer, value)
				return err
			}
			heading := subject.Name
			if metadata(ctx).Heading != "" {
				heading = metadata(ctx).Heading
			}
			if err := write(`<header class="tokens-subject-header"><p class="tokens-back-link"><a class="tokens-link" href="` + subjectListPath(subject.Key.Kind != publish.SubjectComponent) + `">` + materialSymbolSVG("arrow_back") + html.EscapeString(subjectKindLabel(subject.Key.Kind)) + `</a></p><div class="tokens-subject-heading"><h1 class="md-typescale-headline-medium">` + html.EscapeString(heading) + `</h1>`); err != nil {
				return err
			}
			if err := siteDownloadButton().Render(ctx, writer); err != nil {
				return err
			}
			if err := write(`</div><nav class="tokens-subject-actions" aria-label="Component links"><a class="tokens-link" href="` + html.EscapeString(subjectChangesURL(subject)) + `">Change history (` + strconv.Itoa(len(changes)) + `)</a>`); err != nil {
				return err
			}
			if len(subject.SourcePageURLs) > 0 {
				if sourceURL := safeSourceDocumentationURL(subject.SourcePageURLs[0]); sourceURL != "" {
					if err := write(`<a class="tokens-link" href="` + html.EscapeString(sourceURL) + `">Material Design 3 documentation ` + materialSymbolSVG("open_in_new") + `</a>`); err != nil {
						return err
					}
				}
			}
			if err := write(`</nav></header>`); err != nil {
				return err
			}
			if image := metadata(ctx).Illustration; image != nil {
				if err := subjectIllustration(*image).Render(ctx, writer); err != nil {
					return err
				}
			}
			if browser.Preview != nil {
				if err := componentPreviewView(*browser.Preview).Render(ctx, writer); err != nil {
					return err
				}
			}
			if err := write(`<section class="tokens-section" aria-labelledby="tokens-heading"><h2 class="md-typescale-title-medium" id="tokens-heading">Browse tokens</h2>`); err != nil {
				return err
			}
			if err := tokenBrowserView(subject, browser).Render(ctx, writer); err != nil {
				return err
			}
			_, err := io.WriteString(writer, `</section>`)
			return err
		})
		return layout(subject.Name).Render(templ.WithChildren(ctx, content), writer)
	})
}

func safeSourceDocumentationURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Host != "m3.material.io" || parsed.User != nil {
		return ""
	}
	return raw
}

func writeTokenGroup(writer io.Writer, group tokenGroup, state bool) error {
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
		if err := write(`<details class="` + className + `"` + open + `><summary>` + `<span class="tokens-folder-closed">` + tableIcon("folder") + `</span><span class="tokens-folder-open">` + tableIcon("folder-open") + `</span><span>` + html.EscapeString(group.Name) + `</span>` + tableIcon("dropdown") + `</summary>`); err != nil {
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
			if err := writeTokenReferences(writer, token); err != nil {
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
		if err := writeTokenGroup(writer, child, false); err != nil {
			return err
		}
	}
	if group.Name != "" {
		return write(`</details>`)
	}
	return nil
}

func writeTokenReferences(writer io.Writer, token tokenRow) error {
	write := func(value string) error { _, err := io.WriteString(writer, value); return err }
	id := fmt.Sprintf("token-info-%x", sha256.Sum256([]byte(token.SourceName+"\x00"+token.Name+"\x00"+token.Value)))
	label := token.Value
	if label == "" || label == "Color not resolved" || label == "Undefined: true" {
		if len(token.References) > 0 {
			label = token.References[0].Name
		} else {
			label = "—"
		}
	}
	style := ""
	if token.Color != "" {
		style = ` style="--tokens-preview-color:` + html.EscapeString(token.Color) + `;--tokens-preview-on-color:` + tokenColorTextColor(token.Color) + `"`
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
	if err := write(`<div class="tokens-value-cell"><button type="button" class="tokens-large-preview` + colorCellClass(token.Color) + `"` + style + ` popovertarget="` + id + `" aria-label="View ` + html.EscapeString(token.Name) + ` details">` + preview + `<span class="tokens-preview-name" aria-hidden="true">` + html.EscapeString(token.Name) + `</span></button>` + compact + `<button type="button" class="md-icon-button md-standard-icon-button tokens-icon-button tokens-info-button" popovertarget="` + id + `" aria-label="References for ` + html.EscapeString(token.Name) + `">` + tokenTableSymbol("info") + `</button></div>`); err != nil {
		return err
	}
	// Keep the entire reference tree in the document. Duplicate previews and
	// JavaScript-only controls are added from the shared template on first open.
	if err := write(`<div id="` + id + `" class="md-elevation tokens-reference-popover" popover role="dialog" aria-label="` + html.EscapeString(token.Name) + `"><div class="tokens-reference-chain"><code class="tokens-popup-source">` + html.EscapeString(token.SourceName) + `</code>`); err != nil {
		return err
	}
	if err := writeReferenceBranches(writer, token.References); err != nil {
		return err
	}
	return write(`</div>`)
}

func writeReferenceBranches(writer io.Writer, references []publish.TokenReference) error {
	if _, err := io.WriteString(writer, `<ul>`); err != nil {
		return err
	}
	for _, reference := range references {
		name := html.EscapeString(reference.Name)
		if _, err := io.WriteString(writer, `<li><div class="tokens-reference-node">`+tokenTableSymbol("camera")+`<code title="`+name+`">`+name+`</code></div>`+tokenTableSymbol("arrow_downward")); err != nil {
			return err
		}
		if len(reference.References) > 0 {
			if err := writeReferenceBranches(writer, reference.References); err != nil {
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
