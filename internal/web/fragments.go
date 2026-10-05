package web

import (
	"context"
	"html"
	"io"
	"net/url"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/sojupie/m3-source/internal/publish"
)

func tokenBrowserView(subject publish.SubjectView, browser tokenBrowser) templ.Component {
	return templ.ComponentFunc(func(_ context.Context, writer io.Writer) error {
		write := func(value string) error {
			_, err := io.WriteString(writer, value)
			return err
		}
		if len(browser.Sets) == 0 {
			return write(`<p class="sl-empty typo-sublabel">No published tokens for this page.</p>`)
		}
		if err := write(`<div class="sl-token-browser"><form class="sl-token-filters" method="get" action="` + html.EscapeString(subjectURL(subject)) + `"><div class="sl-token-toolbar"><label class="sl-token-set"><span class="sl-visually-hidden">Token set</span><select name="set">`); err != nil {
			return err
		}
		for _, set := range browser.Sets {
			selected := ""
			if set.Key == browser.Selected.Key {
				selected = ` selected`
			}
			if err := write(`<option value="` + html.EscapeString(set.Key) + `"` + selected + `>` + html.EscapeString(set.Label) + `</option>`); err != nil {
				return err
			}
		}
		if err := write(`</select></label><label class="sl-token-search">` + tableIcon("search") + `<span class="sl-visually-hidden">Search tokens</span><input type="search" name="q" value="` + html.EscapeString(browser.Query) + `" placeholder="Search"></label><button class="sl-icon-button" type="button" data-action="previews" aria-label="Toggle previews" aria-pressed="true">` + tableIcon("eye") + `</button><button class="sl-icon-button" type="button" data-action="view" aria-label="Toggle view" aria-pressed="false">` + tableIcon("grid") + `</button><button class="sl-icon-button" type="button" data-action="expand" aria-label="Expand all groups" aria-expanded="false">` + tableIcon("expand") + `</button></div><div class="sl-token-columns"><span>Token</span>`); err != nil {
			return err
		}
		if len(browser.Filters) > 0 {
			if err := write(`<button class="sl-context-button" type="button" data-action="context" aria-haspopup="dialog">` + html.EscapeString(tableContextLabel(browser.Filters)) + tableIcon("dropdown") + `</button></div><dialog class="sl-context-dialog" aria-labelledby="sl-context-title"><h2 id="sl-context-title">See more values</h2><p>Compare token values by selecting another context</p><div class="sl-context-fields">`); err != nil {
				return err
			}
			advanced := false
			for _, filter := range browser.Filters {
				if !advanced && filter.Name != "Theme" && filter.Name != "Contrast" {
					if err := write(`<details class="sl-context-advanced"><summary>More contexts</summary>`); err != nil {
						return err
					}
					advanced = true
				}
				if err := write(`<fieldset><legend>` + html.EscapeString(filter.Name) + `</legend><div class="sl-context-options">`); err != nil {
					return err
				}
				options := filter.Options
				for _, option := range options {
					selected := ""
					if option == filter.Selected {
						selected = ` checked`
					}
					label := option
					if err := write(`<label><input type="radio" name="` + html.EscapeString(filter.Key) + `" value="` + html.EscapeString(option) + `"` + selected + `><span>` + html.EscapeString(label) + `</span></label>`); err != nil {
						return err
					}
				}
				if err := write(`</div></fieldset>`); err != nil {
					return err
				}
				if filter.Name == "Audience" {
					if err := write(`<details class="sl-audience-help"><summary>What do 1P and 3P mean?</summary><p>1P means first-party; 3P means third-party. We interpret <strong>1P Baseline</strong> as the configuration for Google&#39;s own products, and <strong>3P</strong> as the configuration for other developers.</p><p>Material supplies these audience labels but does not define them in the captured data. They are not a baseline-versus-Expressive switch.</p></details>`); err != nil {
						return err
					}
				}
			}
			if advanced {
				if err := write(`</details>`); err != nil {
					return err
				}
			}
			if err := write(`</div><p class="sl-context-status" role="status" aria-live="polite"></p><footer><button type="button" data-action="cancel">Cancel</button><button type="submit">Save</button></footer></dialog>`); err != nil {
				return err
			}
		} else if err := write(`<span>Value</span></div>`); err != nil {
			return err
		}
		if err := write(`<noscript><button type="submit">Apply filters</button><style>.sl-context-dialog:not([open]){display:block;position:static}</style></noscript></form>`); err != nil {
			return err
		}
		if browser.Count == 0 {
			return write(`<p class="sl-empty typo-sublabel">No tokens match these filters.</p></div><script src="/static/token-table.js?v=20261005-contexts" defer></script>`)
		}
		if err := write(`<div class="sl-token-groups">`); err != nil {
			return err
		}
		for _, group := range browser.Selected.Groups {
			if err := writeTokenGroup(writer, group, true); err != nil {
				return err
			}
		}
		return write(`</div></div><script src="/static/token-table.js?v=20261005-contexts" defer></script>`)
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

// Inline SVGs keep the table icons independent of an icon font or UI library.
func tableIcon(name string) string {
	paths := map[string]string{
		"search":   `<circle cx="10.5" cy="10.5" r="6.5"/><path d="m16 16 5 5"/>`,
		"eye":      `<path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12Z"/><circle cx="12" cy="12" r="3"/>`,
		"grid":     `<path d="M3 3h6v6H3zM15 3h6v6h-6zM3 15h6v6H3zM15 15h6v6h-6z"/>`,
		"expand":   `<path d="m7 9 5-5 5 5M7 15l5 5 5-5"/>`,
		"dropdown": `<path d="m8 10 4 4 4-4"/>`,
		"folder":   `<path d="M3 6h6l2 2h10v12H3Z"/>`,
	}
	return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">` + paths[name] + `</svg>`
}

func subjectPageComponent(subject publish.SubjectView, changes []publish.ChangeView, browser tokenBrowser) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) error {
		content := templ.ComponentFunc(func(ctx context.Context, writer io.Writer) error {
			write := func(value string) error {
				_, err := io.WriteString(writer, value)
				return err
			}
			if err := write(`<p class="sl-back-link"><a class="sl-link" href="` + subjectListPath(subject.Key.Kind != publish.SubjectComponent) + `">← ` + html.EscapeString(subjectKindLabel(subject.Key.Kind)) + `</a></p><header class="sl-subject-heading"><h1 class="typo-page-title">` + html.EscapeString(subject.Name) + `</h1><nav class="sl-subject-actions" aria-label="Component links">`); err != nil {
				return err
			}
			if err := writeTokenDownloadMenu(writer, subject, browser); err != nil {
				return err
			}
			if len(changes) > 0 {
				if err := write(`<a class="sl-link" href="` + html.EscapeString(subjectChangesURL(subject)) + `">Change history (` + strconv.Itoa(len(changes)) + `)</a>`); err != nil {
					return err
				}
			}
			if len(subject.SourcePageURLs) > 0 {
				if sourceURL := safeSourceDocumentationURL(subject.SourcePageURLs[0]); sourceURL != "" {
					if err := write(`<a class="sl-link" href="` + html.EscapeString(sourceURL) + `">Material documentation ↗</a>`); err != nil {
						return err
					}
				}
			}
			if err := write(`</nav></header><section class="sl-section" aria-labelledby="tokens-heading"><h2 class="typo-section-header" id="tokens-heading">Browse tokens</h2>`); err != nil {
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
	className := "sl-token-folder"
	if !state {
		className += " sl-token-subfolder"
	}
	open := ""
	if group.Open {
		open = ` open`
	}
	if err := write(`<details class="` + className + `"` + open + `><summary>` + tableIcon("folder") + `<span>` + html.EscapeString(group.Name) + `</span>` + tableIcon("dropdown") + `</summary>`); err != nil {
		return err
	}
	if len(group.Tokens) > 0 {
		if err := write(`<div class="sl-token-table-wrap"><table class="sl-token-table"><caption class="sl-visually-hidden">` + html.EscapeString(group.Name) + ` tokens</caption><colgroup><col><col class="sl-token-value-column"></colgroup><thead class="sl-visually-hidden"><tr><th scope="col">Token</th><th scope="col">Value</th></tr></thead><tbody>`); err != nil {
			return err
		}
		for _, token := range group.Tokens {
			if err := write(`<tr><th scope="row"><span class="sl-token-description">` + html.EscapeString(token.Name) + `</span><code class="sl-token-source-name">` + html.EscapeString(token.SourceName) + `</code>`); err != nil {
				return err
			}
			if token.Lifecycle != "" {
				if err := write(`<span class="sl-token-lifecycle">` + html.EscapeString(token.Lifecycle) + `</span>`); err != nil {
					return err
				}
			}
			if !token.HasValue {
				if err := write(`</th><td class="sl-token-no-value">—</td></tr>`); err != nil {
					return err
				}
				continue
			}
			if err := write(`</th><td>`); err != nil {
				return err
			}
			if len(token.References) > 0 {
				if err := writeTokenReferences(writer, token); err != nil {
					return err
				}
				if err := write(`</td></tr>`); err != nil {
					return err
				}
				continue
			}
			style := ""
			if token.Color != "" {
				style = ` style="--token-color:` + html.EscapeString(token.Color) + `;--token-ink:` + tokenColorInk(token.Color) + `"`
			}
			if err := write(`<span class="sl-token-value` + colorCellClass(token.Color) + `"` + style + `>` + html.EscapeString(token.Value) + `</span></td></tr>`); err != nil {
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
	return write(`</details>`)
}

func writeTokenReferences(writer io.Writer, token tokenRow) error {
	write := func(value string) error { _, err := io.WriteString(writer, value); return err }
	if err := write(`<details class="sl-token-reference"><summary>`); err != nil {
		return err
	}
	for _, reference := range token.References {
		if err := write(`<code>` + html.EscapeString(reference.Name) + `</code>`); err != nil {
			return err
		}
	}
	if err := write(`</summary><div class="sl-reference-chain"><code>` + html.EscapeString(token.SourceName) + `</code>`); err != nil {
		return err
	}
	if err := writeReferenceBranches(writer, token.References); err != nil {
		return err
	}
	if err := write(`</div></details>`); err != nil {
		return err
	}
	if token.Value != "" && token.Value != "Color not resolved" && token.Value != "Undefined: true" {
		style := ""
		if token.Color != "" {
			style = ` style="--token-color:` + html.EscapeString(token.Color) + `"`
		}
		return write(`<span class="sl-reference-preview"` + style + `>` + html.EscapeString(token.Value) + `</span>`)
	}
	return nil
}

func writeReferenceBranches(writer io.Writer, references []publish.TokenReference) error {
	if _, err := io.WriteString(writer, `<ul>`); err != nil {
		return err
	}
	for _, reference := range references {
		if _, err := io.WriteString(writer, `<li><code>`+html.EscapeString(reference.Name)+`</code>`); err != nil {
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
