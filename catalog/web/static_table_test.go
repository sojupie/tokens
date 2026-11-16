package web

import (
	"bytes"
	"context"
	"strings"
	"testing"

	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	"github.com/sojupie/tokens/forge/components/ui"
	"github.com/sojupie/tokens/source/publish"
)

func TestStaticTableUsesSharedColorPreviewAndReferences(t *testing.T) {
	table := staticTokenTable{Groups: []webview.TokenGroup{{
		Name: "Color roles",
		Tokens: []webview.TokenRow{{
			Name: "Primary", SourceName: "brand.sys.color.primary", Color: "#014D34",
			Value: "#014D34", HasValue: true,
			References: []publish.TokenReference{{Name: "brand.ref.palette.green"}},
		}, {
			Name: "Overlay", SourceName: "brand.sys.color.overlay", Color: "#014D3480",
			Value: "#014D3480", HasValue: true,
		}},
	}}}
	var fragment bytes.Buffer
	modules := ui.Assets{}
	if err := writeStaticTable(ui.WithAssets(context.Background(), modules), &fragment, table); err != nil {
		t.Fatal(err)
	}
	for _, module := range []string{"common", "messages"} {
		if !modules[module] {
			t.Errorf("rendered table did not declare %q", module)
		}
	}
	for _, expected := range []string{
		`<details class="tokens-token-folder">`,
		`--tokens-preview-color:rgba(1, 77, 52, 1);--tokens-preview-on-color:var(--tokens-preview-on-dark)`,
		`--tokens-preview-color:rgba(1, 77, 52, 0.5019607843137255);--tokens-preview-on-color:var(--tokens-preview-on-light)`,
		`brand.ref.palette.green`,
		`popovertarget="token-info-`,
		`data-action="expand"`,
		`<div class="tokens-token-columns"><span>Token</span></div>`,
	} {
		if !strings.Contains(fragment.String(), expected) {
			t.Errorf("static table is missing %q", expected)
		}
	}
	if strings.Contains(fragment.String(), "hx-get") || strings.Contains(fragment.String(), "data-export-open") {
		t.Error("static table depends on Catalog HTTP actions")
	}
}
