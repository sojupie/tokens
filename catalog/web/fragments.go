package web

import (
	"context"
	"crypto/sha256"
	"fmt"
	"html"
	"io"
	"net/url"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/sojupie/tokens/source/publish"
)

func tokenBrowserView(subject publish.SubjectView, browser tokenBrowser) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) error {
		write := func(value string) error {
			_, err := io.WriteString(writer, value)
			return err
		}
		if len(browser.Sets) == 0 {
			return write(`<p class="tokens-empty md-typescale-body-large">No published tokens for this page.</p>`)
		}
		if err := write(`<div class="tokens-token-browser"><form class="tokens-token-filters" method="get" action="` + html.EscapeString(subjectURL(subject)) + `"><input type="hidden" name="set" value="` + html.EscapeString(browser.Selected.Key) + `"><div class="tokens-token-toolbar">`); err != nil {
			return err
		}
		if err := tokenSetPicker(subject, browser).Render(ctx, writer); err != nil {
			return err
		}
		if err := write(`<label class="tokens-token-search">` + tableIcon("search") + `<span class="tokens-visually-hidden">Search tokens</span><input type="search" name="q" value="` + html.EscapeString(browser.Query) + `" placeholder="Search"></label><div class="tokens-token-tools" role="group" aria-label="Table display"><button class="tokens-icon-button" type="button" data-action="previews" aria-label="Hide previews" aria-pressed="true">` + `<span data-icon="on">` + tableIcon("eye") + `</span><span data-icon="off" hidden>` + materialSymbolSVG("visibility_off") + `</span></button><button class="tokens-icon-button" type="button" data-action="view" aria-label="Show grid view" aria-pressed="false">` + `<span data-icon="on">` + tableIcon("grid") + `</span><span data-icon="off" hidden>` + materialSymbolSVG("table_rows") + `</span></button><button class="tokens-icon-button" type="button" data-action="expand" aria-label="Expand all groups" aria-expanded="false">` + `<span data-icon="on">` + tableIcon("expand") + `</span><span data-icon="off" hidden>` + materialSymbolSVG("unfold_less") + `</span></button></div></div><div class="tokens-token-columns"><span>Token</span>`); err != nil {
			return err
		}
		if len(browser.Filters) > 0 {
			if err := write(`<button class="tokens-context-button" type="button" data-action="context" aria-haspopup="dialog">` + `<span class="tokens-context-label" title="` + html.EscapeString(tableContextLabel(browser.Filters)) + `">` + html.EscapeString(tableContextLabel(browser.Filters)) + `</span>` + tableIcon("dropdown") + `</button></div><dialog class="tokens-context-dialog" aria-labelledby="tokens-context-title"><h2 id="tokens-context-title">See more values</h2><p>Compare token values by selecting another context</p><div class="tokens-context-fields">`); err != nil {
				return err
			}
			advanced := false
			for _, filter := range browser.Filters {
				if !advanced && filter.Name != "Theme" && filter.Name != "Contrast" {
					if err := write(`<details class="tokens-context-advanced"><summary>More contexts</summary>`); err != nil {
						return err
					}
					advanced = true
				}
				if err := write(`<fieldset><legend>` + html.EscapeString(filter.Name) + `</legend><div class="tokens-context-options">`); err != nil {
					return err
				}
				options := filter.Options
				for _, option := range options {
					selected := ""
					if option == filter.Selected {
						selected = ` checked`
					}
					label := option
					if err := write(`<label class="tokens-context-choice"><input class="tokens-visually-hidden" type="radio" name="` + html.EscapeString(filter.Key) + `" value="` + html.EscapeString(option) + `"` + selected + `><span>` + materialSymbolSVG("check") + html.EscapeString(label) + `</span></label>`); err != nil {
						return err
					}
				}
				if err := write(`</div></fieldset>`); err != nil {
					return err
				}
				if filter.Name == "Audience" {
					if err := write(`<details class="tokens-audience-help"><summary>What do 1P and 3P mean?</summary><p>1P means first-party; 3P means third-party. We interpret <strong>1P Baseline</strong> as the configuration for Google&#39;s own products, and <strong>3P</strong> as the configuration for other developers.</p><p>Material supplies these audience labels but does not define them in the captured data. They are not a baseline-versus-Expressive switch.</p></details>`); err != nil {
						return err
					}
				}
			}
			if advanced {
				if err := write(`</details>`); err != nil {
					return err
				}
			}
			if err := write(`</div><p class="tokens-context-status" role="status" aria-live="polite"></p><footer><button class="md-text-button" type="button" data-action="cancel">Cancel</button><button class="md-text-button" type="submit">Save</button></footer></dialog>`); err != nil {
				return err
			}
		} else if err := write(`<span>Value</span></div>`); err != nil {
			return err
		}
		if err := write(`<noscript><button class="md-filled-button" type="submit">Apply filters</button><style>.tokens-context-dialog:not([open]){display:block;position:static}</style></noscript></form>`); err != nil {
			return err
		}
		if browser.Count == 0 {
			return write(`<p class="tokens-empty md-typescale-body-large">No tokens match these filters.</p></div>`)
		}
		if err := write(`<div class="tokens-token-groups">`); err != nil {
			return err
		}
		for _, group := range browser.Selected.Groups {
			if err := writeTokenGroup(writer, group, true); err != nil {
				return err
			}
		}
		return write(`</div></div>`)
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
	return materialSymbolSVG(symbols[name])
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
			if err := write(`<p class="tokens-back-link"><a class="tokens-link" href="` + subjectListPath(subject.Key.Kind != publish.SubjectComponent) + `">` + materialSymbolSVG("arrow_back") + html.EscapeString(subjectKindLabel(subject.Key.Kind)) + `</a></p><header class="tokens-subject-heading"><h1 class="md-typescale-headline-medium">` + html.EscapeString(heading) + `</h1><nav class="tokens-subject-actions" aria-label="Component links">`); err != nil {
				return err
			}
			if err := writeTokenDownloadMenu(writer, subject, browser, metadata(ctx).SiteDownload); err != nil {
				return err
			}
			if err := write(`<a class="tokens-link" href="` + html.EscapeString(subjectChangesURL(subject)) + `">Change history (` + strconv.Itoa(len(changes)) + `)</a>`); err != nil {
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
		preview = `<span class="tokens-elevation-swatch" style="--_elevation-level:` + html.EscapeString(token.Elevation) + `" aria-hidden="true"></span><span class="tokens-visually-hidden">` + html.EscapeString(label) + `</span>`
	}
	compact := `<span class="tokens-compact-value"` + style + `>`
	if token.Color != "" {
		compact += `<span class="tokens-color-dot" aria-hidden="true"></span>`
	}
	compact += html.EscapeString(label) + `</span>`
	if err := write(`<div class="tokens-value-cell"><button type="button" class="tokens-large-preview` + colorCellClass(token.Color) + `"` + style + ` popovertarget="` + id + `" aria-label="View ` + html.EscapeString(token.Name) + ` details">` + preview + `<span class="tokens-preview-name" aria-hidden="true">` + html.EscapeString(token.Name) + `</span></button>` + compact + `<button type="button" class="tokens-icon-button tokens-info-button" popovertarget="` + id + `" aria-label="References for ` + html.EscapeString(token.Name) + `">` + materialSymbolSVG("info") + `</button></div>`); err != nil {
		return err
	}
	if err := write(`<div id="` + id + `" class="tokens-reference-popover" popover role="dialog" aria-labelledby="` + id + `-title"><header><h3 id="` + id + `-title" class="md-typescale-title-large" title="` + html.EscapeString(token.Name) + `">` + html.EscapeString(token.Name) + `</h3><button type="button" class="tokens-icon-button tokens-popover-close" popovertarget="` + id + `" popovertargetaction="hide" aria-label="Close token details">` + materialSymbolSVG("close") + `</button></header><div class="tokens-popup-value">` + preview + `</div><div class="tokens-reference-chain"><code class="tokens-popup-source">` + html.EscapeString(token.SourceName) + `</code>`); err != nil {
		return err
	}
	if err := writeReferenceBranches(writer, token.References); err != nil {
		return err
	}
	kind := ""
	if token.Color != "" {
		kind = `<span class="tokens-reference-type">Color</span>`
	}
	return write(`<div class="tokens-reference-resolved">` + kind + compact + `</div></div><p class="tokens-reference-status" role="status" aria-live="polite"></p></div>`)
}

func writeReferenceBranches(writer io.Writer, references []publish.TokenReference) error {
	if _, err := io.WriteString(writer, `<ul>`); err != nil {
		return err
	}
	for _, reference := range references {
		name := html.EscapeString(reference.Name)
		if _, err := io.WriteString(writer, `<li><div class="tokens-reference-node">`+materialSymbolSVG("camera")+`<code title="`+name+`">`+name+`</code><button type="button" class="tokens-icon-button" data-copy-reference="`+name+`" aria-label="Copy `+name+`">`+materialSymbolSVG("content_copy")+`</button></div>`+materialSymbolSVG("arrow_downward")); err != nil {
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
