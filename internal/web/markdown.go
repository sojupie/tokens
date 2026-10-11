package web

import (
	"bytes"
	"mime"
	"net/http"
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
// with HTML. Scripts, navigation, controls and duplicated footer text are omitted.
func htmlMarkdown(source []byte) (string, error) {
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
		switch node.Data {
		case "head", "script", "style", "nav", "footer", "svg", "button", "form", "dialog":
			return ""
		}
		var body strings.Builder
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.TextNode && strings.TrimSpace(child.Data) == "" {
				body.WriteByte(' ')
				continue
			}
			body.WriteString(render(child))
		}
		text := strings.TrimSpace(body.String())
		switch node.Data {
		case "h1", "h2", "h3", "h4", "h5", "h6":
			return "\n\n" + strings.Repeat("#", int(node.Data[1]-'0')) + " " + text + "\n\n"
		case "p", "section", "article", "header", "div", "details", "ul", "ol":
			return "\n\n" + text + "\n\n"
		case "li":
			return "\n- " + text + "\n"
		case "br":
			return "\n"
		case "strong", "b":
			return "**" + text + "**"
		case "a":
			href := attrs["href"]
			if href == "" || strings.HasPrefix(href, "javascript:") {
				return text
			}
			return "[" + text + "](" + strings.NewReplacer("(", "%28", ")", "%29", " ", "%20").Replace(href) + ")"
		case "tr":
			cells := []string{}
			heading := false
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				if child.Data == "th" || child.Data == "td" {
					if child.Data == "th" {
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
	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}
	return strings.TrimSpace(text) + "\n", nil
}
