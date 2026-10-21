package web

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestFABMenuUsesIndependentProfilesAndStateTargets(t *testing.T) {
	c := componentCase{Family: "fab-menu", Variant: "secondary", Control: "close", State: "hovered"}
	if c.FABMenuAction(0)["class"] != "md-button md-fab-menu-item" || c.FABMenuAction(-1)["class"] != "md-button md-fab-menu-close" {
		t.Fatal("FAB menu actions must use their own token profiles")
	}
	if c.FABMenuAction(0)["data-component-state"] != "enabled" || c.FABMenuAction(-1)["data-component-state"] != "hovered" {
		t.Fatal("close state must leave action states unchanged")
	}
	c.State = "interactive"
	if c.FABMenuOpener()["popovertarget"] != c.FABMenuPanel()["id"] || c.FABMenuPanel()["popover"] != "auto" || c.FABMenuAction(0)["tabindex"] != "-1" {
		t.Fatal("interactive menu requires a native popover and managed menu focus")
	}
}

func TestInteractiveFABMenuHasOneToggleSurface(t *testing.T) {
	var html bytes.Buffer
	if err := componentSpecimen(componentCase{Family: "fab-menu", Variant: "primary", State: "interactive"}).Render(context.Background(), &html); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(html.String(), "md-fab-menu-close") || strings.Count(html.String(), "md-fab-menu-opener") != 1 {
		t.Fatal("interactive menu must morph its opener instead of overlaying a second button")
	}
}
