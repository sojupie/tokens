package web

import (
	"net/url"
	"testing"
)

func TestSplitPreviewLeadingContentKeepsMenuTrigger(t *testing.T) {
	cases := []componentCase{{Family: "split", Variant: "filled", Size: "small", State: "enabled"}}
	for _, content := range []string{"icon-label", "label", "icon"} {
		preview, err := selectComponentPreview(cases, url.Values{"preview.compare": {"interactive"}, "preview.shape": {content}})
		if err != nil {
			t.Fatal(err)
		}
		if len(preview.Cases) != 1 || preview.Cases[0].Shape != content {
			t.Fatalf("%s: %#v", content, preview)
		}
		c := preview.Cases[0]
		if c.SplitAction(true)["aria-haspopup"] != "menu" || c.SplitAction(true)["aria-expanded"] != "false" {
			t.Fatal("missing menu semantics")
		}
		if content == "icon" && c.SplitAction(false)["aria-label"] != "Add to cart" {
			t.Fatal("icon-only action has no accessible name")
		}
	}
}
