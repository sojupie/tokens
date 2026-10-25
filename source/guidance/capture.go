// Package guidance captures rendered documentation independently of token payloads.
package guidance

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

const ExtractorVersion = 2

type Config struct {
	URLs       []string
	Output     string
	ChromePath string
	Selector   string
	Timeout    time.Duration
}

type Page struct {
	URL            string `json:"url"`
	FinalURL       string `json:"finalUrl,omitempty"`
	BaseURL        string `json:"baseUrl,omitempty"`
	Title          string `json:"title,omitempty"`
	Status         int    `json:"status,omitempty"`
	CapturedAt     string `json:"capturedAt"`
	HTML           string `json:"html,omitempty"`
	Text           string `json:"text,omitempty"`
	Markdown       string `json:"markdown,omitempty"`
	HTMLSHA256     string `json:"htmlSha256,omitempty"`
	TextSHA256     string `json:"textSha256,omitempty"`
	MarkdownSHA256 string `json:"markdownSha256,omitempty"`
	Error          string `json:"error,omitempty"`
}

type Receipt struct {
	SchemaVersion    int    `json:"schemaVersion"`
	ExtractorVersion int    `json:"extractorVersion"`
	StartedAt        string `json:"startedAt"`
	CompletedAt      string `json:"completedAt"`
	Selector         string `json:"selector"`
	Complete         bool   `json:"complete"`
	Pages            []Page `json:"pages"`
}

type document struct {
	BaseURL string `json:"baseUrl"`
	URL     string `json:"url"`
	Title   string `json:"title"`
	Status  int    `json:"status"`
	HTML    string `json:"html"`
	Text    string `json:"text"`
}

// Capture saves a new run, never replacing an earlier capture. Failed pages remain
// in its receipt and make the command fail; successful pages are still retained.
func Capture(ctx context.Context, cfg Config) (string, error) {
	if len(cfg.URLs) == 0 {
		return "", errors.New("provide at least one --url")
	}
	if cfg.Selector == "" {
		cfg.Selector = "main"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 45 * time.Second
	}
	if cfg.Output == "" {
		cfg.Output = "snapshots/material-guidance"
	}
	seen := map[string]bool{}
	var urls []string
	for _, raw := range cfg.URLs {
		u, err := url.Parse(raw)
		if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil {
			return "", fmt.Errorf("invalid documentation URL %q", raw)
		}
		u.Fragment = ""
		if !seen[u.String()] {
			seen[u.String()] = true
			urls = append(urls, u.String())
		}
	}
	if err := os.MkdirAll(cfg.Output, 0755); err != nil {
		return "", err
	}
	started := time.Now().UTC()
	dir, err := os.MkdirTemp(cfg.Output, ".pending-")
	if err != nil {
		return "", err
	}
	opts := append([]chromedp.ExecAllocatorOption{}, chromedp.DefaultExecAllocatorOptions[:]...)
	opts = append(opts, chromedp.Flag("headless", "new"), chromedp.Flag("disable-dev-shm-usage", true))
	if cfg.ChromePath != "" {
		opts = append(opts, chromedp.ExecPath(cfg.ChromePath))
	}
	allocator, cancelAllocator := chromedp.NewExecAllocator(ctx, opts...)
	defer cancelAllocator()
	browser, cancelBrowser := chromedp.NewContext(allocator)
	defer cancelBrowser()
	receipt := Receipt{SchemaVersion: 1, ExtractorVersion: ExtractorVersion, StartedAt: started.Format(time.RFC3339Nano), Selector: cfg.Selector, Complete: true, Pages: []Page{}}
	var failures []error
	for _, raw := range urls {
		result := Page{URL: raw}
		doc, captureErr := capturePage(browser, raw, cfg.Selector, cfg.Timeout)
		result.CapturedAt = time.Now().UTC().Format(time.RFC3339Nano)
		result.FinalURL = doc.URL
		result.BaseURL = doc.BaseURL
		result.Title = doc.Title
		result.Status = doc.Status
		if captureErr == nil {
			captureErr = savePage(dir, &result, doc)
		}
		if captureErr != nil {
			result.Error = captureErr.Error()
			receipt.Complete = false
			failures = append(failures, fmt.Errorf("%s: %w", raw, captureErr))
		}
		receipt.Pages = append(receipt.Pages, result)
	}
	receipt.CompletedAt = time.Now().UTC().Format(time.RFC3339Nano)
	data, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return dir, err
	}
	if err = os.WriteFile(filepath.Join(dir, "capture.json"), append(data, '\n'), 0644); err != nil {
		return dir, err
	}
	final := filepath.Join(cfg.Output, started.Format("20060102T150405.000000000Z")+"-"+strings.TrimPrefix(filepath.Base(dir), ".pending-"))
	if err = os.Rename(dir, final); err != nil {
		return dir, err
	}
	return final, errors.Join(failures...)
}

