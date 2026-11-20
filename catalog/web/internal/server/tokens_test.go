package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	webrender "github.com/sojupie/tokens/catalog/web/internal/render"
	"github.com/sojupie/tokens/source/publish"
)

func TestAudienceHelpExplainsSourceLabels(t *testing.T) {
	subject := publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "buttons"}}
	browser := webview.TokenBrowser{Sets: []webview.TokenSet{{Key: "color", Label: "Color"}}, Filters: []webview.TokenFilter{{Name: "Audience", Key: "context.Audience", Options: []string{"1P Baseline", "3P"}, Selected: "3P"}}}
	var output bytes.Buffer
	if err := webrender.TokenBrowserView(subject, browser).Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"What do 1P and 3P mean?", "first-party", "third-party", "does not define them", "not a baseline-versus-Expressive switch"} {
		if !strings.Contains(output.String(), text) {
			t.Fatalf("missing audience explanation: %q", text)
		}
	}
}

func TestReferenceCellUsesAliasesAndPreservesChain(t *testing.T) {
	row := webview.TokenRow{SourceName: "md.comp.button.shadow", Value: "Color not resolved", References: []publish.TokenReference{{Name: "md.sys.color.shadow", References: []publish.TokenReference{{Name: "md.ref.palette.neutral0"}}}}}
	var output bytes.Buffer
	if err := webview.WriteTokenReferences(t.Context(), &output, row); err != nil {
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
	if err := webview.WriteTokenReferences(t.Context(), &output, row); err != nil {
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
	browser := buildTokenBrowser(subject, url.Values{"set": {"button|color|elevated"}})
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
	row := webview.TokenRow{HasValue: ok, References: value.References}
	if !tokenSearchMatches("elevation.level1", tokenSearchFields(publish.TokenView{}, row, value)) {
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
	filters := webview.TokenFilters(tokens, url.Values{})
	if got := webview.TableContextLabel(filters); got != "Default, Light" {
		t.Fatalf("default context = %q", got)
	}
	filters = webview.TokenFilters(tokens, url.Values{"context.Theme": {"Dark"}, "context.Contrast": {"High contrast"}})
	if got := webview.TableContextLabel(filters); got != "High contrast, Dark" {
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
		{Value: json.RawMessage(`0.38`)},
	}}}
	filters := webview.TokenFilters(tokens, url.Values{"context.iOS dark theme elevation": {"Elevated"}})
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
	filters = webview.TokenFilters(tokens, url.Values{"context.Theme": {"Dark"}, "context.Contrast": {"Default"}, "context.Platform": {"Web"}, "context.Audience": {"3P"}})
	selected := make(map[string]string)
	for _, filter := range filters {
		selected[filter.Name] = filter.Selected
	}
	if selected["Contrast"] != "High contrast" || selected["Platform"] != "iOS" || selected["Audience"] != "1P Baseline" || selected["iOS dark theme elevation"] == "" {
		t.Fatalf("did not normalize to the available dark context: %#v", selected)
	}
}

func TestPreviewTextContrast(t *testing.T) {
	if webview.TokenColorTextColor("rgba(103, 80, 164, 1)") != "var(--tokens-preview-on-dark)" || webview.TokenColorTextColor("rgba(255, 255, 255, 1)") != "var(--tokens-preview-on-light)" {
		t.Fatal("preview labels need contrasting text")
	}
}

func TestValuePreviewAndInfoShareOneNativeReferencePopup(t *testing.T) {
	row := webview.TokenRow{Name: "Button color", SourceName: "md.comp.button.color", Value: "#6750A4", Color: "rgba(103,80,164,1)", HasValue: true, References: []publish.TokenReference{{Name: "md.sys.color.primary"}}}
	var output bytes.Buffer
	if err := webview.WriteTokenReferences(t.Context(), &output, row); err != nil {
		t.Fatal(err)
	}
	body := output.String()
	start := strings.Index(body, `popovertarget="`) + len(`popovertarget="`)
	end := start + strings.Index(body[start:], `"`)
	id := body[start:end]
	if strings.Count(body, `popovertarget="`+id+`"`) != 2 || strings.Count(body, `id="`+id+`"`) != 1 || !strings.Contains(body, `popover role="dialog"`) || !strings.Contains(body, `class="tokens-color-dot"`) || !strings.Contains(body, `class="tokens-value-native-disclosure"`) {
		t.Fatal("preview and info must target one popup while the compact swatch remains present")
	}
}

func TestTableSharesDecorationButIncludesAllReferenceData(t *testing.T) {
	row := webview.TokenRow{Name: "Button color", SourceName: "md.comp.button.color", Value: "#6750A4", HasValue: true, References: []publish.TokenReference{{Name: "md.sys.color.primary", References: []publish.TokenReference{{Name: "md.ref.palette.primary40"}}}}}
	var output bytes.Buffer
	output.WriteString(webrender.TokenTableSharedMarkup())
	for range 2 {
		if err := webview.WriteTokenReferences(t.Context(), &output, row); err != nil {
			t.Fatal(err)
		}
	}
	body := output.String()
	if strings.Count(body, `<template class="tokens-reference-chrome">`) != 1 || strings.Count(body, webview.MaterialSymbolPath("link")) != 1 || strings.Count(body, `data-copy-reference`) != 1 {
		t.Fatal("popover controls and icon paths must be shared, not repeated per reference")
	}
	for _, name := range []string{row.SourceName, "md.sys.color.primary", "md.ref.palette.primary40"} {
		if strings.Count(body, ">"+name+"</code>") != 4 {
			t.Fatalf("each token must include its full reference tree in both disclosures: %s", name)
		}
	}
	markdown, err := htmlMarkdown([]byte("<main>"+body+"</main>"), "https://tokens.test", "")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(markdown, "Color") || !strings.Contains(markdown, "md.ref.palette.primary40") {
		t.Fatalf("Markdown must preserve reference data and omit the shared controls: %s", markdown)
	}
}

func TestMissingTokenValuesHaveExplicitLabels(t *testing.T) {
	subject := publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectStyle, Slug: "spacing"}, Tokens: []publish.TokenView{
		{SourceName: "md.sys.color.primary", Category: "COLOR"},
		{SourceName: "md.sys.color.secondary", Category: "COLOR", Values: []publish.TokenValueView{{Value: json.RawMessage(`{"undefined":true}`)}}},
	}}
	browser := buildTokenBrowser(subject, url.Values{"set": {"all"}})
	var labels []string
	var visit func([]webview.TokenGroup)
	visit = func(groups []webview.TokenGroup) {
		for _, g := range groups {
			for _, row := range g.Tokens {
				labels = append(labels, row.Value)
			}
			visit(g.Children)
		}
	}
	visit(browser.Selected.Groups)
	if len(labels) != 2 || !strings.Contains(strings.Join(labels, "|"), "No captured value") || !strings.Contains(strings.Join(labels, "|"), "No value for selected context") {
		t.Fatal(labels)
	}
	var out bytes.Buffer
	if err := webview.WriteTokenReferences(t.Context(), &out, webview.TokenRow{Name: "Color", Value: "Color not resolved", HasValue: true}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Color not resolved") || strings.Contains(out.String(), ">—<") {
		t.Fatal(out.String())
	}
}

