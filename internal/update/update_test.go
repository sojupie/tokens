package update

import (
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sojupie/tokens/internal/crawl"
	"github.com/sojupie/tokens/internal/publish"
)

func receiptFixture(t *testing.T, width int, partial bool) string {
	t.Helper()
	root := t.TempDir()
	payload := []byte(fmt.Sprintf(`{"system":{"components":[{"name":"designSystems/test/components/snackbar","displayName":"Snackbar","tokenSets":["designSystems/test/tokenSets/snackbar"]}],"tokenSets":[{"name":"designSystems/test/tokenSets/snackbar","displayName":"Snackbar"}],"tokens":[{"name":"designSystems/test/tokenSets/snackbar/tokens/width","tokenName":"md.comp.snackbar.container.width"}],"contextualReferenceTrees":{"designSystems/test/tokenSets/snackbar/tokens/width":{"contextualReferenceTree":[{"referenceTree":{"tokenName":"md.comp.snackbar.container.width"},"resolvedValue":{"length":{"unit":"DIPS","value":%d}}}]}}}}`, width))
	file, err := os.Create(filepath.Join(root, "payload.json.gz"))
	if err != nil {
		t.Fatal(err)
	}
	writer := gzip.NewWriter(file)
	if _, err := writer.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	page := "https://m3.material.io/components/snackbar/specs"
	payloadURL := "https://m3.material.io/_dsm/data/dsdb-m3/TOKEN_TABLE.fixture.json"
	catalog := filepath.Join(root, "catalog.json")
	if err := os.WriteFile(catalog, []byte(`{"schemaVersion":2,"source":"fixture","scope":"component specs","generatedAt":"2026-06-01T12:00:00Z","pages":[{"url":"https://m3.material.io/components/snackbar/specs","area":"components","subject":"snackbar","pageType":"specs"}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	receipt := crawl.CrawlReceipt{SchemaVersion: 2, CrawlID: "fixture", StartedAt: "2026-09-01T00:00:00Z", CompletedAt: "2026-09-01T00:00:00Z", Options: crawl.CrawlOptions{PageCatalogPath: catalog, CacheDir: root}, Pages: []string{page}, Observations: []crawl.Observation{{PageURL: page, Status: 200, SourcePayloadURLs: []string{payloadURL}}}, SourcePayloads: []crawl.SourcePayloadRecord{{PageURLs: []string{page}, SourcePayloadURL: payloadURL, SourcePayloadKind: "component", SourceArea: "components", SourceSubject: "snackbar", SourcePageType: "specs", Hash: fmt.Sprintf("%x", sha256.Sum256(payload)), Bytes: len(payload), Artifact: "payload.json.gz"}}}
	if partial {
		receipt.Pages = append(receipt.Pages, "https://m3.material.io/components/buttons/specs")
		receipt.Observations = append(receipt.Observations, crawl.Observation{PageURL: receipt.Pages[1], Status: 404, Error: "not found"})
	}
	body, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "receipt.json")
	if err := os.WriteFile(path, body, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func testConfig(t *testing.T, width int, partial bool) Config {
	t.Helper()
	return Config{StateDir: t.TempDir(), ReceiptPath: receiptFixture(t, width, partial), MaxChangeFraction: 1, Crawl: crawl.Config{Seed: crawl.DefaultSeed, Scope: crawl.ScopeTokenSourcePages, MaxPages: 500, Concurrency: 2, PageTimeout: time.Second}}
}

func TestBaselineReportNotificationAndTamperRejection(t *testing.T) {
	config := testConfig(t, 40, false)
	var notice map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&notice); err != nil {
			t.Error(err)
		}
		w.WriteHeader(204)
	}))
	defer server.Close()
	config.Webhook = server.URL
	report, err := Run(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != "needs_review" || report.Summary.Tokens != 1 || report.CandidateDigest == "" {
		t.Fatalf("unexpected report: %+v", report)
	}
	if notice["runId"] != report.ID || notice["status"] != "needs_review" {
		t.Fatalf("notice: %v", notice)
	}
	if _, err := os.Stat(filepath.Join(config.StateDir, "current")); !os.IsNotExist(err) {
		t.Fatalf("baseline activated before approval: %v", err)
	}
	saved, err := Read(config.StateDir, report.ID)
	if err != nil || saved.CandidateDigest != report.CandidateDigest {
		t.Fatalf("saved report: %v", err)
	}
	if err := os.WriteFile(filepath.Join(report.CandidatePath, "publication.json"), []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(config.StateDir, report.ID); err == nil || !strings.Contains(err.Error(), "changed after review") {
		t.Fatalf("tampered candidate approval: %v", err)
	}
}

func TestPartialCaptureCannotBeApprovedAndRetryUsesExactReceipt(t *testing.T) {
	config := testConfig(t, 40, true)
	report, err := Run(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != "needs_review" || report.CandidatePath != "" {
		t.Fatalf("partial report: %+v", report)
	}
	if _, err := Approve(config.StateDir, report.ID); err == nil {
		t.Fatal("approved incomplete capture")
	}
	config.RetryRunID = report.ID
	retried, err := run(context.Background(), config, func(ctx context.Context, c crawl.Config) error {
		if !c.RetryFailed || c.RetryFailedFrom != report.ReceiptPath {
			t.Fatalf("retry used wrong receipt: %+v", c)
		}
		var receipt crawl.CrawlReceipt
		if err := readJSON(c.RetryFailedFrom, &receipt); err != nil {
			t.Fatal(err)
		}
		pages, err := crawl.FailedPages(receipt)
		if err != nil || len(pages) != 1 || !strings.Contains(pages[0], "buttons") {
			t.Fatalf("retry pages: %v %v", pages, err)
		}
		if c.ResultsDir == report.Crawl.ResultsDir {
			t.Fatal("retry overwrites parent's receipts")
		}
		return fmt.Errorf("simulated capture failure")
	})
	if err == nil || retried.ID == report.ID || retried.ParentRunID != report.ID || retried.Status != "failed" {
		t.Fatalf("retry: %+v %v", retried, err)
	}
	parent, err := Read(config.StateDir, report.ID)
	if err != nil || parent.Status != report.Status {
		t.Fatalf("parent mutated: %v", err)
	}
}

func TestNotificationFailureIsSavedAndResendWorks(t *testing.T) {
	config := testConfig(t, 40, false)
	code := 503
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(code) }))
	defer server.Close()
	config.Webhook = server.URL
	report, err := Run(context.Background(), config)
	if err == nil || report.Status != "needs_review" || report.NotificationError == "" {
		t.Fatalf("notification error lost: %+v %v", report, err)
	}
	code = 204
	if err := NotifyRun(context.Background(), config.StateDir, report.ID, server.URL); err != nil {
		t.Fatal(err)
	}
	report, err = Read(config.StateDir, report.ID)
	if err != nil || report.NotificationError != "" {
		t.Fatalf("resend: %+v %v", report, err)
	}
}

func TestUpdateLockAndInvalidIDs(t *testing.T) {
	root := t.TempDir()
	unlock, err := lock(root)
	if err != nil {
		t.Fatal(err)
	}
	if second, err := lock(root); err == nil {
		second()
		t.Fatal("concurrent lock accepted")
	}
	unlock()
	unlock, err = lock(root)
	if err != nil {
		t.Fatal(err)
	}
	unlock()
	for _, id := range []string{"../escape", "", "a/b", "a\\b"} {
		if _, err := Read(root, id); err == nil {
			t.Fatalf("invalid ID %q", id)
		}
	}
}

func TestUnchangedCaptureFreshnessAndLegacyApproval(t *testing.T) {
	config := testConfig(t, 40, false)
	baseline, err := Run(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(config.StateDir, baseline.ID); err != nil {
		t.Fatal(err)
	}
	config.ReceiptPath = receiptFixture(t, 48, false)
	changed, err := Run(context.Background(), config)
	if err != nil || changed.Status != "published" {
		t.Fatalf("changed capture: %+v %v", changed, err)
	}
	before, err := publish.Load(filepath.Join(changed.CandidatePath, "tokens.sqlite"))
	if err != nil || len(before.Changes) != 1 {
		t.Fatalf("changed history: %+v %v", before, err)
	}
	fresh, err := Run(context.Background(), config)
	if err != nil || fresh.Status != "published" {
		t.Fatalf("fresh capture: %+v %v", fresh, err)
	}
	after, err := publish.Load(filepath.Join(fresh.CandidatePath, "tokens.sqlite"))
	if err != nil || after.CurrentSnapshot.ID != fresh.ID || after.CurrentSnapshot.CreatedAt != fresh.StartedAt || len(after.Changes) != 1 {
		t.Fatalf("fresh capture metadata/history: %+v %v", after, err)
	}
	oldHistory, _ := json.Marshal(before.Changes)
	newHistory, _ := json.Marshal(after.Changes)
	if string(oldHistory) != string(newHistory) {
		t.Fatalf("unchanged capture rewrote detected changes: %s -> %s", oldHistory, newHistory)
	}
	// Recreate the saved candidate and current link produced by older releases.
	if err := publish.Activate(filepath.Join(config.StateDir, "current"), changed.CandidatePath); err != nil {
		t.Fatal(err)
	}
	fresh.Status = "unchanged"
	fresh.Phase = "review"
	if err := save(config.StateDir, fresh); err != nil {
		t.Fatal(err)
	}
	approved, err := Approve(config.StateDir, fresh.ID)
	if err != nil || approved.Status != "published" {
		t.Fatalf("legacy unchanged approval: %+v %v", approved, err)
	}
	current, err := active(config.StateDir)
	if err != nil || current != fresh.CandidatePath {
		t.Fatalf("legacy unchanged candidate not activated: %s %v", current, err)
	}
}

func TestApprovalHistoryUnchangedAndStaleCandidate(t *testing.T) {
	config := testConfig(t, 40, false)
	probe := filepath.Join(config.StateDir, "probe")
	if err := os.Symlink(config.StateDir, probe); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	first, err := Run(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(config.StateDir, first.ID); err != nil {
		t.Fatal(err)
	}
	unchanged, err := Run(context.Background(), config)
	if err != nil || unchanged.Status != "published" || unchanged.Summary.Added+unchanged.Summary.Updated+unchanged.Summary.Removed != 0 {
		t.Fatalf("unchanged: %+v %v", unchanged, err)
	}
	current, err := active(config.StateDir)
	if err != nil || current != unchanged.CandidatePath {
		t.Fatalf("unchanged capture did not become current: %s %v", current, err)
	}
	refreshed, err := publish.Load(filepath.Join(current, "tokens.sqlite"))
	if err != nil || refreshed.CurrentSnapshot.ID != unchanged.ID || len(refreshed.Changes) != 0 {
		t.Fatalf("freshness or history incorrect after unchanged capture: %+v %v", refreshed, err)
	}
	config.ReceiptPath = receiptFixture(t, 48, false)
	second, err := Run(context.Background(), config)
	if err != nil || second.Status != "published" || second.Summary.Updated != 1 {
		t.Fatalf("update: %+v %v", second, err)
	}
	state, err := publish.Load(filepath.Join(second.CandidatePath, "tokens.sqlite"))
	if err != nil || len(state.Snapshots) != 3 || len(state.Changes) != 1 {
		t.Fatalf("timeline: %+v %v", state, err)
	}
	config.MaxChangeFraction = .25
	config.ReceiptPath = receiptFixture(t, 56, false)
	held, err := Run(context.Background(), config)
	if err != nil || held.Status != "needs_review" {
		t.Fatalf("held: %+v %v", held, err)
	}
	config.MaxChangeFraction = 1
	config.ReceiptPath = receiptFixture(t, 64, false)
	latest, err := Run(context.Background(), config)
	if err != nil || latest.Status != "published" {
		t.Fatalf("latest: %+v %v", latest, err)
	}
	if _, err := Approve(config.StateDir, held.ID); err == nil || !strings.Contains(err.Error(), "live publication changed") {
		t.Fatalf("stale approval: %v", err)
	}
	// Recover a crash after the link switch but before the final report write.
	latest.Status = "running"
	if err := save(config.StateDir, latest); err != nil {
		t.Fatal(err)
	}
	if _, err := Run(context.Background(), config); err == nil || !strings.Contains(err.Error(), "recover interrupted activation") {
		t.Fatalf("lost interrupted activation: %v", err)
	}
	if recovered, err := Approve(config.StateDir, latest.ID); err != nil || recovered.Status != "published" {
		t.Fatalf("recovery: %+v %v", recovered, err)
	}
}
