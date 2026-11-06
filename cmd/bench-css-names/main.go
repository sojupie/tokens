// Compare custom-property name lengths on frozen, rendered Catalog pages.
package main

import (
	"bytes"
	"compress/gzip"
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/performance"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"golang.org/x/net/html"
)

//go:embed measure.js
var measureJS string

type sheet struct {
	URL   string `json:"url"`
	Text  string `json:"text"`
	Order string `json:"order,omitempty"`
}
type frozenPage struct {
	Path   string   `json:"path"`
	HTML   string   `json:"html"`
	Styles []sheet  `json:"styles"`
	Inline []string `json:"inline"`
}
type corpus struct {
	Pages   []frozenPage      `json:"pages"`
	Mapping map[string]string `json:"mapping,omitempty"`
}
type size struct {
	Raw      int `json:"raw"`
	Gzip     int `json:"gzip"`
	HTMLRaw  int `json:"htmlRaw"`
	HTMLGzip int `json:"htmlGzip"`
}
type observation struct {
	Path          string             `json:"path"`
	Names         string             `json:"names"`
	Round         int                `json:"round"`
	FCP           float64            `json:"fcpMs"`
	DCL           float64            `json:"domContentLoadedMs"`
	Load          float64            `json:"loadMs"`
	Style         float64            `json:"styleMs"`
	Layout        float64            `json:"layoutMs"`
	Task          float64            `json:"taskMs"`
	Parse         []float64          `json:"parseMs"`
	Recalculate   []float64          `json:"recalculateMs"`
	Appearance    []appearanceSample `json:"appearance"`
	Appearances   map[string]string  `json:"appearanceFingerprints,omitempty"`
	CSSTransfer   float64            `json:"cssTransferBytes"`
	TotalTransfer float64            `json:"totalTransferBytes"`
	Fingerprint   string             `json:"fingerprint,omitempty"`
	Elements      int                `json:"elements"`
	UserAgent     string             `json:"userAgent"`
	DocumentMode  string             `json:"documentMode"`
}

