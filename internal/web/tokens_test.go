package web

import (
	"net/url"
	"testing"

	"github.com/sojupie/m3-source/internal/publish"
)

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

func TestPreviewTextContrast(t *testing.T) {
	if tokenColorInk("rgba(103, 80, 164, 1)") != "#ffffff" || tokenColorInk("rgba(255, 255, 255, 1)") != "#1d1b20" {
		t.Fatal("preview labels need contrasting text")
	}
}
