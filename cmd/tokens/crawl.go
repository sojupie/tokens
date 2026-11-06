// The crawl command translates CLI options into capture configuration; capture mechanics belong to source/crawl.
package main

import (
	"context"

	"errors"
	"flag"
	"fmt"

	"os"

	"strings"
	"time"

	"github.com/sojupie/tokens/source/crawl"
)

func runCrawl(args []string) error {
	var config crawl.Config
	flags := flag.NewFlagSet("tokens crawl", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() {
		_, _ = fmt.Fprintln(flags.Output(), "Usage: tokens crawl [flags]")
		_, _ = fmt.Fprintln(flags.Output(), "\nCrawls reviewed Material token-source pages and downloads source payloads.")
		_, _ = fmt.Fprintln(flags.Output(), "")
		flags.PrintDefaults()
	}
	flags.StringVar(&config.Seed, "seed", crawl.DefaultSeed, "starting Material page")
	flags.StringVar(&config.Sitemap, "sitemap", "", "sitemap or sitemap-index URL override")
	flags.StringVar(&config.PathPrefix, "path-prefix", "/", "same-origin page path prefix used during live discovery")
	flags.StringVar(&config.Scope, "scope", crawl.ScopeComponentSpecs, "live discovery scope: component-specs or token-source-pages")
	flags.StringVar(&config.PageCatalogPath, "page-catalog", crawl.DefaultPageCatalogPath, "page URL catalog used by default")
	flags.BoolVar(&config.Discover, "discover", false, "bypass the page URL catalog and report page-catalog drift")
	flags.BoolVar(&config.UpdatePageCatalog, "update-page-catalog", false, "replace the page URL catalog with reviewed live discovery")
	flags.BoolVar(&config.NoPageCatalog, "no-page-catalog", false, "bypass the page URL catalog without comparing or updating it")
	flags.IntVar(&config.MaxPages, "max-pages", 500, "maximum candidate pages")
	flags.IntVar(&config.Concurrency, "concurrency", 1, "parallel browser pages")
	flags.DurationVar(&config.PageTimeout, "page-timeout", 30*time.Second, "maximum observation time per page; finishes early after network activity settles")
	flags.StringVar(&config.CacheDir, "cache-dir", ".cache/material-token-table-responses", "source-payload cache directory")
	flags.StringVar(&config.ResultsDir, "results-dir", "crawl-results", "crawl-receipt directory")
	flags.BoolVar(&config.Refresh, "refresh", false, "re-download source payloads already in the cache")
	flags.BoolVar(&config.DryRun, "dry-run", false, "discover pages and payload URLs without downloading payloads")
	flags.BoolVar(&config.RetryFailed, "retry-failed", false, "retry failed pages from the newest crawl receipt")
	flags.StringVar(&config.RetryFailedFrom, "retry-failed-from", "", "retry failed pages from this crawl receipt")
	flags.BoolVar(&config.Headed, "headed", false, "show Chromium while observing pages")
	flags.StringVar(&config.ChromePath, "chrome-path", "", "optional Chrome/Chromium executable path")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() > 0 {
		return fmt.Errorf("crawl does not accept positional arguments: %s", strings.Join(flags.Args(), " "))
	}
	return crawl.Run(context.Background(), config)
}
