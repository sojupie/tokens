// Check rendered behavior requirements against the browser's diagnostic selectors.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"golang.org/x/net/html"
)

type audit struct {
	Undeclared      []string `json:"undeclared"`
	DeclarationOnly []string `json:"declarationOnly"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	catalog := flag.String("catalog", "http://127.0.0.1:8085", "Catalog base URL")
	path := flag.String("path", "", "limit to one component path")
	chrome := flag.String("chrome-path", "", "Chrome or Chromium executable")
	flag.Parse()
	base, err := url.Parse(*catalog)
	if err != nil {
		return err
	}
	if base.Host == "" || (base.Scheme != "http" && base.Scheme != "https") {
		return fmt.Errorf("catalog must be an HTTP(S) URL")
	}
	base.Path = strings.TrimSuffix(base.Path, "/") + "/"
	client := &http.Client{Timeout: 30 * time.Second}
	fetch := func(path string) (string, error) {
		u, err := base.Parse(path)
		if err != nil {
			return "", err
		}
		r, err := client.Get(u.String())
		if err != nil {
			return "", err
		}
		defer r.Body.Close()
		if r.StatusCode != 200 {
			return "", fmt.Errorf("%s: HTTP %d", u, r.StatusCode)
		}
		b, err := io.ReadAll(r.Body)
		return string(b), err
	}
	source, err := fetch("static/asset-loader.js")
	if err != nil {
		return err
	}
	options := append([]chromedp.ExecAllocatorOption{}, chromedp.DefaultExecAllocatorOptions[:]...)
	if *chrome != "" {
		options = append(options, chromedp.ExecPath(*chrome))
	}
	alloc, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()
	browser, cancel := chromedp.NewContext(alloc)
	defer cancel()
	ctx, cancel := context.WithTimeout(browser, 10*time.Minute)
	defer cancel()
	js, _ := json.Marshal(source)
	if err := chromedp.Run(ctx, chromedp.Evaluate(`globalThis.behaviorAudit=new Function('module','document',`+string(js)+`+';return module.exports')({exports:{}},undefined).auditBehaviors`, nil)); err != nil {
		return err
	}
	var paths []string
	if *path != "" {
		paths = []string{strings.TrimPrefix(*path, "/")}
	} else {
		body, err := fetch("components")
		if err != nil {
			return err
		}
		doc, err := html.Parse(strings.NewReader(body))
		if err != nil {
			return err
		}
		seen := map[string]bool{}
		walk(doc, func(n *html.Node) {
			if n.Data != "a" {
				return
			}
			u, err := base.Parse(attr(n, "href"))
			if err != nil || u.Host != base.Host {
				return
			}
			p := strings.TrimPrefix(u.Path, base.Path)
			if strings.HasPrefix(p, "components/") && !strings.Contains(strings.TrimPrefix(p, "components/"), "/") && !seen[p] {
				paths = append(paths, p)
				seen[p] = true
			}
		})
	}
	if len(paths) == 0 {
		return fmt.Errorf("no component routes found")
	}
	queue := []string{"", "components", "styles", "changes"}
	if *path != "" {
		queue = nil
	}
	for _, p := range paths {
		for _, mode := range []string{"", "states", "interactive"} {
			u, err := url.Parse(p)
			if err != nil {
				return err
			}
			if mode != "" {
				q := u.Query()
				q.Set("preview.compare", mode)
				u.RawQuery = q.Encode()
			}
			queue = append(queue, u.String())
		}
	}
	seen := map[string]bool{}
	extras := map[string]int{}
	checked, failures, negative := 0, 0, 0
	// Expand controls only for the starting views, not every combination of options.
	seeds := len(queue)
	for i := 0; i < len(queue); i++ {
		path := queue[i]
		if seen[path] {
			continue
		}
		seen[path] = true
		body, err := fetch(path)
		if err != nil {
			return err
		}
		encoded, _ := json.Marshal(body)
		var result struct {
			Audit    audit
			Negative []string
		}
		expression := `(()=>{const d=new DOMParser().parseFromString(` + string(encoded) + `, 'text/html');const a=behaviorAudit(d);d.querySelectorAll('script[data-component-behaviors]').forEach(n=>n.remove());return {Audit:a,Negative:behaviorAudit(d).undeclared}})()`
		if err := chromedp.Run(ctx, chromedp.Evaluate(expression, &result)); err != nil {
			return err
		}
		checked++
		if len(result.Audit.Undeclared) > 0 {
			failures++
			fmt.Printf("FAIL %s: %v\n", path, result.Audit.Undeclared)
		}
		if len(result.Negative) > 0 {
			negative++
		}
		for _, name := range result.Audit.DeclarationOnly {
			extras[name]++
		}
		if i < seeds {
			doc, err := html.Parse(strings.NewReader(body))
			if err != nil {
				return err
			}
			walk(doc, func(n *html.Node) {
				if n.Data != "form" || !strings.Contains(attr(n, "class"), "tokens-gallery-filters") {
					return
				}
				q := url.Values{}
				walk(n, func(c *html.Node) {
					name := attr(c, "name")
					if name == "" {
						return
					}
					if c.Data == "input" && attr(c, "type") == "hidden" {
						q.Set(name, attr(c, "value"))
					}
					if c.Data == "select" {
						first := true
						walk(c, func(o *html.Node) {
							if o.Data == "option" {
								if first || has(o, "selected") {
									q.Set(name, attr(o, "value"))
								}
								first = false
							}
						})
					}
				})
				walk(n, func(c *html.Node) {
					if c.Data != "select" {
						return
					}
					walk(c, func(o *html.Node) {
						if o.Data != "option" || has(o, "disabled") {
							return
						}
						next := url.Values{}
						for k, v := range q {
							next[k] = append([]string(nil), v...)
						}
						next.Set(attr(c, "name"), attr(o, "value"))
						u, _ := url.Parse(path)
						u.RawQuery = next.Encode()
						queue = append(queue, u.String())
					})
				})
			})
		}
	}
	if negative == 0 {
		return fmt.Errorf("negative control failed: removing declarations detected no missing behavior")
	}
	fmt.Printf("Checked %d views; %d failures; negative control detected missing requirements in %d views.\n", checked, failures, negative)
	data, _ := json.Marshal(extras)
	fmt.Printf("Declaration-only script occurrences: %s\n", data)
	if failures > 0 {
		return fmt.Errorf("rendered behavior requirements are incomplete")
	}
	return nil
}
func walk(n *html.Node, visit func(*html.Node)) {
	visit(n)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walk(c, visit)
	}
}
func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}
func has(n *html.Node, key string) bool {
	for _, a := range n.Attr {
		if a.Key == key {
			return true
		}
	}
	return false
}
