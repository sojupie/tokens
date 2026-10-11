package usage

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRetainedReportAndConcurrentRecording(t *testing.T) {
	dir := t.TempDir()
	var log bytes.Buffer
	store, err := Open(dir, &log)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Truncate(time.Hour)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				store.Record(Event{Time: now, Event: "http_request", Subject: "component/buttons", Format: "text/markdown", Kind: "markdown", Status: 200, Success: true, DurationMS: 2, Bytes: 50})
			}
		}()
	}
	wg.Wait()
	store.lost.Add(3)
	store.Record(Event{Time: now, Event: "http_request", Kind: "prefetch", Status: 200, Success: true})
	store.Record(Event{Time: now, Event: "mcp_call", Method: "tools/call", Tool: "get_tokens", Success: false, Status: 200})
	store.Record(Event{Time: now, Event: "http_request", Kind: "api", Status: 404})
	store.Record(Event{Time: now.AddDate(0, 0, -91), Event: "http_request", Kind: "old", Status: 200})
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	report, err := reopened.Report(now.AddDate(0, 0, -100), now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if report.Lost != 3 || report.Requests != 101 || report.Errors != 1 || report.Prefetches != 1 || report.Bytes != 5000 || report.Subjects["component/buttons"] != 100 || report.MCP["tools/call / get_tokens / failed"] != 1 {
		t.Fatalf("wrong report: %+v", report)
	}
	if strings.Contains(report.Text(), "old") {
		t.Fatal("expired observations survived")
	}
	if err := reopened.SaveDigest(now); err != nil {
		t.Fatal(err)
	}
	last, err := reopened.LastDigest()
	if err != nil || !last.Equal(now) {
		t.Fatalf("checkpoint lost: %v %v", last, err)
	}
	if !strings.Contains(log.String(), `"event":"http_request"`) {
		t.Fatal("missing structured log")
	}
}

func TestDigestLease(t *testing.T) {
	store, err := Open(t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	release, err := store.ClaimDigest()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.ClaimDigest(); err == nil {
		t.Fatal("concurrent digest permitted")
	}
	release()
	next, err := store.ClaimDigest()
	if err != nil {
		t.Fatal(err)
	}
	next()
	_, err = store.db.Exec("INSERT INTO metadata VALUES ('digest_lease','1')")
	if err != nil {
		t.Fatal(err)
	}
	expired, err := store.ClaimDigest()
	if err != nil {
		t.Fatal(err)
	}
	expired()
}
