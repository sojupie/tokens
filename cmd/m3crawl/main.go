package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"material-pain-3/internal/m3crawl"
)

func main() {
	var config m3crawl.Config
	flag.StringVar(&config.Seed, "seed", m3crawl.DefaultSeed, "starting Material page")
	flag.StringVar(&config.Sitemap, "sitemap", "", "sitemap or sitemap-index URL override")
	flag.StringVar(&config.PathPrefix, "path-prefix", "/components/", "same-origin page path prefix")
	flag.BoolVar(&config.SpecsOnly, "specs-only", true, "only open component /specs pages (use --specs-only=false to inspect other pages)")
	flag.StringVar(&config.CatalogPath, "catalog", m3crawl.DefaultCatalogPath, "curated page catalog used by default")
	flag.BoolVar(&config.Discover, "discover", false, "bypass the catalog, crawl live discovery, and report catalog drift")
	flag.BoolVar(&config.UpdateCatalog, "update-catalog", false, "crawl live discovery and replace the curated catalog")
	flag.BoolVar(&config.NoCatalog, "no-catalog", false, "bypass the curated catalog without comparing or updating it")
	flag.IntVar(&config.MaxPages, "max-pages", 500, "maximum candidate pages")
	flag.IntVar(&config.Concurrency, "concurrency", 2, "parallel browser pages")
	flag.DurationVar(&config.PageTimeout, "page-timeout", 10*time.Second, "time to observe each page")
	flag.StringVar(&config.CacheDir, "cache-dir", ".cache/m3", "content-addressed cache directory")
	flag.StringVar(&config.RunDir, "run-dir", "sources/runs/m3", "run manifest directory")
	flag.BoolVar(&config.Refresh, "refresh", false, "refetch payloads already in the cache")
	flag.BoolVar(&config.DryRun, "dry-run", false, "discover pages and payload URLs without fetching payloads")
	flag.BoolVar(&config.Headed, "headed", false, "show Chromium while observing pages")
	flag.StringVar(&config.ChromePath, "chrome-path", "", "optional Chrome/Chromium executable path")
	var all bool
	flag.BoolVar(&all, "all", false, "legacy alias for --no-catalog")
	flag.Parse()
	if all {
		config.NoCatalog = true
	}

	if err := m3crawl.Run(context.Background(), config); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
