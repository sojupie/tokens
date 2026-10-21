package web

import "testing"

func TestDialogProfilesAndTargets(t *testing.T) {
	c := componentCase{Family: "dialog", Variant: "full-screen", Orientation: "bottom", Control: "close", State: "pressed", Width: "middle"}
	if c.DialogPanel()["class"] != "md-full-screen-dialog" || c.DialogPanel()["data-scroll-start"] != "middle" {
		t.Fatal("fullscreen profile missing")
	}
	if c.DialogClose()["data-component-state"] != "pressed" {
		t.Fatal("close preview state missing")
	}
	if _, ok := c.DialogAction(true)["data-component-state"]; ok {
		t.Fatal("untargeted action must retain its native state")
	}
	c.State = "interactive"
	if _, ok := c.DialogPanel()["open"]; ok {
		t.Fatal("interactive dialogs must start closed")
	}
	cases := []componentCase{{Family: "dialog", Variant: "basic", Control: "action", Shape: "plain", State: "enabled"}, {Family: "dialog", Variant: "full-screen", Control: "action", Orientation: "header", Shape: "plain", Width: "top", State: "enabled"}}
	p, err := selectComponentPreview(cases, map[string][]string{"preview.compare": {"variants"}})
	if err != nil || len(p.Cases) != 2 {
		t.Fatalf("dialog variants must both remain in comparison: %v %v", p, err)
	}
}

func TestDialogOptionalHeadlineAndDecorations(t *testing.T) {
	for _, variant := range []string{"basic", "full-screen"} {
		c := componentCase{Family: "dialog", Variant: variant, Shape: "icon-divided", DialogHeadline: "hidden"}
		a := c.DialogPanel()
		if a["aria-label"] != c.DialogHeading() {
			t.Fatal("headline-free dialogs need a name")
		}
		if _, ok := a["aria-labelledby"]; ok {
			t.Fatal("missing headline referenced")
		}
		if !c.DialogIcon() || !c.DialogDividers() {
			t.Fatal("icon and dividers must compose")
		}
		shown := c
		shown.DialogHeadline = "shown"
		if shown.ID() == c.ID() {
			t.Fatal("headline configurations collide")
		}
	}
}

func TestDialogVariantResetsUnsupportedControls(t *testing.T) {
	cases := []componentCase{
		{Family: "dialog", Variant: "basic", Control: "action", Shape: "plain", State: "enabled"},
		{Family: "dialog", Variant: "full-screen", Control: "close", Shape: "plain", State: "enabled"},
	}
	p, err := selectComponentPreview(cases, map[string][]string{"preview.compare": {"states"}, "preview.variant": {"basic"}, "preview.control": {"close"}})
	if err != nil || len(p.Cases) != 1 || p.Cases[0].Variant != "basic" || p.Cases[0].Control != "action" {
		t.Fatalf("stale fullscreen control changed the selected variant: %#v %v", p, err)
	}
	for _, control := range p.Controls {
		if control.Name == "control" {
			t.Fatal("basic dialog offers an absent close control")
		}
	}
}
