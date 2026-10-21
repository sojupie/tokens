package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"github.com/sojupie/tokens/source/publish"
)

func TestAudienceHelpExplainsSourceLabels(t *testing.T) {
	subject := publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "buttons"}}
	browser := tokenBrowser{Sets: []tokenSet{{Key: "color", Label: "Color"}}, Filters: []tokenFilter{{Name: "Audience", Key: "context.Audience", Options: []string{"1P Baseline", "3P"}, Selected: "3P"}}}
	var output bytes.Buffer
	if err := tokenBrowserView(subject, browser).Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"What do 1P and 3P mean?", "first-party", "third-party", "does not define them", "not a baseline-versus-Expressive switch"} {
		if !strings.Contains(output.String(), text) {
			t.Fatalf("missing audience explanation: %q", text)
		}
	}
}

func TestReferenceCellUsesAliasesAndPreservesChain(t *testing.T) {
	row := tokenRow{SourceName: "md.comp.button.shadow", Value: "Color not resolved", References: []publish.TokenReference{{Name: "md.sys.color.shadow", References: []publish.TokenReference{{Name: "md.ref.palette.neutral0"}}}}}
	var output bytes.Buffer
	if err := writeTokenReferences(&output, row); err != nil {
		t.Fatal(err)
	}
	body := output.String()
	if !strings.Contains(body, `<code title="md.sys.color.shadow">md.sys.color.shadow</code>`) || !strings.Contains(body, "md.ref.palette.neutral0") {
		t.Fatalf("missing primary alias or dependency: %s", body)
	}
	if strings.Contains(body, "Color not resolved") {
		t.Fatal("an unresolved preview must not obscure a valid reference")
	}
	row.References[0].Name = `<script>alert(1)</script>`
	output.Reset()
	if err := writeTokenReferences(&output, row); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "<script>") {
		t.Fatal("reference names must be HTML escaped")
	}
}

func TestMaterialColorSetIncludesRelatedValues(t *testing.T) {
	subject := publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectComponent}, Tokens: []publish.TokenView{
		{SourceName: "md.comp.button.elevated.hovered.state-layer.color", Category: "COLOR"},
		{SourceName: "md.comp.button.elevated.selected.hovered.state-layer.opacity", Category: "OPACITY"},
		{SourceName: "md.comp.button.elevated.hovered.container.elevation", Category: "ELEVATION"},
	}}
	browser := buildTokenBrowser(subject, url.Values{})
	if len(browser.Sets) != 1 || browser.Selected.Label != "Button - Color - Elevated" || browser.Count != 3 {
		t.Fatalf("expected one complete color set, got %#v", browser)
	}
	if len(browser.Selected.Groups) != 1 || browser.Selected.Groups[0].Name != "Hovered" || len(browser.Selected.Groups[0].Tokens) != 3 {
		t.Fatalf("toggle values must stay in their interaction state: %#v", browser.Selected.Groups)
	}
	if browser.Selected.Groups[0].Open {
		t.Fatal("state groups should start collapsed")
	}
}

func TestUnresolvedReferencesRemainBrowsable(t *testing.T) {
	references := []publish.TokenReference{{Name: "md.sys.elevation.level1"}}
	value, ok := selectTokenValue([]publish.TokenValueView{{Value: json.RawMessage(`{"undefined":true}`), References: references}}, nil)
	if !ok || len(value.References) != 1 {
		t.Fatal("unresolved alias was dropped")
	}
	if !referencesMatch("elevation.level1", value.References) {
		t.Fatal("reference names must be searchable")
	}
}

func TestMaterialDisplayNames(t *testing.T) {
	token := publish.TokenView{SourceName: "md.comp.button.elevated.selected.container.color", DisplayName: "Button elevated container color - toggle (selected)"}
	if tokenDisplayName(token) != token.DisplayName {
		t.Fatal("upstream label should be preserved")
	}
	token.DisplayName = ""
	if tokenDisplayName(token) != "Button elevated container color - toggle (selected)" {
		t.Fatal("fallback label misplaced toggle state")
	}
}

func TestMaterialContextDefaultsAndSelection(t *testing.T) {
	tokens := []publish.TokenView{{Values: []publish.TokenValueView{
		{ContextLabels: []string{"Theme: Light", "Contrast: Default"}, Value: []byte(`1`)},
		{ContextLabels: []string{"Theme: Dark", "Contrast: High contrast"}, Value: []byte(`2`)},
	}}}
	filters := tokenFilters(tokens, url.Values{})
	if got := tableContextLabel(filters); got != "Default, Light" {
		t.Fatalf("default context = %q", got)
	}
	filters = tokenFilters(tokens, url.Values{"context.Theme": {"Dark"}, "context.Contrast": {"High contrast"}})
	if got := tableContextLabel(filters); got != "High contrast, Dark" {
		t.Fatalf("selected context = %q", got)
	}
	value, ok := selectTokenValue(tokens[0].Values, map[string]string{"Theme": "Dark", "Contrast": "High contrast"})
	if !ok || string(value.Value) != "2" {
		t.Fatalf("wrong context value: %#v", value)
	}
}

