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

func TestProcessingIndicatorsRenderSharedAnatomy(t *testing.T) {
	var out bytes.Buffer
	if err := ProgressIndicator(templ.Attributes{"data-variant": "linear", "data-control": "determinate", "aria-valuenow": "25"}).Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{`role="progressbar"`, `aria-valuenow="25"`, `md-progress-track`, `md-progress-active`, `md-progress-stop`, `<svg width="48" height="4"`} {
		if !strings.Contains(out.String(), required) {
			t.Fatal(required)
		}
	}
	out.Reset()
	if err := LoadingIndicator(templ.Attributes{"aria-label": "Loading video engine"}).Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `<svg width="38" height="38"`) || !strings.Contains(out.String(), `data-control="indeterminate"`) || !strings.Contains(out.String(), initialLoadingPath) || strings.Contains(out.String(), `aria-valuenow`) {
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

func TestDrawerLinkPreservesNativeURLAndCurrentState(t *testing.T) {
	var out bytes.Buffer
	if err := DrawerLink("/md3/tools", templ.Attributes{"aria-current": "page"}, "Tools & tests").Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{`<a class="md-nav-item"`, `href="/md3/tools"`, `aria-current="page"`, `Tools &amp; tests`} {
		if !strings.Contains(out.String(), required) {
			t.Fatalf("missing %s in %s", required, out.String())
		}
	}
}

func TestSelectRendersInitialSelectionBeforeRuntimeLoads(t *testing.T) {
	var out bytes.Buffer
	if err := Select("speed", "Speed", []Option{{"0.5", "Half"}, {"1", "Normal"}}, "1").Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{`data-field-display>Normal`, `data-field-value value="1"`, `value="1" selected`, `aria-selected="true"`} {
		if !strings.Contains(out.String(), required) {
			t.Fatal(required, out.String())
		}
	}
}
