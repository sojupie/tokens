package presentation_test

import (
	"context"
	"strings"
	"testing"

	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	webrender "github.com/sojupie/tokens/catalog/web/internal/render"
)

func TestPlaygroundReplacesHistoryAndKeepsAddressableControls(t *testing.T) {
	var out strings.Builder
	page := webview.ComponentPreview{Action: "/components/buttons"}
	if err := webrender.ComponentPreviewView(page).Render(context.Background(), &out); err != nil {
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
