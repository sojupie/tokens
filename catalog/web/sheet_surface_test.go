package web

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func TestSheetSurfaceDoesNotIncludePlaygroundFrameOrExampleContent(t *testing.T) {
	c := componentCase{Family: "bottom-sheet", Variant: "expanded", Control: "standard", State: "enabled"}
	var out bytes.Buffer
	ctx := templ.WithChildren(context.Background(), templ.Raw("<p>Route details</p>"))
	if err := sheetSurface(c).Render(ctx, &out); err != nil {
		t.Fatal(err)
	}
	body := out.String()
	for _, unwanted := range []string{"md-sheet-demo", "md-sheet-frame", "Supporting text", `value="Sample"`} {
		if strings.Contains(body, unwanted) {
			t.Fatalf("surface includes playground content: %s", unwanted)
		}
	}
	for _, want := range []string{"md-bottom-sheet", "md-sheet-handle", "md-sheet-content", "Route details"} {
		if !strings.Contains(body, want) {
			t.Fatalf("surface lost %s", want)
		}
	}
	out.Reset()
	if err := specimenBottomSheet(c).Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "md-sheet-frame") || !strings.Contains(out.String(), `value="Sample"`) {
		t.Fatal("playground lost its frame or example content")
	}
}
