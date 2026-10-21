package web

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestSearchDisclosure(t *testing.T) {
	c := componentCase{Family: "search", Variant: "docked", Control: "contained", Width: "clear", SearchContent: "text", Expanded: true, State: "enabled"}
	if c.SearchPanel()["data-open"] != "true" || c.SearchPanel()["role"] != "dialog" {
		t.Fatal("focused search must use the persistent surface")
	}
	if c.SearchQuery() != "Buttons" || c.SearchPanel()["data-open"] != "true" {
		t.Fatal("search query and disclosure must match the specimen")
	}
	c.State = "interactive"
	if _, ok := c.SearchPanel()["open"]; ok {
		t.Fatal("search must stay within the preview canvas")
	}
	c.Expanded = false
	c.Width = "none"
	c.SearchContent = "empty"
	if c.SearchQuery() != "" || c.SearchPanel()["data-open"] != "false" {
		t.Fatal("empty collapsed bar must show its placeholder")
	}
}

func TestSearchIndependentConfiguration(t *testing.T) {
	c := componentCase{Family: "search", Variant: "adaptive", Width: "clear-settings"}
	if c.SearchQuery() != "" || !c.SearchHasAction("clear") || !c.SearchHasAction("settings") || c.SearchHasAction("avatar") {
		t.Fatal("search query and action configuration are coupled")
	}
	if c.SearchPanel()["aria-label"] != "Search tokens" || c.SearchPanel()["data-layout"] != "adaptive" {
		t.Fatal("search labeling or adaptive layout lost")
	}
	page := &componentPreview{Cases: []componentCase{c}}
	if err := configureSearchPreview(page, map[string][]string{"preview.content": {"text"}}); err != nil {
		t.Fatal(err)
	}
	if page.Cases[0].SearchQuery() != "Buttons" {
		t.Fatal("explicit sample query missing")
	}
}

func TestSearchComparisons(t *testing.T) {
	var cases []componentCase
	for _, layout := range []string{"adaptive", "docked", "full-screen"} {
		for _, style := range []string{"contained", "baseline"} {
			for _, expanded := range []bool{false, true} {
				cases = append(cases, componentCase{Family: "search", Variant: layout, Control: style, Expanded: expanded, State: "enabled"})
			}
		}
	}
	for _, comparison := range []string{"layouts", "styles"} {
		p, err := selectComponentPreview(cases, map[string][]string{"preview.compare": {comparison}})
		if err != nil {
			t.Fatal(err)
		}
		if len(p.Cases) != 2 {
			t.Fatalf("%s: wanted two cases, got %d", comparison, len(p.Cases))
		}
		for _, c := range p.Cases {
			if !c.Expanded {
				t.Fatal("comparisons must show focused layout/style differences")
			}
		}
	}
}

func TestSearchKeepsOneHeaderAndInput(t *testing.T) {
	for _, expanded := range []bool{false, true} {
		c := componentCase{Family: "search", Variant: "docked", Control: "contained", Shape: "icon", Width: "clear-settings", Expanded: expanded, State: "interactive"}
		var html bytes.Buffer
		if err := componentSpecimen(c).Render(context.Background(), &html); err != nil {
			t.Fatal(err)
		}
		text := html.String()
		if strings.Count(text, `class="md-search-input"`) != 1 || strings.Count(text, `class="md-search-bar md-search-trigger"`) != 1 || strings.Contains(text, `<dialog class="md-search-view`) {
			t.Fatal("search must retain one header/input rather than a duplicate popup")
		}
	}
}

func TestSearchElevationIsExplicit(t *testing.T) {
	for _, style := range []string{"baseline", "contained"} {
		page := &componentPreview{Cases: []componentCase{{Family: "search", Control: style}}}
		if err := configureSearchPreview(page, map[string][]string{}); err != nil {
			t.Fatal(err)
		}
		if page.Cases[0].SearchElevation != "none" || strings.Contains(page.Cases[0].SearchPanel()["class"].(string), "md-elevation") {
			t.Fatal("search must have no shadow by default")
		}
		if err := configureSearchPreview(page, map[string][]string{"preview.shadow": {"token"}}); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(page.Cases[0].SearchBar()["class"].(string), "md-elevation") || !strings.Contains(page.Cases[0].SearchPanel()["class"].(string), "md-elevation") {
			t.Fatal("explicit elevation must be available to both styles")
		}
	}
}

func TestAdaptiveSearchUsesComponentPrimitives(t *testing.T) {
	c := componentCase{Family: "search", Variant: "adaptive", Control: "contained", Shape: "icon", Width: "clear", State: "interactive"}
	var html bytes.Buffer
	if err := componentSpecimen(c).Render(context.Background(), &html); err != nil {
		t.Fatal(err)
	}
	text := html.String()
	for _, family := range []string{"nav-rail", "list", "card", "fab"} {
		if !strings.Contains(text, `data-family="`+family+`"`) {
			t.Errorf("composition missing %s primitive", family)
		}
	}
	if strings.Contains(text, `data-family="nav-drawer"`) {
		t.Fatal("adaptive composition must not include a drawer")
	}
	if strings.Count(text, `data-search-message=`) != len(searchExamples) || strings.Count(text, `class="md-list-avatar-label"`) != len(searchExamples) {
		t.Fatal("mail content must use the shared list avatar slots for every message")
	}
	if !strings.Contains(text, `data-button-case="`+c.ID()+`-reply"`) {
		t.Fatal("reply must use the shared button renderer")
	}
	if strings.Count(text, `class="md-search-input"`) != 1 {
		t.Fatal("composition must keep a single search input")
	}
}
