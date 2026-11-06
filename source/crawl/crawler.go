// Run coordinates discovery, browser observation and payload persistence; capture evidence is recorded even when a page fails.
package crawl

import (
	"context"

	"errors"
	"fmt"

	"path/filepath"

	"time"
)

func Run(ctx context.Context, config Config) error {
	if config.RetryFailedFrom != "" {
		config.RetryFailed = true
	}
	if config.Scope == "" {
		config.Scope = ScopeComponentSpecs
	}
	if config.PathPrefix == "" {
		config.PathPrefix = "/"
	}
	if err := config.validate(); err != nil {
		return err
	}
	if config.PageCatalogPath == "" {
		config.PageCatalogPath = DefaultPageCatalogPath
	}
	started := time.Now().UTC()
	var pages []string
	var pagesToObserve []string
	var discovery DiscoveryReport
	var previous *CrawlReceipt
	if config.RetryFailed {
		receiptPath, err := retryReceiptPath(config)
		if err != nil {
			return err
		}
		receipt, err := readCrawlReceipt(receiptPath)
		if err != nil {
			return err
		}
		failed, err := failedPages(receipt)
		if err != nil {
			return err
		}
		pages = append([]string(nil), receipt.Pages...)
		pagesToObserve = limitPages(failed, config.MaxPages)
		if len(pagesToObserve) == 0 {
			return fmt.Errorf("no failed pages found in crawl receipt %s", receiptPath)
		}
		previous = &receipt
		discovery = DiscoveryReport{Mode: "retry-failed", PageCatalogPath: config.PageCatalogPath, Source: receiptPath}
		fmt.Printf("Retrying %d failed pages; retaining %d total pages.\n", len(pagesToObserve), len(pages))
	} else {
		var err error
		fmt.Printf("Selecting pages (scope %s)...\n", config.Scope)
		pages, discovery, err = discoverPages(ctx, config)
		if err != nil {
			return fmt.Errorf("discover pages: %w", err)
		}
		pagesToObserve = pages
		fmt.Printf("Discovered %d candidate pages.\n", len(pages))
		fmt.Printf("Page source: %s\n", discovery.Mode)
		if len(discovery.Added) > 0 || len(discovery.Removed) > 0 {
			fmt.Printf("Page-catalog drift: %d added, %d removed.\n", len(discovery.Added), len(discovery.Removed))
		}
	}

	observations, err := observePages(ctx, config, pagesToObserve)
	if err != nil {
		return fmt.Errorf("observe pages: %w", err)
	}
	if successfulPageCount(observations) == 0 {
		for _, observation := range observations {
			if observation.Error != "" {
				return fmt.Errorf("capture produced no successful pages: %s", observation.Error)
			}
		}
		return errors.New("capture produced no successful pages")
	}
	sources := uniqueSources(observations)
	fmt.Printf("Discovered %d unique source payloads.\n", len(sources))

	indexPath := filepath.Join(config.CacheDir, "index.json")
	index, err := loadIndex(indexPath)
	if err != nil {
		return err
	}
	sourcePayloads := make([]SourcePayloadRecord, 0, len(sources))
	if !config.DryRun {
		fmt.Printf("Persisting %d unique source payloads...\n", len(sources))
		for payloadNumber, source := range sources {
			started := time.Now()
			fmt.Printf("  [payload %d/%d] processing %s\n", payloadNumber+1, len(sources), source.SourcePayloadURL)
			record, err := persistPayload(ctx, config, index, source)
			if err != nil {
				record = SourcePayloadRecord{
					PageURLs:          append([]string(nil), source.PageURLs...),
					SourcePayloadURL:  source.SourcePayloadURL,
					SourcePayloadKind: source.SourcePayloadKind,
					SourceArea:        source.SourceArea,
					SourceSubject:     source.SourceSubject,
					SourcePageType:    source.SourcePageType,
					Error:             err.Error(),
				}
				fmt.Printf("  [payload %d/%d] failed after %s: %s\n", payloadNumber+1, len(sources), elapsed(time.Since(started)), err)
			} else if record.Reused {
				fmt.Printf("  [payload %d/%d] reused after %s\n", payloadNumber+1, len(sources), elapsed(time.Since(started)))
			} else {
				fmt.Printf("  [payload %d/%d] stored %d bytes after %s\n", payloadNumber+1, len(sources), record.Bytes, elapsed(time.Since(started)))
			}
			sourcePayloads = append(sourcePayloads, record)
		}
		if err := writeJSONAtomic(indexPath, index); err != nil {
			return fmt.Errorf("write cache index: %w", err)
		}
	} else {
		fmt.Printf("Dry run: skipping persistence of %d source payloads.\n", len(sources))
	}
	var captureReportValue CaptureReport
	if previous != nil {
		attemptedObservations := observations
		attemptedSourcePayloads := sourcePayloads
		observations = mergeObservations(previous.Observations, observations, pagesToObserve)
		sourcePayloads = mergeSourcePayloads(previous.SourcePayloads, sourcePayloads, pagesToObserve)
		capture := captureReport(pagesToObserve, attemptedObservations, attemptedSourcePayloads, config.DryRun)
		capture.PagesCarriedForward = len(pages) - len(pagesToObserve)
		capture.PayloadsCarriedForward = len(sourcePayloads) - len(attemptedSourcePayloads)
		capture.Complete = captureComplete(pages, observations, sourcePayloads) && !config.DryRun
		captureReportValue = capture
	}

	crawlID := started.Format("20060102T150405.000000000Z")
	if previous == nil {
		captureReportValue = captureReport(pages, observations, sourcePayloads, config.DryRun)
	}
	receipt := CrawlReceipt{
		SchemaVersion:  2,
		CrawlID:        crawlID,
		StartedAt:      started.Format(time.RFC3339Nano),
		CompletedAt:    time.Now().UTC().Format(time.RFC3339Nano),
		Mode:           mode(config.RetryFailed),
		Options:        crawlOptions(config),
		Discovery:      discovery,
		Capture:        captureReportValue,
		Pages:          pages,
		Observations:   observations,
		SourcePayloads: sourcePayloads,
	}
	receiptPath := filepath.Join(config.ResultsDir, crawlID+".json")
	if err := writeJSONAtomic(receiptPath, receipt); err != nil {
		return fmt.Errorf("write crawl receipt: %w", err)
	}
	fmt.Printf("Wrote crawl receipt to %s\n", receiptPath)
	return nil
}
