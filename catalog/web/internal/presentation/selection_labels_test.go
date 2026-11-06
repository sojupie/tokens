package presentation_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	webrender "github.com/sojupie/tokens/catalog/web/internal/render"
)

func TestSelectionExamplesHaveAssociatedVisibleLabels(t *testing.T) {
	for _, family := range []string{"checkbox", "switch"} {
		c := webview.ComponentCase{Family: family, State: "interactive", Variant: "plain"}
		var specimen templ.Component
		if family == "checkbox" {
			specimen = webrender.SpecimenCheckbox(c)
		} else {
			specimen = webrender.SpecimenSwitch(c)
		}
		var out bytes.Buffer
		if err := specimen.Render(context.Background(), &out); err != nil {
			t.Fatal(err)
		}
		html := out.String()
		for _, part := range []string{`id="` + c.ID() + `-input"`, `for="` + c.ID() + `-input"`, `aria-labelledby="` + c.ID() + `-label"`, `id="` + c.ID() + `-label"`} {
			if !strings.Contains(html, part) {
				t.Errorf("%s missing %s", family, part)
			}
		}
	}
}
