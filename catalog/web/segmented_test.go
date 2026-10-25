package web

import (
	"bytes"
	"context"
	"net/url"
	"strings"
	"testing"
)

func TestSegmentedDensityConfiguration(t *testing.T) {
	for _, variant := range []string{"single", "multiple"} {
		for _, control := range []string{"label", "icon", "icon-label"} {
			for _, density := range []string{"0", "-1", "-2", "-3"} {
				original := componentCase{Family: "segmented", Variant: variant, Control: control}
				page := &componentPreview{Cases: []componentCase{original}}
				if err := configureSegmentedPreview(page, url.Values{"preview.density": {density}}); err != nil {
					t.Fatal(err)
				}
				c := page.Cases[0]
				if c.Attributes()["data-density"] != density || c.Control != control || c.Variant != variant {
					t.Fatal("density must apply independently of variant and content")
				}
				if (c.ID() == original.ID()) != (density == "0") {
					t.Fatal("nondefault densities need distinct specimen IDs")
				}
			}
		}
	}
	if err := configureSegmentedPreview(&componentPreview{}, url.Values{"preview.density": {"-4"}}); err == nil {
		t.Fatal("unsupported density must be rejected")
	}
}

func TestSegmentedSelectionRoles(t *testing.T) {
	for variant, role := range map[string]string{"single": "radiogroup", "multiple": "group"} {
		var out bytes.Buffer
		if err := specimenSegmented(componentCase{Family: "segmented", Variant: variant, Control: "label", State: "interactive"}).Render(context.Background(), &out); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), `role="`+role+`"`) {
			t.Fatal(variant, "missing", role)
		}
	}
}
