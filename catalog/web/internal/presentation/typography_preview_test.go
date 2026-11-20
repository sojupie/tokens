package presentation

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/sojupie/tokens/source/publish"
)

func TestTypographyPreviewKeepsPropertyReferencesSeparate(t *testing.T) {
	const scale = "md.sys.typescale.display-medium"
	value := json.RawMessage(`{"type":{"fontNames":{"values":["Roboto"]},"fontWeight":400,"fontSize":{"value":45,"unit":"POINTS"},"fontTracking":{"unit":"POINTS"},"lineHeight":{"value":52,"unit":"POINTS"}}}`)
	references := []publish.TokenReference{{Name: scale, References: []publish.TokenReference{
		{Name: scale + ".font", References: []publish.TokenReference{{Name: "md.ref.typeface.brand"}}},
		{Name: scale + ".weight"},
	}}}
	preview := TokenTypographyPreview(value, "md.comp.scroll-field.item.unselected.label-text.type", references)
	if preview == nil || len(preview.Properties) != 5 || preview.FallbackFont ||
		!strings.Contains(preview.Style, "--tokens-type-size:45px;--tokens-type-line-height:52px;--tokens-type-tracking:0px") {
		t.Fatalf("unexpected typography preview: %+v", preview)
	}
	font, size, weight := preview.Properties[0], preview.Properties[1], preview.Properties[3]
	if font.Value != "Roboto" || len(font.References) != 1 ||
		font.References[0].Name != scale || font.References[0].References[0].References[0].Name != "md.ref.typeface.brand" ||
		size.Value != "45pt" || len(size.References) != 0 ||
		weight.Value != "400" || len(weight.References) != 1 || weight.References[0].Name != scale || weight.References[0].References[0].Name != scale+".weight" {
		t.Fatalf("unexpected property associations: %+v", preview.Properties)
	}
}
