// Browser capture observes network payloads and page activity together; quiet network traffic alone does not prove a complete capture.
package crawl

import (
	"context"

	"fmt"

	"sort"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type pageProgress struct {
	mu        sync.Mutex
	total     int
	started   int
	completed int
}

func (progress *pageProgress) start(pageURL string) {
	progress.mu.Lock()
	defer progress.mu.Unlock()
	progress.started++
	fmt.Printf("  [page %d/%d] observing %s\n", progress.started, progress.total, pageURL)
}

func (progress *pageProgress) finish(observation Observation, duration time.Duration) {
	progress.mu.Lock()
	defer progress.mu.Unlock()
	progress.completed++
	status := fmt.Sprintf("HTTP %d", observation.Status)
	if observation.Status == 0 {
		status = "no HTTP status"
	}
	if observation.Error != "" {
		status = "failed: " + observation.Error
	} else if len(observation.SourcePayloadURLs) == 0 {
		status += ", no source payload"
	} else {
		status += fmt.Sprintf(", %d source payloads observed", len(observation.SourcePayloadURLs))
	}
	fmt.Printf("  [page %d/%d] %s (%s)\n", progress.completed, progress.total, status, elapsed(duration))
}

func elapsed(duration time.Duration) string {
	return duration.Round(10 * time.Millisecond).String()
}

func observePages(parent context.Context, config Config, pages []string) ([]Observation, error) {
	fmt.Printf("Starting Chromium for %d pages...\n", len(pages))
	headless := chromedp.Flag("headless", "new")
	if config.Headed {
		headless = chromedp.Flag("headless", false)
	}
	allocatorOptions := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		headless,
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-gpu-compositing", true),
		chromedp.Flag("disable-gpu-rasterization", true),
		chromedp.Flag("disable-gpu-sandbox", true),
		chromedp.Flag("disable-3d-apis", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
	}
	if config.ChromePath != "" {
		allocatorOptions = append(allocatorOptions, chromedp.ExecPath(config.ChromePath))
	}
	allocatorContext, cancelAllocator := chromedp.NewExecAllocator(parent, allocatorOptions...)
	defer cancelAllocator()
	browserContext, cancelBrowser := chromedp.NewContext(allocatorContext)
	defer cancelBrowser()
	// Initialize the parent so page contexts reuse one browser instead of
	// allocating a separate Chromium process for each page.
	if err := chromedp.Run(browserContext); err != nil {
		return nil, err
	}

	fmt.Printf("Observing %d pages with concurrency %d...\n", len(pages), config.Concurrency)
	progress := &pageProgress{total: len(pages)}
	return mapLimit(pages, config.Concurrency, func(pageURL string) Observation {
		started := time.Now()
		progress.start(pageURL)
		observation := observePage(browserContext, pageURL, config.PageTimeout)
		progress.finish(observation, time.Since(started))
		return observation
	})
}

func observePage(browserContext context.Context, pageURL string, timeout time.Duration) Observation {
	pageContext, cancel := chromedp.NewContext(browserContext)
	defer cancel()
	observation := Observation{PageURL: pageURL}
	seenPayloadURLs := map[string]bool{}
	var payloadMu sync.Mutex
	activity := captureActivity{pageURL: pageURL, pending: map[network.RequestID]bool{}}
	var statusMu sync.Mutex
	mainStatus := 0
	chromedp.ListenTarget(pageContext, func(event interface{}) {
		payloadMu.Lock()
		activity.observe(event, time.Now())
		payloadMu.Unlock()
		if response, ok := event.(*network.EventResponseReceived); ok {
			if response.Response.MimeType == "text/html" || response.Response.URL == pageURL {
				statusMu.Lock()
				mainStatus = int(response.Response.Status)
				statusMu.Unlock()
			}
			if sourcePayloadRE.MatchString(response.Response.URL) {
				payloadMu.Lock()
				seenPayloadURLs[response.Response.URL] = true
				payloadMu.Unlock()
			}
		}
	})

	operationContext, cancelOperation := context.WithTimeout(pageContext, timeout+5*time.Second)
	defer cancelOperation()
	err := chromedp.Run(operationContext,
		network.Enable(),
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, _, errorText, _, err := page.Navigate(pageURL).Do(ctx)
			if err != nil {
				return err
			}
			if errorText != "" {
				return fmt.Errorf("page load error %s", errorText)
			}
			return nil
		}),
	)
	if err != nil {
		observation.Error = err.Error()
	}
	if err == nil {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		poll := time.NewTicker(250 * time.Millisecond)
		defer poll.Stop()
		started := time.Now()
	waitForCapture:
		for {
			select {
			case <-timer.C:
				break waitForCapture
			case <-operationContext.Done():
				observation.Error = operationContext.Err().Error()
				break waitForCapture
			case now := <-poll.C:
				payloadMu.Lock()
				settled := activity.settled(now, started)
				payloadMu.Unlock()
				if !settled {
					continue
				}
				var loaded bool
				// DOM readiness plus main-frame network quiet avoids waiting for
				// unrelated third-party iframe load events.
				if err := chromedp.Run(operationContext, chromedp.Evaluate(`document.readyState !== "loading"`, &loaded)); err != nil {
					observation.Error = err.Error()
					break waitForCapture
				}
				// Recheck: evaluation can race with another payload request.
				payloadMu.Lock()
				settled = activity.settled(time.Now(), started)
				payloadMu.Unlock()
				if loaded && settled {
					break waitForCapture
				}
			}
		}
	}
	payloadMu.Lock()
	observation.SourcePayloadURLs = make([]string, 0, len(seenPayloadURLs))
	for payloadURL := range seenPayloadURLs {
		observation.SourcePayloadURLs = append(observation.SourcePayloadURLs, payloadURL)
	}
	payloadMu.Unlock()
	sort.Strings(observation.SourcePayloadURLs)
	statusMu.Lock()
	observation.Status = mainStatus
	statusMu.Unlock()
	return observation
}

