package web

import (
	"net/url"
	"testing"
)

func TestToolbarPreviewTreatsFABAsConfigurationAndPreservesOldLinks(t *testing.T) {
	cases := []componentCase{
		{Family: "toolbar", Variant: "docked", Control: "standard", Orientation: "horizontal", Size: "none", Shape: "icon", State: "enabled"},
		{Family: "toolbar", Variant: "floating", Control: "vibrant", Orientation: "vertical", Size: "medium", Shape: "icon", State: "enabled"},
	}
	query := url.Values{"preview.compare": {"interactive"}, "preview.variant": {"floating-fab"}, "preview.control": {"vibrant"}, "preview.orientation": {"vertical"}, "preview.size": {"medium"}}
	page, err := selectComponentPreview(cases, query)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Cases) != 1 || page.Cases[0].Variant != "floating" || page.Cases[0].Size != "medium" {
		t.Fatalf("wrong FAB composition: %#v", page)
	}
	if query.Get("preview.variant") != "floating-fab" {
		t.Fatal("modified caller query")
	}
}

func TestToolbarComparisonsUseCompatibleContexts(t *testing.T) {
	cases := []componentCase{
		{Family: "toolbar", Variant: "docked", Control: "standard", Orientation: "horizontal", Size: "none", Shape: "icon", State: "enabled"},
		{Family: "toolbar", Variant: "floating", Control: "standard", Orientation: "horizontal", Size: "none", Shape: "icon", State: "enabled"},
		{Family: "toolbar", Variant: "floating", Control: "standard", Orientation: "vertical", Size: "medium", Shape: "icon", State: "enabled"},
	}
	page, err := selectComponentPreview(cases, url.Values{"preview.compare": {"variants"}, "preview.orientation": {"vertical"}, "preview.size": {"medium"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Cases) != 2 {
		t.Fatalf("variant comparison lost a variant: %#v", page.Cases)
	}
	for _, control := range page.Controls {
		if control.Name == "orientation" || control.Name == "size" {
			t.Fatal("incompatible comparison filter exposed")
		}
	}
}
