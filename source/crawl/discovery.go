// Discovery separates the curated page catalog from live sitemap and link discovery so a routine capture has a stable scope.
package crawl

import (
	"context"

	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"sort"
	"strings"

	"time"

	"golang.org/x/net/html"
)

type PageCatalogEntry struct {
	URL      string `json:"url"`
	Area     string `json:"area"`
	Subject  string `json:"subject"`
	PageType string `json:"pageType"`
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

func DiscoverPages(ctx context.Context, config Config) ([]string, error) {
	pages, _, err := discoverPages(ctx, config)
	return pages, err
}

func discoverPages(ctx context.Context, config Config) ([]string, DiscoveryReport, error) {
	if config.Scope == "" {
		config.Scope = ScopeComponentSpecs
	}
	if config.PathPrefix == "" {
		config.PathPrefix = "/"
	}
	if config.PageCatalogPath == "" {
		config.PageCatalogPath = DefaultPageCatalogPath
	}
	if !config.NoPageCatalog && !config.Discover && !config.UpdatePageCatalog {
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
	if err := ValidatePageCatalog(body); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &pageCatalog); err != nil {
		return nil, fmt.Errorf("parse page catalog: %w", err)
	}
	seen := map[string]bool{}
	pages := make([]string, 0, len(pageCatalog.Pages))
	for _, pageEntry := range pageCatalog.Pages {
		if !seen[pageEntry.URL] {
			seen[pageEntry.URL] = true
			pages = append(pages, pageEntry.URL)
		}
	}
	sort.Strings(pages)
	return limitPages(pages, maxPages), nil
}

func ValidatePageCatalog(body []byte) error {
	var pageCatalog PageCatalog
	if err := json.Unmarshal(body, &pageCatalog); err != nil {
		return fmt.Errorf("parse page catalog: %w", err)
	}
	if pageCatalog.SchemaVersion != 2 {
		return fmt.Errorf("unsupported page catalog schema version %d", pageCatalog.SchemaVersion)
	}
	for _, pageEntry := range pageCatalog.Pages {
		parsed, err := url.Parse(pageEntry.URL)
		area, subject, pageType := pageMetadata(pageEntry.URL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" || !isTokenSourcePage(pageEntry.URL) {
			return fmt.Errorf("invalid token-source page in page catalog: %q", pageEntry.URL)
		}
		if pageEntry.Area != area || pageEntry.Subject != subject || pageEntry.PageType != pageType {
			return fmt.Errorf("page catalog metadata does not match URL %q", pageEntry.URL)
		}
	}
	return nil
}

func writePageCatalog(path, source string, pages []string) error {
	entries := make([]PageCatalogEntry, 0, len(pages))
	for _, pageURL := range pages {
		area, subject, pageType := pageMetadata(pageURL)
		entries = append(entries, PageCatalogEntry{URL: pageURL, Area: area, Subject: subject, PageType: pageType})
	}
	pageCatalog := PageCatalog{
		SchemaVersion: 2,
		Source:        source,
		Scope:         "component specs and style/foundation token pages",
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339Nano),
		Pages:         entries,
	}
	return writeJSONAtomic(path, pageCatalog)
}

func pageMetadata(page string) (area, subject, pageType string) {
	parsed, err := url.Parse(page)
	if err != nil {
		return "", "", ""
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	// This token table lives under the static scheme guide, not a /tokens URL.
	if strings.Trim(parsed.Path, "/") == "styles/color/static/baseline" {
		return "styles", "color", "tokens"
	}
	if len(parts) >= 3 && parts[0] == "components" && parts[len(parts)-1] == "specs" {
		return "components", strings.Join(parts[1:len(parts)-1], "/"), "specs"
	}
	if len(parts) >= 3 && parts[0] == "styles" {
		last := parts[len(parts)-1]
		pageType := ""
		subject := parts[1 : len(parts)-1]
		switch {
		case last == "specs":
			pageType = "specs"
		case last == "tokens":
			pageType = "tokens"
		case strings.HasSuffix(last, "-tokens"):
			pageType = "tokens"
			subject = append(subject, strings.TrimSuffix(last, "-tokens"))
		case strings.HasSuffix(last, "-scale"):
			pageType = "tokens"
			subject = append(subject, strings.TrimSuffix(last, "-scale"))
		default:
			return "", "", ""
		}
		return "styles", strings.Join(subject, "/"), pageType
	}
	return "", "", ""
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
	if !isPageInScope(seedPage, config.Scope) {
		return nil, "", fmt.Errorf("seed page is outside discovery scope %q: %s", config.Scope, seedPage)
	}
	pages[seedPage] = true
	for _, sitemap := range sitemaps {
		for _, candidate := range collectSitemap(ctx, client, sitemap, 3, map[string]bool{}) {
			canonical := canonicalPageURL(candidate, seed, config.PathPrefix)
			if canonical != "" && isPageInScope(canonical, config.Scope) {
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
		pages = crawlLinks(ctx, client, pages, seed, config.PathPrefix, config.Scope, config.MaxPages)
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
	prefix := strings.TrimSuffix(pathPrefix, "/")
	inPrefix := prefix == "" || resolved.Path == prefix || strings.HasPrefix(resolved.Path, prefix+"/")
	if resolved.Scheme != origin.Scheme || resolved.Host != origin.Host || !inPrefix || strings.Contains(resolved.Path, ".") {
		return ""
	}
	resolved.Fragment = ""
	resolved.RawQuery = ""
	return resolved.String()
}

func isComponentSpecsPage(page string) bool {
	parsed, err := url.Parse(page)
	if err != nil {
		return false
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	return len(parts) >= 3 && parts[0] == "components" && parts[len(parts)-1] == "specs"
}

func isStyleTokenSourcePage(page string) bool {
	area, _, _ := pageMetadata(page)
	return area == "styles"
}

func isTokenSourcePage(page string) bool {
	return isComponentSpecsPage(page) || isStyleTokenSourcePage(page)
}

func isPageInScope(page, scope string) bool {
	switch scope {
	case ScopeComponentSpecs:
		return isComponentSpecsPage(page)
	case ScopeTokenSourcePages:
		return isTokenSourcePage(page)
	default:
		return false
	}
}

func crawlLinks(ctx context.Context, client *http.Client, pages map[string]bool, origin *url.URL, pathPrefix, scope string, maxPages int) map[string]bool {
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
			if canonical != "" && isPageInScope(canonical, scope) && !pages[canonical] {
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
