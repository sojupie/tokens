package web

import (
	"context"
	"strings"
	"testing"
)

func TestPlaygroundReplacesHistoryAndKeepsAddressableControls(t *testing.T) {
	var out strings.Builder
	page := componentPreview{Action: "/components/buttons"}
	if err := componentPreviewView(page).Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	body := out.String()
	for _, want := range []string{`hx-replace-url="true"`, `hx-trigger="change"`, `method="get"`, `action="/components/buttons"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %s", want)
		}
	}
	if strings.Contains(body, `hx-push-url`) {
		t.Fatal("configuration changes must not add history entries")
	}
}
