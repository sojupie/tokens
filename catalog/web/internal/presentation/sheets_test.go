package presentation_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	webrender "github.com/sojupie/tokens/catalog/web/internal/render"
)

func TestSheetModalSemantics(t *testing.T) {
	for _, family := range []string{"bottom-sheet", "side-sheet"} {
		for _, control := range []string{"standard", "modal"} {
			c := webview.ComponentCase{Family: family, Control: control, State: "interactive", Variant: "expanded"}
			var rendered bytes.Buffer
			if err := webrender.ComponentSpecimen(c).Render(context.Background(), &rendered); err != nil {
				t.Fatal(err)
			}
			body := rendered.String()
			if strings.Contains(body, "<dialog") != (control == "modal") {
				t.Fatalf("%s %s: wrong modal element", family, control)
			}
			if control == "standard" && !strings.Contains(body, ` hidden`) {
				t.Fatalf("%s: standard sheet is visible before opening", family)
			}
			if c.SheetOpener()["aria-controls"] != c.SheetAttributes()["id"] {
				t.Fatalf("%s: opener does not identify its sheet", family)
			}
		}
	}
}

func TestSheetActionStateDoesNotDisableDismissal(t *testing.T) {
	c := webview.ComponentCase{Family: "side-sheet", State: "disabled"}
	if c.SheetAction(1)["disabled"] != true || c.SheetAction(0)["disabled"] == true {
		t.Fatal("disabled example must leave Cancel available")
	}
	if c.SheetAction(1)["class"] != "md-button md-sheet-action" {
		t.Fatal("side-sheet actions must use their own color profile")
	}
	c = webview.ComponentCase{Family: "bottom-sheet", Variant: "minimized"}
	if c.SheetHandle()["aria-expanded"] != "false" || c.SheetHandle()["aria-label"] != "Restore sheet" {
		t.Fatal("minimized handle does not describe its action")
	}
}
