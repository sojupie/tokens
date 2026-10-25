package web

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestFieldAffixesAndErrorsAreDescribed(t *testing.T) {
	c := componentCase{Family: "field", Variant: "outlined", Shape: "affixes", Invalid: true}
	var out bytes.Buffer
	if err := specimenField(c).Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`aria-describedby="` + c.ID() + `-units ` + c.ID() + `-help"`, `id="` + c.ID() + `-units"`, `US dollars per kilogram`, `role="alert"`, `Enter a valid value.`, `Price per kilogram`} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %s", want)
		}
	}
	c.Invalid = false
	c.Shape = "plain"
	if c.FieldInput()["aria-describedby"] != c.ID()+"-help" {
		t.Fatal("Plain field must describe only supporting text")
	}
}
