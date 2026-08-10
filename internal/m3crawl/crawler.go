package m3crawl

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
	"github.com/chromedp/chromedp"
	"golang.org/x/net/html"
)

const (
	DefaultSeed        = "https://m3.material.io/components/buttons/specs"
	DefaultCatalogPath = "sources/catalog/m3-spec-pages.json"
	tokenTablePattern  = `(?i)/TOKEN_TABLE(?:\.[^/]+)?\.json(?:\?.*)?$`
)

var tokenTableRE = regexp.MustCompile(tokenTablePattern)

type Config struct {
	Seed          string        `json:"seed"`
	Sitemap       string        `json:"sitemap,omitempty"`
	PathPrefix    string        `json:"pathPrefix"`
	SpecsOnly     bool          `json:"specsOnly"`
	CatalogPath   string        `json:"catalogPath"`
	Discover      bool          `json:"discover"`
	UpdateCatalog bool          `json:"updateCatalog"`
	NoCatalog     bool          `json:"noCatalog"`
	MaxPages      int           `json:"maxPages"`
	Concurrency   int           `json:"concurrency"`
	PageTimeout   time.Duration `json:"-"`
	CacheDir      string        `json:"cacheDir"`
	RunDir        string        `json:"runDir"`
	Refresh       bool          `json:"refresh"`
	DryRun        bool          `json:"dryRun"`
	Headed        bool          `json:"headed"`
	ChromePath    string        `json:"chromePath,omitempty"`
}

type ManifestOptions struct {
	Seed          string `json:"seed"`
	Sitemap       string `json:"sitemap,omitempty"`
	PathPrefix    string `json:"pathPrefix"`
	SpecsOnly     bool   `json:"specsOnly"`
	CatalogPath   string `json:"catalogPath"`
	Discover      bool   `json:"discover"`
	UpdateCatalog bool   `json:"updateCatalog"`
	NoCatalog     bool   `json:"noCatalog"`
	MaxPages      int    `json:"maxPages"`
	Concurrency   int    `json:"concurrency"`
	PageTimeout   string `json:"pageTimeout"`
	CacheDir      string `json:"cacheDir"`
	RunDir        string `json:"runDir"`
	Refresh       bool   `json:"refresh"`
	DryRun        bool   `json:"dryRun"`
	Headed        bool   `json:"headed"`
	ChromePath    string `json:"chromePath,omitempty"`
}

type Observation struct {
	PageURL       string `json:"pageUrl"`
	Status        int    `json:"status"`
	TokenTableURL string `json:"tokenTableUrl,omitempty"`
	Title         string `json:"title,omitempty"`
	Error         string `json:"error,omitempty"`
}

type PayloadRecord struct {
	PageURL       string `json:"pageUrl"`
	TokenTableURL string `json:"tokenTableUrl"`
	Hash          string `json:"hash"`
	Artifact      string `json:"artifact"`
	Bytes         int    `json:"bytes"`
	RetrievedAt   string `json:"retrievedAt"`
	ETag          string `json:"etag,omitempty"`
	LastModified  string `json:"lastModified,omitempty"`
	Reused        bool   `json:"reused"`
	DerivedHash   string `json:"derivedHash,omitempty"`
	TokenSetCount int    `json:"tokenSetCount,omitempty"`
	TokenCount    int    `json:"tokenCount,omitempty"`
	Error         string `json:"error,omitempty"`
}

type CacheIndex struct {
	SchemaVersion int                      `json:"schemaVersion"`
	Payloads      map[string]PayloadRecord `json:"payloads"`
}

type RunManifest struct {
	SchemaVersion int             `json:"schemaVersion"`
	RunID         string          `json:"runId"`
	StartedAt     string          `json:"startedAt"`
	CompletedAt   string          `json:"completedAt"`
	Mode          string          `json:"mode"`
	Options       ManifestOptions `json:"options"`
	Discovery     DiscoveryReport `json:"discovery"`
	Pages         []string        `json:"pages"`
	Observations  []Observation   `json:"observations"`
	Payloads      []PayloadRecord `json:"payloads"`
}

type CatalogPage struct {
	URL       string `json:"url"`
	Component string `json:"component"`
}

type PageCatalog struct {
	SchemaVersion int           `json:"schemaVersion"`
	Source        string        `json:"source"`
	Scope         string        `json:"scope"`
	GeneratedAt   string        `json:"generatedAt"`
	Pages         []CatalogPage `json:"pages"`
}

