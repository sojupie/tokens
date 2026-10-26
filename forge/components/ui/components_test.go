package ui

import (
	"bytes"
	"context"
	"github.com/a-h/templ"
	"strings"
	"testing"
)

func TestDrawerContracts(t *testing.T) {
	var out bytes.Buffer
	ctx := templ.WithChildren(context.Background(), DrawerItem(templ.Attributes{"data-project-id": "one"}, "A < B"))
	if e := Drawer("drawer", true, "Recordings").Render(ctx, &out); e != nil {
		t.Fatal(e)
	}
	for _, required := range []string{`<dialog`, `data-navigation-modal`, `data-navigation-group`, `data-navigation-item`, `class="md-nav-indicator"`, `A &lt; B`} {
		if !strings.Contains(out.String(), required) {
			t.Fatal(required, out.String())
		}
	}
	out.Reset()
	if e := Drawer("drawer", false, "Recordings").Render(context.Background(), &out); e != nil {
		t.Fatal(e)
	}
	if strings.Contains(out.String(), "<dialog") || !strings.Contains(out.String(), "<nav") {
		t.Fatal(out.String())
	}
}
func TestFieldNamesAndButtonState(t *testing.T) {
	var out bytes.Buffer
	if e := Select("position", "Position", []Option{{"x", "Horizontal"}}).Render(context.Background(), &out); e != nil {
		t.Fatal(e)
	}
	for _, required := range []string{`aria-controls="position-options"`, `id="position-options"`, `role="listbox"`, `data-field-value="x"`, `id="position"`} {
		if !strings.Contains(out.String(), required) {
			t.Fatal(required, out.String())
		}
	}
	out.Reset()
	if e := Button(templ.Attributes{"disabled": true, "class": "md-button md-filled-button"}).Render(context.Background(), &out); e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(out.String(), "disabled") {
		t.Fatal(out.String())
	}
}
