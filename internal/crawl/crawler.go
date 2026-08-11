package crawl

import (
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"golang.org/x/net/html"
)

const (
	DefaultSeed            = "https://m3.material.io/components/buttons/specs"
	DefaultPageCatalogPath = "inputs/page-catalog.json"
	tokenTablePattern      = `(?i)/TOKEN_TABLE(?:\.[^/]+)?\.json(?:\?.*)?$`
)

var tokenTableRE = regexp.MustCompile(tokenTablePattern)

type Config struct {
	Seed              string        `json:"seed"`
	Sitemap           string        `json:"sitemap,omitempty"`
	PathPrefix        string        `json:"pathPrefix"`
	SpecsOnly         bool          `json:"specsOnly"`
	PageCatalogPath   string        `json:"pageCatalogPath"`
	Discover          bool          `json:"discover"`
	UpdatePageCatalog bool          `json:"updatePageCatalog"`
	NoPageCatalog     bool          `json:"noPageCatalog"`
	MaxPages          int           `json:"maxPages"`
	Concurrency       int           `json:"concurrency"`
	PageTimeout       time.Duration `json:"-"`
	CacheDir          string        `json:"cacheDir"`
	ResultsDir        string        `json:"resultsDir"`
	Refresh           bool          `json:"refresh"`
	DryRun            bool          `json:"dryRun"`
	RetryFailed       bool          `json:"retryFailed"`
	RetryFailedFrom   string        `json:"retryFailedFrom,omitempty"`
	Headed            bool          `json:"headed"`
	ChromePath        string        `json:"chromePath,omitempty"`
}

type CrawlOptions struct {
	Seed              string `json:"seed"`
	Sitemap           string `json:"sitemap,omitempty"`
	PathPrefix        string `json:"pathPrefix"`
	SpecsOnly         bool   `json:"specsOnly"`
	PageCatalogPath   string `json:"pageCatalogPath"`
	Discover          bool   `json:"discover"`
	UpdatePageCatalog bool   `json:"updatePageCatalog"`
	NoPageCatalog     bool   `json:"noPageCatalog"`
	MaxPages          int    `json:"maxPages"`
	Concurrency       int    `json:"concurrency"`
	PageTimeout       string `json:"pageTimeout"`
	CacheDir          string `json:"cacheDir"`
	ResultsDir        string `json:"resultsDir"`
	Refresh           bool   `json:"refresh"`
	DryRun            bool   `json:"dryRun"`
	RetryFailed       bool   `json:"retryFailed"`
	RetryFailedFrom   string `json:"retryFailedFrom,omitempty"`
	Headed            bool   `json:"headed"`
	ChromePath        string `json:"chromePath,omitempty"`
}

type Observation struct {
	PageURL       string `json:"pageUrl"`
	Status        int    `json:"status"`
	TokenTableURL string `json:"tokenTableUrl,omitempty"`
	Title         string `json:"title,omitempty"`
	Error         string `json:"error,omitempty"`
}

