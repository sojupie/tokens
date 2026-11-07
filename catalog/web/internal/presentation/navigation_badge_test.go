package presentation_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	webrender "github.com/sojupie/tokens/catalog/web/internal/render"
)

func TestNavigationBadgeNamesAndInitialSelection(t *testing.T) {
	c := webview.ComponentCase{Family: "nav-rail", State: "interactive", Selected: true}
	for variant, want := range map[string]string{"small": "New notification", "labeled": "3 unread files"} {
		c.NavBadge = variant
		var out bytes.Buffer
		if err := webrender.NavigationItemContents(c, 1, "Files", "folder").Render(context.Background(), &out); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), `aria-label="`+want+`"`) || c.NavigationItem(1)["aria-label"] != "Files "+want {
			t.Errorf("%s missing %s", variant, want)
		}
	}
	c.Selected = false
	var out bytes.Buffer
	if err := webrender.NavigationItemContents(c, 1, "Files", "folder").Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "md-nav-badge") {
		t.Error("Selection must not change unread state")
	}
}
