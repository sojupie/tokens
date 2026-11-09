package server

import (
	"bytes"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	"github.com/sojupie/tokens/source/publish"
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

type markdownTokenGuideKey struct{}

// Use the same set links and context choices as the HTML table.
func (server *Server) markdownTokenGuide(subject publish.SubjectView, browser webview.TokenBrowser, viewJSON string) string {
	var guide strings.Builder
	fmt.Fprintf(&guide, "**Showing %s** · %d of %d rows.\n\n", markdownEscape(webview.SelectedTableLabel(subject, browser.Selected)), browser.Count, len(browser.Selected.SourceTokens))
	if browser.Query != "" {
		clear := browser
		clear.Query = ""
		fmt.Fprintf(&guide, "Search filter `q` is **%s**. [Clear search](%s).\n\n", markdownEscape(strings.Join(strings.Fields(browser.Query), " ")), markdownLinkURL(server.publicOrigin+webview.TokenSetLinkURL(subject, browser.Selected, clear)))
	}
	fmt.Fprintf(&guide, "[Fetch only this table](%s).\n\n", markdownLinkURL(server.publicOrigin+webview.TableOnlyURL(subject, browser)))
	guide.WriteString("### Token sets\n\n")
	fmt.Fprintf(&guide, "Open a set below, or [view all tokens in this subject](%s). Set links keep the search and any context choices the destination supports.\n\n", markdownLinkURL(server.publicOrigin+webview.TokenSetLinkURL(subject, webview.AllTokenSet(subject), browser)))
	for _, set := range browser.Sets {
		fmt.Fprintf(&guide, "- [%s](%s) · set key `%s`", markdownEscape(webview.TokenTableQualifier(subject, set)), markdownLinkURL(server.publicOrigin+webview.TokenSetLinkURL(subject, set, browser)), set.Key)
		if set.Key == browser.Selected.Key {
			guide.WriteString(" (current)")
		}
		guide.WriteByte('\n')
	}
	guide.WriteString("\n### Query parameters\n\n")
	guide.WriteString("Send `Accept: text/markdown` again when following a link. Add or replace these query parameters on the set URL. Use `?` before the first parameter and `&` between parameters. URL-encode spaces in names and values.\n\n")
	guide.WriteString("- `q` searches the selected set, for example `q=container`. Remove it to return every row in the set.\n")
	guide.WriteString("- `set` selects a table on the subject URL. Use one of the keys above, or `set=all` for all tables. The linked set URLs already select their table.\n")
	guide.WriteString("- `view=table` returns only the token groups and rows in Markdown. HTML also includes the table controls.\n")
	if len(browser.Filters) > 0 {
		guide.WriteString("\nUse one value per context parameter. Change one context at a time and read the updated choices. A change can affect the options for other parameters.\n\n")
		guide.WriteString("| Parameter | Current value | Available values |\n| --- | --- | --- |\n")
		for _, filter := range browser.Filters {
			fmt.Fprintf(&guide, "| `%s` | %s | %s |\n", filter.Key, markdownEscape(filter.Selected), markdownEscape(strings.Join(filter.Options, ", ")))
		}
	} else {
		guide.WriteString("\nThis set has no context parameters.\n")
	}
	if browser.Preview != nil && len(browser.Preview.Controls) > 0 {
		guide.WriteString("\n### Playground parameters\n\n")
		guide.WriteString("Pass these parameters to the implementation export to configure its example. Use `preview.compare=interactive` on this page to see the controls for a single example.\n\n")
		guide.WriteString("| Parameter | Current value | Available values |\n| --- | --- | --- |\n")
		for _, control := range browser.Preview.Controls {
			fmt.Fprintf(&guide, "| `preview.%s` | %s | %s |\n", control.Name, markdownEscape(control.Selected), markdownEscape(strings.Join(control.Options, ", ")))
		}
	}
	download := server.publicOrigin + subjectDownloadURL(subject)
	setDownload := url.Values{"scope": {"set"}, "set": {browser.Selected.Key}}
	fmt.Fprintf(&guide, "\n### Downloads\n\n- [Current table view as JSON](%s)\n- [Selected set as JSON with all captured contexts](%s)\n- [Whole subject as JSON with all captured contexts](%s)\n",
		markdownLinkURL(server.publicOrigin+viewJSON), markdownLinkURL(download+"?"+setDownload.Encode()), markdownLinkURL(download+"?scope=all"))
	options := url.Values{"kind": {string(subject.Key.Kind)}, "slug": {subject.Key.Slug}, "scope": {"view"}, "set": {browser.Selected.Key}, "options": {"1"}}
	if browser.Query != "" {
		options.Set("q", browser.Query)
	}
	for _, filter := range browser.Filters {
		options.Set(filter.Key, filter.Selected)
	}
	fmt.Fprintf(&guide, "- [Token export context choices for this selection](%s/downloads/export?%s)\n", server.publicOrigin, options.Encode())
	if server.publicOrigin != "" {
		fmt.Fprintf(&guide, "- [Download guide for scopes, contexts, and formats](%s/llms-full.txt)\n", server.publicOrigin)
	}
	return "\n\n" + guide.String() + "\n## Token values\n\n"
}

func markdownLinkURL(href string) string {
	return strings.NewReplacer("(", "%28", ")", "%29", " ", "%20").Replace(href)
}

// Convert the rendered view so search, table, context and pagination filters agree
// with HTML. Keep content navigation; omit scripts, controls and site chrome.
// Resolve links against the page URL so the Markdown works outside the browser.
func htmlMarkdown(source []byte, pageURL, tokenGuide string) (string, error) {
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

		if attrs["id"] == "component-preview" || attrs["aria-hidden"] == "true" {
			return ""
		}
		if _, hidden := attrs["hidden"]; hidden {
			return ""
		}
		classes := " " + attrs["class"] + " "
		if strings.Contains(classes, " tokens-table-only-link ") {
			return ""
		}
		if node.Data == "form" && strings.Contains(classes, " tokens-token-filters ") {
			return tokenGuide
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
		case "head", "script", "style", "noscript", "template", "footer", "svg", "button", "form", "dialog", "caption":
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
		summaryHasHeading := false
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.TextNode && strings.TrimSpace(child.Data) == "" {
				body.WriteByte(' ')
				continue
			}
			if node.Data == "summary" && child.Type == html.ElementNode && len(child.Data) == 2 && child.Data[0] == 'h' && child.Data[1] >= '1' && child.Data[1] <= '6' {
				summaryHasHeading = true
			}
			if node.Data != "nav" && child.Data == "a" && child.PrevSibling != nil && child.PrevSibling.Data == "a" {
				body.WriteByte(' ')
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
			if text == "" {
				return ""
			}
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
			if summaryHasHeading {
				return "\n\n" + text + "\n\n"
			}
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
			return "[" + text + "](" + markdownLinkURL(href) + ")"
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
			var rows []string
			for _, line := range strings.Split(text, "\n") {
				if line = strings.TrimSpace(line); line != "" {
					rows = append(rows, line)
				}
			}
			return "\n\n" + strings.Join(rows, "\n") + "\n\n"
		}
		return body.String()
	}
	text := render(root)

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