type SourcePayloadRecord struct {
	PageURL       string `json:"pageUrl"`
	TokenTableURL string `json:"tokenTableUrl"`
	Hash          string `json:"hash"`
	Artifact      string `json:"artifact"`
	Bytes         int    `json:"bytes"`
	RetrievedAt   string `json:"retrievedAt"`
	ETag          string `json:"etag,omitempty"`
	LastModified  string `json:"lastModified,omitempty"`
	Reused        bool   `json:"reused"`
	Error         string `json:"error,omitempty"`
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

type PageCatalogEntry struct {
	URL       string `json:"url"`
	Component string `json:"component"`
}

type PageCatalog struct {
	SchemaVersion int                `json:"schemaVersion"`
	Source        string             `json:"source"`
	Scope         string             `json:"scope"`
	GeneratedAt   string             `json:"generatedAt"`
	Pages         []PageCatalogEntry `json:"pages"`
}

type DiscoveryReport struct {
	Mode            string   `json:"mode"`
	PageCatalogPath string   `json:"pageCatalogPath,omitempty"`
	Source          string   `json:"source,omitempty"`
	Added           []string `json:"added,omitempty"`
	Removed         []string `json:"removed,omitempty"`
}

type robotsFile struct {
	Sitemaps []string
}

type sitemapDocument struct {
	Sitemaps []struct {
		Loc string `xml:"loc"`
	} `xml:"sitemap"`
	URLs []struct {
		Loc string `xml:"loc"`
	} `xml:"url"`
}

func Run(ctx context.Context, config Config) error {
	if config.RetryFailedFrom != "" {
		config.RetryFailed = true
	}
	if err := config.validate(); err != nil {
		return err
	}
	if config.PathPrefix == "" {
		config.PathPrefix = "/"
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
	fmt.Printf("Discovered %d unique token-table payloads.\n", len(sources))

	indexPath := filepath.Join(config.CacheDir, "index.json")
	index, err := loadIndex(indexPath)
	if err != nil {
		return err
	}
	sourcePayloads := make([]SourcePayloadRecord, 0, len(sources))
	if !config.DryRun {
		for _, source := range sources {
			record, err := persistPayload(ctx, config, index, source)
			if err != nil {
				record = SourcePayloadRecord{PageURL: source.PageURL, TokenTableURL: source.TokenTableURL, Error: err.Error()}
			}
			sourcePayloads = append(sourcePayloads, record)
		}
		if err := writeJSONAtomic(indexPath, index); err != nil {
			return fmt.Errorf("write cache index: %w", err)
		}
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
		SchemaVersion:  1,
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
	if receipt.SchemaVersion != 1 {
		return CrawlReceipt{}, fmt.Errorf("unsupported crawl receipt schema version %d", receipt.SchemaVersion)
	}
	if receipt.Options.DryRun {
		return CrawlReceipt{}, errors.New("cannot retry failed pages from a dry-run crawl receipt")
	}
	return receipt, nil
}

func failedPages(receipt CrawlReceipt) ([]string, error) {
	observations := make(map[string]Observation, len(receipt.Observations))
	payloadsByPage := make(map[string][]SourcePayloadRecord)
	successfulPayloads := make(map[string]bool)
	for _, observation := range receipt.Observations {
		observations[observation.PageURL] = observation
	}
	for _, payload := range receipt.SourcePayloads {
		payloadsByPage[payload.PageURL] = append(payloadsByPage[payload.PageURL], payload)
		if payload.TokenTableURL != "" && payload.Error == "" && payload.Hash != "" {
			successfulPayloads[payload.TokenTableURL] = true
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
		pageFailed := !observed || observation.Error != "" || observation.Status < http.StatusOK || observation.Status >= http.StatusBadRequest || observation.TokenTableURL == ""
		if !pageFailed && !successfulPayloads[observation.TokenTableURL] {
			pageFailed = true
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
	merged := make([]SourcePayloadRecord, 0, len(previous)+len(retried))
	for _, payload := range previous {
		if !retriedSet[payload.PageURL] {
			merged = append(merged, payload)
		}
	}
	return append(merged, retried...)
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
	fetched := 0
	for _, payload := range payloads {
		if payload.Hash != "" && payload.Error == "" {
			fetched++
		}
	}
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
	fetched := 0
	for _, payload := range payloads {
		if payload.Hash != "" && payload.Error == "" {
			fetched++
		}
	}
	observed := len(uniqueSources(observations))
	return len(pages) > 0 && successfulPageCount(observations) == len(pages) && observed > 0 && fetched == observed
}

func (config Config) validate() error {
	if _, err := url.ParseRequestURI(config.Seed); err != nil {
		return fmt.Errorf("invalid seed URL: %w", err)
	}
	if config.Discover && config.UpdatePageCatalog {
		return errors.New("discover and update-page-catalog cannot be used together")
	}
	if config.NoPageCatalog && config.UpdatePageCatalog {
		return errors.New("no-page-catalog and update-page-catalog cannot be used together")
	}
	if !config.SpecsOnly && config.UpdatePageCatalog {
		return errors.New("update-page-catalog requires specs-only discovery")
	}
	if config.RetryFailed && config.DryRun {
		return errors.New("retry-failed cannot be combined with dry-run")
	}
	if config.RetryFailed && (config.Discover || config.UpdatePageCatalog || config.NoPageCatalog) {
		return errors.New("retry-failed uses the previous crawl receipt and cannot be combined with discovery or page-catalog selection flags")
	}
	if config.MaxPages < 1 || config.Concurrency < 1 || config.PageTimeout < time.Second {
		return errors.New("max-pages, concurrency, and page-timeout must be positive (page-timeout at least 1s)")
	}
	return nil
}

func crawlOptions(config Config) CrawlOptions {
	pathPrefix := config.PathPrefix
	if pathPrefix == "" {
		pathPrefix = "/"
	}
	return CrawlOptions{
		Seed:              config.Seed,
		Sitemap:           config.Sitemap,
		PathPrefix:        pathPrefix,
		SpecsOnly:         config.SpecsOnly,
		PageCatalogPath:   config.PageCatalogPath,
		Discover:          config.Discover,
		UpdatePageCatalog: config.UpdatePageCatalog,
		NoPageCatalog:     config.NoPageCatalog,
		MaxPages:          config.MaxPages,
		Concurrency:       config.Concurrency,
		PageTimeout:       config.PageTimeout.String(),
		CacheDir:          config.CacheDir,
		ResultsDir:        config.ResultsDir,
		Refresh:           config.Refresh,
		DryRun:            config.DryRun,
		RetryFailed:       config.RetryFailed,
		RetryFailedFrom:   config.RetryFailedFrom,
		Headed:            config.Headed,
		ChromePath:        config.ChromePath,
	}
}

func DiscoverPages(ctx context.Context, config Config) ([]string, error) {
	pages, _, err := discoverPages(ctx, config)
	return pages, err
}

func discoverPages(ctx context.Context, config Config) ([]string, DiscoveryReport, error) {
	if config.PageCatalogPath == "" {
		config.PageCatalogPath = DefaultPageCatalogPath
	}
	if !config.NoPageCatalog && !config.Discover && !config.UpdatePageCatalog {
		if !config.SpecsOnly {
			return nil, DiscoveryReport{}, errors.New("the curated page catalog contains specification pages; use --no-page-catalog for non-spec discovery")
		}
		pages, err := loadPageCatalog(config.PageCatalogPath, config.MaxPages)
		if err != nil {
			return nil, DiscoveryReport{}, err
		}
		return pages, DiscoveryReport{Mode: "page-catalog", PageCatalogPath: config.PageCatalogPath}, nil
	}

	liveConfig := config
	liveConfig.MaxPages = 0
	livePages, source, err := discoverLivePages(ctx, liveConfig)
	if err != nil {
		return nil, DiscoveryReport{}, err
	}
	report := DiscoveryReport{Mode: "live", Source: source}
	if !config.NoPageCatalog {
		report.PageCatalogPath = config.PageCatalogPath
		if pageCatalogURLs, loadErr := loadPageCatalog(config.PageCatalogPath, 0); loadErr == nil {
			report.Added, report.Removed = pageDiff(pageCatalogURLs, livePages)
		} else if !errors.Is(loadErr, os.ErrNotExist) {
			return nil, DiscoveryReport{}, loadErr
		}
	}
	if config.UpdatePageCatalog {
		if config.DryRun {
			report.Mode = "page-catalog-update-preview"
		} else {
			if err := writePageCatalog(config.PageCatalogPath, source, livePages); err != nil {
				return nil, DiscoveryReport{}, err
			}
			report.Mode = "page-catalog-updated"
		}
	}
	return limitPages(livePages, config.MaxPages), report, nil
}

func loadPageCatalog(path string, maxPages int) ([]string, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var pageCatalog PageCatalog
	if err := json.Unmarshal(body, &pageCatalog); err != nil {
		return nil, fmt.Errorf("parse page catalog: %w", err)
	}
	if pageCatalog.SchemaVersion != 1 {
		return nil, fmt.Errorf("unsupported page catalog schema version %d", pageCatalog.SchemaVersion)
	}
	seen := map[string]bool{}
	pages := make([]string, 0, len(pageCatalog.Pages))
	for _, pageEntry := range pageCatalog.Pages {
		parsed, err := url.Parse(pageEntry.URL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" || !isSpecsPage(pageEntry.URL) {
			return nil, fmt.Errorf("invalid specification page in page catalog: %q", pageEntry.URL)
		}
		if !seen[pageEntry.URL] {
			seen[pageEntry.URL] = true
			pages = append(pages, pageEntry.URL)
		}
	}
	sort.Strings(pages)
	return limitPages(pages, maxPages), nil
}

func writePageCatalog(path, source string, pages []string) error {
	entries := make([]PageCatalogEntry, 0, len(pages))
	for _, pageURL := range pages {
		entries = append(entries, PageCatalogEntry{URL: pageURL, Component: componentName(pageURL)})
	}
	pageCatalog := PageCatalog{
		SchemaVersion: 1,
		Source:        source,
		Scope:         "/components/<component>/specs",
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339Nano),
		Pages:         entries,
	}
	return writeJSONAtomic(path, pageCatalog)
}

func componentName(page string) string {
	parsed, err := url.Parse(page)
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) >= 3 && parts[len(parts)-1] == "specs" {
		return parts[len(parts)-2]
	}
	return ""
}

func pageDiff(previous, current []string) (added, removed []string) {
	previousSet := make(map[string]bool, len(previous))
	currentSet := make(map[string]bool, len(current))
	for _, catalogPage := range previous {
		previousSet[catalogPage] = true
	}
	for _, catalogPage := range current {
		currentSet[catalogPage] = true
	}
	for catalogPage := range currentSet {
		if !previousSet[catalogPage] {
			added = append(added, catalogPage)
		}
	}
	for catalogPage := range previousSet {
		if !currentSet[catalogPage] {
			removed = append(removed, catalogPage)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	return added, removed
}

func limitPages(pages []string, maxPages int) []string {
	if maxPages > 0 && len(pages) > maxPages {
		return pages[:maxPages]
	}
	return pages
}

func discoverLivePages(ctx context.Context, config Config) ([]string, string, error) {
	seed, err := url.Parse(config.Seed)
	if err != nil {
		return nil, "", err
	}
	client := &http.Client{Timeout: 20 * time.Second}
	var sitemaps []string
	if config.Sitemap != "" {
		sitemaps = []string{config.Sitemap}
	} else if robots, requestErr := fetchBody(ctx, client, seed.Scheme+"://"+seed.Host+"/robots.txt"); requestErr == nil {
		sitemaps = parseRobotsSitemaps(robots, seed)
	}
	if len(sitemaps) == 0 {
		sitemaps = []string{seed.Scheme + "://" + seed.Host + "/sitemap.xml"}
	}

	pages := map[string]bool{}
	seedPage := canonicalPageURL(config.Seed, seed, config.PathPrefix)
	if seedPage == "" {
		seedPage = config.Seed
	}
	if config.SpecsOnly && !isSpecsPage(seedPage) {
		return nil, "", fmt.Errorf("seed page must end with /specs when specs-only discovery is enabled: %s", seedPage)
	}
	pages[seedPage] = true
	for _, sitemap := range sitemaps {
		for _, candidate := range collectSitemap(ctx, client, sitemap, 3, map[string]bool{}) {
			canonical := canonicalPageURL(candidate, seed, config.PathPrefix)
			if canonical != "" && (!config.SpecsOnly || isSpecsPage(canonical)) {
				pages[canonical] = true
			}
			if config.MaxPages > 0 && len(pages) >= config.MaxPages {
				break
			}
		}
		if config.MaxPages > 0 && len(pages) >= config.MaxPages {
			break
		}
	}
	if len(pages) == 1 {
		pages = crawlLinks(ctx, client, pages, seed, config.PathPrefix, config.SpecsOnly, config.MaxPages)
	}

	result := make([]string, 0, len(pages))
	for catalogPage := range pages {
		result = append(result, catalogPage)
	}
	sort.Strings(result)
	result = limitPages(result, config.MaxPages)
	return result, sitemaps[0], nil
}

func fetchBody(ctx context.Context, client *http.Client, target string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("%s", response.Status)
	}
	return io.ReadAll(response.Body)
}

func parseRobotsSitemaps(body []byte, origin *url.URL) []string {
	result := robotsFile{}
	for _, line := range strings.Split(string(body), "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 && strings.EqualFold(strings.TrimSpace(parts[0]), "sitemap") {
			if parsed, err := url.Parse(strings.TrimSpace(parts[1])); err == nil {
				result.Sitemaps = append(result.Sitemaps, origin.ResolveReference(parsed).String())
			}
		}
	}
	return result.Sitemaps
}

func collectSitemap(ctx context.Context, client *http.Client, target string, depth int, seen map[string]bool) []string {
	if depth < 0 || seen[target] {
		return nil
	}
	seen[target] = true
	body, err := fetchBody(ctx, client, target)
	if err != nil {
		return nil
	}
	var document sitemapDocument
	if xml.Unmarshal(body, &document) != nil {
		return nil
	}
	if len(document.Sitemaps) == 0 {
		result := make([]string, 0, len(document.URLs))
		for _, item := range document.URLs {
			result = append(result, item.Loc)
		}
		return result
	}
	result := []string{}
	for _, nested := range document.Sitemaps {
		result = append(result, collectSitemap(ctx, client, nested.Loc, depth-1, seen)...)
	}
	return result
}

func canonicalPageURL(candidate string, origin *url.URL, pathPrefix string) string {
	parsed, err := url.Parse(candidate)
	if err != nil || parsed.Host != "" && parsed.Host != origin.Host {
		return ""
	}
	resolved := origin.ResolveReference(parsed)
	if resolved.Scheme != origin.Scheme || resolved.Host != origin.Host || !strings.HasPrefix(resolved.Path, pathPrefix) || strings.Contains(resolved.Path, ".") {
		return ""
	}
	resolved.Fragment = ""
	resolved.RawQuery = ""
	return resolved.String()
}

func isSpecsPage(page string) bool {
	parsed, err := url.Parse(page)
	if err != nil {
		return false
	}
	return strings.HasSuffix(strings.TrimSuffix(parsed.Path, "/"), "/specs")
}

func crawlLinks(ctx context.Context, client *http.Client, pages map[string]bool, origin *url.URL, pathPrefix string, specsOnly bool, maxPages int) map[string]bool {
	queue := []string{}
	for catalogPage := range pages {
		queue = append(queue, catalogPage)
	}
	visited := map[string]bool{}
	for len(queue) > 0 && (maxPages <= 0 || len(pages) < maxPages) {
		catalogPage := queue[0]
		queue = queue[1:]
		if visited[catalogPage] {
			continue
		}
		visited[catalogPage] = true
		body, err := fetchBody(ctx, client, catalogPage)
		if err != nil {
			continue
		}
		document, err := html.Parse(strings.NewReader(string(body)))
		if err != nil {
			continue
		}
		for _, link := range documentLinks(document) {
			canonical := canonicalPageURL(link, origin, pathPrefix)
			if canonical != "" && (!specsOnly || isSpecsPage(canonical)) && !pages[canonical] {
				pages[canonical] = true
				queue = append(queue, canonical)
				if maxPages > 0 && len(pages) >= maxPages {
					break
				}
			}
		}
	}
	return pages
}

func documentLinks(root *html.Node) []string {
	var result []string
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" {
			for _, attribute := range node.Attr {
				if attribute.Key == "href" {
					result = append(result, attribute.Val)
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(root)
	return result
}

type pageSource struct {
	PageURL       string
	TokenTableURL string
}

func uniqueSources(observations []Observation) []pageSource {
	seen := map[string]bool{}
	result := []pageSource{}
	for _, observation := range observations {
		if observation.TokenTableURL != "" && !seen[observation.TokenTableURL] {
			seen[observation.TokenTableURL] = true
			result = append(result, pageSource{PageURL: observation.PageURL, TokenTableURL: observation.TokenTableURL})
		}
	}
	return result
}

func observePages(parent context.Context, config Config, pages []string) ([]Observation, error) {
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
	}
	if config.ChromePath != "" {
		allocatorOptions = append(allocatorOptions, chromedp.ExecPath(config.ChromePath))
	}
	allocatorContext, cancelAllocator := chromedp.NewExecAllocator(parent, allocatorOptions...)
	defer cancelAllocator()
	browserContext, cancelBrowser := chromedp.NewContext(allocatorContext)
	defer cancelBrowser()

	return mapLimit(pages, config.Concurrency, func(pageURL string) Observation {
		return observePage(browserContext, pageURL, config.PageTimeout)
	})
}

func observePage(browserContext context.Context, pageURL string, timeout time.Duration) Observation {
	pageContext, cancel := chromedp.NewContext(browserContext)
	defer cancel()
	observation := Observation{PageURL: pageURL}
	responseURL := make(chan string, 1)
	var responseOnce sync.Once
	var statusMu sync.Mutex
	mainStatus := 0
	chromedp.ListenTarget(pageContext, func(event interface{}) {
		if response, ok := event.(*network.EventResponseReceived); ok {
			if response.Response.URL == pageURL {
				statusMu.Lock()
				mainStatus = int(response.Response.Status)
				statusMu.Unlock()
			}
			if tokenTableRE.MatchString(response.Response.URL) {
				responseOnce.Do(func() { responseURL <- response.Response.URL })
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
		select {
		case observation.TokenTableURL = <-responseURL:
		case <-timer.C:
		case <-operationContext.Done():
			observation.Error = operationContext.Err().Error()
		}
	}
	statusMu.Lock()
	observation.Status = mainStatus
	statusMu.Unlock()
	return observation
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

func loadIndex(path string) (CacheIndex, error) {
	index := CacheIndex{SchemaVersion: 1, SourcePayloads: map[string]SourcePayloadRecord{}}
	body, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return index, nil
	}
	if err != nil {
		return index, err
	}
	if err := json.Unmarshal(body, &index); err != nil {
		return index, fmt.Errorf("parse cache index: %w", err)
	}
	if index.SourcePayloads == nil {
		index.SourcePayloads = map[string]SourcePayloadRecord{}
	}
	return index, nil
}

func persistPayload(ctx context.Context, config Config, index CacheIndex, source pageSource) (SourcePayloadRecord, error) {
	if cached, ok := index.SourcePayloads[source.TokenTableURL]; ok && !config.Refresh {
		path := filepath.Join(config.CacheDir, cached.Artifact)
		if _, err := os.Stat(path); err == nil {
			cached.PageURL = source.PageURL
			cached.Reused = true
			return cached, nil
		}
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, source.TokenTableURL, nil)
	if err != nil {
		return SourcePayloadRecord{}, err
	}
	request.Header.Set("Accept", "application/json")
	response, err := (&http.Client{Timeout: 60 * time.Second}).Do(request)
	if err != nil {
		return SourcePayloadRecord{}, err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode >= http.StatusBadRequest {
		return SourcePayloadRecord{}, fmt.Errorf("%s", response.Status)
	}
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return SourcePayloadRecord{}, err
	}
	hashBytes := sha256.Sum256(raw)
	hash := hex.EncodeToString(hashBytes[:])
	artifact := filepath.Join("source-payloads", hash[:2], hash+".json.gz")
	if err := writeGzipIfAbsent(filepath.Join(config.CacheDir, artifact), raw); err != nil {
		return SourcePayloadRecord{}, err
	}
	record := SourcePayloadRecord{
		PageURL:       source.PageURL,
		TokenTableURL: source.TokenTableURL,
		Hash:          hash,
		Artifact:      artifact,
		Bytes:         len(raw),
		RetrievedAt:   time.Now().UTC().Format(time.RFC3339Nano),
		ETag:          response.Header.Get("ETag"),
		LastModified:  response.Header.Get("Last-Modified"),
	}
	index.SourcePayloads[source.TokenTableURL] = record
	return record, nil
}

func writeGzipIfAbsent(path string, body []byte) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".payload-*.tmp")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	writer := gzip.NewWriter(temporary)
	if _, err := writer.Write(body); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	return nil
}

func writeJSONAtomic(path string, value interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".json-*.tmp")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if _, err := temporary.Write(append(body, '\n')); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}
