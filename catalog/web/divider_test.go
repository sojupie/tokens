package web

import "testing"

func TestDividerLayouts(t *testing.T) {
	for _, row := range []struct{ variant, shape, inset string }{{"full-width", "", "false"}, {"inset", "start", "true"}, {"inset", "both", "middle"}} {
		c := componentCase{Family: "divider", Variant: row.variant, Shape: row.shape}
		if c.Attributes()["data-inset"] != row.inset {
			t.Fatalf("%s/%s: %+v", row.variant, row.shape, c.Attributes())
		}
	}
}