func TestContextFiltersOnlyOfferSupportedCombinations(t *testing.T) {
	tokens := []publish.TokenView{{Values: []publish.TokenValueView{
		{ContextLabels: []string{"Theme: Light", "Contrast: Default", "Platform: Web", "Audience: 3P"}},
		{ContextLabels: []string{"Theme: Dark", "Contrast: High contrast", "Platform: iOS", "Audience: 1P Baseline", "iOS dark theme elevation: Elevated"}},
		{ContextLabels: []string{"Theme: Dark", "Contrast: High contrast", "Platform: iOS", "Audience: 1P Baseline", "iOS dark theme elevation: None"}},
		{Value: json.RawMessage(`0.38`)}, // Shared literal does not introduce context options.
	}}}
	filters := tokenFilters(tokens, url.Values{"context.iOS dark theme elevation": {"Elevated"}})
	for _, filter := range filters {
		if filter.Name == "iOS dark theme elevation" {
			t.Fatal("iOS elevation must not be offered for Light/Web")
		}
		if filter.Name == "Platform" && (filter.Selected != "Web" || len(filter.Options) != 1) {
			t.Fatalf("unsupported platform choices: %#v", filter)
		}
		if filter.Selected == "" {
			t.Fatal("context selection must be explicit")
		}
	}
	filters = tokenFilters(tokens, url.Values{"context.Theme": {"Dark"}, "context.Contrast": {"Default"}, "context.Platform": {"Web"}, "context.Audience": {"3P"}})
	selected := make(map[string]string)
	for _, filter := range filters {
		selected[filter.Name] = filter.Selected
	}
	if selected["Contrast"] != "High contrast" || selected["Platform"] != "iOS" || selected["Audience"] != "1P Baseline" || selected["iOS dark theme elevation"] == "" {
		t.Fatalf("did not normalize to the available dark context: %#v", selected)
	}
}

func TestPreviewTextContrast(t *testing.T) {
	if tokenColorTextColor("rgba(103, 80, 164, 1)") != "var(--tokens-preview-on-dark)" || tokenColorTextColor("rgba(255, 255, 255, 1)") != "var(--tokens-preview-on-light)" {
		t.Fatal("preview labels need contrasting text")
	}
}

func TestValuePreviewAndInfoShareOneNativeReferencePopup(t *testing.T) {
	row := tokenRow{Name: "Button color", SourceName: "md.comp.button.color", Value: "#6750A4", Color: "rgba(103,80,164,1)", HasValue: true, References: []publish.TokenReference{{Name: "md.sys.color.primary"}}}
	var output bytes.Buffer
	if err := writeTokenReferences(&output, row); err != nil {
		t.Fatal(err)
	}
	body := output.String()
	start := strings.Index(body, `popovertarget="`) + len(`popovertarget="`)
	end := start + strings.Index(body[start:], `"`)
	id := body[start:end]
	if strings.Count(body, `popovertarget="`+id+`"`) != 2 || strings.Count(body, `id="`+id+`"`) != 1 || !strings.Contains(body, `popover role="dialog"`) || !strings.Contains(body, `class="tokens-color-dot"`) || strings.Contains(body, `<details`) {
		t.Fatal("preview and info must target one popup while the compact swatch remains present")
	}
}

func TestTableSharesDecorationButIncludesAllReferenceData(t *testing.T) {
	row := tokenRow{Name: "Button color", SourceName: "md.comp.button.color", Value: "#6750A4", HasValue: true, References: []publish.TokenReference{{Name: "md.sys.color.primary", References: []publish.TokenReference{{Name: "md.ref.palette.primary40"}}}}}
	var output bytes.Buffer
	output.WriteString(tokenTableSharedMarkup())
	for range 2 {
		if err := writeTokenReferences(&output, row); err != nil {
			t.Fatal(err)
		}
	}
	body := output.String()
	if strings.Count(body, `<template class="tokens-reference-chrome">`) != 1 || strings.Count(body, materialSymbolPaths["camera"]) != 1 || strings.Count(body, `data-copy-reference`) != 1 {
		t.Fatal("popover controls and icon paths must be shared, not repeated per reference")
	}
	for _, name := range []string{row.SourceName, "md.sys.color.primary", "md.ref.palette.primary40"} {
		if strings.Count(body, ">"+name+"</code>") != 2 {
			t.Fatalf("each token must include its full reference tree without JavaScript: %s", name)
		}
	}
	markdown, err := htmlMarkdown([]byte("<main>"+body+"</main>"), "https://tokens.test")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(markdown, "Color") || !strings.Contains(markdown, "md.ref.palette.primary40") {
		t.Fatalf("Markdown must preserve reference data and omit the shared controls: %s", markdown)
	}
}