type DiscoveryReport struct {
	Mode        string   `json:"mode"`
	CatalogPath string   `json:"catalogPath,omitempty"`
	Source      string   `json:"source,omitempty"`
	Added       []string `json:"added,omitempty"`
	Removed     []string `json:"removed,omitempty"`
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

type tokenTable struct {
	System struct {
		TokenSets []struct {
			Name         string `json:"name"`
			TokenSetName string `json:"tokenSetName"`
			DisplayName  string `json:"displayName"`
			State        string `json:"state"`
		} `json:"tokenSets"`
		Tokens []struct {
			TokenName string `json:"tokenName"`
		} `json:"tokens"`
	} `json:"system"`
}

func Run(ctx context.Context, config Config) error {
	if err := config.validate(); err != nil {
		return err
	}
	if config.PathPrefix == "" {
		config.PathPrefix = "/"
	}
	if config.CatalogPath == "" {
		config.CatalogPath = DefaultCatalogPath
	}
	started := time.Now().UTC()
	pages, discovery, err := discoverPages(ctx, config)
	if err != nil {
		return fmt.Errorf("discover pages: %w", err)
	}
	fmt.Printf("Discovered %d candidate pages.\n", len(pages))
	fmt.Printf("Page source: %s\n", discovery.Mode)
	if len(discovery.Added) > 0 || len(discovery.Removed) > 0 {
		fmt.Printf("Catalog drift: %d added, %d removed.\n", len(discovery.Added), len(discovery.Removed))
	}

	observations, err := observePages(ctx, config, pages)
	if err != nil {
		return fmt.Errorf("observe pages: %w", err)
	}
	sources := uniqueSources(observations)
	fmt.Printf("Discovered %d unique token-table payloads.\n", len(sources))

	indexPath := filepath.Join(config.CacheDir, "index.json")
	index, err := loadIndex(indexPath)
	if err != nil {
		return err
	}
	payloads := make([]PayloadRecord, 0, len(sources))
	if !config.DryRun {
		for _, source := range sources {
			record, err := persistPayload(ctx, config, index, source)
			if err != nil {
				record = PayloadRecord{PageURL: source.PageURL, TokenTableURL: source.TokenTableURL, Error: err.Error()}
			} else {
				record.TokenSetCount, record.TokenCount = inspectTokenTable(record, config.CacheDir)
			}
			payloads = append(payloads, record)
		}
		if err := writeJSONAtomic(indexPath, index); err != nil {
			return fmt.Errorf("write cache index: %w", err)
		}
	}

	runID := started.Format("20060102T150405.000000000Z")
	manifest := RunManifest{
		SchemaVersion: 1,
		RunID:         runID,
		StartedAt:     started.Format(time.RFC3339Nano),
		CompletedAt:   time.Now().UTC().Format(time.RFC3339Nano),
		Mode:          "all",
		Options:       manifestOptions(config),
		Discovery:     discovery,
		Pages:         pages,
		Observations:  observations,
		Payloads:      payloads,
	}
	manifestPath := filepath.Join(config.RunDir, runID+".json")
	if err := writeJSONAtomic(manifestPath, manifest); err != nil {
		return fmt.Errorf("write run manifest: %w", err)
	}
	fmt.Printf("Wrote run manifest to %s\n", manifestPath)
	return nil
}

func (config Config) validate() error {
	if _, err := url.ParseRequestURI(config.Seed); err != nil {
		return fmt.Errorf("invalid seed URL: %w", err)
	}
	if config.Discover && config.UpdateCatalog {
		return errors.New("discover and update-catalog cannot be used together")
	}
	if config.NoCatalog && config.UpdateCatalog {
		return errors.New("no-catalog and update-catalog cannot be used together")
	}
	if !config.SpecsOnly && config.UpdateCatalog {
		return errors.New("update-catalog requires specs-only discovery")
	}
	if config.MaxPages < 1 || config.Concurrency < 1 || config.PageTimeout < time.Second {
		return errors.New("max-pages, concurrency, and page-timeout must be positive (page-timeout at least 1s)")
	}
	return nil
}

func manifestOptions(config Config) ManifestOptions {
	pathPrefix := config.PathPrefix
	if pathPrefix == "" {
		pathPrefix = "/"
	}
	return ManifestOptions{
		Seed:          config.Seed,
		Sitemap:       config.Sitemap,
		PathPrefix:    pathPrefix,
		SpecsOnly:     config.SpecsOnly,
		CatalogPath:   config.CatalogPath,
		Discover:      config.Discover,
		UpdateCatalog: config.UpdateCatalog,
		NoCatalog:     config.NoCatalog,
		MaxPages:      config.MaxPages,
		Concurrency:   config.Concurrency,
		PageTimeout:   config.PageTimeout.String(),
		CacheDir:      config.CacheDir,
		RunDir:        config.RunDir,
		Refresh:       config.Refresh,
		DryRun:        config.DryRun,
		Headed:        config.Headed,
		ChromePath:    config.ChromePath,
	}
}

func DiscoverPages(ctx context.Context, config Config) ([]string, error) {
	pages, _, err := discoverPages(ctx, config)
	return pages, err
}

func discoverPages(ctx context.Context, config Config) ([]string, DiscoveryReport, error) {
	if config.CatalogPath == "" {
		config.CatalogPath = DefaultCatalogPath
	}
	if !config.NoCatalog && !config.Discover && !config.UpdateCatalog {
		if !config.SpecsOnly {
			return nil, DiscoveryReport{}, errors.New("the curated catalog contains specs pages; use --no-catalog for non-spec discovery")
		}
		pages, err := loadCatalogPages(config.CatalogPath, config.MaxPages)
		if err != nil {
			return nil, DiscoveryReport{}, err
		}
		return pages, DiscoveryReport{Mode: "catalog", CatalogPath: config.CatalogPath}, nil
	}

	liveConfig := config
	liveConfig.MaxPages = 0
	livePages, source, err := discoverLivePages(ctx, liveConfig)
	if err != nil {
		return nil, DiscoveryReport{}, err
	}
	report := DiscoveryReport{Mode: "live", Source: source}
	if !config.NoCatalog {
		report.CatalogPath = config.CatalogPath
		if catalogPages, loadErr := loadCatalogPages(config.CatalogPath, 0); loadErr == nil {
			report.Added, report.Removed = pageDiff(catalogPages, livePages)
		} else if !errors.Is(loadErr, os.ErrNotExist) {
			return nil, DiscoveryReport{}, loadErr
		}
	}
	if config.UpdateCatalog {
		if config.DryRun {
			report.Mode = "catalog-update-preview"
		} else {
			if err := writeCatalog(config.CatalogPath, source, livePages); err != nil {
				return nil, DiscoveryReport{}, err
			}
			report.Mode = "catalog-updated"
		}
	}
	return limitPages(livePages, config.MaxPages), report, nil
}

func loadCatalogPages(path string, maxPages int) ([]string, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var catalog PageCatalog
	if err := json.Unmarshal(body, &catalog); err != nil {
		return nil, fmt.Errorf("parse page catalog: %w", err)
	}
	if catalog.SchemaVersion != 1 {
		return nil, fmt.Errorf("unsupported page catalog schema version %d", catalog.SchemaVersion)
	}
	seen := map[string]bool{}
	pages := make([]string, 0, len(catalog.Pages))
	for _, page := range catalog.Pages {
		parsed, err := url.Parse(page.URL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" || !isSpecsPage(page.URL) {
			return nil, fmt.Errorf("invalid specs page in catalog: %q", page.URL)
		}
		if !seen[page.URL] {
			seen[page.URL] = true
			pages = append(pages, page.URL)
		}
	}
	sort.Strings(pages)
	return limitPages(pages, maxPages), nil
}

func writeCatalog(path, source string, pages []string) error {
	entries := make([]CatalogPage, 0, len(pages))
	for _, page := range pages {
		entries = append(entries, CatalogPage{URL: page, Component: componentName(page)})
	}
	catalog := PageCatalog{
		SchemaVersion: 1,
		Source:        source,
		Scope:         "/components/<component>/specs",
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339Nano),
		Pages:         entries,
	}
	return writeJSONAtomic(path, catalog)
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
	for _, page := range previous {
		previousSet[page] = true
	}
	for _, page := range current {
		currentSet[page] = true
	}
	for page := range currentSet {
		if !previousSet[page] {
			added = append(added, page)
		}
	}
	for page := range previousSet {
		if !currentSet[page] {
			removed = append(removed, page)
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
	sitemaps := []string{}
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
	for page := range pages {
		result = append(result, page)
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
	defer response.Body.Close()
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
	for page := range pages {
		queue = append(queue, page)
	}
	visited := map[string]bool{}
	for len(queue) > 0 && (maxPages <= 0 || len(pages) < maxPages) {
		page := queue[0]
		queue = queue[1:]
		if visited[page] {
			continue
		}
		visited[page] = true
		body, err := fetchBody(ctx, client, page)
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
	result := []string{}
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
	allocatorOptions := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", !config.Headed))
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
	var title string
	err := chromedp.Run(operationContext,
		network.Enable(),
		chromedp.Navigate(pageURL),
		chromedp.Sleep(timeout),
		chromedp.Title(&title),
	)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		observation.Error = err.Error()
	}
	statusMu.Lock()
	observation.Status = mainStatus
	statusMu.Unlock()
	observation.Title = title
	select {
	case observation.TokenTableURL = <-responseURL:
	default:
	}
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
		wait.Add(1)
		go func() {
			defer wait.Done()
			for index := range jobs {
				results[index] = worker(values[index])
			}
		}()
	}
	for index := range values {
		jobs <- index
	}
	close(jobs)
	wait.Wait()
	return results, nil
}

func loadIndex(path string) (CacheIndex, error) {
	index := CacheIndex{SchemaVersion: 1, Payloads: map[string]PayloadRecord{}}
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
	if index.Payloads == nil {
		index.Payloads = map[string]PayloadRecord{}
	}
	return index, nil
}

func persistPayload(ctx context.Context, config Config, index CacheIndex, source pageSource) (PayloadRecord, error) {
	if cached, ok := index.Payloads[source.TokenTableURL]; ok && !config.Refresh {
		path := filepath.Join(config.CacheDir, cached.Artifact)
		if _, err := os.Stat(path); err == nil {
			cached.PageURL = source.PageURL
			cached.Reused = true
			return cached, nil
		}
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, source.TokenTableURL, nil)
	if err != nil {
		return PayloadRecord{}, err
	}
	request.Header.Set("Accept", "application/json")
	response, err := (&http.Client{Timeout: 60 * time.Second}).Do(request)
	if err != nil {
		return PayloadRecord{}, err
	}
	defer response.Body.Close()
	if response.StatusCode >= http.StatusBadRequest {
		return PayloadRecord{}, fmt.Errorf("%s", response.Status)
	}
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return PayloadRecord{}, err
	}
	hashBytes := sha256.Sum256(raw)
	hash := hex.EncodeToString(hashBytes[:])
	artifact := filepath.Join("payloads", hash[:2], hash+".json.gz")
	if err := writeGzipIfAbsent(filepath.Join(config.CacheDir, artifact), raw); err != nil {
		return PayloadRecord{}, err
	}
	record := PayloadRecord{
		PageURL:       source.PageURL,
		TokenTableURL: source.TokenTableURL,
		Hash:          hash,
		Artifact:      artifact,
		Bytes:         len(raw),
		RetrievedAt:   time.Now().UTC().Format(time.RFC3339Nano),
		ETag:          response.Header.Get("ETag"),
		LastModified:  response.Header.Get("Last-Modified"),
	}
	index.Payloads[source.TokenTableURL] = record
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
	defer os.Remove(temporaryPath)
	writer := gzip.NewWriter(temporary)
	if _, err := writer.Write(body); err != nil {
		temporary.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		temporary.Close()
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

func inspectTokenTable(record PayloadRecord, cacheDir string) (int, int) {
	path := record.Artifact
	if !filepath.IsAbs(path) {
		path = filepath.Join(cacheDir, path)
	}
	file, err := os.Open(path)
	if err != nil {
		return 0, 0
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		return 0, 0
	}
	defer reader.Close()
	body, err := io.ReadAll(reader)
	if err != nil {
		return 0, 0
	}
	var table tokenTable
	if json.Unmarshal(body, &table) != nil {
		return 0, 0
	}
	sets := 0
	for _, tokenSet := range table.System.TokenSets {
		if tokenSet.State == "ACTIVE" && strings.HasPrefix(tokenSet.TokenSetName, "md.comp.button") && !strings.HasPrefix(tokenSet.DisplayName, "[Deprecated]") {
			sets++
		}
	}
	tokens := 0
	for _, token := range table.System.Tokens {
		if strings.HasPrefix(token.TokenName, "md.comp.button") {
			tokens++
		}
	}
	return sets, tokens
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
	defer os.Remove(temporaryPath)
	if _, err := temporary.Write(append(body, '\n')); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}
