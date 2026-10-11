package update

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sojupie/tokens/internal/crawl"
	"github.com/sojupie/tokens/internal/publish"
	"github.com/sojupie/tokens/internal/tokensnapshot"
)

type Config struct {
	StateDir          string
	HistoryDir        string
	Crawl             crawl.Config
	ReceiptPath       string
	RetryRunID        string
	Webhook           string
	MaxChangeFraction float64
	Progress          func(string)
}

// Run saves evidence before promotion. Failed/review runs never replace current.
func Run(ctx context.Context, config Config) (report Report, resultErr error) {
	return run(ctx, config, crawl.Run)
}

func run(ctx context.Context, config Config, capture func(context.Context, crawl.Config) error) (report Report, resultErr error) {
	root, err := stateRoot(config.StateDir)
	if err != nil {
		return Report{}, err
	}
	if config.MaxChangeFraction <= 0 || config.MaxChangeFraction > 1 {
		return Report{}, errors.New("max change fraction must be greater than zero and at most one")
	}
	unlock, err := lock(root)
	if err != nil {
		return Report{}, err
	}
	defer unlock()
	if config.RetryRunID != "" {
		parent, err := Read(root, config.RetryRunID)
		if err != nil {
			return Report{}, err
		}
		if parent.Status != "failed" && parent.Status != "needs_review" && parent.Status != "running" {
			return Report{}, errors.New("only failed, interrupted or held runs can be retried")
		}
		config.Crawl = parent.Crawl
		config.HistoryDir = parent.HistoryDir
		if parent.ReceiptPath != "" {
			var receipt crawl.CrawlReceipt
			if err := readJSON(parent.ReceiptPath, &receipt); err != nil {
				return Report{}, err
			}
			pages, err := crawl.FailedPages(receipt)
			if err != nil {
				return Report{}, err
			}
			if len(pages) == 0 {
				return Report{}, errors.New("this capture has no failed pages; approve the candidate or start a fresh run")
			}
		}
		config.Crawl.PageTimeout, err = time.ParseDuration(parent.PageTimeout)
		if err != nil {
			return Report{}, err
		}
		config.MaxChangeFraction = parent.MaxChangeFraction
		config.Crawl.RetryFailed = parent.ReceiptPath != ""
		config.Crawl.RetryFailedFrom = parent.ReceiptPath
		config.Crawl.Discover = false
		config.Crawl.NoPageCatalog = false
		config.Crawl.UpdatePageCatalog = false
		config.ReceiptPath = ""
	}
	report = newReport(config)
	report.ParentRunID = config.RetryRunID
	report.BaseGeneration, err = active(root)
	if err != nil {
		return report, err
	}
	if report.BaseGeneration != "" {
		reports, err := List(root)
		if err != nil {
			return report, err
		}
		known := false
		for _, previous := range reports {
			if previous.CandidatePath == report.BaseGeneration {
				if config.HistoryDir == "" {
					config.HistoryDir = previous.HistoryDir
				}
				known = true
				if previous.Status != "published" {
					return report, fmt.Errorf("recover interrupted activation first: tokens update approve %s", previous.ID)
				}
			}
		}
		if !known {
			return report, errors.New("current generation is not an accepted update run")
		}
	}
	if config.HistoryDir != "" {
		config.HistoryDir, err = filepath.Abs(config.HistoryDir)
		if err != nil {
			return report, err
		}
	}
	report.HistoryDir = config.HistoryDir
	directory, _ := runDirectory(root, report.ID)
	config.Crawl.ResultsDir = filepath.Join(directory, "receipts")
	if config.Crawl.CacheDir == "" {
		config.Crawl.CacheDir = filepath.Join(root, "cache", "payloads")
	}
	report.Crawl = config.Crawl
	if err := save(root, report); err != nil {
		return report, err
	}
	defer func() {
		if resultErr != nil {
			report.Status = "failed"
			report.Error = resultErr.Error()
		}
		report.FinishedAt = time.Now().UTC().Format(time.RFC3339Nano)
		if err := save(root, report); err != nil {
			resultErr = errors.Join(resultErr, err)
			return
		}
		if report.Status == "failed" || report.Status == "needs_review" {
			if err := Notify(ctx, root, report.ID, config.Webhook); err != nil {
				resultErr = errors.Join(resultErr, err)
			}
			if saved, err := Read(root, report.ID); err == nil {
				report = saved
			}
		}
	}()
	if config.ReceiptPath == "" {
		if err := capture(ctx, config.Crawl); err != nil {
			return report, err
		}
		entries, err := os.ReadDir(config.Crawl.ResultsDir)
		if err != nil {
			return report, err
		}
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
				continue
			}
			if config.ReceiptPath != "" {
				return report, errors.New("capture wrote more than one receipt")
			}
			config.ReceiptPath = filepath.Join(config.Crawl.ResultsDir, entry.Name())
		}
		if config.ReceiptPath == "" {
			return report, errors.New("capture did not write a receipt")
		}
	}
	// Copy imported receipts so external edits cannot change retry evidence.
	body, err := os.ReadFile(config.ReceiptPath)
	if err != nil {
		return report, err
	}
	report.ReceiptPath = filepath.Join(directory, "receipt.json")
	if err := os.WriteFile(report.ReceiptPath, body, 0600); err != nil {
		return report, err
	}
	var receipt crawl.CrawlReceipt
	if err := readJSON(report.ReceiptPath, &receipt); err != nil {
		return report, err
	}
	report.Capture = receipt.Capture
	report.Crawl.CacheDir = receipt.Options.CacheDir
	report.Crawl.PageCatalogPath = receipt.Options.PageCatalogPath
	if receipt.SchemaVersion != 2 || len(receipt.Pages) == 0 {
		return report, errors.New("invalid or empty crawl receipt")
	}
	if receipt.Options.DryRun {
		return report, errors.New("a dry crawl cannot become a publication")
	}
	report.Phase = "snapshot"
	if err := save(root, report); err != nil {
		return report, err
	}
	report.SnapshotPath = filepath.Join(directory, "snapshot")
	manifest, err := tokensnapshot.Build(tokensnapshot.BuildConfig{CrawlReceiptPath: report.ReceiptPath, OutputDir: report.SnapshotPath, SnapshotID: report.ID, CreatedAt: report.StartedAt, ExtractorName: "tokens", ExtractorVersion: "0.1.0"})
	if err != nil {
		return report, err
	}
	if !manifest.Capture.Complete {
		report.Issues = append(report.Issues, Issue{"incomplete_capture", "Some pages or payloads failed. Retry this run before publishing.", true})
	}
	report.Capture.Complete = manifest.Capture.Complete
	report.Capture.PagesAttempted = manifest.Capture.PagesAttempted
	report.Capture.PagesSucceeded = manifest.Capture.PagesSucceeded
	report.Capture.PagesFailed = manifest.Capture.PagesAttempted - manifest.Capture.PagesSucceeded
	var catalog crawl.PageCatalog
	if err := readJSON(filepath.Join(report.SnapshotPath, "page-catalog.json"), &catalog); err != nil {
		return report, err
	}
	capturedPages := make(map[string]bool, len(receipt.Pages))
	for _, page := range receipt.Pages {
		capturedPages[page] = true
	}
	for _, page := range catalog.Pages {
		if !capturedPages[page.URL] {
			report.Issues = append(report.Issues, Issue{"missing_required_page", "Required catalog page was not attempted: " + page.URL, true})
		}
	}
	var pages tokensnapshot.PageResults
	if err := readJSON(filepath.Join(report.SnapshotPath, "page-results.json"), &pages); err != nil {
		return report, err
	}
	for _, page := range pages.Pages {
		if page.ExtractionError != "" || page.PayloadError != "" || page.RecordsExtracted == 0 {
			report.Issues = append(report.Issues, Issue{"invalid_page", page.URL + ": missing records or extraction failure", true})
		}
	}
	for _, issue := range report.Issues {
		if issue.Blocking {
			report.Status = "needs_review"
			return report, nil
		}
	}
	report.SnapshotDigest, err = publish.GenerationDigest(report.SnapshotPath)
	if err != nil {
		return report, err
	}
	report.Phase = "publication"
	if err := ctx.Err(); err != nil {
		return report, err
	}
	if err := save(root, report); err != nil {
		return report, err
	}
	reports, err := List(root)
	if err != nil {
		return report, err
	}
	history := make([]string, 0)
	if config.HistoryDir != "" {
		entries, err := os.ReadDir(config.HistoryDir)
		if err != nil {
			return report, err
		}
		for _, entry := range entries {
			if entry.IsDir() {
				history = append(history, filepath.Join(config.HistoryDir, entry.Name()))
			}
		}
	}
	for _, previous := range reports {
		if previous.Status == "published" {
			history = append(history, previous.SnapshotPath)
		}
	}
	report.CandidatePath = filepath.Join(directory, "publication")
	publication, err := publish.Build(publish.BuildConfig{SnapshotDir: report.SnapshotPath, HistoryPaths: history, HistoryCacheDir: filepath.Join(root, "cache", "history"), OutputDir: report.CandidatePath, Progress: config.Progress})
	if err != nil {
		return report, err
	}
	report.Summary, report.Issues, err = review(publication, report.BaseGeneration, config.MaxChangeFraction)
	if err != nil {
		return report, err
	}
	for _, previous := range reports {
		if previous.CandidatePath != report.BaseGeneration || report.BaseGeneration == "" {
			continue
		}
		var baseline tokensnapshot.Manifest
		if err := readJSON(filepath.Join(previous.SnapshotPath, "manifest.json"), &baseline); err != nil {
			return report, err
		}
		if baseline.Source.PageCatalogSha256 != manifest.Source.PageCatalogSha256 {
			report.Issues = append(report.Issues, Issue{"catalog_changed", "The reviewed page catalog changed since the live capture.", false})
		}
	}
	report.CandidateDigest, err = publish.GenerationDigest(report.CandidatePath)
	if err != nil {
		return report, err
	}
	report.Phase = "review"
	if err := ctx.Err(); err != nil {
		return report, err
	}
	if len(report.Issues) > 0 {
		report.Status = "needs_review"
		return report, nil
	}
	// A complete capture refreshes freshness metadata even without token changes.
	report.Status = "needs_review"
	if err := save(root, report); err != nil {
		return report, err
	}
	if err := activate(root, &report); err != nil {
		return report, fmt.Errorf("activate: %w", err)
	}
	return report, nil
}
