package web

import (
	"testing"
)

func TestChipActionsAndVariants(t *testing.T) {
	for _, variant := range []string{"filter", "assist", "suggestion", "input"} {
		c := componentCase{Family: "chip", Variant: variant, State: "disabled", Shape: "icon", Width: "removable", Selected: true}
		a := c.Attributes()
		if a["class"] != "md-"+variant+"-chip" || a["aria-disabled"] != "true" {
			t.Fatalf("missing native chip class or disabled state: %v", a)
		}
		d := previewDimensions(c)
		if variant == "assist" || variant == "suggestion" {
			if _, ok := d["selection"]; ok {
				t.Fatal("action chips must not expose selection")
			}
			if a["disabled"] != true {
				t.Fatal("action chips need native disabled buttons")
			}
		}
		if variant == "input" {
			if _, ok := d["elevation"]; ok {
				t.Fatal("input chips have no elevated profile")
			}
		}
	}
	cases := []componentCase{{Family: "chip", Variant: "filter", Shape: "icon", Width: "plain", State: "enabled"}, {Family: "chip", Variant: "assist", Shape: "icon", Width: "plain", State: "enabled"}, {Family: "chip", Variant: "suggestion", Shape: "icon", Width: "plain", State: "enabled"}, {Family: "chip", Variant: "input", Shape: "icon", Width: "removable", State: "enabled"}}
	preview, err := selectComponentPreview(cases, map[string][]string{"preview.compare": {"variants"}})
	if err != nil || len(preview.Cases) != 4 {
		t.Fatalf("variant comparison omitted a chip type: %v %v", preview, err)
	}
}
