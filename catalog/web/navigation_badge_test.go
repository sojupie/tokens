package web

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestNavigationBadgeNamesAndInitialSelection(t *testing.T) {
	c := componentCase{Family: "nav-rail", State: "interactive", Selected: true}
	for variant, want := range map[string]string{"small": "New notification", "labeled": "3 unread files"} {
		c.NavBadge = variant
		var out bytes.Buffer
		if err := navigationItemContents(c, 1, "Files", "folder").Render(context.Background(), &out); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), `aria-label="`+want+`"`) || c.NavigationItem(1)["aria-label"] != "Files "+want {
			t.Errorf("%s missing %s", variant, want)
		}
	}
	c.Selected = false
	var out bytes.Buffer
	if err := navigationItemContents(c, 1, "Files", "folder").Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "md-nav-badge") {
		t.Error("Already-selected destination must not show unread badge")
	}
}
