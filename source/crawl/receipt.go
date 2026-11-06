// Receipts preserve both successful and failed observations. Retrying replaces only the selected pages, not unrelated evidence.
package crawl

import (
	"encoding/json"

	"errors"
	"fmt"

	"net/http"

	"os"
	"path/filepath"

	"sort"
)

type Observation struct {
	PageURL           string   `json:"pageUrl"`
	Status            int      `json:"status"`
	SourcePayloadURLs []string `json:"sourcePayloadUrls,omitempty"`
	Title             string   `json:"title,omitempty"`
	Error             string   `json:"error,omitempty"`
}

type SourcePayloadRecord struct {
	PageURLs          []string `json:"pageUrls"`
	SourcePayloadURL  string   `json:"sourcePayloadUrl"`
	SourcePayloadKind string   `json:"sourcePayloadKind,omitempty"`
	SourceArea        string   `json:"sourceArea,omitempty"`
	SourceSubject     string   `json:"sourceSubject,omitempty"`
	SourcePageType    string   `json:"sourcePageType,omitempty"`
	Hash              string   `json:"hash"`
	Artifact          string   `json:"artifact"`
	Bytes             int      `json:"bytes"`
	RetrievedAt       string   `json:"retrievedAt"`
	ETag              string   `json:"etag,omitempty"`
	LastModified      string   `json:"lastModified,omitempty"`
	Reused            bool     `json:"reused"`
	Error             string   `json:"error,omitempty"`
}

type CacheIndex struct {
	SchemaVersion  int                            `json:"schemaVersion"`
	SourcePayloads map[string]SourcePayloadRecord `json:"sourcePayloads"`
}

type CrawlReceipt struct {
	SchemaVersion  int                   `json:"schemaVersion"`
	CrawlID        string                `json:"crawlId"`
	StartedAt      string                `json:"startedAt"`
	CompletedAt    string                `json:"completedAt"`
	Mode           string                `json:"mode"`
	Options        CrawlOptions          `json:"options"`
	Discovery      DiscoveryReport       `json:"discovery"`
	Capture        CaptureReport         `json:"capture"`
	Pages          []string              `json:"pages"`
	Observations   []Observation         `json:"observations"`
	SourcePayloads []SourcePayloadRecord `json:"sourcePayloads"`
}

type CaptureReport struct {
	PagesAttempted         int  `json:"pagesAttempted"`
	PagesSucceeded         int  `json:"pagesSucceeded"`
	PagesFailed            int  `json:"pagesFailed"`
	PagesCarriedForward    int  `json:"pagesCarriedForward,omitempty"`
	PayloadsObserved       int  `json:"payloadsObserved"`
	PayloadsFetched        int  `json:"payloadsFetched"`
	PayloadsCarriedForward int  `json:"payloadsCarriedForward,omitempty"`
	Complete               bool `json:"complete"`
}

func mode(retryFailed bool) string {
	if retryFailed {
		return "retry-failed"
	}
	return "all"
}

func retryReceiptPath(config Config) (string, error) {
	if config.RetryFailedFrom != "" {
		return config.RetryFailedFrom, nil
	}
	resultsDir := config.ResultsDir
	if resultsDir == "" {
		resultsDir = "crawl-results"
	}
	entries, err := os.ReadDir(resultsDir)
	if err != nil {
		return "", fmt.Errorf("read crawl-receipt directory: %w", err)
	}
	paths := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			paths = append(paths, filepath.Join(resultsDir, entry.Name()))
		}
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		return "", fmt.Errorf("no crawl receipts found in %s", resultsDir)
	}
	return paths[len(paths)-1], nil
}

func readCrawlReceipt(path string) (CrawlReceipt, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return CrawlReceipt{}, fmt.Errorf("read crawl receipt: %w", err)
	}
	var receipt CrawlReceipt
	if err := json.Unmarshal(body, &receipt); err != nil {
		return CrawlReceipt{}, fmt.Errorf("parse crawl receipt: %w", err)
	}
	if receipt.SchemaVersion != 2 {
		return CrawlReceipt{}, fmt.Errorf("unsupported crawl receipt schema version %d", receipt.SchemaVersion)
	}
	if receipt.Options.DryRun {
		return CrawlReceipt{}, errors.New("cannot retry failed pages from a dry-run crawl receipt")
	}
	return receipt, nil
}

func failedPages(receipt CrawlReceipt) ([]string, error) {
	observations := make(map[string]Observation, len(receipt.Observations))
	successfulPayloads := make(map[string]bool)
	for _, observation := range receipt.Observations {
		observations[observation.PageURL] = observation
	}
	for _, payload := range receipt.SourcePayloads {
		if payload.SourcePayloadURL != "" && payload.Error == "" && payload.Hash != "" {
			successfulPayloads[payload.SourcePayloadURL] = true
		}
	}
	failed := make([]string, 0)
	seen := make(map[string]bool)
	for _, pageURL := range receipt.Pages {
		if seen[pageURL] {
			continue
		}
		seen[pageURL] = true
		observation, observed := observations[pageURL]
		pageFailed := !observed || observation.Error != "" || observation.Status < http.StatusOK || observation.Status >= http.StatusBadRequest || len(observation.SourcePayloadURLs) == 0
		if !pageFailed {
			for _, payloadURL := range observation.SourcePayloadURLs {
				if !successfulPayloads[payloadURL] {
					pageFailed = true
					break
				}
			}
		}
		if pageFailed {
			failed = append(failed, pageURL)
		}
	}
	if len(receipt.Pages) == 0 {
		return nil, errors.New("crawl receipt contains no pages")
	}
	return failed, nil
}