func TestElevationTableExcludesOnlySupportingColorDefinitions(t *testing.T) {
	subject := publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectStyle, Slug: "elevation"}}
	for _, name := range []string{"md.sys.color.primary", "md.ref.palette.primary20", "md.ref.palette.primary30", "md.ref.palette.primary40", "md.ref.palette.primary80", "md.ref.palette.primary90", "md.ref.palette.primary95", "md.sys.elevation.level1", "md.sys.elevation.surface-tint-color"} {
		subject.Tokens = append(subject.Tokens, publish.TokenView{SourceName: name})
	}
	if got := webview.AllTokenSet(subject); len(got.SourceTokens) != 2 {
		t.Fatalf("all table: %v", got.SourceTokens)
	}
	count := 0
	for _, set := range webview.TokenSets(subject) {
		count += len(set.SourceTokens)
	}
	if count != 2 || len(subject.Tokens) != 9 {
		t.Fatal("table filtering changed source records", count, len(subject.Tokens))
	}
	subject.Key.Slug = "spacing"
	if len(webview.AllTokenSet(subject).SourceTokens) != 9 {
		t.Fatal("exclusion leaked into another style")
	}
	subject.Key = publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "elevation"}
	if len(webview.AllTokenSet(subject).SourceTokens) != 9 {
		t.Fatal("exclusion leaked into components")
	}
}

