package crawl

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func TestCaptureActivitySettlesOnlyAfterPayloadAndCompletedRequests(t *testing.T) {
	started := time.Unix(0, 0)
	activity := captureActivity{pending: map[network.RequestID]bool{}}
	if activity.settled(started.Add(time.Minute), started) {
		t.Fatal("an empty page must not finish early")
	}
	request := &network.EventRequestWillBeSent{RequestID: "table", Request: &network.Request{URL: "https://example.test/_dsm/data/ds/TOKEN_TABLE.first.json"}}
	activity.observe(request, started)
	activity.observe(&network.EventResponseReceived{RequestID: "table", Response: &network.Response{URL: request.Request.URL}}, started.Add(time.Second))
	if activity.settled(started.Add(5*time.Second), started) {
		t.Fatal("response headers are not a completed payload")
	}
	activity.observe(&network.EventLoadingFinished{RequestID: "table"}, started.Add(5*time.Second))
	if activity.settled(started.Add(6*time.Second), started) {
		t.Fatal("must wait for network quiet after completion")
	}
	if !activity.settled(started.Add(7*time.Second), started) {
		t.Fatal("completed capture should now settle")
	}
	activity.observe(&network.EventRequestWillBeSent{RequestID: "late", Request: &network.Request{URL: "https://example.test/module.js"}}, started.Add(7*time.Second))
	if activity.settled(started.Add(20*time.Second), started) {
		t.Fatal("a late request must prevent early exit")
	}
	activity.observe(&network.EventLoadingFailed{RequestID: "late"}, started.Add(20*time.Second))
	if !activity.settled(started.Add(22*time.Second), started) {
		t.Fatal("failed requests must not keep capture pending forever")
	}
	activity.lastActivity = started
	if activity.settled(started.Add(2500*time.Millisecond), started) {
		t.Fatal("must leave a minimum window for initial lazy loads")
	}
}

func TestCaptureActivityIgnoresUnrelatedFramesButWaitsForTheirPayloads(t *testing.T) {
	started := time.Unix(0, 0)
	activity := captureActivity{pageURL: "https://example.test/specs", pending: map[network.RequestID]bool{}}
	activity.observe(&network.EventRequestWillBeSent{RequestID: "main", FrameID: "main", Type: network.ResourceTypeDocument, Request: &network.Request{URL: activity.pageURL}}, started)
	activity.observe(&network.EventLoadingFinished{RequestID: "main"}, started)
	activity.sawPayload = true
	activity.observe(&network.EventRequestWillBeSent{RequestID: "feedback", FrameID: "iframe", Request: &network.Request{URL: "https://feedback.test/proxy.html"}}, started.Add(time.Second))
	if !activity.settled(started.Add(3*time.Second), started) {
		t.Fatal("unrelated iframe must not hold up capture")
	}
	activity.observe(&network.EventRequestWillBeSent{RequestID: "table", FrameID: "iframe", Request: &network.Request{URL: "https://example.test/_dsm/data/ds/TOKEN_TABLE.embedded.json"}}, started.Add(3*time.Second))
	if activity.settled(started.Add(time.Minute), started) {
		t.Fatal("token requests must be awaited even in an iframe")
	}
}

// Opt-in, local-only browser check. Ordinary tests do not launch Chromium.
// TOKENS_TEST_CHROME_PATH=/usr/bin/chromium go test ./internal/crawl -run TestObservePageBrowser -v
func TestObservePageBrowser(t *testing.T) {
	chromePath := os.Getenv("TOKENS_TEST_CHROME_PATH")
	if chromePath == "" {
		t.Skip("set TOKENS_TEST_CHROME_PATH to run the local browser regression check")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/fast":
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<html><script>fetch('/_dsm/data/ds/TOKEN_TABLE.first.json')</script></html>`)
		case "/slow":
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<html><script>fetch('/_dsm/data/ds/TOKEN_TABLE.first.json'); setTimeout(() => fetch('/_dsm/data/ds/TOKEN_TABLE.slow.json'), 1200)</script></html>`)
		case "/empty":
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<html>no token data</html>`)
		case "/feedback":
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<html><script>fetch('/_dsm/data/ds/TOKEN_TABLE.first.json')</script><iframe src='/hanging-frame'></iframe></html>`)
		case "/hanging-frame":
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<html>feedback`)
			w.(http.Flusher).Flush()
			<-r.Context().Done()
		case "/_dsm/data/ds/TOKEN_TABLE.first.json", "/_dsm/data/ds/TOKEN_TABLE.slow.json":
			w.Header().Set("Content-Type", "application/json")
			if strings.Contains(r.URL.Path, "slow") {
				w.WriteHeader(http.StatusOK)
				w.(http.Flusher).Flush()
				time.Sleep(2500 * time.Millisecond)
			}
			fmt.Fprint(w, `{"system":{"tokens":[]}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	options := append([]chromedp.ExecAllocatorOption{}, chromedp.DefaultExecAllocatorOptions[:]...)
	options = append(options, chromedp.ExecPath(chromePath), chromedp.Flag("disable-gpu", true))
	allocator, stopAllocator := chromedp.NewExecAllocator(context.Background(), options...)
	defer stopAllocator()
	browser, stopBrowser := chromedp.NewContext(allocator)
	defer stopBrowser()
	if err := chromedp.Run(browser); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		path     string
		limit    time.Duration
		payloads int
	}{
		{"/fast", 8 * time.Second, 1},
		{"/slow", 8 * time.Second, 2},
		{"/feedback", 8 * time.Second, 1},
		{"/empty", time.Second, 0},
	} {
		t.Run(test.path, func(t *testing.T) {
			started := time.Now()
			observation := observePage(browser, server.URL+test.path, test.limit)
			duration := time.Since(started)
			if observation.Error != "" || observation.Status != http.StatusOK || len(observation.SourcePayloadURLs) != test.payloads {
				t.Fatalf("unexpected observation: %+v", observation)
			}
			if test.payloads > 0 && duration >= test.limit {
				t.Fatalf("capture did not finish early: %s", duration)
			}
			if test.payloads == 0 && duration < test.limit {
				t.Fatalf("empty page ended before its deadline: %s", duration)
			}
			t.Logf("%d payloads in %s (cap %s)", test.payloads, duration.Round(time.Millisecond), test.limit)
		})
	}
	canceled, cancel := context.WithCancel(browser)
	cancel()
	if observation := observePage(canceled, server.URL+"/empty", time.Second); observation.Error == "" {
		t.Fatal("cancellation must be reported, not treated as a settled page")
	}
}
