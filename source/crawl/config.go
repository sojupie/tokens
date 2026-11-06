// Crawl configuration keeps defaults and validation shared by normal captures and targeted retries.
package crawl

import (
	"errors"
	"fmt"

	"net/url"

	"regexp"

	"time"
)

const (
	DefaultSeed            = "https://m3.material.io/components/buttons/specs"
	DefaultPageCatalogPath = "inputs/page-catalog.json"
	ScopeComponentSpecs    = "component-specs"
	ScopeTokenSourcePages  = "token-source-pages"
	sourcePayloadPattern   = `(?i:/_dsm/data/)(?:[^/]+/)+[A-Z][A-Z0-9_-]*(?:\.[A-Za-z0-9_-]+)?\.json(?:\?.*)?$`
)

var sourcePayloadRE = regexp.MustCompile(sourcePayloadPattern)

type Config struct {
	Seed              string        `json:"seed"`
	Sitemap           string        `json:"sitemap,omitempty"`
	PathPrefix        string        `json:"pathPrefix"`
	Scope             string        `json:"scope"`
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
	Scope             string `json:"scope"`
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

func (config Config) validate() error {
	if _, err := url.ParseRequestURI(config.Seed); err != nil {
		return fmt.Errorf("invalid seed URL: %w", err)
	}
	if config.NoPageCatalog && config.UpdatePageCatalog {
		return errors.New("no-page-catalog and update-page-catalog cannot be used together")
	}
	if config.Scope != ScopeComponentSpecs && config.Scope != ScopeTokenSourcePages {
		return fmt.Errorf("unsupported discovery scope %q (use %q or %q)", config.Scope, ScopeComponentSpecs, ScopeTokenSourcePages)
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
		Scope:             config.Scope,
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