func TestSpringInspectorUsesCapturedParametersForTheSelectedContext(t *testing.T) {
	const spring = "md.sys.motion.spring.fast.spatial"
	tokens := []publish.TokenView{
		{SourceName: spring + ".damping", Values: []publish.TokenValueView{
			{Value: json.RawMessage(`{"numeric":0.9}`), ContextLabels: []string{"Motion scheme: Standard"}, IsDefaultContext: true},
			{Value: json.RawMessage(`{"numeric":0.6}`), ContextLabels: []string{"Motion scheme: Expressive"}},
		}},
		{SourceName: spring + ".stiffness", Values: []publish.TokenValueView{
			{Value: json.RawMessage(`{"numeric":1400}`), ContextLabels: []string{"Motion scheme: Standard"}, IsDefaultContext: true},
			{Value: json.RawMessage(`{"numeric":800}`), ContextLabels: []string{"Motion scheme: Expressive"}},
		}},
	}
	standard := springProperties(json.RawMessage(`"`+spring+`"`), tokens, nil)
	if len(standard) != 2 || standard[0].Value != "0.9" || standard[1].Value != "1400" {
		t.Fatalf("unexpected default spring parameters: %+v", standard)
	}
	properties := springProperties(json.RawMessage(`"`+spring+`"`), tokens, map[string]string{"Motion scheme": "Expressive"})
	if len(properties) != 2 || properties[0].Value != "0.6" || properties[1].Value != "800" ||
		properties[0].References[0].References[0].Name != spring+".damping" ||
		properties[1].References[0].References[0].Name != spring+".stiffness" {
		t.Fatalf("unexpected spring parameters: %+v", properties)
	}
	if properties := springProperties(json.RawMessage(`"`+spring+`"`), tokens, map[string]string{"Motion scheme": "Unknown"}); len(properties) != 0 {
		t.Fatalf("missing context must not invent spring parameters: %+v", properties)
	}
}

func TestReferenceArrowsConnectToAnotherAliasOrLiteral(t *testing.T) {
	const arrow = `href="#tokens-table-icon-arrow_downward"`
	const literal = `class="tokens-reference-literal"`
	const alias = "md.sys.duration"
	row := webview.TokenRow{Name: "Duration", Value: alias, SourceName: "md.comp.demo.duration", References: []publish.TokenReference{{Name: alias}}}
	var output bytes.Buffer
	render := func() string {
		t.Helper()
		output.Reset()
		if err := webview.WriteTokenReferences(t.Context(), &output, row); err != nil {
			t.Fatal(err)
		}
		return output.String()
	}
	if body := render(); strings.Contains(body, arrow) || strings.Contains(body, literal) {
		t.Fatal("an unresolved alias must be the last item")
	}
	row.Value = "50ms"
	row.References[0].References = []publish.TokenReference{{Name: "md.ref.duration"}}
	if body := render(); strings.Count(body, arrow) != 4 || strings.Count(body, literal) != 2 {
		t.Fatal("each disclosure must connect two aliases and end at a literal")
	}
	row.References[0].References = append(row.References[0].References, publish.TokenReference{Name: "md.ref.other"})
	if body := render(); strings.Count(body, arrow) != 2 || strings.Count(body, literal) != 2 {
		t.Fatal("a fork must not connect sibling aliases or imply each branch resolves to the shared literal")
	}
}
