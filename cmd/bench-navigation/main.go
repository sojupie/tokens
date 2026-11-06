// Benchmark Catalog interactions in Chrome; write raw samples and DevTools traces.
package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/chromedp/cdproto/emulation"
	cdpio "github.com/chromedp/cdproto/io"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/cdproto/tracing"
	"github.com/chromedp/chromedp"
)

//go:embed probe.js
var probe string

type journey struct{ Name, From, To, Kind string }
type sample struct {
	UserAgent string             `json:"userAgent"`
	Viewport  map[string]int     `json:"viewport"`
	Journey   string             `json:"journey"`
	Scenario  string             `json:"scenario"`
	Round     int                `json:"round"`
	Trace     string             `json:"trace,omitempty"`
	Marks     map[string]float64 `json:"marks"`
	LongTasks []json.RawMessage  `json:"longTasks"`
	Resources []resource         `json:"resources"`
}
type resource struct {
	Protocol      string            `json:"protocol"`
	DecodedBytes  int64             `json:"decodedBytes"`
	URL           string            `json:"url"`
	Start         float64           `json:"start"`
	Duration      float64           `json:"duration"`
	TransferBytes int64             `json:"transferBytes"`
	EncodedBytes  int64             `json:"encodedBytes"`
	ServerTiming  []json.RawMessage `json:"serverTiming"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	baseFlag := flag.String("catalog", "http://127.0.0.1:8085", "Catalog base URL")
	chrome := flag.String("chrome-path", "", "Chrome or Chromium executable")
	runs := flag.Int("runs", 15, "measured runs per journey/scenario")
	out := flag.String("output", "", "output directory; default is a temporary directory")
	cpu := flag.Float64("cpu-rate", 1, "CPU slowdown factor, at least 1")
	latency := flag.Float64("latency-ms", 0, "added network latency; throughput is unchanged")
	only := flag.String("journey", "", "cards, date or download; default all")
	width := flag.Int("width", 1440, "viewport width in CSS pixels")
	height := flag.Int("height", 900, "viewport height in CSS pixels")
	traces := flag.Bool("traces", true, "capture one additional traced run per scenario, excluded from statistics")
	flag.Parse()
	if *width < 320 || *height < 320 || *runs < 1 || *cpu < 1 || *latency < 0 || math.IsNaN(*cpu) || math.IsNaN(*latency) || math.IsInf(*cpu, 0) || math.IsInf(*latency, 0) {
		return fmt.Errorf("invalid run count or throttling")
	}
	base, err := url.Parse(*baseFlag)
	if err != nil {
		return err
	}
	if base.Host == "" || (base.Scheme != "http" && base.Scheme != "https") {
		return fmt.Errorf("catalog must be an HTTP(S) URL")
	}
	base.Path = strings.TrimSuffix(base.Path, "/") + "/"
	if *out == "" {
		*out, err = os.MkdirTemp("", "tokens-navigation-")
		if err != nil {
			return err
		}
	} else if err = os.MkdirAll(*out, 0755); err != nil {
		return err
	}
	options := append([]chromedp.ExecAllocatorOption{}, chromedp.DefaultExecAllocatorOptions[:]...)
	if *chrome != "" {
		options = append(options, chromedp.ExecPath(*chrome))
	}
	alloc, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()
	journeys := []journey{{"cards", "components", "components/cards", "navigation"}, {"date", "components/cards", "components/date-pickers?preview.compare=interactive&preview.variant=modal", "navigation"}, {"download", "components/cards", "", "download"}}
	var results []sample
	write := func() error {
		data, e := json.MarshalIndent(struct {
			Catalog string   `json:"catalog"`
			CPU     float64  `json:"cpuRate"`
			Latency float64  `json:"latencyMs"`
			Created string   `json:"created"`
			Samples []sample `json:"samples"`
		}{base.String(), *cpu, *latency, time.Now().UTC().Format(time.RFC3339), results}, "", "  ")
		if e != nil {
			return e
		}
		return os.WriteFile(filepath.Join(*out, "results.json"), data, 0644)
	}
	found := false
	for _, j := range journeys {
		if *only != "" && j.Name != *only {
			continue
		}
		found = true
		scenarios := []string{"cold", "warm", "prefetch", "expired"}
		if j.Kind == "download" {
			scenarios = scenarios[:2]
		}
		for _, scenario := range scenarios {
			n := *runs
			if *traces {
				n++
			}
			for i := 0; i < n; i++ {
				tracePath := ""
				if i == *runs {
					tracePath = filepath.Join(*out, j.Name+"-"+scenario+".trace.json")
				}
				value, e := measure(alloc, base, j, scenario, *cpu, *latency, *width, *height, tracePath)
				if e != nil {
					_ = write()
					return fmt.Errorf("%s/%s run %d: %w", j.Name, scenario, i+1, e)
				}
				value.Journey = j.Name
				value.Scenario = scenario
				value.Round = i + 1
				value.Trace = tracePath
				results = append(results, value)
				if e = write(); e != nil {
					return e
				}
			}
			var ready, dom []float64
			for _, s := range results {
				if s.Journey == j.Name && s.Scenario == scenario && s.Trace == "" {
					ready = append(ready, s.Marks["ready"])
					dom = append(dom, s.Marks["dom"])
				}
			}
			fmt.Printf("%s %-8s DOM p50/p95 %.1f/%.1f ms; ready %.1f/%.1f ms (n=%d)\n", j.Name, scenario, percentile(dom, .5), percentile(dom, .95), percentile(ready, .5), percentile(ready, .95), len(ready))
		}
	}
	if !found {
		return fmt.Errorf("unknown journey %q", *only)
	}
	fmt.Println("Results:", filepath.Join(*out, "results.json"))
	return nil
}
func percentile(values []float64, p float64) float64 {
	sort.Float64s(values)
	return values[max(0, int(math.Ceil(float64(len(values))*p))-1)]
}
func measure(alloc context.Context, base *url.URL, j journey, scenario string, cpu, latency float64, width, height int, tracePath string) (sample, error) {
	browser, cancel := chromedp.NewContext(alloc)
	defer cancel()
	ctx, cancel := context.WithTimeout(browser, 60*time.Second)
	defer cancel()
	run := func(actions ...chromedp.Action) error { return chromedp.Run(ctx, actions...) }
	evaluate := func(js string) error {
		return run(chromedp.Evaluate(js, nil, func(p *runtime.EvaluateParams) *runtime.EvaluateParams { return p.WithAwaitPromise(true) }))
	}
	dest := func(path string) string { u, _ := base.Parse(path); return u.String() }
	quote := func(s string) string { v, _ := json.Marshal(s); return string(v) }
	ready := func() error {
		return evaluate(`TokenAssets.ready.then(()=>document.fonts.ready).then(()=>new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(()=>r(true)))))`)
	}
	var result sample
	if err := run(network.Enable(), network.ClearBrowserCache(), emulation.SetCPUThrottlingRate(cpu), emulation.SetDeviceMetricsOverride(int64(width), int64(height), 1, false), network.EmulateNetworkConditions(false, latency, -1, -1), chromedp.ActionFunc(func(ctx context.Context) error {
		_, err := page.AddScriptToEvaluateOnNewDocument(probe).Do(ctx)
		return err
	})); err != nil {
		return result, err
	}
	if scenario != "cold" {
		path := j.To
		if j.Kind == "download" {
			path = j.From
		}
		if err := run(chromedp.Navigate(dest(path))); err != nil {
			return result, err
		}
		if err := ready(); err != nil {
			return result, err
		}
	}
	if err := run(chromedp.Navigate(dest(j.From))); err != nil {
		return result, err
	}
	if err := ready(); err != nil {
		return result, err
	}
	if j.Kind == "download" {
		if err := evaluate(`document.querySelector('[data-export-open]').click();true`); err != nil {
			return result, err
		}
		if err := run(chromedp.Poll(`document.querySelector('#tokens-site-download').open && !document.querySelector('#tokens-export-fields fieldset:disabled')`, nil)); err != nil {
			return result, err
		}
	} else {
		path, _ := url.Parse(j.To)
		selector := `a[href="` + strings.TrimSuffix(base.Path, "/") + "/" + path.Path + `"]`
		if err := evaluate(`(()=>{const old=document.querySelector(` + quote(selector) + `);if(!old)throw Error('missing benchmark link');const link=old.cloneNode(true);link.href=` + quote(dest(j.To)) + `;old.replaceWith(link);htmx.process(link);globalThis.benchmarkLink=link;return true})()`); err != nil {
			return result, err
		}
		if scenario == "prefetch" || scenario == "expired" {
			delay := 300
			if scenario == "expired" {
				delay = 5400
			}
			if err := evaluate(fmt.Sprintf(`new Promise(resolve=>{benchmarkLink.dispatchEvent(new PointerEvent('pointerover',{bubbles:true,pointerType:'mouse'}));setTimeout(()=>resolve(true),%d)})`, delay)); err != nil {
				return result, err
			}
		}
	}
	var complete chan *tracing.EventTracingComplete
	if tracePath != "" {
		complete = make(chan *tracing.EventTracingComplete, 1)
		chromedp.ListenTarget(ctx, func(e any) {
			if v, ok := e.(*tracing.EventTracingComplete); ok {
				complete <- v
			}
		})
		if err := run(tracing.Start().WithTransferMode(tracing.TransferModeReturnAsStream).WithTraceConfig(&tracing.TraceConfig{EnableSampling: true, IncludedCategories: []string{"devtools.timeline", "blink.user_timing", "loading", "disabled-by-default-devtools.timeline", "disabled-by-default-devtools.screenshot"}})); err != nil {
			return result, err
		}
	}
	path, _ := url.Parse(dest(j.To))
	action := `benchmarkLink.click()`
	if j.Kind == "download" {
		action = `document.querySelector('input[name=content][value=implementation]').click()`
	}
	expected := ".md-card"
	if j.Name == "date" {
		expected = ".md-date-demo"
	}
	if err := evaluate(`TokenBenchmark.reset(` + quote(j.Kind) + `,` + quote(path.Path) + `,` + quote(expected) + `);` + action + `;true`); err != nil {
		return result, err
	}
	if err := run(chromedp.Poll(`TokenBenchmark.done`, nil, chromedp.WithPollingInterval(10*time.Millisecond)), chromedp.Evaluate(`TokenBenchmark.result()`, &result)); err != nil {
		return result, err
	}
	if tracePath != "" {
		// Keep recording after readiness so screenshots include the presented page
		// and any transition, rather than stopping before its next paint.
		if err := evaluate(`new Promise(resolve=>setTimeout(()=>resolve(true),500))`); err != nil {
			return result, err
		}
		if err := run(tracing.End()); err != nil {
			return result, err
		}
		var e *tracing.EventTracingComplete
		select {
		case e = <-complete:
		case <-ctx.Done():
			return result, ctx.Err()
		}
		if e.DataLossOccurred {
			return result, fmt.Errorf("trace lost events")
		}
		file, err := os.Create(tracePath)
		if err != nil {
			return result, err
		}
		defer file.Close()
		err = run(chromedp.ActionFunc(func(ctx context.Context) error {
			defer cdpio.Close(e.Stream).Do(ctx)
			for {
				data, eof, err := cdpio.Read(e.Stream).Do(ctx)
				if err != nil {
					return err
				}
				if _, err = file.WriteString(data); err != nil {
					return err
				}
				if eof {
					return nil
				}
			}
		}))
		if err != nil {
			return result, err
		}
	}
	if j.Name == "date" {
		if err := evaluate(`document.querySelector('[data-date-open]').click();if(!document.querySelector('.md-date-panel').open)throw Error('date picker not usable');true`); err != nil {
			return result, err
		}
	}
	if j.Kind == "navigation" {
		// Verify the scenario actually exercised a reusable or expired preload.
		requests := 0
		before, after := false, false
		for _, r := range result.Resources {
			if r.URL == dest(j.To) {
				requests++
				before = before || r.Start < 0
				after = after || r.Start >= 0
			}
		}
		if scenario == "prefetch" && (!before || after || requests != 1) {
			return result, fmt.Errorf("prefetch was not reused: %d requests", requests)
		}
		if scenario == "expired" && (!before || !after) {
			return result, fmt.Errorf("expired preload did not fall back to a request")
		}
	}
	return result, nil
}