// FailedPages returns the exact pages eligible for a targeted retry.
func FailedPages(receipt CrawlReceipt) ([]string, error) { return failedPages(receipt) }

func mergeObservations(previous, retried []Observation, retriedPages []string) []Observation {
	retriedSet := make(map[string]bool, len(retriedPages))
	for _, pageURL := range retriedPages {
		retriedSet[pageURL] = true
	}
	updated := make(map[string]Observation, len(retried))
	for _, observation := range retried {
		updated[observation.PageURL] = observation
	}
	merged := make([]Observation, 0, len(previous)+len(retried))
	seen := make(map[string]bool)
	for _, observation := range previous {
		if replacement, ok := updated[observation.PageURL]; retriedSet[observation.PageURL] && ok {
			merged = append(merged, replacement)
		} else {
			merged = append(merged, observation)
		}
		seen[observation.PageURL] = true
	}
	for _, observation := range retried {
		if !seen[observation.PageURL] {
			merged = append(merged, observation)
			seen[observation.PageURL] = true
		}
	}
	return merged
}

func mergeSourcePayloads(previous, retried []SourcePayloadRecord, retriedPages []string) []SourcePayloadRecord {
	retriedSet := make(map[string]bool, len(retriedPages))
	for _, pageURL := range retriedPages {
		retriedSet[pageURL] = true
	}
	byURL := make(map[string]SourcePayloadRecord, len(previous)+len(retried))
	order := make([]string, 0, len(previous)+len(retried))
	for _, payload := range previous {
		if payload.SourcePayloadURL == "" {
			continue
		}
		filteredPages := make([]string, 0, len(payload.PageURLs))
		for _, pageURL := range payload.PageURLs {
			if !retriedSet[pageURL] {
				filteredPages = append(filteredPages, pageURL)
			}
		}
		payload.PageURLs = filteredPages
		if len(filteredPages) > 0 {
			byURL[payload.SourcePayloadURL] = payload
			order = append(order, payload.SourcePayloadURL)
		}
	}
	for _, payload := range retried {
		if payload.SourcePayloadURL == "" {
			continue
		}
		if existing, ok := byURL[payload.SourcePayloadURL]; ok {
			existing.PageURLs = mergeStrings(existing.PageURLs, payload.PageURLs)
			if payload.Error == "" || existing.Error != "" {
				payload.PageURLs = existing.PageURLs
				byURL[payload.SourcePayloadURL] = payload
			} else {
				byURL[payload.SourcePayloadURL] = existing
			}
			continue
		}
		byURL[payload.SourcePayloadURL] = payload
		order = append(order, payload.SourcePayloadURL)
	}
	merged := make([]SourcePayloadRecord, 0, len(byURL))
	for _, payloadURL := range order {
		payload := byURL[payloadURL]
		if len(payload.PageURLs) == 0 {
			continue
		}
		sort.Strings(payload.PageURLs)
		merged = append(merged, payload)
	}
	return merged
}

func successfulPageCount(observations []Observation) int {
	count := 0
	for _, observation := range observations {
		if observation.Status >= http.StatusOK && observation.Status < http.StatusBadRequest && observation.Error == "" {
			count++
		}
	}
	return count
}

func captureReport(pages []string, observations []Observation, payloads []SourcePayloadRecord, dryRun bool) CaptureReport {
	fetched := successfulPayloadCount(payloads)
	observed := len(uniqueSources(observations))
	return CaptureReport{
		PagesAttempted:   len(pages),
		PagesSucceeded:   successfulPageCount(observations),
		PagesFailed:      len(pages) - successfulPageCount(observations),
		PayloadsObserved: observed,
		PayloadsFetched:  fetched,
		Complete:         !dryRun && captureComplete(pages, observations, payloads),
	}
}

func captureComplete(pages []string, observations []Observation, payloads []SourcePayloadRecord) bool {
	fetched := successfulPayloadCount(payloads)
	observed := len(uniqueSources(observations))
	return len(pages) > 0 && successfulPageCount(observations) == len(pages) && observed > 0 && fetched == observed
}

func successfulPayloadCount(payloads []SourcePayloadRecord) int {
	seen := make(map[string]bool, len(payloads))
	count := 0
	for _, payload := range payloads {
		if payload.SourcePayloadURL != "" && payload.Hash != "" && payload.Error == "" && !seen[payload.SourcePayloadURL] {
			seen[payload.SourcePayloadURL] = true
			count++
		}
	}
	return count
}

func mergeStrings(left, right []string) []string {
	seen := make(map[string]bool, len(left)+len(right))
	merged := make([]string, 0, len(left)+len(right))
	for _, value := range append(append([]string(nil), left...), right...) {
		if value != "" && !seen[value] {
			seen[value] = true
			merged = append(merged, value)
		}
	}
	sort.Strings(merged)
	return merged
}
