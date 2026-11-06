package main

import (
	"strings"
	"testing"
)

func TestRenderFreezesScriptsAndStylesWithoutLosingInlineMapping(t *testing.T) {
	p := frozenPage{HTML: `<html><head><script style="--old:1">throw Error()</script><link rel="stylesheet" href="old.css"><style>x{color:red}</style></head><body><div style="color:var(--old)">Text</div></body></html>`, Inline: []string{"--a:1", "color:var(--a)"}}
	result, err := render(p, `<link rel="stylesheet" href="new.css">`)
	if err != nil {
		t.Fatal(err)
	}
	for _, unexpected := range []string{"throw Error", "old.css", "x{color:red}", "--old"} {
		if strings.Contains(result, unexpected) {
			t.Fatal("retained", unexpected)
		}
	}
	if !strings.Contains(result, "color:var(--a)") || !strings.Contains(result, "new.css") {
		t.Fatal(result)
	}
	if !strings.HasPrefix(result, "<!DOCTYPE html>") {
		t.Fatal("capture must replay in standards mode")
	}
	p.HTML = "<!DOCTYPE html>" + p.HTML
	withDoctype, err := render(p, "")
	if err != nil || strings.Count(strings.ToLower(withDoctype), "<!doctype") != 1 {
		t.Fatal("existing doctype must be preserved once", err)
	}
	p.Inline = nil
	if _, err := render(p, ""); err == nil {
		t.Fatal("missing inline styles must fail")
	}
}

func TestMixedPageCollectsSpecimensAndSharedStyles(t *testing.T) {
	first := frozenPage{Path: "/components/buttons", HTML: `<html data-site-drawer="expanded"><head></head><body><nav>old navigation</nav><div data-playground-specimens style="--x:1"><button>Example</button></div></body></html>`, Styles: []sheet{{URL: "http://localhost/core.css", Text: "body{}"}, {URL: "http://localhost/button.css", Text: "button{}", Order: "0010"}}}
	second := frozenPage{Path: "/components/cards", HTML: `<html><body><div data-playground-specimens><article>Card</article></div></body></html>`, Styles: []sheet{{URL: "http://localhost/core.css", Text: "body{}"}, {URL: "http://localhost/card.css", Text: "article{}", Order: "0008"}}}
	p, err := mixedPage([]frozenPage{first, second})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(p.HTML, "old navigation") || strings.Contains(p.HTML, `data-site-drawer`) {
		t.Fatal("retained removed shell")
	}
	if strings.Count(p.HTML, "data-benchmark-family=") != 2 || len(p.Inline) != 1 || p.Inline[0] != "--x:1" {
		t.Fatal("lost specimen structure")
	}
	if len(p.Styles) != 4 || p.Styles[0].URL != "http://localhost/core.css" || p.Styles[1].Order != "0008" {
		t.Fatal("lost stylesheet deduplication/order", p.Styles)
	}
	second.Styles[0].Text = "body{color:red}"
	if _, err := mixedPage([]frozenPage{first, second}); err == nil {
		t.Fatal("changing dependencies must fail")
	}
}
