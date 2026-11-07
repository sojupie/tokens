package main

import (
	"bytes"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"golang.org/x/net/html"
)

const mixedPaths = "/components/buttons,/components/button-groups,/components/cards,/components/carousel,/components/checkbox,/components/chips,/components/lists,/components/progress-indicators,/components/radio-button,/components/segmented-buttons,/components/sliders,/components/switch,/components/tabs,/components/text-fields"

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}
func find(n *html.Node, predicate func(*html.Node) bool) *html.Node {
	if predicate(n) {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if match := find(c, predicate); match != nil {
			return match
		}
	}
	return nil
}

// Compose the actual rendered specimens, rather than maintaining a second set
// of benchmark-only component anatomy. Module links are collected from captures.
func mixedPage(pages []frozenPage) (frozenPage, error) {
	if len(pages) == 0 {
		return frozenPage{}, fmt.Errorf("mixed page needs captures")
	}
	root, err := html.Parse(strings.NewReader(pages[0].HTML))
	if err != nil {
		return frozenPage{}, err
	}
	if doc := find(root, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "html" }); doc != nil {
		attrs := doc.Attr[:0]
		for _, a := range doc.Attr {
			if !strings.HasPrefix(a.Key, "data-site-") {
				attrs = append(attrs, a)
			}
		}
		doc.Attr = attrs
	}
	body := find(root, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "body" })
	if body == nil {
		return frozenPage{}, fmt.Errorf("missing body")
	}
	for body.FirstChild != nil {
		body.RemoveChild(body.FirstChild)
	}
	markup := `<div class="tokens-layout"><main class="tokens-main tokens-layout-content" data-component-assets id="component-preview"><header class="tokens-subject-heading"><h1 class="md-typescale-headline-medium">Component delivery benchmark</h1></header><form class="tokens-gallery-filters" aria-label="Benchmark appearance"><label>Theme <select class="md-native-select" data-benchmark-appearance="theme"><option value="light">Light</option><option value="dark">Dark</option></select></label><label>Contrast <select class="md-native-select" data-benchmark-appearance="contrast"><option value="default">Default</option><option value="medium">Medium</option><option value="high">High</option></select></label><a href="/original/0/index.html">Original names</a><a href="/short/0/index.html">Short names</a></form><p class="md-typescale-body-medium">Rendered component specimens with frozen behavior. Theme and contrast controls remain active.</p></main></div>`
	fragment, err := html.Parse(strings.NewReader(markup))
	if err != nil {
		return frozenPage{}, err
	}
	fragmentBody := find(fragment, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "body" })
	for fragmentBody.FirstChild != nil {
		child := fragmentBody.FirstChild
		fragmentBody.RemoveChild(child)
		body.AppendChild(child)
	}
	main := find(body, func(n *html.Node) bool { return attr(n, "id") == "component-preview" })
	panels := &html.Node{Type: html.ElementNode, Data: "div", Attr: []html.Attribute{{Key: "class", Val: "tokens-layout-grid tokens-benchmark-grid"}}}
	main.AppendChild(panels)
	result := frozenPage{Path: "/benchmark/mixed"}
	seen := map[string]string{}
	for _, p := range pages {
		doc, err := html.Parse(strings.NewReader(p.HTML))
		if err != nil {
			return result, err
		}
		grid := find(doc, func(n *html.Node) bool {
			for _, a := range n.Attr {
				if a.Key == "data-playground-specimens" {
					return true
				}
			}
			return false
		})
		if grid == nil {
			return result, fmt.Errorf("no rendered specimens in %s", p.Path)
		}
		section := &html.Node{Type: html.ElementNode, Data: "section", Attr: []html.Attribute{{Key: "class", Val: "tokens-section"}, {Key: "data-component-assets"}, {Key: "data-benchmark-family", Val: strings.TrimPrefix(strings.Split(p.Path, "?")[0], "/components/")}}}
		heading := &html.Node{Type: html.ElementNode, Data: "h2", Attr: []html.Attribute{{Key: "class", Val: "md-typescale-title-large"}}}
		heading.AppendChild(&html.Node{Type: html.TextNode, Data: strings.ReplaceAll(strings.TrimPrefix(strings.Split(p.Path, "?")[0], "/components/"), "-", " ")})
		section.AppendChild(heading)
		grid.Parent.RemoveChild(grid)
		section.AppendChild(grid)
		panels.AppendChild(section)
		for _, s := range p.Styles {
			if text, ok := seen[s.URL]; ok {
				if text != s.Text {
					return result, fmt.Errorf("stylesheet changed between captures: %s", s.URL)
				}
				continue
			}
			seen[s.URL] = s.Text
			result.Styles = append(result.Styles, s)
		}
	}
	// Core sheets retain their captured order; optional modules follow the same
	// manifest order used by the live loader, independent of capture order.
	sort.SliceStable(result.Styles, func(i, j int) bool {
		a, b := result.Styles[i].Order, result.Styles[j].Order
		if a == "" {
			return b != ""
		}
		if b == "" {
			return false
		}
		return a < b
	})
	u, _ := url.Parse(result.Styles[0].URL)
	u.Path = "/__benchmark.css"
	u.RawQuery = ""
	result.Styles = append(result.Styles, sheet{URL: u.String(), Order: "9999", Text: `.tokens-gallery-filters [data-benchmark-appearance]{display:inline-block}.tokens-benchmark-grid{row-gap:var(--layout-section-gap)}@container tokens-layout (min-width:840px){.tokens-benchmark-grid>section{--layout-span:calc(var(--layout-columns)/2)}}`})
	var collect func(*html.Node)
	collect = func(n *html.Node) {
		for _, a := range n.Attr {
			if a.Key == "style" {
				result.Inline = append(result.Inline, a.Val)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			collect(c)
		}
	}
	collect(root)
	var output bytes.Buffer
	if err := html.Render(&output, root); err != nil {
		return result, err
	}
	result.HTML = output.String()
	return result, nil
}
