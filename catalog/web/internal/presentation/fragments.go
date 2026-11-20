package presentation

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/url"
	"strings"

	"github.com/a-h/templ"
	"github.com/sojupie/tokens/forge/components/ui"
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

func WriteTokenGroups(ctx context.Context, writer io.Writer, browser TokenBrowser) error {
	if len(browser.Sets) == 0 {
		_, err := io.WriteString(writer, `<p class="tokens-empty md-typescale-body-large">No published tokens for this page.</p>`)
		return err
	}
	if browser.Count == 0 && !browser.LocalSearch {
		_, err := io.WriteString(writer, `<p class="tokens-empty md-typescale-body-large">No tokens match these filters.</p>`)
		return err
	}
	if _, err := io.WriteString(writer, `<div class="tokens-token-groups">`); err != nil {
		return err
	}
	for _, group := range browser.Selected.Groups {
		if err := WriteTokenGroup(ctx, writer, group, true); err != nil {
			return err
		}
	}
	ending := `</div>`
	if browser.LocalSearch {
		hidden := ""
		if browser.Count > 0 {
			hidden = ` hidden`
		}
		ending += `<p class="tokens-empty md-typescale-body-large" data-token-search-empty role="status"` + hidden + `>No tokens match these filters.</p>`
	}
	_, err := io.WriteString(writer, ending)
	return err
}

