package web

import (
	"net/url"
)

func configureToolbarPreview(page *componentPreview, query url.Values) error {
	controls := page.Controls[:0]
	for _, control := range page.Controls {
		if page.Dimension == "variant" && (control.Name == "orientation" || control.Name == "size") {
			continue
		}
		if page.Dimension == "size" && control.Name == "variant" {
			continue
		}
		controls = append(controls, control)
	}
	page.Controls = controls
	for i := range page.Controls {
		control := &page.Controls[i]
		if control.Name == "size" {
			control.Label = "FAB"
			control.Labels = map[string]string{"none": "None", "small": "Small (56dp)", "medium": "Medium (80dp)"}
		}
		if control.Name == "compare" {
			control.Labels = map[string]string{"states": "States", "sizes": "FAB sizes", "variants": "Variants", "interactive": "Interactive"}
		}
	}
	return nil
}
