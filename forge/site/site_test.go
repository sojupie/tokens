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
			if component.name == "header" && !strings.Contains(html, `href="https://tokencatalog.dev/md3/components"`) {
				t.Fatal("missing catalog destination")
			}
			if component.name == "header" && (!strings.Contains(html, `aria-current="page"`) || !strings.Contains(html, `aria-controls="tokens-site-menu"`)) {
				t.Fatal("missing active Tools destination or mobile menu")
			}

			if component.name == "header" {
				if strings.Count(html, `class="tokens-brand-logo"`) != 2 || strings.Contains(html, `<img class="tokens-brand-logo"`) || !strings.Contains(html, `fill="var(--md-sys-color-tertiary,var(--tokens-logo-fallback))"`) {
					t.Fatal("header and drawer must share the inline, theme-aware tertiary logo")
				}
				for _, tool := range config.Tools {
					if !strings.Contains(html, `href="`+tool.URL+`"`) {
						t.Fatalf("missing direct tool link %s", tool.Name)
					}
				}
				if !strings.Contains(html, `data-drawer-breakpoint="1100"`) {
					t.Fatal("missing adaptive drawer boundary")
				}
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
				if !strings.Contains(html, `href="https://tokencatalog.dev/md3/llms-full.txt"`) || !strings.Contains(html, `href="https://marmenlind.com"`) {
					t.Fatal("missing footer utility links")
				}
				for _, tool := range config.Tools {
					if strings.Contains(html, `href="`+tool.URL+`"`) {
						t.Fatal("footer repeats drawer destinations")
					}
				}
			}

		})
	}
}

func TestWorkspaceKeepsWorkingWidth(t *testing.T) {
	c := Default()
	c.Workspace = true
	c.ActiveTool = c.Tools[0].URL
	var out bytes.Buffer
	if err := Header(c).Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `data-drawer-breakpoint="1440"`) {
		t.Fatal("workspace needs room for viewer and tools")
	}
	if c.current("/tools") != "false" || toolCurrent(c, c.Tools[0]) != "page" || toolCurrent(c, c.Tools[1]) != "false" {
		t.Fatal("only the current tool should be selected")
	}
}