func capturePage(browser context.Context, raw, selector string, timeout time.Duration) (document, error) {
	tab, cancelTab := chromedp.NewContext(browser)
	defer cancelTab()
	ctx, cancel := context.WithTimeout(tab, timeout)
	defer cancel()
	var doc document
	selected, _ := json.Marshal(selector)
	// Wait for actual article text, not the JavaScript shell. Scrolling requests
	// lazy content; repeated passes stop once the document height stops growing.
	script := `(async()=>{
 const root=document.querySelector(` + string(selected) + `);
 if(!root || root.innerText.trim().length<100) return false;
 for(let pass=0;pass<3;pass++) {
  const height=document.documentElement.scrollHeight;
  for(let y=0;y<height;y+=window.innerHeight) { window.scrollTo(0,y); await new Promise(r=>setTimeout(r,100)); }
  await new Promise(r=>setTimeout(r,350));
  if(document.documentElement.scrollHeight===height) break;
 }
 window.scrollTo(0,0);
 return true;
 })()`
	// Polling the rendered HTML catches delayed tables and text after navigation.
	stable := `(()=>{
 const root=document.querySelector(` + string(selected) + `);
 if(!root || root.innerText.trim().length<100) return false;
 const text=root.innerText;
 const previous=window.__guidanceCapture;
 window.__guidanceCapture={text,at:previous?.text===text?previous.at:Date.now()};
 return previous?.text===text && Date.now()-previous.at>=1500;
 })()`
	err := chromedp.Run(ctx,
		chromedp.Navigate(raw),
		chromedp.WaitReady(selector, chromedp.ByQuery),
		chromedp.Poll(`(()=>{const root=document.querySelector(`+string(selected)+`);if(!root || root.innerText.trim().length<100)return false;return location.hostname!=="m3.material.io" || `+string(selected)+`!=="main" || root.querySelector("article")?.innerText.trim().length>=100;})()`, nil, chromedp.WithPollingInterval(250*time.Millisecond)),
		chromedp.Evaluate(script, nil, func(p *runtime.EvaluateParams) *runtime.EvaluateParams { return p.WithAwaitPromise(true) }),
		chromedp.Poll(stable, nil, chromedp.WithPollingInterval(250*time.Millisecond)),
		chromedp.Evaluate(`(()=>{const root=document.querySelector(`+string(selected)+`);return {url:location.href,baseUrl:document.baseURI,title:document.title,html:root.outerHTML,text:root.innerText,status:performance.getEntriesByType("navigation")[0]?.responseStatus||0};})()`, &doc),
	)
	if err != nil {
		return document{}, fmt.Errorf("render documentation: %w", err)
	}
	if doc.Status < 200 || doc.Status >= 300 {
		return doc, fmt.Errorf("documentation returned HTTP %d", doc.Status)
	}
	return doc, nil
}

func markdown(doc document) (string, error) {
	conv := converter.NewConverter(converter.WithPlugins(base.NewBasePlugin(), commonmark.NewCommonmarkPlugin(), table.NewTablePlugin(table.WithNewlineBehavior(table.NewlineBehaviorPreserve))))
	root, err := html.Parse(strings.NewReader(doc.HTML))
	if err != nil {
		return "", err
	}
	baseURL := doc.BaseURL
	if baseURL == "" {
		baseURL = doc.URL
	}
	origin, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	cleanContent(root, origin)
	result, err := conv.ConvertNode(root)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(result)) + "\n", nil
}

// Material's inline definition popups and section-copy controls are UI chrome,
// not article prose. Raw HTML is retained before these extraction-only removals.
func cleanContent(node *html.Node, baseURL *url.URL) {
	for child := node.FirstChild; child != nil; {
		next := child.NextSibling
		remove := child.Data == "script" || child.Data == "style" || child.Data == "noscript" || child.Data == "nav"
		for i, attr := range child.Attr {
			if attr.Key == "role" && attr.Val == "tab" {
				remove = true
			}
			if baseURL.Hostname() == "m3.material.io" && attr.Key == "class" {
				for _, class := range strings.Fields(attr.Val) {
					if class == "google-symbols" || class == "copy-button-container" || class == "tooltip-parent" {
						remove = true
					}
				}
			}
			if attr.Key == "href" || attr.Key == "src" || attr.Key == "poster" {
				if ref, err := url.Parse(attr.Val); err == nil {
					child.Attr[i].Val = baseURL.ResolveReference(ref).String()
				}
			}
		}
		if remove {
			node.RemoveChild(child)
		} else {
			cleanContent(child, baseURL)
		}
		child = next
	}
}

func digest(data string) string {
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])
}

func savePage(dir string, p *Page, doc document) error {
	md, err := markdown(doc)
	if err != nil {
		return fmt.Errorf("convert documentation: %w", err)
	}
	text := strings.TrimSpace(doc.Text) + "\n"
	if len(strings.TrimSpace(text)) < 100 || len(strings.TrimSpace(md)) < 100 {
		return errors.New("documentation content is empty or too short")
	}
	// URL-derived names avoid collisions between pages with identical titles.
	name := digest(p.URL)[:16]
	for ext, data := range map[string]string{"html": doc.HTML, "txt": text, "md": md} {
		if err := os.WriteFile(filepath.Join(dir, name+"."+ext), []byte(data), 0644); err != nil {
			return err
		}
	}
	p.HTML = name + ".html"
	p.Text = name + ".txt"
	p.Markdown = name + ".md"
	p.HTMLSHA256 = digest(doc.HTML)
	p.TextSHA256 = digest(text)
	p.MarkdownSHA256 = digest(md)
	return nil
}
