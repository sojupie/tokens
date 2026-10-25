package web

import (
	"bytes"
	"context"
	"github.com/a-h/templ"
	"strings"
	"testing"
)

func TestSelectionExamplesHaveAssociatedVisibleLabels(t *testing.T) {
	for _, family := range []string{"checkbox", "switch"} {
		c := componentCase{Family: family, State: "interactive", Variant: "plain"}
		var specimen templ.Component
		if family == "checkbox" {
			specimen = specimenCheckbox(c)
		} else {
			specimen = specimenSwitch(c)
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
