package web

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"github.com/sojupie/tokens/source/publish"
)

func TestSourceTablesRetainMembershipGroupOrderAndTokenOrder(t *testing.T) {
	token := func(name, group string, groupOrder, order int) publish.TokenView {
		return publish.TokenView{SourceName: name, Category: "ELEVATION", Lifecycle: "active", Presentation: &publish.TokenPresentation{SetName: "md.comp.button.elevated", SetLabel: "Button - Color - Elevated", SetOrder: 2, Groups: []publish.DisplayGroup{{Name: group, Order: groupOrder}}, Order: order}}
	}
	subject := publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "buttons"}, Name: "Buttons", Tokens: []publish.TokenView{
		token("md.comp.button.elevated.z", "Enabled", 1, 1),
		token("md.comp.button.elevated.a", "Enabled", 1, 2),
		token("md.comp.button.elevated.disabled.elevation", "Disabled", 2, 0),
	}}
	browser := buildTokenBrowser(subject, url.Values{})
	if len(browser.Sets) != 1 || browser.Selected.Label != "Button - Color - Elevated" {
		t.Fatalf("sets=%+v", browser.Sets)
	}
	if len(browser.Selected.Groups) != 2 || browser.Selected.Groups[0].Name != "Enabled" || browser.Selected.Groups[1].Name != "Disabled" {
		t.Fatalf("groups=%+v", browser.Selected.Groups)
	}
	rows := browser.Selected.Groups[0].Tokens
	if rows[0].SourceName != "md.comp.button.elevated.z" {
		t.Fatalf("source row order lost: %+v", rows)
	}
	if got := tokenSetURL(subject, browser.Selected); got != "/components/buttons/sets/md.comp.button.elevated" {
		t.Fatal(got)
	}
	if tokenSetMenuLabel(browser.Selected) != "Button - Color - Elevated" {
		t.Fatal("menu invented a label")
	}
	canonical, ok := legacyTokenSet(subject, "button/color/elevated")
	if !ok || canonical.sourceName != "md.comp.button.elevated" {
		t.Fatal("legacy table URL lost its source-table replacement")
	}
	subject.Tokens = subject.Tokens[:1]
	subject.Tokens[0].Presentation.Groups = nil
	browser = buildTokenBrowser(subject, url.Values{})
	var flat strings.Builder
	if err := writeTokenGroup(&flat, browser.Selected.Groups[0], true); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(flat.String(), "<details") {
		t.Fatal("a source table without display groups must remain flat")
	}
}

func TestShapeDescriptionAndPreviewPreserveCornerOverrides(t *testing.T) {
	for _, test := range []struct{ raw, text, radius string }{
		{`{"shape":{"family":"SHAPE_FAMILY_ROUNDED_CORNERS","defaultSize":{"value":16,"unit":"DIPS"},"topLeft":{"unit":"DIPS"}}}`, "Default: 16dp · Top-left: 0dp", "0px 16px 16px 16px"},
		{`{"shape":{"family":"SHAPE_FAMILY_ROUNDED_CORNERS","defaultSize":{"unit":"DIPS"},"topRight":{"value":16,"unit":"DIPS"},"bottomRight":{"value":16,"unit":"DIPS"}}}`, "Default: 0dp · Top-right: 16dp · Bottom-right: 16dp", "0px 16px 16px 0px"},
		{`{"shape":{"family":"SHAPE_FAMILY_CIRCULAR"}}`, "Full", "50%"},
		{`{"shape":{"family":"SHAPE_FAMILY_CUT_CORNERS","defaultSize":{"value":16,"unit":"DIPS"}}}`, "Cut Corners · 16dp", ""},
	} {
		if got := tokenValueText(json.RawMessage(test.raw)); got != test.text {
			t.Errorf("text=%q want %q", got, test.text)
		}
		if got := shapePreviewCSS(json.RawMessage(test.raw)); got != test.radius {
			t.Errorf("radius=%q want %q", got, test.radius)
		}
	}
	var output strings.Builder
	if err := writeTokenReferences(&output, tokenRow{Name: "Large end", Value: "Default: 0dp · Top-right: 16dp · Bottom-right: 16dp", Shape: "0px 16px 16px 0px", HasValue: true}); err != nil {
		t.Fatal(err)
	}
	if strings.Count(output.String(), `border-radius:0px 16px 16px 0px`) != 1 {
		t.Fatal("the table must provide one directional shape for reuse in the popup")
	}
	markdown, err := htmlMarkdown([]byte(`<main><button class="tokens-token-set"><span>Shape</span></button><table><tr><th>Large end</th><td>`+output.String()+`</td></tr></table></main>`), "https://example.test/styles/shape")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(markdown, "Default: 0dp") != 1 || !strings.Contains(markdown, "## Shape") {
		t.Fatalf("duplicated popup or missing table label: %s", markdown)
	}
}

func TestSourceDimensionZeroRequiresAConcreteDimension(t *testing.T) {
	for _, test := range []struct {
		raw, want string
		valid     bool
	}{
		{`{"unit":"DIPS"}`, "0dp", true},
		{`{"unit":"POINTS"}`, "0pt", true},
		{`{}`, "", false},
		{`{"unit":"DIPS","value":null}`, "", false},
		{`{"unit":"DIPS","value":"unknown"}`, "", false},
	} {
		var value map[string]any
		if err := json.Unmarshal([]byte(test.raw), &value); err != nil {
			t.Fatal(err)
		}
		got, valid := dimensionValueText(value)
		if valid != test.valid || got != test.want {
			t.Errorf("%s: got %q/%v", test.raw, got, valid)
		}
	}
}