func WriteTokenGroup(ctx context.Context, writer io.Writer, group TokenGroup, state bool) error {
	write := func(value string) error {
		_, err := io.WriteString(writer, value)
		return err
	}
	className := "tokens-token-folder"
	if !state {
		className += " tokens-token-subfolder"
	}
	attributes := ""
	if group.Open {
		attributes += ` open`
	}
	if group.Hidden {
		attributes += ` hidden`
	}
	if group.Name != "" {
		if err := write(`<details class="` + className + `"` + attributes + `><summary>` + `<span class="tokens-folder-closed">` + TableIcon("folder") + `</span><span class="tokens-folder-open">` + TableIcon("folder-open") + `</span><span>` + html.EscapeString(group.Name) + `</span>` + TableIcon("dropdown") + `</summary>`); err != nil {
			return err
		}
	}
	if len(group.Tokens) > 0 {
		hidden := ` hidden`
		for _, row := range group.Tokens {
			if !row.Hidden {
				hidden = ""
				break
			}
		}
		if err := write(`<div class="tokens-token-table-wrap"` + hidden + `><table class="tokens-token-table"><caption class="tokens-visually-hidden">` + html.EscapeString(strings.TrimSpace(group.Name+" tokens")) + `</caption><colgroup><col><col class="tokens-token-value-column"></colgroup><thead class="tokens-visually-hidden"><tr><th scope="col">Token</th><th scope="col">Value</th></tr></thead><tbody>`); err != nil {
			return err
		}
		for _, token := range group.Tokens {
			attributes := ""
			if token.Hidden {
				attributes += ` hidden`
			}
			if token.Search != nil {
				search, err := json.Marshal(token.Search)
				if err != nil {
					return err
				}
				attributes += ` data-token-search-fields="` + html.EscapeString(string(search)) + `"`
			}
			if err := write(`<tr` + attributes + `><th scope="row"><div class="tokens-token-heading"><span class="tokens-token-description">` + html.EscapeString(token.Name) + `</span>`); err != nil {
				return err
			}
			if len(token.Notices) > 0 {
				if err := writeTokenGuidance(ctx, writer, token); err != nil {
					return err
				}
			}
			if err := write(`</div><code class="tokens-token-source-name">` + html.EscapeString(token.SourceName) + `</code>`); err != nil {
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
			if err := WriteTokenReferences(ctx, writer, token); err != nil {
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
		if err := WriteTokenGroup(ctx, writer, child, false); err != nil {
			return err
		}
	}
	if group.Name != "" {
		return write(`</details>`)
	}
	return nil
}

func WriteTokenReferences(ctx context.Context, writer io.Writer, token TokenRow) error {
	write := func(value string) error { _, err := io.WriteString(writer, value); return err }
	id := tokenReferenceID(token)
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
	valueText := `<span class="tokens-preview-value">` + html.EscapeString(label) + `</span>`
	preview := valueText
	popupPreview := valueText
	popupClass := ColorCellClass(token.Color)
	if token.Typography != nil && token.Typography.Style != "" {
		preview = typographySpecimen(token.Typography)
		popupPreview = preview
		popupClass += " tokens-popup-type"
	}
	if token.Shape != "" {
		shapeClass := "tokens-shape-swatch"
		if token.Shape == "50%" {
			shapeClass += " tokens-shape-circular"
		}
		popupPreview = `<span class="` + shapeClass + `" style="border-radius:` + html.EscapeString(token.Shape) + `" aria-hidden="true"></span>`
		preview = popupPreview + valueText
		popupClass += " tokens-popup-shape"
	}
	if token.Elevation != "" {
		popupPreview = `<span class="md-elevation tokens-elevation-swatch" style="--_elevation-level:` + html.EscapeString(token.Elevation) + `" aria-hidden="true"></span>`
		preview = popupPreview + `<span class="tokens-visually-hidden">` + html.EscapeString(label) + `</span>`
	}
	if token.ValueKind == "motionReference" {
		popupPreview = `<span class="tokens-numeric-swatch">` + TokenTableSymbol("numbers") + `</span>`
		preview = popupPreview + `<span class="tokens-visually-hidden">` + html.EscapeString(label) + `</span>`
		popupClass += " tokens-popup-motion"
	}
	compact := `<span class="tokens-compact-value"` + style + `>`
	if token.Color != "" {
		compact += `<span class="tokens-color-dot" aria-hidden="true"></span>`
	}
	compact += html.EscapeString(label) + `</span>`
	if err := write(`<div class="tokens-value-cell"><button type="button" class="tokens-large-preview` + ColorCellClass(token.Color) + `"` + style + ` popovertarget="` + id + `" aria-label="View ` + html.EscapeString(token.Name) + ` details">` + preview + `<span class="tokens-preview-name" aria-hidden="true">` + html.EscapeString(token.Name) + `</span></button>` + compact + `<button type="button" class="md-icon-button md-standard-icon-button tokens-icon-button tokens-info-button" popovertarget="` + id + `" aria-label="References for ` + html.EscapeString(token.Name) + `">` + TokenTableSymbol("info") + `</button></div>`); err != nil {
		return err
	}

	if err := write(`<div id="` + id + `" class="md-elevation tokens-reference-popover" popover role="dialog" aria-label="` + html.EscapeString(token.Name) + `">`); err != nil {
		return err
	}
	if err := write(`<div class="tokens-popup-value` + popupClass + `"` + style + `>` + popupPreview + `</div>`); err != nil {
		return err
	}
	properties := token.ValueProperties
	if token.Typography != nil {
		properties = token.Typography.Properties
		if token.Typography.FallbackFont {
			if err := write(`<p class="tokens-type-font-note">Uses Roboto when the captured font is unavailable.</p>`); err != nil {
				return err
			}
		}
	}
	if len(properties) == 0 {
		properties = []TokenProperty{scalarTokenProperty(token, label)}
	}
	if err := writeTokenProperties(ctx, writer, id, properties, token.SourceName, style); err != nil {
		return err
	}

	return write(`</div>`)
}

func tokenReferenceID(token TokenRow) string {
	return fmt.Sprintf("token-info-%x", sha256.Sum256([]byte(token.SourceName+"\x00"+token.Name+"\x00"+token.Value)))
}

func typographySpecimen(typography *TypographyPreview) string {
	return `<span class="tokens-type-specimen" style="` + html.EscapeString(typography.Style) + `" aria-hidden="true">Aa</span>`
}

func writeTokenProperties(ctx context.Context, writer io.Writer, id string, properties []TokenProperty, source, colorStyle string) error {
	list := ui.List(ui.ListProps{
		Label: "Token properties", Appearance: ui.Expressive, Variant: ui.SegmentedList, Lines: ui.TwoLines, Leading: ui.LeadingIcon,
		Attrs: templ.Attributes{"class": "md-list tokens-value-properties", "data-control": "expand", "data-expand-surface": "tonal", "data-expand-icon": "expanded"},
	})
	rows := templ.ComponentFunc(func(ctx context.Context, writer io.Writer) error {
		for index, property := range properties {
			panelID := fmt.Sprintf("%s-property-%d", id, index)
			if err := writeTokenProperty(ctx, writer, panelID, property, source, colorStyle); err != nil {
				return err
			}
		}
		return nil
	})
	return list.Render(templ.WithChildren(ctx, rows), writer)
}

func writeTokenProperty(ctx context.Context, writer io.Writer, panelID string, property TokenProperty, source, colorStyle string) error {
	leading := ui.Icon(property.Icon)
	if property.Icon == "palette" && colorStyle != "" {
		leading = templ.Raw(`<span class="tokens-color-dot"` + colorStyle + `></span>`)
	}
	item := ui.ListItemProps{
		LeadingKind: ui.LeadingIcon, Leading: leading,
		Overline: property.Label, Label: property.Value, SingleLine: true,
	}
	if len(property.References) == 0 {
		return ui.ListRow(item).Render(ctx, writer)
	}
	item.Trailing = ui.ListIcon("expand_more", true)
	content := ui.ListContent(item)
	var chain strings.Builder
	if err := writeTokenReferenceChain(&chain, source, property.References, resolvedReferenceValue(property.References, property.Value)); err != nil {
		return err
	}
	aliases := chain.String()
	fallback := templ.ComponentFunc(func(ctx context.Context, writer io.Writer) error {
		if _, err := io.WriteString(writer, `<noscript><details class="tokens-value-native-disclosure"><summary class="md-list-item">`); err != nil {
			return err
		}
		if err := content.Render(ctx, writer); err != nil {
			return err
		}
		_, err := io.WriteString(writer, `</summary>`+aliases+`</details></noscript>`)
		return err
	})
	return ui.ListRowFrame(ui.ListRowFrameProps{
		Attrs:         templ.Attributes{"data-list-row": "", "data-list-label": property.Label, "data-list-expanded": "false"},
		ContentAttrs:  templ.Attributes{"class": "md-list-item", "type": "button", "data-list-item": "", "aria-expanded": "false", "aria-controls": panelID},
		ChildrenAttrs: templ.Attributes{"id": panelID, "aria-label": property.Label + " references", "hidden": true, "inert": true},
		Before:        fallback, Content: content, Expanded: templ.Raw(`<li>` + aliases + `</li>`),
	}).Render(ctx, writer)
}

func resolvedReferenceValue(references []publish.TokenReference, value string) string {
	if value == "" {
		return ""
	}
	for _, reference := range references {
		if value == reference.Name || resolvedReferenceValue(reference.References, value) == "" {
			return ""
		}
	}
	return value
}

func writeTokenReferenceChain(writer io.Writer, source string, references []publish.TokenReference, value string) error {
	write := func(text string) error { _, err := io.WriteString(writer, text); return err }
	if err := write(`<div class="tokens-reference-chain">`); err != nil {
		return err
	}
	if source != "" {
		if err := write(`<code class="tokens-popup-source tokens-visually-hidden">` + html.EscapeString(source) + `</code>`); err != nil {
			return err
		}
	}
	if err := writeReferenceBranches(writer, references, value); err != nil {
		return err
	}
	return write(`</div>`)
}

func writeTokenGuidance(ctx context.Context, writer io.Writer, token TokenRow) error {
	var body strings.Builder
	title := "Usage discouraged"
	seen := make(map[string]bool)
	for _, notice := range token.Notices {
		if notice.Message == "" || seen[notice.Message] {
			continue
		}
		seen[notice.Message] = true
		if notice.State == "deprecated" {
			title = "Deprecated token"
		}
		body.WriteString(`<p>` + html.EscapeString(notice.Message) + `</p>`)
	}
	if body.Len() == 0 {
		return nil
	}
	id := tokenReferenceID(token) + "-guidance"
	host := ui.TooltipHost(templ.Attributes{"class": "md-tooltip-demo tokens-token-guidance-anchor", "data-tooltip-hover": ""})
	trigger := ui.IconButton(ui.IconButtonProps{
		Label: "Usage guidance for " + token.Name, Icon: "warning", Variant: "tonal", Size: ui.ExtraSmall, Width: "wide",
		Attrs: templ.Attributes{
			"class":         "md-icon-button md-tonal-icon-button tokens-token-guidance-trigger",
			"popovertarget": id, "popovertargetaction": "show", "aria-haspopup": "dialog",
		},
	})
	tip := ui.Tooltip(ui.TooltipProps{ID: id, Variant: "rich", Title: title, Body: templ.Raw(body.String())})
	content := templ.ComponentFunc(func(ctx context.Context, writer io.Writer) error {
		if err := trigger.Render(ctx, writer); err != nil {
			return err
		}
		return tip.Render(ctx, writer)
	})
	return host.Render(templ.WithChildren(ctx, content), writer)
}

func writeReferenceBranches(writer io.Writer, references []publish.TokenReference, value string) error {
	write := func(text string) error { _, err := io.WriteString(writer, text); return err }
	if err := write(`<ul>`); err != nil {
		return err
	}
	for _, reference := range references {
		name := html.EscapeString(reference.Name)
		if err := write(`<li><div class="tokens-reference-node">` + TokenTableSymbol("link") + `<code title="` + name + `">` + name + `</code></div>`); err != nil {
			return err
		}
		next := reference.References
		// A shared literal belongs to the property, not to each branch in a fork.
		leaf := value != "" && len(references) == 1 && len(next) == 0
		if len(next) > 0 || leaf {
			if err := write(TokenTableSymbol("arrow_downward")); err != nil {
				return err
			}
		}
		if len(next) > 0 {
			branchValue := value
			if len(references) > 1 {
				branchValue = ""
			}
			if err := writeReferenceBranches(writer, next, branchValue); err != nil {
				return err
			}
		}
		if leaf {
			if err := writeReferenceLeaf(writer, value); err != nil {
				return err
			}
		}
		if err := write(`</li>`); err != nil {
			return err
		}
	}
	if len(references) > 1 && value != "" {
		if err := write(`<li>`); err != nil {
			return err
		}
		if err := writeReferenceLeaf(writer, value); err != nil {
			return err
		}
		if err := write(`</li>`); err != nil {
			return err
		}
	}
	return write(`</ul>`)
}

func writeReferenceLeaf(writer io.Writer, value string) error {
	_, err := io.WriteString(writer, `<div class="tokens-reference-literal">`+TokenTableSymbol("deployed_code")+`<span class="tokens-reference-leaf">`+html.EscapeString(value)+`</span></div>`)
	return err
}
