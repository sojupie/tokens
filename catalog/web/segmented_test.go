package web

import (
	"net/url"
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