// captureActivity is guarded by observePage's payload mutex.
type captureActivity struct {
	pageURL      string
	mainFrame    string
	pending      map[network.RequestID]bool
	lastActivity time.Time
	sawPayload   bool
}

func (activity *captureActivity) observe(event interface{}, now time.Time) {
	switch event := event.(type) {
	case *network.EventRequestWillBeSent:
		if event.Type == network.ResourceTypeDocument && event.Request.URL == activity.pageURL {
			activity.mainFrame = string(event.FrameID)
		}
		// Scripts and fetches in the main frame can discover more tables.
		// Iframes must not hold up capture, but always wait for token data.
		if (string(event.FrameID) == activity.mainFrame || sourcePayloadRE.MatchString(event.Request.URL)) &&
			(strings.HasPrefix(event.Request.URL, "http://") || strings.HasPrefix(event.Request.URL, "https://")) {
			activity.pending[event.RequestID] = true
			activity.lastActivity = now
		}
	case *network.EventResponseReceived:
		if sourcePayloadRE.MatchString(event.Response.URL) {
			activity.sawPayload = true
			activity.lastActivity = now
		}
	case *network.EventLoadingFinished:
		if activity.pending[event.RequestID] {
			delete(activity.pending, event.RequestID)
			activity.lastActivity = now
		}
	case *network.EventLoadingFailed:
		if activity.pending[event.RequestID] {
			delete(activity.pending, event.RequestID)
			activity.lastActivity = now
		}
	}
}

func (activity *captureActivity) settled(now, started time.Time) bool {
	// ponytail: quiet is a heuristic, not proof that no later timer can fetch
	// another table. Keep coverage checks; add source-specific readiness if needed.
	return activity.sawPayload && len(activity.pending) == 0 &&
		now.Sub(started) >= 3*time.Second && now.Sub(activity.lastActivity) >= 2*time.Second
}

func mapLimit(values []string, concurrency int, worker func(string) Observation) ([]Observation, error) {
	if len(values) == 0 {
		return []Observation{}, nil
	}
	if concurrency > len(values) {
		concurrency = len(values)
	}
	results := make([]Observation, len(values))
	jobs := make(chan int)
	var wait sync.WaitGroup
	for workerIndex := 0; workerIndex < concurrency; workerIndex++ {
		wait.Go(func() {
			for index := range jobs {
				results[index] = worker(values[index])
			}
		})
	}
	for index := range values {
		jobs <- index
	}
	close(jobs)
	wait.Wait()
	return results, nil
}
