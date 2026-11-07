package presentation_test

import (
	"testing"

	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
)

func TestDividerLayouts(t *testing.T) {
	for _, row := range []struct{ variant, shape, inset string }{{"full-width", "", "false"}, {"inset", "start", "true"}, {"inset", "both", "middle"}} {
		c := webview.ComponentCase{Family: "divider", Variant: row.variant, Shape: row.shape}
		if c.Attributes()["data-inset"] != row.inset {
			t.Fatalf("%s/%s: %+v", row.variant, row.shape, c.Attributes())
		}
	}
}
