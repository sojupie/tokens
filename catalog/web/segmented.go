package web

import (
	"fmt"
	"net/url"
)

func configureSegmentedPreview(page *componentPreview, query url.Values) error {
	density := query.Get("preview.density")
	if density == "" {
		density = "0"
	}
	options := []string{"0", "-1", "-2", "-3"}
	if !containsPreview(options, density) {
		return fmt.Errorf("%w: density %s", errPreviewInput, density)
	}
	page.Controls = append(page.Controls, previewControl{
		Name: "density", Label: "Density", Options: options, Selected: density,
		Labels: map[string]string{"0": "Default (40dp)", "-1": "−1 (36dp)", "-2": "−2 (32dp)", "-3": "−3 (28dp)"},
	})
	for i := range page.Cases {
		page.Cases[i].SegmentDensity = density
	}
	return nil
}
