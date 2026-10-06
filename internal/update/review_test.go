package update

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/sojupie/tokens/internal/crawl"
	"github.com/sojupie/tokens/internal/publish"
)

func TestReviewPolicies(t *testing.T) {
	config := testConfig(t, 40, false)
	report, err := Run(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	var baseline publish.Publication
	if err := readJSON(filepath.Join(report.CandidatePath, "publication.json"), &baseline); err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct {
		name, code string
		mutate     func(*publish.Publication)
	}{
		{"unchanged", "", func(p *publish.Publication) {}},
		{"empty", "empty_publication", func(p *publish.Publication) { p.Subjects = nil }},
		{"subject removed", "subject_removed", func(p *publish.Publication) { p.Subjects = nil }},
		{"token count", "token_count_drop", func(p *publish.Publication) { p.Subjects[0].Tokens = nil }},
		{"unresolved", "unresolved_values", func(p *publish.Publication) {
			p.Subjects[0].Tokens[0].Values[0].Value = json.RawMessage(`{"undefined": true}`)
		}},
		{"large", "large_change", func(p *publish.Publication) {
			p.Changes = []publish.ChangeView{{SnapshotID: p.CurrentSnapshot.ID, Type: "changed"}}
		}},
		{"removed value", "values_removed", func(p *publish.Publication) {
			p.Changes = []publish.ChangeView{{SnapshotID: p.CurrentSnapshot.ID, Type: "removed"}}
		}},
	} {
		t.Run(item.name, func(t *testing.T) {
			body, err := json.Marshal(baseline)
			if err != nil {
				t.Fatal(err)
			}
			var candidate publish.Publication
			if err := json.Unmarshal(body, &candidate); err != nil {
				t.Fatal(err)
			}
			item.mutate(&candidate)
			_, issues, err := review(candidate, report.CandidatePath, .25)
			if err != nil {
				t.Fatal(err)
			}
			if item.code == "" {
				if len(issues) != 0 {
					t.Fatalf("unexpected issues: %v", issues)
				}
				return
			}
			for _, issue := range issues {
				if issue.Code == item.code {
					return
				}
			}
			t.Fatalf("missing %s: %v", item.code, issues)
		})
	}
}

func TestUnattemptedCatalogPageBlocksPublication(t *testing.T) {
	config := testConfig(t, 40, false)
	var receipt crawl.CrawlReceipt
	if err := readJSON(config.ReceiptPath, &receipt); err != nil {
		t.Fatal(err)
	}
	var catalog crawl.PageCatalog
	if err := readJSON(receipt.Options.PageCatalogPath, &catalog); err != nil {
		t.Fatal(err)
	}
	catalog.Pages = append(catalog.Pages, crawl.PageCatalogEntry{URL: "https://m3.material.io/components/buttons/specs", Area: "components", Subject: "buttons", PageType: "specs"})
	body, err := json.Marshal(catalog)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(receipt.Options.PageCatalogPath, body, 0600); err != nil {
		t.Fatal(err)
	}
	report, err := Run(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	for _, issue := range report.Issues {
		if issue.Code == "missing_required_page" && issue.Blocking && report.CandidatePath == "" {
			return
		}
	}
	t.Fatalf("truncated capture not blocked: %+v", report)
}

func TestAlertRecordsInterruptedRun(t *testing.T) {
	root := t.TempDir()
	report := newReport(Config{MaxChangeFraction: 1})
	if err := save(root, report); err != nil {
		t.Fatal(err)
	}
	notice := make(chan map[string]any, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		notice <- body
		w.WriteHeader(204)
	}))
	defer server.Close()
	if err := Alert(context.Background(), root, server.URL); err != nil {
		t.Fatal(err)
	}
	saved, err := Read(root, report.ID)
	if err != nil || saved.Status != "failed" || saved.Error == "" {
		t.Fatalf("interrupted run: %+v %v", saved, err)
	}
	if body := <-notice; body["runId"] != report.ID || body["status"] != "failed" {
		t.Fatalf("notice: %v", body)
	}
}
