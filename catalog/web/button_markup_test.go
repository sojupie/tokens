package web

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/sojupie/tokens/source/publish"
	"golang.org/x/net/html"
)

func TestTableAndDownloadButtonsUseSharedImplementation(t *testing.T) {
	subject := publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "buttons"}, Name: "Buttons"}
	browser := tokenBrowser{Sets: []tokenSet{{Key: "color", Label: "Color"}}, Filters: []tokenFilter{{Name: "Theme", Key: "context.Theme", Options: []string{"Light", "Dark"}, Selected: "Light"}}}
	var output bytes.Buffer
	if err := tokenBrowserView(subject, browser).Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	if err := siteDownloadButton().Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	document, err := html.ParseWithOptions(&output, html.ParseOptionEnableScripting(false))
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		for _, attr := range node.Attr {
			if attr.Key != "class" {
				continue
			}
			classes := strings.Fields(attr.Val)
			base, variant := false, false
			for _, class := range classes {
				if class == "md-button" {
					base = true
				}
				switch class {
				case "md-filled-button", "md-tonal-button", "md-outlined-button", "md-elevated-button", "md-text-button":
					variant = true
				}
			}
			if variant {
				count++
				if !base {
					t.Errorf("button missing shared implementation: %s", attr.Val)
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(document)
	if count != 4 {
		t.Fatalf("checked %d common buttons, want 4", count)
	}
}
