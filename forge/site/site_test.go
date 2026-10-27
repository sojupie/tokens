package site

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestToolsShareCatalogNavigationAndFooter(t *testing.T) {
	config := Default()
	for _, component := range []struct {
		name   string
		render func(context.Context, *bytes.Buffer) error
	}{
		{"header", func(ctx context.Context, out *bytes.Buffer) error { return Header(config).Render(ctx, out) }},
		{"footer", func(ctx context.Context, out *bytes.Buffer) error { return Footer(config).Render(ctx, out) }},
	} {
		t.Run(component.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := component.render(context.Background(), &out); err != nil {
				t.Fatal(err)
			}
			html := out.String()
			if strings.Contains(html, `href="/components"`) {
				t.Fatal("catalog links would resolve inside tool mount")
			}
			if !strings.Contains(html, `href="https://tokencatalog.dev/md3/components"`) {
				t.Fatal("missing catalog destination")
			}
			if component.name == "header" && (!strings.Contains(html, `aria-current="page"`) || !strings.Contains(html, `aria-controls="tokens-site-menu"`)) {
				t.Fatal("missing active Tools destination or mobile menu")
			}

			if component.name == "header" {
				for _, required := range []string{`<dialog id="tokens-site-menu"`, `data-navigation-modal`, `data-navigation-open`, `data-navigation-close`, `aria-expanded="false"`} {
					if !strings.Contains(html, required) {
						t.Fatalf("missing drawer contract %s", required)
					}
				}
				if strings.Contains(html, "popovertarget") {
					t.Fatal("site navigation still uses a popover menu")
				}
			}
			if component.name == "footer" {
				for _, tool := range config.Tools {
					if !strings.Contains(html, `href="`+tool.URL+`"`) {
						t.Fatalf("missing %s", tool.Name)
					}
				}
			}
		})
	}
}
