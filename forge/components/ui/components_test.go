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
	for _, required := range []string{`aria-controls="position-options"`, `id="position-options"`, `role="listbox"`, `data-field-value="x"`, `id="position"`, `class="md-chip-icon" data-slot="trailing-icon"`} {
		if !strings.Contains(out.String(), required) {
			t.Fatal(required, out.String())
		}
	}
	if strings.Count(out.String(), `fill="currentColor"`) != strings.Count(out.String(), `<svg`) {
		t.Fatal("select arrow and option check must paint with their inherited state color")
	}
	out.Reset()
	if e := Button(templ.Attributes{"disabled": true, "class": "md-button md-filled-button"}).Render(context.Background(), &out); e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(out.String(), "disabled") {
		t.Fatal(out.String())
	}
}

func TestCheckboxKeepsItsLabelInTheSharedChoice(t *testing.T) {
	var out bytes.Buffer
	if err := Checkbox("loop", "Loop").Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{`class="md-checkbox-choice"`, `id="loop"`, `<label for="loop">Loop</label></span>`} {
		if !strings.Contains(out.String(), required) {
			t.Fatal(required, out.String())
		}
	}
}

func TestMenuItemSlotsAndSemantics(t *testing.T) {
	var out bytes.Buffer
	err := MenuItem(templ.Attributes{"role": "menuitemradio", "aria-checked": "true", "disabled": true}, "A < B", Button(nil), Button(nil)).Render(context.Background(), &out)
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{`class="md-menu-item"`, `role="menuitemradio"`, `aria-checked="true"`, `disabled`, `data-slot="leading-icon"`, `data-slot="trailing-icon"`, `class="md-menu-item-text"`, `A &lt; B`} {
		if !strings.Contains(out.String(), required) {
			t.Fatal(required, out.String())
		}
	}
	if strings.Contains(out.String(), `role="menuitem"`) {
		t.Fatal("default role overrides selection role")
	}
}

func TestReusableWorkspaceComponents(t *testing.T) {
	cases := []struct {
		component templ.Component
		required  []string
	}{
		{Menu("actions", "Actions", true), []string{`class="md-menu"`, `data-menu-main`, `popover="auto"`, `<section><ul role="presentation"`}},
		{Toolbar(nil), []string{`class="md-toolbar"`, `role="toolbar"`}},
		{ButtonGroup("transport", nil), []string{`class="md-button-group transport"`, `role="group"`}},
		{ListItem(nil, "Recording"), []string{`class="md-list-item"`, `class="md-list-body"`, `class="md-list-label"`}},
		{Dialog("export", "title", "Export", "md-dialog", Button(nil)), []string{`aria-labelledby="title"`, `data-dialog-surface`, `data-dialog-content`, `class="md-basic-actions"`}},
		{Snackbar("notice", Button(nil)), []string{`class="md-snackbar"`, `popover="manual"`, `class="md-feedback-text"`, `data-snackbar-close`}},
	}
	for _, c := range cases {
		var out bytes.Buffer
		if err := c.component.Render(context.Background(), &out); err != nil {
			t.Fatal(err)
		}
		for _, want := range c.required {
			if !strings.Contains(out.String(), want) {
				t.Fatal(want, out.String())
			}
		}
		if strings.Contains(out.String(), "@") {
			t.Fatal("unrendered component call", out.String())
		}
	}
}
