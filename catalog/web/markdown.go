package web

import (
	"bytes"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// Negotiate explicit Markdown against HTML; q=0 never selects a representation.
func wantsMarkdown(request *http.Request) bool {
	markdown, htmlQuality := 0.0, 0.0
	for _, part := range strings.Split(request.Header.Get("Accept"), ",") {
		mediaType, params, err := mime.ParseMediaType(strings.TrimSpace(part))
		if err != nil {
			continue
		}
		quality := 1.0
		if q, ok := params["q"]; ok {
			quality, err = strconv.ParseFloat(q, 64)
			if err != nil || quality < 0 || quality > 1 {
				continue
			}
		}
		if mediaType == "text/markdown" && quality > markdown {
			markdown = quality
		}
		if mediaType == "text/html" && quality > htmlQuality {
			htmlQuality = quality
		}
	}
	return markdown > 0 && markdown >= htmlQuality
}

func markdownEscape(value string) string {
	return strings.NewReplacer("\\", "\\\\", "*", "\\*", "_", "\\_", "[", "\\[", "]", "\\]", "<", "&lt;", ">", "&gt;", "|", "\\|", "`", "\\`").Replace(value)
}

// Convert the rendered view so search, table, context and pagination filters agree
// with HTML. Keep content navigation; omit scripts, controls and site chrome.
// Resolve links against the page URL so the Markdown works outside the browser.
func htmlMarkdown(source []byte, pageURL string) (string, error) {
	base, err := url.Parse(pageURL)
	if err != nil {
		return "", err
	}
	document, err := html.Parse(bytes.NewReader(source))
	if err != nil {
		return "", err
	}
	root := document
	var findMain func(*html.Node)
	findMain = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "main" {
			root = node
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			findMain(child)
		}
	}
	findMain(document)
	var render func(*html.Node) string
	render = func(node *html.Node) string {
		if node.Type == html.TextNode {
			value := markdownEscape(strings.Join(strings.Fields(node.Data), " "))
			if len(node.Data) > 0 && strings.ContainsRune(" \t\r\n", rune(node.Data[0])) {
				value = " " + value
			}
			if len(node.Data) > 0 && strings.ContainsRune(" \t\r\n", rune(node.Data[len(node.Data)-1])) {
				value += " "
			}
			return value
		}
		if node.Type != html.ElementNode && node.Type != html.DocumentNode {
			return ""
		}
		attrs := map[string]string{}
		for _, attr := range node.Attr {
			attrs[attr.Key] = attr.Val
		}
		if _, hidden := attrs["hidden"]; hidden {
			return ""
		}
		classes := " " + attrs["class"] + " "
		if node.Data == "button" && strings.Contains(classes, " tokens-token-set ") {
			var label strings.Builder
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				label.WriteString(render(child))
			}
			return "\n\n## " + strings.TrimSpace(label.String()) + "\n\n"
		}
		if strings.Contains(classes, " tokens-reference-popover ") {
			var references []string
			var collect func(*html.Node, bool)
			collect = func(n *html.Node, inList bool) {
				inList = inList || n.Data == "ul"
				if inList && n.Data == "code" {
					references = append(references, strings.TrimSpace(render(n)))
					return
				}
				for child := n.FirstChild; child != nil; child = child.NextSibling {
					collect(child, inList)
				}
			}
			collect(node, false)
			if len(references) == 0 {
				return ""
			}
			return " (references: " + strings.Join(references, "; ") + ") "
		}
		switch node.Data {
		case "head", "script", "style", "noscript", "footer", "svg", "button", "form", "dialog", "caption":
			return ""
		}
		for _, class := range strings.Fields(attrs["class"]) {
			if class == "tokens-reference-toggle" {
				return ""
			}
			if node.Data == "details" && class == "tokens-token-reference" {
				var chain strings.Builder
				for child := node.FirstChild; child != nil; child = child.NextSibling {
					if child.Data != "summary" {
						chain.WriteString(render(child))
					}
				}
				text := strings.Join(strings.Fields(chain.String()), " ")
				return " (reference chain: " + strings.ReplaceAll(text, " - ", " → ") + ") "
			}
		}
		var body strings.Builder
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.TextNode && strings.TrimSpace(child.Data) == "" {
				body.WriteByte(' ')
				continue
			}
			body.WriteString(render(child))
			if node.Data == "nav" {
				body.WriteByte('\n')
			}
		}
		text := strings.TrimSpace(body.String())
		switch node.Data {
		case "h1", "h2", "h3", "h4", "h5", "h6":
			return "\n\n" + strings.Repeat("#", int(node.Data[1]-'0')) + " " + text + "\n\n"
		case "p", "section", "article", "header", "div", "details", "ul", "ol", "nav":
			return "\n\n" + text + "\n\n"
		case "li":
			return "\n- " + text + "\n"
		case "br":
			return "\n"
		case "strong", "b":
			return "**" + text + "**"
		case "code":
			var raw strings.Builder
			var collect func(*html.Node)
			collect = func(n *html.Node) {
				if n.Type == html.TextNode {
					raw.WriteString(n.Data)
				}
				for child := n.FirstChild; child != nil; child = child.NextSibling {
					collect(child)
				}
			}
			collect(node)
			value := strings.Join(strings.Fields(raw.String()), " ")
			delimiter := "`"
			for strings.Contains(value, delimiter) {
				delimiter += "`"
			}
			if strings.HasPrefix(value, "`") || strings.HasSuffix(value, "`") {
				value = " " + value + " "
			}
			return " " + delimiter + strings.ReplaceAll(value, "|", "\\|") + delimiter + " "
		case "summary":
			return "\n\n### " + text + "\n\n"
		case "a":
			href := attrs["href"]
			link, err := url.Parse(href)
			if href == "" || err != nil {
				return text
			}
			switch strings.ToLower(link.Scheme) {
			case "", "http", "https", "mailto", "tel":
			default:
				return text
			}
			href = base.ResolveReference(link).String()
			return "[" + text + "](" + strings.NewReplacer("(", "%28", ")", "%29", " ", "%20").Replace(href) + ")"
		case "tr":
			cells := []string{}
			heading := false
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				if child.Data == "th" || child.Data == "td" {
					if child.Data == "th" && node.Parent != nil && node.Parent.Data == "thead" {
						heading = true
					}
					cells = append(cells, strings.Join(strings.Fields(render(child)), " "))
				}
			}
			if len(cells) == 0 {
				return ""
			}
			row := "\n| " + strings.Join(cells, " | ") + " |\n"
			if heading {
				row += "| " + strings.TrimSuffix(strings.Repeat("--- | ", len(cells)), " ") + "\n"
			}
			return row
		case "table":
			return "\n\n" + strings.ReplaceAll(text, "\n\n", "\n") + "\n\n"
		}
		return body.String()
	}
	text := render(root)
	// Whitespace-only lines from empty UI elements should not create blank blocks.
	lines := strings.Split(text, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t\r")
	}
	text = strings.Join(lines, "\n")
	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}
	return strings.TrimSpace(text) + "\n", nil
}