type appearanceSample struct {
	Kind  string  `json:"kind"`
	From  string  `json:"from"`
	To    string  `json:"to"`
	Flush float64 `json:"flushMs"`
	Frame float64 `json:"frameMs"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	base := flag.String("catalog", "http://127.0.0.1:8085", "Catalog URL")
	chrome := flag.String("chrome-path", "", "Chrome executable")
	out := flag.String("output", "", "output directory (default temporary)")
	runs := flag.Int("runs", 15, "cold page loads per arm, alternating order")
	cpu := flag.Float64("cpu-rate", 1, "browser CPU slowdown factor")
	names := flag.String("names", "both", "both, original or short")
	paths := flag.String("paths", "/components/cards,/components/date-pickers?preview.compare=interactive,/components/carousel?preview.compare=interactive&preview.content=overlay", "comma-separated pages")
	mixed := flag.Bool("mixed", false, "combine rendered component previews into one benchmark page")
	listen := flag.String("listen", "", "keep generated pages available at this address after benchmarking")
	input := flag.String("input", "", "reuse a saved original.json corpus instead of capturing pages")
	flag.Parse()
	if *mixed {
		customPaths := false
		flag.Visit(func(f *flag.Flag) {
			if f.Name == "paths" {
				customPaths = true
			}
		})
		if !customPaths {
			*paths = mixedPaths
		}
	}
	if *runs < 1 || *cpu < 1 || math.IsNaN(*cpu) || math.IsInf(*cpu, 0) || (*names != "both" && *names != "original" && *names != "short") {
		return fmt.Errorf("invalid runs, cpu-rate or names")
	}
	origin, err := url.Parse(*base)
	if err != nil {
		return err
	}
	if origin.Host == "" || (origin.Scheme != "http" && origin.Scheme != "https") {
		return fmt.Errorf("invalid catalog URL")
	}
	if *out == "" {
		*out, err = os.MkdirTemp("", "tokens-css-names-")
		if err != nil {
			return err
		}
	}
	if err = os.MkdirAll(*out, 0755); err != nil {
		return err
	}
	options := append([]chromedp.ExecAllocatorOption{}, chromedp.DefaultExecAllocatorOptions[:]...)
	if *chrome != "" {
		options = append(options, chromedp.ExecPath(*chrome))
	}
	alloc, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()
	var original corpus
	if *input != "" {
		data, err := os.ReadFile(*input)
		if err != nil {
			return err
		}
		if err = json.Unmarshal(data, &original); err != nil {
			return err
		}
	} else {
		for _, path := range strings.Split(*paths, ",") {
			destination := origin.ResolveReference(&url.URL{Path: "/"})
			u, err := url.Parse(path)
			if err != nil {
				return err
			}
			destination = destination.ResolveReference(u)
			value, err := capture(alloc, destination.String())
			if err != nil {
				return fmt.Errorf("capture %s: %w", path, err)
			}
			value.Path = path
			original.Pages = append(original.Pages, value)
		}
	}
	if len(original.Pages) == 0 {
		return fmt.Errorf("empty corpus")
	}
	if *mixed && !(len(original.Pages) == 1 && original.Pages[0].Path == "/benchmark/mixed") {
		combined, err := mixedPage(original.Pages)
		if err != nil {
			return err
		}
		original.Pages = []frozenPage{combined}
	}
	if err = writeJSON(filepath.Join(*out, "original.json"), original); err != nil {
		return err
	}
	args := []string{"forge/short-css-names.ts", filepath.Join(*out, "original.json"), filepath.Join(*out, "short.json")}
	command := exec.Command("node", args...)
	command.Stderr = os.Stderr
	if err = command.Run(); err != nil {
		return err
	}
	data, err := os.ReadFile(filepath.Join(*out, "short.json"))
	if err != nil {
		return err
	}
	var short corpus
	if err = json.Unmarshal(data, &short); err != nil {
		return err
	}
	if err = writeJSON(filepath.Join(*out, "names.json"), short.Mapping); err != nil {
		return err
	}
	// Freeze one artifact per arm. No parsing, renaming or compression occurs in
	// the measured request path; both arms serve prebuilt gzip at identical URLs.
	fixtures := map[string][]byte{}
	sizes := map[string]map[string]size{}
	for arm, c := range map[string]corpus{"original": original, "short": short} {
		for index, p := range c.Pages {
			prefix := fmt.Sprintf("/%s/%d", arm, index)
			var links []string
			raw, gz := 0, 0
			for _, s := range p.Styles {
				u, _ := url.Parse(s.URL)
				key := prefix + u.Path
				if _, exists := fixtures[key]; exists {
					return fmt.Errorf("duplicate sheet URL %s", s.URL)
				}
				fixtures[key] = []byte(s.Text)
				links = append(links, `<link rel="stylesheet" href="`+key+`">`)
				raw += len(s.Text)
				encoded, _ := compress([]byte(s.Text))
				gz += len(encoded)
			}
			rendered, err := render(p, strings.Join(links, ""))
			if err != nil {
				return err
			}
			fixtures[prefix+"/index.html"] = []byte(rendered)
			if sizes[p.Path] == nil {
				sizes[p.Path] = map[string]size{}
			}
			htmlEncoded, err := compress([]byte(rendered))
			if err != nil {
				return err
			}
			sizes[p.Path][arm] = size{raw, gz, len(rendered), len(htmlEncoded)}
			if err = os.MkdirAll(filepath.Join(*out, arm, fmt.Sprint(index)), 0755); err != nil {
				return err
			}
			if err = os.WriteFile(filepath.Join(*out, arm, fmt.Sprint(index), "index.html"), []byte(rendered), 0644); err != nil {
				return err
			}
			var combined strings.Builder
			for _, s := range p.Styles {
				combined.WriteString(s.Text)
				combined.WriteByte('\n')
			}
			if err = os.WriteFile(filepath.Join(*out, arm, fmt.Sprint(index), "styles.css"), []byte(combined.String()), 0644); err != nil {
				return err
			}
		}
	}
	encoded := map[string][]byte{}
	for key, data := range fixtures {
		encoded[key], err = compress(data)
		if err != nil {
			return err
		}
	}
	client := &http.Client{Timeout: 30 * time.Second}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if body, ok := encoded[r.URL.Path]; ok {
			if strings.HasSuffix(r.URL.Path, ".css") {
				w.Header().Set("Content-Type", "text/css")
			} else {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
			}
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Cache-Control", "no-store")
			_, _ = w.Write(body)
			return
		}
		// Relative font/image URLs retain the original stylesheet directory.
		path := r.URL.Path
		parts := strings.SplitN(strings.TrimPrefix(path, "/"), "/", 3)
		if len(parts) == 3 && (parts[0] == "short" || parts[0] == "original") {
			path = "/" + parts[2]
		}
		remote := origin.ResolveReference(&url.URL{Path: path, RawQuery: r.URL.RawQuery})
		response, err := client.Get(remote.String())
		if err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		defer response.Body.Close()
		for _, key := range []string{"Content-Type", "Cache-Control"} {
			if v := response.Header.Get(key); v != "" {
				w.Header().Set(key, v)
			}
		}
		w.WriteHeader(response.StatusCode)
		_, _ = io.Copy(w, response.Body)
	}))
	if *listen != "" {
		_ = server.Listener.Close()
		server.Listener, err = net.Listen("tcp", *listen)
		if err != nil {
			return err
		}
	}
	server.Start()
	defer server.Close()
	fmt.Printf("Benchmark pages: %s/original/0/index.html and %s/short/0/index.html\n", server.URL, server.URL)
	benchmarkURL := server.URL
	if *listen != "" {
		// Keep measurements on loopback, including the secure-context SHA-256
		// equivalence check, when the review server binds a LAN address.
		local := httptest.NewServer(server.Config.Handler)
		defer local.Close()
		benchmarkURL = local.URL
	}
	var samples []observation
	for index, p := range original.Pages {
		for round := 0; round < *runs; round++ {
			arms := []string{"original", "short"}
			if round%2 == 1 {
				arms = []string{"short", "original"}
			}
			if *names != "both" {
				arms = []string{*names}
			}
			for _, arm := range arms {
				screenshot := ""
				if round == 0 {
					screenshot = filepath.Join(*out, arm, fmt.Sprint(index), "preview.png")
				}
				value, err := measure(alloc, fmt.Sprintf("%s/%s/%d/index.html", benchmarkURL, arm, index), *cpu, screenshot)
				if err != nil {
					return fmt.Errorf("%s %s round %d: %w", p.Path, arm, round, err)
				}
				value.Path = p.Path
				value.Names = arm
				value.Round = round + 1
				samples = append(samples, value)
				if round == 0 {
					if err = verifyFingerprint(samples, p.Path); err != nil {
						return err
					}
				}
			}
		}
		for _, arm := range []string{"original", "short"} {
			var fcp, parse, style []float64
			for _, s := range samples {
				if s.Path == p.Path && s.Names == arm {
					fcp = append(fcp, s.FCP)
					parse = append(parse, s.Parse...)
					style = append(style, s.Style)
				}
			}
			if len(fcp) > 0 {
				fmt.Printf("%s %s: CSS %d raw / %d gzip bytes; FCP %.1f ms; parse %.2f ms; style %.2f ms (medians)\n", p.Path, arm, sizes[p.Path][arm].Raw, sizes[p.Path][arm].Gzip, median(fcp), median(parse), median(style))
				for _, kind := range []string{"theme", "contrast"} {
					var flush, frame []float64
					for _, s := range samples {
						if s.Path == p.Path && s.Names == arm {
							for _, change := range s.Appearance {
								if change.Kind == kind {
									flush = append(flush, change.Flush)
									frame = append(frame, change.Frame)
								}
							}
						}
					}
					if len(flush) > 0 {
						fmt.Printf("  %s: forced style/layout %.2f ms; through two frames %.2f ms (medians, %d changes)\n", kind, median(flush), median(frame), len(flush))
					}
				}
			}
		}
		if err = writeJSON(filepath.Join(*out, "results.json"), map[string]any{"catalog": *base, "viewport": []int{1440, 900}, "cpuRate": *cpu, "names": len(short.Mapping), "sizes": sizes, "samples": samples}); err != nil {
			return err
		}
	}
	fmt.Println("Results:", filepath.Join(*out, "results.json"))
	if *listen != "" {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		fmt.Println("Serving benchmark pages; interrupt to stop.")
		<-ctx.Done()
	}
	return nil
}

func capture(alloc context.Context, destination string) (frozenPage, error) {
	ctx, cancel := chromedp.NewContext(alloc)
	defer cancel()
	ctx, timeout := context.WithTimeout(ctx, 45*time.Second)
	defer timeout()
	var p frozenPage
	js := `(async()=>{await TokenAssets.ready;await document.fonts.ready;await new Promise(r=>setTimeout(r,350));const styles=[];let i=0;for(const sheet of document.styleSheets){if(sheet.disabled||sheet.media.mediaText&&!matchMedia(sheet.media.mediaText).matches)continue;const url=sheet.href||new URL('__inline-'+i+++'.css',location.href).href;const text=sheet.href?await (await fetch(url)).text():sheet.ownerNode.textContent;if(/@import\b/.test(text))throw Error('Import must be captured explicitly');styles.push({url,text,order:sheet.ownerNode.dataset.componentOrder||''});}return {html:document.documentElement.outerHTML,styles,inline:[...document.querySelectorAll('[style]')].map(e=>e.getAttribute('style'))};})()`
	err := chromedp.Run(ctx, emulation.SetDeviceMetricsOverride(1440, 900, 1, false), chromedp.Navigate(destination), chromedp.Evaluate(js, &p, awaitPromise))
	return p, err
}
func awaitPromise(p *runtime.EvaluateParams) *runtime.EvaluateParams { return p.WithAwaitPromise(true) }
func measure(alloc context.Context, destination string, cpu float64, screenshot string) (observation, error) {
	ctx, cancel := chromedp.NewContext(alloc)
	defer cancel()
	ctx, timeout := context.WithTimeout(ctx, 60*time.Second)
	defer timeout()
	var value observation
	var counters []*performance.Metric
	err := chromedp.Run(ctx, network.Enable(), network.ClearBrowserCache(), performance.Enable(), emulation.SetDeviceMetricsOverride(1440, 900, 1, false), emulation.SetCPUThrottlingRate(cpu), chromedp.Navigate(destination),
		chromedp.Evaluate(`(async()=>{await document.fonts.ready;await Promise.all([...document.images].filter(i=>i.getBoundingClientRect().top<innerHeight).map(i=>i.decode().catch(()=>{})));await new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(()=>r(true))));return true})()`, nil, awaitPromise),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			counters, err = performance.GetMetrics().Do(ctx)
			return err
		}), chromedp.Evaluate(measureJS, &value, awaitPromise))
	if err != nil {
		return value, err
	}
	if value.DocumentMode != "CSS1Compat" {
		return value, fmt.Errorf("benchmark requires standards mode, got %q", value.DocumentMode)
	}
	for _, m := range counters {
		switch m.Name {
		case "RecalcStyleDuration":
			value.Style = m.Value * 1000
		case "LayoutDuration":
			value.Layout = m.Value * 1000
		case "TaskDuration":
			value.Task = m.Value * 1000
		}
	}
	// CDP counters are taken before the microbenchmarks and equivalence check.
	if screenshot != "" {
		err = chromedp.Run(ctx, chromedp.Evaluate(`(async()=>{const result={},root=document.documentElement,theme=root.dataset.theme,contrast=root.dataset.contrast;const props=[...getComputedStyle(document.body)].filter(p=>!p.startsWith('--'));for(const t of ['light','dark'])for(const c of ['default','medium','high']){root.dataset.theme=t;root.dataset.contrast=c;const rows=[...document.querySelectorAll('*')].filter(e=>!['SCRIPT','STYLE','LINK','HEAD','META','TITLE'].includes(e.tagName)).map(e=>{const s=getComputedStyle(e),r=e.getBoundingClientRect();return [e.tagName,r.x,r.y,r.width,r.height,...props.map(p=>s.getPropertyValue(p))]});const hash=await crypto.subtle.digest('SHA-256',new TextEncoder().encode(JSON.stringify(rows)));result[t+'/'+c]=[...new Uint8Array(hash)].map(x=>x.toString(16).padStart(2,'0')).join('')}if(theme===undefined)delete root.dataset.theme;else root.dataset.theme=theme;if(contrast===undefined)delete root.dataset.contrast;else root.dataset.contrast=contrast;return result})()`, &value.Appearances, awaitPromise))
		value.Fingerprint = value.Appearances["light/default"]
		if err == nil {
			var shot []byte
			err = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error { var e error; shot, e = page.CaptureScreenshot().Do(ctx); return e }))
			if err == nil {
				err = os.WriteFile(screenshot, shot, 0644)
			}
		}
	}
	return value, err
}
func verifyFingerprint(samples []observation, path string) error {
	var expected string
	appearances := map[string]string{}
	for _, s := range samples {
		if s.Path == path && s.Fingerprint != "" {
			if expected != "" && expected != s.Fingerprint {
				return fmt.Errorf("computed styles or geometry differ for %s", path)
			}
			expected = s.Fingerprint
			for key, value := range s.Appearances {
				if previous, ok := appearances[key]; ok && previous != value {
					return fmt.Errorf("computed styles or geometry differ for %s %s", path, key)
				}
				appearances[key] = value
			}
		}
	}
	return nil
}
func render(p frozenPage, links string) (string, error) {
	root, err := html.Parse(strings.NewReader(p.HTML))
	if err != nil {
		return "", err
	}
	index := 0
	var inlineError error
	var visit func(*html.Node)
	visit = func(n *html.Node) {
		if name := attr(n, "data-benchmark-appearance"); name != "" {
			n.Attr = append(n.Attr, html.Attribute{Key: "aria-label", Val: strings.ToUpper(name[:1]) + name[1:]})
		}
		for i := range n.Attr {
			if n.Attr[i].Key == "style" {
				if index >= len(p.Inline) {
					inlineError = fmt.Errorf("missing inline style %d", index)
					return
				}
				n.Attr[i].Val = p.Inline[index]
				index++
			}
		}
		for c := n.FirstChild; c != nil; {
			next := c.NextSibling
			visit(c)
			if c.Type == html.ElementNode && (c.Data == "script" || c.Data == "style" || c.Data == "link" || c.Data == "base") {
				n.RemoveChild(c)
			}
			c = next
		}
	}
	visit(root)
	if inlineError != nil {
		return "", inlineError
	}
	if index != len(p.Inline) {
		return "", fmt.Errorf("inline style count changed: %d != %d", index, len(p.Inline))
	}
	if p.Path == "/benchmark/mixed" {
		if title := find(root, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "title" }); title != nil {
			for title.FirstChild != nil {
				title.RemoveChild(title.FirstChild)
			}
			title.AppendChild(&html.Node{Type: html.TextNode, Data: "Component delivery benchmark"})
		}
	}
	var out bytes.Buffer
	// outerHTML excludes the Catalog's doctype. Restore standards mode when
	// replaying captures rather than benchmarking quirks-mode layout.
	hasDoctype := false
	for node := root.FirstChild; node != nil; node = node.NextSibling {
		hasDoctype = hasDoctype || node.Type == html.DoctypeNode
	}
	if !hasDoctype {
		out.WriteString("<!DOCTYPE html>")
	}
	if err = html.Render(&out, root); err != nil {
		return "", err
	}
	// Freeze animated decoration in both arms; no runtime translation is installed.
	extra := links + `<style>*,*::before,*::after{animation:none!important;transition:none!important;caret-color:transparent!important}</style>`
	// The benchmark controls change root attributes only, never custom-property
	// names. Both arms execute the same small controller.
	extra += `<script>document.addEventListener('change',e=>{const k=e.target.dataset.benchmarkAppearance;if(k)document.documentElement.dataset[k]=e.target.value});</script>`
	return strings.Replace(out.String(), "</head>", extra+"</head>", 1), nil
}
func compress(data []byte) ([]byte, error) {
	var out bytes.Buffer
	w := gzip.NewWriter(&out)
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
func median(values []float64) float64 {
	sort.Float64s(values)
	n := len(values)
	if n%2 == 1 {
		return values[n/2]
	}
	return (values[n/2-1] + values[n/2]) / 2
}
